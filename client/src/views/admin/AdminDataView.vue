<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type BackupInfo, type ImportResult } from '../../api';
import { useDialogStore } from '../../stores/dialog';
import { useAuthStore } from '../../stores/auth';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSeg from './studio/StSeg.vue';
import StSwitch from './studio/StSwitch.vue';
import StModal from './studio/StModal.vue';
import { refreshCounts, saveBlob } from './studio/state';
import { useRouter } from 'vue-router';
import { toast } from './studio/toast';
import { dateTimeText, formatSize, relTime, sizeParts } from './studio/format';
import { looksLikeBackup } from '../../utils/zipEntries';

/**
 * 数据备份：立即备份（进度环）、备份记录（下载 / 恢复 / 删除）、自动备份间隔；
 * 迁移：从备份包恢复、导出为 Markdown、从 Hexo / Hugo / Jekyll 导入（先识别预览再导入）。
 */
const { t } = useI18n();
const dialog = useDialogStore();
const router = useRouter();

const backups = ref<BackupInfo[]>([]);
const loaded = ref(false);
const running = ref(false);
const progress = ref(0);
const fresh = ref('');
const autoHours = ref(0);
let timer = 0;

const latest = computed(() => backups.value[0] ?? null);
const totalSize = computed(() => backups.value.reduce((s, b) => s + b.size, 0));

async function load(): Promise<void> {
  try {
    const [list, cfg] = await Promise.all([adminApi.backups(), adminApi.settings()]);
    backups.value = [...list].sort((a, b) => b.createdAt.localeCompare(a.createdAt));
    autoHours.value = Number((cfg as { backup?: { autoHours?: number } }).backup?.autoHours) || 0;
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  loaded.value = true;
}

/** 请求期间进度环缓慢逼近 90%，响应后补满 */
async function backupNow(): Promise<void> {
  if (running.value) return;
  running.value = true;
  progress.value = 0;
  timer = window.setInterval(() => {
    progress.value = Math.min(90, progress.value + Math.max(1, (90 - progress.value) * 0.08));
  }, 120);
  try {
    const { name } = await adminApi.createBackup();
    window.clearInterval(timer);
    progress.value = 100;
    await load();
    fresh.value = name;
    toast(t('studio.data.done'), { action: t('studio.data.download'), fn: () => void download(name) });
  } catch {
    toast(t('studio.data.failed'), { icon: 'x' });
  } finally {
    window.clearInterval(timer);
    window.setTimeout(() => (running.value = false), 400);
  }
}

async function download(name: string): Promise<void> {
  try {
    saveBlob(await adminApi.downloadBackup(name), name);
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
}

async function remove(b: BackupInfo): Promise<void> {
  const ok = await dialog.confirm({
    title: t('studio.data.deleteTitle'),
    message: t('studio.data.deleteBody', { name: b.name }),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteBackup(b.name);
    backups.value = backups.value.filter((x) => x.name !== b.name);
    toast(t('studio.deleted'), { icon: 'trash' });
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

/* ===== 自动备份 ===== */
const PRESETS = [6, 12, 24, 168];
const autoOn = computed({
  get: () => autoHours.value > 0,
  set: (v: boolean) => void saveAuto(v ? 24 : 0),
});
const freq = computed({
  get: () => (PRESETS.includes(autoHours.value) ? autoHours.value : 0),
  set: (v: number) => void saveAuto(v),
});

async function saveAuto(h: number): Promise<void> {
  const prev = autoHours.value;
  autoHours.value = Math.max(0, h);
  try {
    await adminApi.saveSettings({ backup: { autoHours: autoHours.value } });
    toast(autoHours.value ? t('studio.data.autoOn', { h: autoLabel(autoHours.value) }) : t('studio.data.autoOff'), { icon: 'clock' });
  } catch {
    autoHours.value = prev;
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

function autoLabel(h: number): string {
  if (h === 24) return t('studio.data.daily');
  if (h === 168) return t('studio.data.weekly');
  return t('studio.data.everyH', { n: h });
}

/* ===== 恢复 ===== */
const restoring = ref(false);
const restorePct = ref<number | null>(null);
/** 恢复完成、即将跳转登录：锁定页面 */
const relogin = ref(false);
const restoreInput = ref<HTMLInputElement | null>(null);

/**
 * 恢复成功：整站数据（含设置与账号）都变了，所有会话已随恢复失效。
 * 先把服务端返回的安全备份插进列表并提示，随后清掉本地登录标记、送去登录页（登录后回到数据页），
 * 期间页面锁定，避免在已失效的会话上继续操作。
 */
function afterRestore(res: { safety: string; backup?: BackupInfo }): void {
  const b = res.backup;
  if (b && !backups.value.some((x) => x.name === b.name)) {
    backups.value = [b, ...backups.value];
    fresh.value = b.name;
  }
  toast(t('studio.data.restored', { name: res.safety }), { icon: 'check' });
  relogin.value = true;
  useAuthStore().clear();
  window.setTimeout(() => window.location.assign('/admin/login?next=/admin/data'), 1600);
}

function restoreError(err: unknown): void {
  const code = (err as Error).message;
  toast(code === 'invalid_backup' ? t('studio.data.restoreBad') : t('studio.data.restoreFailed'), { icon: 'x' });
}

async function restoreFrom(b: BackupInfo): Promise<void> {
  if (restoring.value) return;
  const ok = await dialog.confirm({
    title: t('studio.data.restoreTitle', { name: b.name }),
    message: t('studio.data.restoreBody'),
    confirmText: t('studio.data.restoreConfirm'),
    danger: true,
  });
  if (!ok) return;
  restoring.value = true;
  try {
    afterRestore(await adminApi.restoreBackup(b.name));
  } catch (err) {
    restoreError(err);
    restoring.value = false;
  }
}

async function onRestoreFile(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file || restoring.value) return;
  // 先在本地读 zip 目录：不是本站备份包就直接说明，不必先弹「覆盖全部数据」的确认、也不用上传
  if ((await looksLikeBackup(file)) === false) {
    toast(t('studio.data.restoreBad'), { icon: 'x' });
    return;
  }
  const ok = await dialog.confirm({
    title: t('studio.data.restoreUploadTitle', { name: file.name }),
    message: t('studio.data.restoreBody'),
    confirmText: t('studio.data.restoreConfirm'),
    danger: true,
  });
  if (!ok) return;
  restoring.value = true;
  restorePct.value = 0;
  try {
    afterRestore(await adminApi.restoreUpload(file, (p) => { restorePct.value = p; }));
  } catch (err) {
    restoreError(err);
    restoring.value = false;
  } finally {
    restorePct.value = null;
  }
}

/* ===== 导出 ===== */
const exportMedia = ref(false);
const exporting = ref(false);
async function exportMd(): Promise<void> {
  if (exporting.value) return;
  exporting.value = true;
  try {
    const d = new Date();
    const stamp = `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}`;
    saveBlob(await adminApi.exportMarkdown(exportMedia.value), `myself-markdown-${stamp}.zip`);
    toast(t('studio.data.exported'), { icon: 'download' });
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  } finally {
    exporting.value = false;
  }
}

/* ===== 导入：选择文件夹或 zip → 过滤无关文件 → 识别预览 → 确认导入 ===== */
const importDraft = ref(true);
const importBusy = ref<'' | 'scan' | 'run'>('');
const importPct = ref(0);
const preview = ref<ImportResult | null>(null);
let pending: { file: File; path: string }[] = [];
const folderInput = ref<HTMLInputElement | null>(null);
const zipInput = ref<HTMLInputElement | null>(null);

/** 只上传用得到的：Markdown、图片、站点配置；跳过依赖、构建产物、主题与版本库 */
const SKIP_DIR = /(^|\/)(node_modules|\.git|public|_site|resources|themes|\.github|\.obsidian|\.vscode)(\/|$)/;
const KEEP_EXT = /\.(md|markdown|png|jpe?g|webp|gif|ya?ml|toml|zip)$/i;

async function onImportPick(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const files = [...(input.files ?? [])];
  input.value = '';
  pending = files
    .map((file) => ({ file, path: (file as File & { webkitRelativePath?: string }).webkitRelativePath || file.name }))
    .filter(({ path }) => !SKIP_DIR.test(path) && KEEP_EXT.test(path));
  if (!pending.length) {
    toast(t('studio.data.importNone'), { icon: 'x' });
    return;
  }
  importBusy.value = 'scan';
  importPct.value = 0;
  try {
    const res = await adminApi.importPosts(pending, { draft: importDraft.value, dry: true }, (p) => { importPct.value = p; });
    if (!res.found) toast(t('studio.data.importNone'), { icon: 'x' });
    else preview.value = res;
  } catch {
    toast(t('studio.data.importFailed'), { icon: 'x' });
  } finally {
    importBusy.value = '';
  }
}

async function runImport(): Promise<void> {
  if (!pending.length || importBusy.value) return;
  importBusy.value = 'run';
  importPct.value = 0;
  try {
    const res = await adminApi.importPosts(pending, { draft: importDraft.value, dry: false }, (p) => { importPct.value = p; });
    preview.value = null;
    pending = [];
    toast(t('studio.data.imported', { n: res.created, m: res.images }), {
      icon: 'check',
      action: t('studio.nav.posts'),
      fn: () => void router.push({ name: 'admin-posts' }),
    });
    void refreshCounts();
  } catch {
    toast(t('studio.data.importFailed'), { icon: 'x' });
  } finally {
    importBusy.value = '';
  }
}

const importable = computed(() => preview.value?.items.filter((i) => !i.reason).length ?? 0);

onMounted(load);
onBeforeUnmount(() => window.clearInterval(timer));
</script>

<template>
  <section class="studio view" :class="{ relogin }" :inert="relogin">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.data.title') }}</h1>
        <p>{{ t('studio.data.desc') }}</p>
      </div>
    </div>

    <div class="bk-hero st-rise">
      <div class="ring" :style="{ '--p': running ? progress : latest ? 100 : 0 }">
        <span v-if="running" class="mono">{{ Math.round(progress) }}%</span>
        <SIcon v-else :name="latest ? 'check' : 'archive'" :size="30" />
      </div>
      <div>
        <h2>{{ running ? t('studio.data.running') : latest ? t('studio.data.last', { when: relTime(latest.createdAt) }) : t('studio.data.none') }}</h2>
        <p v-if="latest && !running" class="mono">{{ dateTimeText(latest.createdAt) }} · {{ formatSize(latest.size) }} · {{ autoHours ? t('studio.data.autoEvery', { h: autoLabel(autoHours) }) : t('studio.data.manualOnly') }}</p>
        <p v-else-if="running">{{ t('studio.data.runningSub') }}</p>
        <p v-else>{{ t('studio.data.noneSub') }}</p>
      </div>
      <button type="button" class="st-btn p lg" :disabled="running" @click="backupNow"><SIcon name="archive" :size="18" />{{ t('studio.data.now') }}</button>
    </div>

    <div class="bk-grid">
      <section class="st-card">
        <div class="st-sec-t"><h2>{{ t('studio.data.records') }}</h2><span>{{ t('studio.data.recordsN', { n: backups.length }) }}</span></div>
        <div class="st-stats" style="--n: 3">
          <div class="st-stat"><b>{{ backups.length }}</b><small>{{ t('studio.data.sCount') }}</small></div>
          <div class="st-stat"><b>{{ sizeParts(totalSize)[0] }}<span class="u">{{ sizeParts(totalSize)[1] }}</span></b><small>{{ t('studio.data.sSize') }}</small></div>
          <div class="st-stat"><b class="auto">{{ autoHours ? autoLabel(autoHours) : t('studio.data.sAutoOff') }}</b><small>{{ t('studio.data.sAuto') }}</small></div>
        </div>
        <p v-if="loaded && !backups.length" class="empty">{{ t('studio.data.empty') }}</p>
        <div class="bk-list">
          <div v-for="(b, i) in backups" :key="b.name" class="r st-rise" :class="{ fresh: fresh === b.name }" :style="{ '--i': Math.min(i, 8) }">
            <span class="fi"><SIcon name="archive" /></span>
            <div class="nm"><b class="mono">{{ b.name }}</b><small>{{ dateTimeText(b.createdAt) }}</small></div>
            <span class="num">{{ formatSize(b.size) }}</span>
            <span class="ops">
              <button type="button" class="st-ibtn" :title="t('studio.data.download')" @click="download(b.name)"><SIcon name="download" :size="18" /></button>
              <button type="button" class="st-ibtn" :disabled="restoring" :title="t('studio.data.restoreHere')" @click="restoreFrom(b)"><SIcon name="undo" :size="18" /></button>
              <button type="button" class="st-ibtn" :title="t('studio.delete')" @click="remove(b)"><SIcon name="trash" :size="18" /></button>
            </span>
          </div>
        </div>
      </section>

      <section class="st-card">
        <div class="st-sec-t"><h2>{{ t('studio.data.auto') }}</h2></div>
        <div class="st-opt">
          <div>{{ t('studio.data.autoSwitch') }}<small>{{ t('studio.data.autoSwitchSub') }}</small></div>
          <StSwitch v-model="autoOn" :label="t('studio.data.autoSwitch')" />
        </div>
        <div class="st-opt" :class="{ dim: !autoOn }">
          <div>{{ t('studio.data.freq') }}</div>
          <StSeg
            v-model="freq"
            :options="[
              { value: 6, label: '6h' },
              { value: 12, label: '12h' },
              { value: 24, label: t('studio.data.daily') },
              { value: 168, label: t('studio.data.weekly') },
            ]"
          />
        </div>
        <div class="contents">
          <div class="st-flabel">{{ t('studio.data.contains') }}</div>
          <p><SIcon name="check" :size="16" />{{ t('studio.data.cDb') }}</p>
          <p><SIcon name="check" :size="16" />{{ t('studio.data.cUploads') }}</p>
          <p><SIcon name="check" :size="16" />{{ t('studio.data.cConfig') }}</p>
        </div>
      </section>
    </div>

    <!-- 迁移 -->
    <div class="migrate">
      <div class="st-sec-t"><h2>{{ t('studio.data.migrate') }}</h2></div>
      <div class="st-note-bar"><SIcon name="info" />{{ t('studio.data.migrateNote') }}</div>
      <div class="mg">
        <div class="mc">
          <span class="ic"><SIcon name="upload" :size="20" /></span>
          <div><b>{{ t('studio.data.restore') }}</b><small>{{ t('studio.data.restoreSub') }}</small></div>
          <div class="acts">
            <button type="button" class="st-btn g sm" :disabled="restoring" @click="restoreInput?.click()">
              <template v-if="restoring">{{ restorePct !== null && restorePct < 1 ? `${Math.round(restorePct * 100)}%` : t('studio.data.restoring') }}</template>
              <template v-else>{{ t('studio.data.restoreFrom') }}</template>
            </button>
          </div>
          <span v-if="restorePct !== null" class="bar"><i :style="{ width: `${Math.round(restorePct * 100)}%` }" /></span>
        </div>
        <div class="mc">
          <span class="ic"><SIcon name="markdown" :size="20" /></span>
          <div><b>{{ t('studio.data.export') }}</b><small>{{ t('studio.data.exportSub') }}</small></div>
          <label class="opt"><StSwitch v-model="exportMedia" :label="t('studio.data.exportMedia')" />{{ t('studio.data.exportMedia') }}</label>
          <div class="acts">
            <button type="button" class="st-btn g sm" :disabled="exporting" @click="exportMd">
              <SIcon name="download" :size="16" />{{ exporting ? t('studio.data.exporting') : t('studio.data.exportBtn') }}
            </button>
          </div>
        </div>
        <div class="mc">
          <span class="ic"><SIcon name="layers" :size="20" /></span>
          <div><b>{{ t('studio.data.import') }}</b><small>{{ t('studio.data.importSub') }}</small></div>
          <label class="opt"><StSwitch v-model="importDraft" :label="t('studio.data.importDraft')" /><span>{{ t('studio.data.importDraft') }}<em>{{ t('studio.data.importDraftSub') }}</em></span></label>
          <div class="acts">
            <button type="button" class="st-btn g sm" :disabled="!!importBusy" @click="folderInput?.click()">
              <template v-if="importBusy">{{ importPct < 1 ? `${Math.round(importPct * 100)}%` : importBusy === 'scan' ? t('studio.data.importScanning') : t('studio.data.importing') }}</template>
              <template v-else>{{ t('studio.data.importFolder') }}</template>
            </button>
            <button type="button" class="st-btn g sm" :disabled="!!importBusy" @click="zipInput?.click()">{{ t('studio.data.importZip') }}</button>
          </div>
          <span v-if="importBusy" class="bar"><i :style="{ width: `${Math.round(importPct * 100)}%` }" /></span>
        </div>
      </div>
    </div>

    <input ref="restoreInput" type="file" accept=".zip" hidden @change="onRestoreFile" />
    <input ref="folderInput" type="file" webkitdirectory multiple hidden @change="onImportPick" />
    <input ref="zipInput" type="file" accept=".zip" hidden @change="onImportPick" />

    <!-- 导入预览 -->
    <StModal :open="!!preview" wide panel-class="imp" @close="preview = null">
      <template v-if="preview">
        <h3>{{ t('studio.data.importTitle', { platform: t(`studio.data.platform.${preview.platform}`), n: importable }) }}</h3>
        <p>{{ t('studio.data.importPreviewSub') }}</p>
        <div class="imp-list">
          <div v-for="it in preview.items" :key="it.file" class="ir" :class="{ skip: it.reason }">
            <div class="tt">
              <b>{{ it.title || it.file }}</b>
              <small class="mono">{{ it.reason ? it.file : `/${it.slug}` }}</small>
            </div>
            <span v-if="it.reason" class="why">{{ t(`studio.data.importSkip.${it.reason}`) }}</span>
            <span v-else class="st-badge" :class="`st-${it.status}`"><i class="st-dot" />{{ t(`studio.data.st.${it.status}`) }}</span>
          </div>
        </div>
        <div class="ft">
          <button type="button" class="st-btn g" @click="preview = null">{{ t('studio.cancel') }}</button>
          <button type="button" class="st-btn p" :disabled="!importable || !!importBusy" @click="runImport">
            {{ importBusy === 'run' ? (importPct < 1 ? `${Math.round(importPct * 100)}%` : t('studio.data.importing')) : t('studio.data.importGo', { n: importable }) }}
          </button>
        </div>
      </template>
    </StModal>
  </section>
</template>

<style scoped lang="scss">
.relogin { opacity: 0.6; pointer-events: none; transition: opacity var(--dur); }

.view {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.bk-hero {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 24px;
  align-items: center;
  padding: 20px 24px;
  border-radius: var(--r-lg);
  background: var(--well);
  margin-bottom: 20px;

  h2 { font: 700 24px/1.3 var(--font-serif); margin: 0 0 6px; }
  p { margin: 0; font-size: 14px; color: var(--st-ink-3); }
}

.ring {
  --p: 0;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  position: relative;
  color: var(--ink);
  background: conic-gradient(var(--ink) calc(var(--p) * 1%), var(--well-2) 0);

  &::before { content: ''; position: absolute; inset: 6px; border-radius: 50%; background: var(--well); }
  > * { position: relative; }
  .mono { font-size: 15px; font-weight: 500; color: var(--st-ink); }
}

.bk-grid {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
  align-items: stretch;
  gap: 20px;
  margin-bottom: 36px;

  > .st-card { gap: 6px; }
  .st-sec-t { margin-bottom: 8px; }
  .st-stats { margin-bottom: 8px; }
  .auto { font-family: var(--font-sans); font-size: 24px; line-height: 1.25; }
}

.empty { font-size: 14px; color: var(--st-ink-3); margin: 0; }

.bk-list .r {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr) 90px auto;
  gap: 14px;
  align-items: center;
  padding: 10px 8px;
  border-bottom: 1px solid var(--line);
  border-radius: var(--r-sm);
  transition: background var(--dur-fast);

  &:hover { background: var(--well); }
  &.fresh { animation: fresh 1.8s var(--ease-out); }

  .fi { width: 40px; height: 40px; border-radius: var(--r-sm); display: grid; place-items: center; background: var(--well); color: var(--st-ink-2); }
  .nm { min-width: 0; }
  b { display: block; font-size: 14px; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  small { font-size: 13px; color: var(--st-ink-3); }
  .num { font: 600 13px var(--font-mono); color: var(--st-ink-2); text-align: right; }
  .ops { display: flex; gap: 2px; opacity: 0.55; transition: opacity var(--dur-fast); }
  &:hover .ops { opacity: 1; }
}

@keyframes fresh { 0% { background: var(--tint); } 100% { background: transparent; } }

.dim { opacity: 0.5; pointer-events: none; }

.contents {
  margin-top: auto;
  padding: 14px 16px 6px;
  border-radius: var(--r-md);
  background: var(--well);

  p { display: flex; align-items: center; gap: 8px; margin: 0 0 10px; font-size: 14px; color: var(--st-ink-2); }
  .st-ic { color: color-mix(in oklab, var(--green) 70%, var(--st-ink)); }
}

.migrate .st-note-bar { margin: 0 0 18px; }

.mg {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.mc {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px;
  border-radius: var(--r-lg);
  box-shadow: 0 0 0 1px var(--line-2);

  .ic { width: 44px; height: 44px; border-radius: var(--r-md); display: grid; place-items: center; background: var(--well); color: var(--st-ink-2); }
  b { display: block; font: 700 17px/1.35 var(--font-serif); margin-bottom: 4px; }
  small { font-size: 13.5px; color: var(--st-ink-3); line-height: 1.6; }
  .acts { display: flex; gap: 8px; margin-top: auto; }

  .opt {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 13.5px;
    color: var(--st-ink-2);
    cursor: pointer;

    em { display: block; font-style: normal; font-size: 12px; color: var(--st-ink-4); }
  }

  .bar { height: 4px; border-radius: var(--r-pill); background: var(--well-2); overflow: hidden; }
  .bar i { display: block; height: 100%; background: var(--ink); transition: width var(--dur-fast); }
}

:global(.st-modal.imp) { width: min(720px, calc(100vw - 32px)); }

.imp-list {
  max-height: min(52vh, 520px);
  overflow: auto;
  margin: 0 -8px;
  padding: 0 8px;

  .ir {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 4px;
    border-bottom: 1px solid var(--line);

    &.skip { opacity: 0.55; }
    .tt { flex: 1; min-width: 0; }
    b { display: block; font-size: 14.5px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    small { font-size: 12px; color: var(--st-ink-3); }
    .why { font-size: 12.5px; color: var(--st-ink-3); }
  }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .bk-grid { grid-template-columns: 1fr; }
  .mg { grid-template-columns: 1fr; }
}

@media (max-width: 767px) {
  .bk-hero { grid-template-columns: 56px minmax(0, 1fr); gap: 14px; padding: 16px; }
  .bk-hero h2 { font-size: 20px; }
  .bk-hero p { font-size: 12.5px; line-height: 1.6; }
  .bk-hero > .st-btn { grid-column: 1 / -1; width: 100%; }
  .ring { width: 56px; height: 56px; }
  .bk-grid .auto { font-size: 20px; }
  .bk-list .r { grid-template-columns: 36px minmax(0, 1fr); gap: 8px 10px; padding: 12px 0; }
  .bk-list .fi { width: 36px; height: 36px; }
  .bk-list .nm b { white-space: normal; overflow-wrap: anywhere; font-size: 12px; }
  .bk-list .nm small { white-space: nowrap; font-size: 12px; }
  .bk-list .r .num { grid-column: 2; text-align: left; }
  .bk-list .r .ops { grid-column: 1 / -1; justify-content: flex-end; opacity: 1; }
  .bk-grid :deep(.st-opt) { flex-wrap: wrap; gap: 10px; }
  .mc { padding: 16px; }
  .mc .acts { flex-wrap: wrap; }
  .contents { padding: 12px; }
  .imp-list .ir { flex-wrap: wrap; }
  .imp-list .ir .why { flex-basis: 100%; }
}
</style>

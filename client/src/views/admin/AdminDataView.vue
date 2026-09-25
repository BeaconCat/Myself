<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type BackupInfo } from '../../api';
import { useDialogStore } from '../../stores/dialog';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSeg from './studio/StSeg.vue';
import StSwitch from './studio/StSwitch.vue';
import { saveBlob } from './studio/state';
import { toast } from './studio/toast';
import { dateTimeText, formatSize, relTime, sizeParts } from './studio/format';

/** 数据备份：立即备份（进度环）、备份记录（下载 / 删除）、自动备份间隔；导入导出为预留界面 */
const { t } = useI18n();
const dialog = useDialogStore();

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

onMounted(load);
onBeforeUnmount(() => window.clearInterval(timer));
</script>

<template>
  <section class="studio view">
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
              <button type="button" class="st-ibtn" :title="t('studio.delete')" @click="remove(b)"><SIcon name="trash" :size="18" /></button>
            </span>
          </div>
        </div>
      </section>

      <section class="st-card">
        <div class="st-sec-t"><h2>{{ t('studio.data.auto') }}</h2></div>
        <div class="st-opt">
          <div>{{ t('studio.data.autoSwitch') }}<small>{{ t('studio.data.autoSwitchSub') }}</small></div>
          <StSwitch v-model="autoOn" />
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

    <!-- 迁移（预留） -->
    <div class="migrate">
      <div class="st-sec-t"><h2>{{ t('studio.data.migrate') }}</h2></div>
      <div class="st-note-bar"><SIcon name="info" />{{ t('studio.data.migrateNote') }}</div>
      <div class="mg">
        <div class="mc">
          <span class="ic"><SIcon name="upload" :size="20" /></span>
          <div><b>{{ t('studio.data.restore') }}</b><small>{{ t('studio.data.restoreSub') }}</small></div>
          <button type="button" class="st-btn g sm st-tip" :data-tip="t('studio.data.soonTip')" disabled>{{ t('studio.data.restoreBtn') }}</button>
        </div>
        <div class="mc">
          <span class="ic"><SIcon name="markdown" :size="20" /></span>
          <div><b>{{ t('studio.data.export') }}</b><small>{{ t('studio.data.exportSub') }}</small></div>
          <button type="button" class="st-btn g sm st-tip" :data-tip="t('studio.data.soonTip')" disabled>{{ t('studio.data.exportBtn') }}</button>
        </div>
        <div class="mc">
          <span class="ic"><SIcon name="layers" :size="20" /></span>
          <div><b>{{ t('studio.data.import') }}</b><small>{{ t('studio.data.importSub') }}</small></div>
          <button type="button" class="st-btn g sm st-tip" :data-tip="t('studio.data.soonTip')" disabled>{{ t('studio.data.importBtn') }}</button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
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
  .st-btn { align-self: flex-start; margin-top: auto; }
}

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
  .bk-grid { grid-template-columns: 1fr; }
  .mg { grid-template-columns: 1fr; }
}
</style>

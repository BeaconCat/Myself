<script setup lang="ts">
/**
 * 后台 · 素材：存储占用条 + 三列网格；上传后自动发起压缩任务，
 * 每张新图在网格顶部以模糊预览 + 进度环呈现（上传转圈 → 压缩描边 → 清晰落位），完成后灵动岛汇报节省比例。
 * 点击查看大图（FLIP），可复制链接 / 删除。
 */
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, type CompressJob, type MediaItem, type QualityItem } from '../../../api';
import { useDialogStore } from '../../../stores/dialog';
import MaPage from '../../../components/mobile-admin/MaPage.vue';
import MaIcon from '../../../components/mobile-admin/MaIcon.vue';
import MaRing from '../../../components/mobile-admin/MaRing.vue';
import MaSkeleton from '../../../components/mobile-admin/MaSkeleton.vue';
import MaViewer from '../../../components/mobile-admin/MaViewer.vue';
import MediaTile from '../../admin/studio/MediaTile.vue';
import { mediaKind } from '../../../utils/mediaKind';
import { cache, loadMedia, shell, toast } from '../../../components/mobile-admin/state';
import { formatSize } from '../../../components/mobile-admin/format';
import { copyText as writeClipboard } from '../../../utils/clipboard';

const { t } = useI18n();
const dialog = useDialogStore();

const ready = computed(() => cache.media !== null);
const quality = ref<QualityItem[] | null>(null);

function load(): Promise<unknown> {
  return Promise.allSettled([loadMedia(), adminApi.qualityScan().then((q) => (quality.value = q))]);
}
onMounted(() => {
  void load();
  consumePending();
});
watch(() => shell.bump.media, () => void loadMedia().catch(() => undefined));

function refresh(done: () => void): void {
  void load().finally(done);
}

/* ---------- 存储占用 ---------- */
const extOf = (n: string): string => n.slice(n.lastIndexOf('.') + 1).toLowerCase();
const storage = computed(() => {
  const list = cache.media ?? [];
  const total = list.reduce((a, m) => a + m.size, 0);
  const by = (exts: string[]): number => list.filter((m) => exts.includes(extOf(m.name))).reduce((a, m) => a + m.size, 0);
  const webp = by(['webp']);
  const jpg = by(['jpg', 'jpeg']);
  const png = by(['png']);
  const other = Math.max(0, total - webp - jpg - png);
  const pct = (v: number): string => `${total ? (v / total) * 100 : 0}%`;
  return {
    total,
    count: list.length,
    segs: [
      { key: 'webp', label: 'WebP', size: webp, w: pct(webp), c: 'var(--primary)' },
      { key: 'jpg', label: 'JPG', size: jpg, w: pct(jpg), c: 'var(--accent-yellow)' },
      { key: 'png', label: 'PNG', size: png, w: pct(png), c: 'var(--accent-red)' },
      { key: 'other', label: t('mobileAdmin.media.other'), size: other, w: pct(other), c: 'var(--text-3)' },
    ].filter((s) => s.size > 0),
  };
});
const heavy = computed(() => (quality.value ?? []).filter((q) => q.compressible && q.format !== 'webp'));

/* ---------- 上传 + 压缩 ---------- */
interface UpTile {
  key: number;
  preview: string;
  name: string;
  phase: 'upload' | 'compress' | 'done' | 'error';
  progress: number;
  label: string;
}
const tiles = ref<UpTile[]>([]);
let tileKey = 0;
const COMPRESSIBLE = ['png', 'jpg', 'jpeg'];

function consumePending(): void {
  if (!shell.pendingUploads.length) return;
  const files = shell.pendingUploads;
  shell.pendingUploads = [];
  void start(files);
}
watch(() => shell.pendingUploads.length, consumePending);

function onPick(e: Event): void {
  const input = e.target as HTMLInputElement;
  const files = Array.from(input.files ?? []);
  input.value = '';
  if (files.length) void start(files);
}

async function start(files: File[]): Promise<void> {
  const batch: UpTile[] = files.map((f) =>
    reactive({ key: ++tileKey, preview: f.type.startsWith('image/') ? URL.createObjectURL(f) : '', name: '', phase: 'upload', progress: 0, label: t('mobileAdmin.media.uploading') }),
  );
  tiles.value = [...batch, ...tiles.value];
  let items: MediaItem[];
  try {
    items = await adminApi.uploadMedia(files);
  } catch {
    batch.forEach((b) => {
      b.phase = 'error';
      b.label = t('mobileAdmin.media.failed');
    });
    toast(t('mobileAdmin.common.uploadFailed'), '', 'error');
    window.setTimeout(() => dropTiles(batch), 2200);
    return;
  }
  batch.forEach((b, k) => {
    const it = items[k];
    if (!it) {
      b.phase = 'error';
      b.label = t('mobileAdmin.media.skipped');
      return;
    }
    b.name = it.name;
    const compressible = COMPRESSIBLE.includes(extOf(it.name));
    b.phase = compressible ? 'compress' : 'done';
    b.progress = compressible ? 0.06 : 1;
    b.label = compressible ? t('mobileAdmin.media.compressing') : formatSize(it.size);
  });
  const names = batch.filter((b) => b.phase === 'compress').map((b) => b.name);
  if (!names.length) return finish(batch, null);
  try {
    const job = await adminApi.qualityCompress(names, 80);
    await poll(job.id, batch, items);
  } catch (err) {
    const busy = err instanceof Error && err.message === 'job_running';
    batch.forEach((b) => {
      if (b.phase === 'compress') {
        b.phase = 'done';
        b.progress = 1;
        b.label = formatSize(items.find((i) => i.name === b.name)?.size ?? 0);
      }
    });
    toast(busy ? t('mobileAdmin.media.jobBusy') : t('mobileAdmin.media.compressFailed'), '', 'info');
    finish(batch, null);
  }
}

function poll(id: string, batch: UpTile[], items: MediaItem[]): Promise<void> {
  return new Promise((resolve) => {
    const tick = async (): Promise<void> => {
      let job: CompressJob;
      try {
        job = await adminApi.qualityJob(id);
      } catch {
        window.setTimeout(tick, 600);
        return;
      }
      const results = job.results ?? [];
      for (const b of batch) {
        if (b.phase !== 'compress') continue;
        const r = results.find((x) => x.name === b.name);
        if (r) {
          b.phase = r.error ? 'error' : 'done';
          b.progress = 1;
          b.label = r.error ? t('mobileAdmin.media.failed') : formatSize(r.after ?? 0);
        } else if (job.current === b.name) {
          b.progress = Math.min(0.92, b.progress + 0.14);
        }
      }
      if (job.running) {
        window.setTimeout(tick, 260);
        return;
      }
      const before = results.reduce((a, r) => a + (r.before ?? 0), 0);
      const after = results.reduce((a, r) => a + (r.after ?? 0), 0);
      finish(batch, before ? { before, after, n: items.length } : null);
      resolve();
    };
    void tick();
  });
}

function finish(batch: UpTile[], stat: { before: number; after: number; n: number } | null): void {
  if (stat) {
    const saved = Math.max(0, Math.round((1 - stat.after / stat.before) * 100));
    window.setTimeout(() => toast(t('mobileAdmin.media.compressed', { n: stat.n }), t('mobileAdmin.media.saved', { p: saved })), 250);
  } else {
    toast(t('mobileAdmin.media.uploaded', { n: batch.length }));
  }
  window.setTimeout(async () => {
    await loadMedia().catch(() => undefined);
    adminApi.qualityScan().then((q) => (quality.value = q)).catch(() => undefined);
    dropTiles(batch);
  }, 1100);
}

function dropTiles(batch: UpTile[]): void {
  const keys = new Set(batch.map((b) => b.key));
  batch.forEach((b) => { if (b.preview) URL.revokeObjectURL(b.preview); });
  tiles.value = tiles.value.filter((x) => !keys.has(x.key));
}

/* 已上传完成但仍在 tiles 里的，网格里先不重复显示 */
const gridItems = computed(() => {
  const pending = new Set(tiles.value.map((x) => x.name).filter(Boolean));
  return (cache.media ?? []).filter((m) => !pending.has(m.name));
});
/** 看图器只放图片；视频 / 音频 / 文件点开时在新标签页播放或下载 */
const imageItems = computed(() => gridItems.value.filter((m) => mediaKind(m) === 'image'));

/* ---------- 一键压缩存量 ---------- */
const bulk = ref<{ done: number; total: number } | null>(null);
async function compressAll(): Promise<void> {
  const list = heavy.value;
  if (!list.length || bulk.value) return;
  const ok = await dialog.confirm({
    title: t('mobileAdmin.media.bulkTitle', { n: list.length }),
    message: t('mobileAdmin.media.bulkMsg'),
    confirmText: t('mobileAdmin.media.bulkGo'),
    cancelText: t('mobileAdmin.common.cancel'),
  });
  if (!ok) return;
  try {
    const job = await adminApi.qualityCompress(list.map((q) => q.name), 80);
    bulk.value = { done: 0, total: job.total };
    const tick = async (): Promise<void> => {
      const j = await adminApi.qualityJob(job.id).catch(() => null);
      if (j) bulk.value = { done: j.done, total: j.total };
      if (!j || j.running) {
        window.setTimeout(tick, 400);
        return;
      }
      const before = (j.results ?? []).reduce((a, r) => a + (r.before ?? 0), 0);
      const after = (j.results ?? []).reduce((a, r) => a + (r.after ?? 0), 0);
      toast(t('mobileAdmin.media.compressed', { n: j.total }), t('mobileAdmin.media.saved', { p: before ? Math.round((1 - after / before) * 100) : 0 }));
      bulk.value = null;
      void load();
    };
    void tick();
  } catch (err) {
    toast(err instanceof Error && err.message === 'job_running' ? t('mobileAdmin.media.jobBusy') : t('mobileAdmin.media.compressFailed'), '', 'error');
  }
}

/* ---------- 大图 ---------- */
const viewerOpen = ref(false);
const viewerIndex = ref(0);
const gridEl = ref<HTMLElement | null>(null);
function openAt(k: number): void {
  const m = gridItems.value[k];
  if (!m) return;
  if (mediaKind(m) !== 'image') {
    window.open(m.url, '_blank', 'noopener');
    return;
  }
  viewerIndex.value = imageItems.value.indexOf(m);
  viewerOpen.value = true;
}
function sourceOf(k: number): HTMLElement | null {
  const m = imageItems.value[k];
  const at = m ? gridItems.value.indexOf(m) : -1;
  return gridEl.value?.querySelector<HTMLElement>(`[data-k="${at}"] img`) ?? null;
}
async function copyLink(m: MediaItem): Promise<void> {
  const url = `${window.location.origin}${m.url}`;
  if (await writeClipboard(url)) toast(t('mobileAdmin.media.copied'), m.url);
  else toast(t('mobileAdmin.media.copyFailed'), '', 'error');
}
const viewer = ref<InstanceType<typeof MaViewer> | null>(null);
async function removeItem(m: MediaItem): Promise<void> {
  const ok = await dialog.confirm({
    title: t('mobileAdmin.media.deleteTitle'),
    message: t('mobileAdmin.media.deleteMsg'),
    confirmText: t('mobileAdmin.common.delete'),
    cancelText: t('mobileAdmin.common.cancel'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteMedia(m.name);
    viewer.value?.close();
    window.setTimeout(() => {
      cache.media = (cache.media ?? []).filter((x) => x.name !== m.name);
    }, 430);
    toast(t('mobileAdmin.content.deleted'), m.name);
  } catch {
    toast(t('mobileAdmin.common.deleteFailed'), '', 'error');
  }
}
</script>

<template>
  <MaPage :title="t('mobileAdmin.media.title')" :sub="t('mobileAdmin.media.sub')" @refresh="refresh">
    <template #right>
      <label class="icbtn tap" :aria-label="t('mobileAdmin.create.upload')">
        <MaIcon name="upload" :size="20" />
        <input type="file" multiple hidden @change="onPick" />
      </label>
    </template>

    <MaSkeleton v-if="!ready" variant="grid" :count="15" />
    <template v-else>
      <section class="store">
        <div class="top2">
          <b>{{ formatSize(storage.total) }}</b>
          <small>{{ t('mobileAdmin.media.storeSub', { n: storage.count }) }}</small>
        </div>
        <div class="bar"><i v-for="s in storage.segs" :key="s.key" :style="{ width: s.w, background: s.c }" /></div>
        <div class="leg">
          <span v-for="s in storage.segs" :key="s.key" :style="{ '--c': s.c }">{{ s.label }} {{ formatSize(s.size) }}</span>
        </div>
        <button v-if="heavy.length || bulk" class="squeeze tap" :disabled="!!bulk" @click="compressAll">
          <span class="sq-ic">
            <MaRing v-if="bulk" :value="bulk.total ? bulk.done / bulk.total : 0" :size="22" :stroke="2.4" />
            <MaIcon v-else name="wand" :size="17" />
          </span>
          <span class="sq-t">
            <template v-if="bulk">{{ t('mobileAdmin.media.bulkRunning', { done: bulk.done, total: bulk.total }) }}</template>
            <template v-else>{{ t('mobileAdmin.media.heavy', { n: heavy.length }) }}</template>
          </span>
          <MaIcon v-if="!bulk" name="chev" :size="15" />
        </button>
      </section>

      <div ref="gridEl" class="mgrid">
        <div v-for="u in tiles" :key="`u${u.key}`" class="m up" :class="u.phase">
          <img v-if="u.preview" :src="u.preview" alt="" />
          <div class="prog">
            <MaRing :value="u.progress" :indeterminate="u.phase === 'upload'" :size="34" />
            <MaIcon v-if="u.phase === 'done'" name="check" :size="16" class="ok" />
          </div>
          <small>{{ u.label }}</small>
        </div>
        <button
          v-for="(m, k) in gridItems"
          :key="m.name"
          class="m"
          :class="{ named: mediaKind(m) !== 'image' }"
          :data-k="k"
          :style="{ '--k': Math.min(k, 24) }"
          :aria-label="m.title || m.name"
          @click="openAt(k)"
        >
          <MediaTile :item="m" class="preview" />
          <span v-if="mediaKind(m) !== 'image'" class="file-meta">
            <span class="file-name" :title="m.title || m.name">{{ m.title || m.name }}</span>
            <span class="file-size">{{ formatSize(m.size) }}</span>
          </span>
          <small v-else>{{ formatSize(m.size) }}</small>
        </button>
      </div>
      <div v-if="!gridItems.length && !tiles.length" class="empty">
        <span class="ring"><MaIcon name="image" /></span>
        <b>{{ t('mobileAdmin.media.empty') }}</b>
        <label class="pill-btn tap">{{ t('mobileAdmin.create.upload') }}<input type="file" multiple hidden @change="onPick" /></label>
      </div>
    </template>

    <MaViewer
      ref="viewer"
      v-model:index="viewerIndex"
      :items="imageItems"
      :open="viewerOpen"
      :source="sourceOf"
      @close="viewerOpen = false"
      @copy="copyLink"
      @remove="removeItem"
    />
  </MaPage>
</template>

<style scoped lang="scss">
label.icbtn { cursor: pointer; }

.store {
  margin: 2px 16px 0;
  padding: 18px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);

  .top2 {
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  /* 数字优先：占用总量用大号等宽数字 */
  b {
    font-family: var(--font-mono);
    font-size: 34px;
    font-weight: 600;
    line-height: 1.1;
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.01em;
  }

  small {
    font-size: 13px;
    color: var(--text-3);
  }

  .bar {
    display: flex;
    height: 10px;
    margin-top: 12px;
    border-radius: 999px;
    overflow: hidden;
    background: var(--fill-2);
    gap: 2px;

    i {
      display: block;
      height: 100%;
      min-width: 3px;
      animation: ma-bar-in 0.8s var(--ease-out) backwards;
      transform-origin: left;
    }
  }

  .leg {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 14px;
    margin-top: 10px;
    font-size: 13px;
    color: var(--text-2);

    span::before {
      content: '';
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 5px;
      background: var(--c);
    }
  }
}

@keyframes ma-bar-in {
  from { transform: scaleX(0); }
}

.squeeze {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
  padding: 10px 12px;
  border-radius: var(--r-md);
  background: var(--fill);
  box-shadow: inset 0 0 0 0.5px var(--line);
  color: var(--text);
  font-size: 14.5px;
  text-align: left;

  .sq-ic {
    width: 28px;
    height: 28px;
    border-radius: var(--r-sm);
    display: grid;
    place-items: center;
    background: var(--solid);
    color: var(--on-solid);
    --ring-bg: color-mix(in oklab, var(--on-solid) 30%, transparent);
    --ring-fg: var(--on-solid);
  }

  .sq-t { flex: 1; }
}

.mgrid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 3px;
  margin-top: 14px;
}

.m {
  position: relative;
  min-width: 0;
  aspect-ratio: 1;
  overflow: hidden;
  background: var(--fill);
  animation: ma-m-in 0.45s var(--ease-out) backwards;
  animation-delay: calc(var(--k, 0) * 18ms);

  &.named {
    display: flex;
    flex-direction: column;

    .preview { flex: 1; min-height: 0; height: auto; }
  }

  .file-meta {
    flex: none;
    display: grid;
    gap: 2px;
    padding: 6px 7px;
    background: var(--elev);
    text-align: left;
  }

  .file-name {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    overflow: hidden;
    overflow-wrap: anywhere;
    font-size: 12px;
    line-height: 1.25;
    color: var(--text);
  }

  .file-size { font: 11px/1.2 var(--font-mono); color: var(--text-3); }

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: filter 0.6s var(--ease-out), transform 0.6s var(--ease-out);
  }

  &:active img { transform: scale(0.96); }

  small {
    position: absolute;
    left: 5px;
    bottom: 5px;
    padding: 1px 5px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 11px;
    color: rgba(255, 255, 255, 0.92);
    background: rgb(0 0 0 / 0.42);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
  }

  &.up {
    animation: ma-up-in 0.5s var(--ease-spring) backwards;

    img { filter: blur(4px) brightness(0.55); transform: scale(1.06); }

    &.done img { filter: none; transform: none; }
    &.error img { filter: grayscale(1) brightness(0.4); }
  }

  .prog {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    transition: opacity 0.4s 0.5s;

    > * { grid-area: 1 / 1; }

    .ok {
      color: #fff;
      stroke-width: 2.4;
      animation: ma-ok 0.45s var(--ease-spring);
    }
  }

  &.done .prog { opacity: 0; }
}

@keyframes ma-m-in {
  from { opacity: 0; }
}

@keyframes ma-up-in {
  from { opacity: 0; transform: scale(0.8); }
}

@keyframes ma-ok {
  from { transform: scale(0.3); opacity: 0; }
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 60px 30px;
  color: var(--text-3);

  .ring {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--fill);
  }

  b { font-size: 15px; font-weight: 500; color: var(--text-2); }

  .pill-btn {
    display: inline-grid;
    place-items: center;
    cursor: pointer;
  }
}
</style>

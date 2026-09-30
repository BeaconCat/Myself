<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { adminApi, api, type AdminPost, type MediaItem, type Note } from '../../api';
import type { SiteConfig } from '../../stores/config';
import './studio/i18n';
import SIcon from './studio/SIcon.vue';
import StSeg from './studio/StSeg.vue';
import EmptyArt from './studio/EmptyArt.vue';
import { ImagePlus } from 'lucide';
import MediaViewer from './studio/MediaViewer.vue';
import QualityPanel from './studio/QualityPanel.vue';
import { toast } from './studio/toast';
import { useAuthStore } from '../../stores/auth';
import { formatSize, sizeParts } from './studio/format';
import type { MediaRef } from './studio/types';

/** 素材：瀑布流 + 拖放上传 + 大图查看/裁切；「图片优化」标签页负责扫描与后台压缩 */
const { t } = useI18n();
const isAdmin = computed(() => useAuthStore().isAdmin);
const route = useRoute();
const router = useRouter();

type Tab = 'library' | 'optimize';
const tab = ref<Tab>(route.query.tab === 'optimize' ? 'optimize' : 'library');
const items = ref<MediaItem[]>([]);
const loaded = ref(false);
const uploading = ref(false);
const dragover = ref(false);
const viewing = ref(-1);
const compressible = ref<Set<string>>(new Set());
const compressCount = ref(0);
const refs = ref<Record<string, MediaRef[]>>({});
const stamp = ref(Date.now());
const fileInput = ref<HTMLInputElement | null>(null);

async function load(): Promise<void> {
  try {
    items.value = await adminApi.media();
    stamp.value = Date.now();
  } catch {
    toast(t('studio.loadFailed'), { icon: 'x' });
  }
  loaded.value = true;
}

/** 扫描引用：文章封面 + 随想配图 */
async function loadRefs(): Promise<void> {
  const map: Record<string, MediaRef[]> = {};
  const add = (url: string, r: MediaRef) => (map[url] ??= []).push(r);
  try {
    const posts: AdminPost[] = await adminApi.posts();
    posts.forEach((p) => p.covers.forEach((u) => add(u, { kind: 'post', id: p.id, title: p.title })));
    const notes: Note[] = [];
    for (let page = 1; page <= 10; page += 1) {
      const res = await api.notes({ page, pageSize: 50 });
      notes.push(...res.items);
      if (page * 50 >= res.total) break;
    }
    notes.forEach((n) => n.images.forEach((u) => add(u, { kind: 'note', id: n.id, title: n.contentMd.slice(0, 28) || t('studio.nav.notes') })));
  } catch { /* 引用信息缺失不影响素材浏览 */ }
  // 身份与关于页：删掉它们用着的图同样会让页面缺图
  try {
    const cfg = (await adminApi.settings()) as unknown as SiteConfig;
    const about = cfg.about;
    const identity: [string | undefined, string][] = [
      [cfg.site?.logo, t('studio.identity.logo')],
      [about?.avatar, t('studio.identity.avatar')],
      [about?.portrait?.src, t('studio.identity.portrait')],
      [about?.banner?.src, t('studio.identity.banner')],
    ];
    identity.forEach(([u, label], i) => {
      if (u) add(u, { kind: 'identity', id: i, title: t('studio.media.refIdentity', { what: label }) });
    });
    const raw = JSON.stringify(about?.modules ?? []);
    for (const u of new Set(raw.match(/\/uploads\/[A-Za-z0-9.-]+/g) ?? [])) add(u, { kind: 'about', id: 0, title: t('studio.media.refAbout') });
  } catch { /* 忽略 */ }
  refs.value = map;
}

async function loadQuality(): Promise<void> {
  try {
    const q = await adminApi.qualityScan();
    compressible.value = new Set(q.filter((i) => i.compressible).map((i) => i.name));
    compressCount.value = compressible.value.size;
  } catch { /* 忽略 */ }
}

const stats = computed(() => {
  const total = items.value.reduce((s, i) => s + i.size, 0);
  let posts = 0;
  let notes = 0;
  let postsN = 0;
  let notesN = 0;
  /** 身份与关于页用到的图 */
  let site = 0;
  for (const it of items.value) {
    const r = refs.value[it.url] ?? [];
    if (r.some((x) => x.kind === 'post')) {
      posts += it.size;
      postsN += 1;
    } else if (r.some((x) => x.kind === 'note')) {
      notes += it.size;
      notesN += 1;
    } else if (r.length) {
      site += it.size;
    }
  }
  const other = Math.max(0, total - posts - notes - site);
  const otherN = Math.max(0, items.value.length - postsN - notesN);
  const pct = (n: number) => (total ? `${(n / total) * 100}%` : '0%');
  return { total, posts, notes, site, other, pct, postsN, notesN, otherN };
});

async function upload(files: File[]): Promise<void> {
  const list = files.filter((f) => f.type.startsWith('image/'));
  if (!list.length || uploading.value) return;
  uploading.value = true;
  try {
    const res = await adminApi.uploadMedia(list);
    const dup = res.filter((m) => m.duplicate).length;
    toast(
      dup ? t('studio.media.uploadedDup', { n: res.length - dup, d: dup }) : t('studio.media.uploaded', { n: res.length }),
      { icon: 'upload' },
    );
    await load();
    void loadQuality();
  } catch {
    toast(t('studio.composer.uploadFailed'), { icon: 'x' });
  } finally {
    uploading.value = false;
  }
}

function onPick(e: Event): void {
  const input = e.target as HTMLInputElement;
  if (input.files?.length) void upload([...input.files]);
  input.value = '';
}

function onDragOver(e: DragEvent): void {
  if (tab.value !== 'library' || !e.dataTransfer?.types.includes('Files')) return;
  e.preventDefault();
  dragover.value = true;
}
function onDragLeave(e: DragEvent): void {
  if (!(e.currentTarget as HTMLElement).contains(e.relatedTarget as Node)) dragover.value = false;
}
function onDrop(e: DragEvent): void {
  dragover.value = false;
  if (!e.dataTransfer?.files.length) return;
  e.preventDefault();
  void upload([...e.dataTransfer.files]);
}

function onChanged(): void {
  void load();
  void loadQuality();
}

function openRef(r: MediaRef): void {
  viewing.value = -1;
  if (r.kind === 'post') void router.push({ name: 'admin-write-post', query: { id: String(r.id) } });
  else if (r.kind === 'note') void router.push({ name: 'admin-write-note', query: { id: String(r.id) } });
  else if (r.kind === 'identity') void router.push({ name: 'admin-identity' });
  else void router.push({ name: 'admin-about' });
}

function setTab(v: Tab): void {
  tab.value = v;
  void router.replace({ query: v === 'optimize' ? { tab: v } : {} });
}

onMounted(() => {
  void load();
  void loadRefs();
  void loadQuality();
});
</script>

<template>
  <section class="studio view" @dragover="onDragOver" @dragleave="onDragLeave" @drop="onDrop">
    <div class="st-vh">
      <div>
        <h1>{{ t('studio.media.title') }}</h1>
        <p>{{ tab === 'library' ? t('studio.media.desc') : t('studio.media.qDesc') }}</p>
      </div>
      <div class="act">
        <StSeg
          v-if="isAdmin"
          :model-value="tab"
          :options="[
            { value: 'library', label: t('studio.media.tabLibrary') },
            { value: 'optimize', label: t('studio.media.tabOptimize'), count: compressCount || undefined },
          ]"
          @update:model-value="setTab"
        />
        <button v-if="tab === 'library'" type="button" class="st-btn p" :disabled="uploading" @click="fileInput?.click()">
          <SIcon name="upload" :size="18" />{{ uploading ? t('studio.media.uploading') : t('studio.media.upload') }}
        </button>
      </div>
    </div>

    <template v-if="tab === 'library'">
      <section class="st-card overview st-rise">
        <div class="st-stats" style="--n: 5">
          <div class="st-stat"><b>{{ items.length }}</b><small>{{ t('studio.media.statCount') }}</small></div>
          <div class="st-stat"><b>{{ sizeParts(stats.total)[0] }}<span class="u">{{ sizeParts(stats.total)[1] }}</span></b><small>{{ t('studio.media.statSize') }}</small></div>
          <div class="st-stat"><b>{{ sizeParts(stats.posts)[0] }}<span class="u">{{ sizeParts(stats.posts)[1] }}</span></b><small>{{ t('studio.media.sPosts', { n: stats.postsN }) }}</small></div>
          <div class="st-stat"><b>{{ sizeParts(stats.notes)[0] }}<span class="u">{{ sizeParts(stats.notes)[1] }}</span></b><small>{{ t('studio.media.sNotes', { n: stats.notesN }) }}</small></div>
          <button type="button" class="st-stat" :disabled="!compressCount || !isAdmin" @click="setTab('optimize')">
            <b :class="{ warn: compressCount }">{{ compressCount }}</b><small>{{ t('studio.media.sZip') }}</small>
          </button>
        </div>
        <div class="usage">
          <div class="ub">
            <i :style="{ width: stats.pct(stats.posts), background: 'var(--ink)' }" />
            <i :style="{ width: stats.pct(stats.notes), background: 'color-mix(in oklab, var(--ink) 45%, var(--well-2))' }" />
            <i v-if="stats.site" :style="{ width: stats.pct(stats.site), background: 'color-mix(in oklab, var(--ink) 35%, var(--line-3))' }" />
            <i :style="{ width: stats.pct(stats.other), background: 'var(--line-3)' }" />
          </div>
          <div class="st-legend">
            <span><i class="sw" style="--c: var(--ink)" />{{ t('studio.media.lgPosts', { size: formatSize(stats.posts) }) }}</span>
            <span><i class="sw" style="--c: color-mix(in oklab, var(--ink) 45%, var(--well-2))" />{{ t('studio.media.lgNotes', { size: formatSize(stats.notes) }) }}</span>
            <span v-if="stats.site"><i class="sw" style="--c: color-mix(in oklab, var(--ink) 35%, var(--line-3))" />{{ t('studio.media.lgSite', { size: formatSize(stats.site) }) }}</span>
            <span><i class="sw" style="--c: var(--line-3)" />{{ t('studio.media.lgOther', { size: formatSize(stats.other) }) }}</span>
          </div>
        </div>
      </section>

      <div v-if="loaded && !items.length" class="st-empty">
        <EmptyArt :icon="ImagePlus" />
        <h4>{{ t('studio.media.empty') }}</h4>
        <p>{{ t('studio.media.emptySub') }}</p>
        <button type="button" class="st-btn p" @click="fileInput?.click()"><SIcon name="upload" :size="18" />{{ t('studio.media.upload') }}</button>
      </div>

      <div class="masonry">
        <button
          v-for="(it, i) in items"
          :key="it.name"
          type="button"
          class="mtile st-rise"
          :style="{ '--i': i % 8 }"
          @click="viewing = i"
        >
          <img :src="`${it.thumb}?v=${it.size}-${stamp}`" alt="" loading="lazy" />
          <span v-if="compressible.has(it.name)" class="zip">{{ t('studio.media.compressible') }}</span>
          <span v-else-if="it.crop" class="zip cut">{{ t('studio.media.croppedTag') }}</span>
          <span class="cap"><span class="nm">{{ it.name }}</span><span class="mono">{{ formatSize(it.size) }}</span></span>
        </button>
      </div>
    </template>

    <QualityPanel v-else @done="onChanged" @count="compressCount = $event" />

    <Transition name="veil">
      <div v-if="dragover" class="veil"><SIcon name="upload" :size="28" /><span>{{ t('studio.media.dropVeil') }}</span></div>
    </Transition>

    <MediaViewer v-model:index="viewing" :items="items" :refs="refs" :compressible="compressible" @changed="onChanged" @open="openRef" />
    <input ref="fileInput" type="file" accept=".png,.jpg,.jpeg,.webp,.gif" multiple hidden @change="onPick" />
  </section>
</template>

<style scoped lang="scss">
.view {
  position: relative;
  max-width: 1280px;
  min-height: 100%;
  margin: 0 auto;
  padding: 32px 48px 72px;
}

.overview { margin-bottom: 22px; gap: 16px; }

.overview .st-stat:disabled { cursor: default; }
.overview .st-stat:disabled:hover { background: none; }
.overview b.warn { color: color-mix(in oklab, var(--yellow) 55%, var(--st-ink)); }

/* 空间构成：撑满整宽的分段条 */
.usage {
  display: flex;
  flex-direction: column;
  gap: 10px;

  .ub {
    display: flex;
    height: 12px;
    border-radius: var(--r-xs);
    overflow: hidden;
    background: var(--well-2);

    i { height: 100%; transition: width var(--dur-slow) var(--ease-out); }
    i + i { box-shadow: -2px 0 0 var(--paper); }
  }
}

.masonry {
  columns: 5 180px;
  column-gap: 12px;
}

.mtile {
  position: relative;
  display: block;
  width: 100%;
  margin: 0 0 12px;
  break-inside: avoid;
  border-radius: var(--r-md);
  overflow: hidden;
  cursor: zoom-in;
  background: var(--well-2);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur);

  img { display: block; width: 100%; height: auto; min-height: 80px; }

  &:hover { transform: translateY(-3px) scale(1.01); box-shadow: var(--sh-card-hover); }

  .cap {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    padding: 26px 12px 10px;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 10px;
    font-size: 13px;
    color: #fff;
    background: linear-gradient(transparent, rgba(0, 0, 0, 0.55));
    opacity: 0;
    transform: translateY(6px);
    transition: all var(--dur) var(--ease-out);
    text-align: left;

    .nm { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mono { font-size: 12px; opacity: 0.85; flex: none; }
  }

  &:hover .cap { opacity: 1; transform: none; }

  .zip {
    position: absolute;
    right: 10px;
    top: 10px;
    height: 24px;
    padding: 0 9px;
    border-radius: var(--r-sm);
    font: 500 12px/24px var(--font-sans);
    background: rgba(255, 179, 0, 0.92);
    color: #1e1c19;

    &.cut { background: rgba(255, 255, 255, 0.88); }
  }
}

.veil {
  position: fixed;
  inset: 14px 14px 14px 276px;
  z-index: 30;
  border-radius: var(--r-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font: 500 17px var(--font-serif);
  color: var(--ink);
  background: color-mix(in oklab, var(--paper) 84%, transparent);
  border: 1.5px dashed color-mix(in oklab, var(--ink) 60%, transparent);
  backdrop-filter: blur(4px);
  pointer-events: none;
}

.veil-enter-active, .veil-leave-active { transition: opacity var(--dur-fast); }
.veil-enter-from, .veil-leave-to { opacity: 0; }

@media (max-width: 1180px) {
  .view { padding: 28px 32px 64px; }
}
</style>

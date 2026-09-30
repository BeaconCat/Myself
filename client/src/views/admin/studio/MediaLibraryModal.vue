<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Check, Link, Search } from 'lucide';
import { adminApi, type MediaFolder, type MediaItem, type MediaKindName } from '../../../api';
import Icon from '../../../components/ui/Icon.vue';
import { acceptFor, extOf, kindOfExt, mediaKind } from '../../../utils/mediaKind';
import { safeMediaSrc } from '../../../utils/embeds';
import './i18n';
import SIcon from './SIcon.vue';
import StModal from './StModal.vue';
import MediaTile from './MediaTile.vue';
import FolderTree from './FolderTree.vue';
import StSeg from './StSeg.vue';
import { ALL } from './useFolders';
import { dateText, formatSize } from './format';
import { toast } from './toast';
import { useAuthStore } from '../../../stores/auth';

/**
 * 素材库模态框：左侧文件夹 + 按类型筛选 + 搜索 + 网格 / 列表 + 就地上传进当前文件夹（带进度，走查重）+ 外链地址。
 * 只显示 accept 范围内的类型（选图片时只出现图片），文件夹计数也只数这些类型。
 * - 单选：点一项即选中并关闭；多选：点选排序号，底部「插入」确认（拼图编辑器、正文插图共用）。
 * - accept 限定可选类型（默认只看图片）。外链项以 name = '' 的伪条目返回，url 即外链地址。
 */
const props = withDefaults(defineProps<{
  open: boolean;
  title: string;
  accept?: MediaKindName[];
  multiple?: boolean;
  /** 多选上限 */
  max?: number;
  /** 允许直接填写外链地址（https） */
  external?: boolean;
}>(), { accept: () => ['image'], multiple: false, max: 50, external: true });
const emit = defineEmits<{ pick: [items: MediaItem[]]; close: [] }>();
const { t } = useI18n();

const items = ref<MediaItem[] | null>(null);
/** 协作作者不能浏览整个素材库：只能上传新素材或用外链，本次上传的会列在这里 */
const isAdmin = computed(() => useAuthStore().isAdmin);
const tab = ref<MediaKindName | 'all'>('all');
const q = ref('');
const picked = ref<MediaItem[]>([]);
const progress = ref<number | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const extUrl = ref('');

const folder = ref<string>(ALL);
const allFolders = ref<MediaFolder[]>([]);
const VIEW_KEY = 'myself.studio.libraryView';
const view = ref<'grid' | 'list'>((() => {
  try {
    return localStorage.getItem(VIEW_KEY) === 'list' ? 'list' : 'grid';
  } catch {
    return 'grid';
  }
})());
watch(view, (v) => {
  try {
    localStorage.setItem(VIEW_KEY, v);
  } catch { /* 忽略 */ }
});

/** 可选范围内的素材（类型不在 accept 里的一律不出现） */
const inScope = computed(() => (items.value ?? []).filter((m) => props.accept.includes(mediaKind(m))));
/** 文件夹计数只数可选范围内的素材；没有这类素材的空文件夹也列出，方便上传进去 */
const scopedFolders = computed<MediaFolder[]>(() => {
  const count = new Map<string, number>();
  for (const m of inScope.value) if (m.folder) count.set(m.folder, (count.get(m.folder) ?? 0) + 1);
  return allFolders.value.map((f) => ({ path: f.path, count: count.get(f.path) ?? 0 }));
});
const unfiled = computed(() => inScope.value.filter((m) => !m.folder).length);

const tabs = computed(() => (props.accept.length > 1 ? (['all', ...props.accept] as const) : []));

watch(
  () => props.open,
  async (open) => {
    if (!open) return;
    picked.value = [];
    q.value = '';
    extUrl.value = '';
    tab.value = 'all';
    folder.value = ALL;
    if (!isAdmin.value) {
      items.value = [];
      allFolders.value = [];
      return;
    }
    void adminApi.mediaFolders().then((f) => { allFolders.value = f; }).catch(() => { allFolders.value = []; });
    try {
      items.value = await adminApi.media();
    } catch {
      items.value = [];
      toast(t('studio.loadFailed'), { icon: 'x' });
    }
  },
  { immediate: true },
);

const shown = computed(() => {
  const s = q.value.trim().toLowerCase();
  return inScope.value.filter((m) => {
    const k = mediaKind(m);
    if (folder.value !== ALL && (m.folder ?? '') !== folder.value) return false;
    if (tab.value !== 'all' && k !== tab.value) return false;
    return !s || `${m.title ?? ''} ${m.name}`.toLowerCase().includes(s);
  });
});

const counts = computed(() => {
  const out: Record<string, number> = { all: 0 };
  for (const m of items.value ?? []) {
    const k = mediaKind(m);
    if (!props.accept.includes(k)) continue;
    out.all += 1;
    out[k] = (out[k] ?? 0) + 1;
  }
  return out;
});

function order(m: MediaItem): number {
  return picked.value.findIndex((x) => x.url === m.url) + 1;
}

function choose(m: MediaItem): void {
  if (!props.multiple) {
    emit('pick', [m]);
    emit('close');
    return;
  }
  const i = picked.value.findIndex((x) => x.url === m.url);
  if (i >= 0) picked.value.splice(i, 1);
  else if (picked.value.length < props.max) picked.value.push(m);
  else toast(t('studio.library.max', { n: props.max }), { icon: 'x' });
}

function confirm(): void {
  if (!picked.value.length) return;
  emit('pick', [...picked.value]);
  emit('close');
}

async function onFiles(e: Event): Promise<void> {
  const input = e.target as HTMLInputElement;
  const files = [...(input.files ?? [])];
  input.value = '';
  if (!files.length || progress.value !== null) return;
  progress.value = 0;
  try {
    // 当前在某个文件夹里：上传直接放进去
    const ups = await adminApi.uploadMedia(files, (p) => { progress.value = p; }, folder.value === ALL ? '' : folder.value);
    const dup = ups.filter((m) => m.duplicate).length;
    if (dup) toast(t('studio.picker.reused'), { icon: 'copy' });
    items.value = [...ups.filter((u) => !(items.value ?? []).some((m) => m.name === u.name)), ...(items.value ?? [])];
    const usable = ups.filter((m) => props.accept.includes(mediaKind(m)));
    if (!props.multiple && usable[0]) choose(usable[0]);
    else usable.forEach((m) => { if (!order(m)) choose(m); });
  } catch (err) {
    const code = (err as Error).message;
    toast(t(code === 'unsupported_type' || code === 'file_too_large' || code === 'image_too_large' ? `studio.library.err_${code}` : 'studio.identity.uploadFailed'), { icon: 'x' });
  } finally {
    progress.value = null;
  }
}

/** 外链：只接受 https；按扩展名判断类型，不在允许范围内按「文件」处理 */
function addExternal(): void {
  const url = extUrl.value.trim();
  if (!safeMediaSrc(url) || url.startsWith('/')) {
    toast(t('studio.library.externalBad'), { icon: 'x' });
    return;
  }
  const k = kindOfExt(extOf(url));
  const kind = props.accept.includes(k) ? k : props.accept[0];
  const item: MediaItem = {
    name: '', url, thumb: url, size: 0, hasOriginal: false, crop: null, createdAt: '', kind, ext: extOf(url),
    title: decodeURIComponent(url.split(/[?#]/)[0].split('/').pop() ?? ''),
  };
  extUrl.value = '';
  choose(item);
}
</script>

<template>
  <StModal :open="open" wide panel-class="mlib" @close="emit('close')">
    <div class="hd">
      <h3>{{ title }}</h3>
      <label class="st-field search">
        <Icon :icon="Search" :size="16" />
        <input v-model="q" :placeholder="t('studio.library.search')" />
      </label>
    </div>
    <div class="body" :class="{ solo: !isAdmin }">
    <aside v-if="isAdmin" class="side">
      <FolderTree v-model="folder" compact :folders="scopedFolders" :total="inScope.length" :unfiled="unfiled" />
    </aside>
    <div class="main">
    <div class="tabs">
      <template v-if="tabs.length">
        <button v-for="k in tabs" :key="k" type="button" class="st-chip" :class="{ on: tab === k }" @click="tab = k">
          {{ t(`studio.library.kind.${k}`) }}<span class="n">{{ counts[k] ?? 0 }}</span>
        </button>
      </template>
      <span class="sp" />
      <StSeg
        v-model="view"
        icon-only
        :options="[
          { value: 'grid', icon: 'grid', title: t('studio.posts.grid') },
          { value: 'list', icon: 'list', title: t('studio.posts.list') },
        ]"
      />
    </div>

    <div class="grid" :class="{ list: view === 'list' }">
      <button type="button" class="up" :disabled="progress !== null" @click="fileInput?.click()">
        <template v-if="progress === null">
          <SIcon name="upload" :size="20" />
          <span>{{ t('studio.library.upload') }}</span>
        </template>
        <template v-else>
          <span class="bar"><i :style="{ width: `${Math.round(progress * 100)}%` }" /></span>
          <span class="mono">{{ Math.round(progress * 100) }}%</span>
        </template>
      </button>
      <button
        v-for="m in shown"
        :key="m.name"
        type="button"
        class="it"
        :class="{ on: order(m) > 0 }"
        :title="m.title || m.name"
        @click="choose(m)"
      >
        <MediaTile :item="m" />
        <span v-if="view === 'list'" class="row-nm"><b>{{ m.title || m.name }}</b><small class="mono">{{ formatSize(m.size) }} · {{ m.folder || t('studio.folder.unfiled') }} · {{ dateText(m.createdAt) }}</small></span>
        <span v-else-if="mediaKind(m) !== 'image'" class="nm">{{ m.title || m.name }}</span>
        <span v-if="order(m)" class="tick">
          <template v-if="multiple">{{ order(m) }}</template>
          <Icon v-else :icon="Check" :size="14" />
        </span>
      </button>
    </div>
    <p v-if="items && !shown.length" class="empty">{{ t(isAdmin ? 'studio.library.empty' : 'studio.library.authorHint') }}</p>
    </div>
    </div>

    <div class="ft">
      <label v-if="external" class="st-field ext">
        <Icon :icon="Link" :size="16" />
        <input v-model="extUrl" :placeholder="t('studio.library.externalPh')" @keydown.enter.prevent="addExternal" />
        <button type="button" class="st-link" :disabled="!extUrl.trim()" @click="addExternal">{{ t('studio.library.externalAdd') }}</button>
      </label>
      <span class="sp" />
      <template v-if="multiple">
        <span class="cnt">{{ t('studio.library.picked', { n: picked.length }) }}</span>
        <button type="button" class="st-btn g" @click="emit('close')">{{ t('studio.cancel') }}</button>
        <button type="button" class="st-btn p" :disabled="!picked.length" @click="confirm">{{ t('studio.library.insert') }}</button>
      </template>
    </div>
    <input ref="fileInput" type="file" :accept="acceptFor(accept)" multiple hidden @change="onFiles" />
  </StModal>
</template>

<style scoped lang="scss">
:global(.st-modal.mlib) { width: min(980px, calc(100vw - 48px)); max-height: min(86vh, 900px); display: flex; flex-direction: column; }

.hd {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;

  h3 { margin: 0; flex: 1; }
  .search { width: 260px; }
}

.body {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 190px minmax(0, 1fr);
  gap: 16px;
}

.body.solo { grid-template-columns: minmax(0, 1fr); }

.side {
  min-height: 0;
  overflow: auto;
  padding: 8px 4px;
  border-radius: var(--r-md);
  background: var(--well);
}

.main { min-width: 0; min-height: 0; display: flex; flex-direction: column; }

.tabs { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; margin-bottom: 12px; }
.tabs .sp { flex: 1; }

/* 列表视图：一行一项，缩略图 + 名称 + 大小 / 文件夹 / 日期 */
.grid.list {
  grid-template-columns: 1fr;
  gap: 4px;

  .it, .up {
    aspect-ratio: auto;
    height: 56px;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 6px 10px 6px 6px;
    text-align: left;
    background: none;

    &:hover { transform: none; background: var(--hover); }
    :deep(.mt) { width: 44px; height: 44px; flex: none; border-radius: var(--r-xs); overflow: hidden; }
  }

  .up { justify-content: flex-start; flex-direction: row; }
  .it.on { box-shadow: none; background: var(--tint); }
  .tick { position: static; margin-left: auto; }
  .row-nm { min-width: 0; display: flex; flex-direction: column; }
  .row-nm b { font-size: 14px; font-weight: 500; color: var(--st-ink); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .row-nm small { font-size: 11.5px; color: var(--st-ink-3); }
}

@media (max-width: 760px) {
  .body { grid-template-columns: 1fr; }
  .side { max-height: 140px; }
}

.grid {
  flex: 1;
  min-height: 200px;
  overflow: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(128px, 1fr));
  gap: 10px;
  padding: 3px;
}

.it, .up {
  position: relative;
  aspect-ratio: 1;
  border: 0;
  padding: 0;
  border-radius: var(--r-md);
  overflow: hidden;
  cursor: pointer;
  background: var(--well-2);
  transition: transform var(--dur) var(--ease-spring), box-shadow var(--dur-fast);

  &:hover { transform: translateY(-2px); }
}

.it.on { box-shadow: 0 0 0 2.5px var(--ink); }

.it .nm {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 16px 8px 6px;
  font-size: 12px;
  text-align: left;
  color: var(--st-ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  background: linear-gradient(transparent, color-mix(in oklab, var(--paper) 92%, transparent) 55%);
}

.tick {
  position: absolute;
  right: 6px;
  top: 6px;
  min-width: 22px;
  height: 22px;
  padding: 0 6px;
  display: grid;
  place-items: center;
  border-radius: var(--r-pill);
  font: 600 12px/1 var(--font-mono);
  color: var(--on-solid);
  background: var(--solid);
  animation: st-pop var(--dur) var(--ease-spring);
}

.up {
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 8px;
  color: var(--st-ink-3);
  box-shadow: 0 0 0 1.5px var(--line-2) inset;
  background: none;
  font-size: 13px;

  &:hover:not(:disabled) { color: var(--ink); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }

  .bar { width: 80px; height: 4px; border-radius: var(--r-pill); background: var(--well-2); overflow: hidden; }
  .bar i { display: block; height: 100%; background: var(--ink); transition: width var(--dur-fast); }
}

.empty { margin: 12px 0 0; font-size: 13px; color: var(--st-ink-3); text-align: center; }

.ft {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--line);

  .ext { width: min(420px, 100%); }
  .ext input { min-width: 0; }
  .sp { flex: 1; }
  .cnt { font-size: 13px; color: var(--st-ink-3); }
}
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Check, Link, Search } from 'lucide';
import { adminApi, type MediaItem, type MediaKindName } from '../../../api';
import Icon from '../../../components/ui/Icon.vue';
import { acceptFor, extOf, kindOfExt, mediaKind } from '../../../utils/mediaKind';
import { safeMediaSrc } from '../../../utils/embeds';
import './i18n';
import SIcon from './SIcon.vue';
import StModal from './StModal.vue';
import MediaTile from './MediaTile.vue';
import { toast } from './toast';

/**
 * 素材库模态框：按类型筛选 + 搜索 + 就地上传（带进度，走查重）+ 外链地址。
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
const tab = ref<MediaKindName | 'all'>('all');
const q = ref('');
const picked = ref<MediaItem[]>([]);
const progress = ref<number | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const extUrl = ref('');

const tabs = computed(() => (props.accept.length > 1 ? (['all', ...props.accept] as const) : []));

watch(
  () => props.open,
  async (open) => {
    if (!open) return;
    picked.value = [];
    q.value = '';
    extUrl.value = '';
    tab.value = 'all';
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
  return (items.value ?? []).filter((m) => {
    const k = mediaKind(m);
    if (!props.accept.includes(k)) return false;
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
    const ups = await adminApi.uploadMedia(files, (p) => { progress.value = p; });
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
    <div v-if="tabs.length" class="tabs">
      <button v-for="k in tabs" :key="k" type="button" class="st-chip" :class="{ on: tab === k }" @click="tab = k">
        {{ t(`studio.library.kind.${k}`) }}<span class="n">{{ counts[k] ?? 0 }}</span>
      </button>
    </div>

    <div class="grid">
      <button type="button" class="up" :disabled="progress !== null" @click="fileInput?.click()">
        <template v-if="progress === null">
          <SIcon name="upload" :size="20" />
          <span>{{ t('studio.picker.upload') }}</span>
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
        <span v-if="mediaKind(m) !== 'image'" class="nm">{{ m.title || m.name }}</span>
        <span v-if="order(m)" class="tick">
          <template v-if="multiple">{{ order(m) }}</template>
          <Icon v-else :icon="Check" :size="14" />
        </span>
      </button>
    </div>
    <p v-if="items && !shown.length" class="empty">{{ t('studio.library.empty') }}</p>

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

.tabs { display: flex; gap: 6px; flex-wrap: wrap; margin-bottom: 12px; }

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

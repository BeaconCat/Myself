<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { adminApi, thumbOf, type MediaItem } from '../../../api';
import { useAuthStore } from '../../../stores/auth';
import { useDialogStore } from '../../../stores/dialog';
import './i18n';
import SIcon from './SIcon.vue';
import { toast } from './toast';
import type { MediaRef } from './types';
import { dateTimeText, formatSize } from './format';
import { mediaKind } from '../../../utils/mediaKind';
import { renderMediaHtml } from '../../../utils/embeds';

/**
 * 素材大图查看器 + 裁切：始终基于原图与上次裁切框。
 * 裁切框以原图像素存储，渲染时按显示比例换算，窗口缩放不失真。
 */
const props = defineProps<{
  items: MediaItem[];
  refs: Record<string, MediaRef[]>;
  compressible: Set<string>;
}>();
const index = defineModel<number>('index', { default: -1 });
const emit = defineEmits<{ changed: []; open: [ref: MediaRef] }>();

const { t } = useI18n();
/** 裁剪与删除只对站长开放（协作作者只能浏览和上传） */
const isAdmin = computed(() => useAuthStore().isAdmin);
const dialog = useDialogStore();

const item = computed(() => props.items[index.value] ?? null);
/** 非图片（视频 / 音频 / 压缩包 / 文件）：舞台直接播放或显示文件卡片，没有裁切 */
const kind = computed(() => (item.value ? mediaKind(item.value) : 'image'));
const isImage = computed(() => kind.value === 'image');
const fileHtml = computed(() => {
  const it = item.value;
  if (!it || kind.value === 'image' || kind.value === 'video') return '';
  const k = kind.value === 'audio' ? 'audio' : kind.value === 'archive' && it.ext === 'zip' ? 'archive' : 'file';
  return renderMediaHtml({ kind: k, src: it.url, title: it.title || it.name }, {
    download: t('studio.media.download'), preview: t('content.embed.preview'), open: t('content.embed.open'),
  });
});
const open = computed(() => index.value >= 0 && !!item.value);
const itemRefs = computed(() => (item.value ? props.refs[item.value.url] ?? [] : []));

const src = ref('');
const loadingImg = ref(false);
const imgEl = ref<HTMLImageElement | null>(null);
const natural = reactive({ w: 1, h: 1 });
const display = reactive({ w: 1, h: 1 });
/** 原图像素坐标 */
const box = reactive({ x: 0, y: 0, w: 0, h: 0 });
/** 当前裁切比例：null = 未进入裁切（不显示裁切框），0 = 自由比例 */
const ratio = ref<number | null>(null);
const busy = ref(false);
const active = ref(false);

const scale = computed(() => display.w / natural.w);
const boxStyle = computed(() => ({
  left: `${box.x * scale.value}px`,
  top: `${box.y * scale.value}px`,
  width: `${box.w * scale.value}px`,
  height: `${box.h * scale.value}px`,
}));
const format = computed(() => (item.value?.name.split('.').pop() ?? '').toUpperCase());
const RATIOS = [
  { v: 0, label: 'free', w: 18, h: 13 },
  { v: 1, label: '1:1', w: 14, h: 14 },
  { v: 4 / 3, label: '4:3', w: 18, h: 13.5 },
  { v: 16 / 9, label: '16:9', w: 20, h: 11.25 },
];

let blobUrl = '';
async function loadOriginal(): Promise<void> {
  const it = item.value;
  if (!it) return;
  if (mediaKind(it) !== 'image') {
    loadingImg.value = false;
    ratio.value = null;
    return;
  }
  loadingImg.value = true;
  ratio.value = null;
  try {
    const blob = await adminApi.mediaOriginal(it.name);
    if (item.value?.name !== it.name) return;
    if (blobUrl) URL.revokeObjectURL(blobUrl);
    blobUrl = URL.createObjectURL(blob);
    src.value = blobUrl;
  } catch {
    src.value = it.url;
  }
}

function onImgLoad(): void {
  const el = imgEl.value;
  const it = item.value;
  if (!el || !it) return;
  natural.w = el.naturalWidth || 1;
  natural.h = el.naturalHeight || 1;
  measure();
  if (it.crop) Object.assign(box, { x: it.crop.left, y: it.crop.top, w: it.crop.width, h: it.crop.height });
  else Object.assign(box, { x: natural.w * 0.08, y: natural.h * 0.08, w: natural.w * 0.84, h: natural.h * 0.84 });
  loadingImg.value = false;
}

function measure(): void {
  if (!imgEl.value) return;
  display.w = imgEl.value.clientWidth || 1;
  display.h = imgEl.value.clientHeight || 1;
}

watch(item, (v, old) => {
  if (v && v.name !== old?.name) void loadOriginal();
});
watch(open, (v) => {
  if (v) {
    document.addEventListener('keydown', onKey);
    window.addEventListener('resize', measure);
  } else {
    document.removeEventListener('keydown', onKey);
    window.removeEventListener('resize', measure);
  }
});

function close(): void {
  index.value = -1;
}

function step(d: number): void {
  const n = props.items.length;
  if (!n) return;
  index.value = (index.value + d + n) % n;
}

function onKey(e: KeyboardEvent): void {
  if (document.querySelector('.modal-mask')) return;
  if (e.key === 'Escape') close();
  if (e.key === 'ArrowRight') step(1);
  if (e.key === 'ArrowLeft') step(-1);
}

/* ===== 裁切框拖拽 ===== */
type Kind = 'move' | 'nw' | 'ne' | 'sw' | 'se';
let drag: { kind: Kind; sx: number; sy: number; snap: { x: number; y: number; w: number; h: number } } | null = null;
const MIN = 24;

function startDrag(e: PointerEvent, kind: Kind): void {
  e.preventDefault();
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  drag = { kind, sx: e.clientX, sy: e.clientY, snap: { ...box } };
  active.value = true;
}

function onDrag(e: PointerEvent): void {
  if (!drag) return;
  const k = 1 / scale.value;
  const dx = (e.clientX - drag.sx) * k;
  const dy = (e.clientY - drag.sy) * k;
  const s = drag.snap;
  const W = natural.w;
  const H = natural.h;
  const min = MIN * k;
  const clamp = (v: number, a: number, b: number) => Math.min(b, Math.max(a, v));
  if (drag.kind === 'move') {
    box.x = clamp(s.x + dx, 0, W - s.w);
    box.y = clamp(s.y + dy, 0, H - s.h);
    return;
  }
  let { x, y, w, h } = s;
  if (drag.kind.includes('w')) {
    x = clamp(s.x + dx, 0, s.x + s.w - min);
    w = s.x + s.w - x;
  }
  if (drag.kind.includes('e')) w = clamp(s.w + dx, min, W - s.x);
  if (drag.kind.includes('n')) {
    y = clamp(s.y + dy, 0, s.y + s.h - min);
    h = s.y + s.h - y;
  }
  if (drag.kind.includes('s')) h = clamp(s.h + dy, min, H - s.y);
  if (ratio.value) {
    let nh = w / ratio.value;
    const maxH = drag.kind.includes('n') ? s.y + s.h : H - y;
    if (nh > maxH) {
      nh = maxH;
      const nw = nh * ratio.value;
      if (drag.kind.includes('w')) x = s.x + s.w - nw;
      w = nw;
    }
    if (drag.kind.includes('n')) y = s.y + s.h - nh;
    h = nh;
  }
  Object.assign(box, { x, y, w, h });
}

function endDrag(): void {
  drag = null;
  active.value = false;
}

const animating = ref(false);
/** 点击比例：进入裁切并套用该比例；再点已选中的比例则退出裁切（隐藏裁切框） */
function setRatio(r: number): void {
  if (ratio.value === r) {
    ratio.value = null;
    return;
  }
  ratio.value = r;
  if (!r) return;
  let w = natural.w * 0.84;
  let h = w / r;
  if (h > natural.h * 0.84) {
    h = natural.h * 0.84;
    w = h * r;
  }
  animating.value = true;
  Object.assign(box, { x: (natural.w - w) / 2, y: (natural.h - h) / 2, w, h });
  window.setTimeout(() => (animating.value = false), 460);
}

function resetFull(): void {
  animating.value = true;
  ratio.value = 0;
  Object.assign(box, { x: 0, y: 0, w: natural.w, h: natural.h });
  window.setTimeout(() => (animating.value = false), 460);
}

async function applyCrop(): Promise<void> {
  const it = item.value;
  if (!it || busy.value) return;
  busy.value = true;
  try {
    await adminApi.cropMedia(it.name, {
      left: Math.round(box.x),
      top: Math.round(box.y),
      width: Math.round(box.w),
      height: Math.round(box.h),
    });
    toast(t('studio.media.cropped'), { icon: 'crop' });
    emit('changed');
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

async function remove(): Promise<void> {
  const it = item.value;
  if (!it) return;
  const n = itemRefs.value.length;
  const ok = await dialog.confirm({
    title: t('studio.media.deleteTitle'),
    message: n ? t('studio.media.deleteBodyRef', { n }) : t('studio.media.deleteBody'),
    confirmText: t('studio.delete'),
    danger: true,
  });
  if (!ok) return;
  try {
    await adminApi.deleteMedia(it.name);
    toast(t('studio.deleted'), { icon: 'trash' });
    close();
    emit('changed');
  } catch {
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

/* ---------- 回退压缩：恢复压缩前的原图（可能改回原文件名），交给父级刷新列表 ---------- */
async function revert(): Promise<void> {
  const it = item.value;
  if (!it?.compressed || busy.value) return;
  const ok = await dialog.confirm({
    title: t('studio.media.revertTitle'),
    message: t('studio.media.revertBody', { size: formatSize(it.compressed.before) }),
    confirmText: t('studio.media.revert'),
  });
  if (!ok) return;
  busy.value = true;
  try {
    const res = await adminApi.revertMedia([it.name]);
    if (res.failed) throw new Error('revert');
    toast(t('studio.media.reverted', { n: 1 }), { icon: 'check' });
    close();
    emit('changed');
  } catch {
    toast(t('studio.media.revertFailed', { n: 1 }), { icon: 'x' });
  } finally {
    busy.value = false;
  }
}

/* ---------- 重命名（只改显示名） ---------- */
const titleDraft = ref('');
watch(item, (it) => { titleDraft.value = it ? it.title || it.name : ''; }, { immediate: true });

function blurTarget(e: KeyboardEvent): void {
  (e.target as HTMLInputElement).blur();
}

function cancelTitle(e: KeyboardEvent): void {
  titleDraft.value = item.value ? item.value.title || item.value.name : '';
  (e.target as HTMLInputElement).blur();
}

async function saveTitle(): Promise<void> {
  const it = item.value;
  if (!it) return;
  const next = titleDraft.value.trim();
  if (next === (it.title || it.name)) return;
  try {
    // 改回文件名或清空：清掉显示名
    const res = await adminApi.renameMedia(it.name, next === it.name ? '' : next);
    it.title = res.title;
    titleDraft.value = res.title || res.name;
    toast(t('studio.media.renamed'), { icon: 'check' });
  } catch {
    titleDraft.value = it.title || it.name;
    toast(t('studio.saveFailed'), { icon: 'x' });
  }
}

onBeforeUnmount(() => {
  if (blobUrl) URL.revokeObjectURL(blobUrl);
  document.removeEventListener('keydown', onKey);
  window.removeEventListener('resize', measure);
});
</script>

<template>
  <Teleport to="body">
    <Transition name="viewer">
      <div v-if="open && item" class="studio viewer">
        <div class="v-stage" @click.self="close">
          <div class="v-close">
            <button type="button" class="st-ibtn" :title="t('studio.media.close')" @click="close"><SIcon name="x" /></button>
            <span class="mono">{{ index + 1 }} / {{ items.length }}</span>
          </div>
          <button type="button" class="nav prev" :title="t('studio.media.prev')" @click="step(-1)"><SIcon name="arrowL" /></button>
          <button type="button" class="nav next" :title="t('studio.media.next')" @click="step(1)"><SIcon name="arrowR" /></button>
          <div v-if="kind === 'video'" class="v-media">
            <video :key="item.url" :src="item.url" controls playsinline preload="metadata" />
          </div>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <div v-else-if="!isImage" class="v-file" v-html="fileHtml" />
          <div v-else class="v-img" :class="{ ready: !loadingImg }">
            <img ref="imgEl" :src="src || thumbOf(item.url)" alt="" draggable="false" @load="onImgLoad" />
            <div
              v-if="!loadingImg && ratio !== null"
              class="crop"
              :class="{ act: active, anim: animating }"
              :style="boxStyle"
              @pointerdown.self="startDrag($event, 'move')"
              @pointermove="onDrag"
              @pointerup="endDrag"
              @pointercancel="endDrag"
            >
              <div class="g" />
              <i
                v-for="c in (['nw', 'ne', 'sw', 'se'] as const)"
                :key="c"
                class="h"
                :class="c"
                @pointerdown.stop="startDrag($event, c)"
                @pointermove="onDrag"
                @pointerup="endDrag"
              />
              <span class="sz mono">{{ Math.round(box.w) }} × {{ Math.round(box.h) }}</span>
            </div>
            <span v-if="loadingImg" class="spin" />
          </div>
        </div>

        <aside class="v-panel">
          <div>
            <!-- 显示名：站长点击即可改名（回车 / 失焦保存，Esc 放弃）；文件地址不变 -->
            <input
              v-if="isAdmin"
              v-model="titleDraft"
              class="title-in"
              :placeholder="item.name"
              :title="t('studio.media.renameHint')"
              maxlength="120"
              @keydown.enter.prevent="blurTarget"
              @keydown.esc.stop.prevent="cancelTitle"
              @blur="saveTitle"
            />
            <h3 v-else>{{ item.title || item.name }}</h3>
            <div class="sub">{{ t('studio.media.uploadedAt', { when: dateTimeText(item.createdAt) }) }}</div>
          </div>
          <dl class="kv">
            <dt>{{ t('studio.media.kvFile') }}</dt><dd class="mono file">{{ item.name }}</dd>
            <template v-if="isImage"><dt>{{ t('studio.media.kvSize') }}</dt><dd class="mono">{{ natural.w }} × {{ natural.h }}</dd></template>
            <dt>{{ t('studio.media.kvBytes') }}</dt>
            <dd class="mono">{{ formatSize(item.size) }}<span v-if="compressible.has(item.name)" class="zip">{{ t('studio.media.compressible') }}</span></dd>
            <dt>{{ t('studio.media.kvFormat') }}</dt><dd>{{ format }}</dd>
            <template v-if="item.compressed">
              <dt>{{ t('studio.media.kvCompress') }}</dt>
              <dd class="comp">
                <span class="mono">{{ t('studio.media.compressedFrom', { size: formatSize(item.compressed.before) }) }} → {{ formatSize(item.size) }}</span>
                <button v-if="isAdmin" type="button" class="st-link" :disabled="busy" @click="revert">{{ t('studio.media.revert') }}</button>
              </dd>
            </template>
            <template v-if="isImage">
              <dt>{{ t('studio.media.kvCrop') }}</dt>
              <dd>{{ item.crop ? t('studio.media.cropYes', { w: item.crop.width, h: item.crop.height }) : t('studio.media.cropNo') }}</dd>
            </template>
          </dl>
          <div v-if="isAdmin && isImage">
            <div class="st-flabel"><span>{{ t('studio.media.ratio') }}</span><button type="button" class="st-link" @click="resetFull">{{ t('studio.media.full') }}</button></div>
            <div class="ratios">
              <button v-for="r in RATIOS" :key="r.label" type="button" :class="{ on: ratio === r.v }" @click="setRatio(r.v)">
                <i :style="{ width: `${r.w}px`, height: `${r.h}px`, borderStyle: r.v ? 'solid' : 'dashed' }" />
                {{ r.v ? r.label : t('studio.media.free') }}
              </button>
            </div>
          </div>
          <div class="refs">
            <div class="st-flabel">{{ t('studio.media.usedIn') }}</div>
            <button v-for="r in itemRefs" :key="`${r.kind}-${r.id}-${r.title}`" type="button" @click="emit('open', r)">
              <SIcon :name="r.kind === 'post' ? 'doc' : r.kind === 'note' ? 'feather' : r.kind === 'identity' ? 'user' : 'layers'" :size="16" /><span>{{ r.title }}</span>
            </button>
            <p v-if="!itemRefs.length">{{ t('studio.media.unused') }}</p>
          </div>
          <div class="acts">
            <button v-if="isAdmin && isImage" type="button" class="st-btn p" :disabled="busy || loadingImg || ratio === null" @click="applyCrop"><SIcon name="crop" :size="16" />{{ t('studio.media.applyCrop') }}</button>
            <a class="st-ibtn ring" :href="item.url" :download="item.title || item.name" :title="t('studio.media.download')"><SIcon name="download" /></a>
            <button v-if="isAdmin" type="button" class="st-ibtn ring" :disabled="busy" :title="t('studio.delete')" @click="remove"><SIcon name="trash" /></button>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.viewer {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  grid-template-columns: 1fr 340px;
  background: rgba(14, 13, 12, 0.84);
  backdrop-filter: blur(14px) saturate(1.1);
}

.viewer-enter-active, .viewer-leave-active { transition: opacity var(--dur) var(--ease-out); }
.viewer-enter-from, .viewer-leave-to { opacity: 0; }
.viewer-enter-active .v-img { transition: transform var(--dur-slow) var(--ease-spring), opacity var(--dur-slow); }
.viewer-enter-from .v-img { transform: scale(0.92); opacity: 0; }
.viewer-enter-active .v-panel { transition: transform var(--dur-slow) var(--ease-spring) 0.05s, opacity var(--dur-slow) 0.05s; }
.viewer-enter-from .v-panel, .viewer-leave-to .v-panel { transform: translateX(40px); opacity: 0; }

.v-stage {
  position: relative;
  display: grid;
  place-items: center;
  padding: 72px 80px;
  min-width: 0;
}

.v-close {
  position: absolute;
  left: 24px;
  top: 22px;
  color: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;

  .st-ibtn { color: rgba(255, 255, 255, 0.8); background: rgba(255, 255, 255, 0.08); }
  .st-ibtn:hover { background: rgba(255, 255, 255, 0.16); color: #fff; }
}

.nav {
  position: absolute;
  top: 50%;
  width: 40px;
  height: 40px;
  margin-top: -20px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: rgba(255, 255, 255, 0.75);
  background: rgba(255, 255, 255, 0.06);
  transition: all var(--dur-fast);

  &:hover { background: rgba(255, 255, 255, 0.16); color: #fff; }
  &.prev { left: 22px; }
  &.next { right: 22px; }
}

.v-img {
  position: relative;
  line-height: 0;
  border-radius: var(--r-xs);
  box-shadow: 0 40px 80px -30px rgba(0, 0, 0, 0.8);
  user-select: none;
  overflow: hidden;

  img {
    display: block;
    max-width: min(100%, 1100px);
    max-height: calc(100vh - 150px);
    opacity: 0.35;
    transition: opacity var(--dur);
  }

  &.ready img { opacity: 1; }
}

.spin {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 28px;
  height: 28px;
  margin: -14px 0 0 -14px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.25);
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.crop {
  position: absolute;
  border-radius: calc(var(--r-xs) / 2);
  box-shadow: 0 0 0 9999px rgba(8, 8, 10, 0.55), 0 0 0 1px rgba(255, 255, 255, 0.9);
  cursor: move;
  touch-action: none;

  &.anim { transition: all 0.45s var(--ease-spring); }

  .g {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background:
      linear-gradient(90deg, transparent calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% + 0.5px), transparent calc(33.33% + 0.5px), transparent calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% + 0.5px), transparent calc(66.66% + 0.5px)),
      linear-gradient(transparent calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% - 0.5px), rgba(255, 255, 255, 0.35) calc(33.33% + 0.5px), transparent calc(33.33% + 0.5px), transparent calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% - 0.5px), rgba(255, 255, 255, 0.35) calc(66.66% + 0.5px), transparent calc(66.66% + 0.5px));
    opacity: 0;
    transition: opacity var(--dur);
  }

  &.act .g, &:hover .g { opacity: 1; }

  .h {
    position: absolute;
    width: 18px;
    height: 18px;
    border: 3px solid #fff;
    touch-action: none;

    &.nw { left: -3px; top: -3px; border-right: 0; border-bottom: 0; cursor: nwse-resize; }
    &.ne { right: -3px; top: -3px; border-left: 0; border-bottom: 0; cursor: nesw-resize; }
    &.sw { left: -3px; bottom: -3px; border-right: 0; border-top: 0; cursor: nesw-resize; }
    &.se { right: -3px; bottom: -3px; border-left: 0; border-top: 0; cursor: nwse-resize; }
  }

  .sz {
    position: absolute;
    left: 50%;
    bottom: 10px;
    transform: translateX(-50%);
    font-size: 11px;
    line-height: 1.4;
    color: #fff;
    background: rgba(0, 0, 0, 0.6);
    padding: 3px 8px;
    border-radius: var(--r-xs);
    white-space: nowrap;
    pointer-events: none;
  }
}

.v-panel {
  background: var(--paper);
  margin: 14px 14px 14px 0;
  border-radius: var(--r-lg);
  padding: 26px 24px;
  display: flex;
  flex-direction: column;
  gap: 22px;
  overflow: auto;
  color: var(--st-ink);

  h3 { font: 600 18px/1.4 var(--font-serif); margin: 0; word-break: break-all; }

  .title-in {
    width: calc(100% + 8px);
    margin: -4px 0 -4px -8px;
    padding: 4px 8px;
    border: 0;
    border-radius: var(--r-xs);
    font: 600 18px/1.4 var(--font-serif);
    color: var(--st-ink);
    background: transparent;
    outline: none;
    transition: background var(--dur-fast), box-shadow var(--dur-fast);

    &:hover { background: var(--hover); }
    &:focus { background: var(--well); box-shadow: 0 0 0 1.5px color-mix(in oklab, var(--ink) 55%, transparent) inset; }
  }
  .sub { font-size: 12.5px; color: var(--st-ink-3); margin-top: 4px; }
}

/* 非图片的舞台：视频限高居中，其余用正文同款文件卡片 */
.v-media {
  width: min(100%, 1100px);

  video { display: block; width: 100%; max-height: 80vh; border-radius: var(--r-md); background: #000; }
}

.v-file { width: min(100%, 560px); }

.kv .comp { display: flex; align-items: baseline; gap: 10px; flex-wrap: wrap; }

.kv .file { word-break: break-all; font-size: 12px; color: var(--st-ink-3); }

.kv {
  display: grid;
  grid-template-columns: 72px 1fr;
  gap: 9px 12px;
  font-size: 13px;
  margin: 0;

  dt { color: var(--st-ink-3); }
  dd { margin: 0; color: var(--st-ink); }

  .zip {
    margin-left: 8px;
    font: 500 11px var(--font-sans);
    padding: 1px 7px;
    border-radius: var(--r-xs);
    background: color-mix(in oklab, var(--yellow) 18%, var(--paper));
    color: color-mix(in oklab, var(--yellow) 50%, var(--st-ink));
  }
}

.ratios {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;

  button {
    height: 54px;
    border-radius: var(--r-sm);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 5px;
    font-size: 11.5px;
    color: var(--st-ink-3);
    box-shadow: 0 0 0 1px var(--line-2) inset;
    transition: all var(--dur-fast);

    i { display: block; border: 1.5px solid currentColor; border-radius: calc(var(--r-xs) / 2); }

    &.on { color: var(--lift-fg); box-shadow: var(--lift-shadow); background: var(--lift); }
    &:active { transform: scale(0.95); }
  }
}

.refs {
  button {
    display: flex;
    align-items: center;
    gap: 10px;
    width: calc(100% + 20px);
    padding: 8px 10px;
    margin: 0 -10px;
    border-radius: var(--r-sm);
    font: 500 14px var(--font-serif);
    color: var(--st-ink);
    text-align: left;

    span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .st-ic { color: var(--st-ink-3); }
    &:hover { background: var(--well); }
  }

  p { font-size: 13px; color: var(--st-ink-3); margin: 0; }
}

.acts {
  display: flex;
  gap: 8px;
  margin-top: auto;

  .st-btn { flex: 1; }
}
</style>

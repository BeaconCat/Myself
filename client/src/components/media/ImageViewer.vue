<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

/**
 * 全屏图片查看器：
 * 缩放：PC 滚轮 / 移动端双指；拖动：长按图片后拖拽；
 * 退出：双击空白 / 右上角关闭键；左下角操作指南逐条滑入滑出。
 */
export interface OriginRect {
  left: number;
  top: number;
  width: number;
  height: number;
}

const props = defineProps<{
  images: string[];
  startIndex: number;
  /** 提供时：开启 FLIP——图片从该矩形飞入中央，关闭时飞回 */
  originRect?: OriginRect;
}>();
const emit = defineEmits<{ close: [] }>();

const { t } = useI18n();

const index = ref(props.startIndex);
const scale = ref(1);
const tx = ref(0);
const ty = ref(0);
const closing = ref(false);
const dragging = ref(false);
/** FLIP 飞行中：用慢过渡接管 transform */
const flying = ref(false);
const imgReady = ref(!props.originRect);
const imgEl = ref<HTMLImageElement | null>(null);

/** 由 originRect 计算相对屏幕中央展示位的位移/缩放 */
function flyTransformFromOrigin(): { tx: number; ty: number; s: number } | null {
  const rect = props.originRect;
  const el = imgEl.value;
  if (!rect || !el) return null;
  const display = el.getBoundingClientRect();
  const baseW = display.width / scale.value;
  const baseH = display.height / scale.value;
  return {
    tx: rect.left + rect.width / 2 - window.innerWidth / 2,
    ty: rect.top + rect.height / 2 - window.innerHeight / 2,
    s: Math.max(rect.width / baseW, rect.height / baseH),
  };
}

function onImgLoad(): void {
  if (!props.originRect || imgReady.value) {
    imgReady.value = true;
    return;
  }
  // 起点摆到卡片矩形上（无过渡），下一帧飞向中央
  const from = flyTransformFromOrigin();
  if (from) {
    tx.value = from.tx;
    ty.value = from.ty;
    scale.value = from.s;
  }
  imgReady.value = true;
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      flying.value = true;
      tx.value = 0;
      ty.value = 0;
      scale.value = 1;
      window.setTimeout(() => { flying.value = false; }, 520);
    });
  });
}

const isTouch = 'ontouchstart' in window;

const guide = computed(() => [
  isTouch ? t('viewer.pinch') : t('viewer.wheel'),
  t('viewer.drag'),
  t('viewer.exit'),
]);

const imgStyle = computed(() => ({
  transform: `translate(${tx.value}px, ${ty.value}px) scale(${scale.value})`,
}));

function resetTransform(): void {
  scale.value = 1;
  tx.value = 0;
  ty.value = 0;
}

function clampScale(v: number): number {
  return Math.min(8, Math.max(0.5, v));
}

/* ===== 关闭（退场动画后卸载；有 originRect 时飞回卡片） ===== */
function requestClose(): void {
  if (closing.value) return;
  closing.value = true;
  if (props.originRect && index.value === props.startIndex) {
    const to = flyTransformFromOrigin();
    if (to) {
      flying.value = true;
      tx.value = to.tx;
      ty.value = to.ty;
      scale.value = to.s;
    }
  }
  window.setTimeout(() => emit('close'), 460);
}

/* 双击/双触空白退出 */
let lastTap = 0;
function onBackdropTap(e: MouseEvent | PointerEvent): void {
  if (e.target !== e.currentTarget) return; // 只认空白区
  const now = performance.now();
  if (now - lastTap < 320) requestClose();
  lastTap = now;
}

/* ===== 滚轮缩放（以光标为锚点） ===== */
function onWheel(e: WheelEvent): void {
  e.preventDefault();
  const factor = e.deltaY < 0 ? 1.12 : 1 / 1.12;
  const next = clampScale(scale.value * factor);
  const ratio = next / scale.value;
  const cx = e.clientX - window.innerWidth / 2;
  const cy = e.clientY - window.innerHeight / 2;
  tx.value = cx - (cx - tx.value) * ratio;
  ty.value = cy - (cy - ty.value) * ratio;
  scale.value = next;
}

/* ===== 指针：即时拖动 + 双指捏合 ===== */
const pointers = new Map<number, { x: number; y: number }>();
let panReady = false;
let last = { x: 0, y: 0 };
let pinchDist = 0;

function onPointerDown(e: PointerEvent): void {
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });

  if (pointers.size === 1) {
    last = { x: e.clientX, y: e.clientY };
    // 即时拖动：按下即可平移
    panReady = true;
    dragging.value = true;
  } else if (pointers.size === 2) {
    panReady = false;
    dragging.value = false;
    const [a, b] = [...pointers.values()];
    pinchDist = Math.hypot(a.x - b.x, a.y - b.y);
  }
}

function onPointerMove(e: PointerEvent): void {
  if (!pointers.has(e.pointerId)) return;
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });

  if (pointers.size === 2) {
    // 双指捏合：以两指中心为锚点
    const [a, b] = [...pointers.values()];
    const dist = Math.hypot(a.x - b.x, a.y - b.y);
    if (pinchDist > 0) {
      const next = clampScale(scale.value * (dist / pinchDist));
      const ratio = next / scale.value;
      const cx = (a.x + b.x) / 2 - window.innerWidth / 2;
      const cy = (a.y + b.y) / 2 - window.innerHeight / 2;
      tx.value = cx - (cx - tx.value) * ratio;
      ty.value = cy - (cy - ty.value) * ratio;
      scale.value = next;
    }
    pinchDist = dist;
    return;
  }

  if (panReady) {
    tx.value += e.clientX - last.x;
    ty.value += e.clientY - last.y;
  }
  last = { x: e.clientX, y: e.clientY };
}

function onPointerUp(e: PointerEvent): void {
  pointers.delete(e.pointerId);
  if (pointers.size < 2) pinchDist = 0;
  if (pointers.size === 0) {
    panReady = false;
    dragging.value = false;
  }
}

/* ===== 多图切换 ===== */
function go(delta: number): void {
  const len = props.images.length;
  index.value = (index.value + delta + len) % len;
  resetTransform();
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') requestClose();
  if (e.key === 'ArrowLeft') go(-1);
  if (e.key === 'ArrowRight') go(1);
}

onMounted(() => {
  window.addEventListener('keydown', onKey);
  document.documentElement.style.overflow = 'hidden';
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  document.documentElement.style.overflow = '';
});
</script>

<template>
  <Teleport to="body">
    <div
      class="viewer"
      :class="{ closing }"
      @click="onBackdropTap"
      @wheel.prevent="onWheel"
    >
      <!-- 图片舞台 -->
      <div class="stage">
        <img
          :key="index"
          ref="imgEl"
          class="stage-img"
          :class="{ dragging, flying, flip: !!originRect, ready: imgReady }"
          :src="images[index]"
          alt=""
          draggable="false"
          :style="imgStyle"
          @load="onImgLoad"
          @pointerdown.prevent="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
        />
      </div>

      <!-- 关闭 -->
      <button class="close" :aria-label="t('viewer.close')" @click="requestClose">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">
          <path d="M6 6l12 12M18 6L6 18" />
        </svg>
      </button>

      <!-- 多图切换 -->
      <template v-if="images.length > 1">
        <button class="nav prev" aria-label="prev" @click.stop="go(-1)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7" /></svg>
        </button>
        <button class="nav next" aria-label="next" @click.stop="go(1)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7" /></svg>
        </button>
        <span class="counter">{{ index + 1 }} / {{ images.length }}</span>
      </template>

      <!-- 操作指南：逐条丝滑滑入/退出 -->
      <ul class="guide">
        <li
          v-for="(line, i) in guide"
          :key="line"
          :style="{ '--i': i }"
        >{{ line }}</li>
      </ul>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.viewer {
  position: fixed;
  inset: 0;
  z-index: 9500;
  background: rgba(8, 8, 12, 0.82);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  overflow: hidden;
  touch-action: none;
  animation: viewer-in 0.35s var(--ease-out) both;

  &.closing {
    animation: viewer-out 0.4s var(--ease-out) both;
  }
}

@keyframes viewer-in {
  from { opacity: 0; }
}
@keyframes viewer-out {
  to { opacity: 0; }
}

.stage {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  pointer-events: none; /* 空白点击落到 viewer 本体 */
}

.stage-img {
  max-width: 88vw;
  max-height: 86vh;
  pointer-events: auto;
  user-select: none;
  cursor: grab;
  will-change: transform;
  transition: transform 0.08s linear;
  animation: img-in 0.45s var(--ease-out) both;

  &.dragging { cursor: grabbing; transition: none; }

  /* FLIP 模式：不用缩放入场动画，加载前隐藏，飞行时慢过渡 */
  &.flip { animation: none; opacity: 0; }
  &.flip.ready { opacity: 1; }
  &.flying { transition: transform 0.5s var(--ease-out); }
}

@keyframes img-in {
  from { opacity: 0; scale: 0.9; }
}

.viewer.closing .stage-img:not(.flip) {
  animation: img-out 0.35s var(--ease-out) both;
}

@keyframes img-out {
  to { opacity: 0; scale: 0.92; }
}

/* 关闭键 */
.close {
  position: absolute;
  top: 22px;
  right: 22px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  display: grid;
  place-items: center;
  transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);

  svg { width: 20px; height: 20px; }

  &:hover { transform: scale(1.12) rotate(90deg); background: rgba(var(--primary-rgb), 0.35); }
}

/* 多图切换 */
.nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 46px;
  height: 46px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  display: grid;
  place-items: center;
  transition: transform var(--dur-fast) var(--ease-spring), background var(--dur-fast);

  svg { width: 22px; height: 22px; }

  &.prev { left: 20px; }
  &.next { right: 20px; }

  &:hover { transform: translateY(-50%) scale(1.12); background: rgba(var(--primary-rgb), 0.35); }
}

.counter {
  position: absolute;
  top: 30px;
  left: 50%;
  transform: translateX(-50%);
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.1em;
}

/* 操作指南 */
.guide {
  position: absolute;
  left: 24px;
  bottom: 24px;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;

  li {
    color: rgba(255, 255, 255, 0.78);
    font-size: 13px;
    padding: 7px 14px;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.14);
    backdrop-filter: blur(8px);
    width: fit-content;
    animation: guide-in 0.5s var(--ease-out) both;
    animation-delay: calc(0.25s + var(--i) * 0.12s);
  }
}

@keyframes guide-in {
  from { opacity: 0; transform: translateX(-28px); filter: blur(6px); }
  to { opacity: 1; transform: none; filter: blur(0); }
}

.viewer.closing .guide li {
  animation: guide-out 0.3s var(--ease-out) both;
  animation-delay: calc(var(--i) * 0.06s);
}

@keyframes guide-out {
  to { opacity: 0; transform: translateX(-28px); filter: blur(6px); }
}

@media (max-width: 768px) {
  .nav { display: none; }
  .guide { left: 14px; bottom: 18px; }
}
</style>

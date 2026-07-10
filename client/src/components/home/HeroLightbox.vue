<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

export interface OriginRect {
  left: number;
  top: number;
  width: number;
  height: number;
}

/**
 * 首页轮播专属 Lightbox：
 * 打开：卡片从原位飞向屏幕中央，圆角渐隐展开为全尺寸；
 * 切换：当前卡飞回原位，新卡从其所在位置飞入中央；
 * 退出：卡片飞回原位，圆角套回。
 * 中央态支持滚轮/双指缩放、长按拖动。
 */
const props = defineProps<{
  images: string[];
  startIndex: number;
  /** 由父组件提供：查询某张图当前卡片的屏幕矩形 */
  getRect: (index: number) => OriginRect | null;
}>();
const emit = defineEmits<{ close: [finalIndex: number]; change: [index: number] }>();

const { t } = useI18n();

const index = ref(props.startIndex);
const closing = ref(false);
const settled = ref(false);
const dragging = ref(false);

/* 当前卡与飞回中的旧卡 */
const flight = reactive({ src: props.images[props.startIndex], style: {} as Record<string, string> });
const outgoing = reactive({ src: '', style: {} as Record<string, string>, active: false });

/* 缩放/拖动（仅中央态） */
const scale = ref(1);
const tx = ref(0);
const ty = ref(0);

const isTouch = 'ontouchstart' in window;
const FLY_MS = 550;

const guide = computed(() => [
  isTouch ? t('viewer.pinch') : t('viewer.wheel'),
  t('viewer.drag'),
  t('viewer.exit'),
]);

const zoomStyle = computed(() => ({
  transform: `translate(${tx.value}px, ${ty.value}px) scale(${scale.value})`,
  /* 中央态覆盖飞行过渡，缩放拖动即时响应 */
  transition: dragging.value ? 'none' : 'transform 0.08s linear',
}));

function rectStyle(r: OriginRect): Record<string, string> {
  return {
    left: `${r.left}px`,
    top: `${r.top}px`,
    width: `${r.width}px`,
    height: `${r.height}px`,
    borderRadius: 'var(--radius-lg)',
  };
}

/** 按图片自然比例算中央目标框 */
function centerStyle(src: string): Promise<Record<string, string>> {
  return new Promise((resolve) => {
    const probe = new Image();
    probe.onload = () => {
      const maxW = window.innerWidth * 0.86;
      const maxH = window.innerHeight * 0.8;
      const ratio = Math.min(maxW / probe.naturalWidth, maxH / probe.naturalHeight);
      const w = probe.naturalWidth * ratio;
      const h = probe.naturalHeight * ratio;
      resolve({
        left: `${(window.innerWidth - w) / 2}px`,
        top: `${(window.innerHeight - h) / 2}px`,
        width: `${w}px`,
        height: `${h}px`,
        borderRadius: '0px',
      });
    };
    probe.onerror = () => resolve(rectStyle({
      left: window.innerWidth * 0.2,
      top: window.innerHeight * 0.2,
      width: window.innerWidth * 0.6,
      height: window.innerHeight * 0.6,
    }));
    probe.src = src;
  });
}

function fallbackRect(): OriginRect {
  return {
    left: window.innerWidth / 2 - 120,
    top: window.innerHeight / 2 - 90,
    width: 240,
    height: 180,
  };
}

async function flyIn(i: number): Promise<void> {
  settled.value = false;
  resetZoom(false);
  flight.src = props.images[i];
  flight.style = { ...rectStyle(props.getRect(i) ?? fallbackRect()), transition: 'none' };
  const target = await centerStyle(flight.src);
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      flight.style = { ...target, transition: `all ${FLY_MS}ms cubic-bezier(0.2, 0.8, 0.3, 1)` };
      window.setTimeout(() => { settled.value = true; }, FLY_MS);
    });
  });
}

function flyOutCurrent(toRect: OriginRect): void {
  outgoing.src = flight.src;
  outgoing.style = { ...flight.style, transition: 'none' };
  outgoing.active = true;
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      outgoing.style = {
        ...rectStyle(toRect),
        opacity: '0.9',
        transition: `all ${FLY_MS}ms cubic-bezier(0.2, 0.8, 0.3, 1)`,
      };
      window.setTimeout(() => { outgoing.active = false; }, FLY_MS);
    });
  });
}

/** 切换：旧卡飞回原位，新卡飞入中央 */
function go(delta: number): void {
  if (!settled.value || closing.value) return;
  const len = props.images.length;
  const next = (index.value + delta + len) % len;
  flyOutCurrent(props.getRect(index.value) ?? fallbackRect());
  index.value = next;
  emit('change', next);
  void flyIn(next);
}

function requestClose(): void {
  if (closing.value) return;
  closing.value = true;
  settled.value = false;
  resetZoom(true);
  flight.style = {
    ...rectStyle(props.getRect(index.value) ?? fallbackRect()),
    transition: `all ${FLY_MS}ms cubic-bezier(0.2, 0.8, 0.3, 1)`,
  };
  window.setTimeout(() => emit('close', index.value), FLY_MS - 30);
}

function resetZoom(animated: boolean): void {
  if (!animated) {
    scale.value = 1;
    tx.value = 0;
    ty.value = 0;
    return;
  }
  scale.value = 1;
  tx.value = 0;
  ty.value = 0;
}

/* 双击空白退出 */
let lastTap = 0;
function onBackdropTap(e: MouseEvent): void {
  if (e.target !== e.currentTarget) return;
  const now = performance.now();
  if (now - lastTap < 320) requestClose();
  lastTap = now;
}

function clampScale(v: number): number {
  return Math.min(8, Math.max(0.5, v));
}

function onWheel(e: WheelEvent): void {
  if (!settled.value) return;
  const factor = e.deltaY < 0 ? 1.12 : 1 / 1.12;
  const next = clampScale(scale.value * factor);
  const ratio = next / scale.value;
  const cx = e.clientX - window.innerWidth / 2;
  const cy = e.clientY - window.innerHeight / 2;
  tx.value = cx - (cx - tx.value) * ratio;
  ty.value = cy - (cy - ty.value) * ratio;
  scale.value = next;
}

/* 长按拖动 + 双指捏合 */
const pointers = new Map<number, { x: number; y: number }>();
let holdTimer = 0;
let panReady = false;
let last = { x: 0, y: 0 };
let pinchDist = 0;
const HOLD_MS = 180;

function onPointerDown(e: PointerEvent): void {
  if (!settled.value) return;
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
  pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
  if (pointers.size === 1) {
    last = { x: e.clientX, y: e.clientY };
    panReady = false;
    holdTimer = window.setTimeout(() => {
      panReady = true;
      dragging.value = true;
    }, HOLD_MS);
  } else if (pointers.size === 2) {
    window.clearTimeout(holdTimer);
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
  window.clearTimeout(holdTimer);
  if (pointers.size < 2) pinchDist = 0;
  if (pointers.size === 0) {
    panReady = false;
    dragging.value = false;
  }
}

function onKey(e: KeyboardEvent): void {
  if (e.key === 'Escape') requestClose();
  if (e.key === 'ArrowLeft') go(-1);
  if (e.key === 'ArrowRight') go(1);
}

onMounted(() => {
  window.addEventListener('keydown', onKey);
  document.documentElement.style.overflow = 'hidden';
  void flyIn(index.value);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  document.documentElement.style.overflow = '';
});
</script>

<template>
  <Teleport to="body">
    <div
      class="hlb"
      :class="{ closing }"
      @click="onBackdropTap"
      @wheel.prevent="onWheel"
    >
      <!-- 飞回中的旧卡 -->
      <div v-if="outgoing.active" class="card ghost" :style="outgoing.style">
        <img :src="outgoing.src" alt="" draggable="false" />
      </div>

      <!-- 当前卡（中央态支持缩放拖动） -->
      <div
        class="card"
        :class="{ dragging, settled }"
        :style="[flight.style, settled ? zoomStyle : {}]"
        @pointerdown.prevent="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
      >
        <img :src="flight.src" alt="" draggable="false" />
      </div>

      <!-- 关闭 -->
      <button class="close" :aria-label="t('viewer.close')" @click="requestClose">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">
          <path d="M6 6l12 12M18 6L6 18" />
        </svg>
      </button>

      <!-- 切换 -->
      <template v-if="images.length > 1">
        <button class="nav prev" aria-label="prev" @click.stop="go(-1)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7" /></svg>
        </button>
        <button class="nav next" aria-label="next" @click.stop="go(1)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7" /></svg>
        </button>
        <span class="counter">{{ index + 1 }} / {{ images.length }}</span>
      </template>

      <!-- 操作指南 -->
      <ul class="guide">
        <li v-for="(line, i) in guide" :key="line" :style="{ '--i': i }">{{ line }}</li>
      </ul>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.hlb {
  position: fixed;
  inset: 0;
  z-index: 9500;
  background: rgba(8, 8, 12, 0.82);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  overflow: hidden;
  touch-action: none;
  animation: hlb-in 0.35s var(--ease-out) both;

  &.closing { animation: hlb-out 0.45s var(--ease-out) both; }
}

@keyframes hlb-in { from { opacity: 0; } }
@keyframes hlb-out { to { opacity: 0; } }

/* 飞行卡片：位置尺寸圆角全程过渡 */
.card {
  position: absolute;
  overflow: hidden;
  will-change: left, top, width, height, transform;
  box-shadow: 0 30px 80px -20px rgba(0, 0, 0, 0.55);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    user-select: none;
    display: block;
  }

  &.settled { cursor: grab; }
  &.dragging { cursor: grabbing; }
  &.ghost { pointer-events: none; }
}

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
  transition: transform var(--dur-fast) var(--ease-out), background var(--dur-fast);

  svg { width: 20px; height: 20px; }

  &:hover { transform: scale(1.12) rotate(90deg); background: rgba(var(--primary-rgb), 0.35); }
}

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
  transition: transform var(--dur-fast) var(--ease-out), background var(--dur-fast);

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
    animation-delay: calc(0.3s + var(--i) * 0.12s);
  }
}

@keyframes guide-in {
  from { opacity: 0; transform: translateX(-28px); filter: blur(6px); }
  to { opacity: 1; transform: none; filter: blur(0); }
}

.hlb.closing .guide li {
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

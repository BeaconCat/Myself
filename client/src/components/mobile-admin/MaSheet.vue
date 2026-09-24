<script setup lang="ts">
/**
 * 可拖拽 detent 底部 sheet（iOS 近似）：
 * - 抓手 / 头部任意方向拖；正文在顶部继续下拉、或未到最高档时上推，接管为拖 sheet
 * - 抬手按「当前位置 + 速度 × 180ms」投射到最近 detent；投射低于首档 55% 或快速下甩则关闭
 * - 超过最高档橡皮筋阻尼，落位带轻微回弹
 * - stackable：从首档到最高档输出 0..1 后退量，外壳据此把后台页面缩小后退
 */
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { bindDrag, clamp, haptic, rubber } from './gesture';

type Detent = 'half' | 'full' | number;

const props = withDefaults(
  defineProps<{
    open: boolean;
    detents?: Detent[];
    initial?: number;
    stackable?: boolean;
    /** 关闭前拦截（如未保存内容确认），返回 false 回弹 */
    beforeClose?: () => boolean | Promise<boolean>;
    label?: string;
  }>(),
  { detents: () => ['half', 'full'], initial: 0, stackable: false, beforeClose: undefined, label: '' },
);
const emit = defineEmits<{
  'update:open': [v: boolean];
  stack: [value: number, animated: boolean];
  detent: [index: number];
  closed: [];
}>();

const sheetEl = ref<HTMLElement | null>(null);
const bodyEl = ref<HTMLElement | null>(null);
const grabEl = ref<HTMLElement | null>(null);
const headEl = ref<HTMLElement | null>(null);

const visible = ref(false);
const h = ref(0);
const anim = ref(false);
const bounce = ref(false);
const vh = ref(window.innerHeight);

/** 读取 env(safe-area-inset-top)（JS 无法直接取 env，用探针元素） */
function safeTop(): number {
  const probe = document.createElement('div');
  probe.style.cssText = 'position:fixed;top:0;left:0;width:0;visibility:hidden;height:env(safe-area-inset-top)';
  document.body.appendChild(probe);
  const v = probe.offsetHeight;
  probe.remove();
  return v;
}
const sat = safeTop();

function px(d: Detent): number {
  /* 全屏档顶部留出后退舞台的露边（约 52px + 安全区） */
  if (d === 'full') return vh.value - sat - 52;
  if (d === 'half') return Math.round(vh.value * 0.56);
  return Math.min(d, vh.value - 22);
}
const heights = computed(() => props.detents.map(px).sort((a, b) => a - b));
const maxH = computed(() => heights.value[heights.value.length - 1]);
const detentIndex = ref(0);

function setH(v: number, animated: boolean, springy = animated): void {
  const max = maxH.value;
  const val = v > max ? max + rubber(v - max, 50) : Math.max(0, v);
  h.value = val;
  anim.value = animated;
  bounce.value = springy;
  if (props.stackable) {
    const d0 = heights.value[0];
    const s = max > d0 ? clamp((val - d0) / (max - d0), 0, 1.05) : 0;
    emit('stack', s, animated);
  }
}

const scrimOpacity = computed(() => clamp(h.value / Math.max(1, heights.value[0]), 0, 1));

async function show(): Promise<void> {
  vh.value = window.innerHeight;
  visible.value = true;
  anim.value = false;
  h.value = 0;
  await nextTick();
  void sheetEl.value?.offsetHeight;
  requestAnimationFrame(() => snapTo(clamp(props.initial, 0, heights.value.length - 1)));
}

function snapTo(i: number): void {
  if (detentIndex.value !== i) haptic(6);
  detentIndex.value = i;
  setH(heights.value[i], true);
  emit('detent', i);
}

let hideTimer = 0;
function hide(): void {
  anim.value = true;
  bounce.value = false;
  h.value = 0;
  if (props.stackable) emit('stack', 0, true);
  window.clearTimeout(hideTimer);
  hideTimer = window.setTimeout(() => {
    if (!props.open) {
      visible.value = false;
      emit('closed');
    }
  }, 560);
}

async function requestClose(): Promise<void> {
  const ok = props.beforeClose ? await props.beforeClose() : true;
  if (ok) emit('update:open', false);
  else snapTo(detentIndex.value);
}

/** 抬手结算：速度投射到最近 detent */
function settle(vy: number): void {
  const proj = h.value - vy * 180;
  const ds = heights.value;
  if (proj < ds[0] * 0.55 || vy > 1.4) {
    void requestClose();
    return;
  }
  let best = 0;
  ds.forEach((d, i) => {
    if (Math.abs(d - proj) < Math.abs(ds[best] - proj)) best = i;
  });
  snapTo(best);
}

watch(
  () => props.open,
  (v) => {
    if (v) void show();
    else if (visible.value) hide();
  },
  { immediate: true },
);

/** 外部指定档位（如聚焦输入时展开到全屏） */
function expand(): void {
  if (detentIndex.value < heights.value.length - 1) snapTo(heights.value.length - 1);
}
defineExpose({ expand, snapTo, close: requestClose });

/* ---------- 抓手 / 头部：指针拖拽 ---------- */
const unbinds: (() => void)[] = [];
function bindHandles(): void {
  for (const el of [grabEl.value, headEl.value]) {
    if (!el) continue;
    unbinds.push(
      bindDrag(el, {
        axis: 'y',
        down: (e) => ((e.target as HTMLElement).closest('button:not(.grab-btn),input,textarea,a,label') ? null : { h0: h.value }),
        move: (c, _dx, dy) => setH(c.h0 - dy, false),
        end: (_c, _vx, vy) => settle(vy),
      }),
    );
  }
}

/* ---------- 正文：触摸接管（顶部下拉 / 未满时上推） ---------- */
let touch: { y0: number; h0: number; mode: '' | 'sheet' | 'scroll'; hist: { t: number; y: number }[] } | null = null;
function onTouchStart(e: TouchEvent): void {
  if (e.touches.length !== 1) return;
  touch = { y0: e.touches[0].clientY, h0: h.value, mode: '', hist: [{ t: performance.now(), y: e.touches[0].clientY }] };
}
function onTouchMove(e: TouchEvent): void {
  if (!touch || !bodyEl.value) return;
  const y = e.touches[0].clientY;
  const dy = y - touch.y0;
  if (!touch.mode) {
    if (Math.abs(dy) < 6) return;
    const atTop = bodyEl.value.scrollTop <= 0;
    const canGrow = touch.h0 < maxH.value - 1;
    touch.mode = (atTop && dy > 0) || (canGrow && dy < 0) ? 'sheet' : 'scroll';
  }
  if (touch.mode !== 'sheet') return;
  e.preventDefault();
  touch.hist.push({ t: performance.now(), y });
  if (touch.hist.length > 6) touch.hist.shift();
  setH(touch.h0 - dy, false);
}
function onTouchEnd(): void {
  if (touch?.mode === 'sheet') {
    const a = touch.hist[0];
    const b = touch.hist[touch.hist.length - 1];
    const stale = performance.now() - b.t > 90;
    settle(stale ? 0 : (b.y - a.y) / Math.max(16, b.t - a.t));
  }
  touch = null;
}

const headH = ref(64);
watch(visible, async (v) => {
  if (!v) return;
  await nextTick();
  headH.value = (grabEl.value?.offsetHeight ?? 22) + (headEl.value?.offsetHeight ?? 42);
  unbinds.splice(0).forEach((u) => u());
  bindHandles();
  bodyEl.value?.addEventListener('touchstart', onTouchStart, { passive: true });
  bodyEl.value?.addEventListener('touchmove', onTouchMove, { passive: false });
  bodyEl.value?.addEventListener('touchend', onTouchEnd);
  bodyEl.value?.addEventListener('touchcancel', onTouchEnd);
});

function onResize(): void {
  vh.value = window.innerHeight;
  if (visible.value && props.open) setH(heights.value[detentIndex.value], false);
}
window.addEventListener('resize', onResize);
onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize);
  unbinds.forEach((u) => u());
  window.clearTimeout(hideTimer);
});
</script>

<template>
  <Teleport to="#ma-overlay" defer>
    <div v-if="visible" class="ma-sheet-wrap">
      <div
        class="ma-scrim"
        :class="{ anim }"
        :style="{ opacity: scrimOpacity }"
        @click="requestClose"
      />
      <section
        ref="sheetEl"
        class="ma-sheet"
        :class="{ anim, bounce }"
        role="dialog"
        aria-modal="true"
        :aria-label="label"
        :style="{ height: `${maxH + 100}px`, transform: `translateY(${maxH + 100 - h - 100}px)` }"
      >
        <div ref="grabEl" class="sh-grab"><i /></div>
        <div ref="headEl" class="sh-head"><slot name="head" :close="requestClose" /></div>
        <div ref="bodyEl" class="sh-body" :style="{ maxHeight: `${Math.max(120, h - headH)}px` }">
          <slot :close="requestClose" :detent="detentIndex" />
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.ma-sheet-wrap {
  position: absolute;
  inset: 0;
  z-index: 70;
  pointer-events: none;
}

.ma-scrim {
  position: absolute;
  inset: 0;
  background: var(--scrim);
  pointer-events: auto;

  &.anim {
    transition: opacity 0.5s var(--ease-sheet);
  }
}

.ma-sheet {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -100px;
  display: flex;
  flex-direction: column;
  border-radius: 30px 30px 0 0;
  background: var(--sheet);
  box-shadow: 0 -10px 40px -10px rgba(0, 0, 0, 0.4), inset 0 0.5px 0 var(--glass-line);
  pointer-events: auto;
  will-change: transform;
  padding-bottom: 100px;

  &.anim {
    transition: transform 0.55s var(--ease-sheet);
  }

  &.anim.bounce {
    transition: transform 0.6s var(--ease-bounce);
  }
}

.sh-grab {
  height: 22px;
  display: grid;
  place-items: center;
  flex: none;
  cursor: grab;
  touch-action: none;

  i {
    width: 38px;
    height: 5px;
    border-radius: 3px;
    background: var(--line-2);
  }
}

.sh-head {
  flex: none;
  touch-action: none;
}

.sh-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
  padding-bottom: calc(var(--safe-b) + 16px);

  &::-webkit-scrollbar { display: none; }
}
</style>

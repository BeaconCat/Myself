<script setup lang="ts">
/**
 * 后台页面外壳：iOS 大标题（滚动折叠进毛玻璃导航条）+ 吸顶附加条 + 下拉刷新。
 * 下拉刷新：顶部继续下拉时接管触摸，橡皮筋位移 + 圆环按拉距描边，过阈值松手转圈，完成后收回。
 */
import { computed, onBeforeUnmount, onMounted, ref, useSlots } from 'vue';
import { useI18n } from 'vue-i18n';
import { clamp, haptic, rubber } from './gesture';

const props = withDefaults(
  defineProps<{ title: string; eyebrow?: string; sub?: string; refreshable?: boolean; bare?: boolean }>(),
  { eyebrow: '', sub: '', refreshable: true, bare: false },
);
const emit = defineEmits<{ refresh: [done: () => void] }>();
const { t } = useI18n();
const slots = useSlots();

const scrollEl = ref<HTMLElement | null>(null);
const ltEl = ref<HTMLElement | null>(null);
const y = ref(0);
const ltH = ref(80);

const scrolled = computed(() => y.value > ltH.value - 4);
const titled = computed(() => y.value > ltH.value * 0.62);

function onScroll(): void {
  y.value = scrollEl.value?.scrollTop ?? 0;
}

/* ---------- 下拉刷新 ---------- */
const THRESHOLD = 72;
const pull = ref(0);
const refreshing = ref(false);
const settling = ref(false);
let startY = 0;
let tracking: '' | 'pull' | 'skip' = '';
let armed = false;

function onTouchStart(e: TouchEvent): void {
  if (!props.refreshable || refreshing.value || e.touches.length !== 1) return;
  startY = e.touches[0].clientY;
  tracking = '';
  armed = false;
}
function onTouchMove(e: TouchEvent): void {
  if (!props.refreshable || refreshing.value || tracking === 'skip') return;
  const dy = e.touches[0].clientY - startY;
  if (!tracking) {
    if (Math.abs(dy) < 4) return;
    tracking = dy > 0 && (scrollEl.value?.scrollTop ?? 0) <= 0 ? 'pull' : 'skip';
    if (tracking === 'skip') return;
  }
  e.preventDefault();
  settling.value = false;
  pull.value = rubber(Math.max(0, dy), 110);
  const nowArmed = pull.value >= THRESHOLD;
  if (nowArmed && !armed) haptic(10);
  armed = nowArmed;
}
function onTouchEnd(): void {
  if (tracking !== 'pull') return;
  tracking = '';
  settling.value = true;
  if (armed) {
    refreshing.value = true;
    pull.value = 56;
    emit('refresh', () => {
      window.setTimeout(() => {
        refreshing.value = false;
        settling.value = true;
        pull.value = 0;
      }, 280);
    });
  } else {
    pull.value = 0;
  }
}

const ringDash = computed(() => 62.8 * (1 - clamp(pull.value / THRESHOLD, 0, 1) * 0.86));

function measure(): void {
  ltH.value = ltEl.value?.offsetHeight ?? 80;
}

onMounted(() => {
  measure();
  const el = scrollEl.value;
  if (!el) return;
  el.addEventListener('touchstart', onTouchStart, { passive: true });
  el.addEventListener('touchmove', onTouchMove, { passive: false });
  el.addEventListener('touchend', onTouchEnd);
  el.addEventListener('touchcancel', onTouchEnd);
});
onBeforeUnmount(() => {
  const el = scrollEl.value;
  el?.removeEventListener('touchmove', onTouchMove);
});

function scrollTop(): void {
  scrollEl.value?.scrollTo({ top: 0, behavior: 'smooth' });
}
defineExpose({ scrollTop, measure, scrollEl });
</script>

<template>
  <div class="ma-page" :class="{ scrolled, titled, 'has-extra': !!slots.extra }">
    <header class="nav">
      <div class="nav-bg" />
      <div class="nav-row">
        <slot name="left" />
        <b class="nav-title">{{ title }}</b>
        <div class="nav-r"><slot name="right" /></div>
      </div>
    </header>

    <div ref="scrollEl" class="scroll" @scroll.passive="onScroll">
      <div class="ptr" :class="{ spinning: refreshing, settling }" :style="{ opacity: clamp(pull / 40, 0, 1) }" aria-hidden="true">
        <svg viewBox="0 0 24 24" :style="{ transform: `rotate(${pull * 3}deg)` }">
          <circle cx="12" cy="12" r="10" :style="{ strokeDashoffset: refreshing ? 42 : ringDash }" />
        </svg>
        <span class="sr-only">{{ t('mobileAdmin.common.refresh') }}</span>
      </div>
      <div class="content" :class="{ settling }" :style="pull ? { transform: `translateY(${pull}px)` } : undefined">
        <div class="nav-spacer" />
        <div v-if="!bare" ref="ltEl" class="lt-block">
          <div v-if="eyebrow" class="lt-eyebrow">{{ eyebrow }}</div>
          <h1 class="lt">{{ title }}</h1>
          <p v-if="sub" class="lt-sub">{{ sub }}</p>
        </div>
        <div v-if="slots.extra" class="extra"><slot name="extra" /></div>
        <slot />
        <div class="tail" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ma-page {
  position: absolute;
  inset: 0;
  background: var(--bg);
  overflow: hidden;
}

.scroll {
  position: absolute;
  inset: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar { display: none; }
}

.content.settling {
  transition: transform 0.45s var(--ease-sheet);
}

.nav-spacer {
  height: calc(var(--safe-t) + var(--nav-row));
}

.tail {
  height: calc(var(--tab-h) + var(--safe-b) + 48px);
}

/* 导航：大标题折叠 */
.nav {
  position: absolute;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  pointer-events: none;

  > * { pointer-events: auto; }
}

.nav-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: 0;
  background: var(--glass-2);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  box-shadow: 0 0.5px 0 var(--line-2);
  transition: opacity var(--dur) var(--ease-out);

  .scrolled & { opacity: 1; }
  .has-extra & { box-shadow: none; }
}

.nav-row {
  position: relative;
  height: var(--nav-row);
  margin-top: var(--safe-t);
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 8px;
}

.nav-title {
  position: absolute;
  left: 80px;
  right: 80px;
  text-align: center;
  font-size: 16.5px;
  font-weight: 600;
  opacity: 0;
  transform: translateY(6px);
  transition: all var(--dur) var(--ease-out);
  pointer-events: none;

  .titled & {
    opacity: 1;
    transform: none;
  }
}

.nav-r {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.lt-block {
  padding: 2px 20px 14px;
  transition: opacity var(--dur) var(--ease-out);

  .titled & { opacity: 0.2; }
}

.lt-eyebrow {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-3);
  letter-spacing: 0.04em;
  margin-bottom: 2px;
}

.lt {
  font-family: var(--font-serif);
  font-size: 34px;
  font-weight: 700;
  line-height: 1.25;
  letter-spacing: 0.01em;
}

.lt-sub {
  margin-top: 6px;
  font-size: 14px;
  color: var(--text-2);
  line-height: 1.6;
}

/* 吸顶附加条（分段控件等）：吸住时与导航条同一块玻璃 */
.extra {
  position: sticky;
  z-index: 8;
  top: calc(var(--safe-t) + var(--nav-row));
  padding: 4px 0 10px;
  transition: background var(--dur), box-shadow var(--dur);

  .scrolled & {
    background: var(--glass-2);
    backdrop-filter: blur(24px) saturate(180%);
    -webkit-backdrop-filter: blur(24px) saturate(180%);
    box-shadow: 0 0.5px 0 var(--line-2);
  }
}

/* 下拉刷新圆环 */
.ptr {
  position: absolute;
  z-index: 5;
  left: 50%;
  top: calc(var(--safe-t) + var(--nav-row) + 12px);
  width: 28px;
  height: 28px;
  margin-left: -14px;
  pointer-events: none;

  svg {
    width: 28px;
    height: 28px;
  }

  circle {
    fill: none;
    stroke: var(--primary);
    stroke-width: 2.4;
    stroke-linecap: round;
    stroke-dasharray: 62.8;
    transition: stroke-dashoffset 0.1s linear;
  }

  &.spinning svg {
    animation: ma-spin 0.8s linear infinite;
  }

  &.settling {
    transition: opacity 0.3s;
  }
}

@keyframes ma-spin {
  to { transform: rotate(360deg); }
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
}
</style>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, useSlots, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import { rubber } from './gesture';
import { openDrawer, shell } from './shell';

/**
 * iOS 大标题页：独立滚动容器 + 顶部导航（头像开抽屉 / 右侧动作）+ 大标题折叠为居中小标题
 * + 可选吸顶附加栏（chips / segmented）+ 下拉刷新（橡皮筋 + 进度环）。
 * 滚动方向驱动底栏收缩；KeepAlive 切走再回来时恢复滚动位置；重复点当前 tab 平滑回顶。
 */
const props = withDefaults(
  defineProps<{
    title: string;
    eyebrow?: string;
    sub?: string;
    /** false：不渲染大标题块（关于页自带身份区） */
    large?: boolean;
    /** 提供则启用下拉刷新 */
    refresh?: () => Promise<unknown>;
  }>(),
  { eyebrow: '', sub: '', large: true, refresh: undefined },
);

const { t } = useI18n();
const slots = useSlots();
const config = useConfigStore();
const avatar = computed(() => config.cfg.about.avatar || '/favicon-256.png');

const scroller = ref<HTMLElement | null>(null);
const extraEl = ref<HTMLElement | null>(null);
const y = ref(0);
const stuck = ref(false);
let lastY = 0;
let saved = 0;
let active = true;

const scrolled = computed(() => y.value > 6);
const titled = computed(() => y.value > (props.large ? 36 : 120));

function onScroll(): void {
  const el = scroller.value;
  if (!el) return;
  const top = el.scrollTop;
  y.value = top;
  const dy = top - lastY;
  lastY = top;
  if (dy > 6 && top > 80) shell.barMin = true;
  else if (dy < -6 || top < 40) shell.barMin = false;
  /* 吸顶判定：粘性定位相对滚动容器内容区（已扣除导航高度的 padding），越过自身偏移即吸住 */
  if (extraEl.value) stuck.value = top > 4 && top >= extraEl.value.offsetTop - 0.5;
}

/* ---------- 下拉刷新 ---------- */
const THRESHOLD = 64;
const pull = ref(0);
const refreshing = ref(false);
const releasing = ref(false);
let tracking = false;
let pulling = false;
let x0 = 0;
let y0 = 0;

function onTouchStart(e: TouchEvent): void {
  if (!props.refresh || refreshing.value) return;
  if ((scroller.value?.scrollTop ?? 1) > 0) return;
  tracking = true;
  pulling = false;
  x0 = e.touches[0].clientX;
  y0 = e.touches[0].clientY;
}

function onTouchMove(e: TouchEvent): void {
  if (!tracking) return;
  const dx = e.touches[0].clientX - x0;
  const dy = e.touches[0].clientY - y0;
  if (!pulling) {
    if (Math.hypot(dx, dy) < 6) return;
    if (dy > 0 && Math.abs(dy) > Math.abs(dx) && (scroller.value?.scrollTop ?? 1) <= 0) {
      pulling = true;
      releasing.value = false;
    } else {
      tracking = false;
      return;
    }
  }
  if (e.cancelable) e.preventDefault();
  pull.value = Math.max(0, rubber(dy, 150));
}

async function onTouchEnd(): Promise<void> {
  if (!tracking) return;
  tracking = false;
  if (!pulling) return;
  pulling = false;
  releasing.value = true;
  if (pull.value >= THRESHOLD && props.refresh) {
    refreshing.value = true;
    pull.value = 56;
    const started = performance.now();
    try {
      await props.refresh();
    } finally {
      const wait = Math.max(0, 520 - (performance.now() - started));
      window.setTimeout(() => {
        refreshing.value = false;
        pull.value = 0;
      }, wait);
    }
  } else {
    pull.value = 0;
  }
}

const ringProgress = computed(() => Math.min(1, pull.value / THRESHOLD));

onMounted(() => {
  const el = scroller.value;
  if (!el) return;
  el.addEventListener('touchstart', onTouchStart, { passive: true });
  el.addEventListener('touchmove', onTouchMove, { passive: false });
  el.addEventListener('touchend', onTouchEnd);
  el.addEventListener('touchcancel', onTouchEnd);
});

onBeforeUnmount(() => {
  const el = scroller.value;
  if (!el) return;
  el.removeEventListener('touchstart', onTouchStart);
  el.removeEventListener('touchmove', onTouchMove);
  el.removeEventListener('touchend', onTouchEnd);
  el.removeEventListener('touchcancel', onTouchEnd);
});

/* ---------- KeepAlive：保存 / 恢复滚动 ---------- */
onDeactivated(() => {
  active = false;
  saved = scroller.value?.scrollTop ?? 0;
});
onActivated(() => {
  active = true;
  shell.barMin = false;
  if (scroller.value && saved) scroller.value.scrollTop = saved;
});

watch(
  () => shell.scrollTopSeq,
  () => {
    if (!active || !scroller.value) return;
    scroller.value.scrollTo({ top: 0, behavior: 'smooth' });
  },
);

function scrollToTop(): void {
  scroller.value?.scrollTo({ top: 0, behavior: 'smooth' });
}

defineExpose({ scroller, scrollToTop });
</script>

<template>
  <section class="lt-page" :class="{ scrolled, titled, stuck, 'has-extra': !!slots.extra }">
    <div ref="scroller" class="lt-scroll" @scroll.passive="onScroll">
      <div
        v-if="refresh"
        class="ptr"
        :class="{ spinning: refreshing, releasing }"
        :style="{ opacity: Math.min(1, pull / 36), transform: `translateY(${pull * 0.5 - 18}px) rotate(${pull * 3}deg)` }"
        aria-hidden="true"
      >
        <svg viewBox="0 0 24 24">
          <circle class="bg" cx="12" cy="12" r="9" />
          <circle class="fg" cx="12" cy="12" r="9" :style="{ strokeDashoffset: refreshing ? 40 : 56.55 * (1 - ringProgress) }" />
        </svg>
      </div>
      <div class="lt-body" :class="{ releasing }" :style="pull ? { transform: `translateY(${pull}px)` } : undefined">
        <div v-if="large" class="lt-block" :style="pull ? { transform: `scale(${1 + pull / 900})` } : undefined">
          <div v-if="eyebrow" class="lt-eyebrow">{{ eyebrow }}</div>
          <div class="lt-row">
            <h1 class="lt">{{ title }}</h1>
            <p v-if="sub" class="lt-sub">{{ sub }}</p>
          </div>
        </div>
        <slot name="hero" />
        <div v-if="slots.extra" ref="extraEl" class="lt-extra">
          <slot name="extra" />
        </div>
        <slot />
        <div class="lt-tail" />
      </div>
    </div>

    <header class="lt-nav">
      <div class="lt-nav-bg" />
      <div class="lt-nav-row">
        <button class="av-btn m-tap" :aria-label="t('mobile.openDrawer')" @click="openDrawer">
          <img class="m-avatar" :src="avatar" alt="" draggable="false" />
        </button>
        <b class="lt-nav-title">{{ title }}</b>
        <div class="lt-nav-r"><slot name="right" /></div>
      </div>
    </header>
  </section>
</template>

<style scoped lang="scss">
.lt-page {
  position: absolute;
  inset: 0;
  background: var(--bg);
  overflow: hidden;
}

.lt-scroll {
  position: absolute;
  inset: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior-y: contain;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
  padding-top: var(--m-nav-h);

  &::-webkit-scrollbar { display: none; }
}

.lt-body {
  position: relative;
  will-change: transform;

  &.releasing { transition: transform 0.5s var(--m-ease-sheet); }
}

.lt-tail { height: calc(var(--m-safe-b) + 120px); }

/* 大标题（高密度：标题与副标题同一基线行，放不下时副标题换到下一行） */
.lt-block {
  padding: 0 16px 10px;
  transform-origin: left center;
}

.lt-eyebrow {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-3);
  letter-spacing: 0.04em;
  margin-bottom: 0;
}

.lt-row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  column-gap: 12px;
}

.lt {
  flex: none;
  font-family: var(--font-serif);
  font-size: 30px;
  font-weight: 700;
  line-height: 1.3;
  letter-spacing: 0.01em;
  transition: opacity var(--dur) var(--ease-out);

  .titled & { opacity: 0; }
}

.lt-sub {
  flex: 1 1 12em;
  min-width: 0;
  font-size: 13px;
  color: var(--text-3);
  line-height: 1.55;
}

/* 吸顶附加栏 */
.lt-extra {
  position: sticky;
  z-index: 6;
  top: 0;
  padding: 4px 0 10px;
  margin-top: -2px;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    z-index: -1;
    background: var(--m-glass-2);
    backdrop-filter: blur(24px) saturate(180%);
    -webkit-backdrop-filter: blur(24px) saturate(180%);
    box-shadow: 0 0.5px 0 var(--line-2);
    opacity: 0;
    transition: opacity var(--dur) var(--ease-out);
  }

  .stuck &::before { opacity: 1; }
}

/* 顶部导航 */
.lt-nav {
  position: absolute;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  pointer-events: none;

  > * { pointer-events: auto; }
}

.lt-nav-bg {
  position: absolute;
  inset: 0;
  height: var(--m-nav-h);
  pointer-events: none;
  opacity: 0;
  background: var(--m-glass-2);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  box-shadow: 0 0.5px 0 var(--line-2);
  transition: opacity var(--dur) var(--ease-out);

  .scrolled & { opacity: 1; }
  .stuck & { box-shadow: none; }
}

.lt-nav-row {
  position: relative;
  height: var(--m-nav-row);
  margin-top: calc(var(--m-safe-t) + 6px);
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 8px;
}

.av-btn {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  flex: none;

  &::after {
    content: '';
    position: absolute;
    right: -1px;
    bottom: 1px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #2bd46a;
    box-shadow: 0 0 0 2px var(--bg);
  }
}

.lt-nav-title {
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;

  .titled & {
    opacity: 1;
    transform: none;
  }
}

.lt-nav-r {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

/* 下拉刷新指示器 */
.ptr {
  position: absolute;
  z-index: 1;
  left: 50%;
  top: calc(var(--m-nav-h) - 4px);
  width: 28px;
  height: 28px;
  margin-left: -14px;
  pointer-events: none;

  &.releasing { transition: transform 0.5s var(--m-ease-sheet), opacity 0.3s; }

  svg {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg);
  }

  circle {
    fill: none;
    stroke-width: 2.4;
  }

  .bg { stroke: var(--fill-3); }

  .fg {
    stroke: var(--ink);
    stroke-linecap: round;
    stroke-dasharray: 56.55;
  }

  &.spinning svg { animation: ptr-spin 0.8s linear infinite; }
}

@keyframes ptr-spin {
  from { transform: rotate(-90deg); }
  to { transform: rotate(270deg); }
}
</style>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { LanguagesData } from '../types';
import type { ModProps } from './props';
import ModHead from '../parts/ModHead.vue';

/**
 * 语言占比（languages）：堆叠条（段间 3px 缝）或环形图；图例与图形悬停互相高亮。
 * 配色 = 主色单色阶梯（按占比排名从实到淡），不再使用每项自带的彩色，保持低饱和。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as LanguagesData);
const { t } = useI18n();
const hover = ref<number | null>(null);

/** 合计不足 100 时按比例归一 */
const items = computed(() => {
  const sum = d.value.items.reduce((a, b) => a + Math.max(0, b.percent), 0) || 1;
  return d.value.items.map((it, i) => ({ ...it, share: (Math.max(0, it.percent) / sum) * 100, color: ladder(i) }));
});

/** 主色单色阶梯：第 i 段主色不透明度递减（叠在卡面上即为同色相明度阶梯） */
const STEPS = [88, 62, 44, 30, 20];
function ladder(i: number): string {
  return `color-mix(in oklab, var(--primary) ${STEPS[Math.min(i, STEPS.length - 1)]}%, transparent)`;
}

/** 图例条长度以最大占比为满格，便于比较 */
const maxShare = computed(() => Math.max(1, ...items.value.map((x) => x.share)));

const R = 50;
const C = 2 * Math.PI * R;
const arcs = computed(() => {
  let off = 0;
  return items.value.map((it) => {
    const len = (it.share / 100) * C;
    const arc = { color: it.color, dash: `${Math.max(0, len - 3)} ${C}`, offset: -off };
    off += len;
    return arc;
  });
});
</script>

<template>
  <ModHead :title="title">{{ d.unit }}</ModHead>
  <div :class="variant === 'ring' ? 'lg-ring' : 'lg-barwrap'" @pointerleave="hover = null">
    <div v-if="variant === 'ring'" class="ctr">
      <svg viewBox="0 0 124 124">
        <circle
          v-for="(a, i) in arcs"
          :key="i"
          r="50"
          cx="62"
          cy="62"
          :stroke="a.color"
          :style="{ '--dash': a.dash, '--c': `0 ${C}` }"
          :stroke-dashoffset="a.offset"
          :class="{ dim: hover !== null && hover !== i }"
          @pointerenter="hover = i"
        />
      </svg>
      <span><b>{{ items.length }}<small>{{ t('aboutKit.languages.kinds') }}</small></b></span>
    </div>
    <div v-else class="lg-bar" :class="{ hov: hover !== null }">
      <i
        v-for="(it, i) in items"
        :key="it.name"
        :class="{ on: hover === i }"
        :style="{ '--c': it.color, '--w': `${it.share}%` }"
        @pointerenter="hover = i"
      />
    </div>
    <ul class="lg-leg">
      <li
        v-for="(it, i) in items"
        :key="it.name"
        :class="{ on: hover === i }"
        :style="{ '--c': it.color }"
        @pointerenter="hover = i"
      >
        <i /><span>{{ it.name }}</span><b>{{ it.percent }}<small>%</small></b>
        <span v-if="variant !== 'ring'" class="lg-track"><em :style="{ '--w': `${(it.share / maxShare) * 100}%`, '--k': i }" /></span>
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
/* 堆叠条撑满宽度 → 每种语言一行：色块 + 名称 + 等宽百分比，下方同宽比较条；行在模块内均匀铺开 */
.lg-barwrap { display: flex; flex-direction: column; flex: 1; }

.lg-bar {
  display: flex;
  gap: 3px;
  height: 14px;
  border-radius: var(--r-pill);
  overflow: hidden;

  i {
    position: relative;
    width: var(--w);
    height: 100%;
    border-radius: calc(var(--r-xs) * 0.5);
    background: var(--c);
    transition: opacity var(--dur);

    &:first-child { border-top-left-radius: var(--r-pill); border-bottom-left-radius: var(--r-pill); }
    &:last-child { border-top-right-radius: var(--r-pill); border-bottom-right-radius: var(--r-pill); }
  }

  &.hov i { opacity: 0.35; }
  &.hov i.on { opacity: 1; }
}

.in .lg-bar i { animation: lg-grow 1.2s var(--ease-out) both; }

@keyframes lg-grow { from { width: 0; } }

.lg-leg {
  list-style: none;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  flex: 1;
  gap: 4px;
  margin-top: 16px;

  li {
    display: grid;
    grid-template-columns: 10px minmax(0, 1fr) auto;
    column-gap: 10px;
    row-gap: 6px;
    align-items: center;
    padding: 6px 10px;
    margin: 0 -10px;
    border-radius: var(--r-sm);
    font-size: 15px;
    font-weight: 500;
    cursor: default;
    transition: background var(--dur-fast);

    &:hover, &.on { background: var(--ak-sunken); }
    > i { width: 10px; height: 10px; border-radius: max(2px, calc(var(--r-xs) * 0.6)); background: var(--c); }
    > b { font: 600 15px var(--ak-mono); font-variant-numeric: tabular-nums; }
    > b small { font-size: 12px; font-weight: 400; color: var(--ak-text-3); }
  }
}

.lg-track {
  grid-column: 2 / -1;
  height: 4px;
  overflow: hidden;
  border-radius: var(--r-pill);
  background: var(--fill-2);

  em { display: block; width: 0; height: 100%; border-radius: inherit; background: var(--c); transition: width 1s var(--ease-out); transition-delay: calc(var(--k) * 70ms + 200ms); }
}

.in .lg-track em { width: var(--w); }

@container (min-width: 520px) {
  .lg-barwrap .lg-leg { display: grid; grid-template-columns: 1fr 1fr; column-gap: 28px; align-content: space-between; }
}

.lg-ring {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 22px;
  align-items: center;

  .lg-leg { margin-top: 0; }

  .ctr { position: relative; }
  svg { display: block; width: 140px; height: 140px; transform: rotate(-90deg); }

  circle {
    fill: none;
    stroke-width: 12;
    stroke-dasharray: var(--dash);
    transition: opacity var(--dur);

    &.dim { opacity: 0.25; }
  }

  .ctr span {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    text-align: center;
    font: 600 26px/1.1 var(--ak-mono);

    small { display: block; margin-top: 2px; font: 400 12px var(--font-sans); color: var(--ak-text-3); }
  }
}

.in .lg-ring circle { animation: lg-dash 1.2s var(--ease-out) both; }

@keyframes lg-dash { from { stroke-dasharray: var(--c); } }

@container (max-width: 300px) { .lg-ring { grid-template-columns: 1fr; justify-items: center; } }

@media (prefers-reduced-motion: reduce) { .lg-track em { width: var(--w); transition: none; } }
</style>

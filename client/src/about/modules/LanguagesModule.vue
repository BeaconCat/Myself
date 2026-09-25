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
        <i /><span>{{ it.name }}</span><b>{{ it.percent }}%</b>
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
.lg-bar {
  display: flex;
  gap: 3px;
  height: 12px;
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
  display: grid;
  grid-template-columns: 1fr;
  gap: 2px;
  margin-top: 18px;

  li {
    display: grid;
    grid-template-columns: 10px minmax(0, 1fr) auto;
    gap: 10px;
    align-items: center;
    padding: 6px 8px;
    margin: 0 -8px;
    border-radius: var(--r-sm);
    font-size: 13.5px;
    cursor: default;
    transition: background var(--dur-fast);

    &:hover, &.on { background: var(--ak-sunken); }
    i { width: 10px; height: 10px; border-radius: calc(var(--r-xs) * 0.6); background: var(--c); }
    b { font: 500 12.5px var(--ak-mono); color: var(--text-2); }
  }
}

@container (min-width: 520px) { .lg-barwrap .lg-leg { grid-template-columns: 1fr 1fr; column-gap: 24px; } }

.lg-ring {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 22px;
  align-items: center;

  .lg-leg { margin-top: 0; }

  .ctr { position: relative; }
  svg { display: block; width: 124px; height: 124px; transform: rotate(-90deg); }

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
    font: 500 20px/1.1 var(--ak-mono);

    small { display: block; font: 400 10px var(--font-sans); color: var(--ak-text-3); }
  }
}

.in .lg-ring circle { animation: lg-dash 1.2s var(--ease-out) both; }

@keyframes lg-dash { from { stroke-dasharray: var(--c); } }

@container (max-width: 300px) { .lg-ring { grid-template-columns: 1fr; justify-items: center; } }
</style>

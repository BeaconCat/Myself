<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { YearData } from '../types';
import type { ModProps } from './props';
import CountUp from '../parts/CountUp.vue';
import ModHead from '../parts/ModHead.vue';

/** 年度回顾（year）：按年切换 —— 大号年份 + 统计条 + 撑满宽度的 12 个月柱状图（峰值实色、未来月虚线、黄点标大事）+ 高光 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as YearData);
const { t } = useI18n();

const years = computed(() => Object.keys(d.value.years ?? {}).sort().reverse());
const picked = ref('');
const year = computed(() => (years.value.includes(picked.value) ? picked.value : years.value[0] ?? ''));
const Y = computed(() => d.value.years?.[year.value]);

const bars = computed(() => {
  const months = Y.value?.months ?? [];
  const max = Math.max(1, ...months.filter((v): v is number => v != null));
  return months.map((v, i) => ({
    v,
    future: v == null,
    peak: v != null && v === max && max > 0,
    h: v == null ? 18 : Math.max(6, (v / max) * 100),
    pin: (Y.value?.pins ?? []).includes(i),
  }));
});
</script>

<template>
  <ModHead :title="title">
    <div v-if="years.length > 1" class="ak-seg">
      <button v-for="y in years" :key="y" :class="{ on: y === year }" @click="picked = y">{{ y }}</button>
    </div>
    <template v-else>{{ d.metric }}</template>
  </ModHead>
  <div v-if="Y" :key="year" class="yr">
    <div class="yr-top">
      <div class="yr-big">{{ year }}<small>{{ Y.sub }}</small></div>
      <div class="ak-statbar yr-nums" :style="{ '--n': Y.nums.length || 4 }">
        <div v-for="([k, v], i) in Y.nums" :key="i" class="ak-stat"><b><CountUp :value="Number(v) || 0" /></b><span>{{ k }}</span></div>
      </div>
    </div>
    <div class="yr-chart" :data-metric="d.metric">
      <div v-for="(b, i) in bars" :key="i" class="c">
        <div class="b" :class="{ fut: b.future, peak: b.peak }" :style="{ '--k': i, '--h': `${b.h}%` }">
          <span v-if="!b.future" class="v">{{ b.v }}</span>
        </div>
      </div>
    </div>
    <div class="yr-mo">
      <span v-for="(b, i) in bars" :key="i" :class="{ pin: b.pin }">{{ t('aboutKit.github.month', { m: i + 1 }) }}</span>
    </div>
    <ul v-if="Y.highlights.length" class="yr-hl">
      <li v-for="([m, text], i) in Y.highlights" :key="i"><b>{{ m }}</b>{{ text }}</li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
.yr { display: flex; flex-direction: column; flex: 1; animation: ak-fade-up 0.45s var(--ease-out); }

.yr-top { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: center; gap: 28px; margin-bottom: 20px; }

.yr-big {
  font: 700 56px/0.95 var(--font-serif);
  letter-spacing: -0.04em;

  small { display: block; margin-top: 10px; font: 400 13px var(--font-sans); letter-spacing: 0; color: var(--ak-text-3); }
}

.yr-nums { width: 100%; max-width: 640px; justify-self: end; }

/* 柱状图撑满模块宽度：柱子随列宽放大（不设上限），数值 12px 等宽 */
.yr-chart {
  position: relative;
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 10px;
  align-items: end;
  height: 190px;
  padding-top: 26px;
  border-bottom: 1px solid var(--ak-line-2);

  .c { position: relative; display: flex; flex-direction: column; justify-content: flex-end; align-items: center; height: 100%; }

  .b {
    position: relative;
    width: 100%;
    height: var(--h);
    border-radius: var(--r-xs) var(--r-xs) calc(var(--r-xs) * 0.4) calc(var(--r-xs) * 0.4);
    background: color-mix(in oklab, var(--primary) 36%, var(--fill-2));

    &.fut {
      background: repeating-linear-gradient(135deg, var(--ak-line-2) 0 1px, transparent 1px 6px);
      border: 1px dashed var(--ak-line-2);
      border-bottom: 0;
    }

    &.peak { background: color-mix(in oklab, var(--primary) 80%, var(--fill-2)); }
  }

  .v { position: absolute; top: -21px; left: 50%; translate: -50% 0; font: 500 12.5px var(--ak-mono); color: var(--text-2); }
  .peak .v { color: var(--ak-ink); font-weight: 600; }
}

.in .yr-chart .b { animation: yr-grow 1s var(--ease-out) both; animation-delay: calc(var(--k) * 60ms); }

@keyframes yr-grow { from { height: 0; } }

.yr-mo {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 10px;
  margin-top: 8px;
  font: 400 12px var(--ak-mono);
  color: var(--ak-text-3);
  text-align: center;

  .pin { color: var(--text); }

  .pin::before {
    content: '';
    display: inline-block;
    width: 6px;
    height: 6px;
    margin-right: 4px;
    border-radius: 50%;
    vertical-align: middle;
    background: var(--ak-yellow);
  }
}

.yr-hl {
  list-style: none;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 24px;
  margin-top: auto;
  padding-top: 20px;

  li { font-size: 15px; line-height: 1.6; color: var(--text); }

  b {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
    font: 500 12.5px var(--ak-mono);
    letter-spacing: 0.04em;
    color: var(--ak-text-3);

    &::before { content: ''; width: 6px; height: 6px; border-radius: 50%; background: var(--ak-yellow); }
  }
}

@container (max-width: 760px) {
  .yr-top { grid-template-columns: 1fr; gap: 16px; }
  .yr-nums { max-width: none; }
}

@container (max-width: 640px) {
  .yr-hl { grid-template-columns: 1fr; gap: 12px; }
  .yr-big { font-size: 44px; }
  .yr-chart { height: 160px; }
  .yr-chart, .yr-mo { gap: 4px; }
  .yr-chart .v { display: none; }
  .yr-mo { font-size: 10px; }
}
</style>

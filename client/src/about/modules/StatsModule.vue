<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { api } from '../../api';
import { useConfigStore } from '../../stores/config';
import type { StatsData } from '../types';
import type { ModProps } from './props';
import CountUp from '../parts/CountUp.vue';
import ModHead from '../parts/ModHead.vue';

/**
 * 站点数字（stats）：运行天数 + 文章 / 随想 / 标签实时统计（统计条：大号等宽数字 + 三级灰标签），
 * 下方本年进度 = 12 段月份条撑满模块宽度（已过月份实填、本月按天数部分填充）。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as StatsData);
const { t } = useI18n();
const config = useConfigStore();

const live = reactive({ posts: 0, notes: 0, tags: 0 });

const days = computed(() => {
  const from = new Date(`${config.cfg.about.foundedAt || '2026-01-01'}T00:00:00`).getTime();
  return Math.max(1, Math.floor((Date.now() - from) / 864e5));
});

const items = computed(() =>
  d.value.items.map((it) => ({
    key: it.key,
    value: it.key === 'days' ? days.value : (live[it.key as 'posts' | 'notes' | 'tags'] ?? 0),
    label: it.label || t(`aboutKit.stats.${it.key}`),
    hint: it.hint ?? '',
  })),
);

const year = new Date().getFullYear();
const pct = computed(() => {
  const y0 = new Date(year, 0, 1).getTime();
  const y1 = new Date(year + 1, 0, 1).getTime();
  return Math.min(100, ((Date.now() - y0) / (y1 - y0)) * 100);
});

/** 12 个月的填充比例：已过 = 1，本月按日期，未来 = 0 */
const months = computed(() => {
  const now = new Date();
  return Array.from({ length: 12 }, (_, m) => {
    if (m < now.getMonth()) return 1;
    if (m > now.getMonth()) return 0;
    const dim = new Date(year, m + 1, 0).getDate();
    return now.getDate() / dim;
  });
});
const daysLeft = computed(() => Math.max(0, Math.ceil((new Date(year + 1, 0, 1).getTime() - Date.now()) / 864e5)));

onMounted(async () => {
  try {
    const [posts, notes, tags] = await Promise.all([api.posts({ pageSize: 1 }), api.notes({ pageSize: 1 }), api.tags()]);
    live.posts = posts.total;
    live.notes = notes.total;
    live.tags = tags.length;
  } catch { /* 后端未启动时统计留零 */ }
});
</script>

<template>
  <ModHead :title="title"><span class="ak-dot live" />{{ t('aboutKit.live') }}</ModHead>
  <div class="ak-statbar st" :class="{ hero: variant === 'hero' }" :style="{ '--n': items.length }">
    <div v-for="it in items" :key="it.key" class="ak-stat">
      <b><CountUp :value="it.value" /><sup v-if="it.key === 'days'">{{ t('aboutKit.stats.dayUnit') }}</sup></b>
      <span>{{ it.label }}</span>
      <em v-if="it.hint">{{ it.hint }}</em>
    </div>
  </div>
  <div v-if="d.showYearProgress" class="st-year">
    <div class="yh">
      <span>{{ t('aboutKit.stats.yearPassed', { y: year }) }}</span>
      <b class="ak-mono">{{ pct.toFixed(1) }}<small>%</small></b>
      <em>{{ t('aboutKit.stats.yearLeft', { n: daysLeft }) }}</em>
    </div>
    <div class="ym">
      <div v-for="(f, m) in months" :key="m" class="c" :class="{ cur: f > 0 && f < 1 }">
        <i :style="{ '--f': f, '--k': m }" />
        <span>{{ t('aboutKit.github.month', { m: m + 1 }) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.st { flex: 1; min-height: 104px; }

/* 与同排卡片等高时，统计格纵向拉伸，数字与标签成组垂直居中（不留上下空洞） */
.st .ak-stat { display: flex; flex-direction: column; justify-content: center; }

.st.hero .ak-stat > b { font-size: clamp(30px, 6cqi, 44px); }

/* 本年进度：标题行（百分比大号等宽）+ 12 段月份条撑满宽度 */
.st-year {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: auto;
  padding-top: 20px;
}

.yh {
  display: flex;
  align-items: baseline;
  gap: 10px;
  font-size: 13px;
  color: var(--ak-text-3);

  b { font-size: 22px; font-weight: 600; color: var(--text); font-variant-numeric: tabular-nums; }
  small { margin-left: 1px; font-size: 13px; color: var(--ak-text-3); }
  em { margin-left: auto; font-style: normal; }
}

.ym {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 4px;

  .c { display: flex; flex-direction: column; gap: 6px; min-width: 0; }

  i {
    position: relative;
    display: block;
    height: 10px;
    overflow: hidden;
    border-radius: var(--r-xs);
    background: var(--fill-2);

    &::after {
      content: '';
      position: absolute;
      inset: 0;
      border-radius: inherit;
      background: color-mix(in oklab, var(--primary) 62%, var(--fill-2));
      transform: scaleX(0);
      transform-origin: left;
      transition: transform 0.7s var(--ease-out);
      transition-delay: calc(var(--k) * 45ms + 200ms);
    }
  }

  .cur i::after { background: var(--ink); }

  span { overflow: hidden; font: 400 12px var(--ak-mono); color: var(--ak-text-3); white-space: nowrap; text-align: center; }
  .cur span { color: var(--text); }
}

.in .ym i::after { transform: scaleX(var(--f)); }

@container (max-width: 520px) {
  .st { flex: none; }
  .ym span { font-size: 10px; }
  .ym .c:nth-child(odd) span { visibility: hidden; }
}

@media (prefers-reduced-motion: reduce) {
  .ym i::after { transform: scaleX(var(--f)); transition: none; }
}
</style>

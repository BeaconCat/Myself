<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { api } from '../../api';
import { useConfigStore } from '../../stores/config';
import type { StatsData } from '../types';
import type { ModProps } from './props';
import CountUp from '../parts/CountUp.vue';
import ModHead from '../parts/ModHead.vue';

/** 站点数字（stats）：运行天数 + 文章 / 随想 / 标签实时统计，细分隔线 + 本年进度条 */
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
  <div class="st" :class="{ hero: variant === 'hero' }" :style="{ '--n': items.length }">
    <div v-for="it in items" :key="it.key">
      <b><CountUp :value="it.value" /><sup v-if="it.key === 'days'">{{ t('aboutKit.stats.dayUnit') }}</sup></b>
      <span>{{ it.label }}</span>
      <em v-if="it.hint">{{ it.hint }}</em>
    </div>
  </div>
  <div v-if="d.showYearProgress" class="st-foot">
    <span>{{ t('aboutKit.stats.yearPassed', { y: year }) }}</span>
    <span class="bar"><i :style="{ '--w': `${pct.toFixed(1)}%` }" /></span>
    <span class="ak-mono">{{ pct.toFixed(1) }}%</span>
  </div>
</template>

<style scoped lang="scss">
.st {
  display: grid;
  grid-template-columns: repeat(var(--n, 4), minmax(0, 1fr));
  margin-bottom: 26px;

  > div { padding: 4px 22px 2px; border-left: 1px solid var(--ak-line); min-width: 0; }
  > div:first-child { border-left: 0; padding-left: 0; }

  b {
    display: block;
    font: 500 50px/1 var(--ak-mono);
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.04em;
    color: var(--text);
    white-space: nowrap;
  }

  sup {
    position: relative;
    top: 0.3em;
    margin-left: 3px;
    font-size: 0.36em;
    letter-spacing: 0;
    vertical-align: top;
    color: var(--ak-ink);
    font-family: var(--font-sans);
  }

  > div > span { display: block; margin-top: 12px; font-size: 13px; color: var(--text-2); }
  > div > em { display: block; margin-top: 3px; font: normal 11.5px/1.4 var(--ak-mono); color: var(--ak-text-3); }

  &.hero b { font-size: clamp(40px, 8cqi, 64px); }
}

.st-foot {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: auto;
  padding-top: 16px;
  border-top: 1px dashed var(--ak-line-2);
  font-size: 12.5px;
  color: var(--ak-text-3);

  .bar { flex: 1; height: 4px; border-radius: 4px; background: var(--ak-sunken); overflow: hidden; }

  .bar i {
    display: block;
    height: 100%;
    width: 0;
    border-radius: inherit;
    background: linear-gradient(90deg, rgba(var(--primary-rgb), 0.3), var(--primary));
    transition: width 1.6s 0.3s var(--ease-out);
  }
}

.in .st-foot .bar i { width: var(--w); }

@container (max-width: 560px) {
  .st { grid-template-columns: 1fr 1fr; row-gap: 22px; }
  .st > div:nth-child(3) { border-left: 0; padding-left: 0; }
  .st b { font-size: 38px; }
}

@container (max-width: 300px) {
  .st > div { padding: 4px 12px 2px; }
  .st b { font-size: 30px; }
}
</style>

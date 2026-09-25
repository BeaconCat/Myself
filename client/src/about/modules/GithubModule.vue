<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { api } from '../../api';
import { useConfigStore } from '../../stores/config';
import type { GithubData } from '../types';
import type { ModProps } from './props';
import CountUp from '../parts/CountUp.vue';
import KitIcon from '../parts/KitIcon.vue';
import ModHead from '../parts/ModHead.vue';
import { ago } from '../useClock';

/**
 * GitHub：四项计数 + 53 周贡献热力图（悬停提示日期与次数）+ 最近动态。
 * 数据来自 /api/v1/github-status（服务端缓存）；窄容器裁剪周数而不压扁格子。
 */
const props = defineProps<ModProps>();
const d = computed(() => props.mod.data as GithubData);
const { t } = useI18n();
const config = useConfigStore();

interface HeatDay { date: string; count: number; level: number }
interface Activity { type: string; repo: string; text: string; time: string }

const remote = ref<{
  stats: { repos: number; stars: number; followers: number; commits: number };
  activities?: Activity[];
  heatmap?: HeatDay[];
} | null>(null);

onMounted(async () => {
  try { remote.value = await api.githubStatus(); } catch { /* 回退配置数字 */ }
});

const user = computed(() => config.cfg.github.username);
const stats = computed(() => remote.value?.stats ?? config.cfg.github.stats);
const head = computed(() => [
  [t('aboutKit.github.repos'), stats.value.repos],
  [t('aboutKit.github.stars'), stats.value.stars],
  [t('aboutKit.github.followers'), stats.value.followers],
  [t('aboutKit.github.commits'), stats.value.commits],
] as [string, number][]);

const WEEKS = 53;

function iso(dt: Date): string {
  return `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`;
}

/** 以今天为末列，按周日起排列 53 周；无数据时为空格子 */
const grid = computed(() => {
  const byDate = new Map((remote.value?.heatmap ?? []).map((h) => [h.date, h]));
  const end = new Date();
  end.setHours(0, 0, 0, 0);
  const start = new Date(end);
  start.setDate(start.getDate() - end.getDay() - (WEEKS - 1) * 7);
  const cells: { key: string; level: number; tip: string; future: boolean; week: number }[] = [];
  const months: { label: string; week: number }[] = [];
  let lastMonth = -1;
  for (let i = 0; i < WEEKS * 7; i++) {
    const dt = new Date(start);
    dt.setDate(start.getDate() + i);
    const week = Math.floor(i / 7);
    if (i % 7 === 0) {
      const m = dt.getMonth();
      months.push({ label: m !== lastMonth && week < WEEKS - 2 ? t('aboutKit.github.month', { m: m + 1 }) : '', week });
      lastMonth = m;
    }
    const key = iso(dt);
    const h = byDate.get(key);
    const count = h?.count ?? 0;
    const level = h ? Math.max(0, Math.min(4, h.level)) : 0;
    cells.push({
      key,
      level,
      future: dt > end,
      week,
      tip: `${key.replaceAll('-', '.')} · ${t('aboutKit.github.tip', { n: count })}`,
    });
  }
  return { cells, months };
});

const olderCls = (week: number) => ({ o1: week < WEEKS - 33, o2: week < WEEKS - 20 });

const commits = computed(() => (remote.value?.activities ?? []).slice(0, d.value.commitCount || 3));
const shortType = (type: string) => (type || 'event').replace(/Event$/, '').toLowerCase().slice(0, 7);
</script>

<template>
  <ModHead :title="title">
    <a class="ak-link-arrow gh-user" :href="`https://github.com/${user}`" target="_blank" rel="noopener">@{{ user }}<KitIcon name="arrow" :size="15" /></a>
  </ModHead>
  <div class="ak-statbar gh-head">
    <div v-for="[k, v] in head" :key="k" class="ak-stat"><b><CountUp :value="v" /></b><span>{{ k }}</span></div>
  </div>
  <div class="hm-wrap">
    <div class="hm-months">
      <span v-for="m in grid.months" :key="m.week" :class="olderCls(m.week)">{{ m.label }}</span>
    </div>
    <div class="hm">
      <i
        v-for="c in grid.cells"
        :key="c.key"
        :class="[olderCls(c.week), { f: c.future }]"
        :data-l="c.level"
        :data-tip="c.future ? undefined : c.tip"
      />
    </div>
    <div class="hm-legend">
      <span>{{ t('aboutKit.github.pastYear', { n: stats.commits }) }}</span>
      <span class="sc">{{ t('aboutKit.github.less') }}<i /><i data-l="1" /><i data-l="2" /><i data-l="3" /><i data-l="4" />{{ t('aboutKit.github.more') }}</span>
    </div>
  </div>
  <ul v-if="variant !== 'map' && d.showCommits && commits.length" class="gh-commits">
    <li v-for="(c, i) in commits" :key="i">
      <span class="sha">{{ shortType(c.type) }}</span>
      <span class="msg">{{ c.text }}<small>{{ c.repo }}</small></span>
      <time>{{ ago(c.time, t) }}</time>
    </li>
  </ul>
</template>

<style scoped lang="scss">
/* 统计条 → 撑满模块宽度的 53 周热力图 → 最近动态；与首页 GithubStatusCard 同一写法 */
.gh-user { display: flex; align-items: center; gap: 4px; font-size: 13px; color: var(--ak-ink); transition: color var(--dur-fast); }
.gh-user:hover { color: var(--text); }

.gh-head { margin-bottom: 20px; }

.hm-wrap { position: relative; }

.hm-months {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  gap: 3px;
  margin-bottom: 6px;
  font: 400 12px var(--ak-mono);
  color: var(--ak-text-3);

  span { width: 0; overflow: visible; white-space: nowrap; }
}

.hm {
  display: grid;
  grid-template-rows: repeat(7, 1fr);
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  gap: 3px;

  i {
    aspect-ratio: 1;
    border-radius: max(2px, calc(var(--r-xs) * 0.6));
    background: var(--fill-2);
    transition: transform var(--dur-fast) var(--ease-spring);

    &:hover { position: relative; z-index: 2; transform: scale(1.5); }
    &.f { visibility: hidden; }
  }
}

/* 热力图：主色单色阶梯（向 --fill-2 混合），数据信号，不发光 */
.hm i, .sc i {
  &[data-l='1'] { background: color-mix(in oklab, var(--primary) 28%, var(--fill-2)); }
  &[data-l='2'] { background: color-mix(in oklab, var(--primary) 52%, var(--fill-2)); }
  &[data-l='3'] { background: color-mix(in oklab, var(--primary) 76%, var(--fill-2)); }
  &[data-l='4'] { background: var(--primary); }
}

.hm-legend {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
  font: 400 12px var(--ak-mono);
  color: var(--ak-text-3);

  .sc { display: flex; align-items: center; gap: 4px; }
  .sc i { width: 11px; height: 11px; border-radius: max(2px, calc(var(--r-xs) * 0.5)); background: var(--fill-2); }
}

@container (max-width: 680px) { .hm .o1, .hm-months .o1 { display: none; } }

@container (max-width: 440px) { .hm .o2, .hm-months .o2 { display: none; } }

.gh-commits {
  list-style: none;
  margin-top: auto;
  padding-top: 16px;

  li {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 12px;
    align-items: baseline;
    padding: 10px 0;
    box-shadow: 0 -1px 0 var(--ak-line);
    font-size: 14px;

    &:last-child { padding-bottom: 0; }
  }

  .sha { font: 500 12.5px var(--ak-mono); color: var(--ak-ink); }
  .msg { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .msg small { margin-left: 8px; font: 12.5px var(--ak-mono); color: var(--ak-text-3); }
  time { font: 400 12.5px var(--ak-mono); color: var(--ak-text-3); }
}

@container (max-width: 420px) {
  .gh-commits .sha { display: none; }
  .gh-commits li { grid-template-columns: minmax(0, 1fr) auto; }
}
</style>

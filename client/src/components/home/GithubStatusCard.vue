<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import { formatDateTime } from '../../utils/date';
import type { GhActivity, GhStatus, HeatDay } from './format';

/**
 * GitHub Status：数据由首页统一经后端 /github-status 获取后传入
 * （manual 模式回配置数字；api 模式服务端拉 GitHub API 并缓存）。
 * 中性卡片面；数字等宽；热力格用主色阶作数据信号，无发光。
 */
const props = defineProps<{ data: GhStatus | null; loading?: boolean }>();

const { t } = useI18n();
const config = useConfigStore();

const stats = computed(() => {
  const s = props.data?.stats ?? config.cfg.github.stats;
  return [
    { label: '仓库', value: s.repos },
    { label: 'Stars', value: s.stars },
    { label: '关注者', value: s.followers },
    { label: '年度提交', value: s.commits },
  ];
});

const activities = computed<GhActivity[]>(() => props.data?.activities ?? []);

function timeOf(iso: string): string {
  try {
    return formatDateTime(iso.replace('T', ' ').replace('Z', '').slice(0, 19));
  } catch {
    return iso.slice(0, 10);
  }
}

const WEEKS = 53;

/**
 * 真实贡献热力：取最近 53 周（一年），按 GitHub 周对齐（列 = 周，行 = 周日起）。
 * 无数据（manual 模式 / 拉取失败）时隐藏热力图。
 */
const heatmap = computed<HeatDay[]>(() => {
  const days = props.data?.heatmap ?? [];
  if (!days.length) return [];
  const tail = days.slice(-WEEKS * 7);
  const firstWeekday = new Date(`${tail[0].date}T00:00:00Z`).getUTCDay();
  const pad: HeatDay[] = Array.from({ length: firstWeekday }, (_, i) => ({ date: `pad-${i}`, count: -1, level: -1 }));
  return [...pad, ...tail];
});
</script>

<template>
  <section class="gh-card">
    <header class="gh-head">
      <div class="gh-title">
        <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" /></svg>
        <h2>GitHub</h2>
      </div>
      <a class="gh-user" :href="`https://github.com/${config.cfg.github.username}`" target="_blank" rel="noopener">
        {{ '@' + config.cfg.github.username }}
      </a>
    </header>

    <div class="gh-stats">
      <div v-for="s in stats" :key="s.label" class="stat">
        <span v-if="loading" class="sk num-sk" />
        <strong v-else>{{ s.value }}</strong>
        <span class="lbl">{{ s.label }}</span>
      </div>
    </div>

    <div v-if="heatmap.length" class="heat-wrap">
      <div class="heatmap" role="img" :aria-label="t('home.contrib')">
        <span
          v-for="day in heatmap"
          :key="day.date"
          class="cell"
          :class="day.level < 0 ? 'pad' : `l${day.level}`"
          :title="day.level < 0 ? undefined : (day.count > 0 ? t('home.contribTip', { d: day.date, n: day.count }) : day.date)"
        />
      </div>
      <div class="legend">
        <span>{{ t('home.less') }}</span>
        <i class="cell l0" /><i class="cell l1" /><i class="cell l2" /><i class="cell l3" /><i class="cell l4" />
        <span>{{ t('home.more') }}</span>
      </div>
    </div>

    <ul v-if="activities.length" class="activity">
      <li v-for="(a, i) in activities.slice(0, 3)" :key="i">
        <span class="a-repo">{{ a.repo }}</span>
        <span class="a-text">{{ a.text }}</span>
        <span class="a-time">{{ timeOf(a.time) }}</span>
      </li>
    </ul>
  </section>
</template>

<style scoped lang="scss">
/* 标题 → 四个大号数字 → 撑满整宽的大格热力图 → 最近提交；中性卡面，热力格用主色阶作数据信号 */
.gh-card {
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
  padding: 24px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
}

.gh-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.gh-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text);

  svg { width: 26px; height: 26px; }
  h2 { font-family: var(--font-serif); font-size: 24px; font-weight: 700; line-height: 1.2; }
}

.gh-user {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--ink);

  &:hover { text-decoration: underline; text-underline-offset: 3px; }
}

.gh-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-radius: var(--r-md);
  background: var(--fill);
}

.stat {
  padding: 16px 16px 14px;
  min-width: 0;

  & + & { box-shadow: -1px 0 0 var(--line); }

  strong {
    display: block;
    font-family: var(--font-mono);
    font-size: 30px;
    line-height: 1.1;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--text);
  }

  .lbl {
    display: block;
    margin-top: 6px;
    font-size: 12.5px;
    color: var(--text-3);
  }
}

.num-sk { display: block; width: 50%; height: 33px; border-radius: var(--r-xs); }

/* 热力图撑满卡片宽度：53 周 x 7 天，格子随宽度等比放大 */
.heat-wrap { display: flex; flex-direction: column; gap: 8px; }

.heatmap {
  display: grid;
  grid-template-rows: repeat(7, auto);
  grid-auto-flow: column;
  grid-auto-columns: minmax(0, 1fr);
  gap: 3px;
}

.cell {
  aspect-ratio: 1;
  border-radius: max(2px, calc(var(--r-xs) * 0.6));
  background: var(--fill-2);

  &.pad { visibility: hidden; }
  &.l1 { background: color-mix(in oklab, var(--primary) 28%, var(--fill-2)); }
  &.l2 { background: color-mix(in oklab, var(--primary) 52%, var(--fill-2)); }
  &.l3 { background: color-mix(in oklab, var(--primary) 76%, var(--fill-2)); }
  &.l4 { background: var(--primary); }
}

.legend {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  font-size: 12px;
  color: var(--text-3);

  .cell { width: 11px; }
  span:first-child { margin-right: 4px; }
  span:last-child { margin-left: 4px; }
}

.activity {
  display: flex;
  flex-direction: column;
  margin: 0;
  padding: 0;
  list-style: none;

  li {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 12px;
    align-items: baseline;
    padding: 9px 0;
    font-size: 13.5px;
    box-shadow: 0 -1px 0 var(--line);
  }

  .a-repo { font-family: var(--font-mono); font-size: 12.5px; color: var(--ink); }
  .a-text { color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .a-time { font-size: 12px; color: var(--text-3); }
}

@media (max-width: 560px) {
  .gh-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .stat:nth-child(3) { box-shadow: 0 -1px 0 var(--line); }
  .stat:nth-child(4) { box-shadow: -1px 0 0 var(--line), 0 -1px 0 var(--line); }
}
</style>

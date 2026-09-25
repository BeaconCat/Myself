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
        <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
          <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" />
        </svg>
        <h2>GitHub</h2>
      </div>
      <a class="gh-user" :href="`https://github.com/${config.cfg.github.username}`" target="_blank" rel="noopener">
        {{ '@' + config.cfg.github.username }}
      </a>
    </header>

    <div class="gh-body" :class="{ heat: heatmap.length }">
    <div class="gh-stats">
      <div v-for="s in stats" :key="s.label" class="stat">
        <span v-if="loading" class="sk sk-line num-sk" />
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
    </div>

    <ul v-if="activities.length" class="activity">
      <li v-for="(a, i) in activities" :key="i">
        <span class="a-dot" :class="a.type" />
        <span class="a-repo">{{ a.repo }}</span>
        <span class="a-text">{{ a.text }}</span>
        <span class="a-time">{{ timeOf(a.time) }}</span>
      </li>
    </ul>
  </section>
</template>

<style scoped lang="scss">
.gh-card {
  padding: 28px 32px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
}

:root[data-mode='dark'] .gh-card {
  background: linear-gradient(180deg, color-mix(in oklab, var(--surface) 90%, white), var(--surface) 70%);
}

.gh-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 22px;
}

.gh-title {
  display: flex;
  align-items: center;
  gap: 10px;

  svg { width: 20px; height: 20px; color: var(--text); }

  h2 {
    font-family: var(--font-serif);
    font-size: 20px;
    font-weight: 700;
  }
}

/* 用户名：信号色文字链接（无底色块） */
.gh-user {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--ink);
  padding: 2px 0;
  border-radius: var(--r-xs);
  background: linear-gradient(currentColor, currentColor) 0 100% / 0 1px no-repeat;
  transition: background-size var(--dur) var(--ease-out);

  &:hover { background-size: 100% 1px; }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

/* 有热力图：左侧 2×2 统计 + 右侧热力；否则统计一行四列 */
.gh-body.heat {
  display: grid;
  grid-template-columns: minmax(0, 260px) minmax(0, 1fr);
  gap: 40px;
  align-items: center;
  margin-bottom: 8px;

  .gh-stats {
    grid-template-columns: repeat(2, 1fr);
    row-gap: 18px;
    margin-bottom: 0;
  }

  .stat:nth-child(3) {
    padding-left: 0;
    box-shadow: none;
  }
}

.heat-wrap {
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

/* 统计：发丝分隔的一行数字 */
.gh-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  margin-bottom: 22px;
}

.stat {
  padding: 2px 20px;
  box-shadow: inset 0.5px 0 0 var(--line-2);

  &:first-child {
    padding-left: 0;
    box-shadow: none;
  }

  strong {
    display: block;
    font-family: var(--font-mono);
    font-size: 22px;
    font-weight: 500;
    line-height: 1.2;
    color: var(--text);
    font-variant-numeric: tabular-nums;
  }

  .lbl {
    font-size: 12px;
    color: var(--text-3);
  }

  .num-sk {
    display: block;
    width: 44px;
    height: 20px;
    margin: 3px 0 4px;
  }
}

/* 热力图：53 周 × 7 天，格子上限 ~14px（宽屏不再撑成大方块） */
.heatmap {
  display: grid;
  grid-template-rows: repeat(7, auto);
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  gap: 3px;
  width: 100%;
  max-width: 820px;
  margin-bottom: 10px;
}

.cell {
  aspect-ratio: 1;
  border-radius: var(--r-xs);
  min-width: 0;

  &.pad { background: transparent; }
  &.l0 { background: var(--fill-2); }
  &.l1 { background: color-mix(in oklab, var(--primary) 28%, var(--fill-2)); }
  &.l2 { background: color-mix(in oklab, var(--primary) 50%, var(--fill-2)); }
  &.l3 { background: color-mix(in oklab, var(--primary) 75%, var(--fill-2)); }
  &.l4 { background: var(--primary); }
}

.heatmap .cell { border-radius: min(var(--r-xs), 3px); }

.legend {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: flex-end;
  margin-bottom: 18px;

  span { font-size: 11px; color: var(--text-3); margin: 0 4px; }

  i {
    width: 10px;
    height: 10px;
    display: inline-block;
    border-radius: min(var(--r-xs), 3px);
  }
}

.activity {
  list-style: none;

  li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 11px 0;
    box-shadow: inset 0 0.5px 0 var(--line-2);
    font-size: 13px;
  }
}

.a-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background: var(--text-3);

  &.commit { background: var(--accent-blue); }
  &.star { background: var(--accent-yellow); }
}

.a-repo {
  font-weight: 500;
  flex-shrink: 0;
}

.a-text {
  color: var(--text-2);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.a-time {
  color: var(--text-3);
  font-family: var(--font-mono);
  font-size: 12px;
  flex-shrink: 0;
}

@media (max-width: 1000px) {
  .gh-body.heat { grid-template-columns: 1fr; gap: 22px; }
}

@media (max-width: 768px) {
  .gh-card { padding: 22px 20px; }

  .gh-stats {
    grid-template-columns: repeat(2, 1fr);
    row-gap: 14px;
  }

  .stat:nth-child(3) { padding-left: 0; box-shadow: none; }
  .a-text { display: none; }
}
</style>

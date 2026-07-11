<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useConfigStore } from '../../stores/config';
import { formatDateTime } from '../../utils/date';

/**
 * GitHub Status 展示卡：经后端 /github-status 获取
 * （manual 模式回配置数字；api 模式服务端拉 GitHub API 并缓存）。
 */
const config = useConfigStore();

interface GhActivity {
  type: string;
  repo: string;
  text: string;
  time: string;
}

interface HeatDay {
  date: string;
  count: number;
  level: number;
}

const remote = ref<{
  stats: { repos: number; stars: number; followers: number; commits: number };
  activities: GhActivity[];
  heatmap?: HeatDay[];
} | null>(null);

onMounted(async () => {
  try {
    const res = await fetch('/api/v1/github-status');
    if (res.ok) remote.value = await res.json();
  } catch { /* 回退配置数字 */ }
});

const stats = computed(() => {
  const s = remote.value?.stats ?? config.cfg.github.stats;
  return [
    { label: '仓库', value: s.repos },
    { label: 'Stars', value: s.stars },
    { label: '关注者', value: s.followers },
    { label: '年度提交', value: s.commits },
  ];
});

const activities = computed<GhActivity[]>(() => remote.value?.activities ?? []);

function timeOf(iso: string): string {
  try {
    return formatDateTime(iso.replace('T', ' ').replace('Z', '').slice(0, 19));
  } catch {
    return iso.slice(0, 10);
  }
}

const WEEKS = 26;

/**
 * 真实贡献热力：取最近 26 周，按 GitHub 周对齐（列 = 周，行 = 周日起）。
 * 无数据（manual 模式 / 拉取失败）时隐藏热力图。
 */
const heatmap = computed<HeatDay[]>(() => {
  const days = remote.value?.heatmap ?? [];
  if (!days.length) return [];
  const tail = days.slice(-WEEKS * 7);
  // 首格对齐到所在周的周日：前面补空位
  const firstWeekday = new Date(`${tail[0].date}T00:00:00Z`).getUTCDay();
  const pad: HeatDay[] = Array.from({ length: firstWeekday }, (_, i) => ({
    date: `pad-${i}`,
    count: -1,
    level: -1,
  }));
  return [...pad, ...tail];
});
</script>

<template>
  <section v-reveal class="gh-card">
    <header class="gh-head">
      <div class="gh-title">
        <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
          <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" />
        </svg>
        <h2>GitHub Status</h2>
      </div>
      <a
        class="gh-user"
        :href="`https://github.com/${config.cfg.github.username}`"
        target="_blank"
        rel="noopener"
      >{{ '@' + config.cfg.github.username }}</a>
    </header>

    <!-- 统计行 -->
    <div class="gh-stats">
      <div v-for="s in stats" :key="s.label" class="stat">
        <strong>{{ s.value }}</strong>
        <span>{{ s.label }}</span>
      </div>
    </div>

    <!-- 贡献热力图（真实数据；无数据时隐藏） -->
    <template v-if="heatmap.length">
      <div class="heatmap" role="img" aria-label="贡献热力图">
        <span
          v-for="day in heatmap"
          :key="day.date"
          class="cell"
          :class="day.level < 0 ? 'pad' : `l${day.level}`"
          :title="day.level < 0 ? undefined : `${day.date}${day.count > 0 ? ` · ${day.count} 次贡献` : ''}`"
        />
      </div>
      <div class="legend">
        <span>少</span>
        <i class="cell l0" /><i class="cell l1" /><i class="cell l2" /><i class="cell l3" /><i class="cell l4" />
        <span>多</span>
      </div>
    </template>

    <!-- 最新动态（api 模式实时） -->
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
  padding: 28px;
  background: var(--surface);
  border: 1px solid var(--border);
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

  svg { width: 22px; height: 22px; color: var(--text); }
  h2 { font-size: 21px; }
}

.gh-user {
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);
  padding: 4px 12px;
  border-radius: 999px;
  background: rgba(var(--primary-rgb), 0.1);
  transition: transform var(--dur-fast) var(--ease-spring);

  &:hover { transform: scale(1.06); }
}

/* 统计 */
.gh-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 22px;
}

.stat {
  text-align: center;
  padding: 14px 8px;
  background: var(--surface-2);

  strong {
    display: block;
    font-size: 22px;
    font-family: var(--font-serif);
    background: var(--grad-title);
    background-clip: text;
    -webkit-background-clip: text;
    color: transparent;
  }

  span {
    font-size: 12px;
    color: var(--text-2);
  }
}

/* 热力图：26 周 × 7 天 */
.heatmap {
  display: grid;
  grid-template-rows: repeat(7, 1fr);
  grid-auto-flow: column;
  gap: 3px;
  margin-bottom: 10px;
}

.cell {
  aspect-ratio: 1;
  border-radius: 2px;
  min-width: 0;

  &.pad { background: transparent; }
  &.l0 { background: var(--surface-2); }
  &.l1 { background: rgba(var(--primary-rgb), 0.25); }
  &.l2 { background: rgba(var(--primary-rgb), 0.5); }
  &.l3 { background: rgba(var(--primary-rgb), 0.75); }
  &.l4 { background: var(--primary); box-shadow: 0 0 4px rgba(var(--primary-rgb), 0.5); }
}

.legend {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: flex-end;
  margin-bottom: 20px;

  span { font-size: 11px; color: var(--text-2); margin: 0 4px; }
  i { width: 10px; height: 10px; display: inline-block; }
}

/* 动态 */
.activity {
  list-style: none;

  li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 11px 4px;
    border-top: 1px solid var(--border);
    font-size: 13px;
  }
}

.a-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;

  &.commit { background: var(--accent-blue); }
  &.star { background: var(--accent-yellow); }
}

.a-repo {
  font-weight: 700;
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
  color: var(--text-2);
  font-size: 12px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .gh-stats { grid-template-columns: repeat(2, 1fr); }
  .a-text { display: none; }
}
</style>

<script setup lang="ts">
import { computed } from 'vue';

/**
 * GitHub Status 展示卡（设计稿阶段用演示数据渲染）。
 * TODO(P3+)：数据改由后端配置接口下发（/api/v1/widgets/github）。
 */
const mock = {
  username: 'BeaconCat',
  stats: [
    { label: '仓库', value: 32 },
    { label: 'Stars', value: 218 },
    { label: '关注者', value: 47 },
    { label: '年度提交', value: 1286 },
  ],
  activities: [
    { type: 'commit', repo: 'Myself', text: 'feat: hero 3D album carousel with progress bar', time: '2 小时前' },
    { type: 'commit', repo: 'Myself', text: 'fix: circular theme transition flash', time: '5 小时前' },
    { type: 'star', repo: 'vuejs/core', text: 'Starred 仓库', time: '1 天前' },
  ],
};

const WEEKS = 26;
const DAYS = 7;

/** 确定性伪随机热力数据（种子固定，设计稿稳定渲染） */
const heatmap = computed(() => {
  const cells: number[] = [];
  let seed = 20260710;
  for (let i = 0; i < WEEKS * DAYS; i++) {
    seed = (seed * 1103515245 + 12345) % 2147483648;
    const r = seed / 2147483648;
    cells.push(r < 0.32 ? 0 : r < 0.55 ? 1 : r < 0.75 ? 2 : r < 0.9 ? 3 : 4);
  }
  return cells;
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
      <a class="gh-user" href="javascript:;">{{ '@' + mock.username }}</a>
    </header>

    <!-- 统计行 -->
    <div class="gh-stats">
      <div v-for="s in mock.stats" :key="s.label" class="stat">
        <strong>{{ s.value }}</strong>
        <span>{{ s.label }}</span>
      </div>
    </div>

    <!-- 贡献热力图 -->
    <div class="heatmap" role="img" aria-label="贡献热力图">
      <span
        v-for="(level, i) in heatmap"
        :key="i"
        class="cell"
        :class="`l${level}`"
      />
    </div>
    <div class="legend">
      <span>少</span>
      <i class="cell l0" /><i class="cell l1" /><i class="cell l2" /><i class="cell l3" /><i class="cell l4" />
      <span>多</span>
    </div>

    <!-- 最新动态 -->
    <ul class="activity">
      <li v-for="(a, i) in mock.activities" :key="i">
        <span class="a-dot" :class="a.type" />
        <span class="a-repo">{{ a.repo }}</span>
        <span class="a-text">{{ a.text }}</span>
        <span class="a-time">{{ a.time }}</span>
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

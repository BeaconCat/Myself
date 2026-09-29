<script setup lang="ts">
import { useIdentity } from '../../about/useIdentity';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';

/**
 * 首页「关于」小卡：头像 + 名字 + 座右铭 + 一行统计（运行天数 / 文章 / 随想 / 年提交），
 * 右侧社交图标 + 次按钮。中性卡片面，无发光；技能标签为极轻描边胶囊。
 */
const props = defineProps<{
  posts: number;
  notes: number;
  commits: number;
  loading?: boolean;
}>();

const { t } = useI18n();
const config = useConfigStore();
const me = useIdentity();
const about = computed(() => config.cfg.about);

const days = computed(() => {
  const start = Date.parse(`${about.value.foundedAt}T00:00:00`);
  if (Number.isNaN(start)) return 0;
  return Math.max(1, Math.floor((Date.now() - start) / 86400000) + 1);
});

const stats = computed(() => [
  { v: days.value, k: t('aboutKit.stats.days'), tip: t('home.since', { d: about.value.foundedAt.replaceAll('-', '.') }) },
  { v: props.posts, k: t('home.posts') },
  { v: props.notes, k: t('home.notes') },
  { v: props.commits, k: t('home.commits') },
]);

const ghUrl = computed(() => `https://github.com/${config.cfg.github.username}`);
</script>

<template>
  <section class="me-card">
    <header class="head">
      <span class="av"><img :src="me.avatar.value" alt="" draggable="false" /></span>
      <div class="who">
        <b>{{ me.fullName.value }}</b>
        <q v-if="about.motto">{{ about.motto }}</q>
      </div>
      <div class="soc">
        <a class="ib" :href="ghUrl" target="_blank" rel="noopener" :aria-label="t('home.github')">
          <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" /></svg>
        </a>
        <a class="ib" href="/feed" target="_blank" rel="noopener" :aria-label="t('home.rss')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 5a14 14 0 0 1 14 14M5 11a8 8 0 0 1 8 8" /><circle cx="6" cy="18" r="1.4" fill="currentColor" stroke="none" /></svg>
        </a>
      </div>
    </header>

    <div class="stats">
      <div v-for="s in stats" :key="s.k" class="stat" :title="(s as { tip?: string }).tip">
        <span v-if="loading" class="sk num-sk" />
        <strong v-else>{{ s.v }}</strong>
        <small>{{ s.k }}</small>
      </div>
    </div>

    <footer class="foot">
      <div v-if="about.skills.length" class="skills">
        <span v-for="s in about.skills" :key="s">{{ s }}</span>
      </div>
      <router-link to="/about" class="btn-2nd">
        {{ t('home.aboutMore') }}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
      </router-link>
    </footer>
  </section>
</template>

<style scoped lang="scss">
/* 紧凑三段：身份（大头像 + 名字格言 + 社交）→ 四个大号数字 → 标签与入口。中性卡面，无发光 */
.me-card {
  display: flex;
  flex-direction: column;
  gap: 22px;
  height: 100%;
  padding: var(--card-pad);
  border-radius: var(--card-r);
  background: var(--card-bg);
  box-shadow: var(--card-shadow);
}

.head {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 18px;
}

.av {
  width: 76px;
  height: 76px;
  border-radius: var(--r-lg);
  overflow: hidden;
  box-shadow: 0 0 0 1px var(--line);
  background: var(--fill);

  img { width: 100%; height: 100%; object-fit: cover; display: block; }
}

.who {
  min-width: 0;

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 28px;
    line-height: 1.2;
    font-weight: 700;
    color: var(--text);
  }

  q {
    display: block;
    margin-top: 6px;
    font-family: var(--font-serif);
    font-size: 15px;
    color: var(--text-2);
    quotes: '\300C' '\300D';
  }
}

.soc { display: flex; gap: 8px; align-self: start; }

.ib {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: var(--r-pill);
  background: var(--fill);
  color: var(--text-2);
  transition: background var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  svg { width: 18px; height: 18px; }
  &:hover { background: var(--fill-2); color: var(--text); transform: translateY(-1px); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

.stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--stat-gap);
  padding: var(--statbar-pad);
  border-radius: var(--statbar-r);
  background: var(--statbar-bg);
  box-shadow: var(--statbar-shadow);
}

.stat {
  padding: var(--stat-pad);
  min-width: 0;

  & + & { box-shadow: var(--stat-sep); }

  strong {
    display: block;
    font-family: var(--font-mono);
    font-size: 30px;
    line-height: 1.1;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--text);
  }

  small {
    display: block;
    margin-top: 6px;
    font-size: 12.5px;
    color: var(--text-3);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.num-sk { display: block; width: 56%; height: 33px; border-radius: var(--r-xs); }

.foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-top: auto;
}

.skills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;

  span {
    padding: 5px 12px;
    border-radius: var(--r-pill);
    font-size: 13px;
    color: var(--text-2);
    box-shadow: 0 0 0 1px var(--line-2) inset;
  }
}

.btn-2nd {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex: none;
  height: 38px;
  padding: 0 16px;
  border-radius: var(--r-pill);
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  background: var(--fill);
  transition: background var(--dur-fast), transform var(--dur-fast) var(--ease-out);

  svg { width: 16px; height: 16px; transition: transform var(--dur-fast) var(--ease-out); }
  &:hover { background: var(--fill-2); }
  &:hover svg { transform: translateX(2px); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

@media (max-width: 560px) {
  .stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .stat:nth-child(3) { box-shadow: 0 -1px 0 var(--line); }
  .stat:nth-child(4) { box-shadow: -1px 0 0 var(--line), 0 -1px 0 var(--line); }
}
/* 简洁风格：纯数字行，窄屏换行时不画格线 */
:root[data-style='clean'] .stat { box-shadow: none; }
</style>

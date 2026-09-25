<script setup lang="ts">
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
const about = computed(() => config.cfg.about);

const days = computed(() => {
  const start = Date.parse(`${about.value.foundedAt}T00:00:00`);
  if (Number.isNaN(start)) return 0;
  return Math.max(1, Math.floor((Date.now() - start) / 86400000) + 1);
});

const stats = computed(() => [
  { v: days.value, k: t('home.since', { d: about.value.foundedAt.replaceAll('-', '.') }) },
  { v: props.posts, k: t('home.posts') },
  { v: props.notes, k: t('home.notes') },
  { v: props.commits, k: t('home.commits') },
]);

const ghUrl = computed(() => `https://github.com/${config.cfg.github.username}`);
</script>

<template>
  <section class="me-card">
    <span class="av"><img :src="about.avatar || '/favicon-256.png'" alt="" draggable="false" /></span>

    <div class="who">
      <b>{{ about.name || config.cfg.site.title }}</b>
      <q v-if="about.motto">{{ about.motto }}</q>
      <div class="stats">
        <div v-for="s in stats" :key="s.k">
          <span v-if="loading" class="sk sk-line num-sk" />
          <strong v-else>{{ s.v }}</strong>
          <small>{{ s.k }}</small>
        </div>
      </div>
      <div v-if="about.skills.length" class="skills">
        <span v-for="s in about.skills" :key="s">{{ s }}</span>
      </div>
    </div>

    <div class="go">
      <div class="soc">
        <a class="ib" :href="ghUrl" target="_blank" rel="noopener" :aria-label="t('home.github')">
          <svg viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8Z" /></svg>
        </a>
        <a class="ib" href="/feed" target="_blank" :aria-label="t('home.rss')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M5 5a14 14 0 0 1 14 14M5 11a8 8 0 0 1 8 8" /><circle cx="6" cy="18" r="1.4" fill="currentColor" stroke="none" /></svg>
        </a>
      </div>
      <router-link to="/about" class="btn-2nd">
        {{ t('home.aboutMore') }}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
      </router-link>
    </div>
  </section>
</template>

<style scoped lang="scss">
.me-card {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  gap: 36px;
  align-items: center;
  padding: 30px 32px;
  border-radius: var(--r-lg);
  background: var(--elev);
  box-shadow: var(--shadow-card);
  overflow: hidden;
}

:root[data-mode='dark'] .me-card {
  background: linear-gradient(180deg, color-mix(in oklab, var(--surface) 90%, white), var(--surface) 70%);
}

.av {
  width: 72px;
  height: 72px;
  border-radius: var(--r-xl);
  overflow: hidden;
  flex: none;
  background: #060b16;
  box-shadow: 0 0 0 0.5px rgb(255 255 255 / 0.14);
  transition: transform var(--dur) var(--ease-spring);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
}

.me-card:hover .av { transform: rotate(-4deg) scale(1.04); }

.who {
  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 24px;
    font-weight: 700;
  }

  q {
    display: block;
    margin-top: 4px;
    font-family: var(--font-serif);
    font-size: 15px;
    color: var(--text-2);
    quotes: '「' '」';

    &::before,
    &::after { color: var(--text-3); }
  }
}

.stats {
  display: flex;
  flex-wrap: wrap;
  row-gap: 12px;
  margin-top: 18px;

  div {
    padding: 0 24px;
    box-shadow: inset 0.5px 0 0 var(--line-2);

    &:first-child {
      padding-left: 0;
      box-shadow: none;
    }
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

  small {
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

/* 技能：极轻描边胶囊（无色块） */
.skills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 16px;

  span {
    display: inline-flex;
    align-items: center;
    height: 26px;
    padding: 0 11px;
    border-radius: var(--r-pill);
    font-size: 12.5px;
    color: var(--text-2);
    box-shadow: inset 0 0 0 1px var(--line);
  }
}

.go {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
}

.soc {
  display: flex;
  gap: 6px;
}

.ib {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--text-2);
  background: var(--fill);
  transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  svg { width: 17px; height: 17px; }

  &:hover { color: var(--text); background: var(--fill-2); }
  &:active { transform: scale(0.94); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

/* 次按钮：中性填充 + 发丝描边 */
.btn-2nd {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 38px;
  padding: 0 18px;
  border-radius: var(--r-pill);
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  background: var(--fill-2);
  box-shadow: inset 0 0 0 0.5px var(--line);
  transition: background-color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

  svg {
    width: 15px;
    height: 15px;
    transition: transform var(--dur) var(--ease-spring);
  }

  &:hover { background: var(--fill-3); }
  &:hover svg { transform: translateX(2px); }
  &:active { transform: scale(0.97); }
  &:focus-visible { outline: none; box-shadow: var(--focus); }
}

:root[data-mode='light'] .btn-2nd {
  background: #fff;
  box-shadow: 0 0 0 0.5px var(--line-2), 0 1px 2px rgb(16 24 40 / 0.06);

  &:hover { background: color-mix(in oklab, var(--bg) 50%, white); }
}

@media (max-width: 1100px) {
  .me-card { grid-template-columns: auto minmax(0, 1fr); }

  .go {
    grid-column: 1 / -1;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    padding-top: 18px;
    box-shadow: inset 0 0.5px 0 var(--line);
  }
}

@media (max-width: 640px) {
  .me-card {
    grid-template-columns: 1fr;
    gap: 20px;
    padding: 24px 20px;
  }

  .stats div { padding: 0 14px; }
}
</style>

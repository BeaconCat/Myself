<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import { useThemeStore } from '../../stores/theme';
import UiIcon from '../ui/UiIcon.vue';

/** 桌面前台页脚：logo + 站点名 / 格言、导航、RSS / GitHub、版权 + 当前色盘 */
const { t } = useI18n();
const config = useConfigStore();
const theme = useThemeStore();

const siteName = computed(() => config.cfg.site.title || t('common.siteName'));
const motto = computed(() => config.cfg.about.motto || config.cfg.site.subtitle);
const owner = computed(() => config.cfg.about.name || siteName.value);
const year = new Date().getFullYear();
const github = computed(() => {
  const u = config.cfg.github.username;
  return u ? `https://github.com/${encodeURIComponent(u)}` : '';
});

const links = [
  { to: '/articles', key: 'nav.articles' },
  { to: '/thoughts', key: 'nav.thoughts' },
  { to: '/about', key: 'nav.about' },
];
</script>

<template>
  <footer class="site-footer">
    <div class="in">
      <router-link to="/" class="who">
        <img class="logo" src="/favicon-256.png" alt="" width="40" height="40" />
        <span>
          <b>{{ siteName }}</b>
          <small>{{ motto }}</small>
        </span>
      </router-link>

      <nav class="links" :aria-label="t('footer.nav')">
        <router-link v-for="l in links" :key="l.to" :to="l.to">{{ t(l.key) }}</router-link>
        <a href="/feed" target="_blank" rel="noopener"><UiIcon name="rss" class="s" />{{ t('footer.rss') }}</a>
        <a v-if="github" :href="github" target="_blank" rel="noopener"><UiIcon name="github" class="s" />{{ t('footer.github') }}</a>
      </nav>

      <div class="cp">
        <span>&copy; {{ year }} {{ owner }} · {{ t('footer.poweredBy') }}</span>
        <span class="mono">{{ theme.palette?.nameKey }}</span>
      </div>
    </div>
  </footer>
</template>

<style scoped lang="scss">
.site-footer {
  margin-top: 120px;
  padding: 40px 0 44px;
  box-shadow: inset 0 0.5px 0 var(--line);
  font-size: 13px;
  color: var(--text-3);
}

.in {
  width: min(1200px, calc(100% - 80px));
  margin-inline: auto;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 28px 40px;
  align-items: center;
}

.who {
  justify-self: start;
  display: flex;
  align-items: center;
  gap: 14px;
  margin-left: -4px;
  padding: 4px 10px 4px 4px;
  border-radius: var(--r-md);
  transition: background-color var(--dur-fast);

  &:hover { background: var(--fill); }

  b {
    display: block;
    font-family: var(--font-serif);
    font-size: 17px;
    font-weight: 700;
    color: var(--text);
  }

  small { font-size: 13px; }
}

.logo {
  width: 40px;
  height: 40px;
  border-radius: var(--r-sm);
  box-shadow: 0 0 0 0.5px var(--line);
}

.links {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-right: -12px;

  a {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 32px;
    padding: 0 12px;
    border-radius: var(--r-pill);
    color: var(--text-2);
    transition: background-color var(--dur-fast), color var(--dur-fast), transform var(--dur-fast) var(--ease-spring);

    &:hover { background: var(--fill-2); color: var(--text); }
    &:active { transform: scale(0.96); }
    &:focus-visible { outline: none; box-shadow: var(--focus); }
  }
}

.cp {
  grid-column: 1 / -1;
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding-top: 20px;
  box-shadow: inset 0 0.5px 0 var(--line);
}

.mono {
  font-family: var(--font-mono);
  font-size: 12px;
}

@media (max-width: 1100px) {
  .in { width: calc(100% - 64px); }
}

@media (max-width: 860px) {
  .in { grid-template-columns: 1fr; }
  .cp { flex-direction: column; }
}
</style>

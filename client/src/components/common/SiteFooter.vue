<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/config';
import { useThemeStore } from '../../stores/theme';
import UiIcon from '../ui/UiIcon.vue';
import { useIdentity } from '../../about/useIdentity';

/** 桌面前台页脚：站点身份（头像 + 名字（别名）+ 签名）、导航、RSS / GitHub、版权（网站名称）+ 当前色盘 */
const { t } = useI18n();
const config = useConfigStore();
const theme = useThemeStore();

const siteName = computed(() => config.cfg.site.title || t('common.siteName'));
const { avatar, hasAvatar, fullName: name, sign: tagline } = useIdentity();
const sign = computed(() => tagline.value || config.cfg.site.subtitle);
const year = new Date().getFullYear();
/** 引擎源码仓库（固定，不随站长配置变化） */
const REPO = 'BeaconCat/Myself';
const REPO_URL = `https://github.com/${REPO}`;

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
        <img class="logo" :class="{ face: hasAvatar }" :src="avatar" alt="" width="40" height="40" />
        <span>
          <b>{{ name }}</b>
          <small>{{ sign }}</small>
        </span>
      </router-link>

      <nav class="links" :aria-label="t('footer.nav')">
        <router-link v-for="l in links" :key="l.to" :to="l.to">{{ t(l.key) }}</router-link>
        <a href="/feed" target="_blank" rel="noopener"><UiIcon name="rss" class="s" />{{ t('footer.rss') }}</a>
        <a v-if="github" :href="github" target="_blank" rel="noopener"><UiIcon name="github" class="s" />{{ t('footer.github') }}</a>
      </nav>

      <div class="cp">
        <span class="pw">
          &copy; {{ year }} {{ siteName }} · {{ t('footer.poweredBy') }}
          <a class="repo" :href="REPO_URL" target="_blank" rel="noopener" :title="t('footer.repoTitle')">
            <UiIcon name="github" class="s" />{{ REPO }}
          </a>
        </span>
        <span class="mono">{{ theme.palette?.nameKey }}</span>
      </div>
    </div>
  </footer>
</template>

<style scoped lang="scss">
.site-footer {
  margin-top: 64px;
  padding: 32px 0 36px;
  box-shadow: inset 0 0.5px 0 var(--line);
  font-size: 13px;
  color: var(--text-3);
}

/* 简洁风格：页脚与内容之间留白更舒展 */
:root[data-style='clean'] .site-footer {
  margin-top: 120px;
  padding: 40px 0 44px;
}

.in {
  width: min(1200px, calc(100% - 80px));
  margin-inline: auto;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 20px 40px;
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
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 16px;
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

.pw { display: inline-flex; align-items: center; flex-wrap: wrap; gap: 6px; }

/* 源码仓库：图标 + 仓库名，悬停抬升为链接色 */
.repo {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 8px 2px 6px;
  border-radius: var(--r-pill);
  font-family: var(--font-mono);
  font-size: 12.5px;
  color: var(--text-2);
  box-shadow: inset 0 0 0 1px var(--line);
  transition: color var(--dur-fast), background var(--dur-fast), box-shadow var(--dur-fast);

  &:hover { color: var(--ink); background: var(--fill); box-shadow: inset 0 0 0 1px var(--line-2); }
}

/* 上传了头像：圆形头像（未上传时仍是方形站点 logo） */
.logo.face { border-radius: 50%; object-fit: cover; }
</style>

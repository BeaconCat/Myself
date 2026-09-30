import { computed } from 'vue';
import { useConfigStore } from '../stores/config';

/** 内置 logo（未在「身份」里上传站点 logo 时使用） */
export const BUILTIN_LOGO = '/favicon-256.png';
export const BUILTIN_LOGO_LARGE = '/logo-1024.webp';

/**
 * 站点 logo：后台「身份 · 形象」里上传，全站唯一来源（后台品牌、登录页、头像 / 形象图的回退、浏览器标签图标）。
 * large = 大尺寸位置（形象图回退）用的版本；未上传时各自回落到内置图。
 */
export function useSiteLogo() {
  const config = useConfigStore();
  const custom = computed(() => config.cfg.site.logo || '');
  return {
    logo: computed(() => custom.value || BUILTIN_LOGO),
    logoLarge: computed(() => custom.value || BUILTIN_LOGO_LARGE),
    hasLogo: computed(() => !!custom.value),
  };
}

/** 浏览器标签图标跟随站点 logo（index.html 里的内置图标只负责首帧） */
export function applyFavicon(url: string): void {
  const href = url || '';
  document.querySelectorAll<HTMLLinkElement>('link[rel="icon"]').forEach((link) => {
    if (!link.dataset.builtin) link.dataset.builtin = link.getAttribute('href') ?? '';
    link.href = href || link.dataset.builtin;
  });
}

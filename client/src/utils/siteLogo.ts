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

const ICON_SIZE = 64;
let faviconSeq = 0;

/**
 * 浏览器标签图标跟随站点 logo，并按全局圆角裁成圆角方块（index.html 里的内置图标只负责首帧）。
 * 圆角与站内 30–40px 的 logo 同比例：每 40px 对应 1 倍圆角基准（默认基准 10 → 64px 图标圆角 16px）。
 */
export function applyFavicon(url: string, radiusBase = 10): void {
  const links = [...document.querySelectorAll<HTMLLinkElement>('link[rel="icon"]')];
  links.forEach((link) => {
    if (!link.dataset.builtin) link.dataset.builtin = link.getAttribute('href') ?? '';
  });
  const src = url || links.at(-1)?.dataset.builtin || BUILTIN_LOGO;
  const seq = ++faviconSeq;
  const img = new Image();
  img.onload = () => {
    if (seq !== faviconSeq) return;
    const canvas = document.createElement('canvas');
    canvas.width = canvas.height = ICON_SIZE;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    const r = Math.max(0, Math.min(ICON_SIZE / 2, (radiusBase * ICON_SIZE) / 40));
    ctx.beginPath();
    ctx.roundRect(0, 0, ICON_SIZE, ICON_SIZE, r);
    ctx.clip();
    // 按「铺满」裁切成正方形
    const s = Math.min(img.naturalWidth, img.naturalHeight);
    ctx.drawImage(img, (img.naturalWidth - s) / 2, (img.naturalHeight - s) / 2, s, s, 0, 0, ICON_SIZE, ICON_SIZE);
    let href = '';
    try {
      href = canvas.toDataURL('image/png');
    } catch {
      href = src;
    }
    links.forEach((link) => {
      link.href = href;
    });
  };
  img.onerror = () => {
    if (seq === faviconSeq) links.forEach((link) => { link.href = src; });
  };
  img.src = src;
}

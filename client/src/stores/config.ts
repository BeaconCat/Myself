import { defineStore } from 'pinia';

export interface ThemePreset {
  id: string;
  name: string;
  primary: string;
  primaryDeep: string;
}

export interface SiteConfig {
  site: { title: string; subtitle: string };
  loading: { bootText: string; routeText: string };
  theme: {
    defaultPaletteId: string;
    defaultMode: 'light' | 'dark';
    autoSwitch: 'off' | 'season';
    allowUserPalette: boolean;
    displayCount: number;
    presets: ThemePreset[];
  };
  timezone: string;
  github: {
    username: string;
    stats: { repos: number; stars: number; followers: number; commits: number };
  };
  about: { name: string; tagline: string; bio: string; skills: string[] };
}

export const FALLBACK_CONFIG: SiteConfig = {
  site: { title: 'Myself', subtitle: '个人博客' },
  loading: { bootText: 'Myself', routeText: '加载中' },
  theme: {
    defaultPaletteId: 'summer',
    defaultMode: 'dark',
    autoSwitch: 'off',
    allowUserPalette: true,
    displayCount: 4,
    presets: [
      { id: 'spring', name: '春 · 新绿', primary: '#00c853', primaryDeep: '#00a344' },
      { id: 'summer', name: '夏 · 炽红', primary: '#ff0032', primaryDeep: '#d40029' },
      { id: 'autumn', name: '秋 · 暖阳', primary: '#ffb300', primaryDeep: '#e09600' },
      { id: 'winter', name: '冬 · 霜蓝', primary: '#0078ff', primaryDeep: '#005fd6' },
    ],
  },
  timezone: 'Asia/Shanghai',
  github: { username: 'your-github', stats: { repos: 0, stars: 0, followers: 0, commits: 0 } },
  about: {
    name: 'Myself',
    tagline: '开源个人博客引擎',
    bio: '这里是 Myself 的默认介绍，可在后台「设置 → 关于信息」修改。',
    skills: ['写作', '摄影', '编程'],
  },
};

export const useConfigStore = defineStore('config', {
  state: () => ({
    cfg: FALLBACK_CONFIG as SiteConfig,
    loaded: false,
  }),
  getters: {
    /** 展示给访客的色盘（前 displayCount 个） */
    visiblePresets: (s) =>
      s.cfg.theme.presets.slice(0, Math.max(1, s.cfg.theme.displayCount)),
  },
  actions: {
    async load() {
      try {
        const res = await fetch('/api/v1/site-config');
        if (res.ok) {
          this.cfg = (await res.json()) as SiteConfig;
        }
      } catch { /* 后端未启动时用回退配置 */ }
      this.loaded = true;
      document.title = this.cfg.site.title;
    },
  },
});

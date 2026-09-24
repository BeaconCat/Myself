import { defineStore } from 'pinia';
import { api } from '../api';
import type { CardChoreoId, TextChoreoId } from '../components/home/hero/choreo/types';
import DEFAULT_ABOUT_MODULES from '../about/default-modules.json';

/**
 * 关于页可排序模块（data 结构由模块注册表约定，见 about/types.ts）。
 * span：1|2|3 = 12 栏 bento 中占 4/8/12 栏；variant：模块变体；title：覆盖默认标题；hidden：前台不渲染。
 */
export interface AboutModule {
  id: string;
  type: string;
  span?: 1 | 2 | 3;
  variant?: string;
  title?: string;
  hidden?: boolean;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any;
}

export interface ThemePreset {
  id: string;
  name: string;
  primary: string;
  primaryDeep: string;
}

export interface SiteConfig {
  site: { title: string; subtitle: string; listEndText: string };
  loading: { bootText: string; routeText: string };
  theme: {
    defaultPaletteId: string;
    defaultMode: 'light' | 'dark';
    autoSwitch: 'off' | 'season';
    allowUserPalette: boolean;
    displayCount: number;
    presets: ThemePreset[];
  };
  hero: {
    intervalMs: number;
    count: number;
    pinnedRule: 'pinned-first' | 'ignore';
    /** 左侧文字切换动效（独立于卡组，可自由搭配） */
    textAnim: TextChoreoId;
    /** 右侧卡组切换动效 */
    cardAnim: CardChoreoId;
  };
  thoughts: { subtitle: string };
  covers: { expandMs: number };
  timezone: string;
  github: {
    username: string;
    mode: 'manual' | 'api';
    token?: string;
    refreshMinutes: number;
    proxy: string;
    insecureTls: boolean;
    stats: { repos: number; stars: number; followers: number; commits: number };
  };
  about: {
    avatar: string;
    name: string;
    tagline: string;
    bio: string;
    skills: string[];
    foundedAt: string;
    motto: string;
    modules: AboutModule[];
  };
}

export const FALLBACK_CONFIG: SiteConfig = {
  site: { title: 'Myself', subtitle: '个人博客', listEndText: '—— 到底啦 ——' },
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
  hero: { intervalMs: 3000, count: 4, pinnedRule: 'pinned-first', textAnim: 'lightscan', cardAnim: 'hinge' },
  thoughts: { subtitle: '碎片化的想法、心情与瞬间，短到装不下一篇文章。' },
  covers: { expandMs: 10000 },
  timezone: 'Asia/Shanghai',
  github: {
    username: 'your-github',
    mode: 'manual',
    refreshMinutes: 30,
    proxy: '',
    insecureTls: false,
    stats: { repos: 0, stars: 0, followers: 0, commits: 0 },
  },
  about: {
    avatar: '',
    name: 'Myself',
    tagline: '开源个人博客引擎',
    bio: '这里是 Myself 的默认介绍，可在后台「关于管理」修改。',
    skills: ['写作', '摄影', '编程'],
    foundedAt: '2026-01-01',
    motto: '记录本身，就是意义。',
    modules: DEFAULT_ABOUT_MODULES as AboutModule[],
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
        this.cfg = await api.siteConfig<SiteConfig>();
      } catch { /* 后端未启动时用回退配置 */ }
      this.loaded = true;
      document.title = this.cfg.site.title;
    },
  },
});

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
  hero: {
    intervalMs: number;
    count: number;
    pinnedRule: 'pinned-first' | 'ignore';
  };
  thoughts: { subtitle: string };
  timezone: string;
  github: {
    username: string;
    mode: 'manual' | 'api';
    token?: string;
    refreshMinutes: number;
    stats: { repos: number; stars: number; followers: number; commits: number };
  };
  about: {
    avatar: string;
    name: string;
    tagline: string;
    bio: string;
    skills: string[];
    foundedAt: string;
    skillGroups: { title: string; items: string[] }[];
    milestones: { year: string; text: string }[];
    socials: { name: string; url: string; icon: string }[];
    motto: string;
  };
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
  hero: { intervalMs: 3000, count: 4, pinnedRule: 'pinned-first' },
  thoughts: { subtitle: '碎片化的想法、心情与瞬间，短到装不下一篇文章。' },
  timezone: 'Asia/Shanghai',
  github: {
    username: 'your-github',
    mode: 'manual',
    refreshMinutes: 30,
    stats: { repos: 0, stars: 0, followers: 0, commits: 0 },
  },
  about: {
    avatar: '',
    name: 'Myself',
    tagline: '开源个人博客引擎',
    bio: '这里是 Myself 的默认介绍，可在后台「关于管理」修改。',
    skills: ['写作', '摄影', '编程'],
    foundedAt: '2026-01-01',
    skillGroups: [
      { title: '创作', items: ['文章', '随想', '摄影'] },
      { title: '工具', items: ['Markdown', '主题系统', 'API 中心'] },
      { title: '兴趣', items: ['阅读', '旅行', '音乐'] },
    ],
    milestones: [
      { year: '01', text: '在后台「设置」里换上你的名字、简介与主题色' },
      { year: '02', text: '发布第一篇文章，或用随想记录此刻' },
      { year: '03', text: '创建 APIKey，把日常发文托管给你的 AI 助手' },
    ],
    socials: [
      { name: 'GitHub', url: 'https://github.com/your-github', icon: 'github' },
      { name: 'Email', url: 'mailto:hi@example.com', icon: 'mail' },
      { name: 'RSS', url: '/feed', icon: 'rss' },
    ],
    motto: '记录本身，就是意义。',
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

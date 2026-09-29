import { defineStore } from 'pinia';
import { api } from '../api';
import type { CardChoreoId, RotateChoreoId, TextChoreoId } from '../components/home/hero/choreo/types';
import DEFAULT_ABOUT_MODULES from '../about/default-modules.json';
import { normalizeIdentity, type Identity } from '../about/identity';

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
  /** 站点尚未初始化（首次启动）；由 /site-config 下发 */
  needsSetup?: boolean;
  /** 用户系统开关。公开配置里是生效后的子集；后台设置接口返回完整字段 */
  users?: UsersConfig;
  /** 发信（仅后台设置接口返回） */
  mail?: MailConfig;
  /** 第三方登录（仅后台设置接口返回） */
  oauth?: { github: { clientId: string; clientSecret: string } };
  site: { title: string; subtitle: string; listEndText: string };
  loading: { bootText: string; routeText: string };
  theme: {
    defaultPaletteId: string;
    defaultMode: 'light' | 'dark';
    /** 默认界面风格：cards = 高密度卡片；clean = 透明背景简洁版式 */
    defaultStyle: 'cards' | 'clean';
    autoSwitch: 'off' | 'season';
    allowUserPalette: boolean;
    /** 是否允许访客自行切换界面风格 */
    allowUserStyle: boolean;
    displayCount: number;
    presets: ThemePreset[];
    /** 全局圆角基准（px，0–24，默认 10）；前台、移动端、后台共用 */
    radius: number;
  };
  hero: {
    intervalMs: number;
    count: number;
    pinnedRule: 'pinned-first' | 'ignore';
    /** 左侧文字切换动效（独立于卡组，可自由搭配） */
    textAnim: TextChoreoId;
    /** 右侧卡组切换动效 */
    cardAnim: CardChoreoId;
    /** 组内多封面的轮转动效 */
    rotateAnim: RotateChoreoId;
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
  /** 站点身份字段见 about/identity.ts（全站唯一来源）；modules 为关于页模块 */
  about: Identity & {
    skills: string[];
    modules: AboutModule[];
  };
}

export interface UsersConfig {
  enabled: boolean;
  readers: { enabled: boolean; signup: 'open' | 'invite' | 'closed'; requireVerify?: boolean };
  authors: { enabled: boolean; directPublish?: boolean };
  comments: { enabled: boolean; anonymous: boolean; moderation: 'all' | 'first' | 'none' };
  /** 公开配置：github = GitHub 登录可用；mailReset = 可用邮件找回密码 */
  login: { github: boolean; mailReset?: boolean };
  /** 访客回应（喜欢 / 灵感 / 会心 / 共鸣），缺省开启 */
  reactions?: boolean;
}

export interface MailConfig {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  password: string;
  from: string;
  security: 'starttls' | 'tls' | 'none';
}

export const FALLBACK_CONFIG: SiteConfig = {
  site: { title: 'Myself', subtitle: '个人博客', listEndText: '—— 到底啦 ——' },
  loading: { bootText: 'Myself', routeText: '加载中' },
  theme: {
    defaultPaletteId: 'summer',
    defaultMode: 'dark',
    defaultStyle: 'clean',
    autoSwitch: 'off',
    allowUserPalette: true,
    allowUserStyle: true,
    displayCount: 4,
    radius: 10,
    presets: [
      { id: 'spring', name: '春 · 新绿', primary: '#00c853', primaryDeep: '#00a344' },
      { id: 'summer', name: '夏 · 炽红', primary: '#ff0032', primaryDeep: '#d40029' },
      { id: 'autumn', name: '秋 · 暖阳', primary: '#ffb300', primaryDeep: '#e09600' },
      { id: 'winter', name: '冬 · 霜蓝', primary: '#0078ff', primaryDeep: '#005fd6' },
    ],
  },
  hero: { intervalMs: 3000, count: 4, pinnedRule: 'pinned-first', textAnim: 'lightscan', cardAnim: 'hinge', rotateAnim: 'lift' },
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
    alias: '',
    hello: '你好，我是',
    tagline: '开源个人博客引擎',
    bio: '这里是 Myself 的默认介绍。前往后台「身份」写下你自己的故事：你是谁、在做什么、热爱什么。',
    skills: ['写作', '摄影', '编程'],
    foundedAt: '2026-01-01',
    motto: '记录本身，就是意义。',
    mottoSign: '',
    status: { doing: '', city: '', tz: 8 },
    links: [
      { name: 'GitHub', handle: '@your-github', icon: 'github', url: 'https://github.com/your-github', primary: true },
      { name: '邮件', handle: 'hi@example.com', icon: 'mail', url: 'mailto:hi@example.com' },
      { name: 'RSS', handle: '/feed', icon: 'rss', url: '/feed' },
    ],
    portrait: { src: '', fade: 'left' },
    banner: { show: false, src: '' },
    modules: DEFAULT_ABOUT_MODULES as AboutModule[],
  },
};

let markLoaded: () => void = () => undefined;
/** 首次载入完成（成功或回退）即兑现；路由守卫据此判断是否需要初始化 */
export const configLoaded = new Promise<void>((resolve) => { markLoaded = resolve; });

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
        const cfg = await api.siteConfig<SiteConfig>();
        normalizeIdentity(cfg.about);
        this.cfg = cfg;
      } catch { /* 后端未启动时用回退配置 */ }
      this.loaded = true;
      markLoaded();
      document.title = this.cfg.site.title;
    },
  },
});

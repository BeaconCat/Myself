import { defineStore } from 'pinia';
import { api } from '../api';
import type { CardChoreoId, RotateChoreoId, TextChoreoId } from '../components/home/hero/choreo/types';
import DEFAULT_ABOUT_MODULES from '../about/default-modules.json';
import { normalizeIdentity, type Identity } from '../about/identity';
import { applyFavicon } from '../utils/siteLogo';

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
  /** 登录保持时长（仅后台设置接口返回） */
  session?: { duration: SessionDuration };
  /** 第三方登录（仅后台设置接口返回） */
  oauth?: { github: { clientId: string; clientSecret: string } };
  /** url：站点对外地址（邮件链接、RSS、第三方登录回调用；留空时取当前访问地址） */
  site: { title: string; subtitle: string; listEndText: string; url?: string; logo?: string };
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
  /** 全站动画：route = 切页动效档位 */
  /** shockwave：色盘 / 界面风格切换时扩散圆上的冲击波强调边（默认关闭） */
  motion?: { route: RouteMotion; shockwave?: boolean };
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
  thoughts: { subtitle: string; showAlias: boolean; showUsername: boolean };
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

/**
 * 切页动效：rich = 每次换页都播全屏遮罩；standard = 只在大栏目之间（首页 / 文章 / 随想 / 关于）播全屏遮罩，
 * 栏目内（列表 ↔ 详情等）用轻量渐入；minimal = 从不播切页遮罩，只保留首次载入的全屏揭幕。
 */
export type RouteMotion = 'rich' | 'standard' | 'minimal';

/** 登录保持时长：1 天 / 7 天 / 30 天 / 1 年 / 永久（访问时续期） */
export type SessionDuration = '1d' | '7d' | '30d' | '1y' | 'forever';

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
  motion: { route: 'standard', shockwave: false },
  hero: { intervalMs: 3000, count: 4, pinnedRule: 'pinned-first', textAnim: 'lightscan', cardAnim: 'hinge', rotateAnim: 'lift' },
  thoughts: { subtitle: '碎片化的想法、心情与瞬间，短到装不下一篇文章。', showAlias: true, showUsername: true },
  covers: { expandMs: 10000 },
  timezone: 'Asia/Shanghai',
  github: {
    username: '',
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
      { name: 'RSS', handle: '/feed', icon: 'rss', url: '/feed', primary: true },
    ],
    portrait: { src: '', fade: 'left' },
    banner: { show: false, src: '' },
    modules: DEFAULT_ABOUT_MODULES as AboutModule[],
  },
};

/* ---------- 站点配置载入失败时的退避重试（1s → 2s → 4s … 封顶 15s；恢复联网 / 回到页面时立即重试） ---------- */
let retryTimer = 0;
let retryDelay = 1000;
let retryFn: (() => void) | null = null;
const retryNow = (): void => {
  if (!retryFn) return;
  window.clearTimeout(retryTimer);
  retryFn();
};

function scheduleRetry(fn: () => void): void {
  if (!retryFn) {
    window.addEventListener('online', retryNow);
    window.addEventListener('focus', retryNow);
  }
  retryFn = fn;
  window.clearTimeout(retryTimer);
  retryTimer = window.setTimeout(fn, retryDelay);
  retryDelay = Math.min(15000, retryDelay * 2);
}

function stopRetry(): void {
  if (!retryFn) return;
  window.clearTimeout(retryTimer);
  window.removeEventListener('online', retryNow);
  window.removeEventListener('focus', retryNow);
  retryFn = null;
  retryDelay = 1000;
}

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
        cfg.thoughts = { ...FALLBACK_CONFIG.thoughts, ...cfg.thoughts };
        this.cfg = cfg;
        stopRetry();
      } catch {
        // 后端暂时不可用（重启、网络抖动）：先用回退配置渲染，后台退避重试，拿到后替换，
        // 避免整页一直停在出厂默认内容上
        scheduleRetry(() => this.load());
      }
      this.loaded = true;
      markLoaded();
      document.title = this.cfg.site.title;
      applyFavicon(this.cfg.site.logo ?? '', this.cfg.theme.radius ?? 10);
    },
  },
});

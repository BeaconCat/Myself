import { db } from './db.js';

/**
 * 站点配置：settings 表中 key = 'site_config'，整体 JSON。
 * 公开子集经 /site-config 下发给前台。
 */
export const DEFAULT_CONFIG = {
  site: {
    title: 'Myself',
    subtitle: '个人博客',
  },
  loading: {
    bootText: 'Myself',
    routeText: '加载中',
  },
  theme: {
    defaultPaletteId: 'summer',
    defaultMode: 'dark',
    /** off | season（按月份春夏秋冬自动） */
    autoSwitch: 'off',
    /** 是否展示色盘切换条；关闭则只留深浅切换 */
    allowUserPalette: true,
    /** 预设最多 10 组，前 displayCount 个展示 */
    displayCount: 4,
    presets: [
      { id: 'spring', name: '春 · 新绿', primary: '#00c853', primaryDeep: '#00a344' },
      { id: 'summer', name: '夏 · 炽红', primary: '#ff0032', primaryDeep: '#d40029' },
      { id: 'autumn', name: '秋 · 暖阳', primary: '#ffb300', primaryDeep: '#e09600' },
      { id: 'winter', name: '冬 · 霜蓝', primary: '#0078ff', primaryDeep: '#005fd6' },
    ],
  },
  hero: {
    /** 每张照片停留毫秒 */
    intervalMs: 3000,
    /** 取最新 n 条 */
    count: 4,
    /** pinned-first = 置顶优先；ignore = 无视置顶按规则 */
    pinnedRule: 'pinned-first',
  },
  thoughts: {
    subtitle: '碎片化的想法、心情与瞬间，短到装不下一篇文章。',
  },
  timezone: 'Asia/Shanghai',
  github: {
    username: 'your-github',
    /** manual = 手填统计；api = 服务端经 GitHub API 拉取 */
    mode: 'manual',
    /** 可选只读 PAT：提升配额；留空走匿名公开接口 */
    token: '',
    /** api 模式缓存刷新间隔（分钟） */
    refreshMinutes: 30,
    /** 出站代理，如 http://127.0.0.1:7890；留空自动读 HTTPS_PROXY 环境变量 */
    proxy: '',
    /** 跳过 TLS 校验（仅 GitHub 只读拉取）：本机代理/安全软件注入证书时的兜底 */
    insecureTls: false,
    stats: { repos: 0, stars: 0, followers: 0, commits: 0 },
  },
  about: {
    /** 头像 URL（素材库上传），留空用站点 logo */
    avatar: '',
    name: 'Myself',
    tagline: '开源个人博客引擎',
    bio: '这里是 Myself 的默认介绍。前往后台「关于管理」写下你自己的故事：你是谁、在做什么、热爱什么。',
    skills: ['写作', '摄影', '编程'],
    /** 建站日期：关于页展示运行天数 */
    foundedAt: '2026-01-01',
    /** 关于页格言（身份卡下方引用块） */
    motto: '记录本身，就是意义。',
    /**
     * 模块化区块（有序，身份卡固定在最前不入列）。
     * 每项 { id, type, data }，type 见前端模块注册表。
     */
    modules: [
      { id: 'm-stats', type: 'stats', data: {} },
      {
        id: 'm-skills',
        type: 'skills',
        data: {
          groups: [
            { title: '创作', items: ['文章', '随想', '摄影'] },
            { title: '工具', items: ['Markdown', '主题系统', 'API 中心'] },
            { title: '兴趣', items: ['阅读', '旅行', '音乐'] },
          ],
        },
      },
      {
        id: 'm-milestones',
        type: 'milestones',
        data: {
          items: [
            { year: '01', text: '在后台「设置」里换上你的名字、简介与主题色' },
            { year: '02', text: '发布第一篇文章，或用随想记录此刻' },
            { year: '03', text: '创建 APIKey，把日常发文托管给你的 AI 助手' },
          ],
        },
      },
      { id: 'm-github', type: 'github', data: {} },
      {
        id: 'm-socials',
        type: 'socials',
        data: {
          items: [
            { name: 'GitHub', url: 'https://github.com/your-github', icon: 'github' },
            { name: 'Email', url: 'mailto:hi@example.com', icon: 'mail' },
            { name: 'RSS', url: '/feed', icon: 'rss' },
          ],
        },
      },
      {
        id: 'm-stack',
        type: 'stack',
        data: {
          items: [
            { name: 'Vue 3', role: '前端框架' },
            { name: 'Vite', role: '构建工具' },
            { name: 'Express', role: 'API 服务' },
            { name: 'SQLite', role: '数据存储' },
            { name: 'Markdown', role: '内容规范' },
          ],
        },
      },
    ],
  },
  backup: {
    /** 自动备份间隔小时数，0 = 关闭 */
    autoHours: 0,
  },
};

function deepMerge(base, patch) {
  if (Array.isArray(base) || Array.isArray(patch) || typeof base !== 'object' || base === null) {
    return patch === undefined ? base : patch;
  }
  const out = { ...base };
  for (const key of Object.keys(patch ?? {})) {
    out[key] = deepMerge(base[key], patch[key]);
  }
  return out;
}

/** 旧版扁平 about 字段 → 模块化迁移 */
function migrateAboutModules(cfg) {
  const about = cfg.about ?? {};
  if (Array.isArray(about.modules) && about.modules.length) return cfg;
  const defaults = structuredClone(DEFAULT_CONFIG.about.modules);
  for (const mod of defaults) {
    if (mod.type === 'skills' && Array.isArray(about.skillGroups)) {
      mod.data.groups = about.skillGroups;
    }
    if (mod.type === 'milestones' && Array.isArray(about.milestones)) {
      mod.data.items = about.milestones;
    }
    if (mod.type === 'socials' && Array.isArray(about.socials)) {
      mod.data.items = about.socials;
    }
  }
  about.modules = defaults;
  return cfg;
}

export function getConfig() {
  const raw = db.prepare(`SELECT value FROM settings WHERE key = 'site_config'`).get()?.value;
  if (!raw) return structuredClone(DEFAULT_CONFIG);
  try {
    return migrateAboutModules(deepMerge(DEFAULT_CONFIG, JSON.parse(raw)));
  } catch {
    return structuredClone(DEFAULT_CONFIG);
  }
}

export function saveConfig(patch) {
  const merged = deepMerge(getConfig(), patch);
  db.prepare(`
    INSERT INTO settings (key, value) VALUES ('site_config', ?)
    ON CONFLICT(key) DO UPDATE SET value = excluded.value
  `).run(JSON.stringify(merged));
  return merged;
}

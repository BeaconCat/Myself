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
  timezone: 'Asia/Shanghai',
  github: {
    username: 'BeaconCat',
    stats: { repos: 0, stars: 0, followers: 0, commits: 0 },
  },
  about: {
    name: 'BeaconCat',
    tagline: '写代码的灯塔看守人',
    bio: '白天造界面，晚上写随想。相信简约的惊艳、克制的动效，和 Markdown 能装下的一切。',
    skills: ['Vue', 'TypeScript', 'Node.js'],
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

export function getConfig() {
  const raw = db.prepare(`SELECT value FROM settings WHERE key = 'site_config'`).get()?.value;
  if (!raw) return structuredClone(DEFAULT_CONFIG);
  try {
    return deepMerge(DEFAULT_CONFIG, JSON.parse(raw));
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

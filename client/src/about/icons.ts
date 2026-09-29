/**
 * 关于页线性图标集（24 网格，1.5px stroke）。内容为内部常量，经 KitIcon 以 innerHTML 渲染。
 * 名称同时是 socials / uses 数据里 icon 字段的取值。
 */
export const ICONS: Record<string, string> = {
  github: '<path d="M9 19c-4 1.5-4-2.5-6-3m12 5v-3.5c0-1 .1-1.4-.5-2 2.8-.3 5.5-1.4 5.5-6a4.6 4.6 0 0 0-1.3-3.2 4.2 4.2 0 0 0-.1-3.2s-1.1-.3-3.5 1.3a12.3 12.3 0 0 0-6.2 0C6.5 2.8 5.4 3.1 5.4 3.1a4.2 4.2 0 0 0-.1 3.2A4.6 4.6 0 0 0 4 9.5c0 4.6 2.7 5.7 5.5 6-.6.6-.6 1.2-.5 2V21"/>',
  mail: '<rect x="3" y="5" width="18" height="14" rx="3"/><path d="m4 7 8 6 8-6"/>',
  rss: '<path d="M5 11a8 8 0 0 1 8 8M5 5a14 14 0 0 1 14 14"/><circle cx="6" cy="18" r="1.2"/>',
  x: '<path d="M5 4l14 16M19 4 5 20"/>',
  telegram: '<path d="M21 4 3 11l6 2 2 6 3-4 5 4 2-15Z"/><path d="m9 13 8-6"/>',
  weibo: '<path d="M10 19c-4 0-7-2-7-4.8C3 10.8 7.6 7 11 7c1.8 0 2 1.2 1.6 2.6 1.8-.8 4.4-.8 4.4 1.2 0 .5-.1.9-.4 1.3 1.3.4 2.4 1.2 2.4 2.5C19 17 14.8 19 10 19Z"/><path d="M16 4a5 5 0 0 1 5 5"/>',
  bilibili: '<rect x="3" y="6" width="18" height="13" rx="3"/><path d="m8 3 2 3M16 3l-2 3M9 12v1.5M15 12v1.5"/>',
  link: '<path d="M10 14a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1.5 1.5M14 10a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7L12.5 19"/>',
  arrow: '<path d="M7 17 17 7M9 7h8v8"/>',
  star: '<path d="m12 3 2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1L3.2 9.5l6.1-.9Z"/>',
  copy: '<rect x="8" y="8" width="12" height="12" rx="3"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/>',
  ok: '<path d="m5 12 5 5 9-10"/>',
  play: '<path d="M8 5v14l11-7Z" fill="currentColor"/>',
  pause: '<path d="M8 5v14M16 5v14" stroke-width="3"/>',
  prev: '<path d="M18 6 9 12l9 6V6ZM6 6v12"/>',
  next: '<path d="m6 6 9 6-9 6V6ZM18 6v12"/>',
  left: '<path d="m15 6-6 6 6 6"/>',
  right: '<path d="m9 6 6 6-6 6"/>',
  close: '<path d="M6 6l12 12M18 6 6 18"/>',
  heart: '<path d="M12 20s-7-4.4-9-8.6C1.6 8.3 3.4 5 6.6 5 8.6 5 10 6 12 8c2-2 3.4-3 5.4-3 3.2 0 5 3.3 3.6 6.4C19 15.6 12 20 12 20Z"/>',
  send: '<path d="M4 12 20 4l-6 16-3-7-7-1Z"/>',
  laptop: '<rect x="4" y="5" width="16" height="11" rx="2"/><path d="M2 19h20"/>',
  keyboard: '<rect x="2" y="6" width="20" height="12" rx="3"/><path d="M6 10h.01M10 10h.01M14 10h.01M18 10h.01M7 14h10"/>',
  camera: '<path d="M4 8h3l2-3h6l2 3h3v11H4Z"/><circle cx="12" cy="13" r="3.5"/>',
  monitor: '<rect x="3" y="4" width="18" height="12" rx="2"/><path d="M9 20h6M12 16v4"/>',
  code: '<path d="m8 7-5 5 5 5M16 7l5 5-5 5"/>',
  pen: '<path d="M4 20h4L19 9l-4-4L4 16Z"/>',
  figma: '<circle cx="15" cy="12" r="3"/><path d="M9 3h3v6H9a3 3 0 0 1 0-6ZM9 9h3v6H9a3 3 0 0 1 0-6ZM12 3h3a3 3 0 0 1 0 6h-3ZM9 15h3v3a3 3 0 1 1-3-3Z"/>',
  command: '<path d="M9 6a3 3 0 1 0-3 3h12a3 3 0 1 0-3-3v12a3 3 0 1 0 3-3H6a3 3 0 1 0 3 3Z"/>',
  cloud: '<path d="M7 18a4 4 0 0 1-.5-8A6 6 0 0 1 18 9a4.5 4.5 0 0 1-.5 9Z"/>',
  terminal: '<rect x="3" y="4" width="18" height="16" rx="3"/><path d="m7 9 3 3-3 3M13 15h4"/>',
  headphones: '<path d="M4 15v-3a8 8 0 0 1 16 0v3"/><rect x="3" y="14" width="4" height="6" rx="1.5"/><rect x="17" y="14" width="4" height="6" rx="1.5"/>',
  phone: '<rect x="6" y="2.5" width="12" height="19" rx="3"/><path d="M11 18.5h2"/>',
  book: '<path d="M4 5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2Z"/><path d="M4 19V5M8 7h7"/>',
  music: '<path d="M9 18V5l11-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="17" cy="16" r="3"/>',
  cycle: '<path d="M4 12a8 8 0 0 1 14-5l2 2M20 12a8 8 0 0 1-14 5l-2-2M20 4v5h-5M4 20v-5h5"/>',
};

/** 可供后台选择的图标名（社交 / 工作台） */
export const SOCIAL_ICONS = ['github', 'mail', 'rss', 'x', 'telegram', 'weibo', 'bilibili', 'link'] as const;
export const USES_ICONS = ['laptop', 'monitor', 'keyboard', 'camera', 'headphones', 'phone', 'code', 'terminal', 'figma', 'command', 'cloud', 'pen', 'book', 'music', 'link'] as const;

/** 无真实照片时的默认图：12 幅默认封面（client/public/covers/NN.webp），后台下拉按名字选 */
export const SCENES = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12'] as const;
export const SCENE_LABELS: Record<string, string> = {
  '01': '地平线', '02': '月相', '03': '叠纸', '04': '拱窗', '05': '潮汐', '06': '格物',
  '07': '光斑', '08': '山影', '09': '书脊', '10': '圆舞', '11': '雨线', '12': '远帆',
};
/** 旧版 CSS 光影场景名 → 相近的默认图（读取旧数据时兼容） */
export const LEGACY_SCENES: Record<string, string> = {
  door: '05', window: '04', sea: '12', tunnel: '02', dusk: '01', snow: '08', page: '03', grid: '06',
};

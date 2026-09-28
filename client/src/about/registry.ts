import type { AboutModule } from '../stores/config';
import type { Span } from './types';

/**
 * 关于页模块注册表：模块选择器、后台编辑器、前台渲染器的唯一元数据源。
 * - icon：内联 SVG path（24 viewBox，1.5px stroke 风格）
 * - spans：允许的宽度（1/2/3 = 4/8/12 栏），defaultSpan 为新增时的默认值
 * - variants：变体，第一项为默认
 * - chrome：该变体是否使用统一卡片外壳（profile / chapter / principles / motto 收尾为开放排版）
 */
export interface ModuleVariant { id: string; label: string }

export type ModuleGroup = '身份' | '数据' | '经历' | '喜好' | '工具' | '互动';

export interface ModuleMeta {
  type: string;
  name: string;
  desc: string;
  icon: string;
  group: ModuleGroup;
  variants: ModuleVariant[];
  spans: Span[];
  defaultSpan: Span;
  chrome: (variant: string) => boolean;
  /** 新增时的默认 data */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  defaultData: () => any;
  /** 管理列表行摘要 */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  summary: (data: any) => string;
}

const yes = () => true;
const no = () => false;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const count = (arr: any, unit: string) => `${Array.isArray(arr) ? arr.length : 0} ${unit}`;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const names = (arr: any, key = 'name') => (Array.isArray(arr) ? arr.map((x) => x?.[key]).filter(Boolean).join(' · ') : '');

export const MODULE_GROUPS: ModuleGroup[] = ['身份', '数据', '经历', '喜好', '工具', '互动'];

export const MODULE_REGISTRY: ModuleMeta[] = [
  /* ---------------- 身份 ---------------- */
  {
    type: 'profile',
    name: '身份区',
    desc: '大字宋体名字、一句话、自述、实时状态行与社交入口；右侧形象图全尺寸显示、左缘渐隐',
    icon: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM4 21a8 8 0 0 1 16 0',
    group: '身份',
    variants: [{ id: 'portrait', label: '形象图' }, { id: 'plain', label: '纯文字' }],
    spans: [3],
    defaultSpan: 3,
    chrome: no,
    defaultData: () => ({
      hello: '你好，我是',
      name: '你的名字',
      lede: '一句话介绍你自己，*星号* 包裹的词会被高亮。',
      bio: '写一段自述：你是谁、在做什么、热爱什么。',
      status: { doing: '正在做的一件事', city: '所在城市', tz: 8 },
      links: [{ name: 'GitHub', handle: '@your-github', icon: 'github', url: 'https://github.com/', primary: true }],
      portrait: { src: '', fade: 'left', focus: '50% 40%' },
    }),
    summary: (d) => [d.name, d.status?.city].filter(Boolean).join(' · '),
  },
  {
    type: 'status',
    name: '在线状态',
    desc: '此刻的状态：在线 / 专注 / 离开，附当前活动与最后活跃时间',
    icon: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18ZM12 7v5l3 2',
    group: '身份',
    variants: [{ id: 'card', label: '卡片' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ state: 'online', activity: '在写一篇新文章', app: 'VS Code', lastActive: '', device: '' }),
    summary: (d) => `${({ online: '在线', focus: '专注中', away: '离开' } as Record<string, string>)[d.state] ?? ''} · ${d.activity ?? ''}`,
  },
  {
    type: 'chapter',
    name: '章节',
    desc: '编辑式分隔：编号 + 宋体大标题 + 副题，把页面切成有节奏的段落',
    icon: 'M4 7h16M4 12h10M4 17h6',
    group: '身份',
    variants: [{ id: 'rule', label: '细线' }],
    spans: [3],
    defaultSpan: 3,
    chrome: no,
    defaultData: () => ({ no: '', title: '新章节', subtitle: '一句副题', meta: '' }),
    summary: (d) => [d.title, d.subtitle].filter(Boolean).join(' — '),
  },
  {
    type: 'socials',
    name: '社交',
    desc: '找到我：图标 + 名称 + handle，悬停箭头飞出；紧凑变体为胶囊行',
    icon: 'M10 14a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1.5 1.5M14 10a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7L12.5 19',
    group: '身份',
    variants: [{ id: 'list', label: '列表' }, { id: 'pills', label: '胶囊' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ items: [{ name: 'GitHub', handle: '@your-github', url: 'https://github.com/', icon: 'github' }] }),
    summary: (d) => names(d.items),
  },
  {
    type: 'contact',
    name: '联系',
    desc: 'CTA 卡：宋体大标题、邮箱一键复制、主按钮与回复时效',
    icon: 'M4 6h16v12H4zM4 7l8 6 8-6',
    group: '身份',
    variants: [{ id: 'card', label: '卡片' }, { id: 'wide', label: '通栏' }],
    spans: [1, 2, 3],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({
      title: '想聊聊？',
      text: '合作、提问，或者只是打个招呼，都欢迎。',
      email: 'hi@example.com',
      buttonText: '写邮件',
      url: 'mailto:hi@example.com',
      sla: '通常 24 小时内回复',
    }),
    summary: (d) => d.title ?? '',
  },
  /* ---------------- 数据 ---------------- */
  {
    type: 'stats',
    name: '站点数字',
    desc: '运行天数 / 文章 / 随想 / 标签，等宽数字滚动计数与本年进度',
    icon: 'M4 20V10M10 20V4M16 20v-8M22 20H2',
    group: '数据',
    variants: [{ id: 'row', label: '横排' }, { id: 'hero', label: '大号' }],
    spans: [1, 2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({
      items: [{ key: 'days', label: '运行天数' }, { key: 'posts', label: '文章' }, { key: 'notes', label: '随想' }, { key: 'tags', label: '标签' }],
      showYearProgress: true,
    }),
    summary: () => '实时统计 · 文章 / 随想 / 标签 / 运行天数',
  },
  {
    type: 'github',
    name: 'GitHub',
    desc: '四项计数 + 贡献热力图 + 最近动态，数据来自设置中的 GitHub 配置',
    icon: 'M9 19c-4 1.5-4-2.5-6-3m12 5v-3.5c0-1 .1-1.4-.5-2 2.8-.3 5.5-1.4 5.5-6a4.6 4.6 0 0 0-1.3-3.2 4.2 4.2 0 0 0-.1-3.2s-1.1-.3-3.5 1.3a12.3 12.3 0 0 0-6.2 0C6.5 2.8 5.4 3.1 5.4 3.1a4.2 4.2 0 0 0-.1 3.2A4.6 4.6 0 0 0 4 9.5c0 4.6 2.7 5.7 5.5 6-.6.6-.6 1.2-.5 2V21',
    group: '数据',
    variants: [{ id: 'full', label: '热力图 + 动态' }, { id: 'map', label: '仅热力图' }],
    spans: [1, 2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({ showCommits: true, commitCount: 3 }),
    summary: () => '数据来自设置中的 GitHub 配置',
  },
  {
    type: 'languages',
    name: '语言占比',
    desc: '精致堆叠条 + 图例，悬停互相高亮；变体为环形图',
    icon: 'M3 12h5l2-6 4 12 2-6h5',
    group: '数据',
    variants: [{ id: 'bar', label: '堆叠条' }, { id: 'ring', label: '环形' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({
      unit: '过去一年代码行',
      items: [
        { name: 'TypeScript', percent: 45, color: '#0078ff' },
        { name: 'Vue', percent: 35, color: '#00c853' },
        { name: 'CSS', percent: 20, color: '#ffb300' },
      ],
    }),
    summary: (d) => names(d.items),
  },
  {
    type: 'skillbars',
    name: '技能刻度',
    desc: '20 格刻度尺逐格点亮，末格发光，右侧数值与段位；变体为点阵',
    icon: 'M4 6h12M4 12h16M4 18h8',
    group: '数据',
    variants: [{ id: 'ruler', label: '刻度尺' }, { id: 'dots', label: '点阵' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ items: [{ name: '写作', level: 80 }, { name: '编程', level: 70 }, { name: '摄影', level: 60 }] }),
    summary: (d) => count(d.items, '项能力'),
  },
  {
    type: 'year',
    name: '年度回顾',
    desc: '按年切换：四个大数 + 12 个月柱状图 + 三条高光',
    icon: 'M4 20h16M7 16V9M12 16V5M17 16v-4',
    group: '数据',
    variants: [{ id: 'chart', label: '柱状' }],
    spans: [2, 3],
    defaultSpan: 3,
    chrome: yes,
    defaultData: () => {
      const y = String(new Date().getFullYear());
      return {
        metric: '提交',
        years: {
          [y]: {
            sub: '1 月 1 日 → 今天',
            nums: [['文章', 0], ['随想', 0], ['提交', 0], ['读完', 0]],
            months: Array.from({ length: 12 }, (_, i) => (i <= new Date().getMonth() ? 0 : null)),
            pins: [],
            highlights: [['一月', '写下这一年的第一件事']],
          },
        },
      };
    },
    summary: (d) => Object.keys(d.years ?? {}).sort().reverse().join(' / '),
  },
  /* ---------------- 经历 ---------------- */
  {
    type: 'now',
    name: '此刻',
    desc: 'Now 页：在做 / 在学 / 在读 / 在玩，附注释、进度与更新时间',
    icon: 'M12 3v3M12 18v3M3 12h3M18 12h3M12 12m-4 0a4 4 0 1 0 8 0a4 4 0 1 0-8 0',
    group: '经历',
    variants: [{ id: 'list', label: '列表' }, { id: 'cols', label: '两栏' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ updatedAt: new Date().toISOString().slice(0, 10), items: [{ kind: '在做', text: '正在做的一件事', note: '' }] }),
    summary: (d) => count(d.items, '件事'),
  },
  {
    type: 'milestones',
    name: '历程',
    desc: '时间线：最新一条点亮；竖排适合窄栏，横排适合宽栏',
    icon: 'M12 4v16M12 7h6M12 13h-6M12 19h6',
    group: '经历',
    variants: [{ id: 'vertical', label: '竖排' }, { id: 'horizontal', label: '横排' }],
    spans: [1, 2, 3],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ items: [{ date: String(new Date().getFullYear()), title: '写下第一条历程', text: '', now: true }] }),
    summary: (d) => count(d.items, '条历程'),
  },
  {
    type: 'projects',
    name: '作品',
    desc: '一大两小：精选项目大封面，其余横向卡；无图时用光影构图',
    icon: 'M3 7h18v13H3zM8 7V4h8v3',
    group: '经历',
    variants: [{ id: 'feature', label: '一大两小' }],
    spans: [2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({
      items: [
        { name: '项目名', desc: '一句话介绍这个项目。', url: '', cover: '', scene: 'door', lang: 'Go', color: '#0078ff', stars: 0, featured: true },
      ],
    }),
    summary: (d) => names(d.items),
  },
  {
    type: 'principles',
    name: '信条',
    desc: '编号大字：宋体渐隐数字 + 一句信条 + 一行解释，开放排版',
    icon: 'M5 4h3v16M12 8h7M12 12h7M12 16h5',
    group: '经历',
    variants: [{ id: 'cols', label: '四列' }, { id: 'big', label: '大字竖排' }],
    spans: [1, 2, 3],
    defaultSpan: 3,
    chrome: no,
    defaultData: () => ({ items: [{ title: '先写下来，再写好。', text: '草稿比完美重要。' }] }),
    summary: (d) => count(d.items, '条信条'),
  },
  /* ---------------- 喜好 ---------------- */
  {
    type: 'bookshelf',
    name: '书架',
    desc: '书脊视图：竖排书名、高矮不一，在读挂书签；点击抽出显示书卡',
    icon: 'M5 4h3v16H5zM10 4h3v16h-3zM15.5 5l3-.8 3.6 15.2-3 .8z',
    group: '喜好',
    variants: [{ id: 'spine', label: '书脊' }],
    spans: [1, 2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({ items: [{ title: '书名', author: '作者', color: '#1d3a5f', height: 90, width: 34, status: '在读', progress: 30, note: '' }] }),
    summary: (d) => count(d.items, '本书'),
  },
  {
    type: 'listening',
    name: '最近在听',
    desc: '黑胶唱片：播放时旋转、唱臂落下，进度实时走；宽版附最近曲目',
    icon: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18ZM12 14a2 2 0 1 0 0-4 2 2 0 0 0 0 4Z',
    group: '喜好',
    variants: [{ id: 'vinyl', label: '黑胶' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ playing: true, now: { title: '曲名', artist: '艺术家', album: '', duration: 240, position: 60 }, recent: [] }),
    summary: (d) => [d.now?.title, d.now?.artist].filter(Boolean).join(' — '),
  },
  {
    type: 'quotes',
    name: '语录',
    desc: '轮播：宋体大引文模糊切入，分段进度可悬停暂停；列表变体保留旧版',
    icon: 'M6 15c-1.5 0-2.5-1-2.5-2.5S4.5 10 6 10c.3-2.5 2-4 2-4M15 15c-1.5 0-2.5-1-2.5-2.5S13.5 10 15 10c.3-2.5 2-4 2-4',
    group: '喜好',
    variants: [{ id: 'rotator', label: '轮播' }, { id: 'list', label: '列表' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ interval: 6, items: [{ text: '把喜欢的话收藏在这里。', from: '出处' }] }),
    summary: (d) => count(d.items, '条语录'),
  },
  {
    type: 'motto',
    name: '格言',
    desc: '页尾收束：整行宋体大字逐字浮现，收尾装饰四选一（落款线 / 地平线 / 引号 / 无）；卡片变体可放进网格',
    icon: 'M8 10c0-2 1.5-4 4-4M8 10v4h4v-4H8ZM16 10c0-2 1.5-4 4-4M16 10v4h4v-4h-4Z',
    group: '喜好',
    variants: [{ id: 'closing', label: '收尾大字' }, { id: 'card', label: '卡片' }],
    spans: [1, 3],
    defaultSpan: 3,
    chrome: (v) => v === 'card',
    defaultData: () => ({ text: '写下你的座右铭', sign: '' }),
    summary: (d) => d.text ?? '',
  },
  {
    type: 'gallery',
    name: '画廊',
    desc: '不等宫格：首图 2×2，悬停浮出地点日期；点击从缩略图飞入大图',
    icon: 'M4 5h16v14H4zM4 15l4-4 3 3 5-5 4 4',
    group: '喜好',
    variants: [{ id: 'mosaic', label: '拼贴' }, { id: 'strip', label: '横滑' }],
    spans: [1, 2, 3],
    defaultSpan: 3,
    chrome: yes,
    defaultData: () => ({ images: [] }),
    summary: (d) => count(d.images, '张照片'),
  },
  {
    type: 'places',
    name: '足迹',
    desc: '点阵地图：城市发光脉冲，常住地为黄色，悬停与列表联动',
    icon: 'M12 21s-7-6.2-7-11.5a7 7 0 0 1 14 0C19 14.8 12 21 12 21ZM12 12a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
    group: '喜好',
    variants: [{ id: 'map', label: '点阵' }],
    spans: [1, 2],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({ region: 'china', items: [{ name: '杭州', lon: 120.2, lat: 30.3, year: '常住', home: true }] }),
    summary: (d) => names(d.items),
  },
  {
    type: 'favorites',
    name: '喜好',
    desc: '分组 Tab 切换，条目带作者 / 年份 / 一句评语',
    icon: 'M12 21s-7-4.6-9.5-9C.9 8.5 2.7 5 6 5c2 0 3.4 1 4 2 .6-1 2-2 4-2 3.3 0 5.1 3.5 3.5 7C19 16.4 12 21 12 21Z',
    group: '喜好',
    variants: [{ id: 'tabs', label: 'Tab' }],
    spans: [1, 2],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ groups: [{ title: '电影', items: [{ name: '最爱的一部', by: '', year: '', note: '' }] }] }),
    summary: (d) => names(d.groups, 'title'),
  },
  /* ---------------- 工具 ---------------- */
  {
    type: 'skills',
    name: '技能',
    desc: '三组键帽标签：按下有真实行程；每组可标一个主技能',
    icon: 'M3 7h18v10H3zM7 11h.01M11 11h.01M15 11h.01M8 14h8',
    group: '工具',
    variants: [{ id: 'keys', label: '键帽' }],
    spans: [1, 2, 3],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ groups: [{ title: '新分组', items: ['标签'], star: '' }] }),
    summary: (d) => names(d.groups, 'title'),
  },
  {
    type: 'stack',
    name: '技术栈',
    desc: '等宽字母徽标 + 名称 + 角色，品牌色仅作淡底',
    icon: 'M12 3 3 8l9 5 9-5-9-5ZM3 13l9 5 9-5',
    group: '工具',
    variants: [{ id: 'grid', label: '网格' }],
    spans: [1, 2, 3],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ items: [{ name: 'Vue 3', role: '前端框架', glyph: 'Vu', color: '#42b883' }] }),
    summary: (d) => count(d.items, '项技术'),
  },
  {
    type: 'uses',
    name: '工作台',
    desc: '/uses：硬件与软件分组，图标 + 名称 + 规格 + 标签（取代装备清单）',
    icon: 'M4 6h16v10H4zM8 20h8M12 16v4',
    group: '工具',
    variants: [{ id: 'groups', label: '分组' }],
    spans: [1, 2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({
      groups: [
        { title: '硬件', items: [{ icon: 'laptop', name: '主力电脑', desc: '型号 / 配置', tag: '主力' }] },
        { title: '软件', items: [{ icon: 'code', name: '编辑器', desc: '' }] },
      ],
    }),
    summary: (d) => (Array.isArray(d.groups) ? d.groups.map((g: { title: string; items: unknown[] }) => `${g.title} ${g.items?.length ?? 0}`).join(' / ') : ''),
  },
  /* ---------------- 互动 ---------------- */
  {
    type: 'faq',
    name: '问答',
    desc: '手风琴：高度平滑展开，加号旋转为减号；同时只展开一条',
    icon: 'M9 9a3 3 0 1 1 4 2.8c-.8.4-1 .9-1 1.7M12 17h.01',
    group: '互动',
    variants: [{ id: 'accordion', label: '手风琴' }],
    spans: [1, 2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({ single: true, items: [{ q: '一个常见的问题？', a: '一个真诚的回答。' }] }),
    summary: (d) => count(d.items, '组问答'),
  },
  {
    type: 'guestbook',
    name: '留言墙',
    desc: '输入框、留言卡、站长回复与喜欢；审核后公开',
    icon: 'M4 5h16v11H8l-4 4z',
    group: '互动',
    variants: [{ id: 'wall', label: '墙' }],
    spans: [2, 3],
    defaultSpan: 2,
    chrome: yes,
    defaultData: () => ({
      pageSize: 4,
      requireLogin: false,
      total: 1,
      items: [{ name: '访客', color: '#0078ff', at: '刚刚', text: '第一条留言。', likes: 0 }],
    }),
    summary: (d) => `每页 ${d.pageSize ?? 4} 条${d.requireLogin ? ' · 需登录' : ''}`,
  },
];

/** 旧类型（仅用于后台列表显示名称；读取时会被迁移为新类型） */
const LEGACY: ModuleMeta[] = [
  {
    type: 'devices',
    name: '装备（旧）',
    desc: '已合并入工作台 uses，读取时自动迁移',
    icon: 'M4 6h16v10H4zM8 20h8M12 16v4',
    group: '工具',
    variants: [{ id: 'groups', label: '分组' }],
    spans: [1, 2, 3],
    defaultSpan: 1,
    chrome: yes,
    defaultData: () => ({ items: [] }),
    summary: (d) => count(d.items, '件装备'),
  },
];

export function metaOf(type: string): ModuleMeta | undefined {
  return MODULE_REGISTRY.find((m) => m.type === type) ?? LEGACY.find((m) => m.type === type);
}

/** 模块当前生效的变体（未设置或无效时取第一项） */
export function variantOf(mod: Pick<AboutModule, 'type' | 'variant'>): string {
  const meta = metaOf(mod.type);
  if (!meta) return mod.variant ?? '';
  return meta.variants.some((v) => v.id === mod.variant) ? (mod.variant as string) : meta.variants[0].id;
}

/** 模块当前生效的 span（不在允许列表时取最接近的允许值） */
export function spanOf(mod: Pick<AboutModule, 'type' | 'span'>): Span {
  const meta = metaOf(mod.type);
  if (!meta) return (mod.span ?? 3) as Span;
  const want = mod.span ?? meta.defaultSpan;
  if (meta.spans.includes(want)) return want;
  return meta.spans.reduce((a, b) => (Math.abs(b - want) < Math.abs(a - want) ? b : a));
}

/** 模块标题：覆盖值优先，否则取注册名 */
export function titleOf(mod: Pick<AboutModule, 'type' | 'title'>): string {
  return mod.title?.trim() || metaOf(mod.type)?.name || mod.type;
}

export function createModule(type: string): AboutModule {
  const meta = metaOf(type);
  return {
    id: `m-${type}-${Date.now().toString(36)}-${Math.floor(Math.random() * 1e4)}`,
    type,
    span: meta?.defaultSpan ?? 3,
    variant: meta?.variants[0]?.id,
    data: meta ? meta.defaultData() : {},
  };
}

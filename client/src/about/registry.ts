import type { AboutModule } from '../stores/config';

/**
 * 关于页模块注册表：模块选择器与渲染器的唯一数据源。
 * icon 为内联 SVG path（24 viewBox，stroke 风格）。
 */
export interface ModuleMeta {
  type: string;
  name: string;
  desc: string;
  icon: string;
  /** 新增时的默认 data */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  defaultData: () => any;
  /** 管理列表行摘要 */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  summary: (data: any) => string;
}

export const MODULE_REGISTRY: ModuleMeta[] = [
  {
    type: 'stats',
    name: '站点数字',
    desc: '运行天数、文章、随想、标签的实时统计条',
    icon: 'M4 20V10M10 20V4M16 20v-8M22 20H2',
    defaultData: () => ({}),
    summary: () => '自动统计，无需配置',
  },
  {
    type: 'motto',
    name: '格言块',
    desc: '一句话座右铭，双边框宋体引用样式',
    icon: 'M8 10c0-2 1.5-4 4-4M8 10v4h4v-4H8ZM16 10c0-2 1.5-4 4-4M16 10v4h4v-4h-4Z',
    defaultData: () => ({ text: '写下你的座右铭' }),
    summary: (d) => d.text ?? '',
  },
  {
    type: 'skills',
    name: '技能分组',
    desc: '多组标签卡：创作 / 工具 / 兴趣…',
    icon: 'M4 6h7M4 12h16M4 18h10',
    defaultData: () => ({ groups: [{ title: '新分组', items: ['标签'] }] }),
    summary: (d) => (d.groups ?? []).map((g: { title: string }) => g.title).join(' / '),
  },
  {
    type: 'skillbars',
    name: '技能图表',
    desc: '横向能力条形图，0–100 熟练度可视化',
    icon: 'M4 6h12M4 6v0M4 12h16M4 18h8',
    defaultData: () => ({
      items: [
        { name: '写作', level: 80 },
        { name: '摄影', level: 60 },
        { name: '编程', level: 70 },
      ],
    }),
    summary: (d) => `${(d.items ?? []).length} 项能力`,
  },
  {
    type: 'languages',
    name: '语言占比',
    desc: '堆叠比例条：编程语言 / 时间分配等占比图',
    icon: 'M3 12h5l2-6 4 12 2-6h5',
    defaultData: () => ({
      items: [
        { name: 'TypeScript', percent: 45, color: '#0078ff' },
        { name: 'Vue', percent: 35, color: '#00c853' },
        { name: 'CSS', percent: 20, color: '#ffb300' },
      ],
    }),
    summary: (d) => (d.items ?? []).map((x: { name: string }) => x.name).join(' · '),
  },
  {
    type: 'milestones',
    name: '历程时间线',
    desc: '按年份/序号排布的人生与站点大事记',
    icon: 'M12 4v16M12 7h6M12 13h-6M12 19h6',
    defaultData: () => ({ items: [{ year: '2026', text: '写下第一条历程' }] }),
    summary: (d) => `${(d.items ?? []).length} 条历程`,
  },
  {
    type: 'gallery',
    name: '照片墙',
    desc: '生活照片宫格，点击可全屏查看',
    icon: 'M4 5h16v14H4zM4 15l4-4 3 3 5-5 4 4',
    defaultData: () => ({ images: [] }),
    summary: (d) => `${(d.images ?? []).length} 张照片`,
  },
  {
    type: 'quotes',
    name: '语录集',
    desc: '喜欢的话：多条引用与出处',
    icon: 'M6 15c-1.5 0-2.5-1-2.5-2.5S4.5 10 6 10c.3-2.5 2-4 2-4M15 15c-1.5 0-2.5-1-2.5-2.5S13.5 10 15 10c.3-2.5 2-4 2-4',
    defaultData: () => ({ items: [{ text: '把喜欢的话收藏在这里', from: '出处' }] }),
    summary: (d) => `${(d.items ?? []).length} 条语录`,
  },
  {
    type: 'devices',
    name: '装备清单',
    desc: '在用的设备与工具：主机 / 相机 / 键盘…',
    icon: 'M4 6h16v10H4zM8 20h8M12 16v4',
    defaultData: () => ({ items: [{ name: '设备名', desc: '型号 / 用途' }] }),
    summary: (d) => `${(d.items ?? []).length} 件装备`,
  },
  {
    type: 'favorites',
    name: '喜好清单',
    desc: '分组列出电影 / 音乐 / 书籍 / 游戏…',
    icon: 'M12 21s-7-4.6-9.5-9C.9 8.5 2.7 5 6 5c2 0 3.4 1 4 2 .6-1 2-2 4-2 3.3 0 5.1 3.5 3.5 7C19 16.4 12 21 12 21Z',
    defaultData: () => ({ groups: [{ title: '电影', items: ['最爱的一部'] }] }),
    summary: (d) => (d.groups ?? []).map((g: { title: string }) => g.title).join(' / '),
  },
  {
    type: 'faq',
    name: '问答 FAQ',
    desc: '关于我的常见问题与回答',
    icon: 'M9 9a3 3 0 1 1 4 2.8c-.8.4-1 .9-1 1.7M12 17h.01',
    defaultData: () => ({ items: [{ q: '一个常见的问题？', a: '一个真诚的回答。' }] }),
    summary: (d) => `${(d.items ?? []).length} 组问答`,
  },
  {
    type: 'now',
    name: '正在做',
    desc: 'Now 页理念：此刻在学、在做、在玩什么',
    icon: 'M12 3v3M12 18v3M3 12h3M18 12h3M12 12m-4 0a4 4 0 1 0 8 0a4 4 0 1 0-8 0',
    defaultData: () => ({ items: ['正在做的一件事'] }),
    summary: (d) => `${(d.items ?? []).length} 件事`,
  },
  {
    type: 'github',
    name: 'GitHub 状态',
    desc: '仓库/Stars/关注者统计、贡献热力图与最新动态',
    icon: 'M9 19c-4 1.5-4-2.5-6-3m12 5v-3.5c0-1 .1-1.4-.5-2 2.8-.3 5.5-1.4 5.5-6a4.6 4.6 0 0 0-1.3-3.2 4.2 4.2 0 0 0-.1-3.2s-1.1-.3-3.5 1.3a12.3 12.3 0 0 0-6.2 0C6.5 2.8 5.4 3.1 5.4 3.1a4.2 4.2 0 0 0-.1 3.2A4.6 4.6 0 0 0 4 9.5c0 4.6 2.7 5.7 5.5 6-.6.6-.6 1.2-.5 2V21',
    defaultData: () => ({}),
    summary: () => '数据来自设置中的 GitHub 配置',
  },
  {
    type: 'socials',
    name: '社交链接',
    desc: '找到我：GitHub / 邮箱 / RSS / 任意链接',
    icon: 'M10 14a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1.5 1.5M14 10a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7L12.5 19',
    defaultData: () => ({ items: [{ name: 'GitHub', url: 'https://github.com/', icon: 'github' }] }),
    summary: (d) => (d.items ?? []).map((x: { name: string }) => x.name).join(' · '),
  },
  {
    type: 'stack',
    name: '技术栈',
    desc: '本站或本人使用的技术卡片',
    icon: 'M12 3 3 8l9 5 9-5-9-5ZM3 13l9 5 9-5',
    defaultData: () => ({ items: [{ name: 'Vue 3', role: '前端框架' }] }),
    summary: (d) => `${(d.items ?? []).length} 项技术`,
  },
  {
    type: 'contact',
    name: '联系我 CTA',
    desc: '大按钮行动号召卡：合作 / 邮件 / 订阅',
    icon: 'M4 6h16v12H4zM4 7l8 6 8-6',
    defaultData: () => ({
      title: '想聊聊？',
      text: '合作、提问或只是打个招呼，都欢迎。',
      buttonText: '发邮件',
      url: 'mailto:hi@example.com',
    }),
    summary: (d) => d.title ?? '',
  },
];

export function metaOf(type: string): ModuleMeta | undefined {
  return MODULE_REGISTRY.find((m) => m.type === type);
}

export function createModule(type: string): AboutModule {
  const meta = metaOf(type);
  return {
    id: `m-${type}-${Date.now().toString(36)}-${Math.floor(Math.random() * 1e4)}`,
    type,
    data: meta ? meta.defaultData() : {},
  };
}

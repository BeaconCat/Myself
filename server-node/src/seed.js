/**
 * 演示种子数据：全部围绕 Myself 平台自身的功能介绍，
 * 便于开源分发后新用户直接看到可用的示例内容。
 */
const img = (from, to, label) => `/api/v1/img/${from}/${to}/${label}`;

export const seedNotes = [
  {
    contentMd: '欢迎使用 **Myself**。这是一条「随想」——短内容流，支持 Markdown、心情标签和最多九张配图。',
    mood: '欢迎',
    images: [img('0078ff', '00295c', 'Welcome')],
    createdAt: '2026-07-10 12:00:00',
  },
  {
    contentMd: '随想的配图支持双拼、三拼、四宫格直到九宫格，点开可全屏缩放拖拽查看。',
    mood: '功能',
    images: [
      img('ff0032', '7a0020', 'Grid-1'),
      img('ffb300', '7a5200', 'Grid-2'),
      img('00c853', '00512a', 'Grid-3'),
      img('0078ff', '00295c', 'Grid-4'),
    ],
    createdAt: '2026-07-09 18:30:00',
  },
  {
    contentMd: '试试右上角的主题切换：四季色盘 × 深浅模式，切换时有圆形蒙版扩散动画。',
    mood: '主题',
    createdAt: '2026-07-08 09:15:00',
  },
  {
    contentMd: '后台「API 中心」可以创建 APIKey，把发文托管给你的 AI 助手——内容统一走 Markdown。',
    mood: 'API',
    createdAt: '2026-07-07 21:40:00',
  },
];

export const seedPosts = [
  {
    slug: 'welcome-to-myself',
    title: '欢迎使用 Myself',
    excerpt: '一个开源的个人博客引擎：Vue 3 + Express + SQLite，内容皆 Markdown，主题随四季流转。',
    tags: ['指南'],
    covers: [],
    createdAt: '2026-07-01 10:00:00',
    contentMd: `# 欢迎使用 Myself

Myself 是一个开源的个人博客引擎，前端 Vue 3 + Vite，后端 Express + SQLite，所有内容以 **Markdown** 统一存储。

## 快速开始

1. 登录后台（\`/admin\`，默认账号 \`admin / myself-admin\`，请立即修改密码）
2. 在「设置」里配置站点标题、主题色与关于信息
3. 在「文章管理」发布你的第一篇文章

> 这篇文章本身就是演示数据，可以在后台随时删除。`,
  },
  {
    slug: 'theme-system-guide',
    title: '主题系统指南',
    excerpt: '两层主题架构：深浅模式与色盘正交组合；后台可自定义最多十组主题色并拖拽排序。',
    tags: ['指南', '主题'],
    covers: [],
    createdAt: '2026-07-02 14:00:00',
    contentMd: `# 主题系统指南

主题 = **模式**（浅色 / 深色）×**色盘**（最多十组预设，对访客展示前四组）。

## 自定义

后台「设置 → 主题」中：

- 修改每组预设的主色与深主色，整套界面色由主色自动派生
- 拖拽调整顺序，前 N 组对访客可见
- 可关闭访客换肤，只保留深浅切换
- 支持按季节自动切换色盘

\`\`\`css
/* 组件只消费变量，换主题零改动 */
color: var(--primary);
\`\`\``,
  },
  {
    slug: 'writing-with-markdown',
    title: '用 Markdown 写作',
    excerpt: '文章与随想共用一条 Markdown 管线：可移植、可版本化、对 AI 友好。',
    tags: ['指南', '写作'],
    covers: [],
    createdAt: '2026-07-03 09:30:00',
    contentMd: `# 用 Markdown 写作

数据库只存 Markdown 原文，渲染发生在展示端。

| 能力 | 说明 |
| --- | --- |
| 代码块 | 支持围栏语法 |
| 表格 | 就像这张 |
| 引用 | 见下方 |

> 内容不锁在任何编辑器里——随时导出，随时迁移。`,
  },
  {
    slug: 'api-center-for-agents',
    title: '把发文托管给 AI：API 中心',
    excerpt: '创建 APIKey，让外部 AI 通过 X-Api-Key 认证的 REST 接口替你发布文章与随想。',
    tags: ['指南', 'API'],
    covers: [],
    createdAt: '2026-07-04 16:20:00',
    contentMd: `# 把发文托管给 AI：API 中心

后台「API 中心」创建 Key 后，外部程序即可投稿：

\`\`\`bash
curl -X POST https://your-site/api/v1/ext/posts \\
  -H "X-Api-Key: myk_xxx" \\
  -H "Content-Type: application/json" \\
  -d '{"slug":"hello","title":"你好","contentMd":"# 来自 AI 的第一篇"}'
\`\`\`

- Key 只在创建时显示一次，哈希入库
- 可随时吊销，带最后使用时间追踪
- 外部投稿默认进草稿箱，人工审核后发布`,
  },
];

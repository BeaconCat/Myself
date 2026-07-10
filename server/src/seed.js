/** 演示随想（短内容信息流，images 为配图 URL 数组，演示走本地占位图接口） */
const img = (from, to, label) => `/api/v1/img/${from}/${to}/${label}`;

export const seedNotes = [
  {
    contentMd: '主题系统的圆形蒙版切换终于顺滑了。`View Transition API` 真香，但 `fill: forwards` 的残留动画坑了我两次。',
    mood: '技术',
    images: [img('0078ff', '00295c', 'ViewTransition')],
    createdAt: '2026-07-10 21:40:00',
  },
  {
    contentMd: '夏天的傍晚，窗外的云被烧成 `#ff0032`。切到炽红主题，网站和天空同色。',
    mood: '生活',
    images: [
      img('ff0032', '7a0020', 'Sunset-1'),
      img('ff5c7a', 'a3002a', 'Sunset-2'),
    ],
    createdAt: '2026-07-09 19:22:00',
  },
  {
    contentMd: '想清楚了：博客的一切内容都走 Markdown。**文章**是长文，**随想**是短流，同一套渲染管线。',
    mood: '思考',
    createdAt: '2026-07-07 23:05:00',
  },
  {
    contentMd: '给首页做了个立体相册轮播，卡片从三个方向飞进来的瞬间，值了。三张设计稿：',
    mood: '技术',
    images: [
      img('ffb300', '7a5200', 'Draft-1'),
      img('e09600', '3a2800', 'Draft-2'),
      img('ffd166', '7a5200', 'Draft-3'),
    ],
    createdAt: '2026-07-05 15:48:00',
  },
  {
    contentMd: '雨。适合把收藏夹里的长文清一清，顺便给灯塔图标画了六版草稿。',
    mood: '生活',
    images: [
      img('ff0032', '7a0020', 'Icon-1'),
      img('ffb300', '7a5200', 'Icon-2'),
      img('0078ff', '00295c', 'Icon-3'),
      img('00c853', '00512a', 'Icon-4'),
      img('ff5c7a', 'a3002a', 'Icon-5'),
      img('4d9fff', '003d80', 'Icon-6'),
    ],
    createdAt: '2026-07-03 11:30:00',
  },
  {
    contentMd: '整理了一版九图测试，宫格拼图压力测试专用。',
    mood: '测试',
    images: [
      img('ff0032', '7a0020', 'G-1'), img('ffb300', '7a5200', 'G-2'), img('0078ff', '00295c', 'G-3'),
      img('00c853', '00512a', 'G-4'), img('ff5c7a', 'a3002a', 'G-5'), img('4d9fff', '003d80', 'G-6'),
      img('e09600', '3a2800', 'G-7'), img('d40029', '40000d', 'G-8'), img('005fd6', '001b3d', 'G-9'),
    ],
    createdAt: '2026-07-01 09:12:00',
  },
];

/** 演示种子文章（Markdown 正文） */
export const seedPosts = [
  {
    slug: 'seasonal-theme-system',
    title: '用 CSS Variables 打造季节主题系统',
    excerpt: '两层主题架构：深浅模式与季节色盘正交组合，四季配色一键切换，品牌荧光三色贯穿始终。',
    tags: ['技术', '前端'],
    covers: [],
    createdAt: '2026-06-18 10:00:00',
    contentMd: `# 用 CSS Variables 打造季节主题系统

主题系统的核心是把「模式」和「色盘」拆成两个正交维度：

- **mode**：light / dark
- **palette**：spring / summer / autumn / winter，以及后续任意自定义色盘

## 变量注入

所有颜色都通过 CSS 变量下发到 \`:root\`：

\`\`\`ts
root.style.setProperty('--bg', colors.bg);
root.style.setProperty('--primary', colors.primary);
\`\`\`

组件层永远只消费变量，不写死颜色。换主题 = 换一组变量，零组件改动。

## 品牌常量

三个荧光强调色不随主题变化：

| 颜色 | 值 |
| --- | --- |
| 红 | #ff0032 |
| 黄 | #ffb300 |
| 蓝 | #0078ff |

> 主题会变，品牌不变。`,
  },
  {
    slug: 'glass-and-3d',
    title: '毛玻璃与 3D：简约的惊艳',
    excerpt: 'backdrop-filter 胶囊导航、透视卡片飞入飞出，克制的动效如何撑起高级感。',
    tags: ['设计', '前端'],
    covers: [],
    createdAt: '2026-06-25 21:30:00',
    contentMd: `# 毛玻璃与 3D：简约的惊艳

「简约的惊艳」不是堆特效，而是把每个动效都放在用户注意力的路径上。

## 毛玻璃

\`\`\`css
backdrop-filter: blur(14px) saturate(1.5);
\`\`\`

饱和度提升 1.5 倍是关键——纯 blur 会发灰，加饱和后玻璃后面的颜色会透出来。

## 3D 卡片

透视来自父容器的 \`perspective\`，卡片自身只做 \`rotateY\`。入场用回弹曲线，出场用标准出射曲线，时长压在 0.45s 内。`,
  },
  {
    slug: 'markdown-as-content-spec',
    title: 'Markdown 作为博客的统一内容规范',
    excerpt: '从后台编辑到 API 发文，一切内容皆 Markdown：可移植、可版本化、对 AI 友好。',
    tags: ['架构'],
    covers: [],
    createdAt: '2026-07-02 14:12:00',
    contentMd: `# Markdown 作为博客的统一内容规范

数据库里只存 Markdown 原文，渲染永远发生在展示端。

## 为什么

1. **可移植**：导出即纯文本，不锁在任何编辑器
2. **可版本化**：diff 友好，配 Git 天然合拍
3. **对 AI 友好**：API 中心接收外部 AI 投稿时，Markdown 是所有模型的母语

## 接口约定

\`POST /api/v1/posts\` 只接受 \`contentMd\` 字段，服务端不做任何 HTML 转换。`,
  },
  {
    slug: 'source-han-typography',
    title: '思源黑体与宋体的排版实践',
    excerpt: '黑体承担 UI 与正文，宋体点睛标题与文章，离线 woff2 子集化控制体积。',
    tags: ['排版', '设计'],
    covers: [],
    createdAt: '2026-07-08 09:45:00',
    contentMd: `# 思源黑体与宋体的排版实践

## 分工

- **思源黑体**：UI 控件、正文段落——中性、耐读
- **思源宋体**：大标题、文章标题——笔画的衬线感在大字号下最出彩

## 体积控制

全量 CJK 字体单个 weight 就要 8–16MB，必须子集化：

\`\`\`bash
pyftsubset SourceHanSansSC-Regular.otf --text-file=charset.txt --flavor=woff2
\`\`\`

常用 3500 字 + 标点后，单字重可压到 1MB 上下。`,
  },
];

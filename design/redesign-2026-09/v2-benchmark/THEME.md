# Myself v2 · 标杆迁移 · Design Tokens

前台参照 rauno.me（深底、灰阶层级、单列编辑感、被触摸才回应的微交互）；后台参照 Ghost Admin（极简侧栏、编辑器为中心、状态用小点）。所有值落地即 `:root[data-mode][data-palette]` 变量。

## 色彩
| token | dark（前台默认） | light（后台默认） | 用途 |
|---|---|---|---|
| `--bg` | `#070B14` | `#F7F6F3` 前台 / `#F4F4F2` 后台画布 | 页面底 |
| `--surface` / `--surface-2` | `#0E1524` / `#131B2D` | `#FFFFFF` / `#F7F7F5` | 卡片、输入框、抬升层 |
| `--line` / `--line-2` | `rgba(255,255,255,.07/.13)` | `rgba(16,21,34,.08/.16)` | hairline 分隔、描边 |
| `--text` / `--text-2` / `--text-3` | `#E9EDF5` / `#9CA6BA` / `#5C6780` | `#12172A` / `#525B70` / `#8B93A6` | 三档灰阶承担层级 |
| `--accent-red/yellow/blue` | `#ff0032` / `#ffb300` / `#0078ff` | 同 | 信号色：6px 点、状态、标记，不铺面 |
| `--primary` / `--primary-deep` | 冬 `#0078ff/#0052c2` | 夏 `#ff0032/#b8001f` | 焦点环、当前项指示、链接下划线、「光」 |
| `--primary-soft` / `--primary-line` | `color-mix(primary 16% / 45%, transparent)` | 同 | 运行时派生，不维护独立色盘 |
| 季节预设 | 春 `#00c853/#00893a` · 夏 `#ff0032/#b8001f` · 秋 `#ffb300/#b87d00` · 冬 `#0078ff/#0052c2` | | 后台最多 10 组，可拖拽 |
| 状态映射 | 蓝 = 已发布/完成 · 黄 = 定时/进行中/待审 · 红 = 草稿/危险 · `#00c853` = 健康 | | 一律「小点 + 文字」 |

## 「光」母题（唯一允许的大面积色）
- `.beam`：2px 竖向光柱 `linear-gradient(#fff .55 → 0)` + 顶部 `radial-gradient` 辉光 `--glow`（dark 6% / light 10%）+ 地面椭圆 `primary 14%` blur 10px。
- 封面占位 = `/api/v1/img/<from>/<to>/<label>`：`linear-gradient(135deg, deep, primary 55%, tint)` + soft-light 高光 + 门缝细线。

## 字体
- UI/正文：`Noto Sans SC` 400/500/600（落地 @fontsource/noto-sans-sc）
- 标题/文章正文/预览：`Noto Serif SC` 500/600/700
- 元信息/数字/日期/编号/代码：`JetBrains Mono` 400/500（不用 Inter/Roboto 做 display）
- 字号阶：11 / 12 / 13.5 / 15 / 16.5（文章）/ 19 / 24 / 26 / 34–40（详情 h1）/ 52（Hero）/ 96（404）
- 行高：UI 1.6 · 正文 1.75 · 文章 1.95；等宽 label 字距 `.12–.14em`
- 阅读栏 `--col: 680px`，页面栏 `--wide: 1040px`；后台侧栏 `--sb: 232px`

## 圆角 / 阴影 / 间距
- 圆角：`--r-s 6` 缩略图/输入框 · `--r-m 10` 卡片/表格 · `--r-l 14` 封面/浮层 · `--r-pill 999` 胶囊导航/chip
- 阴影：dark `inset 0 1px 0 rgba(255,255,255,.04), 0 20px 50px -20px rgba(0,0,0,.7)`；light `0 1px 2px rgba(16,21,34,.05), 0 12px 32px -16px rgba(16,21,34,.18)`；浮层 `--shadow-lg` 更深一档
- 间距尺度：4 / 8 / 12 / 16 / 22 / 28 / 40 / 56 / 64；屏内段落间距 44，卡片内边距 18–24
- 毛玻璃：`backdrop-filter: blur(18px) saturate(1.4)`，底 `color-mix(surface 72%, transparent)`
- 禁区：圆角卡片 + 左侧彩色 border；紫色渐变；emoji；大面积信号色

## 动效（统一 motion tokens）
- `--dur-fast .2s` hover 变色 · `--dur .35s` 位移/展开 · `--dur-slow .6s` 进场/reveal
- `--ease-out cubic-bezier(.2,.8,.3,1)` · `--ease-spring cubic-bezier(.2,.8,.3,1.2)`（hover 上浮、卡片抬起）
- 清单：
  1. 页面路由过渡：模糊 8px + 位移 10px 淡入（`.w` 逐字随机延迟 0–.45s，Hero 标题/摘要）
  2. 滚动 reveal：IntersectionObserver，`.rv` 14px 上浮 + 淡入，`rootMargin -8%`
  3. 悬停微动：行标题 `translateX(3px)` spring；卡片 `translateY(-3px)`；链接下划线 `scaleX` 光扫
  4. 3D 封面卡：`perspective 1400px`，栈位 `rotateY(-14°) rotateX(6°)`，飞出 `rotateY(-40°) translate3d(-260,-40,200) blur 6px`；组内 3.5s 轮转，10s 换篇
  5. 封面手风琴：`flex 1 ⇄ 2.6`，`--dur-slow`
  6. 主题切换：View Transition 圆形蒙版从点击点扩散 600ms（无支持则直接切换）
  7. 命令面板：`pop` 18px 上浮 + `scale(.97→1)` spring
  8. 404 门开：两扇 `rotateY(∓58°)` 1.6s，地面光晕
  9. 后台：抽屉滑入、toggle 圆点 spring、拖拽行 `rotate(-.6°) scale(1.01)` + 深阴影、表格行 hover 才显示操作
- `prefers-reduced-motion: reduce` → 全部时长 .01ms，reveal 直接可见

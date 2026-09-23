# Myself · v1-cosmic · Token 清单（复古未来太空图录）

## 色彩（所有值均为 CSS variables，挂在 `:root[data-mode][data-palette]`）
| Token | Dark（默认） | Light | 用途 |
|---|---|---|---|
| `--bg` / `--bg-2` / `--bg-3` / `--bg-4` | `#070B14` / `#0B1220` / `#101A2C` / `#16233A` | `#F0EAD8` / `#E8E1CC` / `#F7F3E8` / `#DED6BE` | 页面底 / 卡片底 / 浮层底 / 控件底 |
| `--ink` / `--ink-2` / `--ink-3` / `--ink-4` | `#F0EAD8` / 72% / 46% / 26% | `#0B1220` / 74% / 50% / 28% | 正文 / 次级 / 标签 / 占位（rgba 透明度） |
| `--line` / `--line-2` | ink 14% / 8% | ink 16% / 8% | 分隔线 / 极淡分隔 |
| `--glass` / `--glass-line` | `rgba(11,18,32,.62)` / ink 16% | `rgba(240,234,216,.7)` / ink 14% | 毛玻璃导航、浮层 |
| `--star` | `#F0EAD8` | `#0B1220` | 星点填充 |
| `--primary` / `--primary-deep` | 由 palette 决定 | 同 | 冬 `#0078ff/#0052b3` 春 `#00c853/#00803a` 夏 `#ff0032/#b8001f` 秋 `#ffb300/#b87d00` |
| `--primary-soft/-line/-glow/-ink` | `color-mix(primary 14% / 40% / 32% / 78%+ink)` | 同 | 浅底 / 描边 / 辉光 / 可读文字色（运行时派生，对应 derive.ts） |
| `--accent-red/-yellow/-blue` | `#ff0032` / `#ffb300` / `#0078ff` | 常量 | 仅作状态点 / 星点 / 标记，绝不铺面 |

品牌母题翻译：灯塔 → `.beam` 锥形光（conic-gradient + blur 14px）；门/光 → 星图上的高光点；封面 = 渐变 + 线描轨道 SVG（data-uri）。

## 字体
- 中文标题 / 文章：`Noto Serif SC`（思源宋体）600
- 西文 display / 数字：`Cormorant Garamond` 500，斜体 400 作副题；字距 `.04–.16em`
- UI / 正文：`Noto Sans SC` 400/500，正文 16px/1.8（后台 14px/1.7）
- 编号 / label / 代码：`JetBrains Mono` 500，`10.5–11px`，`letter-spacing:.16em`，大写
- 字号阶：Display 64 · H1 44–48 · H2 28–30 · H3 19–22 · Body 16 · Small 13 · Label 11

## 圆角 / 阴影 / 间距
- 圆角：`--r-xs 2` `--r-sm 4`（按钮、输入）`--r-md 8`（封面、小卡）`--r-lg 14`（卡片，后台 12）`--r-pill 999`（导航、chip）
- 阴影：`--shadow 0 20px 60px rgba(0,0,0,.45)`、`--shadow-sm 0 8px 24px rgba(0,0,0,.35)`；浅色分别 `.14/.10`；主色辉光 `0 10px 30px var(--primary-glow)`
- 间距（8 基数）：4 · 8 · 12 · 16 · 20 · 24 · 28 · 34 · 44 · 56 · 72；内容最大宽 1200，gutter 40；后台侧栏 232
- 网格：前台 12 栏 gap 20–24；后台 12 栏 gap 16
- 禁用组合：圆角卡 + 左彩条；紫渐变；emoji

## 动效
- Tokens：`--dur-fast .2s` `--dur .35s` `--dur-slow .6s`；`--ease-out cubic-bezier(.2,.8,.3,1)`；`--ease-spring cubic-bezier(.2,.8,.3,1.2)`
- 页面进场：`.rv` → IntersectionObserver 加 `.in`（opacity 0→1，translateY 18→0，`--dur-slow`）
- Hero 文字：`wIn` 0.8s blur 14px→0 + translateY 14，每个词随机延迟 0–.35s；`wOut` .55s 反向
- Hero 封面：`flyIn`（translateX 220 / translateZ -400 / rotateY -45 → 静止，`--ease-spring`）；`flyOut` .6s；每 10s 换一篇；进度条 `pipRun` 10s linear
- 悬停：卡片 `translateY(-4px)` + 边框提亮 + `--shadow-sm`；按钮 `translateY(-2px)`；主按钮辉光
- 3D 倾斜：`mousemove` → `rotateY(±8deg) rotateX(±8deg)`，`perspective 900px`
- 毛玻璃：`backdrop-filter: blur(18px) saturate(1.3)`（导航）/ `blur(24px)`（浮层）/ `blur(12px)`（后台顶栏）
- 封面手风琴：`flex 1 → 2.6`，`--dur-slow`；进度条流光 `shine` 1.4s
- 主题切换：`data-mode/data-palette` 切换 + 圆形蒙版扩散（View Transition，落地时实现）
- `prefers-reduced-motion: reduce` → 全部 duration .001ms，`.rv` 直接可见，Hero 不做飞入

# Myself · v3「白」token 清单（Hara × Rams）

## 色彩（`:root[data-mode][data-palette]`）
| token | light（纸） | dark（夜间的纸） |
|---|---|---|
| `--paper` / `--paper-2` / `--paper-3` | `#F6F5F1` / `#EEEDE8` / `#E6E4DE` | `#0B1220` / `#101A2C` / `#162238` |
| `--ink` / `--ink-2` / `--ink-3` / `--ink-4` | `#161616` / `#4B4A46` / `#8B8A84` / `#B8B6AF` | `#E6E1D8` / `#B3AEA5` / `#77746E` / `#4B4E57` |
| `--line` / `--line-2` / `--line-3` | `rgba(22,22,22,.10/.05/.22)` | `rgba(230,225,216,.12/.06/.26)` |
| `--glass`（胶囊导航 / 发布条） | `rgba(246,245,241,.72)` + blur 14px | `rgba(11,18,32,.70)` + blur 14px |
| `--light`（光路） / `--floor`（地面光晕） | `rgba(255,255,255,.95)` / `rgba(0,120,255,.10)` | `rgba(255,255,255,.85)` / `rgba(0,120,255,.18)` |
| `--code-bg` / `--field` | `#EFEEE9` / `#FBFAF8` | `#0F1727` / `#0E1626` |
| `--shadow`（唯一一档浮层阴影） | `0 1px 2px rgba(20,20,20,.04), 0 18px 48px -18px rgba(20,20,20,.16)` | `0 1px 2px rgba(0,0,0,.3), 0 24px 60px -20px rgba(0,0,0,.6)` |

- 品牌常量：`--accent-red:#ff0032` `--accent-yellow:#ffb300` `--accent-blue:#0078ff`（不随主题变）
- 季节 `--primary / --primary-deep`：spring `#00c853/#009440` · summer `#ff0032/#c20026` · autumn `#ffb300/#c98a00` · winter `#0078ff/#0052b3`；其余色全部由 `--paper/--ink` 系派生，不再为色盘维护变量
- 「朱砂印」用法约定：当前导航 / 激活 tab / 筛选项 = `--primary` 4px 圆点；置顶 = `--accent-yellow`；发布 / 主要动作 = `--accent-red` 实底；链接 / API / 进行中 = `--accent-blue`；语义 ok/warn/err = green/yellow/red。三色永不铺面。
- 封面占位：`linear-gradient(150deg, <季节色或品牌色>, <其深色>)` + 左上 `radial-gradient` 白光 + 底部 30% 压暗，左下 mono 标签

## 字体
- 标题 / 文章正文 / 随想正文 / 引用：`Noto Serif SC`（思源宋体）400 / 500 / 600；display 54px/1.24，h1 40px/1.35，h2 26px，h3 20px，正文 17px/2.0（后台预览 14.5px/1.9）
- UI / 摘要 / 表格：`Noto Sans SC`（思源黑体）400 / 500；前台 15px/1.75，后台 13.5px/1.6
- 日期 / 编号 / 标签 / Key / 日志 / 代码：`IBM Plex Mono` 400 / 500；11–12px，`letter-spacing:.06–.14em`，label 大写
- 大数字（404、统计、概览）：`Cormorant Garamond` 500 / 600；不用于任何正文
- 全局 `font-feature-settings:"palt" 1`（后台加 `"tnum" 1`）

## 尺度
- 容器：前台 `--w:1120px`，正文 `--measure:680px`，封面带 1280px；后台侧栏 `--side:216px`，内容 `--w:1160px`
- 间距（4 基数）：`4 / 8 / 12 / 16 / 24 / 32 / 48 / 64 / 96 / 144`（`--s-1 … --s-10`）
- 圆角：`--r-1:2px`（控件、代码、缩略图）`--r-2:4px`（封面、面板）`--r-pill:999px`（胶囊导航、开关）；头像 26px 或 50%
- 边框：一律 1px hairline；分区标题用 `1px solid var(--ink)` 上边线，列表项用 `var(--line)` 下边线；无卡片描边 + 左彩条组合
- 阴影：仅浮层（搜索面板、模态、hero 封面卡、手机 mock）用 `--shadow`；其余零阴影

## 动效（尊重 `prefers-reduced-motion`：动画/过渡时长归零，reveal 直接可见）
- token：`--dur-fast:.2s` `--dur:.35s` `--dur-slow:.6s` `--dur-breath:6s`；`--ease-out:cubic-bezier(.2,.8,.3,1)` `--ease-spring:cubic-bezier(.2,.8,.3,1.2)`
- 滚动入场 `.rv`：opacity 0→1 + translateY 12px→0，`--dur-slow --ease-out`，IntersectionObserver rootMargin `-6%`，一次性
- Hero 文字：`blurIn`（blur 10px→0，随机 0–260ms 延迟）/ `blurOut`（.4s）；每 10s 换文章，组内封面每 3.2s 轮转
- Hero 封面卡：3 层 `translate3d(44px·n, 26px·n, -90px·n) rotateY(-10deg)`；`flyIn .7s --ease-spring` / `flyOut .5s`；鼠标视差 rotateY ±6° / rotateX ±4°
- 呼吸光晕 `breath`：opacity .7↔1 + scaleX 1↔1.08，6s ease-in-out 无限（hero 地面、404、登录页）
- 悬停：列表项 translateX 6px；封面 translateY -3px；hero CTA 横线 36→56px；按钮 active scale .98 / translateY 1px
- 封面手风琴：flex 1↔3.2，`--dur-slow --ease-out`；进度条 `fill` scaleX 0→1 线性（8–10s）
- 浮层：搜索面板 / 模态 `panelIn`（opacity + translateY -10px + scale .985，`--dur-slow`）；进度条 `shimmer` 2.4s 线性
- 主题切换：圆形蒙版扩散（由 Vue 侧 View Transition 实现，设计稿只切变量）

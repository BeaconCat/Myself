# 第三轮落地 · 并行开发约定

设计稿：`design/round3/desktop/index.html`（顶部工具条可切页面、深浅、色盘、强调方式；文件顶部注释有全部 token 与设计决策）。

## 用户定稿（必须遵守）

1. **强调系统 = 抬升 + 轻染**（原型里的默认方式 `data-accent="blend"`）：
   - 选中态（导航、tab、segmented、chip、日历选中、色盘选中等）= `var(--lift)` 背景 + `var(--lift-shadow)` + `var(--lift-fg)` 文字；选中块在项之间平滑 morph（前缘 .30s ease-out，后缘 .46s spring 延迟 .05s）。
   - **导航选中不挂信号点**；chip 选中保留前置 4px 主色圆点；segmented 选中项图标着色 `var(--ink)`。
   - 主按钮 = `var(--solid)` 实底 + `var(--on-solid)` 文字 + `var(--btn-shadow)`；hover `var(--solid-hover)`；按下 scale(.97)。**禁止**渐变填充、彩色外发光、文字投影（`text-shadow`）。
   - 主色只做信号：链接与当前位置文字用 `var(--ink)`；焦点环 `box-shadow: var(--focus)`。
   - 发光只属于品牌：只有封面光影与 Hero 门缝光可以 blur/glow；控件、卡片阴影一律用中性的 `--shadow-card` / `--shadow-pop` / `--btn-shadow`。
   - 标签：列表里「# 名称」纯文字（# 用 `--text-3`）；需要点击的区域用极轻描边胶囊。
2. **顶栏保留现有结构**：居中胶囊导航 + 右侧色盘圆点选择器 + 深浅切换（`components/layout/NavBar.vue`、`ThemeSwitcher.vue`），只按新强调系统重做视觉。另加 **Ctrl/⌘+K 唤起搜索浮层**（结构参照原型 ⌘K 与移动端 `components/mobile/SearchOverlay.vue`），导航里不强制加搜索按钮。
3. **圆角全局改小且可配置**：只用 token `--r-xs / --r-sm / --r-md / --r-lg / --r-xl / --r-pill`（基准 `--r-base` 默认 10，由站点配置 `theme.radius` 写入，`stores/theme.ts` 的 `applyRadius`）。旧的 `--radius` / `--radius-lg` 已映射到 `--r-md` / `--r-lg`。**清除你目录里所有写死的圆角像素值**（胶囊 999px 与 50% 圆形除外），换成 token。
4. 其余页面改动按原型：首页 Hero 高度随内容（约 min(78vh, 720px)，去掉现在约 1300px 的空壳）、最新文章「首篇大卡 + 缩略列表」、文章列表统计副标题 + 吸顶搜索 chip + 按月分组、详情页 TOC 信号线 + 阅读进度、随想 segmented + 右栏日历、路由切换用同形骨架替代整页缩小 + 全屏遮罩。
5. 待修：`design/BACKLOG.md` 中 Hero 卡片着色阴影硬切问题（首页线负责）。

## 已完成的地基（勿改，只可追加）

- `client/src/styles/tokens.scss`：强调 token（`--lift --lift-hi --lift-fg --lift-shadow --tint --solid --on-solid --ink --solid-hover --btn-shadow --shadow-card --shadow-card-hover --shadow-pop --focus --line --line-2 --fill --fill-2 --fill-3 --elev --text-3 --scrim`）与圆角 token、`--font-mono`、`--ease-sheet`。
- `client/src/themes/derive.ts`：底色掺主色降到 2–3%；按 WCAG 派生 `solid / onSolid / ink`。
- 配置 `theme.radius`（前后端默认 10）。
- 光影封面公共组件 `client/src/components/common/CoverArt.vue`（props `src? seed thumb?`），桌面与移动共用。

## 共享约定（壳层线实现，其他线直接用类名）

- 骨架：给任意元素加 `class="sk"` 即渲染为同形骨架块（shimmer，尊重 reduced-motion）；文字行用 `class="sk sk-line"`。
- 入场：容器加 `class="rise-stagger"`，其直接子元素依次 rise(14px) 淡入，间隔 55ms；单个元素用 `class="rise"`。
- 页面数据就绪：沿用 `stores/loading.ts` 现有的 hold / release 门闩接口（壳层线可以重做实现，但保持接口名不变或在报告里说明迁移）。

## 目录归属

| 线 | 负责 |
|---|---|
| 壳层 | `components/layout/**`、`components/loading/**`、`components/ui/**`、新建 `components/search/**`、新建 `components/common/SiteFooter.vue`、`App.vue`、`router/index.ts`（仅 loading 钩子）、`stores/loading.ts`、`styles/base.scss`、`styles/motion.scss`、`main.ts` |
| 首页 | `views/HomeView.vue`、`components/home/**`（含 hero 全部） |
| 内容 | `views/ArticlesView.vue`、`views/ArticleView.vue`、`views/ThoughtsView.vue`、`components/post/**`、`components/thoughts/**`、`components/media/**`、`composables/useNotesFeed.ts`、`components/common/CoverArt.vue`、`styles/highlight.scss` |
| 关于 | `about/**`、`views/AboutView.vue`、`components/admin/modules/**`、`components/admin/ModulePicker.vue` |
| 移动 | `components/mobile/**`、`views/mobile/**`（含 `admin/`）、`components/mobile-admin/**`、`router/mobile-*.ts` |
| 后台 | `views/admin/**`、`views/write/**`、`components/admin/*.vue`（除 `modules/` 与 `ModulePicker.vue`）、`styles/admin.scss`、`router/admin.ts` |

共享文件（`api/index.ts`、`stores/*`（loading 除外）、`i18n/locales/zh-CN.ts`、`styles/tokens.scss`）只允许追加。

## 质量与自检

- 与原型对齐或更好；深浅 + 四季色盘（重点看秋黄：按钮字应为深色）都要验证；圆角在 `theme.radius` = 4 / 10 / 18 三档下都成立（可在浏览器里 `document.documentElement.style.setProperty('--r-base','4')` 快速检查）。
- 全仓 `grep -n "text-shadow"` 与主色渐变填充 / 彩色外发光在你的目录里应清零（品牌封面光影除外，并在代码里注释说明）。
- `cd client && pnpm exec vue-tsc --noEmit -p tsconfig.app.json`：你的文件零错误，忽略别人正在改的文件。
- 浏览器验证：开发服务器已在 http://localhost:5173（后端 3100）运行，直接用它；Playwright 在 scratchpad：`require('C:/Users/BEACON~1/AppData/Local/Temp/claude/E--Users-BeaconCat-Downloads-Myself/912c003d-238c-4dfd-b291-4c78a8fc534e/scratchpad/node_modules/playwright')`。截图放 `design/round3/impl-shots/<线名>/`，至少两轮迭代。后台登录 admin / myself-admin。**不要重启或关闭 5173 / 3100 上的服务**；测试产生的数据结束前删除，不要改开发库里的站点配置（需要时用 Playwright route 拦截 site-config）。
- 严禁 emoji、原生 alert/confirm/prompt、CDN、「待完善」类文案。
- 不要 git commit；结束时报告改动文件、未完成项、集成注意事项。

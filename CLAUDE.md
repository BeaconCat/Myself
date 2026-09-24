# Myself — 个人博客（内部代号 Myself）

## 总体
- Monorepo：`client/`（Vue 3 + Vite 7 + PNPM）+ `server/`（Go 1.26 net/http + SQLite，纯 Go 无 cgo）；旧 Node 实现归档于 `server-node/`
- 遵循 Google 开发规范；代码简洁专业、易维护
- 内容统一 Markdown 存储与渲染
- 每个功能开发完 → 运行测试服务器人工测试 → 通过后 commit 到 GitHub 私有仓 `BeaconCat/Myself`

## 前端规范
- 依赖全部离线 npm 包（pnpm 安装），严禁 CDN / 联网拉取资源
- 严禁 emoji；图标用离线安装的 iconfont
- 严禁原生 alert/confirm/prompt：一律使用全局模态组件
  （`stores/dialog.ts` + `components/ui/AppModal.vue`，带灵动进出场动画）
- 字体：思源黑体（UI/正文）+ 思源宋体（标题/文章），离线 woff2 打包
- i18n：vue-i18n，当前仅 zh-CN（`i18n/locales/zh-CN.ts`）；多语言暂缓，新增文案一律走字典不写死
- 桌面：顶部毛玻璃胶囊导航。移动端（≤767px）不是响应式缩放，而是独立外壳：前台 `components/mobile/MobileShell.vue`（悬浮玻璃胶囊底栏 + 跟手侧滑抽屉 + iOS 式 push/返回手势 + 骨架屏），后台 `views/mobile/admin/`（贴底底栏 + 中央「+」写作）；路由用命名视图 `mobile`，映射在 `router/mobile-public.ts` / `router/mobile-admin.ts`
- 后台：Studio 设计（`views/admin/`，token 在 `styles/admin.scss` 的 `.studio` 作用域，底色/纸面随主题色 3–8% 交叠）；子路由契约 `router/admin.ts`（name 稳定）
- 首页 Hero：文字动效与卡组动效为两个独立注册表（`components/home/hero/choreo/{text,card}`），配置 `hero.textAnim` / `hero.cardAnim` 自由搭配，后台「外观」页用 `HeroMixer.vue` 实时预览
- 关于页：模块化组件库（`about/`，26 种模块，12 栏 bento，`span/variant/title/hidden`），旧配置经 `about/migrate.ts` 读取时迁移；默认模块唯一来源 `about/default-modules.json`

## 主题系统
- 两层：`mode`（light/dark）×`palette`（季节色盘）
- 品牌常量（所有主题贯穿）：`--accent-red:#ff0032` `--accent-yellow:#ffb300` `--accent-blue:#0078ff`
- 内置四季预设：spring / summer / autumn / winter，由后端站点配置 `theme.presets` 下发（每组仅 primary / primaryDeep）
- 色盘运行时派生：`client/src/themes/derive.ts` 从主色推导整套 light+dark 变量并注入，不维护独立色盘文件；后台可自定义最多 10 组预设
- 全部通过 CSS variables 注入 `:root[data-mode][data-palette]`

## 动画规范（全站）
- 全站要有丰富流畅动画：页面路由过渡、滚动入场（IntersectionObserver reveal）、
  悬停微动效（translateY/辉光）、3D 卡片倾斜、毛玻璃层
- 统一 motion tokens：时长 `--dur-fast:.2s / --dur:.35s / --dur-slow:.6s`，
  缓动 `--ease-out:cubic-bezier(.2,.8,.3,1)`、回弹 `--ease-spring:cubic-bezier(.2,.8,.3,1.2)`
- 尊重 `prefers-reduced-motion`
- 首页 Hero 轮播：左文右卡，切换编舞见上文注册表；组内多卡数秒轮转

## 后端规范
- RESTful API，前缀 `/api/v1`
- JWT 管理员认证（后台编辑）；API 中心：APIKey 认证（`X-Api-Key`），供外部 AI 发文
- SQLite 存储（modernc.org/sqlite），文章正文为 Markdown
- 布局：`cmd/myself-server` 入口，`internal/{store,auth,config,imaging,httpapi}`；标准库 ServeMux 方法路由
- 图片处理纯 Go（stdlib + gen2brain/webp），备份 archive/zip；本机无 gcc，勿引入 cgo 依赖
- 启动：根目录 `pnpm dev`（concurrently 拉起 client + `go -C server run`）；`go test ./...`
- 生产：`pnpm build` → client 产物输出 `server/web/dist` → go:embed 单二进制托管 SPA + API；`/feed` RSS；`/uploads/thumbs/<name>.webp` 按需缩略图
- 前端 Markdown 统一走 `client/src/utils/markdown.ts`（highlight.js 离线 + TOC 锚点），请求统一经 `client/src/api/index.ts`

## 分期
- P0 骨架：主题系统、i18n、响应式布局、motion 系统、Hero 轮播 ✅进行中
- P1 博客核心：列表/详情/标签/随想（原归档已并入随想），Markdown 渲染 + 代码高亮 + TOC + RSS ✅
- P2 后台：登录 + 帖文 CRUD（TipTap 富文本，Markdown 存储）✅
- P3 API 中心：APIKey 管理 ✅
- P4 预留：用户系统、评论

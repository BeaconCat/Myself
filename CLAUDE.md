# Myself — 个人博客（内部代号 Myself）

## 总体
- Monorepo：`client/`（Vue 3 + Vite 8 + PNPM）+ `server/`（Go 1.26 net/http，纯 Go 无 cgo）。生产前端、字体与 SQLite 驱动均内置在二进制中，运行时不依赖系统安装 SQLite。
- 遵循 Google 开发规范；代码简洁专业、易维护
- 内容统一 Markdown 存储与渲染
- 每个功能开发完 → 运行测试服务器人工测试 → 通过后 commit 到 GitHub 私有仓 `BeaconCat/Myself`

## 前端规范
- 依赖全部离线 npm 包（pnpm 安装），严禁 CDN / 联网拉取资源
- 严禁 emoji；严禁手绘 SVG 图标：线性图标统一用离线 `lucide`（成套具名图标走各自组件 SIcon / MaIcon / MIcon / UiIcon / ContentIcon / EngageIcon / KitIcon 的映射表，零散图标用 `components/ui/Icon.vue` 直接传 Lucide 图标），品牌 logo 用离线 `simple-icons`（注册表 `about/brands.ts`）；图表、地图、插画不算图标
- 严禁原生 alert/confirm/prompt：一律使用全局模态组件
  （`stores/dialog.ts` + `components/ui/AppModal.vue`，带灵动进出场动画）
- 字体：思源黑体（UI/正文）+ 思源宋体（标题/文章），离线 woff2 打包
- i18n：vue-i18n，当前仅 zh-CN（`i18n/locales/zh-CN.ts`）；多语言暂缓，新增文案一律走字典不写死
- 桌面：顶部毛玻璃胶囊导航。移动端（≤767px）不是响应式缩放，而是独立外壳：前台 `components/mobile/MobileShell.vue`（悬浮玻璃胶囊底栏 + 跟手侧滑抽屉 + iOS 式 push/返回手势 + 骨架屏），后台 `views/mobile/admin/`（贴底底栏 + 中央「+」写作）；路由用命名视图 `mobile`，映射在 `router/mobile-public.ts` / `router/mobile-admin.ts`
- 后台：Studio 设计（`views/admin/`，token 在 `styles/admin.scss` 的 `.studio` 作用域，底色/纸面随主题色 3–8% 交叠）；子路由契约 `router/admin.ts`（name 稳定）
- 设计语言：抬升 + 轻染、实底主按钮、无彩色发光、圆角 token `--r-*`，基准由 `theme.radius` 配置；高密度排版使用统计条、大号等宽数字、撑满区域的图表与收紧的留白。
- 首页 Hero：文字动效与卡组动效为两个独立注册表（`components/home/hero/choreo/{text,card}`），配置 `hero.textAnim` / `hero.cardAnim` 自由搭配，后台「外观」页用 `HeroMixer.vue` 实时预览
- 关于页：模块化组件库（`about/`，26 种模块，12 栏 bento，`span/variant/title/hidden`），旧配置经 `about/migrate.ts` 读取时迁移；出厂模块 `about/default-modules.json`（仅身份 / 统计 / 格言），演示模块 `about/demo-modules.json`（初始化选了 Demo 才写入数据库），修改时同步 Go 配置默认值与 `server/internal/store/demo_about.json`；演示数据不得写死在组件里。
- 留言墙模块是真实评论（target=guestbook），模块只存展示选项
- 站点身份：头像 / 形象图 / 名片头图（`banner`，默认关）/ 名字 / 别名 / 签名 / 自述 / 状态 / 链接 / 格言存于 `about` 顶层，是全站唯一来源（`about/identity.ts`，展示值统一取 `about/useIdentity.ts`），后台「身份」页（`views/admin/AdminIdentityView.vue`）编辑；profile / motto 模块只存展示选项（kicker、收尾装饰），渲染时 `injectIdentity` 注入内容，勿把内容写回模块；链接的 `card` 标记决定名片按钮（最多 3 个，桌面与移动共用 `cardLinks`）

## 主题系统
- 三维正交：`mode`（light/dark）×`palette`（季节色盘）×`style`（`clean` 透明背景简洁，默认 / `cards` 高密度卡片，可选）；风格 token 在 `styles/tokens.scss` 末尾（`--card-*` `--statbar-*` `--section-gap` 等），站点默认 `theme.defaultStyle`，`theme.allowUserStyle` 控制访客切换
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
- 交接原则：loading 退场与页面入场重叠（揭幕开始即放开页面动画，`stores/loading.ts` 的 curtain）；揭幕前须目标页已挂载 + 数据门闩（holdRoute）清零 + 首帧渲染，6s 兜底；后台内部切页离场与入场同时进行，不播遮罩
- 首页 Hero 轮播：左文右卡，切换编舞见上文注册表；组内多卡数秒轮转

## 后端规范
- RESTful API，前缀 `/api/v1`
- 会话：HttpOnly Cookie `myself_session`（JWT，`sub` 用户 id + `tv` 令牌版本），写操作要带 `X-Requested-With: myself` 且 Origin 与 Host 一致（开发代理保持浏览器 Host）；中间件 `requireRole(admin|author|reader)`
- 用户系统（`users.*` 配置）：总开关 `enabled` → 读者（注册 open/invite/closed、邮箱验证）/ 协作作者（直接发布）/ 评论（审核 all/first/none、匿名）各自独立；访客回应 `reactions` 不受总开关约束；登录：密码 / GitHub OAuth / 站长生成重置链接，发信走 `mail.*` SMTP；前端路由守卫按角色预判（`stores/auth.ts` 的 `canEnterAdmin`），作者只进文章 / 写作 / 评论（写作时可用素材库选取与上传，素材页的管理操作只给站长）
- 互动：回应（喜欢 / 灵感 / 会心 / 共鸣，访客 Cookie 去重）与评论（post / note / guestbook）；前端 `stores/engage.ts` 批量取摘要，`components/engage/`（EngageBar、CommentSection）
- API 中心：APIKey 认证（`X-Api-Key`），供外部 AI 发文
- SQLite 存储（modernc.org/sqlite），文章正文为 Markdown
- 布局：`cmd/myself-server` 入口，`internal/{store,auth,config,imaging,httpapi}`；标准库 ServeMux 方法路由
- 图片处理纯 Go（stdlib + gen2brain/webp），备份 archive/zip；本机无 gcc，勿引入 cgo 依赖
- 启动：根目录 `pnpm dev`（concurrently 拉起 client + `go -C server run`）；`go test ./...`
- 生产：`pnpm build` → client 产物输出 `server/web/dist` → go:embed 单二进制托管 SPA + API；`/feed` RSS；`/uploads/thumbs/<name>.webp` 按需缩略图
- 前端 Markdown 统一走 `client/src/utils/markdown.ts`（highlight.js 离线 + TOC 锚点），请求统一经 `client/src/api/index.ts`

## 安全与首次启动
- 不内置默认管理员：首次启动进入 `/setup` 向导（初始化码 → 管理员 → 站点与身份 → 可选 Demo 数据），初始化码只打印在服务端启动日志；旧库仍用历史默认口令时强制改密（`admin_must_change`）
- JWT 7 天、带 `tv` 令牌版本，改密即全部吊销；登录 / 初始化码按 IP 失败 5 次锁 15 分钟（反向代理后设 `MYSELF_TRUST_PROXY=1`）
- API Key 两档：`contrib` 仅投稿（默认，只写草稿、只碰自己创建的草稿）/ `full` 全托管
- 响应头：页面 CSP 由 `server/web/web.go` 按 index.html 内联脚本哈希生成（勿新增内联脚本）；`/uploads/` 只按单段安全文件名直出，不列目录；配置里的外链统一走 `utils/safeUrl.ts` 的 `safeHref`
- `backups/`、`uploads/`、`data/` 不入库；Demo 数据在 `server/internal/store/seed.go`
- 默认封面：已生成的资源位于 `client/public/covers/NN(-s).webp`，无封面时经 `utils/defaultCovers.ts` 按种子稳定取图。

## 分期
- P0 骨架：主题系统、i18n、响应式布局、motion 系统、Hero 轮播 ✅进行中
- P1 博客核心：列表/详情/标签/随想（原归档已并入随想；随想详情 `/thoughts/:id`，列表保活回到原位置），Markdown 渲染 + 代码高亮 + TOC + RSS ✅
- P2 后台：登录 + 帖文 CRUD（TipTap 富文本，Markdown 存储）✅
- P3 API 中心：APIKey 管理 ✅
- P4 用户系统与评论 ✅（后台用户 / 评论页、设置 · 邮件与登录方式、前台 `/account/*`）

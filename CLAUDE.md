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
- i18n：vue-i18n，默认 zh-CN，字典结构预留其他语言
- 响应式：PC 顶部毛玻璃胶囊导航；移动端悬浮顶栏 + 汉堡侧栏

## 主题系统
- 两层：`mode`（light/dark）×`palette`（季节色盘）
- 品牌常量（所有主题贯穿）：`--accent-red:#ff0032` `--accent-yellow:#ffb300` `--accent-blue:#0078ff`
- 内置四季色盘：spring / summer / autumn / winter，每套含 light+dark 两组变量
- 色盘定义为独立文件（`client/src/themes/*.ts`），支持后续新增自定义色盘
- 全部通过 CSS variables 注入 `:root[data-mode][data-palette]`

## 动画规范（全站）
- 全站要有丰富流畅动画：页面路由过渡、滚动入场（IntersectionObserver reveal）、
  悬停微动效（translateY/辉光）、3D 卡片倾斜、毛玻璃层
- 统一 motion tokens：时长 `--dur-fast:.2s / --dur:.35s / --dur-slow:.6s`，
  缓动 `--ease-out:cubic-bezier(.2,.8,.3,1)`、回弹 `--ease-spring:cubic-bezier(.2,.8,.3,1.2)`
- 尊重 `prefers-reduced-motion`
- 首页 Hero 轮播：左侧大标题+简介（文字随机模糊切入/飞出），右侧封面卡 3D 飞入飞出，
  组内多卡数秒轮转，每 10 秒切换一组内容

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

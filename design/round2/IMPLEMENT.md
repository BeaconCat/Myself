# 第二轮设计落地 · 并行开发约定

五条线同时在同一工作树里开发 Vue 前端。**严格只改自己名下的目录**；需要改共享文件时只改下表标明你可以改的那一段。别人的文件出现类型错误时忽略，只保证你自己的文件零错误。

## 用户定稿（必须遵守）

1. **后台**：用 `design/round2/admin-b-studio/index.html` 的 Studio 设计。但不要锁死暖白：桌面底色、纸面、侧栏、阴影都要跟当前主题色轻微交叠（选冬蓝就是浅浅的蓝调，选春绿就是浅绿调；深色模式同理）。用 `color-mix(in oklab, var(--primary) N%, <基色>)` 派生，N 在 3–8% 之间。
2. **移动端**：用 `design/round2/mobile/index.html` 的设计语言。前台用 V1 悬浮玻璃胶囊底栏，后台用 V2 贴底底栏 + 中间凸起「+」。**补上页面切换动画与 loading**（路由切换过渡、数据加载骨架屏、下拉刷新手感）。
3. **关于页**：用 `design/round2/about-kit/index.html` 的组件库。**删掉身份区右侧的门与门光**，改成：右侧是一张圆角的用户头像/形象图，左边缘向左做不透明度渐变融入背景（mask-image 线性渐变），图片与渐变方向可在后台配置；未配置时用站点 logo。
4. **首页 Hero**：左侧文字动效与右侧卡组动效**完全独立**，两个注册表任意搭配，站点配置 `hero.textAnim` / `hero.cardAnim` 选择。后台「外观」页提供实时预览混搭器。
   - 文字动效取 `design/round2/hero-motion/index.html` 中 1、2、6、7、8 各组的文字部分。
   - 卡组动效取 1、2、6、7、8 各组的卡片部分，其中：
     - **1 门缝开合**的离场改为：最前面的卡片以**左边缘**为轴向内开，后面两张卡片以**右边缘**为轴向内开。
     - **8 推门见光**的门改为**向外开**（朝观者方向），不要向内推。

## 目录归属

| 线 | 负责目录 / 文件 |
|---|---|
| Hero | `client/src/components/home/hero/**`、`client/src/components/home/HeroCarousel.vue`、新建 `client/src/components/home/hero/HeroMixer.vue` |
| 关于 | `client/src/about/**`、`client/src/views/AboutView.vue`、`client/src/components/admin/modules/**`（模块编辑器）、`client/src/components/admin/ModulePicker.vue`；可改 `stores/config.ts` 中 `about` 相关类型与 FALLBACK、`server/internal/config/config.go` 默认 JSON 的 `about` 段 |
| 桌面后台 | `client/src/views/admin/**`（`AdminRoot.vue` 除外）、`client/src/views/write/**`、`client/src/components/admin/*.vue`（`modules/` 与 `ModulePicker.vue` 除外）、`client/src/styles/admin.scss`、`client/src/router/admin.ts` |
| 移动前台 | `client/src/components/mobile/**`、`client/src/views/mobile/*.vue`（不含 `admin/`）、`client/src/router/mobile-public.ts` |
| 移动后台 | `client/src/views/mobile/admin/**`、`client/src/components/mobile-admin/**`、`client/src/router/mobile-admin.ts` |

共享文件（`api/index.ts`、`stores/*`、`i18n/locales/zh-CN.ts`、`styles/base.scss|tokens.scss|motion.scss`、`App.vue`、`router/index.ts`、`main.ts`）：**只允许追加**，不要改动或删除别人已有的内容；i18n 文案请放到以你的线命名的新顶层键下（如 `heroLab`、`aboutKit`、`studio`、`mobile`、`mobileAdmin`）。

## 已有契约

- `client/src/composables/useDevice.ts`：`useDevice().isMobile`（767px 断点）。
- `App.vue`：移动端且非 bare 路由时渲染 `components/mobile/MobileShell.vue`，它内部用 `<router-view name="mobile">`。
- `router/index.ts`：前台路由 name 为 `home / articles / article / thoughts / about`；每条路由的 `mobile` 命名视图取自 `router/mobile-public.ts` 的映射，缺省回落桌面组件。
- 后台 `/admin` 由 `views/admin/AdminRoot.vue` 按设备选择 `views/admin/AdminLayout.vue`（桌面，渲染默认 `<router-view>`）或 `views/mobile/admin/MobileAdminLayout.vue`（移动，渲染 `<router-view name="mobile">`）。子路由在 `router/admin.ts` 登记，**name 是稳定契约**：`admin-today`（/admin）、`admin-posts`、`admin-write-post`（/admin/write/post?id=）、`admin-write-note`、`admin-notes`、`admin-media`、`admin-about`、`admin-appearance`、`admin-settings`、`admin-apikeys`、`admin-data`、`admin-comments`、`admin-users`；登录页 `/admin/login`。
- Hero 注册表契约：`client/src/components/home/hero/choreo/types.ts`（`ChoreoMeta`、`TextChoreoId`、`CardChoreoId`、默认 `lightscan` + `hinge`）。Hero 线需导出 `TEXT_CHOREOS` / `CARD_CHOREOS`（`choreo/index.ts`）以及 `HeroMixer.vue`（props：`text`、`card`，事件 `update:text`、`update:card`；自带示例文章数据，可独立渲染预览舞台与播放控制）。桌面后台在「外观」页直接引入 `HeroMixer`。
- 站点配置：`hero.textAnim`、`hero.cardAnim` 已加入前后端默认配置。
- 后端 API 全部在 `client/src/api/index.ts`；评论、用户功能后端尚未实现，界面按预留功能画完整（示例数据），不要写「待完善」类文案。

## 质量要求

- 与对应 HTML 原型的视觉与交互保持一致或更好；深浅模式 + 四季色盘都要验证。
- 全部动效尊重 `prefers-reduced-motion`；用 transform / opacity。
- 严禁 emoji、原生 alert/confirm/prompt（用 `stores/dialog.ts`）、CDN。字体用已安装的 `@fontsource` 思源黑/宋；等宽数字可用系统等宽栈。
- 自检：`cd client && pnpm exec vue-tsc --noEmit -p tsconfig.app.json`（你的文件零错误）；然后启动 `cd server && go run ./cmd/myself-server`（端口 3100 已占用时用 `PORT=31xx`，每条线用不同端口：Hero 3111、关于 3112、桌面后台 3113、移动前台 3114、移动后台 3115）配合 `cd client && pnpm exec vite --port 51xx`（同上 5111–5115，并把 vite 代理指向你的后端端口：用环境变量 `VITE_API=http://localhost:31xx`，vite.config.ts 已支持）进行浏览器截图验证（Playwright 在 scratchpad 已安装，见 BRIEF.md），至少两轮迭代。截图放 `design/round2/impl-shots/<你的线名>/`。
- 不要 git commit；结束时报告改动文件清单、未完成项、需要集成方注意的事项。

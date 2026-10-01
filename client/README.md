# Myself 前端

Vue 3 + TypeScript + Vite 8，使用 Pinia、Vue Router、Vue I18n 和 Tiptap。

完整功能、截图和部署见 [项目 README](../README.md)，组件展示见 [关于页图鉴](../docs/about-modules.md)。

运行 `pnpm install --frozen-lockfile` 安装前端依赖，`pnpm build` 执行类型检查与构建，产物写入 `../server/web/dist`。回到根目录执行 `pnpm dev` 同时启动前后端。本目录的 `pnpm dev` 只启动 Vite。

开发地址默认 `http://localhost:5173`，API 代理目标默认 `http://localhost:3100`，可用 `VITE_API` 覆盖。

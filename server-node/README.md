# server-node（已归档）

旧版 Node.js（Express + better-sqlite3 + sharp）后端实现，已于 2026-09 全量迁移到 Go（见 `../server`）。
保留仅作对照与回滚参考，不再维护、不参与 `pnpm dev`。

如需临时启动：`pnpm install && PORT=3100 node src/index.js`（数据目录需自行指向 `../server/data`）。

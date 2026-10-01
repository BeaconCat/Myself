# myself-server（Go）

Myself 博客后端：Go 1.26 + `net/http`，内置 SQLite（modernc），可选 MySQL；发行构建使用纯 Go 路径，无 cgo 或动态库依赖。

完整功能、截图、初始化和部署见 [项目 README](../README.md)。首次执行 Go 命令前，先在仓库根目录运行 `pnpm -C client build` 生成 `web/dist`。

```
cmd/myself-server   入口
internal/store      SQLite/MySQL 连接、建表、快照、种子数据与行模型
internal/auth       管理员凭据（scrypt，与旧 Node 哈希兼容）+ JWT
internal/config     站点配置（settings.site_config，深合并 + 旧字段迁移）
internal/imaging    纯 Go 图片解码/裁切/压缩（png/jpeg/gif/webp）
internal/httpapi    /api/v1 路由与全部 handler、/uploads 静态服务（thumbs/ 按需缩略图）、/feed RSS
internal/updater    GitHub 版本检查、二进制校验、替换、重启检查与失败回退
web                 go:embed 内嵌前端构建产物（client 构建输出到 web/dist），SPA 回落
```

运行：`go run -tags nodynamic ./cmd/myself-server`。可用 `--config` 指定 JSON 运行配置，缺省沿用环境变量和当前工作目录。
环境变量：`PORT`（默认 3100）、`MYSELF_ROOT`（覆盖数据根目录）、`MYSELF_CONFIG`（配置文件）。
测试：`go test -tags nodynamic ./...`。设置 `MYSELF_TEST_MYSQL_ADDR`（可附带 `MYSELF_TEST_MYSQL_USER`、`MYSELF_TEST_MYSQL_PASSWORD`）可让 HTTP 回归套件改用隔离 MySQL 测试库，测试账户需有创建/删除测试数据库权限。

生产部署：根目录 `pnpm build` → 前端产物进 `server/web/dist` → Go 编译为单二进制 `server/bin/myself-server`，
直接托管前端 + API + 素材，无需额外反代。

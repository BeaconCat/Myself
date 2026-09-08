# myself-server（Go）

Myself 博客后端：Go 1.26 + `net/http` + SQLite（modernc，纯 Go，无 cgo）。

```
cmd/myself-server   入口
internal/store      SQLite 连接、建表迁移、种子数据、行模型
internal/auth       管理员凭据（scrypt，与旧 Node 哈希兼容）+ JWT
internal/config     站点配置（settings.site_config，深合并 + 旧字段迁移）
internal/imaging    纯 Go 图片解码/裁切/压缩（png/jpeg/gif/webp）
internal/httpapi    /api/v1 路由与全部 handler、/uploads 静态服务
```

运行：`go run ./cmd/myself-server`（工作目录即数据根：`data/`、`uploads/`、`backups/`）。
环境变量：`PORT`（默认 3100）、`MYSELF_ROOT`（覆盖数据根目录）。
测试：`go test ./...`

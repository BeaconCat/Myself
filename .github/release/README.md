# Release 维护说明

工作流 `release.yml` 在推送 main、创建 PR 和手动触发时生成预览包，在推送 `vMAJOR.MINOR.PATCH` 或 `vMAJOR.MINOR.PATCH-rc.1` 标签时创建 Release。包含连字符的版本标记为预发布，自动更新器只使用稳定版。

所有平台共享一次前端构建，然后以 `CGO_ENABLED=0`、`-tags nodynamic` 交叉编译。SQLite、WebP 的纯 Go 实现和前端都内置在程序中。10 个目标分别提供原始二进制与完整部署压缩包，附上 SHA256SUMS 和更新清单。

发布前的门禁：

- 前端类型检查和生产构建。
- SQLite 全量 Go 测试、MySQL 8.4 HTTP 回归套件。
- Linux amd64 原生、386/ARM/ARM64/RISC-V64 的 QEMU 安装包启动与 SQLite/图片处理测试。
- Windows amd64/386/ARM64、macOS amd64/ARM64 的原生安装包测试。
- Linux 上真实二进制替换和故障候选版本的数据回退演练。

只有上述检查都通过，标签构建才上传 GitHub Release。构建使用只读权限，发布 job 单独取得 contents:write。现有 Release 不自动覆盖；重试前应先检查失败位置。

## 本地打包

安装前端依赖并构建后：

```sh
python .github/release/package_release.py --version v1.0.0
```

版本号用于演示，实际发布时替换为准备发布的版本。也可以指定单个目标和空输出目录：

```sh
python .github/release/package_release.py --version v1.0.0 --target windows-amd64 --output dist-release
python .github/release/smoke_release.py dist-release windows-amd64
```

打包器只复制明确列出的程序、模板、文档与许可证，并创建空数据目录，不会递归复制开发环境数据。它从锁定的 npm 依赖和 Go 模块收集第三方许可证，缺少文本会中止打包。补充许可证的来源记录在 `deploy/licenses/README.md`。

## 更新协议

`release-manifest.json` 的 `schemaVersion` 与 `updateProtocol` 当前均为 1。每个原始二进制条目包含 `os`、`arch`、`kind: binary`、`size` 与 `sha256`。程序通过 `--version-json` 输出版本、提交、系统、架构与更新协议，安装前与清单逐项核对。

维护协议 1 时，后续数据库迁移必须保持可重入和向后恢复能力。破坏兼容的迁移应改变协议并提供单独迁移步骤，不能让旧更新器直接安装。更新失败会先恢复程序，再由旧程序恢复升级前整站备份；成功启动并通过检查后才开放写入。

正常 GitHub 更新源固定在运行配置中，后台接口不接受任意下载 URL。私有仓库 token 只从 `MYSELF_UPDATE_TOKEN` 环境变量读取，跨主机下载重定向不携带该 token。

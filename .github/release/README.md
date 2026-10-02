# Release 维护说明

工作流 `release.yml` 在推送 main、创建 PR 和手动触发时生成预览包，在推送 `v` 开头的数字版本标签时创建 Release，例如 `v0.0.1`、`v0.1.0.9`。数字版本支持 2–16 段；为兼容旧产物也接受 `-rc.1` 等后缀，但后缀与 GitHub 预发布标记均不参与数字比较。

开发代号通过仓库 Actions 变量 `RELEASE_CODENAME` 设置，手动运行时也可填写 `codename`。CI 自动生成 `Myself Astra v0.1.0.9` 这样的标题；已存在的 Release 会保留其标题。发布后可以任意改名，更新判断以清单中的数字版本为准。

所有平台共享一次前端构建，然后以 `CGO_ENABLED=0`、`-tags nodynamic` 交叉编译。SQLite、WebP 的纯 Go 实现和前端都内置在程序中。10 个目标分别提供原始二进制与完整部署压缩包，附上 SHA256SUMS 和更新清单。

发布前的门禁：

- 前端类型检查和生产构建。
- npm 官方漏洞库检查与 Go 可达漏洞扫描；Go 最低版本为 1.26.7。
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

自定义版本与开发代号：

```sh
python .github/release/package_release.py --version 0.1.0.9 --codename Astra --repository your-name/Myself
```

`--repository` 默认读取 `GITHUB_REPOSITORY`，本地构建则从 origin 推断；Fork 无需把更新地址留在上游仓库。程序数字版本和代号编入二进制，不能通过编辑后台显示值伪造。Release 的任意标题与标签不会覆盖清单版本；默认 CI 使用数字标签构建，自定义非数字标签时需自行指定 `--version` 构建并上传产物。

打包器只复制明确列出的程序、模板、文档与许可证，并创建空数据目录，不会递归复制开发环境数据。它从锁定的 npm 依赖和 Go 模块收集第三方许可证，缺少文本会中止打包。补充许可证的来源记录在 `deploy/licenses/README.md`。

## 更新协议

`release-manifest.json` 的 `schemaVersion` 与 `updateProtocol` 当前均为 1。每个原始二进制条目包含 `os`、`arch`、`kind: binary`、`size` 与 `sha256`。程序通过 `--version-json` 输出版本、提交、系统、架构与更新协议，安装前与清单逐项核对。

维护协议 1 时，后续数据库迁移必须保持可重入和向后恢复能力。破坏兼容的迁移应改变协议并提供单独迁移步骤，不能让旧更新器直接安装。更新失败会先恢复程序，再由旧程序恢复升级前整站备份；成功启动并通过检查后才开放写入。

更新源可在后台选择 GitHub 仓库，接口不接受任意二进制下载 URL。版本列表按 GitHub 的发布顺序分页，检查更新会完整扫描（最多 2,000 条）后按数字大小选取；超出范围或检查中断不会把部分结果当作最新版本。清单缓存减少重复 API 请求，安装前重新读取并校验目标 Release。

私有仓库 token 可从后台配置，保存在数据根的私有 `.updates/preferences.json`，不通过状态接口回传；无后台保存值时使用 `MYSELF_UPDATE_TOKEN` 环境变量。跨主机下载重定向不携带 token。程序中的时区数据随二进制内置，凌晨检查不依赖目标机器安装 Go 或时区数据库。

# Myself 部署包

前端、字体、SQLite 和图像处理均随可执行文件提供，不必展开前端或单独安装动态库。默认运行无需 Node.js、Go 编译器或外置数据库；也可连接 MySQL。

```text
myself                     # Windows 为 myself.exe
config.example.json        # 可选运行配置模板
config.mysql.example.json  # MySQL 配置示例
runtime/
  data/                    # SQLite 数据库与站点配置
  uploads/                 # 上传素材
  backups/                 # 站点备份
LICENSE
NOTICE
THIRD_PARTY_LICENSES.txt
README.md
```

## 首次安装

下载适合系统和 CPU 架构的部署包，校验 SHA256SUMS 后解压。进入解压后的目录：

Linux / macOS：

```sh
cp config.example.json config.json
./myself --config ./config.json
```

Windows PowerShell：

```powershell
Copy-Item config.example.json config.json
.\myself.exe --config .\config.json
```

访问 `http://localhost:3100`，从服务日志取得初始化码，完成 `/setup` 向导。不要用模板覆盖已有配置。

运行配置包含端口、持久化根目录、数据库连接和更新源；站点名称、外观、账户等保存在选定数据库中，由网页管理。

```json
{
  "port": "3100",
  "root": "./runtime",
  "database": { "driver": "sqlite" },
  "updates": { "repository": "", "disabled": false }
}
```

`root` 的相对路径以配置文件所在目录为基准。配置文件选择顺序为 `--config` → `MYSELF_CONFIG` → 当前工作目录的 `config.json`。显式指定但不存在的文件会报错，默认文件不存在时沿用原来的启动行为。

`PORT`、`MYSELF_ROOT` 环境变量优先于文件；没有指定数据根时使用当前工作目录。无需模板也能启动单独二进制。是否进入初始化向导由数据库中的初始化状态决定，删除配置文件不会重置站点。

SQLite 无需连接信息。选择 MySQL 时，先创建 MySQL 8.0.19+ 的专用空库和有建表/读写权限的账户，然后在向导中填写连接信息，或复制 MySQL 配置模板。`tls: "true"` 会验证服务器证书。运行配置含数据库连接凭据，请单独保管；站点密码以哈希保存于数据库中。端口为 1–65535 的字符串。

完整站点备份可在 SQLite 与 MySQL 之间恢复。迁移时先保存原站点备份，初始化新数据库，再恢复备份，最后用原账号登录。恢复只替换站点数据与素材，不更改当前数据库连接配置。

## 服务与反向代理

长期运行请使用 systemd、Windows 服务管理工具或进程管理器。固定配置文件和数据根的路径。程序提供 HTTP，公网 HTTPS 可交给反向代理。

在可信代理后可设置 `MYSELF_TRUST_PROXY=1`，并限制后端端口直接对公网开放。代理保留 Host 并设置 X-Forwarded-For、X-Forwarded-Proto。在后台填写正式站点地址。

Linux systemd 示例（先创建 `myself` 服务用户并设置目录权限）：

```ini
[Unit]
Description=Myself personal blog
After=network.target

[Service]
User=myself
Group=myself
WorkingDirectory=/opt/myself
ExecStart=/opt/myself/myself --config /opt/myself/config.json
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```

## 更新与回退

后台“设置 → 系统与更新”支持自定义 GitHub 仓库、分页版本列表，以及选择兼容版本进行升级、重新安装或回退。数字版本逐段比较，支持 `0.0.1`、`0.1.0.9` 等格式；开发代号与 Release 标题可自由命名，不参与比较。

程序会下载适合本机架构的二进制，核对清单、SHA-256、数字版本、代号与提交，然后进入维护状态，备份整站、替换程序、重启并检查数据库连接。目标版本启动失败时自动恢复切换前的程序与备份；正常回退保留当前内容，并暂停自动升级。

页面会显示自动识别的服务器系统和架构，Release 说明按 Markdown 展示。每次成功切换后，原程序保存在 `.updates/history/<记录编号>/program.bak`，可在“更新历史与回滚”中分页查看并直接回滚，无需下载。默认保留 3 份，可改为 1、3、5、10 或 20 份；减少保留数量只清理旧程序备份，不会删除 `backups/` 中的整站备份。本地回滚同样先校验和备份，保留当前内容，暂停自动更新。

自动更新和邮件订阅可独立开启，按站点时区每日 00:00 检查，停机错过时在恢复后补查。订阅要求 SMTP 已启用并配置完整，还需填写站长邮箱。相同仓库与数字版本的通知在发信前持久化去重；发送失败或状态不确定时不会自动重发，以免重复。偏好和去重记录在 `.updates/` 中，不随站点备份恢复而回退。

“更新通道”同时控制版本列表、检查更新、自动更新与邮件订阅：正式版通道只接收 GitHub 未标记为 Pre-release 的发布；预览版通道同时接收预发布和正式版，再按数字版本选取更新。新安装默认正式版，已有配置升级时保留此前包含预发布的范围。当前 Alpha 需选择预览版通道。通道变化在保存后生效，不会自动降级或重装相同数字版本。

运行配置的仓库留空时使用构建来源，Fork 的 CI 构建默认使用 Fork 仓库。后台修改后的偏好优先，切换仓库会先暂停自动更新。无 Release、私有仓库权限不足、缺少平台包等情况有独立提示。

自动更新需要对程序目录、配置文件与数据根有写权限。Linux/macOS 更新保持服务主 PID；Windows 直接运行的程序通过辅助进程完成替换。若 Windows 服务包装器会自动重启退出的进程，应设 `MYSELF_DISABLE_SELF_UPDATE=1` 或 `updates.disabled: true`，由服务管理器执行下面的手动步骤。

更新日志位于数据根的 `.updates/install.log`、`.updates/server.log`。Windows 重启后的服务日志写入该目录，Linux/macOS 最终服务继续使用原来的日志通道。突发断电或更新辅助进程被强制结束时，可结合这里的状态、已下载程序与整站备份手动恢复。

私有仓库可在后台“系统与更新”填写有该仓库读取权限的 GitHub Update Token，也可通过 `MYSELF_UPDATE_TOKEN` 环境变量提供。后台保存值优先且立即生效；密钥不回显，留空保留原值，清除保存值后回退到环境变量（若有）。后台密钥保存在数据根的私有 `.updates/preferences.json`，不随站点备份导出。公开仓库可以不配置，受 GitHub 匿名请求限额约束；大量历史版本建议配置读取 token。检查按所选通道筛选，再比较数字版本。持久化运行的开发构建可以手动安装发行版；自动更新需要数字版本构建，`go run` 的临时程序需先构建并运行持久化文件。同一数字版本重新发布后，在版本列表点击“重新安装”获取新程序。

手动升级：

1. 在后台生成并下载整站备份，保留旧程序。
2. 停止服务。
3. 下载对应平台的新二进制，校验 SHA256SUMS，替换程序，保留 `config.json` 和完整数据根。
4. Linux/macOS 确保程序有执行权限：`chmod +x myself`。
5. 用原来的配置启动，检查版本、登录、内容和素材。

Windows 必须先停止进程再替换。不要把新部署包整体解压覆盖已有数据。数据库会自动迁移；回退时应使用与旧程序匹配的数据备份。

`./myself --version`（Windows 为 `.\myself.exe --version`）显示版本、提交和构建时间。

Release 还提供单独二进制、`release-manifest.json` 与 `SHA256SUMS`。清单列出版本、提交、更新协议、系统、架构、文件类型、字节数与 SHA-256，自动更新器使用其中的单独二进制。SHA-256 用于校验下载完整性，不是数字签名。

发行目标：Windows 386/amd64/arm64，macOS amd64/arm64，Linux 386/amd64/arm（ARMv6）/arm64/riscv64。每个目标都内置 SQLite。不存在对应 Go 目标的系统与架构组合不生成安装包。

完整使用说明：https://github.com/BeaconCat/Myself

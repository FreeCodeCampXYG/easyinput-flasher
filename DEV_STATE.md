# EasyInput Flasher 开发状态

## 2026-08-31：S3 验身与 Release 拉取修复

- ESP32-S3 下载端口的只读复核已确认可连通；旧实现错误调用 `chip-id`，该子命令在 S3 上会在读取 MAC 后返回非零，导致 UI 将有效端口误报为工具失败。
- 现改为 `read-mac` 与 `flash-id` 的 ROM 兼容组合，写入同样使用受限的 `--no-stub` 路径；仍只接受 manifest 声明的三段镜像，且保留人工确认与写前验身。
- 默认及旧版设置均使用用户指定的 `http://127.0.0.1:1080`，烧录页启动时自动请求 `FreeCodeCampXYG/easy-input-maker` 的 Release；读取失败会进入运行日志和界面提示，不再静默显示空列表。
- 已核验 `firmware-v0.2.1` Release 含 `firmware-manifest.json`、bootloader、分区表和应用镜像；实际下载、写入及恢复仍待新版桌面包与实板验证。
- 已调研纯 Go 候选 `tinygo.org/x/espflasher v0.8.1`：其声明支持 ESP32-S3 / USB-JTAG；由于项目依赖治理库无该条目，本轮未将其加入正式依赖或发布包。

## 当前目标

- 构建六平台 Wails 烧录器首版：从已验证的 GitHub Release 下载固件，完成设备验身、人工确认、写入、恢复检查与脱敏日志。

## 已完成

- 主题系统：默认恢复协调的深色工作台；侧栏新增“浅色/深色主题”切换并用浏览器本地存储保持选择，浅色主题整体覆盖主区与控件，不再出现半明半暗混搭。
- 正常设备扫描：Windows PnP 同时匹配 USB VID/PID 与本机实际枚举的 `EasyInput AI` BLE 正常模式；正常模式仅作为 BOOT 前确认，进入下载模式后才允许芯片/MAC 验身。
- 烧录流程改为分阶段人机交互：正常 HID 识别后提示短按并松开 BOOT，用户点击刷新下载端口后才进入 ESP32-S3/MAC 验身；正常 HID 不再被当作检测失败。
- 明亮主题：将近黑色背景改为浅灰/白色工作台，提升侧栏、正文、控件、日志和底部状态栏对比度，保留青绿/琥珀/红色状态语义。
- Windows 硬件检测：扫描同时识别正常 HID（VID 303A / PID 1006）和下载模式串口（PID 1001）；正常 HID 仅展示设备，不开放烧录，进入 BOOT 后才允许 ESP32-S3/MAC 验身。
- 创建独立仓库骨架，未修改 `easy-input-maker` 固件工作区。
- 后端已实现 manifest 目标校验、三段镜像 SHA-256 校验、公开 GitHub Release 读取、下载缓存、串口扫描、纯 Go ESP32-S3 协议适配、MAC 脱敏和确认写入状态机。
- Go 1.27 已安装；Go 全量测试与 `go vet`、前端类型检查和生产构建均通过，Wails Windows x64 原生构建成功。
- 已使用用户提供图标作为应用图标源文件。
- 已补齐 Windows x64 CI/标签 Release 工作流、Issue/PR 模板、项目 flow/docs 协作骨架；Wails 结构审计 11 PASS / 3 WARN / 0 FAIL。
- 发布矩阵已扩展为 Windows x64/ARM64、macOS Intel/Apple Silicon、Linux x64/ARM64；每个平台原生构建 Wails，烧录协议编译进主程序，并生成归档、SHA-256 和构建溯源。
- 自有代码采用 PolyForm Noncommercial 1.0.0，版权归 StarLine；固件、`espflasher` 和其他依赖的上游许可证已在 `THIRD_PARTY_NOTICES.md` 分开记录。

## 2026-08-31：官方 main 固件 Release 已发布

- 已回读 `firmware-v0.2.3-main`：公开 Release、Actions 成功，包含 bootloader、partition-table、application、manifest 和 SHA256SUMS。
- manifest 核对通过：`board=easyinput-v2`、`chip=esp32s3`、`idfVersion=5.5.5`、`commit=846114041ed7df6dfb393cb939798c30ea491b1f`，三段偏移符合烧录器合同；当前只验证了 Release/清单层，尚未完成桌面端下载、实板写入和 HID 恢复。

## 已知边界

- `FreeCodeCampXYG/easy-input-maker` 已有 `firmware-v0.2.1` 的受信 manifest Release；新版桌面包尚待实板验证下载、写入和恢复结果。
- 纯 Go 烧录器已完成本地 Windows Wails 编译；六平台 CI、发布包以及真实设备写入/HID 恢复仍待验证。当前无 Python、PyInstaller 或 GPL helper 分发义务。
- Wails 结构审计 13 PASS / 1 WARN；唯一 WARN 是首版不提供 NSIS 安装器，仅提供 Windows 便携 ZIP。
- 前端主题和 Windows 检测改动已通过 npm typecheck/build、Go test/vet 和 `git diff --check`；真实 HID/串口设备回归仍待实板。
- 本轮新增 BOOT 引导提示与“刷新下载端口”动作，前端 typecheck/build、Go test/vet、差异检查通过；尚未推送或发布。
- 提交 `49e7ead` 已推送 main；Go test/vet、前端 typecheck/build 和 `git diff --check` 通过，未创建新 Release tag。
- 提交 `c7a0009` 已推送 main：正常 HID/BLE 只确认设备在线，芯片字段留待下载模式 esptool 验身；BOOT 引导不再显示为红色错误。Go/前端验证通过，未创建新 Release tag。
- 当前电脑没有 Git Bash，Unix 打包脚本只能完成静态检查，须以 GitHub macOS/Linux native runner 作为实际验证证据。
- `v0.1.0` Release 的 Linux 资产构建成功，macOS 两架构在默认 Python 上构建 PyInstaller helper 失败；`v0.1.1` 已修复并验证 macOS/Linux/Windows x64 资产，Windows ARM64 因 cryptography 缺少原生 OpenSSL wheel 失败，Release workflow 现改用 x64 Python/helper，待新 tag 验证。
- `v0.1.3` 六个平台 package jobs 全部成功，publish job 因未 checkout 仓库而无法让 `gh release create --notes-file` 找到 `.git` 失败；新增 `publish-existing-release.yml` 复用该 run 的 artifacts 发布，避免重跑六平台构建。
- `v0.1.3` 发布后复现发现 Windows x64 包实际为 ARM64：Windows 打包脚本硬编码 archive 名称，恢复发布时同名资产覆盖。已修复脚本按架构命名，并让恢复 workflow 分 artifact 下载、仅对旧 ARM64 资产做名称兼容转换；不重编译，待以新 tag/恢复发布验证。
- 一次 Wails 重建曾因先前手动启动的同名可执行文件占用 `build/bin` 失败；精确结束该项目进程后重建成功。构建前必须退出正在运行的应用。

## 下一步

1. 在 Maker `main` 合入固件发布 workflow/manifest 脚本，并创建绑定 main SHA 的 `firmware-v-main-*` Release。
2. 在烧录器新包中验证官方 main 基线显示、下载哈希、受信门禁和实板恢复 HID。
3. 后续再实现精确 GitHub Release URL 导入与逐版本社区审批，不开放分支、裸 bin 或任意 URL。

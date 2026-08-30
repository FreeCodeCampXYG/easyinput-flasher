# EasyInput Flasher 开发状态

## 当前目标

- 构建六平台 Wails 烧录器首版：从已验证的 GitHub Release 下载固件，完成设备验身、人工确认、写入、恢复检查与脱敏日志。

## 已完成

- 创建独立仓库骨架，未修改 `easy-input-maker` 固件工作区。
- 后端已实现 manifest 目标校验、三段镜像 SHA-256 校验、公开 GitHub Release 读取、下载缓存、串口扫描、esptool 命令白名单、MAC 脱敏和确认写入状态机。
- Go 1.27 已安装；Go 全量测试与 `go vet`、前端类型检查和生产构建均通过，Wails Windows x64 原生构建成功。
- 已使用用户提供图标作为应用图标源文件。
- 已补齐 Windows x64 CI/标签 Release 工作流、Issue/PR 模板、项目 flow/docs 协作骨架；Wails 结构审计 11 PASS / 3 WARN / 0 FAIL。
- 发布矩阵已扩展为 Windows x64/ARM64、macOS Intel/Apple Silicon、Linux x64/ARM64；每个平台原生构建 Wails 与 esptool helper，并生成归档、SHA-256 和构建溯源。
- 自有代码采用 PolyForm Noncommercial 1.0.0，版权归 StarLine；固件、esptool 和其他依赖的上游许可证已在 `THIRD_PARTY_NOTICES.md` 分开记录。

## 已知边界

- 当前没有可用的 `firmware-manifest.json` Release，因此不能真实下载或烧录。
- esptool helper、六平台 CI、发布包与真实设备写入尚未验证；GPL helper 的对应源代码与分发义务仍需在正式发布前复核。
- Wails 结构审计 13 PASS / 1 WARN；唯一 WARN 是首版不提供 NSIS 安装器，仅提供 Windows 便携 ZIP。
- 当前电脑没有 Git Bash，Unix 打包脚本只能完成静态检查，须以 GitHub macOS/Linux native runner 作为实际验证证据。
- `v0.1.0` Release 的 Linux 资产构建成功，macOS 两架构在默认 Python 上构建 PyInstaller helper 失败；`v0.1.1` 已修复并验证 macOS/Linux/Windows x64 资产，Windows ARM64 因 cryptography 缺少原生 OpenSSL wheel 失败，Release workflow 现改用 x64 Python/helper，待新 tag 验证。
- 一次 Wails 重建曾因先前手动启动的同名可执行文件占用 `build/bin` 失败；精确结束该项目进程后重建成功。构建前必须退出正在运行的应用。

## 下一步

1. 在 GitHub Actions 完成六平台 CI 与 Release 演练，验证各平台打包的 esptool helper。
2. 推送 Maker 的 manifest 工作流后，以 `firmware-v*` 完成首次云端固件发布。
3. 在真实 Windows 和 EasyInput V2.0 上验证设备验身、写入、取消、恢复 HID 与功能边界。
4. 审查第三方 notices 后，按 Git 治理提交、推送、标注 `v0.1.0` 与创建桌面端 Release。

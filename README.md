# EasyInput Flasher

EasyInput Flasher 是 EasyInput V2.0 的桌面烧录工具。它从受信任 GitHub Release 下载已构建固件，在本机校验版本化清单和 SHA-256 后，识别 ESP32-S3 下载模式设备并以显式确认写入。

它不编译固件、不上传设备数据、不执行 OTA，也不会默认擦除 Flash。

## 当前状态

首版开发中，目标平台是 Windows x64。GitHub Actions 负责构建桌面应用；尚未产生可安装 Release，也尚未完成实机烧录器验证。

## 安全流程

1. 仅显示带 `firmware-manifest.json` 的 Release。
2. 清单必须声明 `easyinput-v2`、`esp32s3` 和完整的三段写入范围。
3. 下载后逐段 SHA-256 校验，失败时不启用烧录。
4. 扫描串口后，通过 esptool 只读读取 ESP32-S3 信息和 MAC 尾号。
5. 必须输入界面显示的精确确认文本，才会开始写入。
6. 写入结束后提示恢复正常启动，并检查预期 USB HID 是否重新出现。

烧录成功、HID 恢复和具体功能正常是三种不同证据，界面会分别呈现。

## 使用前提

- EasyInput V2.0 / ESP32-S3。
- 支持数据传输的 USB 线。
- 当自动连接失败时：开发板保持开机，短按并松开一次 BOOT；烧录后关机再正常开机。
- 固件 Release 必须包含工具认可的 `firmware-manifest.json` 和三段镜像。

不要按住 BOOT 再上电，不要把 GPIO0 当作业务按键，不要把擦除整片 Flash 当作常规操作。

## 下载与平台说明

Release 提供六种原生构建资产：Windows x64/ARM64 便携 ZIP、macOS Intel/Apple Silicon `.app` ZIP、Linux x64/ARM64 TAR.GZ。请选择与系统架构一致的文件，并先校验同名 `.sha256` 文件。

Windows x64 是当前唯一完成本机构建验证的桌面目标；其余五个平台会由 GitHub Actions 原生 runner 打包，欢迎使用者提交安装、串口权限、下载模式和烧录结果。CI 打包成功不等于真实设备烧录已经验证。

macOS 包在未使用 Apple Developer ID 签名和 notarization 前会被系统标记为未验证开发者。使用者需要在系统“隐私与安全性”中自行确认是否允许运行，并按系统提示授权 USB/串口访问；应用不会绕过 Gatekeeper、SIP 或任何系统权限。

## 文档

- [架构](docs/ARCHITECTURE.md)
- [本地开发与构建](docs/BUILDING.md)
- [发布流程](docs/RELEASING.md)
- [故障排查](docs/TROUBLESHOOTING.md)
- [变更日志](CHANGELOG.md)

## 许可证

自有代码采用 [PolyForm Noncommercial 1.0.0](LICENSE)，版权归 StarLine，仅允许该许可证定义的非商业用途。固件和烧录辅助程序保留其各自上游许可证，详见 [第三方声明](THIRD_PARTY_NOTICES.md)。

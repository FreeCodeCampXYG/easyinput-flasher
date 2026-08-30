# 发布流程

1. 确认许可证、版本、变更日志和 Git 工作树。
2. 推送经过审阅的 `vX.Y.Z` annotated tag。
3. GitHub Actions 在六个原生 runner 构建：Windows 为便携 ZIP，macOS 为 `.app` ZIP，Linux 为 TAR.GZ；每项资产生成 SHA-256 与构建溯源。esptool helper 使用 `actions/setup-python` 固定 Python 3.13，并通过 `python -m` 调用，避免平台启动器选择其他 Python。
4. 下载公开资产，重新计算 SHA-256 并检查压缩包内容。
5. 在真实 Windows 和 EasyInput V2.0 上分别验证：安装、设备识别、错误芯片拒绝、下载校验、烧录、恢复 HID、取消和异常拔线后的提示。

Windows、macOS 和 Linux 资产在没有证书、Developer ID 或 notarization 时均必须标记为未签名。macOS 使用者可能需要在系统隐私与安全设置中明确允许未签名应用运行，并为 USB/串口访问授予系统要求的权限；这不是本项目可替用户绕过的权限。不要移动已发布 tag；修复应使用新的递增版本。

# 发布流程

1. 确认许可证、版本、变更日志和 Git 工作树。
2. 推送经过审阅的 `vX.Y.Z` annotated tag。
3. GitHub Actions 在六个原生 runner 构建：Windows 为便携 ZIP，macOS 为 `.app` ZIP，Linux 为 TAR.GZ；每项资产生成 SHA-256 与构建溯源。烧录协议由固定版本的纯 Go `espflasher` 编译进主程序，不再构建或携带 Python helper。
4. 下载公开资产，重新计算 SHA-256 并检查压缩包内容。
5. 在真实 Windows 和 EasyInput V2.0 上分别验证：安装、设备识别、错误芯片拒绝、下载校验、烧录、恢复 HID、取消和异常拔线后的提示。

如果某一平台卡住或失败，可取消原 Release run，使用 `Retry Platform Build` 只重跑该平台；随后在 `Publish Existing Release` 填入原 run、重试 run 和缺失 artifact 名称，将两次 run 的六平台资产合并为同一个已有 tag 的 Release。不要重推 tag 或重复构建已成功的平台。

例如 Linux x64 缺失时，先在 `Retry Platform Build` 选择原 tag 和 `linux-x64`；待该 run 成功后，在 `Publish Existing Release` 填入原 Release run ID、重试 run ID 以及 `easyinput-flasher-linux-x64`。发布前会检查六个平台 artifact 都存在，缺任一项即失败，不会发布不完整资产。

Windows、macOS 和 Linux 资产在没有证书、Developer ID 或 notarization 时均必须标记为未签名。macOS 使用者可能需要在系统隐私与安全设置中明确允许未签名应用运行，并为 USB/串口访问授予系统要求的权限；这不是本项目可替用户绕过的权限。不要移动已发布 tag；修复应使用新的递增版本。

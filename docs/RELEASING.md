# 发布流程

1. 确认许可证、版本、变更日志和 Git 工作树。
2. 推送经过审阅的 `vX.Y.Z` annotated tag。
3. GitHub Actions 当前只在 Windows x64 runner 构建便携 ZIP，并生成 SHA-256 与构建溯源。烧录协议由固定版本的纯 Go `espflasher` 编译进主程序，不再构建或携带 Python helper。
4. 下载公开资产，重新计算 SHA-256 并检查压缩包内容。
5. 在真实 Windows 和 EasyInput V2.0 上分别验证：安装、设备识别、错误芯片拒绝、下载校验、烧录、恢复 HID、取消和异常拔线后的提示。

如果 Windows x64 package job 已成功而 publish job 失败，使用 `Publish Existing Release` workflow 指定原 Release run ID 和 tag 复用已有 artifact，不要为了发布重新运行构建。

当前 Windows x64 资产未使用代码签名，必须标记为未签名。不要移动已发布 tag；修复应使用新的递增版本。

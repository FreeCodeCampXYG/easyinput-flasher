# 决策日志 (decisions)

## 2026-09-01 · 不自动拆分社区 factory 镜像
- 背景: `CY-CHENYUE/easy-input-maker` 当前 Release 只有整片 `factory.bin`，缺少 Flasher 合同要求的 manifest 与三段镜像。
- 决定: 发现页识别并提示 factory 镜像，但不在桌面端自动拆分、猜测偏移或把它伪装成标准包；补齐由固件仓库 Actions 生成标准 Release。
- 原因: 整片镜像边界可能覆盖配置分区，自动补齐会扩大不可逆写入风险；同时保持桌面端不集成本地编译器的范围。

## 2026-08-31 · 正式切换为纯 Go `espflasher`
- 背景: 用户授权依赖准入并要求移除嵌套 Python 烧录助手；依赖治理查询此前未命中记录，但本轮以用户明确授权作为项目准入决定。
- 决定: 固定 `tinygo.org/x/espflasher v0.8.1`，由 Go 后端直连 ESP32-S3 ROM/stub；手动 BOOT 后使用 `ResetNoReset`，写入仍只消费 manifest 固定的三段镜像。
- 影响: 删除 esptool/PyInstaller、GPL helper、相关打包步骤和许可证副本；保留 SHA-256 下载校验、MAC 尾号确认、取消和关机再开机恢复门禁。

## 2026-08-31 · ESP32-S3 采用 ROM 兼容验身与写入路径
- 背景: 实板 COM 下载端口可连接，但旧版先执行 `chip-id`；ESP32-S3 对该子命令会读 MAC 后返回非零，界面因此误判为烧录工具故障。
- 决定: 继续使用受控 esptool helper，但对 S3 只执行 `--no-stub read-mac` 和 `--no-stub flash-id`；写入同样以 `--no-stub` 执行 manifest 固定的三段镜像。
- 否决的方案 & 原因: 本轮不直接引入 `tinygo.org/x/espflasher v0.8.1`。它是有前景的纯 Go 替代项，但依赖治理查询未发现已准入记录，不能以修复故障为由绕过依赖准入与实板回归。

> 记"过程决策 + 为什么"。**追加,不删改**——下一棒最值钱的上下文。

## 2026-08-30 · 六平台原生构建，烧录证据按平台分层
- 背景：用户希望提供 Windows、macOS 和 Linux 包，但当前只在 Windows 进行过本机构建，未在任何平台完成真实设备烧录。
- 决定：CI/Release 使用 Windows x64/ARM64、macOS Intel/Apple Silicon、Linux x64/ARM64 的原生 runner；各自产生 ZIP/TAR.GZ、SHA-256 和构建溯源。
- 边界：六平台 CI 只证明构建和归档；Windows x64 以外的串口、权限和烧录均标为社区待验证。macOS 未签名包需要使用者自行按系统安全设置授权。

## 2026-08-30 · 首版限定 Windows x64 与受信 Release
- 背景：烧录器涉及本机串口、原生 helper 与物理设备，跨平台声明或任意 GitHub 链接都会扩大未验证风险。
- 决定：首版仅发布 Windows x64；仅接受具有 `firmware-manifest.json` 的可信来源，Go 后端从清单生成写入参数并在本轮 MAC 尾号确认后写入。
- 否决的方案 & 原因：不在前端执行浏览器串口烧录，也不手写 ESP ROM 协议或把任何 GitHub 搜索结果默认设为可信。
> (架构 / 产品级的"为什么"按 `flow/规范/文档维护SOP.md` 进 `AGENTS.md`;这里记项目怎么推进的过程决策。)

## 2026-08-31 · 社区合并采用分层门禁而非机器人直接写入
- 背景：项目需要开放 Issue/PR 共建，但固件写入与发布改动具有较高设备和供应链风险。
- 决定：使用 DBX 风格的类型/模块/平台/优先级标签，GitHub Actions 负责自动标注、PR 合同检查和 CodeQL；GitHub 原生 auto-merge 仅在维护者添加 `automerge` 标签、人工审核和所有检查通过后执行。
- 边界：Codex/Copilot 等 AI 审核只提供建议，不能替代批准、CODEOWNERS 或实机验证；不静默安装读取全部仓库的第三方机器人。

## 2026-08-31 · 官方 main 与社区固件采用同一 Release 合同
- 决定：桌面端继续只下载预编译 Release，不集成 ESP-IDF/Python；main 恢复通过 Maker 主分支 CI 产出绑定提交 SHA 的独立 Release。
- 决定：社区包沿用 manifest + 三段镜像 + SHA-256；后端校验精确受信来源和固定写入偏移，不能只依赖前端按钮状态。

<!-- 模板:
## YYYY-MM-DD · <决策标题>
- 背景:
- 决定:
- 否决的方案 & 原因:
-->

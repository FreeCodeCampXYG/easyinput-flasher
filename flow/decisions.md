# 决策日志 (decisions)

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

<!-- 模板:
## YYYY-MM-DD · <决策标题>
- 背景:
- 决定:
- 否决的方案 & 原因:
-->

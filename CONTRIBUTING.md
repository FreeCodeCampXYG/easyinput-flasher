# 贡献指南

谢谢你帮助 EasyInput Flasher 变得更可靠。此项目涉及真实设备写入，欢迎所有改进，但任何新功能都不能绕过 manifest、设备验身、确认文本或恢复启动边界。

## 先选对入口

- 使用 **Bug report**：设备识别、下载、校验、写入、恢复或界面行为异常。
- 使用 **Build report**：GitHub Actions、本地构建、打包或平台依赖问题。
- 使用 **Firmware source proposal**：申请评估一个社区/Fork 的标准 Release 来源；只接受 GitHub Release + manifest，不接受裸 bin、分支名或私有下载链接。
- 使用 **Feature request**：讨论新体验或产品能力。
- 使用 **Question**：使用、兼容性或开发疑问。

维护者会补充四类标签：`type:`、`area:`、`platform:` 和 `priority:`。请不要为安全问题创建公开 Issue，改按 [SECURITY.md](SECURITY.md) 的私下报告渠道处理。

## 提交 Pull Request

1. 从最新 `main` 创建主题分支，保持一个 PR 只解决一个问题。
2. 不提交固件二进制、缓存、完整 MAC、代理配置、Token 或其他凭据。
3. 在本机运行与改动相符的验证；界面改动附截图，设备行为改动写清未验证的实机边界。
4. 完整填写 PR 模板。自动标签与 CI 会先运行，随后由非作者维护者审核。

## 合并规则

- PR 必须通过六平台质量检查和 CodeQL 安全扫描。
- 写入、固件来源、manifest、发布或工作流变更需要 CODEOWNERS 审核。
- AI 审核意见用于发现风险，不是人工批准的替代品，也不能自行合并代码。
- 仅维护者加上 `automerge` 标签后，GitHub 原生自动合并才会在所有门禁满足时生效。

## 本地验证

`npm --prefix frontend ci`

`npm --prefix frontend run typecheck`

`npm --prefix frontend run build`

`go test ./...`

`go vet ./...`

`wails build -clean -trimpath -platform windows/amd64`

本机构建、CI 构建、实际写入、HID 恢复和具体功能验证必须分别报告，不能相互替代。

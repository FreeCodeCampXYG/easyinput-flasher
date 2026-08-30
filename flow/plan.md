# 计划 (plan) —— 契约

> 经确认后执行。要偏离,**先改这里**再动手。

## 里程碑
- [ ] M1: 后端安全合同、前端工作台与 Windows CI 构建通过。
- [ ] M2: Maker Release 清单/哈希/溯源资产就绪，并完成真实设备烧录回归。
- [ ] M3: 用户确认许可证、提交、推送 `main` 和 `v0.1.0`，创建首个 Release。

## 任务拆解
| 任务 | 负责角色 / 工具 | 输入 | 产出(落哪个文件) | 验收标准 |
|---|---|---|---|---|
| T01 | Codex | EasyInput 烧录合同与 GitHub Release | Wails 后端、Actions、文档 | Go 测试、前端构建、CI YAML 和安全状态机检查通过 |
| T02 | Sol worker | 用户 UI 参考与绑定契约 | `frontend/` | 静态 UI 覆盖烧录、固件、发现、日志与状态栏 |
| T03 | Codex | Maker 现有构建工作流 | firmware manifest 与 Release 门禁 | 静态验证和 GitHub Actions 成功 |
| T04 | 用户 + Codex | Windows 设备 | 实机验身、写入、恢复证据 | 不把写入成功误写为功能验证 |

## 实时进展 / 交接棒
→ 见 `flow/进展.md` 顶部(每棒收工在那追加一条:做了什么 / 为什么 / 产出路径 / 下一步)。
（plan.md 只管"计划=契约";"现在到哪了"在进展日志,不在这儿覆盖。）

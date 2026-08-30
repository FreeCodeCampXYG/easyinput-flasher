# Third-party Notices

根目录 [LICENSE](LICENSE) 仅适用于 EasyInput Flasher 的自有 Go 代码、前端代码、脚本、文档和经用户授权的应用图标。它不替代、不收窄或覆盖下列第三方和上游材料的许可证。

| 组件 | 用途 | 许可证与处理 |
| --- | --- | --- |
| EasyInput firmware | GitHub Release 下载的固件 | 固件始终保留其来源仓库的许可证和 notices；桌面端不把它重新许可为 StarLine 作品。 |
| esptool 5.3.0 | 独立的串口烧录辅助程序 | GPL-2.0-or-later；随正式包提供 `LICENSES/GPL-2.0-or-later.txt`、版本与官方来源，不纳入根 PolyForm 许可证。 |
| Wails v2 | Go 桌面框架 | MIT；由 Go module lock 固定。 |
| React、Vite、Lucide | 本地前端 UI 与构建 | 各自许可证由 npm lock 固定；发布包保留所需 notices。 |
| PyInstaller | 生成独立 esptool helper | GPL-2.0 with commercial exception；只用于构建辅助程序，具体边界随其发布许可证处理。 |

## esptool source and license

- 上游：<https://github.com/espressif/esptool>
- 固定版本：`5.3.0`
- 许可证文本：`LICENSES/GPL-2.0-or-later.txt`

发布 esptool helper 前必须复核其对应源代码、依赖与 GPL 分发义务。若无法满足该义务，Release 流程必须停止，不得删除 notices 后继续发布。

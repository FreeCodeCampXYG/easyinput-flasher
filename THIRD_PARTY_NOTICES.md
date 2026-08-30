# Third-party Notices

根目录 [LICENSE](LICENSE) 仅适用于 EasyInput Flasher 的自有 Go 代码、前端代码、脚本、文档和经用户授权的应用图标。它不替代、重授权或覆盖下列第三方和上游材料的许可证。

| 组件 | 用途 | 许可证与处理 |
| --- | --- | --- |
| EasyInput firmware | GitHub Release 下载的固件 | 固件始终保留其来源仓库的许可证和 notices；桌面端不将其重新许可为 StarLine 作品。 |
| `tinygo.org/x/espflasher` 0.8.1 | 内嵌的 ESP32-S3 串口烧录协议实现 | BSD-3-Clause；版本由 Go module lock 固定，用于替代外部 Python/esptool helper。 |
| Wails v2 | Go 桌面框架 | MIT；由 Go module lock 固定。 |
| React、Vite、Lucide | 本地前端 UI 与构建 | 各自许可证由 npm lock 固定；发布包保留所需 notices。 |

## 纯 Go 烧录边界

`espflasher` 在应用进程内实现 ESP ROM/stub 通信，不再随发行包携带 Python、PyInstaller 或独立 esptool 可执行文件。它只能使用 Go 后端根据受信 manifest 生成的固定三段镜像；前端不能传入命令、偏移或任意本地路径。

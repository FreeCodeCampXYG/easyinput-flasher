# 架构

```text
Wails 前端
  -> internal/application       状态机与受控绑定
  -> internal/firmware          GitHub Release、manifest、SHA-256、缓存
  -> internal/device            串口候选、下载模式验身、HID 恢复检查
  -> internal/flasher           纯 Go ESP32-S3 协议适配、manifest 固定镜像与写入门禁
  -> internal/config            用户来源和代理设置
```

前端输入均不可信。Go 后端不接受前端传入的任意命令、偏移或本地文件路径；写入参数只能从校验过的 manifest 生成。

正常模式下 EasyInput 是 HID，不保证提供 CDC 串口。下载模式的串口仅用于本轮只读验身与写入；端口变化后必须重新识别，MAC 尾号不一致时停止任务。

应用不读取或保存完整 MAC、用户快捷键、Wi-Fi、Token 或代理凭据。设备日志只保留必要的阶段和脱敏错误摘要。

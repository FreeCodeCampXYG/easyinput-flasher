# 项目宪章 (charter)

> 立项填。所有 Agent 开工要读的第一份。

- **项目名**: EasyInput Flasher
- **目标**(做成什么样算成功,一句话): 让 EasyInput V2.0 用户无需本地编译环境，即可从经过校验的 GitHub Release 完成安全、可追踪的固件烧录。
- **范围**:
  - 做: Windows x64 Wails 应用、Release 清单与哈希校验、ESP32-S3 验身、人工确认、写入状态、HID 恢复检查、收藏与发现入口。
  - 不做: 本地固件编译、OTA、默认整片擦除、未经验证的跨平台宣称、读取或上传用户凭据。
- **约束**(时间 / 资源 / 必须遵守的): EasyInput V2.0 的 BOOT/GPIO/身份边界；GitHub Actions 云端构建；写入前每次重验设备。
- **成功标准**(尽量可衡量): CI 可产出 Windows x64 ZIP 与 SHA-256；错误板型/错误芯片/损坏固件均被拒绝；真实板完成一次受控烧录和正常 HID 恢复验证。
- **角色**:拍板 = 用户 / 主控 = Codex / 前端专项 = Sol worker / 发布确认 = 用户

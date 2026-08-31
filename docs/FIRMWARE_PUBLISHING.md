# 可烧录固件发布规范

EasyInput Flasher 不包含 ESP-IDF、Python 或本地编译器。可烧录固件必须先在固件仓库的 GitHub Actions 中完成构建，再通过公开 Release 提供；烧录器只下载 Release 资产并校验清单。

## 官方 main 基线

“还原 main”不是从 GitHub 分支下载源码，也不是恢复出厂设置。固件仓库需要在 `main` 合入发布工作流后，按以下顺序操作：

1. 从 `main` 创建不可移动的新标签，例如 `firmware-v-main-20260831.1`（匹配现有 `firmware-v*` 触发规则）。
2. Actions 固定 ESP-IDF 版本，先跑 host tests，再构建 ESP32-S3 固件。
3. 由构建脚本生成 `firmware-manifest.json`，其中 `commit` 必须等于标签指向的 main 提交。
4. Release 上传 manifest、三段镜像、`SHA256SUMS.txt` 和构建溯源证明。

Release 发布后，烧录器会自动从官方仓库列表读取它；不需要桌面端重新编译。写入前仍会重新下载并校验全部资产。

## 社区或 Fork 固件

贡献者应复用固件仓库提供的 workflow 和清单生成脚本，不要手工填写偏移，也不要上传裸 `.bin` 作为“可烧录包”。最小合同如下：

```json
{
  "schemaVersion": 1,
  "product": "easyinput-firmware",
  "board": "easyinput-v2",
  "chip": "esp32s3",
  "tag": "firmware-v1.0.0",
  "commit": "构建提交 SHA",
  "idfVersion": "5.5.5",
  "files": [
    {"name": "bootloader.bin", "offset": "0x0", "sha256": "...", "size": 0},
    {"name": "partition-table.bin", "offset": "0x8000", "sha256": "...", "size": 0},
    {"name": "easy_input_keyboard.bin", "offset": "0x10000", "sha256": "...", "size": 0}
  ]
}
```

Actions 至少应执行 host tests、ESP-IDF 5.5.5 构建、清单生成和 SHA-256 计算。功能说明只能作为发布者自声明；构建成功不等于实板功能验证。

当前版本的烧录器只允许后端已加载且标记为受信的来源进入写入流程。未来接入社区 Release URL 时，建议按单个精确 Release 审批，不把整个仓库永久升级为官方来源；未审核来源只能查看和缓存，不能写入设备。

## 用户操作提示

使用别人的固件时，优先索取公开 Release URL，确认仓库、标签、commit、目标板型和变更说明，再在烧录器中选择该版本。不要提供分支名、任意下载 URL 或裸镜像，也不要把“烧录完成”当作“功能已验证”。

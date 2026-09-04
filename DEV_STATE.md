# EasyInput Flasher 开发状态

## 2026-09-05：自动模式代理端口保存修正

- 根因：`auto` 模式只探测、不保存 UI 当前输入的代理地址；配置为空或旧地址时会导致自动探测失败，而切换 `custom` 正常。
- 修复：`ConfigureNetwork` 现在对 `auto` 与 `custom` 都归一化并保存当前端口，再执行探测。
- 验证：Go 配置/固件测试、前端 typecheck/build、`git diff --check` 通过；尚未发布新包。

## 2026-09-05：网络设置按钮样式修正

- 修复 `.about-grid button` 覆盖 `.button secondary/ghost` 的 CSS 冲突；网络设置的“保存并测试”“重新测试”恢复标准按钮边框、间距和可读性。
- 验证：前端 typecheck/build、`git diff --check` 通过；尚未提交或发布新包。

## 2026-09-05：网络设置入口与自定义选项可见性修正

- 网络设置卡片补充明确入口文案；下拉框现在同时显示自动探测、系统/全局代理、直连网络、自定义代理四项。
- 自动探测与自定义代理均启用端口输入框，避免用户误以为没有代理端口配置。
- 验证：前端 typecheck/build、`git diff --check` 通过；尚未提交或发布新包。

## 2026-09-05：修复 Release publish 幂等性

- 根因：`v0.1.24` 的六个 package job、artifact 下载和整理均成功，publish 因 Release 已预先存在而执行 `gh release create` 失败。
- 修复：`.github/workflows/release.yml` 现在先检查 Release；已存在时使用 `gh release upload --clobber` 补齐资产，不存在时才创建。
- 提交 `3e758d5`、tag `v0.1.25` 已推送；workflow `33894580410` 已入队，待回读六平台和 Release 资产。

## 2026-09-05：v0.1.24 已提交并触发发布

- 提交 `b660118` 已推送 `origin/main`，annotated tag `v0.1.24` 已推送，GitHub Release 已创建。
- Release workflow `33893360565` 正在运行；Release 页面已创建但暂时没有资产，六平台打包结果待回读。

## 2026-09-04：网络代理自动探测与可配置端口

- 后端配置新增 `auto/system/direct/custom` 网络策略；自动模式按自定义代理（默认 `127.0.0.1:1080`，可改端口）→系统/全局代理→直连探测 GitHub。
- “关于与帮助”页新增网络设置面板，可保存策略、端口并重新测试；快照展示真实 GitHub 可达性与实际采用路径。前端不构造命令或下载地址。
- 验证：`go test ./internal/config ./internal/firmware`、application 测试二进制编译、前端 typecheck/build、`git diff --check` 通过；application 测试直接执行受 Windows `Access is denied` 限制；未进行不同网络环境和实板烧录验证。

## 2026-09-02：六平台单平台重试与合并发布

- 已取消 `v0.1.22` 的 Release run `33642934892`；Linux x64 卡在 WebKit 系统依赖安装，未生成 GitHub Release。
- 已恢复 CI/Release 的 Windows x64/ARM64、macOS Intel/Apple Silicon、Linux x64/ARM64 矩阵；新增单平台重试构建和跨 run artifact 合并发布。CI 忽略纯 workflow、文档与状态记录提交，避免本次流程配置本身重跑六平台。
- 验证：全部 workflow YAML 解析、`git diff --check` 通过；Linux x64 重试 run `33647617987` 成功，合并发布 run `33648149667` 成功，`v0.1.22` Release 已包含六平台归档及六个 SHA-256 文件。

## 2026-09-02：v0.1.22 排序修复已发布

- 已确认已发布的 `v0.1.21` 在排序修复提交前创建，截图中的程序仍是该旧包；不能修改或重推既有 tag。
- 后端语义版本排序之外，前端适配层也按同一规则排序，兼容旧后端或缓存快照；下一包应显示 `v0.2.12`、`v0.2.11`、`v0.2.10`、`v0.2.9` 的正确降序。
- 验证：前端 `npm run typecheck`、`npm run build`、`go test ./internal/firmware`、定向 `go vet` 与差异检查通过；`v0.1.22` 六平台 Release 资产已回读，桌面端排序运行时复核仍待安装新包。

## 2026-09-02：固件 Release 语义版本排序

- 修复 GitHub Release 列表直接沿用 API 返回顺序的问题；后端现在按 `vX.Y.Z` / `firmware-vX.Y.Z` 的三段语义版本降序排列，`v0.2.10` 会位于 `v0.2.9` 和 `v0.2.8` 之前。
- 同版本稳定版本优先于带后缀版本；无法识别版本号时以 `publishedAt` 作为稳定回退。排序发生在后端列表边界，前端和后续调用方使用同一顺序。
- 验证：新增 `internal/firmware` 排序测试通过，前端 typecheck/build、`go vet ./...`、`git diff --check` 通过；全量 Go 测试仍仅因既有 Windows `internal/application` 测试二进制 `Access is denied` 失败。

## 2026-09-02：v0.1.21 交付计划待执行

- 已核对本轮仅提交烧录页紧凑布局、状态文档与交接日志；`main` 与 `origin/main` 当前对齐，最新远端 tag 为 `v0.1.20`，建议新增 annotated tag `v0.1.21`。
- `.codegraph/` 为本地未跟踪索引目录，必须保持排除。推送 `v0.1.21` 会自动触发六平台 Release workflow；构建、Release 创建及实板验证均尚未发生，须分别回读。

## 2026-09-02：烧录页首屏紧凑化

- 烧录页顶部留白、标题说明、六步状态条、分栏间距和左右模块间距已收紧，让右侧确认与状态信息更早出现在常见桌面高度内。
- “等待检测设备”进度区域补足左右 `12px` 内边距；该项及镜像地址、校验状态等辅助文本统一调整为 `9px`，主按钮、输入框、关键状态和设备字段字号不变。
- 验证：前端 `npm run typecheck`、`npm run build`、`git diff --check` 通过；以 `2048×1152` 本地预览截图确认右栏主按钮和日志入口位于首屏，进度条左右留白可见。未进行实板流程验证。

## 2026-09-02：v0.1.20 已推送并触发六平台发布

- 提交 `c3342fd` 已推送到 `origin/main`，annotated tag `v0.1.20` 已推送。
- GitHub Release workflow run `33581134052` 已进入 `in_progress`，负责六平台编译、打包和 Release 创建；最终成功状态待回读。

## 2026-09-02：烧录确认与诊断交互提示

- 烧录右栏改为桌面端滚动时保持可见的 sticky 控制区，确认口令增加“一键填入”，仍保留后端确认门禁。
- 烧录下载/写入/恢复阶段显示固定操作提醒；完成状态自动隐藏，提醒内容明确断电重启和禁止断连边界。
- 诊断页对“开始诊断”、教学固件切换和逐项标记增加可见状态反馈，避免点击后看不到结果；教学固件仍只改变说明，不会自动烧录。
- 验证：前端 `npm run build` 通过；Go 定向测试中 `internal/flasher`、`internal/firmware` 通过，`internal/application` 执行被 Windows `Access is denied` 拦截；`go vet ./...` 与 `git diff --check` 已执行。

## 2026-09-02：烧录段日志语义修正

- 烧录段在开始时显示“开始写入 + 地址”，完成时显示“写入完成 + 100%”，不再把开始事件显示成含义不明的 `0%`。
- 完整烧录明细根据开始/完成事件分别显示“已开始、进行中、已完成”；最近活动仍只保留预览，完整记录可从设备与日志页查看。
- 验证：前端 typecheck/build、Go 定向测试和 go vet 通过；application 测试执行仍被 Windows `Access is denied` 拦截。

## 2026-09-02：完整烧录明细面板

- 烧录页右侧新增完整烧录明细，固定列出 bootloader、partition table、application 三段镜像及地址、当前段进度、完成状态和对应烧录日志。
- 明细状态来自后端真实进度/日志字段，不再只显示总进度 100%；查看全部仍跳转设备与日志页。
- 验证：前端 typecheck/build、git diff --check 通过。

## 2026-09-02：烧录页右栏间距与日志入口

- 右侧“写入控制”栏补齐内边距和模块间距，避免完成状态、进度条和日志贴边拥挤。
- “查看全部”现在会跳转到“设备与日志”页，不再是无动作的装饰按钮。
- 验证：前端 typecheck/build、git diff --check 通过；未做桌面截图回归。

## 2026-09-02：诊断请求改用 Maker Output Report

- 已确认无诊断效果的根因：Maker 原先把 `0x13` 定义为 Feature Report，而 Go HID 依赖只能通过 Output Report 写入。
- Maker 端已将 `0x13` 改为 Output 请求；Flasher 保持 16 字节请求和现有分片响应解析。
- 需要先构建并烧录包含该 Maker 修复的新固件；旧固件不会响应当前 Flasher 的状态请求。
- 验证：Maker 状态协议单测通过；Flasher Go 定向测试、go vet、前端构建通过；application 测试受 Windows `Access is denied` 拦截，实板待验证。

## 2026-09-01：Windows ARM64 原生构建的 HID 依赖降级

- 根因：ARM64 runner 没有可用的 Windows ARM64 hidapi/CGO 编译器，`karalabe/hid` 在 Wails 原生构建阶段调用 x64 GCC，报 `gcc_arm64.S` 指令错误。
- 修复：真实 HID 诊断读取限定在 `!windows || amd64`；Windows ARM64 使用纯 Go 明确降级实现并提示未包含 hidapi 诊断能力，不伪造自动检测结果。
- 验证：`GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build ./...` 通过；最新 CI 的 ARM64 测试已通过，原生构建在该依赖路径失败，待修复提交后重新验证。

## 2026-09-01：Windows ARM64 CI 的 CGO 测试修复

- 根因：Windows ARM64 runner 的 `go test ./...` 启用 CGO 后，`karalabe/hid` 的 C 源文件被 x64 GCC 汇编，报 `stp x29,x30` 等 ARM64 指令错误。
- 修复：CI 的 Go 单测和 vet 阶段设置 `CGO_ENABLED=0`，使用 HID 库的纯 Go 不支持实现；原生 Wails 构建保持 CGO，不影响实际 Windows HID 能力。
- 验证：本地 `CGO_ENABLED=0 go test ./...` 中非 application 包通过；application 测试仍被 Windows `Access is denied` 拦截；已通过 `go vet` 与差异检查。

## 2026-09-01：正常 HID 遥测读取与真实输入联动

- 新增纯 Go HID 状态读取：复用 Maker 现有 Vendor HID `0x13` 状态请求和 `0x11/0x04` 分片回报，读取 `diag.last`、输入事件计数、旋钮步进、电池和供电状态。
- 诊断页在正常模式轮询遥测；检测到输入事件计数或旋钮步进增加时，才把对应项目更新为真实事件已收到。没有 `diag` 的旧固件保持不支持自动遥测。
- 新增 `github.com/karalabe/hid v1.0.0`，与现有 Studio 版本一致；未新增串口、WebHID 或常驻全局键盘监听。
- 验证：Flasher Go 定向测试、go vet、前端 typecheck/build、Maker 状态测试通过；Flasher application 测试执行仍受 Windows `Access is denied` 拦截；真实 HID 回报和实板功能未验证。
- Maker 状态请求新增 `diagnostics` 标志，仍复用原 16 字节请求和 512 字节响应预算；Flasher 请求该标志以稳定获取详细 `diag` 字段。

## 2026-09-01：用户学习排查动态观察台

- 诊断页在没有专用固件协议的前提下增加 S1-S8 输入观察台：用户先按实体键并确认系统/固件真实反馈，再点击对应编号记录观察次数。
- 观察台的高亮和轻微动画仅表示“已记录用户观察”，不会自动改变硬件诊断通过状态，也不会把点击 Flasher 当作 HID 事件。
- 验证：前端 typecheck/build、git diff --check 通过；未做实板输入回归。

## 2026-09-01：无线、电池与真实输入反馈诊断项

- 诊断清单新增 BLE、Wi-Fi、USB HID 和电池剩余容量项目；电池容量明确为固件基于电压/负载/校准的估算，板上没有独立电量计。
- 按键、旋钮和 USB HID 测试步骤要求用户观察真实设备反馈（系统输入、固件日志或外设响应）后再标记；Flasher 自身按钮点击不算硬件信号通过。
- BLE/Wi-Fi 只在正常运行态、且当前固件声明支持时验证；板载无线存在不等于每个固件启用。
- 验证：前端 typecheck/build、Go 定向包测试、go vet、git diff --check 通过；application 测试执行受 Windows `Access is denied` 拦截，未做实板回归。

## 2026-09-01：诊断测试步骤与 BOOT/运行态切换提示

- 诊断页明确区分下载模式自动验身与正常模式外设测试：自动验身完成后需关机再正常开机，才能测试按键、旋钮、灯光、音频和 HID。
- 为用户可操作项目增加逐项测试步骤和“看到什么算正常”的白话说明；固件未定义的业务动作仍提示以当前固件说明和实板表现为准。
- 验证：前端 typecheck/build、git diff --check 通过；未做真实设备回归。

## 2026-09-01：诊断页教学提示与固件切换

- 诊断页新增“教学固件”选择器；根据当前 Release 的 board/chip、Factory 标记、features/changelog 展示用途摘要，切换只影响教学说明，不改变烧录门禁。
- 每个板级诊断项目新增白话“怎么用”提示，内容只采用 EasyInput V2.0 板级事实和项目边界：S1-S8、编码器 A/B、WS2812、GPIO8 共享电源、USB/BLE HID、VBAT 估算等。
- 未提供固件特有功能时显示“以当前固件说明和实板表现为准”，避免把板级存在误写成产品能力；Factory 镜像明确标为恢复用途。
- 验证：前端 typecheck/build、Go 定向测试、go vet、git diff --check 通过；未做实板教学回归。

## 2026-09-01：诊断页可读性与设备检测收紧

- 诊断页按 EasyInput V2.0 的真实证据边界重排为“自动验身 / 用户操作 / 运行态观察 / 待固件验证”分组，增加推荐顺序、设备模式提示和结果汇总；未验证项仍不会伪造通过。
- 修复浅色主题诊断卡标题和说明继承近白文字的问题，补齐卡片、说明、选择框和操作按钮的对比度覆盖。
- Windows PnP 查询成功但未命中 EasyInput 时不再把系统所有 COM 口列为目标板；仅在 PnP 查询失败时保留串口候选，并继续要求 ESP32-S3 ROM 验身。
- README 增加火绒/Defender 解压隔离处理边界：先核对官方来源与 SHA-256，再仅允许已校验文件，不关闭实时防护或宽泛加白。
- 验证：前端 `npm run typecheck`、`npm run build` 通过；Go 定向包编译通过，`internal/application` 测试执行仍被 Windows `Access is denied` 拦截；未做实板检测。

## 2026-09-01：更新日志时间轴增强

- 更新与通知页采用纵向版本时间轴，当前版本 v0.1.12 以高亮节点和卡片突出显示，按日期展示版本、变更内容与规划状态。
- 深色/浅色主题均已适配当前版本节点、连接线和卡片背景；前端 typecheck/build 与差异检查通过，尚未提交。

## 2026-09-01：产品定位与更新日志展示

- 烧录页和关于页统一使用“EasyInput 固件烧录与硬件诊断工具”定位；README 同步更新产品描述。
- 更新与通知页改为展示 v0.1.10、v0.1.11、v0.1.12 的真实功能变更，以及本地工程构建扩展规划。
- 前端 typecheck/build 与差异检查通过；本轮仅修改展示文案和更新日志，未触发新提交或发布。

## 2026-09-01：硬件诊断页

- 新增“硬件诊断”页：自动验身读取 ESP32-S3 与 Flash ID，并将 PSRAM、按键、编码器、WS2812/状态灯、麦克风、扬声器、VIN/CHRG/VBAT、USB/BLE 分为自动检测、用户操作、用户观察和未知边界。
- 用户可对需操作项目标记通过/异常；未知项不会伪造通过，下载模式不足时自动检查明确阻断并提示正确 BOOT 流程。
- 验证：Go vet、application 测试编译、定向 Go 测试、前端 typecheck/build、差异检查通过；未做实板元件诊断。

## 2026-09-01：CY Factory Release 兼容与项目链接

- 关于页新增 easyinput-flasher GitHub 仓库和 CY-CHENYUE/easy-input-maker v0.4.53 Release 链接。
- GitHub Release 读取支持仅含 `factory*.bin` 的 Factory 发布：标记为恢复项，下载后固定 `0x0` 写入；来源信任仍需用户确认，写入口令独立为“确认恢复出厂 <MAC>”。
- 普通 manifest 三段 Release 合同未改变；Factory 远程兼容未做实板写入验证。

## 2026-09-01：烧录详情与页内日志

- 烧录页现在直接展示当前镜像、manifest 固定写入地址、当前段已写入/总字节数和设备端 Hash 校验方式；进度来自纯 Go `espflasher` 的真实写入回调，而非阶段估算。
- 页内“最近活动”扩展为最近六条，保留各镜像开始/完成记录；详细日志页仍可导出诊断。
- 深浅主题均为新增详情卡提供边框、背景和文本覆盖。验证：Go vet、application 测试编译、定向 Go 测试、前端 typecheck/build、差异检查通过；未实板验证。

## 2026-09-01：Factory 独立恢复入口

- 固件库支持导入单文件 Factory 镜像并标记为独立恢复项；固定写入 `0x0`，界面显示清除 NVS 配置和蓝牙绑定的风险，确认口令改为“确认恢复出厂 <MAC>”。
- 普通三段固件入口和校验合同未改变；Factory 入口仅接受本地导入的 `.bin`，未进行实板验证。
- 验证：Go vet、application 测试编译、配置/固件/烧录器测试和差异检查通过；前端构建此前已通过。

## 2026-09-01：本地 ZIP 导入与实时写入进度

- 固件库新增“导入本地 ZIP”：后端解码固定 ZIP、拒绝目录穿越、解析并校验 `firmware-manifest.json` 与三段镜像 SHA-256，成功后加入本地固件列表；不接受任意路径或偏移。
- 纯 Go 烧录调用已接入 `espflasher` 进度回调，前端可看到实时写入百分比与字节数；本地工程编译调度和 Factory 恢复入口本轮未实现。
- 验证：`go vet ./...`、application 测试编译、前端 typecheck/build、`git diff --check` 通过；application 测试执行仍被 Windows `Access is denied` 拦截，未进行实板烧录。

## 2026-09-01：社区 factory 镜像提示与内嵌说明

- 发现页审计现在会识别公开 Release 中的整片 `factory*.bin`，并明确提示不能安全自动拆分；`Ready` 仍要求 manifest + bootloader + partition table + application 四类标准资产。
- 关于与帮助页新增内嵌快速说明，覆盖 BOOT 下载模式、完整固件选择、MAC 确认、关机恢复和社区仓库文件不足时的补齐路径；不依赖外部网页才能查看。
- 已同步 `docs/FIRMWARE_PUBLISHING.md` 与 `docs/COMMUNITY_FIRMWARE_GUIDE.md`，说明整片镜像可能覆盖配置分区，补齐应由固件仓库 Actions 生成标准三段镜像与 manifest。
- 已核对 `CY-CHENYUE/easy-input-maker` 当前公开 Release 只有 `easy-input-maker-v0.4.53-esp32s3-factory.bin` 与 `SHA256SUMS.txt`，且未发现标准固件发布 workflow；未将其误标为可烧录来源。
- 验证：Go firmware/config/flasher 定向测试、`go vet ./...`、前端 typecheck/build、`git diff --check` 通过；未进行实板写入或远端仓库修改。

## 2026-08-31：首次对话设计复盘文档

- 已新增 `docs/首次对话设计复盘.md`，基于本机可访问的首次方案会话、项目进展、决策、Git 历史和项目记忆，记录需求到设计、实现边界、踩坑、复用步骤及社区接入路线。
- 文档将公开答复与用户原始提示按左右对照呈现；不包含内部推理、工具日志或敏感资料，并列出会话文件与消息序号供本机追溯。
- 内容明确区分可构建/可审计的软件状态与尚待真实板卡验证的下载、写入、HID 恢复和功能回归，未改变任何烧录、发布或设备逻辑。
- 文末新增面向 GitHub 的共创邀请，将产品定位为硬件应用 Hub；欢迎应用、兼容性、文档和测试贡献，但可写入固件仍须遵循 Release、manifest、哈希、审核和实机证据边界。
- 复盘文档已改为“复刻指南 + 设计复盘”：开头新增可按七步执行的最小闭环、目录边界、硬件/固件/发布/桌面端合同、验收顺序和复刻前检查表，历史对话与设计原因保留在后文。
- 复刻指南的写入门禁已明确要求：真正写入前再次只读验身，MAC 尾号必须与本轮确认对象一致；不一致时保留脱敏诊断并终止，避免同端口设备替换窗口。

## 2026-08-31：S3 验身与 Release 拉取修复

- ESP32-S3 下载端口的只读复核已确认可连通；旧实现错误调用 `chip-id`，该子命令在 S3 上会在读取 MAC 后返回非零，导致 UI 将有效端口误报为工具失败。
- 现改为 `read-mac` 与 `flash-id` 的 ROM 兼容组合，写入同样使用受限的 `--no-stub` 路径；仍只接受 manifest 声明的三段镜像，且保留人工确认与写前验身。
- 默认及旧版设置均使用用户指定的 `http://127.0.0.1:1080`，烧录页启动时自动请求 `FreeCodeCampXYG/easy-input-maker` 的 Release；读取失败会进入运行日志和界面提示，不再静默显示空列表。
- 已核验 `firmware-v0.2.1` Release 含 `firmware-manifest.json`、bootloader、分区表和应用镜像；实际下载、写入及恢复仍待新版桌面包与实板验证。
- 已调研纯 Go 候选 `tinygo.org/x/espflasher v0.8.1`：其声明支持 ESP32-S3 / USB-JTAG；由于项目依赖治理库无该条目，本轮未将其加入正式依赖或发布包。

## 2026-09-01：macOS 设备枚举与恢复检测

- macOS 新增专用设备扫描：串口标记为下载候选，`ioreg`/`system_profiler` 识别 EasyInput USB/Bluetooth 正常模式；恢复检查不再固定返回失败。
- 下载端口仍由 ESP32-S3 ROM 读取芯片与 MAC 进行最终验身，未放宽 manifest、哈希和人工确认门禁。
- 已通过 macOS amd64/arm64 Go 测试二进制交叉编译、Windows 本机定向测试、`go vet` 与 `git diff --check`；未在真实 Mac 或实板上验证。
- 已提交 `0e8e578` 并推送 `main`，创建并推送 annotated tag `v0.1.14`；Release/CI/CodeQL Actions 已启动，等待六平台结果。

## 当前目标

- 构建六平台 Wails 烧录器首版：从已验证的 GitHub Release 下载固件，完成设备验身、人工确认、写入、恢复检查与脱敏日志。

## 已完成

- 主题系统：默认恢复协调的深色工作台；侧栏新增“浅色/深色主题”切换并用浏览器本地存储保持选择，浅色主题整体覆盖主区与控件，不再出现半明半暗混搭。
- 正常设备扫描：Windows PnP 同时匹配 USB VID/PID 与本机实际枚举的 `EasyInput AI` BLE 正常模式；正常模式仅作为 BOOT 前确认，进入下载模式后才允许芯片/MAC 验身。
- 烧录流程改为分阶段人机交互：正常 HID 识别后提示短按并松开 BOOT，用户点击刷新下载端口后才进入 ESP32-S3/MAC 验身；正常 HID 不再被当作检测失败。
- 明亮主题：将近黑色背景改为浅灰/白色工作台，提升侧栏、正文、控件、日志和底部状态栏对比度，保留青绿/琥珀/红色状态语义。
- Windows 硬件检测：扫描同时识别正常 HID（VID 303A / PID 1006）和下载模式串口（PID 1001）；正常 HID 仅展示设备，不开放烧录，进入 BOOT 后才允许 ESP32-S3/MAC 验身。
- 创建独立仓库骨架，未修改 `easy-input-maker` 固件工作区。
- 后端已实现 manifest 目标校验、三段镜像 SHA-256 校验、公开 GitHub Release 读取、下载缓存、串口扫描、纯 Go ESP32-S3 协议适配、MAC 脱敏和确认写入状态机。
- Go 1.27 已安装；Go 全量测试与 `go vet`、前端类型检查和生产构建均通过，Wails Windows x64 原生构建成功。
- 已使用用户提供图标作为应用图标源文件。
- 已补齐 Windows x64 CI/标签 Release 工作流、Issue/PR 模板、项目 flow/docs 协作骨架；Wails 结构审计 11 PASS / 3 WARN / 0 FAIL。
- 发布矩阵已扩展为 Windows x64/ARM64、macOS Intel/Apple Silicon、Linux x64/ARM64；每个平台原生构建 Wails，烧录协议编译进主程序，并生成归档、SHA-256 和构建溯源。
- 自有代码采用 PolyForm Noncommercial 1.0.0，版权归 StarLine；固件、`espflasher` 和其他依赖的上游许可证已在 `THIRD_PARTY_NOTICES.md` 分开记录。

## 2026-08-31：官方 main 固件 Release 已发布

- 已回读 `firmware-v0.2.3-main`：公开 Release、Actions 成功，包含 bootloader、partition-table、application、manifest 和 SHA256SUMS。
- manifest 核对通过：`board=easyinput-v2`、`chip=esp32s3`、`idfVersion=5.5.5`、`commit=846114041ed7df6dfb393cb939798c30ea491b1f`，三段偏移符合烧录器合同；当前只验证了 Release/清单层，尚未完成桌面端下载、实板写入和 HID 恢复。

## 2026-08-31：固件下拉标题去重

- 前端对 Release 标题末尾重复的 tag 做显示层归一化，避免出现 `EasyInput Firmware firmware-v0.2.3-main · firmware-v0.2.3-main`；底层 `repository@tag` 标识和选择逻辑不变。
- 验证：前端 `typecheck`、生产构建和 `git diff --check` 通过；尚未推送或重新打包。

## 2026-08-31：烧录状态显示目标版本

- 后端下载、写入、失败和恢复状态现在携带选中的 `firmwareId`，并在状态文案中明确显示 manifest 绑定的 tag；异步轮询时仍能确认实际目标版本。
- 验证：Go config/firmware/flasher 测试、`go vet ./...`、前端 typecheck/build 和差异检查通过；application 测试执行仍受本机 Windows `Access is denied` 限制，尚未推送或重新打包。

## 2026-08-31：完成状态保留实际固件版本

- 修复 `CheckRecovery` 完成状态清空 `firmwareId` 的问题；恢复、完成、取消和失败状态会继承本次烧录目标，前端不会再因空值回退到列表第一项 main。
- 验证：Go config/firmware/flasher 测试、`go vet ./...`、前端 typecheck、差异检查通过；尚未提交、推送或重新打包。

## 2026-08-31：固件摘要卡片跟随浅色主题

- 为 `.firmware-summary`、Release 元数据和完整性提示补充浅色主题覆盖，避免切换皮肤后仍保留深色卡片背景和低对比文字。
- 验证：前端 typecheck、生产构建和 `git diff --check` 通过；尚未提交、推送或重新打包。

## 2026-08-31：开源共建入口与审核门禁

- README 重写为下载、BOOT、身份读取、版本选择、写入、恢复和社区固件发布的可操作流程，并加入脱敏的 UI 选择截图；新增贡献指南。
- Issue 改为类型/模块/平台/优先级分流，新增社区固件来源和问题模板；PR 自动按变更路径标注模块，并校验描述含验证说明。
- 新增 CODEOWNERS、CodeQL、维护者 `automerge` 标签触发的原生自动合并和可选 Codex 审核。Codex 仅在仓库配置 `OPENAI_API_KEY` 后提供建议，不能批准或直接合并代码。
- 待远端完成标签创建、分支保护、auto-merge 设置和首次 CI 回读；本机 application 测试仍受 Windows `Access is denied` 执行限制。
- 远端已创建分类标签，启用 squash-only、自动删分支、原生 auto-merge 与 `main` 分支保护；首次 CodeQL 因 Go 不支持 `build-mode: none` 失败，已改为 Go autobuild / JS buildless 并升级 Action v4，待推送回读。

## 已知边界

- `FreeCodeCampXYG/easy-input-maker` 已有 `firmware-v0.2.1` 的受信 manifest Release；新版桌面包尚待实板验证下载、写入和恢复结果。
- 纯 Go 烧录器已完成本地 Windows Wails 编译；六平台 CI、发布包以及真实设备写入/HID 恢复仍待验证。当前无 Python、PyInstaller 或 GPL helper 分发义务。
- Wails 结构审计 13 PASS / 1 WARN；唯一 WARN 是首版不提供 NSIS 安装器，仅提供 Windows 便携 ZIP。
- 前端主题和 Windows 检测改动已通过 npm typecheck/build、Go test/vet 和 `git diff --check`；真实 HID/串口设备回归仍待实板。
- 本轮新增 BOOT 引导提示与“刷新下载端口”动作，前端 typecheck/build、Go test/vet、差异检查通过；尚未推送或发布。
- 提交 `49e7ead` 已推送 main；Go test/vet、前端 typecheck/build 和 `git diff --check` 通过，未创建新 Release tag。
- 提交 `c7a0009` 已推送 main：正常 HID/BLE 只确认设备在线，芯片字段留待下载模式 esptool 验身；BOOT 引导不再显示为红色错误。Go/前端验证通过，未创建新 Release tag。
- 当前电脑没有 Git Bash，Unix 打包脚本只能完成静态检查，须以 GitHub macOS/Linux native runner 作为实际验证证据。
- `v0.1.0` Release 的 Linux 资产构建成功，macOS 两架构在默认 Python 上构建 PyInstaller helper 失败；`v0.1.1` 已修复并验证 macOS/Linux/Windows x64 资产，Windows ARM64 因 cryptography 缺少原生 OpenSSL wheel 失败，Release workflow 现改用 x64 Python/helper，待新 tag 验证。
- `v0.1.3` 六个平台 package jobs 全部成功，publish job 因未 checkout 仓库而无法让 `gh release create --notes-file` 找到 `.git` 失败；新增 `publish-existing-release.yml` 复用该 run 的 artifacts 发布，避免重跑六平台构建。
- `v0.1.3` 发布后复现发现 Windows x64 包实际为 ARM64：Windows 打包脚本硬编码 archive 名称，恢复发布时同名资产覆盖。已修复脚本按架构命名，并让恢复 workflow 分 artifact 下载、仅对旧 ARM64 资产做名称兼容转换；不重编译，待以新 tag/恢复发布验证。
- 一次 Wails 重建曾因先前手动启动的同名可执行文件占用 `build/bin` 失败；精确结束该项目进程后重建成功。构建前必须退出正在运行的应用。

## 下一步

1. 在 Maker `main` 合入固件发布 workflow/manifest 脚本，并创建绑定 main SHA 的 `firmware-v-main-*` Release。
2. 在烧录器新包中验证官方 main 基线显示、下载哈希、受信门禁和实板恢复 HID。
3. 后续再实现精确 GitHub Release URL 导入与逐版本社区审批，不开放分支、裸 bin 或任意 URL。

package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/config"
	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/device"
	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/firmware"
	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/flasher"
)

type FlashRequest struct {
	DeviceID     string `json:"deviceId"`
	FirmwareID   string `json:"firmwareId"`
	Confirmation string `json:"confirmation"`
}

type App struct {
	version string
	ctx     context.Context

	mu             sync.RWMutex
	settings       config.Settings
	devices        map[string]domain.DeviceInfo
	firmware       []domain.FirmwareRelease
	status         domain.FlashStatus
	cancel         context.CancelFunc
	logs           []domain.ActivityLog
	local          map[string]string
	networkOnline  bool
	networkMode    string
	networkAddress string
}

func New(version string) *App {
	return &App{version: version, devices: make(map[string]domain.DeviceInfo), local: make(map[string]string), status: newStatus(domain.FlashStageIdle, "等待检测设备", 0), logs: make([]domain.ActivityLog, 0, 64)}
}

// ImportLocalBundle 校验固定 ZIP 合同后加入本地固件库；失败时保留原文件，不执行任何写入。
func (a *App) ImportLocalBundle(encoded string) (domain.FirmwareRelease, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return domain.FirmwareRelease{}, fmt.Errorf("本地固件包编码无效: %w", err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return domain.FirmwareRelease{}, fmt.Errorf("本地固件包不是有效 ZIP: %w", err)
	}
	root, err := os.MkdirTemp("", "easyinput-flasher-bundle-")
	if err != nil {
		return domain.FirmwareRelease{}, err
	}
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() || filepath.Base(entry.Name) != entry.Name {
			continue
		}
		reader, openErr := entry.Open()
		if openErr != nil {
			return domain.FirmwareRelease{}, openErr
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, 32<<20))
		_ = reader.Close()
		if readErr != nil {
			return domain.FirmwareRelease{}, readErr
		}
		if writeErr := os.WriteFile(filepath.Join(root, entry.Name), content, 0o600); writeErr != nil {
			return domain.FirmwareRelease{}, writeErr
		}
	}
	manifestData, err := os.ReadFile(filepath.Join(root, "firmware-manifest.json"))
	if err != nil {
		return domain.FirmwareRelease{}, fmt.Errorf("本地固件包缺少 firmware-manifest.json")
	}
	manifest, err := firmware.ParseManifest(manifestData)
	if err != nil {
		return domain.FirmwareRelease{}, err
	}
	if err := firmware.VerifyBundle(root, manifest); err != nil {
		return domain.FirmwareRelease{}, err
	}
	id := "local@" + manifest.Tag
	release := domain.FirmwareRelease{ID: id, Repository: "local", Tag: manifest.Tag, Name: "本地固件 " + manifest.Tag, PublishedAt: "本地导入", Manifest: manifest, Trusted: true}
	a.mu.Lock()
	a.local[id] = root
	a.firmware = append(a.firmware, release)
	a.mu.Unlock()
	a.appendLog("success", "固件", "已导入并校验本地固件包："+manifest.Tag)
	return release, nil
}

// ImportFactoryBundle 导入单文件恢复镜像；固定写入 0x0，并显式标记会清除 NVS/蓝牙绑定。
func (a *App) ImportFactoryBundle(encoded string) (domain.FirmwareRelease, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) == 0 || len(data) > 32<<20 {
		return domain.FirmwareRelease{}, fmt.Errorf("Factory 镜像无效或超过大小限制")
	}
	root, err := os.MkdirTemp("", "easyinput-flasher-factory-")
	if err != nil {
		return domain.FirmwareRelease{}, err
	}
	name := "factory.bin"
	if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
		return domain.FirmwareRelease{}, err
	}
	manifest := domain.FirmwareManifest{SchemaVersion: 1, Product: "easyinput-factory", Board: domain.BoardID, Chip: domain.ChipType, Tag: "local-factory", IDFVersion: "未知（Factory 镜像）", Files: []domain.FlashFile{{Name: name, Offset: "0x0", Size: int64(len(data))}}}
	release := domain.FirmwareRelease{ID: "local-factory@" + fmt.Sprint(len(data)), Repository: "local-factory", Tag: "Factory 恢复", Name: "本地 Factory 恢复（会清除配置）", PublishedAt: "本地导入", Manifest: manifest, Trusted: true, IsFactory: true}
	a.mu.Lock()
	a.local[release.ID] = root
	a.firmware = append(a.firmware, release)
	a.mu.Unlock()
	a.appendLog("warning", "固件", "已导入 Factory 恢复镜像；写入将清除 NVS 配置和蓝牙绑定")
	return release, nil
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	settings, err := config.Load()
	if err != nil {
		a.setStatus(domain.FlashStageFailed, err.Error(), 0, false)
		return
	}
	a.mu.Lock()
	a.settings = settings
	a.mu.Unlock()
}

func (a *App) Shutdown(context.Context) { a.CancelFlash() }

func (a *App) GetDashboardSnapshot() domain.DashboardSnapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()
	devices := make([]domain.DeviceInfo, 0, len(a.devices))
	for _, item := range a.devices {
		devices = append(devices, item)
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Mode != devices[j].Mode {
			return devices[i].Mode == "download"
		}
		return devices[i].Port < devices[j].Port
	})
	mode := a.networkMode
	if mode == "" {
		mode = a.settings.ProxyMode
	}
	return domain.DashboardSnapshot{AppVersion: a.version, Status: a.status, Devices: devices, Firmware: a.firmware, ProxyMode: mode, NetworkOnline: a.networkOnline, ProxyAddress: a.networkAddress, Logs: append([]domain.ActivityLog(nil), a.logs...)}
}

// ConfigureNetwork 保存网络策略并立即探测 GitHub；探测失败仍保留用户设置，便于离线机器下次继续调整。
func (a *App) ConfigureNetwork(mode, address string) (domain.DashboardSnapshot, error) {
	mode = strings.TrimSpace(mode)
	if mode != "auto" && mode != "system" && mode != "direct" && mode != "custom" {
		return domain.DashboardSnapshot{}, fmt.Errorf("网络模式无效")
	}
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	settings.ProxyMode = mode
	if mode == "custom" {
		normalized, err := config.NormalizeProxyURL(address)
		if err != nil {
			return domain.DashboardSnapshot{}, err
		}
		settings.ProxyURL = normalized
	}
	if err := config.Save(settings); err != nil {
		return domain.DashboardSnapshot{}, err
	}
	result := firmware.ProbeNetwork(a.ctx, settings)
	a.mu.Lock()
	a.settings = settings
	a.networkOnline = result.Online
	a.networkMode = result.ProxyMode
	a.networkAddress = result.ProxyAddress
	a.mu.Unlock()
	if result.Online {
		a.appendLog("success", "网络", "GitHub 可访问："+result.ProxyAddress)
	} else {
		a.appendLog("warning", "网络", "GitHub 暂不可访问，请检查代理或系统网络")
	}
	return a.GetDashboardSnapshot(), nil
}

// CheckNetwork 重新按当前策略探测，不修改用户配置。
func (a *App) CheckNetwork() domain.DashboardSnapshot {
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	result := firmware.ProbeNetwork(a.ctx, settings)
	a.mu.Lock()
	a.networkOnline = result.Online
	a.networkMode = result.ProxyMode
	a.networkAddress = result.ProxyAddress
	a.mu.Unlock()
	return a.GetDashboardSnapshot()
}

// ExportDiagnostics 导出脱敏阶段日志到用户缓存目录，便于提交 Issue 而不暴露完整设备身份。
func (a *App) ExportDiagnostics() (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	directory := filepath.Join(root, "easyinput-flasher", "exports")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", err
	}
	a.mu.RLock()
	logs := append([]domain.ActivityLog(nil), a.logs...)
	a.mu.RUnlock()
	path := filepath.Join(directory, "diagnostics-"+time.Now().UTC().Format("20060102-150405")+".log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer file.Close()
	for _, item := range logs {
		if _, err := fmt.Fprintf(file, "%s [%s] %s: %s\n", item.Time, item.Level, item.Scope, item.Message); err != nil {
			return "", err
		}
	}
	return path, nil
}

func (a *App) ScanDevices() ([]domain.DeviceInfo, error) {
	devices, err := device.ListPorts()
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.devices = make(map[string]domain.DeviceInfo, len(devices))
	for _, item := range devices {
		a.devices[item.ID] = item
	}
	a.mu.Unlock()
	message := "未发现 EasyInput；请确认数据线、蓝牙连接或设备电源"
	hasDownloadPort := false
	for _, item := range devices {
		if item.Mode == "download" {
			hasDownloadPort = true
			break
		}
	}
	if hasDownloadPort {
		message = "已发现下载端口；现在可读取芯片与 MAC"
	} else if len(devices) > 0 {
		message = "已发现正常模式设备；短按并松开 BOOT 后刷新下载端口"
	}
	// 扫描阶段不属于写入任务，进度保持 0，避免前端把“检测设备”误显示为烧录中。
	a.setStatus(domain.FlashStageIdle, message, 0, false)
	a.appendLog("info", "设备", fmt.Sprintf("发现 %d 个设备候选", len(devices)))
	return devices, nil
}

// RunHardwareDiagnostics 汇总自动验身与需用户观察的板级检查，未知项明确保留为待验证。
func (a *App) RunHardwareDiagnostics(deviceID string) (domain.HardwareDiagnosticSnapshot, error) {
	a.mu.RLock()
	item, found := a.devices[deviceID]
	a.mu.RUnlock()
	if !found {
		return domain.HardwareDiagnosticSnapshot{}, fmt.Errorf("设备已变化，请重新扫描")
	}
	items := []domain.HardwareDiagnosticItem{
		{Key: "chip", Label: "ESP32-S3 芯片", Evidence: "自动检测", Status: "pending", Detail: "进入下载模式后读取"},
		{Key: "flash", Label: "Flash / W25Q128", Evidence: "自动检测", Status: "pending", Detail: "读取 Flash ID；容量需结合芯片响应确认"},
		{Key: "psram", Label: "8 MB PSRAM", Evidence: "固件自检", Status: "unknown", Detail: "烧录器无法从 ROM 单独证明，需运行态固件自检"},
		{Key: "keys", Label: "S1-S8 按键", Evidence: "用户操作", Status: "pending", Detail: "逐个按下并确认界面收到事件"},
		{Key: "encoder", Label: "编码器 A/B 与按压", Evidence: "用户操作", Status: "pending", Detail: "顺时针、逆时针旋转并按压一次"},
		{Key: "led", Label: "WS2812 与状态灯", Evidence: "用户观察", Status: "pending", Detail: "运行态测试灯效，确认 5 颗灯与绿色状态灯"},
		{Key: "audio_in", Label: "I2S 麦克风", Evidence: "用户观察", Status: "pending", Detail: "运行态显示输入电平或录音回波"},
		{Key: "audio_out", Label: "I2S 功放/扬声器", Evidence: "用户观察", Status: "pending", Detail: "播放测试音并确认无明显杂音"},
		{Key: "power", Label: "VIN / CHRG / VBAT", Evidence: "固件读数", Status: "unknown", Detail: "需运行态固件读取；VBAT 只能估算电压，不能等同剩余电量"},
		{Key: "battery_capacity", Label: "电池剩余容量", Evidence: "固件估算", Status: "unknown", Detail: "板上没有独立电量计；只能由固件结合电压、负载和校准估算，不能从烧录器直接证明百分比"},
		{Key: "bluetooth", Label: "蓝牙 BLE", Evidence: "运行态观察", Status: "pending", Detail: "正常开机后确认 EasyInput BLE 广播、配对和输入；具体设备名与协议由固件决定"},
		{Key: "wifi", Label: "Wi-Fi", Evidence: "运行态观察", Status: "pending", Detail: "运行支持 Wi-Fi 的固件后确认扫描或联网结果；板载无线存在不等于当前固件已启用"},
		{Key: "usb_hid", Label: "USB HID", Evidence: "系统枚举", Status: "pending", Detail: "恢复正常模式后确认 USB HID 枚举，并在文本框中验证真实按键输入"},
		{Key: "usb_ble", Label: "USB / BLE HID", Evidence: "系统枚举", Status: "pending", Detail: "恢复正常模式后确认 HID 枚举与输入；Flasher 不读取 HID 私有按键事件"},
	}
	if item.Mode != "download" {
		for i := range items {
			if items[i].Evidence == "自动检测" {
				items[i].Status = "blocked"
				items[i].Detail = "请先开机短按并松开 BOOT，再刷新下载端口"
			}
		}
	} else {
		runner, err := flasher.NewRunner()
		if err != nil {
			return domain.HardwareDiagnosticSnapshot{}, err
		}
		ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
		defer cancel()
		info, err := runner.InspectHardware(ctx, item.Port)
		if err != nil {
			return domain.HardwareDiagnosticSnapshot{}, err
		}
		items[0].Status, items[0].Detail = "passed", info.Chip
		items[1].Status, items[1].Detail = "passed", fmt.Sprintf("厂商 0x%02X · 设备 0x%04X", info.FlashManufacturer, info.FlashDevice)
	}
	if item.Mode == "normal" {
		ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
		telemetry, err := device.ReadHIDDiagnostics(ctx)
		cancel()
		if err == nil && telemetry.Supported {
			for index := range items {
				switch items[index].Key {
				case "usb_hid":
					items[index].Status = "passed"
					items[index].Detail = "已读取正常模式 HID 状态；后续按键事件会更新输入计数"
				case "usb_ble":
					items[index].Status = "passed"
					items[index].Detail = "已读取正常模式 BLE/HID 状态；具体功能仍以固件声明为准"
				}
			}
			return domain.HardwareDiagnosticSnapshot{DeviceID: deviceID, Items: items, Telemetry: &domain.HardwareDiagnosticTelemetry{Supported: telemetry.Supported, Firmware: telemetry.Firmware, LastInput: telemetry.LastInput, InputEvents: telemetry.InputEvents, EncoderSteps: telemetry.EncoderSteps, BatteryMV: telemetry.BatteryMV, BatteryPercent: telemetry.BatteryPct, BatteryState: telemetry.BatteryState, VIN: telemetry.VIN, Charge: telemetry.Charge, LEDGPIO: telemetry.LEDGPIO}}, nil
		}
	}
	a.appendLog("info", "硬件诊断", fmt.Sprintf("完成 %s 的自动验身与板级检查清单", item.Port))
	return domain.HardwareDiagnosticSnapshot{DeviceID: deviceID, Items: items}, nil
}

// ReadHardwareDiagnostics 读取正常模式 HID 的最新遥测；只在用户进入诊断页轮询，避免常驻监听键盘输入。
func (a *App) ReadHardwareDiagnostics(deviceID string) (domain.HardwareDiagnosticTelemetry, error) {
	a.mu.RLock()
	item, found := a.devices[deviceID]
	a.mu.RUnlock()
	if !found || item.Mode != "normal" {
		return domain.HardwareDiagnosticTelemetry{}, fmt.Errorf("请先让设备正常开机并重新检测")
	}
	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
	telemetry, err := device.ReadHIDDiagnostics(ctx)
	cancel()
	if err != nil {
		return domain.HardwareDiagnosticTelemetry{}, err
	}
	return domain.HardwareDiagnosticTelemetry{Supported: telemetry.Supported, Firmware: telemetry.Firmware, LastInput: telemetry.LastInput, InputEvents: telemetry.InputEvents, EncoderSteps: telemetry.EncoderSteps, BatteryMV: telemetry.BatteryMV, BatteryPercent: telemetry.BatteryPct, BatteryState: telemetry.BatteryState, VIN: telemetry.VIN, Charge: telemetry.Charge, LEDGPIO: telemetry.LEDGPIO}, nil
}

func (a *App) InspectDevice(deviceID string) (domain.DeviceInfo, error) {
	a.mu.RLock()
	item, found := a.devices[deviceID]
	a.mu.RUnlock()
	if !found {
		return domain.DeviceInfo{}, fmt.Errorf("端口已变化，请重新扫描")
	}
	if item.Mode != "download" {
		return item, fmt.Errorf("当前是正常 HID 模式；请开机短按并松开 BOOT 后重新扫描")
	}
	runner, err := flasher.NewRunner()
	if err != nil {
		return item, err
	}
	inspectCtx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()
	output, err := runner.Inspect(inspectCtx, item.Port)
	if err != nil {
		a.appendLog("error", "验身", err.Error())
		return item, err
	}
	if !strings.Contains(strings.ToLower(output), "esp32-s3") {
		return item, fmt.Errorf("目标端口不是 ESP32-S3，已停止")
	}
	item.Mode = "download"
	item.Chip = domain.ChipType
	item.MACSuffix = macSuffix(output)
	item.Verified = item.MACSuffix != ""
	item.ObservedAt = time.Now().UTC().Format(time.RFC3339)
	a.mu.Lock()
	a.devices[item.ID] = item
	a.mu.Unlock()
	// 仅在本轮下载模式验身成功后开放前端确认；StartFlash 仍会再次校验设备和确认文本。
	a.setStatus(domain.FlashStageConfirm, "设备已验证，请核对 MAC 尾号后确认烧录", 20, true)
	a.appendLog("success", "验身", "ESP32-S3 下载模式身份读取成功，等待人工确认")
	return item, nil
}

func (a *App) ListFirmware() ([]domain.FirmwareRelease, error) {
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	result := firmware.ProbeNetwork(a.ctx, settings)
	a.mu.Lock()
	a.networkOnline = result.Online
	a.networkMode = result.ProxyMode
	a.networkAddress = result.ProxyAddress
	a.mu.Unlock()
	if settings.ProxyMode == "auto" && result.Online {
		settings.ProxyMode = result.ProxyMode
		if result.ProxyMode == "custom" {
			settings.ProxyURL = result.ProxyAddress
		}
	}
	client, err := firmware.NewGitHubClient(settings)
	if err != nil {
		return nil, err
	}
	var releases []domain.FirmwareRelease
	var failures []string
	for _, source := range settings.Sources {
		items, listErr := client.ListReleases(a.ctx, source)
		if listErr != nil {
			failure := source.Repository + ": " + listErr.Error()
			failures = append(failures, failure)
			a.appendLog("warning", "固件", failure)
			continue
		}
		releases = append(releases, items...)
	}
	if len(releases) == 0 && len(failures) > 0 {
		return nil, fmt.Errorf("未能读取 GitHub Release，请检查网络模式或代理端口: %s", strings.Join(failures, "；"))
	}
	a.mu.Lock()
	a.firmware = releases
	a.mu.Unlock()
	a.appendLog("info", "固件", fmt.Sprintf("读取到 %d 个带清单的公开 Release", len(releases)))
	return releases, nil
}

// AuditFirmwareSource 在添加社区来源前检查 Release 与自动发布契约；审计过程不会写入设备或设置文件。
func (a *App) AuditFirmwareSource(value string) (domain.FirmwareSourceAudit, error) {
	repository, err := config.NormalizeRepository(value)
	if err != nil {
		return domain.FirmwareSourceAudit{}, err
	}
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	client, err := firmware.NewGitHubClient(settings)
	if err != nil {
		return domain.FirmwareSourceAudit{}, err
	}
	audit, err := client.AuditSource(a.ctx, repository)
	if err != nil {
		a.appendLog("warning", "固件来源", repository+": "+err.Error())
		return audit, err
	}
	a.appendLog("info", "固件来源", fmt.Sprintf("%s 审计完成：%d 个可烧录 Release", repository, audit.ValidReleases))
	return audit, nil
}

// TrustFirmwareSource 仅在用户确认且来源已有完整 Release 后持久化信任，避免把搜索结果直接变成写入权限。
func (a *App) TrustFirmwareSource(value, confirmation string) error {
	repository, err := config.NormalizeRepository(value)
	if err != nil {
		return err
	}
	if confirmation != "信任来源 "+repository {
		return fmt.Errorf("确认文本不匹配")
	}
	audit, err := a.AuditFirmwareSource(repository)
	if err != nil {
		return err
	}
	if !audit.Ready {
		return fmt.Errorf("该仓库没有完整的可烧录 Release，不能加入来源")
	}
	a.mu.Lock()
	settings := a.settings
	found := false
	for index := range settings.Sources {
		if settings.Sources[index].Repository == repository {
			settings.Sources[index].Enabled = true
			settings.Sources[index].Trusted = true
			found = true
			break
		}
	}
	if !found {
		settings.Sources = append(settings.Sources, config.Source{Repository: repository, Trusted: true, Enabled: true})
	}
	if err := config.Save(settings); err != nil {
		a.mu.Unlock()
		return err
	}
	a.settings = settings
	a.mu.Unlock()
	a.appendLog("success", "固件来源", repository+" 已由用户确认并加入受信来源")
	return nil
}

func (a *App) StartFlash(request FlashRequest) error {
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return fmt.Errorf("已有烧录任务正在执行")
	}
	item, found := a.devices[request.DeviceID]
	if !found || !item.Verified || item.Chip != domain.ChipType {
		a.mu.Unlock()
		return fmt.Errorf("设备尚未完成 ESP32-S3 验身")
	}
	selected, found := findFirmware(a.firmware, request.FirmwareID)
	if !found {
		a.mu.Unlock()
		return fmt.Errorf("固件选择已过期，请刷新固件列表后重试")
	}
	repository, tag, ok := strings.Cut(request.FirmwareID, "@")
	if !ok {
		a.mu.Unlock()
		return fmt.Errorf("固件选择无效")
	}
	expected := "确认烧录 " + item.MACSuffix
	if selected.IsFactory && (repository == "local-factory" || repository == "local") {
		expected = "确认恢复出厂 " + item.MACSuffix
	}
	if request.Confirmation != expected {
		a.mu.Unlock()
		return fmt.Errorf("确认文本不匹配")
	}
	// 写入权限由后端按本轮受信列表签发，避免 Wails 调用绕过前端的社区来源提示。
	if !selected.Trusted {
		a.mu.Unlock()
		return fmt.Errorf("该固件来源尚未审核，不能写入设备")
	}
	jobCtx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mu.Unlock()
	defer a.finishJob()

	// 状态文案使用清单绑定的 tag，避免 Release 自定义标题重复或误导版本判断。
	firmwareLabel := selected.Tag
	a.setFlashStatus(domain.FlashStageDownload, fmt.Sprintf("正在下载并校验固件：%s，请保持网络连接", firmwareLabel), 30, false, item.ID, request.FirmwareID)
	a.mu.RLock()
	settings := a.settings
	a.mu.RUnlock()
	client, err := firmware.NewGitHubClient(settings)
	if err != nil {
		return err
	}
	var cacheRoot string
	var manifest domain.FirmwareManifest
	if selected.IsFactory && (repository == "local-factory" || repository == "local") {
		a.mu.RLock()
		cacheRoot = a.local[request.FirmwareID]
		a.mu.RUnlock()
		if cacheRoot == "" {
			return fmt.Errorf("Factory 镜像已失效，请重新导入")
		}
		manifest = selected.Manifest
	} else if repository == "local" {
		a.mu.RLock()
		cacheRoot = a.local[request.FirmwareID]
		a.mu.RUnlock()
		if cacheRoot == "" {
			return fmt.Errorf("本地固件包已失效，请重新导入")
		}
		manifest = selected.Manifest
	} else {
		cacheRoot, err = bundlePath(repository, tag)
		if err != nil {
			return err
		}
		manifest, err = client.DownloadBundle(jobCtx, repository, tag, cacheRoot)
		if err != nil {
			a.setFlashStatus(domain.FlashStageFailed, fmt.Sprintf("固件 %s 下载失败：%s", firmwareLabel, err), 0, false, item.ID, request.FirmwareID)
			return err
		}
	}
	a.setFlashStatus(domain.FlashStageWrite, fmt.Sprintf("正在烧录固件：%s；请勿拔出 USB 数据线、关闭电源或再次按 BOOT", firmwareLabel), 55, false, item.ID, request.FirmwareID)
	runner, err := flasher.NewRunner()
	if err != nil {
		return err
	}
	if _, err := runner.FlashWithDetails(jobCtx, item.Port, cacheRoot, manifest, func(index, current, total int) {
		if total <= 0 || index >= len(manifest.Files) {
			return
		}
		percent := 55 + (current+manifestPrefixSize(manifest, index))*33/manifestTotalSize(manifest)
		file := manifest.Files[index]
		a.setFlashStatus(domain.FlashStageWrite, fmt.Sprintf("正在写入 %s：%d%%（%d / %d 字节）", file.Name, percent, current, total), percent, false, item.ID, request.FirmwareID)
		a.updateFlashDetails(file.Name, file.Offset, current, total)
	}); err != nil {
		a.setFlashStatus(domain.FlashStageFailed, fmt.Sprintf("固件 %s 写入失败：%s", firmwareLabel, err), 0, false, item.ID, request.FirmwareID)
		return err
	}
	a.setFlashStatus(domain.FlashStageRecovery, fmt.Sprintf("固件 %s 写入和工具校验完成。若手动进入下载模式，请关机后重新开机", firmwareLabel), 90, false, item.ID, request.FirmwareID)
	return nil
}

func findFirmware(items []domain.FirmwareRelease, id string) (domain.FirmwareRelease, bool) {
	for _, item := range items {
		if item.Repository+"@"+item.Tag == id {
			return item, true
		}
	}
	return domain.FirmwareRelease{}, false
}

func (a *App) CheckRecovery() (bool, error) {
	result, err := device.HasExpectedHID(a.ctx)
	if err != nil {
		return false, err
	}
	if result {
		a.setStatus(domain.FlashStageCompleted, "烧录完成，已检测到正常 HID；具体功能仍需按测试范围验证", 100, false)
	}
	return result, nil
}

func (a *App) CancelFlash() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
		a.setStatus(domain.FlashStageCancelled, "烧录任务已请求取消；请勿拔线，等待当前工具退出", 0, false)
	}
}

func (a *App) finishJob() {
	a.mu.Lock()
	a.cancel = nil
	a.mu.Unlock()
}

func (a *App) setStatus(stage domain.FlashStage, message string, progress int, canFlash bool) {
	a.mu.Lock()
	status := domain.FlashStatus{Stage: stage, Message: message, Progress: progress, CanFlash: canFlash, UpdatedAt: time.Now().UTC()}
	if stage == domain.FlashStageRecovery || stage == domain.FlashStageCompleted || stage == domain.FlashStageCancelled || stage == domain.FlashStageFailed {
		// 收尾状态继续绑定本次目标，避免前端因 firmwareId 为空回退到列表第一项。
		status.DeviceID = a.status.DeviceID
		status.FirmwareID = a.status.FirmwareID
	}
	a.status = status
	a.mu.Unlock()
}

func (a *App) setFlashStatus(stage domain.FlashStage, message string, progress int, canFlash bool, deviceID, firmwareID string) {
	a.mu.Lock()
	// 烧录状态保留设备和版本指针，避免异步轮询或失败日志无法确认实际写入目标。
	a.status = domain.FlashStatus{Stage: stage, Message: message, Progress: progress, CanFlash: canFlash, DeviceID: deviceID, FirmwareID: firmwareID, UpdatedAt: time.Now().UTC()}
	a.mu.Unlock()
}

func (a *App) updateFlashDetails(image, address string, current, total int) {
	a.mu.Lock()
	a.status.CurrentImage, a.status.CurrentAddress = image, address
	a.status.CurrentBytes, a.status.TotalBytes = current, total
	a.mu.Unlock()
	if current == 0 || current >= total {
		if current >= total {
			a.appendLog("success", "烧录", fmt.Sprintf("写入完成 %s at %s（100%%）", image, address))
		} else {
			a.appendLog("info", "烧录", fmt.Sprintf("开始写入 %s at %s", image, address))
		}
	}
}

func manifestTotalSize(manifest domain.FirmwareManifest) int {
	total := 0
	for _, file := range manifest.Files {
		total += int(file.Size)
	}
	if total == 0 {
		for _, file := range manifest.Files {
			if file.Size > 0 {
				total += int(file.Size)
			}
		}
	}
	return maxInt(total, 1)
}
func manifestPrefixSize(manifest domain.FirmwareManifest, index int) int {
	total := 0
	for i := 0; i < index && i < len(manifest.Files); i++ {
		total += int(manifest.Files[i].Size)
	}
	return total
}
func percentOf(current, total int) int {
	if total <= 0 {
		return 0
	}
	return current * 100 / total
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) appendLog(level, scope, message string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logs = append(a.logs, domain.ActivityLog{Time: time.Now().UTC().Format(time.RFC3339), Level: level, Scope: scope, Message: message})
	if len(a.logs) > 100 {
		a.logs = a.logs[len(a.logs)-100:]
	}
}

func newStatus(stage domain.FlashStage, message string, progress int) domain.FlashStatus {
	return domain.FlashStatus{Stage: stage, Message: message, Progress: progress, UpdatedAt: time.Now().UTC()}
}

func macSuffix(output string) string {
	for _, part := range strings.Fields(output) {
		clean := strings.Trim(part, " ,.;")
		if strings.Count(clean, ":") == 5 && len(clean) >= 5 {
			return strings.ToUpper(clean[len(clean)-5:])
		}
	}
	return ""
}

func bundlePath(repository, tag string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	safe := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(repository + "-" + tag)
	return filepath.Join(cache, "easyinput-flasher", "firmware", safe), nil
}

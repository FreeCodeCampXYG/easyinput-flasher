package application

import (
	"context"
	"fmt"
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

	mu       sync.RWMutex
	settings config.Settings
	devices  map[string]domain.DeviceInfo
	firmware []domain.FirmwareRelease
	status   domain.FlashStatus
	cancel   context.CancelFunc
	logs     []domain.ActivityLog
}

func New(version string) *App {
	return &App{version: version, devices: make(map[string]domain.DeviceInfo), status: newStatus(domain.FlashStageIdle, "等待检测设备", 0), logs: make([]domain.ActivityLog, 0, 64)}
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
	return domain.DashboardSnapshot{AppVersion: a.version, Status: a.status, Devices: devices, Firmware: a.firmware, ProxyMode: a.settings.ProxyMode, Logs: append([]domain.ActivityLog(nil), a.logs...)}
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
		return nil, fmt.Errorf("未能读取 GitHub Release，请检查 127.0.0.1:1080 代理或网络连接: %s", strings.Join(failures, "；"))
	}
	a.mu.Lock()
	a.firmware = releases
	a.mu.Unlock()
	a.appendLog("info", "固件", fmt.Sprintf("读取到 %d 个带清单的公开 Release", len(releases)))
	return releases, nil
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
	expected := "确认烧录 " + item.MACSuffix
	if request.Confirmation != expected {
		a.mu.Unlock()
		return fmt.Errorf("确认文本不匹配")
	}
	selected, found := findFirmware(a.firmware, request.FirmwareID)
	if !found {
		a.mu.Unlock()
		return fmt.Errorf("固件选择已过期，请刷新固件列表后重试")
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

	repository, tag, ok := strings.Cut(request.FirmwareID, "@")
	if !ok {
		return fmt.Errorf("固件选择无效")
	}
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
	cacheRoot, err := bundlePath(repository, tag)
	if err != nil {
		return err
	}
	manifest, err := client.DownloadBundle(jobCtx, repository, tag, cacheRoot)
	if err != nil {
		a.setFlashStatus(domain.FlashStageFailed, fmt.Sprintf("固件 %s 下载失败：%s", firmwareLabel, err), 0, false, item.ID, request.FirmwareID)
		return err
	}
	a.setFlashStatus(domain.FlashStageWrite, fmt.Sprintf("正在烧录固件：%s；请勿拔出 USB 数据线、关闭电源或再次按 BOOT", firmwareLabel), 55, false, item.ID, request.FirmwareID)
	runner, err := flasher.NewRunner()
	if err != nil {
		return err
	}
	if _, err := runner.Flash(jobCtx, item.Port, cacheRoot, manifest); err != nil {
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

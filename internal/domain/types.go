package domain

import "time"

const (
	BoardID  = "easyinput-v2"
	ChipType = "esp32s3"
)

type FlashFile struct {
	Name   string `json:"name"`
	Offset string `json:"offset"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type FirmwareManifest struct {
	SchemaVersion int         `json:"schemaVersion"`
	Product       string      `json:"product"`
	Board         string      `json:"board"`
	Chip          string      `json:"chip"`
	Tag           string      `json:"tag"`
	Commit        string      `json:"commit"`
	IDFVersion    string      `json:"idfVersion"`
	ReleaseNotes  string      `json:"releaseNotes"`
	Files         []FlashFile `json:"files"`
}

type FirmwareRelease struct {
	ID          string           `json:"id"`
	Repository  string           `json:"repository"`
	Tag         string           `json:"tag"`
	Name        string           `json:"name"`
	PublishedAt string           `json:"publishedAt"`
	Manifest    FirmwareManifest `json:"manifest"`
	Trusted     bool             `json:"trusted"`
	IsFactory   bool             `json:"isFactory"`
}

type FirmwareSourceCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type FirmwareSourceAudit struct {
	Repository    string                `json:"repository"`
	ValidReleases int                   `json:"validReleases"`
	Ready         bool                  `json:"ready"`
	Checks        []FirmwareSourceCheck `json:"checks"`
}

type DeviceInfo struct {
	ID         string `json:"id"`
	Port       string `json:"port"`
	Label      string `json:"label"`
	Mode       string `json:"mode"`
	Chip       string `json:"chip"`
	MACSuffix  string `json:"macSuffix"`
	FlashSize  string `json:"flashSize"`
	Verified   bool   `json:"verified"`
	ObservedAt string `json:"observedAt"`
}

type FlashStage string

const (
	FlashStageIdle      FlashStage = "idle"
	FlashStageDownload  FlashStage = "download"
	FlashStageInspect   FlashStage = "inspect"
	FlashStageConfirm   FlashStage = "confirmation"
	FlashStageWrite     FlashStage = "writing"
	FlashStageVerify    FlashStage = "verify"
	FlashStageRecovery  FlashStage = "recovery"
	FlashStageCompleted FlashStage = "completed"
	FlashStageFailed    FlashStage = "failed"
	FlashStageCancelled FlashStage = "cancelled"
)

type FlashStatus struct {
	Stage          FlashStage `json:"stage"`
	Message        string     `json:"message"`
	Progress       int        `json:"progress"`
	CanFlash       bool       `json:"canFlash"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeviceID       string     `json:"deviceId,omitempty"`
	FirmwareID     string     `json:"firmwareId,omitempty"`
	CurrentImage   string     `json:"currentImage,omitempty"`
	CurrentAddress string     `json:"currentAddress,omitempty"`
	CurrentBytes   int        `json:"currentBytes,omitempty"`
	TotalBytes     int        `json:"totalBytes,omitempty"`
}

type DashboardSnapshot struct {
	AppVersion string            `json:"appVersion"`
	Status     FlashStatus       `json:"status"`
	Devices    []DeviceInfo      `json:"devices"`
	Firmware   []FirmwareRelease `json:"firmware"`
	ProxyMode  string            `json:"proxyMode"`
	Logs       []ActivityLog     `json:"logs"`
}

type ActivityLog struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Scope   string `json:"scope"`
	Message string `json:"message"`
}

type HardwareDiagnosticItem struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Evidence string `json:"evidence"`
	Status   string `json:"status"`
	Detail   string `json:"detail"`
}

type HardwareDiagnosticSnapshot struct {
	DeviceID  string                       `json:"deviceId"`
	Items     []HardwareDiagnosticItem     `json:"items"`
	Telemetry *HardwareDiagnosticTelemetry `json:"telemetry,omitempty"`
}

type HardwareDiagnosticTelemetry struct {
	Supported      bool   `json:"supported"`
	Firmware       string `json:"firmware"`
	LastInput      string `json:"lastInput"`
	InputEvents    uint32 `json:"inputEvents"`
	EncoderSteps   uint32 `json:"encoderSteps"`
	BatteryMV      uint16 `json:"batteryMv"`
	BatteryPercent uint8  `json:"batteryPercent"`
	BatteryState   string `json:"batteryState"`
	VIN            int    `json:"vin"`
	Charge         int    `json:"charge"`
	LEDGPIO        int    `json:"ledGpio"`
}

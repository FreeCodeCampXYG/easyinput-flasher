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
	Stage      FlashStage `json:"stage"`
	Message    string     `json:"message"`
	Progress   int        `json:"progress"`
	CanFlash   bool       `json:"canFlash"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeviceID   string     `json:"deviceId,omitempty"`
	FirmwareID string     `json:"firmwareId,omitempty"`
}

type DashboardSnapshot struct {
	AppVersion string            `json:"appVersion"`
	Status     FlashStatus       `json:"status"`
	Devices    []DeviceInfo      `json:"devices"`
	Firmware   []FirmwareRelease `json:"firmware"`
	ProxyMode  string            `json:"proxyMode"`
}

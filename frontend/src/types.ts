export type PageId =
  | "flash"
  | "library"
  | "discover"
  | "devices"
  | "updates"
  | "about";

export type DeviceMode = "normal" | "download" | "unknown";
export type FlashStage =
  | "idle"
  | "detecting"
  | "verified"
  | "downloading"
  | "flashing"
  | "restarting"
  | "complete"
  | "failed";

export interface DeviceInfo {
  id: string;
  port: string;
  product: string;
  chip: string;
  mode: DeviceMode;
  macSuffix: string;
  usbVidPid: string;
  firmwareVersion?: string;
  connected: boolean;
  verified: boolean;
}

export interface FirmwareFeature {
  name: string;
  status: "declared" | "verified" | "unavailable";
}

export interface FirmwareRelease {
  id: string;
  repository: string;
  sourceName: string;
  tag: string;
  commit: string;
  publishedAt: string;
  board: string;
  chip: string;
  idfVersion: string;
  size: string;
  trusted: boolean;
  checksumVerified: boolean;
  channel: "stable" | "preview" | "custom";
  features: FirmwareFeature[];
  changelog: string[];
}

export interface FirmwareSourceAudit {
  repository: string;
  validReleases: number;
  ready: boolean;
  checks: Array<{ name: string; passed: boolean; message: string }>;
}

export interface FlashProgress {
  stage: FlashStage;
  percent: number;
  step: string;
  message: string;
  currentImage?: string;
  canCancel: boolean;
  hidRecovered: boolean;
  functionalVerification: "pending" | "passed" | "failed";
}

export interface ActivityLog {
  id: string;
  time: string;
  level: "info" | "success" | "warning" | "error";
  scope: string;
  message: string;
}

export interface DashboardSnapshot {
  backendReady: boolean;
  demoMode?: boolean;
  appVersion: string;
  canFlash: boolean;
  gateReasons: string[];
  selectedDeviceId?: string;
  selectedFirmwareId?: string;
  devices: DeviceInfo[];
  releases: FirmwareRelease[];
  progress: FlashProgress;
  logs: ActivityLog[];
  network: {
    online: boolean;
    proxyMode: "system" | "custom" | "direct";
    proxyAddress?: string;
  };
  cache: {
    items: number;
    size: string;
  };
  updateAvailable?: string;
}

export interface StartFlashInput {
  deviceId: string;
  firmwareId: string;
  confirmation: string;
}

export interface OperationResult {
  ok: boolean;
  message?: string;
}

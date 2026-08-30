import type {
  ActivityLog,
  DashboardSnapshot,
  DeviceInfo,
  FirmwareRelease,
  FlashStage,
  OperationResult,
  StartFlashInput,
} from "./types";

interface RawFlashStatus {
  stage: string;
  message: string;
  progress: number;
  canFlash: boolean;
  deviceId?: string;
  firmwareId?: string;
}

interface RawDeviceInfo {
  id: string;
  port: string;
  label: string;
  mode: string;
  chip: string;
  macSuffix: string;
  flashSize: string;
  verified: boolean;
  observedAt: string;
}

interface RawFirmwareRelease {
  id: string;
  repository: string;
  tag: string;
  name: string;
  publishedAt: string;
  trusted: boolean;
  manifest?: {
    board?: string;
    chip?: string;
    commit?: string;
    idfVersion?: string;
    releaseNotes?: string;
    files?: Array<{ size: number }>;
  };
}

interface RawDashboardSnapshot {
  appVersion: string;
  status: RawFlashStatus;
  devices: RawDeviceInfo[];
  firmware: RawFirmwareRelease[];
  proxyMode: string;
  logs?: Array<{ time: string; level: string; scope: string; message: string }>;
}

interface WailsAppBridge {
  GetDashboardSnapshot?: () => Promise<RawDashboardSnapshot>;
  ScanDevices?: () => Promise<RawDeviceInfo[]>;
  InspectDevice?: (deviceId: string) => Promise<RawDeviceInfo>;
  ListFirmware?: () => Promise<RawFirmwareRelease[]>;
  StartFlash?: (input: StartFlashInput) => Promise<void>;
  CancelFlash?: () => Promise<void>;
  CheckRecovery?: () => Promise<boolean>;
  ExportDiagnostics?: () => Promise<string>;
}

declare global {
  interface Window {
    go?: {
      main?: { App?: WailsAppBridge };
      application?: { App?: WailsAppBridge };
    };
  }
}

const appBridge = (): WailsAppBridge | undefined => window.go?.application?.App ?? window.go?.main?.App;

export async function getDashboardSnapshot(): Promise<DashboardSnapshot | undefined> {
  const operation = appBridge()?.GetDashboardSnapshot;
  if (!operation) return undefined;
  return adaptSnapshot(await operation());
}

export async function scanDevices(): Promise<DashboardSnapshot | undefined> {
  const bridge = appBridge();
  if (!bridge?.ScanDevices) return undefined;
  await bridge.ScanDevices();
  return refreshFrom(bridge);
}

export async function inspectDevice(deviceId: string): Promise<DashboardSnapshot | undefined> {
  const bridge = appBridge();
  if (!bridge?.InspectDevice) return undefined;
  await bridge.InspectDevice(deviceId);
  return refreshFrom(bridge);
}

export async function listFirmware(): Promise<DashboardSnapshot | undefined> {
  const bridge = appBridge();
  if (!bridge?.ListFirmware) return undefined;
  await bridge.ListFirmware();
  return refreshFrom(bridge);
}

export async function checkForUpdates(): Promise<DashboardSnapshot | undefined> {
  // 首版应用没有独立更新服务；复用公开 Release 刷新，不引入额外 Token 或后台 API。
  return listFirmware();
}

export async function startFlash(input: StartFlashInput): Promise<OperationResult> {
  const operation = appBridge()?.StartFlash;
  // 真实写入绝不提供浏览器回退，后端缺席时保持失败，避免预览数据误触设备。
  if (!operation) return { ok: false, message: "烧录后端尚未连接" };
  try {
    await operation(input);
    return { ok: true };
  } catch (error) {
    return { ok: false, message: errorMessage(error) };
  }
}

export async function cancelFlash(): Promise<OperationResult> {
  const operation = appBridge()?.CancelFlash;
  if (!operation) return { ok: false, message: "当前没有可取消的烧录任务" };
  try {
    await operation();
    return { ok: true };
  } catch (error) {
    return { ok: false, message: errorMessage(error) };
  }
}

export async function checkRecovery(): Promise<DashboardSnapshot | undefined> {
  const bridge = appBridge();
  if (!bridge?.CheckRecovery) return undefined;
  await bridge.CheckRecovery();
  return refreshFrom(bridge);
}

export async function exportDiagnostics(): Promise<OperationResult> {
  const operation = appBridge()?.ExportDiagnostics;
  if (!operation) return { ok: false, message: "诊断导出后端尚未连接" };
  try {
    const path = await operation();
    return { ok: true, message: `诊断日志已导出到 ${path}` };
  } catch (error) {
    return { ok: false, message: errorMessage(error) };
  }
}

async function refreshFrom(bridge: WailsAppBridge): Promise<DashboardSnapshot | undefined> {
  if (!bridge.GetDashboardSnapshot) return undefined;
  return adaptSnapshot(await bridge.GetDashboardSnapshot());
}

function adaptSnapshot(raw: RawDashboardSnapshot): DashboardSnapshot {
  const releases = (raw.firmware ?? []).map(adaptFirmware);
  const devices = (raw.devices ?? []).map(adaptDevice);
  const stage = adaptStage(raw.status?.stage);
  return {
    backendReady: true,
    appVersion: raw.appVersion || "dev",
    canFlash: Boolean(raw.status?.canFlash),
    gateReasons: raw.status?.canFlash || stage === "idle" || stage === "detecting" ? [] : [raw.status?.message || "等待后端签发写入许可"],
    selectedDeviceId: raw.status?.deviceId || devices[0]?.id,
    selectedFirmwareId: raw.status?.firmwareId || releases[0]?.id,
    devices,
    releases,
    progress: {
      stage,
      percent: clampProgress(raw.status?.progress),
      step: stageLabel(stage),
      message: raw.status?.message || "等待操作",
      canCancel: stage === "downloading" || stage === "flashing",
      hidRecovered: stage === "complete",
      functionalVerification: "pending",
    },
    logs: raw.status?.message
      ? [{ id: `status-${raw.status.stage}`, time: "现在", level: stage === "failed" ? "error" : "info", scope: "状态", message: raw.status.message }, ...(raw.logs ?? []).map((item, index) => ({ id: `log-${index}`, time: formatLogTime(item.time), level: adaptLogLevel(item.level), scope: item.scope, message: item.message }))]
      : (raw.logs ?? []).map((item, index) => ({ id: `log-${index}`, time: formatLogTime(item.time), level: adaptLogLevel(item.level), scope: item.scope, message: item.message })),
    network: {
      online: true,
      proxyMode: adaptProxyMode(raw.proxyMode),
      proxyAddress: raw.proxyMode === "custom" ? "已配置" : undefined,
    },
    cache: { items: 0, size: "按需缓存" },
  };
}

function adaptLogLevel(value: string): ActivityLog["level"] {
  if (value === "error" || value === "success" || value === "warning") return value;
  return "info";
}

function formatLogTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: false });
}

function adaptDevice(raw: RawDeviceInfo): DeviceInfo {
  const mode = raw.mode === "download" || raw.mode === "normal" ? raw.mode : "unknown";
  return {
    id: raw.id,
    port: raw.port,
    product: raw.label || "串口设备",
    chip: raw.chip || "未读取",
    mode,
    macSuffix: raw.macSuffix || "",
    usbVidPid: "未读取",
    firmwareVersion: "未读取",
    connected: true,
    verified: Boolean(raw.verified),
  };
}

function adaptFirmware(raw: RawFirmwareRelease): FirmwareRelease {
  const manifest = raw.manifest ?? {};
  const totalSize = manifest.files?.reduce((sum, file) => sum + (file.size || 0), 0) ?? 0;
  return {
    // Go 后端以 repository@tag 定位下载目标；GitHub 数字 ID 仅用于列表展示，不能参与写入。
    id: `${raw.repository}@${raw.tag}`,
    repository: raw.repository,
    sourceName: raw.name || raw.repository,
    tag: raw.tag,
    commit: manifest.commit || "下载后核对",
    publishedAt: raw.publishedAt ? raw.publishedAt.slice(0, 10) : "未知",
    board: manifest.board || "下载后核对",
    chip: manifest.chip || "ESP32-S3",
    idfVersion: manifest.idfVersion || "下载后核对",
    size: totalSize > 0 ? formatBytes(totalSize) : "下载后核对",
    trusted: Boolean(raw.trusted),
    checksumVerified: false,
    channel: raw.tag.includes("beta") || raw.tag.includes("rc") ? "preview" : "stable",
    features: [],
    changelog: manifest.releaseNotes ? [manifest.releaseNotes] : [],
  };
}

function adaptStage(stage?: string): FlashStage {
  switch (stage) {
    case "inspect": return "detecting";
    case "confirmation": return "verified";
    case "download": return "downloading";
    case "writing":
    case "verify": return "flashing";
    case "recovery": return "restarting";
    case "completed": return "complete";
    case "failed":
    case "cancelled": return "failed";
    default: return "idle";
  }
}

function stageLabel(stage: FlashStage): string {
  const labels: Record<FlashStage, string> = {
    idle: "等待开始",
    detecting: "检测设备",
    verified: "设备验身完成",
    downloading: "下载并校验固件",
    flashing: "写入与工具校验",
    restarting: "等待恢复正常启动",
    complete: "烧录完成",
    failed: "操作已停止",
  };
  return labels[stage];
}

function adaptProxyMode(value?: string): DashboardSnapshot["network"]["proxyMode"] {
  if (value === "custom") return "custom";
  if (value === "disabled" || value === "direct") return "direct";
  return "system";
}

function clampProgress(value?: number): number {
  return Math.max(0, Math.min(100, Number.isFinite(value) ? value! : 0));
}

function formatBytes(value: number): string {
  if (value < 1024 * 1024) return `${Math.round(value / 1024)} KB`;
  return `${(value / 1024 / 1024).toFixed(2)} MB`;
}

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return "操作失败，请查看设备日志";
}

import { useCallback, useEffect, useState, type ReactNode } from "react";
import {
  Activity,
  BadgeCheck,
  Bell,
  BookOpen,
  Box,
  Cable,
  Check,
  CheckCircle2,
  ChevronRight,
  CircleHelp,
  CircleOff,
  Clock3,
  CloudDownload,
  Code2,
  Cpu,
  Database,
  Download,
  ExternalLink,
  FileCheck2,
  Github,
  HardDrive,
  Info,
  Library,
  LoaderCircle,
  Menu,
  MonitorCog,
  Network,
  PackageCheck,
  RefreshCw,
  Search,
  Settings2,
  ShieldAlert,
  ShieldCheck,
  Sparkles,
  Star,
  Sun,
  Moon,
  TerminalSquare,
  Usb,
  X,
  XCircle,
  Zap,
} from "lucide-react";
import {
  cancelFlash,
  checkRecovery,
  checkForUpdates,
  auditFirmwareSource,
  getDashboardSnapshot,
  inspectDevice,
  exportDiagnostics,
  importLocalBundle,
  importFactoryBundle,
  runHardwareDiagnostics,
  readHardwareDiagnostics,
  listFirmware,
  scanDevices,
  startFlash,
  trustFirmwareSource,
} from "./bridge";
import { demoSnapshot } from "./demo-data";
import type {
  ActivityLog,
  DashboardSnapshot,
  DeviceInfo,
  FirmwareFeature,
  FirmwareRelease,
  FirmwareSourceAudit,
  FlashStage,
  HardwareDiagnosticSnapshot,
  PageId,
} from "./types";

const NAV_ITEMS: Array<{ id: PageId; label: string; icon: typeof Zap }> = [
  { id: "flash", label: "烧录", icon: Zap },
  { id: "library", label: "固件库", icon: Library },
  { id: "discover", label: "发现", icon: Search },
  { id: "devices", label: "设备与日志", icon: MonitorCog },
  { id: "diagnostics", label: "硬件诊断", icon: Activity },
  { id: "updates", label: "更新与通知", icon: Bell },
  { id: "about", label: "关于与帮助", icon: CircleHelp },
];

const STAGES: Array<{ id: FlashStage; label: string }> = [
  { id: "detecting", label: "检测设备" },
  { id: "verified", label: "核对身份" },
  { id: "downloading", label: "准备固件" },
  { id: "flashing", label: "写入校验" },
  { id: "restarting", label: "恢复启动" },
  { id: "complete", label: "完成" },
];

const STAGE_INDEX: Record<FlashStage, number> = {
  idle: 0,
  detecting: 0,
  verified: 1,
  downloading: 2,
  flashing: 3,
  restarting: 4,
  complete: 5,
  failed: 0,
};

const isBusyStage = (stage: FlashStage) =>
  stage === "downloading" || stage === "flashing" || stage === "restarting";

function App() {
  const [theme, setTheme] = useState<"dark" | "light">(() => window.localStorage.getItem("easyinput-flasher-theme") === "light" ? "light" : "dark");
  const [page, setPage] = useState<PageId>("flash");
  const [snapshot, setSnapshot] = useState<DashboardSnapshot>(demoSnapshot);
  const [selectedDeviceId, setSelectedDeviceId] = useState(demoSnapshot.selectedDeviceId ?? "");
  const [selectedFirmwareId, setSelectedFirmwareId] = useState(demoSnapshot.selectedFirmwareId ?? "");
  const [confirmation, setConfirmation] = useState("");
  const [pendingAction, setPendingAction] = useState<string>();
  const [notice, setNotice] = useState<string>();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const applySnapshot = useCallback((next?: DashboardSnapshot) => {
    if (!next) return;
    setSnapshot(next);
    setSelectedDeviceId((current) =>
      next.selectedDeviceId ?? (next.devices.some((item) => item.id === current) ? current : next.devices[0]?.id ?? ""),
    );
    setSelectedFirmwareId((current) =>
      next.selectedFirmwareId ?? (next.releases.some((item) => item.id === current) ? current : next.releases[0]?.id ?? ""),
    );
  }, []);

  useEffect(() => {
    let disposed = false;
    const loadInitialData = async () => {
      try {
        applySnapshot(await getDashboardSnapshot());
        if (!disposed) applySnapshot(await listFirmware());
      } catch (error) {
        if (!disposed) setNotice(error instanceof Error ? error.message : "固件列表读取失败，请检查网络与代理设置");
      }
    };
    void loadInitialData();
    return () => { disposed = true; };
  }, [applySnapshot]);

  useEffect(() => {
    if (!isBusyStage(snapshot.progress.stage) || !snapshot.backendReady) return undefined;
    const timer = window.setInterval(() => {
      getDashboardSnapshot().then(applySnapshot).catch(() => undefined);
    }, 900);
    return () => window.clearInterval(timer);
  }, [applySnapshot, snapshot.backendReady, snapshot.progress.stage]);

  const device = snapshot.devices.find((item) => item.id === selectedDeviceId);
  const firmware = snapshot.releases.find((item) => item.id === selectedFirmwareId);
  const expectedConfirmation = device?.macSuffix ? `${firmware?.isFactory ? "确认恢复出厂" : "确认烧录"} ${device.macSuffix}` : "";
  const confirmationMatches = expectedConfirmation.length > 0 && confirmation.trim() === expectedConfirmation;
  const canStart = Boolean(
    snapshot.backendReady &&
      snapshot.canFlash &&
      device?.verified &&
      firmware?.trusted &&
      confirmationMatches &&
      !isBusyStage(snapshot.progress.stage),
  );

  const runSnapshotAction = async (name: string, action: () => Promise<DashboardSnapshot | undefined>) => {
    setPendingAction(name);
    setNotice(undefined);
    try {
      applySnapshot(await action());
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "操作失败，请查看设备日志");
    } finally {
      setPendingAction(undefined);
    }
  };

  const handleStartFlash = async () => {
    if (!canStart || !device || !firmware) return;
    setPendingAction("flash");
    setNotice(undefined);
    // 后端执行是长任务，先进入受控忙碌态以锁定设备和来源；最终状态始终以后端轮询结果覆盖。
    setSnapshot((current) => ({
      ...current,
      progress: {
        ...current.progress,
        stage: "downloading",
        percent: Math.max(30, current.progress.percent),
        step: "下载并校验固件",
        message: "正在准备固件；写入开始后请勿断开设备。",
        canCancel: true,
      },
    }));
    const result = await startFlash({
      deviceId: device.id,
      firmwareId: firmware.id,
      confirmation: confirmation.trim(),
    });
    if (!result.ok) setNotice(result.message ?? "后端拒绝了本次写入");
    applySnapshot(await getDashboardSnapshot());
    setPendingAction(undefined);
  };

  const handleInspect = () => {
    if (!selectedDeviceId) return;
    return runSnapshotAction("inspect", () => inspectDevice(selectedDeviceId));
  };

  const handleRecovery = () => runSnapshotAction("recovery", checkRecovery);

  const handleCancel = async () => {
    setPendingAction("cancel");
    const result = await cancelFlash();
    if (!result.ok) setNotice(result.message ?? "当前阶段无法取消");
    applySnapshot(await getDashboardSnapshot());
    setPendingAction(undefined);
  };

  const renderPage = () => {
    switch (page) {
      case "library":
        return (
          <LibraryPage
            releases={snapshot.releases}
            selectedId={selectedFirmwareId}
            onSelect={(id) => {
              setSelectedFirmwareId(id);
              setPage("flash");
            }}
            onRefresh={() => runSnapshotAction("firmware", listFirmware)}
            refreshing={pendingAction === "firmware"}
            onImport={() => {
              const input = document.createElement("input"); input.type = "file"; input.accept = ".zip";
              input.onchange = () => { const file = input.files?.[0]; if (file) void runSnapshotAction("import", () => importLocalBundle(file)); };
              input.click();
            }}
            onImportFactory={() => {
              const input = document.createElement("input"); input.type = "file"; input.accept = ".bin";
              input.onchange = () => { const file = input.files?.[0]; if (file) void runSnapshotAction("import-factory", () => importFactoryBundle(file)); };
              input.click();
            }}
          />
        );
      case "discover":
        return <DiscoverPage onFirmwareChanged={applySnapshot} />;
      case "devices":
        return <DevicesPage devices={snapshot.devices} logs={snapshot.logs} onExport={exportDiagnostics} />;
      case "diagnostics":
        return <DiagnosticsPage devices={snapshot.devices} releases={snapshot.releases} selectedFirmwareId={selectedFirmwareId} onSelectFirmware={setSelectedFirmwareId} onRun={runHardwareDiagnostics} onRead={readHardwareDiagnostics} />;
      case "updates":
        return (
          <UpdatesPage
            currentVersion={snapshot.appVersion}
            availableVersion={snapshot.updateAvailable}
            onCheck={() => runSnapshotAction("updates", checkForUpdates)}
            checking={pendingAction === "updates"}
          />
        );
      case "about":
        return <AboutPage version={snapshot.appVersion} />;
      default:
        return (
          <FlashPage
            snapshot={snapshot}
            device={device}
            firmware={firmware}
            selectedDeviceId={selectedDeviceId}
            selectedFirmwareId={selectedFirmwareId}
            confirmation={confirmation}
            expectedConfirmation={expectedConfirmation}
            canStart={canStart}
            pendingAction={pendingAction}
            notice={notice}
            onSelectDevice={(id) => {
              setSelectedDeviceId(id);
              setConfirmation("");
            }}
            onSelectFirmware={setSelectedFirmwareId}
            onConfirmation={setConfirmation}
            onScan={() => runSnapshotAction("scan", scanDevices)}
            onInspect={handleInspect}
            onStart={handleStartFlash}
            onCancel={handleCancel}
            onRecovery={handleRecovery}
            onViewLogs={() => setPage("devices")}
          />
        );
    }
  };

  const toggleTheme = () => {
    setTheme((current) => {
      const next = current === "dark" ? "light" : "dark";
      window.localStorage.setItem("easyinput-flasher-theme", next);
      return next;
    });
  };

  return (
    <div className={`app-shell theme-${theme}`}>
      <Sidebar
        page={page}
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onNavigate={(next) => {
          setPage(next);
          setSidebarOpen(false);
        }}
        theme={theme}
        onToggleTheme={toggleTheme}
      />
      <main className="main-workspace">
        <button className="mobile-menu" type="button" onClick={() => setSidebarOpen(true)} aria-label="打开导航">
          <Menu size={19} />
        </button>
        {renderPage()}
      </main>
      <StatusBar snapshot={snapshot} device={device} />
    </div>
  );
}

interface SidebarProps {
  page: PageId;
  open: boolean;
  onClose: () => void;
  onNavigate: (page: PageId) => void;
  theme: "dark" | "light";
  onToggleTheme: () => void;
}

function Sidebar({ page, open, onClose, onNavigate, theme, onToggleTheme }: SidebarProps) {
  return (
    <>
      {open && <button className="sidebar-backdrop" type="button" onClick={onClose} aria-label="关闭导航" />}
      <aside className={`sidebar ${open ? "is-open" : ""}`}>
        <div className="brand-lockup">
          <img src="/app-icon.png" alt="" className="brand-icon" />
          <div>
            <strong>EasyInput</strong>
            <span>Flasher</span>
          </div>
        </div>
        <nav className="primary-nav" aria-label="主导航">
          <span className="nav-caption">工作区</span>
          {NAV_ITEMS.slice(0, 4).map((item) => (
            <NavButton key={item.id} item={item} active={page === item.id} onClick={() => onNavigate(item.id)} />
          ))}
          <span className="nav-caption nav-caption-spaced">应用</span>
          {NAV_ITEMS.slice(4).map((item) => (
            <NavButton key={item.id} item={item} active={page === item.id} onClick={() => onNavigate(item.id)} />
          ))}
        </nav>
        <button className="theme-toggle" type="button" onClick={onToggleTheme}>
          {theme === "dark" ? <Sun size={16} /> : <Moon size={16} />}
          <span>{theme === "dark" ? "切换浅色主题" : "切换深色主题"}</span>
        </button>
        <div className="sidebar-source">
          <Github size={16} />
          <div>
            <span>当前主源</span>
            <strong>FreeCodeCampXYG</strong>
          </div>
          <ChevronRight size={16} />
        </div>
      </aside>
    </>
  );
}

function NavButton({
  item,
  active,
  onClick,
}: {
  item: (typeof NAV_ITEMS)[number];
  active: boolean;
  onClick: () => void;
}) {
  const Icon = item.icon;
  return (
    <button className={`nav-button ${active ? "active" : ""}`} type="button" onClick={onClick}>
      <Icon size={18} />
      <span>{item.label}</span>
    </button>
  );
}

interface FlashPageProps {
  snapshot: DashboardSnapshot;
  device?: DeviceInfo;
  firmware?: FirmwareRelease;
  selectedDeviceId: string;
  selectedFirmwareId: string;
  confirmation: string;
  expectedConfirmation: string;
  canStart: boolean;
  pendingAction?: string;
  notice?: string;
  onSelectDevice: (id: string) => void;
  onSelectFirmware: (id: string) => void;
  onConfirmation: (value: string) => void;
  onScan: () => void;
  onInspect: () => void;
  onStart: () => void;
  onCancel: () => void;
  onRecovery: () => void;
  onViewLogs: () => void;
}

function FlashPage(props: FlashPageProps) {
  const {
    snapshot,
    device,
    firmware,
    selectedDeviceId,
    selectedFirmwareId,
    confirmation,
    expectedConfirmation,
    canStart,
    pendingAction,
    notice,
    onSelectDevice,
    onSelectFirmware,
    onConfirmation,
    onScan,
    onInspect,
    onStart,
    onCancel,
    onRecovery,
    onViewLogs,
  } = props;
  const busy = isBusyStage(snapshot.progress.stage);
  const complete = snapshot.progress.stage === "complete";

  return (
    <div className="page page-flash">
      <PageHeader
        eyebrow="固件交付工作台"
        title="EasyInput 固件烧录与硬件诊断工具"
        description="从受信 Release 或本地固件包完成设备验身、风险确认、地址级写入和恢复检查。"
        actions={
          <button className="button secondary" type="button" onClick={onScan} disabled={pendingAction === "scan" || busy}>
            <RefreshCw size={16} className={pendingAction === "scan" ? "spin" : ""} />
            重新检测
          </button>
        }
      />

      {snapshot.demoMode && (
        <div className="preview-strip">
          <Info size={16} />
          当前展示界面预览数据；Wails 后端接入后才允许执行真实设备写入。
        </div>
      )}

      {device?.mode === "normal" && !device.verified && (
        <div className="boot-instruction" role="status">
          <div className="boot-instruction-icon"><Usb size={20} /></div>
          <div>
            <strong>已检测到正常 HID 设备</strong>
            <span>请保持开发板开机，短按并松开一次 BOOT。等待端口重新出现后，点击“刷新下载端口”继续验身。</span>
          </div>
          <button className="button primary" type="button" onClick={onScan} disabled={pendingAction === "scan" || busy}>
            <RefreshCw size={15} className={pendingAction === "scan" ? "spin" : ""} />
            刷新下载端口
          </button>
        </div>
      )}

      <FlashStepper stage={snapshot.progress.stage} />

      {busy && (
        <div className="connection-warning" role="alert">
          <ShieldAlert size={22} />
          <div>
            <strong>正在烧录，请保持设备连接</strong>
            <span>请勿拔出 USB 数据线、关闭电源、按 BOOT 或退出应用。写入完成后会提示恢复方式。</span>
          </div>
          <span className="progress-number">{snapshot.progress.percent}%</span>
        </div>
      )}

      <section className="workspace-grid">
        <div className="workspace-column">
          <SectionHeading icon={Usb} title="目标设备" meta={device?.verified ? "身份已核对" : "等待验身"} />
          <div className="selector-row">
            <label className="field grow">
              <span>设备端口</span>
              <select value={selectedDeviceId} onChange={(event) => onSelectDevice(event.target.value)} disabled={busy}>
                {snapshot.devices.length === 0 && <option value="">未发现设备</option>}
                {snapshot.devices.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.port} · {item.product}
                  </option>
                ))}
              </select>
            </label>
            <button
              className="button secondary inspect-button"
              type="button"
              onClick={onInspect}
              disabled={!device || busy || pendingAction === "inspect"}
            >
              {pendingAction === "inspect" ? <LoaderCircle size={15} className="spin" /> : <Cpu size={15} />}
              {device?.mode === "normal" ? "短按 BOOT 后刷新" : "读取信息"}
            </button>
            <StatusPill tone={device?.verified ? "success" : "neutral"}>
              {device?.verified ? <ShieldCheck size={15} /> : <Clock3 size={15} />}
              {device?.verified ? "已验身" : "待检测"}
            </StatusPill>
          </div>
          <dl className="fact-grid">
            <Fact label="芯片" value={device?.chip ?? "未读取"} icon={Cpu} />
            <Fact label="工作模式" value={modeLabel(device?.mode)} icon={Cable} />
            <Fact label="设备尾号" value={device?.macSuffix ?? "未读取"} icon={BadgeCheck} mono />
            <Fact label="USB 身份" value={device?.usbVidPid ?? "未读取"} icon={Usb} mono />
            <Fact label="当前固件" value={device?.firmwareVersion ?? "未读取"} icon={Code2} />
            <Fact label="连接状态" value={device?.connected ? "已连接" : "未连接"} icon={Activity} />
          </dl>

          <SectionHeading icon={PackageCheck} title="目标固件" meta={firmware?.trusted ? "受信来源" : "未受信"} />
          <label className="field">
            <span>GitHub Release</span>
            <select value={selectedFirmwareId} onChange={(event) => onSelectFirmware(event.target.value)} disabled={busy}>
              {snapshot.releases.map((release) => (
                <option key={release.id} value={release.id}>
                  {release.sourceName} · {release.tag}
                </option>
              ))}
            </select>
          </label>
          {firmware && <FirmwareSummary release={firmware} />}
        </div>

        <div className="workspace-column action-column">
          <SectionHeading icon={Zap} title="写入控制" meta={snapshot.progress.step} />
          <div className="progress-panel">
            <div className="progress-meta">
              <span>{snapshot.progress.message}</span>
              <strong>{snapshot.progress.percent}%</strong>
            </div>
            <div className="progress-track" aria-label={`烧录进度 ${snapshot.progress.percent}%`}>
              <span style={{ width: `${snapshot.progress.percent}%` }} />
            </div>
            {snapshot.progress.currentImage && <div className="flash-detail-grid">
              <span>当前镜像 <code>{snapshot.progress.currentImage}</code></span>
              <span>写入地址 <code>{snapshot.progress.currentAddress ?? "读取中"}</code></span>
              <span>当前段 <code>{formatBytes(snapshot.progress.currentBytes ?? 0)} / {formatBytes(snapshot.progress.totalBytes ?? 0)}</code></span>
              <span>校验方式 <code>设备端 Hash</code></span>
            </div>}
          </div>
          <FlashDetailPanel snapshot={snapshot} />

          {complete ? (
            <CompletionPanel snapshot={snapshot} firmware={firmware} />
          ) : (
            <>
              <div className="gate-list">
                <GateRow passed={Boolean(firmware?.trusted)} text="固件来源已信任" />
                <GateRow
                  passed={Boolean(firmware?.checksumVerified)}
                  pending={Boolean(firmware?.trusted && !firmware?.checksumVerified)}
                  text="清单与 SHA-256 在写入前校验"
                />
                <GateRow passed={Boolean(device?.verified)} text="ESP32-S3 与设备身份已核对" />
                <GateRow passed={snapshot.canFlash} text="后端已签发本次写入许可" />
              </div>

              <label className="field confirmation-field">
                <span>写入确认</span>
                <input
                  value={confirmation}
                  onChange={(event) => onConfirmation(event.target.value)}
                  placeholder={expectedConfirmation || "检测设备后生成确认口令"}
                  disabled={busy || !device?.verified}
                  spellCheck={false}
                  autoComplete="off"
                />
                <small>为避免误刷，请完整输入：{expectedConfirmation || "等待设备身份"}</small>
              </label>

              {snapshot.gateReasons.length > 0 && (
                <div className="gate-reasons">
                  {snapshot.gateReasons.map((reason) => (
                    <span key={reason}>
                      <CircleOff size={14} /> {reason}
                    </span>
                  ))}
                </div>
              )}
              {notice && <div className="inline-error">{notice}</div>}

              <div className="action-row">
                {snapshot.progress.stage === "restarting" ? (
                  <button className="button primary wide" type="button" onClick={onRecovery} disabled={pendingAction === "recovery"}>
                    {pendingAction === "recovery" ? <LoaderCircle size={17} className="spin" /> : <Usb size={17} />}
                    检测正常 HID 恢复
                  </button>
                ) : busy && snapshot.progress.canCancel ? (
                  <button className="button danger" type="button" onClick={onCancel} disabled={pendingAction === "cancel"}>
                    <X size={17} />请求停止
                  </button>
                ) : (
                  <button className="button primary wide" type="button" onClick={onStart} disabled={!canStart || pendingAction === "flash"}>
                    {pendingAction === "flash" ? <LoaderCircle size={17} className="spin" /> : <Zap size={17} />}
                    确认并开始烧录
                  </button>
                )}
              </div>
            </>
          )}

          <div className="log-peek">
            <div className="log-peek-head">
              <span><TerminalSquare size={15} />最近活动</span>
              <button type="button" onClick={onViewLogs}>查看全部 <ChevronRight size={14} /></button>
            </div>
            {snapshot.logs.slice(-6).map((log) => <LogRow key={log.id} log={log} />)}
          </div>
        </div>
      </section>
    </div>
  );
}

function FlashDetailPanel({ snapshot }: { snapshot: DashboardSnapshot }) {
  const flashLogs = snapshot.logs.filter((log) => log.scope === "烧录");
  const images = [
    { name: "bootloader.bin", address: "0x0" },
    { name: "partition-table.bin", address: "0x8000" },
    { name: "easy_input_keyboard.bin", address: "0x10000" },
  ];
  return <section className="flash-detail-panel">
    <div className="flash-detail-heading"><strong>完整烧录明细</strong><span>{flashLogs.length ? `${flashLogs.length} 条写入记录` : "等待写入"}</span></div>
    <div className="flash-image-list">{images.map((image) => {
      const related = flashLogs.filter((log) => log.message.includes(image.name));
      const started = related.some((log) => /开始写入|0%/.test(log.message));
      const done = related.some((log) => /写入完成|100%|校验通过/.test(log.message));
      const active = snapshot.progress.currentImage === image.name;
      return <div className={`flash-image-row ${done ? "done" : ""} ${active ? "active" : ""}`} key={image.name}><span className="flash-image-dot">{done ? <Check size={11} /> : active ? <LoaderCircle size={11} className="spin" /> : <span />}</span><div><strong>{image.name}</strong><span>{image.address}{active && snapshot.progress.currentBytes !== undefined ? ` · ${formatBytes(snapshot.progress.currentBytes)} / ${formatBytes(snapshot.progress.totalBytes ?? 0)}` : related[related.length - 1]?.message ?? "尚未开始"}</span></div><StatusPill tone={done ? "success" : active ? "warning" : "neutral"}>{done ? "已完成" : active ? `${snapshot.progress.percent}%` : started ? "已开始" : "等待"}</StatusPill></div>;
    })}</div>
    {flashLogs.length > 0 && <div className="flash-log-lines">{flashLogs.map((log) => <div key={log.id}><time>{log.time}</time><span>{log.message}</span></div>)}</div>}
  </section>;
}

function FlashStepper({ stage }: { stage: FlashStage }) {
  const current = STAGE_INDEX[stage];
  return (
    <div className="flash-stepper" aria-label="烧录阶段">
      {STAGES.map((item, index) => {
        const done = stage === "complete" || index < current;
        const active = index === current && stage !== "complete";
        return (
          <div className={`step ${done ? "done" : ""} ${active ? "active" : ""}`} key={item.id}>
            <span className="step-dot">{done ? <Check size={13} /> : index + 1}</span>
            <span>{item.label}</span>
          </div>
        );
      })}
    </div>
  );
}

function FirmwareSummary({ release }: { release: FirmwareRelease }) {
  return (
    <div className="firmware-summary">
      <div className="release-title-row">
        <div>
          <strong>{release.tag}</strong>
          <span>{release.repository}</span>
        </div>
        <StatusPill tone={release.trusted ? "success" : "warning"}>
          {release.trusted ? <ShieldCheck size={14} /> : <ShieldAlert size={14} />}
          {release.trusted ? "受信" : "需审核"}
        </StatusPill>
      </div>
      <div className="release-metadata">
        <span><Code2 size={14} />{release.commit}</span>
        <span><Cpu size={14} />{release.board}</span>
        <span><Box size={14} />{release.idfVersion}</span>
        <span><HardDrive size={14} />{release.size}</span>
      </div>
      <div className="checksum-line">
        {release.checksumVerified ? <FileCheck2 size={16} /> : <XCircle size={16} />}
        {release.checksumVerified ? "固件清单与文件哈希已通过" : "固件完整性尚未验证"}
      </div>
      {release.isFactory && <div className="factory-warning"><ShieldAlert size={15} />Factory 恢复固定写入 0x0，会清除 NVS 配置和蓝牙绑定。</div>}
    </div>
  );
}

function CompletionPanel({ snapshot, firmware }: { snapshot: DashboardSnapshot; firmware?: FirmwareRelease }) {
  return (
    <div className="completion-panel">
      <div className="completion-title">
        <CheckCircle2 size={24} />
        <div><strong>固件写入已完成</strong><span>设备恢复结果与功能验证单独记录</span></div>
      </div>
      <GateRow passed text="镜像写入与数据校验通过" />
      <GateRow passed={snapshot.progress.hidRecovered} text="正常 USB / HID 模式已发现" />
      <GateRow
        passed={snapshot.progress.functionalVerification === "passed"}
        pending={snapshot.progress.functionalVerification === "pending"}
        text="真实按键、USB / BLE 与 App 功能验证"
      />
      {firmware && (
        <div className="feature-list">
          <span>本版声明功能</span>
          {firmware.features.map((feature) => <FeatureTag key={feature.name} feature={feature} />)}
        </div>
      )}
    </div>
  );
}

function GateRow({ passed, pending, text }: { passed: boolean; pending?: boolean; text: string }) {
  return (
    <div className={`gate-row ${passed ? "passed" : pending ? "pending" : "blocked"}`}>
      {passed ? <CheckCircle2 size={17} /> : pending ? <Clock3 size={17} /> : <CircleOff size={17} />}
      <span>{text}</span>
      <small>{passed ? "通过" : pending ? "待验证" : "未通过"}</small>
    </div>
  );
}

function LibraryPage({
  releases,
  selectedId,
  onSelect,
  onRefresh,
  refreshing,
  onImport,
  onImportFactory,
}: {
  releases: FirmwareRelease[];
  selectedId: string;
  onSelect: (id: string) => void;
  onRefresh: () => void;
  refreshing: boolean;
  onImport: () => void;
  onImportFactory: () => void;
}) {
  return (
    <div className="page">
      <PageHeader
        eyebrow="已收藏来源"
        title="固件库"
        description="集中管理官方与社区 Release；未受信来源默认不能进入烧录流程。"
        actions={<button className="button secondary" type="button" onClick={onRefresh}><RefreshCw size={16} className={refreshing ? "spin" : ""} />刷新</button>}
      />
      <div className="toolbar-row">
        <div className="search-box"><Search size={16} /><input placeholder="搜索仓库、标签或提交" /></div>
        <button className="button secondary" type="button" onClick={onImport}><PackageCheck size={16} />导入本地 ZIP</button>
        <button className="button danger" type="button" onClick={onImportFactory}>导入 Factory 恢复</button>
      </div>
      <section className="table-section">
        <div className="table-heading firmware-table-grid">
          <span>来源与版本</span><span>目标</span><span>完整性</span><span>发布时间</span><span />
        </div>
        {releases.map((release) => (
          <div className={`table-row firmware-table-grid ${selectedId === release.id ? "selected" : ""}`} key={release.id}>
            <div className="primary-cell"><strong>{release.sourceName}</strong><span>{release.repository} · {release.tag}</span></div>
            <div><strong>{release.board}</strong><span>{release.chip} · {release.idfVersion}</span></div>
            <div><StatusPill tone={release.trusted ? "success" : "warning"}>{release.trusted ? <ShieldCheck size={14} /> : <ShieldAlert size={14} />}{release.trusted ? "来源受信" : "需审核"}</StatusPill></div>
            <div><strong>{release.publishedAt}</strong><span>{release.commit}</span></div>
            <button className="icon-button" type="button" onClick={() => onSelect(release.id)} aria-label={`选择 ${release.tag}`}><ChevronRight size={17} /></button>
          </div>
        ))}
      </section>
    </div>
  );
}

function DiscoverPage({ onFirmwareChanged }: { onFirmwareChanged: (snapshot?: DashboardSnapshot) => void }) {
  const [repository, setRepository] = useState("");
  const [audit, setAudit] = useState<FirmwareSourceAudit>();
  const [confirmation, setConfirmation] = useState("");
  const [notice, setNotice] = useState<string>();
  const [pending, setPending] = useState<"audit" | "trust">();
  const expected = audit ? `信任来源 ${audit.repository}` : "";

  const auditSource = async () => {
    setPending("audit");
    setNotice(undefined);
    setAudit(undefined);
    try {
      setAudit(await auditFirmwareSource(repository));
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "仓库审计失败");
    } finally {
      setPending(undefined);
    }
  };

  const trustSource = async () => {
    if (!audit || confirmation.trim() !== expected) return;
    setPending("trust");
    setNotice(undefined);
    try {
      onFirmwareChanged(await trustFirmwareSource(audit.repository, confirmation.trim()));
      setNotice(`${audit.repository} 已加入来源；现在可在固件库刷新并选择 Release`);
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "添加来源失败");
    } finally {
      setPending(undefined);
    }
  };

  return (
    <div className="page">
      <PageHeader eyebrow="GitHub 公开生态" title="添加社区固件来源" description="输入 owner/repository，软件会先只读检查 Release 和自动构建合同，再由你决定是否信任。" />
      <div className="discover-search"><Github size={19} /><input value={repository} onChange={(event) => setRepository(event.target.value)} placeholder="例如 FreeCodeCampXYG/easy-input-maker" /><button type="button" className="button primary" onClick={auditSource} disabled={!repository.trim() || Boolean(pending)}>{pending === "audit" ? <LoaderCircle size={16} className="spin" /> : <Search size={16} />}检查仓库</button></div>
      {notice && <div className="inline-error">{notice}</div>}
      {audit && <section className="source-list">
        <article className="source-row">
          <div className="source-mark tone-2"><Github size={21} /></div>
          <div className="source-main"><strong>{audit.repository}</strong><span>{audit.validReleases} 个可烧录 Release</span><p>通过后会显示在固件库；每次写入仍会重新校验 manifest 和 SHA-256。</p></div>
          <StatusPill tone={audit.ready ? "success" : "warning"}>{audit.ready ? <ShieldCheck size={14} /> : <ShieldAlert size={14} />}{audit.ready ? "可申请信任" : "暂不可用"}</StatusPill>
        </article>
        <div className="gate-list">{audit.checks.map((check) => <GateRow key={check.name} passed={check.passed} pending={!check.passed} text={`${check.name}：${check.message}`} />)}</div>
        {audit.ready && <label className="field"><span>输入 “{expected}” 后加入来源</span><input value={confirmation} onChange={(event) => setConfirmation(event.target.value)} placeholder={expected} /><button type="button" className="button primary" onClick={trustSource} disabled={confirmation.trim() !== expected || pending === "trust"}>{pending === "trust" ? <LoaderCircle size={16} className="spin" /> : <ShieldCheck size={16} />}信任并加入</button></label>}
      </section>}
      <div className="safety-note"><ShieldAlert size={19} /><div><strong>只添加完整 Release，不烧录分支或裸 bin</strong><span>缺少三段镜像时，按 README 的图文教程为 Fork 补 GitHub Actions 自动编译与 manifest 发布流程。</span></div></div>
    </div>
  );
}

function DevicesPage({ devices, logs, onExport }: { devices: DeviceInfo[]; logs: ActivityLog[]; onExport: () => Promise<{ ok: boolean; message?: string }> }) {
  const [exportNotice, setExportNotice] = useState<{ ok: boolean; message?: string }>();
  const exportLogs = async () => setExportNotice(await onExport());
  return (
    <div className="page">
      <PageHeader eyebrow="本机诊断" title="设备与日志" description="查看端口、设备身份与脱敏运行记录。日志不会保存完整 MAC、用户文本或快捷键内容。" actions={<button className="button secondary" type="button" onClick={exportLogs}><Download size={16} />导出诊断</button>} />
      {exportNotice?.message && <div className={`diagnostic-notice ${exportNotice.ok ? "success" : "error"}`}>{exportNotice.message}</div>}
      <div className="device-strip">
        {devices.map((device) => (
          <div className="device-summary" key={device.id}>
            <div className="device-icon"><Usb size={22} /></div>
            <div><strong>{device.product}</strong><span>{device.port} · {device.chip} · 尾号 {device.macSuffix}</span></div>
            <StatusPill tone={device.verified ? "success" : "neutral"}>{device.verified ? <ShieldCheck size={14} /> : <Clock3 size={14} />}{device.verified ? "身份已核对" : "待检测"}</StatusPill>
          </div>
        ))}
      </div>
      <section className="log-console">
        <div className="console-toolbar"><span><TerminalSquare size={16} />运行日志</span><div><button type="button">全部</button><button type="button">设备</button><button type="button">固件</button><button type="button">错误</button></div></div>
        <div className="console-lines">
          {logs.map((log) => <LogRow key={log.id} log={log} detailed />)}
        </div>
      </section>
    </div>
  );
}

const BOARD_DIAGNOSTIC_TIPS: Record<string, string> = {
  chip: "它是这块板的主控。只有进入下载模式后，烧录器才能确认型号；普通 HID 在线不等于芯片已验身。",
  flash: "它保存启动程序、分区表和固件。烧录器只按受信清单写固定地址，不会根据文件名猜写入位置。",
  psram: "它是固件运行时使用的高速缓存内存。烧录器读不到“8 MB 已可用”，要看固件自检结果。",
  keys: "S1-S8 是八个独立低有效按键，可同时按下；具体快捷键由当前固件决定。",
  encoder: "旋钮的 A/B 是方向相位，按压是独立按键。顺时针、逆时针各转一次，再按一下；方向和功能由固件决定。",
  led: "5 颗 WS2812 共用一根数据线，另有一颗绿色状态灯。颜色和动画不是板级固定含义。",
  audio_in: "麦克风把声音送进固件，可用于录音、音量或语音功能；是否支持取决于当前固件。",
  audio_out: "扬声器用于播放提示音或测试音。测试时从小音量开始，音频格式由固件决定。",
  power: "这里能观察供电、充电和电池端电压。VBAT 只是电压估算，不是精确剩余电量。",
  battery_capacity: "板子没有独立电量计，剩余百分比只能由固件根据电压、负载和校准估算；这个百分比不能当成精确电量。",
  bluetooth: "蓝牙是 ESP32-S3 自带的无线能力。正常开机后，在系统蓝牙列表找设备并做一次输入测试；设备名和协议由固件决定。",
  wifi: "Wi-Fi 也是 ESP32-S3 自带能力。只有当前固件明确启用联网功能时才测试扫描或联网，板子有模块不代表每个固件都开启。",
  usb_hid: "USB HID 是有线输入通道。正常开机后打开文本框，按一个已知有动作的按键，看到真实字符或快捷键进入系统后再标记。",
  usb_ble: "烧录后关机再开机，电脑应重新看到正常 USB/BLE HID；这不等于每个快捷键都已验收。",
};

const BOARD_DIAGNOSTIC_STEPS: Record<string, string[]> = {
  keys: ["先关机，再正常开机退出下载模式", "依次按 S1 到 S8，每个按键按下并松开一次", "观察当前固件的输入反馈（例如电脑按键、界面日志或灯效）后再标记"],
  encoder: ["先关机，再正常开机退出下载模式", "顺时针转一格、逆时针转一格，再按下旋钮一次", "观察当前固件是否有对应反馈；方向和动作由固件定义"],
  led: ["先关机，再正常开机运行固件", "按当前固件提供的灯光测试入口触发灯效", "确认 5 颗灯和绿色状态灯有响应；不要把灯灭当作断电证明"],
  audio_in: ["先关机，再正常开机运行固件", "打开当前固件的录音、音量或麦克风测试入口", "对着麦克风说话，观察电平或录音回放是否变化"],
  audio_out: ["先关机，再正常开机运行固件", "从小音量开始触发当前固件的测试音", "确认扬声器有声音且无明显杂音，测试后再恢复合适音量"],
  usb_ble: ["先关机，再正常开机退出下载模式", "等待电脑重新枚举 USB/BLE HID", "打开文本框按一个已知有动作的键，确认系统收到输入"],
  battery_capacity: ["先关机，再正常开机运行固件", "打开固件的电池页面或状态输出", "将显示的百分比当作估算值，与实际电压/充电状态一起判断"],
  bluetooth: ["先关机，再正常开机运行固件", "在系统蓝牙列表查找并连接当前固件广播的设备", "用一个已知有动作的按键验证真实输入，不要只看“能搜到”"],
  wifi: ["先关机，再正常开机运行支持 Wi-Fi 的固件", "打开固件提供的扫描或联网入口", "确认扫描结果或联网状态；没有对应固件功能时保持待验证"],
  usb_hid: ["先关机，再正常开机退出下载模式", "等待电脑重新枚举 USB HID", "打开文本框按一个已知有动作的按键，确认系统收到真实输入"],
};

function DiagnosticsPage({ devices, releases, selectedFirmwareId, onSelectFirmware, onRun, onRead }: { devices: DeviceInfo[]; releases: FirmwareRelease[]; selectedFirmwareId: string; onSelectFirmware: (id: string) => void; onRun: (deviceId: string) => Promise<HardwareDiagnosticSnapshot>; onRead: (deviceId: string) => Promise<import("./types").HardwareDiagnosticTelemetry> }) {
  const [deviceId, setDeviceId] = useState(devices[0]?.id ?? "");
  const [report, setReport] = useState<HardwareDiagnosticSnapshot>();
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const mark = (key: string, status: "passed" | "blocked") => setReport((current) => current ? { ...current, items: current.items.map((item) => item.key === key ? { ...item, status, detail: status === "passed" ? "用户已确认本项表现正常" : "用户标记为异常，请导出诊断并检查硬件" } : item) } : current);
  const run = async () => { if (!deviceId) return; setBusy(true); setNotice(""); try { setReport(await onRun(deviceId)); } catch (error) { setNotice(error instanceof Error ? error.message : "硬件诊断失败"); } finally { setBusy(false); } };
  const selected = devices.find((item) => item.id === deviceId);
  useEffect(() => {
    if (!report || selected?.mode !== "normal") return undefined;
    let disposed = false;
    const poll = async () => {
      try {
        const telemetry = await onRead(deviceId);
        if (disposed) return;
        setReport((current) => {
          if (!current) return current;
          const previous = current.telemetry;
          const next = { ...current, telemetry };
          if (previous && telemetry.inputEvents > previous.inputEvents) {
            next.items = next.items.map((item) => item.key === "keys" && telemetry.lastInput.startsWith("KEY") ? { ...item, status: "passed", detail: `已收到真实输入事件：${telemetry.lastInput}` } : item);
          }
          if (previous && telemetry.encoderSteps > previous.encoderSteps) {
            next.items = next.items.map((item) => item.key === "encoder" ? { ...item, status: "passed", detail: `已收到旋钮步进事件（累计 ${telemetry.encoderSteps}）` } : item);
          }
          return next;
        });
      } catch { /* 设备切换或 HID 暂不可读时保留上一份证据，下一轮继续尝试。 */ }
    };
    void poll();
    const timer = window.setInterval(() => void poll(), 900);
    return () => { disposed = true; window.clearInterval(timer); };
  }, [deviceId, onRead, report, selected?.mode]);
  const groups = [
    { key: "auto", title: "自动验身", description: "只读取下载模式能证明的芯片与存储身份。", items: report?.items.filter((item) => item.evidence === "自动检测") ?? [] },
    { key: "operate", title: "用户操作", description: "按提示逐项操作；通过/异常均是用户确认，不是 ROM 自动结论。", items: report?.items.filter((item) => item.evidence === "用户操作") ?? [] },
    { key: "observe", title: "运行态观察", description: "这些项目必须运行固件后观察，烧录器不会伪造通过。", items: report?.items.filter((item) => ["用户观察", "运行态观察", "系统枚举"].includes(item.evidence)) ?? [] },
    { key: "unknown", title: "待固件验证", description: "板级存在，但当前工具无法从 ROM 单独证明。", items: report?.items.filter((item) => item.status === "unknown") ?? [] },
  ].filter((group) => group.items.length > 0);
  const counts = report?.items.reduce((result, item) => { result[item.status] += 1; return result; }, { passed: 0, pending: 0, blocked: 0, unknown: 0 } as Record<string, number>);
  return <div className="page">
    <PageHeader eyebrow="EasyInput V2.0 / 板级体检" title="硬件诊断" description="按开发板真实能力分层：先验身，再做用户操作，最后进行运行态观察。每项都保留证据来源和未验证边界。" actions={<button className="button primary" type="button" onClick={run} disabled={!deviceId || busy}>{busy ? <LoaderCircle size={16} className="spin" /> : <Activity size={16} />}开始诊断</button>} />
    <section className="diagnostic-intro"><div><strong>推荐顺序</strong><span>正常模式先确认设备在线 → 开机短按并松开一次 BOOT → 刷新下载端口 → 读取 ESP32-S3 / Flash ID。自动验身完成后，请关机再正常开机，回到正常 HID 模式，再测试按键、旋钮、灯光、音频和 HID。</span></div><StatusPill tone={selected?.mode === "download" ? "success" : "warning"}>{selected?.mode === "download" ? "当前可做自动验身" : "正常模式可做外设测试"}</StatusPill></section>
    <div className="diagnostic-toolbar"><label className="field grow"><span>目标设备</span><select value={deviceId} onChange={(event) => { setDeviceId(event.target.value); setReport(undefined); }}><option value="">未发现 EasyInput 设备</option>{devices.map((item) => <option key={item.id} value={item.id}>{item.port} · {item.product} · {item.mode === "download" ? "下载模式" : item.mode === "normal" ? "正常 HID" : "待确认"}</option>)}</select></label><label className="field grow"><span>教学固件</span><select value={selectedFirmwareId} onChange={(event) => onSelectFirmware(event.target.value)}><option value="">未选择固件</option>{releases.map((release) => <option key={release.id} value={release.id}>{release.sourceName} · {release.tag}</option>)}</select></label><div className="diagnostic-legend"><span><CheckCircle2 size={14} />已通过</span><span><Clock3 size={14} />待操作</span><span><Info size={14} />待验证</span></div></div>
    {(() => { const firmware = releases.find((item) => item.id === selectedFirmwareId); return firmware ? <section className="diagnostic-firmware-tip"><div><strong>{firmware.sourceName} · {firmware.tag}</strong><span>{firmware.isFactory ? "Factory 恢复镜像：用于恢复出厂，不是普通升级。" : `适配 ${firmware.board} / ${firmware.chip}；具体功能以该版本清单和发布说明为准。`}</span></div><div className="diagnostic-firmware-notes">{(firmware.features.length ? firmware.features.map((feature) => feature.name) : firmware.changelog).slice(0, 4).map((text) => <span key={text}>{text}</span>)}</div></section> : null; })()}
    {notice && <div className="inline-error">{notice}</div>}
    {!report ? <div className="diagnostic-empty"><Activity size={28} /><strong>{deviceId ? "点击“开始诊断”读取板级证据" : "先选择已识别的 EasyInput 设备"}</strong><span>不会扫描或写入无关设备；没有下载端口时，自动验身会明确停在 BOOT 引导，不会显示伪造芯片信息。</span></div> : <><div className="diagnostic-summary"><strong>本次诊断结果</strong><span>通过 {counts?.passed ?? 0} · 待操作 {counts?.pending ?? 0} · 阻断 {counts?.blocked ?? 0} · 待验证 {counts?.unknown ?? 0}</span></div>{report.telemetry?.supported && <div className="diagnostic-live"><strong>正在读取真实固件事件</strong><span>最近输入：{report.telemetry.lastInput || "等待按键或旋钮"} · 输入事件：{report.telemetry.inputEvents} · 旋钮步进：{report.telemetry.encoderSteps} · 电池：{report.telemetry.batteryMv} mV / 估算 {report.telemetry.batteryPercent}%</span></div>}<section className="diagnostic-groups">{groups.map((group) => <div className="diagnostic-group" key={group.key}><div className="diagnostic-group-heading"><div><strong>{group.title}</strong><span>{group.description}</span></div><span>{group.items.length} 项</span></div><div className="diagnostic-grid">{group.items.map((item) => <article className={`diagnostic-card ${item.status}`} key={item.key}><div className="diagnostic-card-heading"><div><strong>{item.label}</strong><span>{item.evidence}</span></div><StatusPill tone={item.status === "passed" ? "success" : item.status === "unknown" ? "neutral" : "warning"}>{item.status === "passed" ? "通过" : item.status === "blocked" ? "需先验身" : item.status === "unknown" ? "待验证" : "待操作"}</StatusPill></div><p>{item.detail}</p><div className="diagnostic-tip"><strong>怎么用</strong><span>{BOARD_DIAGNOSTIC_TIPS[item.key] ?? "请以当前固件说明和实板表现为准。"}</span></div>{BOARD_DIAGNOSTIC_STEPS[item.key] && <div className="diagnostic-steps"><strong>测试步骤</strong><ol>{BOARD_DIAGNOSTIC_STEPS[item.key].map((step) => <li key={step}>{step}</li>)}</ol></div>}{item.status === "pending" && <div className="diagnostic-actions"><button type="button" className="button secondary" onClick={() => mark(item.key, "passed")}>标记通过</button><button type="button" className="button ghost" onClick={() => mark(item.key, "blocked")}>标记异常</button></div>}</article>)}</div></div>)}</section></>}
  </div>;
}

function UpdatesPage({ currentVersion, availableVersion, onCheck, checking }: { currentVersion: string; availableVersion?: string; onCheck: () => void; checking: boolean }) {
  const timeline = [
    { version: "v0.1.12", date: "2026-09-01", title: "本地固件与 Factory 恢复", items: ["固定格式本地 ZIP 校验", "CY Factory Release 兼容与 0x0 恢复写入", "Factory 清除 NVS/蓝牙绑定的独立风险确认"] },
    { version: "v0.1.11", date: "2026-08-31", title: "社区固件来源门禁", items: ["Release、workflow 和 manifest 脚本审计", "受信来源 allow-list 与三段镜像固定偏移", "社区共建规范和内嵌帮助说明"] },
    { version: "v0.1.10", date: "2026-08-31", title: "烧录详情与主题优化", items: ["当前镜像、写入地址、字节进度和 Hash 校验", "烧录页最近活动日志", "深色/浅色主题详情卡适配"] },
    { version: "规划", date: "后续版本", title: "本地工程构建扩展", items: ["检测本机 ESP-IDF 并调用 idf.py build", "读取 flasher_args 生成临时清单", "更多实板元件诊断记录"] },
  ];
  return (
    <div className="page">
      <PageHeader eyebrow="版本时间轴" title="更新与通知" description="应用更新与固件源更新分别展示，任何更新都不会自动触发设备写入。" actions={<button className="button secondary" type="button" onClick={onCheck}><RefreshCw size={16} className={checking ? "spin" : ""} />检查更新</button>} />
      <div className="update-status"><div className="update-icon"><Sparkles size={22} /></div><div><strong>{availableVersion ? `发现新版本 ${availableVersion}` : "当前已是最新版本"}</strong><span>当前版本 {currentVersion} · GitHub Release 通道</span></div>{availableVersion && <button className="button primary" type="button"><CloudDownload size={16} />查看更新</button>}</div>
      <section className="timeline">
        {timeline.map((item) => (
          <article className={`timeline-item ${item.version === "v0.1.12" ? "current" : ""}`} key={`${item.version}-${item.date}`}>
            <div className="timeline-marker" />
            <div className="timeline-date">{item.date}</div>
            <div className="timeline-content"><div><strong>{item.title}</strong><StatusPill tone={item.version === "规划" ? "neutral" : "success"}>{item.version}</StatusPill></div><ul>{item.items.map((entry) => <li key={entry}>{entry}</li>)}</ul></div>
          </article>
        ))}
      </section>
    </div>
  );
}

function AboutPage({ version }: { version: string }) {
  return (
    <div className="page">
      <PageHeader eyebrow="产品与支持" title="EasyInput 固件烧录与硬件诊断工具" description="面向 EasyInput V2.0 的可审计烧录与板级体检工作台。" />
      <section className="about-intro">
        <img src="/app-icon.png" alt="EasyInput Flasher 图标" />
        <div><h2>EasyInput 固件烧录与硬件诊断工具</h2><p>EasyInput Flasher · 版本 {version}</p><span>Wails + Go 桌面端 · React 工作台</span></div>
      </section>
      <div className="about-grid">
        <section><BookOpen size={20} /><div><strong>使用手册（内嵌）</strong><p>设备检测、下载模式、写入确认、恢复启动与故障排查。</p><a className="button secondary" href="#flasher-help">查看本页说明 <ChevronRight size={14} /></a></div></section>
        <section><Github size={20} /><div><strong>EasyInput Flasher 仓库</strong><p>查看烧录器源码、Release、已知问题与公开路线图。</p><a className="button secondary" href="https://github.com/FreeCodeCampXYG/easyinput-flasher" target="_blank" rel="noreferrer">访问 Flasher GitHub <ExternalLink size={14} /></a></div></section>
        <section><Github size={20} /><div><strong>CY 老师固件仓库</strong><p>查看官方 Maker 固件与 Factory 恢复 Release。</p><a className="button secondary" href="https://github.com/CY-CHENYUE/easy-input-maker/releases/tag/v0.4.53" target="_blank" rel="noreferrer">查看 v0.4.53 Release <ExternalLink size={14} /></a></div></section>
        <section><ShieldCheck size={20} /><div><strong>安全边界</strong><p>不默认整片擦除；每次写入都重新验身并要求明确确认。</p><button type="button">查看烧录规范 <ChevronRight size={14} /></button></div></section>
        <section><Settings2 size={20} /><div><strong>开源组件</strong><p>Wails、React、Lucide 与经过版本锁定的烧录工具。</p><button type="button">第三方声明 <ChevronRight size={14} /></button></div></section>
      </div>
      <section id="flasher-help" className="help-embed">
        <div className="section-heading"><BookOpen size={17} /><span>快速使用说明</span><small>无需打开外部网页</small></div>
        <div className="help-steps">
          <details open><summary>1. 进入下载模式</summary><p>设备正常开机并连接数据线后，先确认界面识别到正常 HID；然后开机短按并松开 BOOT，点击“刷新下载端口”。不要按住 BOOT 再上电。</p></details>
          <details><summary>2. 选择完整固件</summary><p>在“固件库”选择带有 <code>firmware-manifest.json</code>、<code>bootloader.bin</code>、<code>partition-table.bin</code> 和 <code>easy_input_keyboard.bin</code> 的 Release。整片 <code>factory.bin</code> 不能直接当作安全补齐包。</p></details>
          <details><summary>3. 验身、确认与恢复</summary><p>下载后软件会重新校验 manifest 与 SHA-256，并显示 ESP32-S3 与 MAC 尾号。输入界面给出的确认文本后开始写入；完成后关机再正常开机，再点击“检测正常 HID 恢复”。</p></details>
          <details><summary>4. 社区仓库文件不足怎么办</summary><p>“发现”页会提示缺少的发布资产。若仓库只有 factory.bin，软件不会自动拆分或猜测偏移，以免覆盖配置分区；请按仓库 Actions 生成标准三段镜像和 manifest，再重新检查 Release。</p></details>
        </div>
      </section>
      <div className="about-footnote">EasyInput Flasher 不会把“构建成功”“写入成功”“设备枚举”和“真实功能验证”合并为同一结论。</div>
    </div>
  );
}

function StatusBar({ snapshot, device }: { snapshot: DashboardSnapshot; device?: DeviceInfo }) {
  return (
    <footer className="status-bar">
      <span className={snapshot.network.online ? "ok" : "error"}><span className="status-dot" />{snapshot.network.online ? "GitHub 可访问" : "网络不可用"}</span>
      <span><Network size={13} />{snapshot.network.proxyMode === "custom" ? snapshot.network.proxyAddress : proxyLabel(snapshot.network.proxyMode)}</span>
      <span><Database size={13} />缓存 {snapshot.cache.items} 项 · {snapshot.cache.size}</span>
      <span><Usb size={13} />{device ? `${device.port} · ${device.chip}` : "未连接设备"}</span>
      <span className="status-version">v{snapshot.appVersion}</span>
    </footer>
  );
}

function PageHeader({ eyebrow, title, description, actions }: { eyebrow: string; title: string; description: string; actions?: ReactNode }) {
  return (
    <header className="page-header"><div><span className="eyebrow">{eyebrow}</span><h1>{title}</h1><p>{description}</p></div>{actions && <div className="header-actions">{actions}</div>}</header>
  );
}

function SectionHeading({ icon: Icon, title, meta }: { icon: typeof Usb; title: string; meta?: string }) {
  return <div className="section-heading"><span><Icon size={17} />{title}</span>{meta && <small>{meta}</small>}</div>;
}

function Fact({ label, value, icon: Icon, mono }: { label: string; value: string; icon: typeof Usb; mono?: boolean }) {
  return <div className="fact"><dt><Icon size={14} />{label}</dt><dd className={mono ? "mono" : ""}>{value}</dd></div>;
}

function StatusPill({ tone, children }: { tone: "success" | "warning" | "neutral"; children: ReactNode }) {
  return <span className={`status-pill ${tone}`}>{children}</span>;
}

function FeatureTag({ feature }: { feature: FirmwareFeature }) {
  return <span className={`feature-tag ${feature.status}`}><Check size={12} />{feature.name}<small>{feature.status === "verified" ? "已验证" : feature.status === "declared" ? "固件声明" : "不可用"}</small></span>;
}

function LogRow({ log, detailed }: { log: ActivityLog; detailed?: boolean }) {
  return <div className={`log-row log-${log.level} ${detailed ? "detailed" : ""}`}><time>{log.time}</time><span className="log-level" /><strong>{log.scope}</strong><p>{log.message}</p></div>;
}

function modeLabel(mode?: DeviceInfo["mode"]) {
  if (mode === "download") return "下载模式";
  if (mode === "normal") return "正常 HID";
  return "未识别";
}

function formatBytes(value: number): string {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / (1024 * 1024)).toFixed(2)} MB`;
}

function proxyLabel(mode: DashboardSnapshot["network"]["proxyMode"]) {
  if (mode === "system") return "系统代理";
  if (mode === "direct") return "直连";
  return "自定义代理";
}

export default App;

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
  getDashboardSnapshot,
  inspectDevice,
  listFirmware,
  scanDevices,
  startFlash,
} from "./bridge";
import { demoSnapshot } from "./demo-data";
import type {
  ActivityLog,
  DashboardSnapshot,
  DeviceInfo,
  FirmwareFeature,
  FirmwareRelease,
  FlashStage,
  PageId,
} from "./types";

const NAV_ITEMS: Array<{ id: PageId; label: string; icon: typeof Zap }> = [
  { id: "flash", label: "烧录", icon: Zap },
  { id: "library", label: "固件库", icon: Library },
  { id: "discover", label: "发现", icon: Search },
  { id: "devices", label: "设备与日志", icon: MonitorCog },
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
  stage === "detecting" || stage === "downloading" || stage === "flashing" || stage === "restarting";

function App() {
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
    getDashboardSnapshot().then(applySnapshot).catch(() => undefined);
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
  const expectedConfirmation = device?.macSuffix ? `确认烧录 ${device.macSuffix}` : "";
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
          />
        );
      case "discover":
        return <DiscoverPage />;
      case "devices":
        return <DevicesPage devices={snapshot.devices} logs={snapshot.logs} />;
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
          />
        );
    }
  };

  return (
    <div className="app-shell">
      <Sidebar
        page={page}
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onNavigate={(next) => {
          setPage(next);
          setSidebarOpen(false);
        }}
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
}

function Sidebar({ page, open, onClose, onNavigate }: SidebarProps) {
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
  } = props;
  const busy = isBusyStage(snapshot.progress.stage);
  const complete = snapshot.progress.stage === "complete";

  return (
    <div className="page page-flash">
      <PageHeader
        eyebrow="固件交付工作台"
        title="安全烧录"
        description="从受信 GitHub Release 获取固件，完成设备验身、写入校验和启动恢复。"
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
              读取信息
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
            {snapshot.progress.currentImage && <code>{snapshot.progress.currentImage}</code>}
          </div>

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
              <button type="button">查看全部 <ChevronRight size={14} /></button>
            </div>
            {snapshot.logs.slice(-4).map((log) => <LogRow key={log.id} log={log} />)}
          </div>
        </div>
      </section>
    </div>
  );
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
}: {
  releases: FirmwareRelease[];
  selectedId: string;
  onSelect: (id: string) => void;
  onRefresh: () => void;
  refreshing: boolean;
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
        <button className="button secondary" type="button"><Github size={16} />添加 GitHub 源</button>
      </div>
      <section className="table-section">
        <div className="table-heading firmware-table-grid">
          <span>来源与版本</span><span>目标</span><span>完整性</span><span>发布时间</span><span />
        </div>
        {releases.map((release) => (
          <div className={`table-row firmware-table-grid ${selectedId === release.id ? "selected" : ""}`} key={release.id}>
            <div className="primary-cell"><strong>{release.sourceName}</strong><span>{release.repository} · {release.tag}</span></div>
            <div><strong>{release.board}</strong><span>{release.chip} · {release.idfVersion}</span></div>
            <div><StatusPill tone={release.trusted && release.checksumVerified ? "success" : "warning"}>{release.trusted ? <ShieldCheck size={14} /> : <ShieldAlert size={14} />}{release.trusted ? "已校验" : "需审核"}</StatusPill></div>
            <div><strong>{release.publishedAt}</strong><span>{release.commit}</span></div>
            <button className="icon-button" type="button" onClick={() => onSelect(release.id)} aria-label={`选择 ${release.tag}`}><ChevronRight size={17} /></button>
          </div>
        ))}
      </section>
    </div>
  );
}

function DiscoverPage() {
  const sources = [
    { name: "EasyInput Maker", repo: "FreeCodeCampXYG/easy-input-maker", stars: "官方", releases: 3, trusted: true },
    { name: "Community Layouts", repo: "community/easyinput-layouts", stars: "128", releases: 8, trusted: false },
    { name: "Macro Profiles", repo: "makers/easyinput-macro-pack", stars: "86", releases: 5, trusted: false },
  ];
  return (
    <div className="page">
      <PageHeader eyebrow="GitHub 公开生态" title="发现" description="按仓库、Topic 与 Release 浏览兼容项目；收藏不代表受信，烧录前仍需清单校验。" />
      <div className="discover-search"><Search size={19} /><input placeholder="搜索 GitHub 仓库或粘贴 Release 链接" /><button type="button" className="button primary">搜索</button></div>
      <div className="filter-tabs"><button className="active">精选</button><button>最新发布</button><button>最多收藏</button><button>兼容 V2</button></div>
      <section className="source-list">
        {sources.map((source, index) => (
          <article className="source-row" key={source.repo}>
            <div className={`source-mark tone-${index + 1}`}><Github size={21} /></div>
            <div className="source-main"><strong>{source.name}</strong><span>{source.repo}</span><p>{index === 0 ? "EasyInput V2.0 公共固件与 Host Action 基线。" : "社区提供的扩展固件来源，使用前需要审核发布清单和目标板型。"}</p></div>
            <div className="source-stats"><span><Star size={14} />{source.stars}</span><span><PackageCheck size={14} />{source.releases} 个版本</span></div>
            <StatusPill tone={source.trusted ? "success" : "neutral"}>{source.trusted ? <ShieldCheck size={14} /> : <Clock3 size={14} />}{source.trusted ? "官方受信" : "未审核"}</StatusPill>
            <button className="icon-button" type="button" aria-label={`收藏 ${source.name}`}><Star size={17} /></button>
          </article>
        ))}
      </section>
      <div className="safety-note"><ShieldAlert size={19} /><div><strong>发现结果不会自动获得写入权限</strong><span>社区项目必须声明 ESP32-S3、EasyInput V2 板型、分区偏移与 SHA-256；不完整的 Release 只能收藏和查看。</span></div></div>
    </div>
  );
}

function DevicesPage({ devices, logs }: { devices: DeviceInfo[]; logs: ActivityLog[] }) {
  return (
    <div className="page">
      <PageHeader eyebrow="本机诊断" title="设备与日志" description="查看端口、设备身份与脱敏运行记录。日志不会保存完整 MAC、用户文本或快捷键内容。" actions={<button className="button secondary" type="button"><Download size={16} />导出诊断</button>} />
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

function UpdatesPage({ currentVersion, availableVersion, onCheck, checking }: { currentVersion: string; availableVersion?: string; onCheck: () => void; checking: boolean }) {
  const timeline = [
    { version: "v0.1.0", date: "2026-08-30", title: "首个公开预览版", items: ["受信 GitHub Release 下载", "ESP32-S3 设备验身与烧录状态机", "固件库、发现、日志与帮助中心"] },
    { version: "规划", date: "下一版本", title: "社区来源审核与自动更新", items: ["来源订阅与变更通知", "构建溯源证明展示", "更多实机验证结果"] },
  ];
  return (
    <div className="page">
      <PageHeader eyebrow="版本时间轴" title="更新与通知" description="应用更新与固件源更新分别展示，任何更新都不会自动触发设备写入。" actions={<button className="button secondary" type="button" onClick={onCheck}><RefreshCw size={16} className={checking ? "spin" : ""} />检查更新</button>} />
      <div className="update-status"><div className="update-icon"><Sparkles size={22} /></div><div><strong>{availableVersion ? `发现新版本 ${availableVersion}` : "当前已是最新版本"}</strong><span>当前版本 {currentVersion} · GitHub Release 通道</span></div>{availableVersion && <button className="button primary" type="button"><CloudDownload size={16} />查看更新</button>}</div>
      <section className="timeline">
        {timeline.map((item) => (
          <article className="timeline-item" key={`${item.version}-${item.date}`}>
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
      <PageHeader eyebrow="产品与支持" title="关于 EasyInput Flasher" description="面向 EasyInput V2.0 的可审计固件交付工具，让使用者无需本地 ESP-IDF 环境即可获取已构建产物。" />
      <section className="about-intro">
        <img src="/app-icon.png" alt="EasyInput Flasher 图标" />
        <div><h2>EasyInput Flasher</h2><p>版本 {version}</p><span>Wails + Go 桌面端 · React 工作台</span></div>
      </section>
      <div className="about-grid">
        <section><BookOpen size={20} /><div><strong>使用手册</strong><p>设备检测、下载模式、写入确认、恢复启动与故障排查。</p><button type="button">打开帮助 <ExternalLink size={14} /></button></div></section>
        <section><Github size={20} /><div><strong>开源仓库</strong><p>查看源码、Release、已知问题与公开路线图。</p><button type="button">访问 GitHub <ExternalLink size={14} /></button></div></section>
        <section><ShieldCheck size={20} /><div><strong>安全边界</strong><p>不默认整片擦除；每次写入都重新验身并要求明确确认。</p><button type="button">查看烧录规范 <ChevronRight size={14} /></button></div></section>
        <section><Settings2 size={20} /><div><strong>开源组件</strong><p>Wails、React、Lucide 与经过版本锁定的烧录工具。</p><button type="button">第三方声明 <ChevronRight size={14} /></button></div></section>
      </div>
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

function proxyLabel(mode: DashboardSnapshot["network"]["proxyMode"]) {
  if (mode === "system") return "系统代理";
  if (mode === "direct") return "直连";
  return "自定义代理";
}

export default App;

/**
 * ReflectorPage — full Reflector control surface.
 *
 * After the #66 redesign this page owns the controls that used to
 * live in the legacy App.tsx top bar: interface picker (with the
 * usable-filter toggle), profile picker (NetAlly / MSN / All /
 * Custom), Start/Stop buttons, live counters, and the selected
 * interface detail card.
 *
 * Wraps everything in a RoleGuard so a Test-Master stem prompts the
 * operator to switch roles before using the reflector.
 */
import {
  Activity,
  AlertTriangle,
  Clock,
  Gauge,
  Play,
  RefreshCw,
  Repeat,
  Square,
  Wifi,
} from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { HeaderInterfaceSelector } from '../components/HeaderInterfaceSelector';
import { RoleGuard } from '../components/RoleGuard';
import { ReflectorSection } from '../components/settings/ReflectorSection';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { useAppContext } from '../contexts/AppContext';
import { useRole } from '../contexts/RoleContext';
import { useCapabilities } from '../hooks/useCapabilities';
import type { InterfaceInfo, Stats } from '../types/api';
import { Breadcrumbs } from '../ui/Breadcrumbs';
import { PageHeader } from '../ui/PageHeader';

function formatNumber(num: number): string {
  if (num >= 1e9) {
    return `${(num / 1e9).toFixed(2)}B`;
  }
  if (num >= 1e6) {
    return `${(num / 1e6).toFixed(2)}M`;
  }
  if (num >= 1e3) {
    return `${(num / 1e3).toFixed(2)}K`;
  }
  return num.toString();
}

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}

function getStatusClassName(status: Stats['testStatus']): string {
  switch (status) {
    case 'running':
      return 'text-status-success';
    case 'error':
      return 'text-status-error';
    case 'completed':
      return 'text-status-info';
    case 'starting':
      return 'text-status-info';
    case 'cancelled':
      return 'text-status-warning';
    default:
      return 'text-text-muted';
  }
}

interface StatsCardProps {
  icon: React.ReactNode;
  title: string;
  value: string;
  subvalue: string;
}

function StatsCard({ icon, title, value, subvalue }: StatsCardProps): ReactElement {
  return (
    <div className="card">
      <div className="card-header">
        {icon}
        {title}
      </div>
      <div className="card-value">{value}</div>
      <div className="card-subvalue">{subvalue}</div>
    </div>
  );
}

interface InterfaceDetailsProps {
  iface: InterfaceInfo;
}

function InterfaceDetails({ iface }: InterfaceDetailsProps): ReactElement {
  const stateClassName = iface.state === 'up' ? 'text-status-success' : 'text-status-error';
  return (
    <div className="card mb-2">
      <div className="card-header">
        <Wifi className="w-4 h-4" />
        Interface Details
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-comfortable text-sm">
        <div>
          <div className="text-text-muted">Name</div>
          <div className="font-medium">{iface.name}</div>
        </div>
        <div>
          <div className="text-text-muted">MAC</div>
          <div className="font-mono">{iface.mac}</div>
        </div>
        <div>
          <div className="text-text-muted">Speed</div>
          <div>
            {iface.speed} Mbps / {iface.duplex}
          </div>
        </div>
        <div>
          <div className="text-text-muted">Driver</div>
          <div>{iface.driver}</div>
        </div>
        <div>
          <div className="text-text-muted">State</div>
          <div className={stateClassName}>{iface.state}</div>
        </div>
        <div>
          <div className="text-text-muted">XDP Support</div>
          <div>{iface.xdp ? 'Yes' : 'No'}</div>
        </div>
        <div>
          <div className="text-text-muted">DPDK Support</div>
          <div>{iface.dpdk ? 'Yes' : 'No'}</div>
        </div>
        <div>
          <div className="text-text-muted">Score</div>
          <div>{iface.score}</div>
        </div>
      </div>
    </div>
  );
}

interface PlatformBannerProps {
  reason: string;
  onSwitchToTestMaster: () => void;
}

// Surfaces when the backend reports reflector.supported=false (macOS / Windows
// builds ship without the CGO + Linux dataplane). Splits the rendering out of
// the main page function to keep its complexity below the Biome ceiling.
function PlatformBanner({ reason, onSwitchToTestMaster }: PlatformBannerProps): ReactElement {
  const { t } = useTranslation();
  return (
    <Alert status="warning" className="flex-wrap">
      <div className="flex flex-1 flex-wrap items-center gap-default">
        <span className="flex-1 min-w-[16rem]">
          <strong className="font-semibold">
            {t('role.platform.bannerTitle', 'Reflector mode is not available on this platform.')}
          </strong>{' '}
          {t(
            'role.platform.bannerBody',
            'macOS and Windows builds use the pure-Go networking stack, which supports Test Master mode but not the line-rate Reflector dataplane. Use the Linux build to act as a Reflector node, or switch this stem to Test Master mode.',
          )}
          {reason ? <span className="ml-tight opacity-80">({reason})</span> : null}
        </span>
        <Button variant="outline" tone="violet" size="sm" onClick={onSwitchToTestMaster}>
          {t('role.platform.switchToTestMaster', 'Switch to Test Master')}
        </Button>
      </div>
    </Alert>
  );
}

interface RunningStatusProps {
  testStatus: Stats['testStatus'];
}

// Inline status badges (running / cancelled / error) rendered to the right
// of the start/stop controls. Extracted to keep ReflectorPage simple.
function RunningStatus({ testStatus }: RunningStatusProps): ReactElement | null {
  const { t } = useTranslation();
  const running = testStatus === 'running' || testStatus === 'starting';

  if (running) {
    return (
      <output className="status-badge success flex items-center gap-compact">
        <span className="w-2 h-2 rounded-full bg-status-success animate-pulse" aria-hidden="true" />
        {testStatus === 'starting'
          ? t('status.starting', 'Starting')
          : t('status.running', 'Running')}
      </output>
    );
  }
  if (testStatus === 'cancelled') {
    return <output className="status-badge warning">{t('status.stopped', 'Stopped')}</output>;
  }
  if (testStatus === 'error') {
    return (
      <output className="status-badge error" role="alert">
        {t('status.error', 'Error')}
      </output>
    );
  }
  return null;
}

export function ReflectorPage(): ReactElement {
  const { t } = useTranslation();
  const {
    interfaces,
    selectedInterface,
    setSelectedInterface,
    stats,
    reflectorProfile,
    setReflectorProfile,
    onStartReflector,
    onStopReflector,
    isStartingReflector,
    isStoppingReflector,
    reflectorStartError,
  } = useAppContext();
  const capabilities = useCapabilities();
  const { setRole } = useRole();

  const selectedIface = interfaces.find((i) => i.name === selectedInterface);
  const reflectorRunning = stats.testStatus === 'running' || stats.testStatus === 'starting';
  const { supported: reflectorSupported, reason: platformReasonRaw } = capabilities.reflector;
  const platformReason = platformReasonRaw ?? '';
  const unsupportedTooltip = t(
    'role.platform.startDisabledTooltip',
    'Reflector mode is not available on this platform. Use the Linux build to act as a Reflector node.',
  );

  const handleSwitchToTestMaster = (): void => {
    setRole('test_master');
  };

  return (
    <section className="stack-xl">
      <Breadcrumbs />
      <PageHeader
        icon={Repeat}
        title="Reflector"
        description="Loopback reflector — bounces frames back to the test master for end-to-end measurement."
        iconColorClass="text-module-reflector"
      />

      <RoleGuard requires="reflector">
        {!reflectorSupported ? (
          <PlatformBanner reason={platformReason} onSwitchToTestMaster={handleSwitchToTestMaster} />
        ) : null}

        {/* Control row: interface picker + start/stop + status */}
        <div className="flex flex-wrap items-start gap-default">
          <HeaderInterfaceSelector
            interfaces={interfaces}
            selectedInterface={selectedInterface}
            onSelectInterface={setSelectedInterface}
          />

          {reflectorRunning ? (
            <button
              type="button"
              onClick={onStopReflector}
              className="btn btn-secondary"
              disabled={isStoppingReflector}
              aria-busy={isStoppingReflector}
            >
              {isStoppingReflector ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                  {t('status.stopped', 'Stopping...')}
                </>
              ) : (
                <>
                  <Square className="w-4 h-4" aria-hidden="true" />
                  {t('buttons.stop', 'Stop')} Reflector
                </>
              )}
            </button>
          ) : (
            <button
              type="button"
              onClick={onStartReflector}
              className="btn btn-primary"
              disabled={!selectedInterface || isStartingReflector || !reflectorSupported}
              aria-busy={isStartingReflector}
              aria-disabled={!reflectorSupported}
              title={!reflectorSupported ? unsupportedTooltip : undefined}
              data-testid="reflector-start-button"
            >
              {isStartingReflector ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                  {t('status.starting', 'Starting...')}
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" aria-hidden="true" />
                  {t('buttons.start', 'Start')} Reflector
                </>
              )}
            </button>
          )}

          {reflectorStartError ? (
            <div
              className="text-sm text-status-error flex items-center gap-compact"
              role="alert"
              aria-live="assertive"
            >
              <AlertTriangle className="w-4 h-4" aria-hidden="true" />
              {reflectorStartError}
            </div>
          ) : null}

          <div
            className="flex items-center gap-default ml-auto"
            aria-live="polite"
            aria-atomic="true"
          >
            <RunningStatus testStatus={stats.testStatus} />
          </div>
        </div>

        {/* Stats grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-comfortable">
          <StatsCard
            icon={<Activity className="w-4 h-4" />}
            title="Packets Received"
            value={formatNumber(stats.packetsReceived)}
            subvalue={`${formatNumber(stats.bytesReceived)} bytes`}
          />
          <StatsCard
            icon={<Activity className="w-4 h-4" />}
            title="Packets Sent"
            value={formatNumber(stats.packetsSent)}
            subvalue={`${formatNumber(stats.bytesSent)} bytes`}
          />
          <StatsCard
            icon={<Gauge className="w-4 h-4" />}
            title="Current Rate"
            value={`${formatNumber(stats.currentPps)} pps`}
            subvalue={`${stats.currentMbps.toFixed(2)} Mbps`}
          />
          <div className="card">
            <div className="card-header">
              <Clock className="w-4 h-4" />
              Uptime
            </div>
            <div className="card-value font-mono">{formatUptime(stats.uptime)}</div>
            <div className="card-subvalue">
              Status:{' '}
              <span className={getStatusClassName(stats.testStatus)}>{stats.testStatus}</span>
            </div>
          </div>
        </div>

        {/* Interface details */}
        {selectedIface ? <InterfaceDetails iface={selectedIface} /> : null}

        {/* Reflector profile picker (moved out of Settings drawer) */}
        <ReflectorSection profile={reflectorProfile} onProfileChange={setReflectorProfile} />
      </RoleGuard>
    </section>
  );
}

export default ReflectorPage;

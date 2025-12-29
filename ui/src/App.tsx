/**
 * @fileoverview The Stem - Main Application Component
 * @description The primary React component that renders the test suite dashboard.
 *              Handles connection state, test execution, and real-time statistics display.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import {
  Activity,
  AlertTriangle,
  Clock,
  Gauge,
  HelpCircle,
  Moon,
  Play,
  RefreshCw,
  Settings,
  Square,
  Sun,
  Wifi,
  WifiOff,
} from 'lucide-react';
import type { ReactElement } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { HelpDrawer } from './components/HelpDrawer';
import { SettingsDrawer } from './components/SettingsDrawer';

interface Stats {
  packetsReceived: number;
  packetsSent: number;
  bytesReceived: number;
  bytesSent: number;
  currentPps: number;
  currentMbps: number;
  uptime: number;
  testStatus: 'idle' | 'running' | 'completed' | 'error';
  currentTest: string | null;
}

interface InterfaceInfo {
  name: string;
  mac: string;
  speed: number;
  duplex: string;
  state: string;
  driver: string;
  physical: boolean;
  xdp: boolean;
  dpdk: boolean;
  score: number;
}

const initialStats: Stats = {
  packetsReceived: 0,
  packetsSent: 0,
  bytesReceived: 0,
  bytesSent: 0,
  currentPps: 0,
  currentMbps: 0,
  uptime: 0,
  testStatus: 'idle',
  currentTest: null,
};

function formatNumber(num: number): string {
  if (num >= 1e9) return `${(num / 1e9).toFixed(2)}B`;
  if (num >= 1e6) return `${(num / 1e6).toFixed(2)}M`;
  if (num >= 1e3) return `${(num / 1e3).toFixed(2)}K`;
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
      return 'text-[var(--color-status-success)]';
    case 'error':
      return 'text-[var(--color-status-error)]';
    case 'completed':
      return 'text-[var(--color-status-info)]';
    default:
      return 'text-[var(--color-text-muted)]';
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
  const stateClassName =
    iface.state === 'up'
      ? 'text-[var(--color-status-success)]'
      : 'text-[var(--color-status-error)]';

  return (
    <div className="card mb-6">
      <div className="card-header">
        <Wifi className="w-4 h-4" />
        Interface Details
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <div className="text-[var(--color-text-muted)]">Name</div>
          <div className="font-medium">{iface.name}</div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">MAC</div>
          <div className="font-mono">{iface.mac}</div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">Speed</div>
          <div>
            {iface.speed} Mbps / {iface.duplex}
          </div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">Driver</div>
          <div>{iface.driver}</div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">State</div>
          <div className={stateClassName}>{iface.state}</div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">XDP Support</div>
          <div>{iface.xdp ? 'Yes' : 'No'}</div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">DPDK Support</div>
          <div>{iface.dpdk ? 'Yes' : 'No'}</div>
        </div>
        <div>
          <div className="text-[var(--color-text-muted)]">Score</div>
          <div>{iface.score}</div>
        </div>
      </div>
    </div>
  );
}

interface TestResultsProps {
  testStatus: Stats['testStatus'];
}

function TestResults({ testStatus }: TestResultsProps): ReactElement {
  let message: string;
  switch (testStatus) {
    case 'idle':
      message = 'No tests running. Configure tests in Settings and click Start.';
      break;
    case 'running':
      message = 'Test in progress... Results will appear here when complete.';
      break;
    case 'completed':
      message = 'Test completed. Detailed results coming soon.';
      break;
    default:
      message = 'An error occurred during the test.';
  }

  return (
    <div className="card">
      <div className="card-header">
        <AlertTriangle className="w-4 h-4" />
        Test Results
      </div>
      <div className="text-center py-12 text-[var(--color-text-muted)]">
        <p>{message}</p>
      </div>
    </div>
  );
}

function App(): ReactElement {
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [helpOpen, setHelpOpen] = useState(false);
  const [darkMode, setDarkMode] = useState(() => {
    if (typeof window !== 'undefined') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return false;
  });
  const [connected, setConnected] = useState(false);
  const [stats, setStats] = useState<Stats>(initialStats);
  const [interfaces, setInterfaces] = useState<InterfaceInfo[]>([]);
  const [selectedInterface, setSelectedInterface] = useState<string>('');
  const [mode, setMode] = useState<'reflector' | 'test_master'>('test_master');
  const [selectedTests, setSelectedTests] = useState<string[]>([
    'rfc2544_throughput',
    'rfc2544_latency',
    'rfc2544_frame_loss',
    'rfc2544_back_to_back',
  ]);
  const [reflectorProfile, setReflectorProfile] = useState<string>('all');

  const fetchInterfaces = useCallback(async () => {
    try {
      const response = await fetch('/api/interfaces');
      if (response.ok) {
        const data = await response.json();
        setInterfaces(data);
        // Auto-select highest scored interface
        if (data.length > 0) {
          setSelectedInterface((prev) => {
            if (prev) return prev;
            const best = data.reduce((a: InterfaceInfo, b: InterfaceInfo) =>
              a.score > b.score ? a : b,
            );
            return best.name;
          });
        }
        setConnected(true);
      }
    } catch (_error) {
      setConnected(false);
    }
  }, []);

  // Toggle dark mode
  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  // Fetch interfaces on mount
  useEffect(() => {
    fetchInterfaces();
  }, [fetchInterfaces]);

  // Poll stats when connected
  useEffect(() => {
    if (!connected) return;

    const interval = setInterval(async () => {
      try {
        const response = await fetch('/api/stats');
        if (response.ok) {
          const data = await response.json();
          setStats(data);
        }
      } catch (_error) {
        // Silently ignore polling errors
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [connected]);

  const handleStartTest = async (): Promise<void> => {
    try {
      await fetch('/api/test/start', { method: 'POST' });
    } catch (_error) {
      // Silently ignore test start errors
    }
  };

  const handleStopTest = async (): Promise<void> => {
    try {
      await fetch('/api/test/stop', { method: 'POST' });
    } catch (_error) {
      // Silently ignore test stop errors
    }
  };

  const toggleDarkMode = (): void => {
    setDarkMode(!darkMode);
  };

  const openHelp = (): void => {
    setHelpOpen(true);
  };

  const openSettings = (): void => {
    setSettingsOpen(true);
  };

  const selectedIface = interfaces.find((i) => i.name === selectedInterface);

  return (
    <div className="min-h-screen bg-[var(--color-surface-base)]">
      {/* Header */}
      <header className="bg-[var(--color-surface-raised)] border-b border-[var(--color-surface-border)] px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-[var(--color-stem-green)] flex items-center justify-center">
                <Activity className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">The Stem</h1>
                <p className="text-xs text-[var(--color-text-muted)]">Mustard Seed Networks</p>
              </div>
            </div>

            {/* Connection Status */}
            <div className={`status-badge ${connected ? 'success' : 'error'}`}>
              {connected ? (
                <>
                  <Wifi className="w-3 h-3" /> Connected
                </>
              ) : (
                <>
                  <WifiOff className="w-3 h-3" /> Disconnected
                </>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Theme Toggle */}
            <button
              type="button"
              onClick={toggleDarkMode}
              className="btn btn-ghost"
              title={darkMode ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {darkMode ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
            </button>

            {/* Refresh */}
            <button
              type="button"
              onClick={fetchInterfaces}
              className="btn btn-ghost"
              title="Refresh interfaces"
            >
              <RefreshCw className="w-5 h-5" />
            </button>

            {/* Help */}
            <button
              type="button"
              onClick={openHelp}
              className="btn btn-ghost"
              title="Help & Documentation"
            >
              <HelpCircle className="w-5 h-5" />
            </button>

            {/* Settings */}
            <button type="button" onClick={openSettings} className="btn btn-secondary">
              <Settings className="w-4 h-4" />
              Settings
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="p-6">
        {/* Test Controls */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-4">
            <select
              value={selectedInterface}
              onChange={(e) => setSelectedInterface(e.target.value)}
              className="w-48"
            >
              <option value="">Select Interface</option>
              {interfaces.map((iface) => (
                <option key={iface.name} value={iface.name}>
                  {iface.name} ({iface.speed}Mbps)
                </option>
              ))}
            </select>

            {stats.testStatus === 'running' ? (
              <button type="button" onClick={handleStopTest} className="btn btn-secondary">
                <Square className="w-4 h-4" />
                Stop Test
              </button>
            ) : (
              <button
                type="button"
                onClick={handleStartTest}
                className="btn btn-primary"
                disabled={!selectedInterface}
              >
                <Play className="w-4 h-4" />
                Start Test
              </button>
            )}
          </div>

          {stats.currentTest && (
            <div className="status-badge warning">Running: {stats.currentTest}</div>
          )}
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
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

        {/* Interface Details */}
        {selectedIface && <InterfaceDetails iface={selectedIface} />}

        {/* Results Area */}
        <TestResults testStatus={stats.testStatus} />
      </main>

      {/* Settings Drawer */}
      <SettingsDrawer
        isOpen={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        mode={mode}
        setMode={setMode}
        interfaces={interfaces}
        selectedInterface={selectedInterface}
        setSelectedInterface={setSelectedInterface}
        selectedTests={selectedTests}
        setSelectedTests={setSelectedTests}
        reflectorProfile={reflectorProfile}
        setReflectorProfile={setReflectorProfile}
      />

      {/* Help Drawer */}
      <HelpDrawer isOpen={helpOpen} onClose={() => setHelpOpen(false)} />
    </div>
  );
}

export default App;

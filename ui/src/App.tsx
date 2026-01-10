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
  History,
  Lock,
  LogOut,
  Moon,
  Play,
  RefreshCw,
  Settings,
  Square,
  Sun,
  Wifi,
  WifiOff,
} from 'lucide-react';
import type { FormEvent, ReactElement } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { HelpDrawer } from './components/HelpDrawer';
import { ModuleCard } from './components/ModuleCard';
import { ResultHistory } from './components/ResultHistory';
import { defaultRFC2544Config, type RFC2544Config } from './components/RFC2544ConfigForm';
import { defaultRFC2889Config, type RFC2889Config } from './components/RFC2889ConfigForm';
import { defaultRFC6349Config, type RFC6349Config } from './components/RFC6349ConfigForm';
import { SettingsDrawer } from './components/SettingsDrawer';
import { TestProgressBar, useTestProgress } from './components/TestProgressBar';
import { defaultTrafficGenConfig, type TrafficGenConfig } from './components/TrafficGenConfigForm';
import { defaultTSNConfig, type TSNConfig } from './components/TSNConfigForm';
import { defaultY1564Config, type Y1564Config } from './components/Y1564ConfigForm';
import { defaultY1731Config, type Y1731Config } from './components/Y1731ConfigForm';
import { ModuleSettingsProvider, useModuleSettings } from './context/ModuleSettingsContext';
import {
  type InterfaceInfo,
  initialStats,
  isValidAuthResponse,
  isValidInterfaceArray,
  isValidStats,
  type Stats,
  type TestEventPayload,
  type TestResult,
} from './types/api';

const TOKEN_STORAGE_KEY = 'stem-token';
const REFRESH_TOKEN_KEY = 'stem-refresh-token';
const WS_BASE_URL = import.meta.env.VITE_WS_URL || '';

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
    case 'starting':
      return 'text-[var(--color-status-info)]';
    case 'cancelled':
      return 'text-[var(--color-status-warning)]';
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
  result: TestResult | null;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = ((ms % 60000) / 1000).toFixed(0);
  return `${minutes}m ${seconds}s`;
}

function TestResults({ testStatus, result }: TestResultsProps): ReactElement {
  // Show placeholder messages when no result data
  if (!result) {
    let message: string;
    switch (testStatus) {
      case 'idle':
        message = 'No tests running. Configure tests in Settings and click Start.';
        break;
      case 'starting':
        message = 'Test is starting. Results will stream in shortly.';
        break;
      case 'running':
        message = 'Test in progress... Results will appear here when complete.';
        break;
      case 'cancelled':
        message = 'Test cancelled. Adjust settings or restart when ready.';
        break;
      case 'error':
        message = 'An error occurred during the test.';
        break;
      default:
        message = 'Waiting for the backend to report a status.';
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

  // Show actual test results
  const statusColor = result.success
    ? 'text-[var(--color-status-success)]'
    : 'text-[var(--color-status-error)]';

  return (
    <div className="card">
      <div className="card-header">
        <Activity className="w-4 h-4" />
        Test Results
      </div>

      {/* Test Header */}
      <div className="flex items-center justify-between mb-4 pb-4 border-b border-[var(--color-surface-border)]">
        <div>
          <div className="text-lg font-semibold text-[var(--color-text-primary)]">
            {result.testType}
          </div>
          <div className="text-sm text-[var(--color-text-muted)]">Module: {result.module}</div>
        </div>
        <div className="text-right">
          <div className={`text-lg font-semibold ${statusColor}`}>
            {result.success ? 'PASSED' : 'FAILED'}
          </div>
          {result.duration !== undefined && (
            <div className="text-sm text-[var(--color-text-muted)]">
              Duration: {formatDuration(result.duration)}
            </div>
          )}
        </div>
      </div>

      {/* Error Message */}
      {result.error && (
        <div className="mb-4 p-3 rounded-lg bg-[var(--color-status-error)]/10 border border-[var(--color-status-error)]/20">
          <div className="text-sm font-medium text-[var(--color-status-error)]">Error</div>
          <div className="text-sm text-[var(--color-text-primary)]">{result.error}</div>
        </div>
      )}

      {/* Metrics Grid */}
      {result.metrics && Object.keys(result.metrics).length > 0 && (
        <div className="mb-4">
          <div className="text-sm font-semibold text-[var(--color-text-muted)] mb-2">Metrics</div>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {Object.entries(result.metrics).map(([key, value]) => (
              <div
                key={key}
                className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]"
              >
                <div className="text-xs text-[var(--color-text-muted)] capitalize">
                  {key.replace(/_/g, ' ')}
                </div>
                <div className="text-lg font-semibold text-[var(--color-text-primary)]">
                  {typeof value === 'number' ? formatNumber(value) : String(value)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Timestamps */}
      <div className="text-xs text-[var(--color-text-muted)] flex gap-4">
        {result.startedAt && <span>Started: {new Date(result.startedAt).toLocaleString()}</span>}
        {result.completedAt && (
          <span>Completed: {new Date(result.completedAt).toLocaleString()}</span>
        )}
      </div>
    </div>
  );
}

function normalizeTestStatus(status?: string): Stats['testStatus'] {
  switch (status) {
    case 'starting':
      return 'starting';
    case 'running':
      return 'running';
    case 'completed':
      return 'completed';
    case 'cancelled':
      return 'cancelled';
    case 'error':
      return 'error';
    default:
      return 'idle';
  }
}

function mapStatsPayload(payload: Partial<Stats>): Stats {
  return {
    packetsReceived: Number(payload.packetsReceived ?? 0),
    packetsSent: Number(payload.packetsSent ?? 0),
    bytesReceived: Number(payload.bytesReceived ?? 0),
    bytesSent: Number(payload.bytesSent ?? 0),
    currentPps: Number(payload.currentPps ?? 0),
    currentMbps: Number(payload.currentMbps ?? 0),
    uptime: Number(payload.uptime ?? 0),
    testStatus: normalizeTestStatus(payload.testStatus),
    currentTest: payload.currentTest ?? null,
  };
}

// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Main app component with many states
function AppContent(): ReactElement {
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [helpOpen, setHelpOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [darkMode, setDarkMode] = useState(() => {
    if (typeof window !== 'undefined') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return false;
  });
  const [token, setToken] = useState<string | null>(() => {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem(TOKEN_STORAGE_KEY);
    }
    return null;
  });
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);
  const [isStoppingTest, setIsStoppingTest] = useState(false);
  const [testStartError, setTestStartError] = useState<string | null>(null);
  const statsIntervalRef = useRef<number | null>(null);
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('admin');
  const [connected, setConnected] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      return Boolean(window.localStorage.getItem(TOKEN_STORAGE_KEY));
    }
    return false;
  });
  const [stats, setStats] = useState<Stats>(initialStats);
  const [testResult, setTestResult] = useState<TestResult | null>(null);
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
  const [isStartingTest, setIsStartingTest] = useState(false);
  const [rfc2544Config, setRFC2544Config] = useState<RFC2544Config>(defaultRFC2544Config);
  const [rfc2889Config, setRFC2889Config] = useState<RFC2889Config>(defaultRFC2889Config);
  const [rfc6349Config, setRFC6349Config] = useState<RFC6349Config>(defaultRFC6349Config);
  const [y1564Config, setY1564Config] = useState<Y1564Config>(defaultY1564Config);
  const [y1731Config, setY1731Config] = useState<Y1731Config>(defaultY1731Config);
  const [tsnConfig, setTSNConfig] = useState<TSNConfig>(defaultTSNConfig);
  const [trafficGenConfig, setTrafficGenConfig] =
    useState<TrafficGenConfig>(defaultTrafficGenConfig);

  // Module settings context
  const {
    modules,
    moduleStatuses,
    moduleResults,
    toggleModule,
    toggleAutoStart,
    toggleTest,
    setModuleStatus,
    setModuleResults,
    clearModuleResults,
  } = useModuleSettings();

  // Calculate expected test duration based on config
  const expectedDuration =
    (rfc2544Config.duration + rfc2544Config.warmup) *
    rfc2544Config.trials *
    rfc2544Config.frameSizes.length *
    selectedTests.filter((t) => t.startsWith('rfc2544')).length;

  // Track test progress
  const testProgress = useTestProgress(stats.testStatus, stats.currentTest, expectedDuration);

  const expireSession = useCallback((message = 'Session expired. Please sign in again.') => {
    // Clear polling interval to prevent continued API calls
    if (statsIntervalRef.current !== null) {
      clearInterval(statsIntervalRef.current);
      statsIntervalRef.current = null;
    }
    setToken(null);
    setConnected(false);
    setLoginError(message);
    // Clear refresh token from storage
    window.localStorage.removeItem(REFRESH_TOKEN_KEY);
  }, []);

  // Token refresh function - attempts to get a new access token using refresh token
  const refreshAccessToken = useCallback(async (): Promise<boolean> => {
    const refreshToken = window.localStorage.getItem(REFRESH_TOKEN_KEY);
    if (!refreshToken) {
      return false;
    }

    try {
      const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
      });

      if (!response.ok) {
        return false;
      }

      const data: unknown = await response.json();
      if (!isValidAuthResponse(data)) {
        return false;
      }

      setToken(data.token);
      if (data.refreshToken) {
        window.localStorage.setItem(REFRESH_TOKEN_KEY, data.refreshToken);
      }
      return true;
    } catch {
      return false;
    }
  }, []);

  const authFetch = useCallback(
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Auth logic with retry requires branching
    async (input: RequestInfo, init: RequestInit = {}) => {
      if (!token) {
        throw new Error('Missing authentication token');
      }
      const headers = new Headers(init.headers || {});
      if (!headers.has('Authorization')) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json');
      }
      let response = await fetch(input, { ...init, headers });

      // On 401, attempt token refresh before expiring session
      if (response.status === 401) {
        const refreshed = await refreshAccessToken();
        if (refreshed) {
          // Retry request with new token
          const newToken = window.localStorage.getItem(TOKEN_STORAGE_KEY);
          if (newToken) {
            headers.set('Authorization', `Bearer ${newToken}`);
            response = await fetch(input, { ...init, headers });
            if (response.ok) {
              return response;
            }
          }
        }
        expireSession();
        throw new Error('Unauthorized');
      }

      if (response.status === 403) {
        expireSession('Access forbidden. Please sign in again.');
        throw new Error('Forbidden');
      }
      return response;
    },
    [expireSession, refreshAccessToken, token],
  );

  const fetchInterfaces = useCallback(
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Interface fetch with validation
    async () => {
      if (!token) {
        return;
      }
      try {
        const response = await authFetch('/api/v1/interfaces');
        if (!response.ok) {
          throw new Error('Failed to load interfaces');
        }
        const data: unknown = await response.json();
        if (!isValidInterfaceArray(data)) {
          throw new Error('Invalid interface data received from server');
        }
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
      } catch (error) {
        const err = error instanceof Error ? error : new Error('Unknown error');
        if (err.message === 'Unauthorized') {
          return;
        }
        setConnected(false);
      }
    },
    [authFetch, token],
  );

  const fetchStats = useCallback(async () => {
    try {
      const response = await authFetch('/api/v1/stats');
      if (!response.ok) {
        throw new Error('Failed to refresh stats');
      }
      const data: unknown = await response.json();
      if (!isValidStats(data)) {
        throw new Error('Invalid stats data received from server');
      }
      setStats(mapStatsPayload(data as Partial<Stats>));
    } catch (error) {
      if ((error as Error).message === 'Unauthorized') {
        return;
      }
      // Silently ignore other polling failures
    }
  }, [authFetch]);

  const handleLogin = useCallback(
    async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setLoginLoading(true);
      setLoginError(null);
      try {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password }),
        });
        if (!response.ok) {
          const text = (await response.text()) || 'Authentication failed';
          setLoginError(text);
          setConnected(false);
          return;
        }
        const data: unknown = await response.json();
        if (!isValidAuthResponse(data)) {
          setLoginError('Authentication failed');
          setConnected(false);
          return;
        }
        setToken(data.token);
        // Store refresh token for token refresh functionality
        if (data.refreshToken) {
          window.localStorage.setItem(REFRESH_TOKEN_KEY, data.refreshToken);
        }
        setLoginError(null);
        setConnected(true);
      } catch (_error) {
        setLoginError('Unable to reach authentication server.');
        setConnected(false);
      } finally {
        setLoginLoading(false);
      }
    },
    [password, username],
  );

  const handleStartTest = useCallback(
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Test config selection requires many branches
    async (): Promise<void> => {
      if (!token) return;
      setIsStartingTest(true);
      setTestStartError(null);

      try {
        // Determine test type based on mode
        const testType =
          mode === 'reflector' ? 'reflect' : (selectedTests[0] ?? 'rfc2544_throughput');

        // Build test configuration based on test type
        let config: Record<string, unknown> | undefined;
        if (testType.startsWith('rfc2544')) {
          config = { rfc2544: rfc2544Config };
        } else if (testType.startsWith('rfc2889')) {
          config = { rfc2889: rfc2889Config };
        } else if (testType.startsWith('rfc6349')) {
          config = { rfc6349: rfc6349Config };
        } else if (testType.startsWith('y1564')) {
          config = { y1564: y1564Config };
        } else if (testType.startsWith('y1731')) {
          config = { y1731: y1731Config };
        } else if (testType.startsWith('tsn')) {
          config = { tsn: tsnConfig };
        } else if (testType === 'custom_stream') {
          config = { trafficGen: trafficGenConfig };
        }

        const response = await authFetch('/api/v1/test/start', {
          method: 'POST',
          body: JSON.stringify({
            interface: selectedInterface,
            testType,
            mode,
            profile: mode === 'reflector' ? reflectorProfile : undefined,
            tests: selectedTests,
            config,
          }),
        });

        // Check for validation errors in response
        if (!response.ok) {
          const errorData = await response.json().catch(() => null);
          const errorMessage = (errorData as { error?: string })?.error || 'Failed to start test';
          setTestStartError(errorMessage);
          return;
        }

        // Status updates will come from WebSocket - don't update optimistically
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to start test';
        setTestStartError(message);
      } finally {
        setIsStartingTest(false);
      }
    },
    [
      authFetch,
      mode,
      reflectorProfile,
      rfc2544Config,
      rfc2889Config,
      rfc6349Config,
      selectedInterface,
      selectedTests,
      token,
      trafficGenConfig,
      tsnConfig,
      y1564Config,
      y1731Config,
    ],
  );

  const handleStopTest = useCallback(async (): Promise<void> => {
    if (!token) return;
    setIsStoppingTest(true);
    try {
      await authFetch('/api/v1/test/stop', { method: 'POST' });
      // Status update will come from WebSocket
    } catch (_error) {
      // Silently ignore test stop errors
    } finally {
      setIsStoppingTest(false);
    }
  }, [authFetch, token]);

  // Module-specific start handler
  const handleModuleStart = useCallback(
    async (moduleName: string): Promise<void> => {
      if (!token || !selectedInterface) return;

      // Get enabled tests for this module
      const moduleConfig = modules.find((m) => m.name === moduleName);
      if (!moduleConfig) return;

      const enabledTests = moduleConfig.tests.filter((t) => t.enabled);
      if (enabledTests.length === 0) return;

      const testType = enabledTests[0].id;

      // Clear previous results and initialize new results structure
      clearModuleResults(moduleName);

      // Initialize frame size results for RFC 2544 style tests
      if (moduleName === 'benchmark' || moduleName === 'certify') {
        const frameSizes = rfc2544Config.frameSizes;
        setModuleResults(moduleName, {
          testType,
          startedAt: new Date().toISOString(),
          frameSizeResults: frameSizes.map((size) => ({
            frameSize: size,
            status: 'pending' as const,
          })),
        });
      }

      // Update module status
      setModuleStatus(moduleName, { status: 'starting', currentTest: testType });

      try {
        await authFetch('/api/v1/test/start', {
          method: 'POST',
          body: JSON.stringify({
            interface: selectedInterface,
            testType,
            module: moduleName,
            tests: enabledTests.map((t) => t.id),
          }),
        });

        setModuleStatus(moduleName, { status: 'running', currentTest: testType });
        setStats((prev) => ({
          ...prev,
          testStatus: 'running',
          currentTest: testType,
        }));
      } catch (_error) {
        setModuleStatus(moduleName, { status: 'error', currentTest: null });
      }
    },
    [
      authFetch,
      clearModuleResults,
      modules,
      rfc2544Config.frameSizes,
      selectedInterface,
      setModuleResults,
      setModuleStatus,
      token,
    ],
  );

  // Module-specific stop handler
  const handleModuleStop = useCallback(
    async (moduleName: string): Promise<void> => {
      if (!token) return;

      try {
        await authFetch('/api/v1/test/stop', { method: 'POST' });
        setModuleStatus(moduleName, { status: 'cancelled', currentTest: null });
      } catch (_error) {
        // Silently ignore
      }
    },
    [authFetch, setModuleStatus, token],
  );

  // Open settings drawer for specific module
  const handleModuleConfigure = useCallback((_moduleName: string): void => {
    // TODO: Pre-select module section in settings drawer
    setSettingsOpen(true);
  }, []);

  // Logout handler - clears token and resets state
  const handleLogout = useCallback(() => {
    // Clear token from state and localStorage
    setToken(null);
    setConnected(false);
    setLoginError(null);
    // Reset stats
    setStats({
      packetsReceived: 0,
      packetsSent: 0,
      bytesReceived: 0,
      bytesSent: 0,
      currentPps: 0,
      currentMbps: 0,
      uptime: 0,
      testStatus: 'idle',
      currentTest: null,
    });
  }, []);

  const handleWsMessage = useCallback((event: MessageEvent<string>) => {
    try {
      const payload = JSON.parse(event.data) as TestEventPayload;
      const normalizedStatus = normalizeTestStatus(payload.status);
      setStats((prev) => ({
        ...prev,
        testStatus: normalizedStatus,
        currentTest: payload.testType ?? prev.currentTest,
      }));

      // Capture test result data when test completes
      if (normalizedStatus === 'completed' || normalizedStatus === 'error') {
        const result: TestResult = {
          testType: payload.testType ?? 'unknown',
          module: payload.module ?? 'unknown',
          status: normalizedStatus,
          success: payload.success ?? normalizedStatus === 'completed',
          error: payload.error,
          ...(payload.data ?? {}),
        };
        setTestResult(result);
      } else if (normalizedStatus === 'starting') {
        // Clear previous results when starting a new test
        setTestResult(null);
      }
    } catch (_error) {
      // Ignore malformed websocket messages
    }
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (token) {
      window.localStorage.setItem(TOKEN_STORAGE_KEY, token);
    } else {
      window.localStorage.removeItem(TOKEN_STORAGE_KEY);
    }
  }, [token]);

  // Toggle dark mode
  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  // Handle mode changes - update selected tests accordingly
  useEffect(() => {
    if (mode === 'reflector') {
      // In reflector mode, always use 'reflect' test type
      setSelectedTests(['reflect']);
    } else if (mode === 'test_master') {
      // When switching back to test_master, restore default tests if empty
      setSelectedTests((prev) => {
        if (prev.length === 0 || (prev.length === 1 && prev[0] === 'reflect')) {
          return [
            'rfc2544_throughput',
            'rfc2544_latency',
            'rfc2544_frame_loss',
            'rfc2544_back_to_back',
          ];
        }
        return prev;
      });
    }
  }, [mode]);

  // Fetch interfaces on mount
  useEffect(() => {
    fetchInterfaces();
  }, [fetchInterfaces]);

  // Poll stats when connected - uses ref for proper cleanup on session expire
  useEffect(() => {
    if (!connected) {
      // Clear any existing interval when disconnected
      if (statsIntervalRef.current !== null) {
        clearInterval(statsIntervalRef.current);
        statsIntervalRef.current = null;
      }
      return;
    }

    statsIntervalRef.current = window.setInterval(() => {
      void fetchStats();
    }, 1000);
    void fetchStats();

    return () => {
      if (statsIntervalRef.current !== null) {
        clearInterval(statsIntervalRef.current);
        statsIntervalRef.current = null;
      }
    };
  }, [connected, fetchStats]);

  useEffect(() => {
    if (!token) return;

    let ws: WebSocket | null = null;
    let reconnectTimer: number | undefined;

    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: WebSocket URL construction with fallbacks
    const connect = () => {
      let wsUrl: string;
      if (WS_BASE_URL) {
        // Use configured WebSocket URL (e.g., from VITE_WS_URL env var)
        wsUrl = `${WS_BASE_URL}/api/v1/ws/test-results?token=${token}`;
      } else {
        // Auto-detect based on current page location
        const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const hostname = window.location.hostname;
        const port = window.location.port === '3000' ? '8080' : window.location.port;
        const portSegment = port ? `:${port}` : '';
        wsUrl = `${protocol}://${hostname}${portSegment}/api/v1/ws/test-results?token=${token}`;
      }
      ws = new WebSocket(wsUrl);
      ws.onmessage = handleWsMessage;
      ws.onclose = () => {
        reconnectTimer = window.setTimeout(connect, 3000);
      };
      ws.onerror = () => {
        ws?.close();
      };
    };

    connect();

    return () => {
      if (reconnectTimer) {
        window.clearTimeout(reconnectTimer);
      }
      ws?.close();
    };
  }, [handleWsMessage, token]);

  const toggleDarkMode = (): void => {
    setDarkMode(!darkMode);
  };

  const openHelp = (): void => {
    setHelpOpen(true);
  };

  const openSettings = (): void => {
    setSettingsOpen(true);
  };

  const openHistory = (): void => {
    setHistoryOpen(true);
  };

  const selectedIface = interfaces.find((i) => i.name === selectedInterface);

  const iconButtonClass =
    'p-2 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-primary focus-visible:ring-offset-1 focus-visible:ring-offset-surface-raised';

  return (
    <div className="min-h-screen bg-[var(--color-surface-base)]">
      {/* Header */}
      <header className="border-b border-surface-border bg-surface-raised">
        <div className="mx-auto max-w-7xl px-4 py-3 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-primary text-text-inverse">
                  <Activity className="h-5 w-5" />
                </div>
                <div>
                  <h1 className="heading-4 text-text-primary">The Stem</h1>
                  <p className="caption text-text-muted">Mustard Seed Networks</p>
                </div>
              </div>

              {/* Connection Status */}
              <div className={`status-badge ${connected ? 'success' : 'error'}`}>
                {connected ? (
                  <>
                    <Wifi className="h-3 w-3" /> Connected
                  </>
                ) : (
                  <>
                    <WifiOff className="h-3 w-3" /> Disconnected
                  </>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              {/* Theme Toggle */}
              <button
                type="button"
                onClick={toggleDarkMode}
                className={iconButtonClass}
                title={darkMode ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {darkMode ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
              </button>

              {/* Refresh */}
              <button
                type="button"
                onClick={fetchInterfaces}
                className={iconButtonClass}
                title="Refresh interfaces"
              >
                <RefreshCw className="h-5 w-5" />
              </button>

              {/* History */}
              <button
                type="button"
                onClick={openHistory}
                className={iconButtonClass}
                title="Test History"
              >
                <History className="h-5 w-5" />
              </button>

              {/* Help */}
              <button
                type="button"
                onClick={openHelp}
                className={iconButtonClass}
                title="Help & Documentation"
              >
                <HelpCircle className="h-5 w-5" />
              </button>

              {/* Settings */}
              <button
                type="button"
                onClick={openSettings}
                className={iconButtonClass}
                title="Settings"
              >
                <Settings className="h-5 w-5" />
              </button>

              {/* Logout */}
              <button
                type="button"
                onClick={handleLogout}
                className={iconButtonClass}
                title="Logout"
              >
                <LogOut className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
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

            {stats.testStatus === 'running' || stats.testStatus === 'starting' ? (
              <button
                type="button"
                onClick={handleStopTest}
                className="btn btn-secondary"
                disabled={isStoppingTest}
              >
                {isStoppingTest ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    Stopping...
                  </>
                ) : (
                  <>
                    <Square className="w-4 h-4" />
                    Stop {mode === 'reflector' ? 'Reflector' : 'Test'}
                  </>
                )}
              </button>
            ) : (
              <button
                type="button"
                onClick={handleStartTest}
                className="btn btn-primary"
                disabled={!selectedInterface || isStartingTest}
              >
                {isStartingTest ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    Starting...
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4" />
                    Start {mode === 'reflector' ? 'Reflector' : 'Test'}
                  </>
                )}
              </button>
            )}

            {/* Test Start Error Display */}
            {testStartError && (
              <div className="text-sm text-[var(--color-status-error)] flex items-center gap-2">
                <AlertTriangle className="w-4 h-4" />
                {testStartError}
              </div>
            )}
          </div>

          {/* Test Status Indicator */}
          <div className="flex items-center gap-3">
            {(stats.testStatus === 'running' || stats.testStatus === 'starting') && (
              <div className="status-badge success flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-[var(--color-status-success)] animate-pulse" />
                {stats.testStatus === 'starting' ? 'Starting' : 'Running'}:{' '}
                {stats.currentTest || mode}
              </div>
            )}
            {stats.testStatus === 'completed' && (
              <div className="status-badge info">Completed: {stats.currentTest}</div>
            )}
            {stats.testStatus === 'error' && (
              <div className="status-badge error">Error: {stats.currentTest || 'Test failed'}</div>
            )}
            {stats.testStatus === 'cancelled' && (
              <div className="status-badge warning">Stopped: {stats.currentTest}</div>
            )}
          </div>
        </div>

        {/* Test Progress Bar */}
        <TestProgressBar progress={testProgress} />

        {/* Module Cards */}
        <div className="mb-6">
          <h2 className="section-title mb-4">Test Modules</h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            {modules.map((moduleConfig) => (
              <ModuleCard
                key={moduleConfig.name}
                config={moduleConfig}
                status={moduleStatuses[moduleConfig.name] ?? { status: 'idle', currentTest: null }}
                results={moduleResults[moduleConfig.name]}
                onToggleModule={(enabled) => toggleModule(moduleConfig.name, enabled)}
                onToggleAutoStart={(enabled) => toggleAutoStart(moduleConfig.name, enabled)}
                onToggleTest={(testId, enabled) => toggleTest(moduleConfig.name, testId, enabled)}
                onStart={() => handleModuleStart(moduleConfig.name)}
                onStop={() => handleModuleStop(moduleConfig.name)}
                onConfigure={() => handleModuleConfigure(moduleConfig.name)}
              />
            ))}
          </div>
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
        <TestResults testStatus={stats.testStatus} result={testResult} />
      </main>

      {/* Footer */}
      <footer className="mt-8">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="rounded-2xl border border-surface-border bg-surface-raised p-6">
            <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
              {/* Product Info */}
              <div>
                <h3 className="heading-4 text-text-primary mb-2">The Stem</h3>
                <p className="body-small text-text-muted mb-1">by Mustard Seed Networks</p>
                <p className="caption text-text-muted">Version 0.1.0</p>
              </div>

              {/* Contact */}
              <div>
                <h4 className="body-small font-medium text-text-primary mb-2">Contact</h4>
                <div className="space-y-1">
                  <a
                    href="mailto:support@mustardseednetworks.com"
                    className="body-small text-brand-primary hover:underline block"
                  >
                    support@mustardseednetworks.com
                  </a>
                  <a
                    href="tel:+18005551234"
                    className="body-small text-text-muted hover:text-text-primary block"
                  >
                    1-800-555-1234
                  </a>
                </div>
              </div>

              {/* Website */}
              <div>
                <h4 className="body-small font-medium text-text-primary mb-2">Website</h4>
                <a
                  href="https://mustardseednetworks.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="body-small text-brand-primary hover:underline"
                >
                  mustardseednetworks.com
                </a>
              </div>

              {/* Legal */}
              <div>
                <h4 className="body-small font-medium text-text-primary mb-2">Legal</h4>
                <div className="flex flex-wrap gap-x-3 gap-y-1">
                  <a href="/terms" className="body-small text-text-muted hover:text-brand-primary">
                    Terms of Service
                  </a>
                  <a
                    href="/privacy"
                    className="body-small text-text-muted hover:text-brand-primary"
                  >
                    Privacy Policy
                  </a>
                  <a
                    href="/license"
                    className="body-small text-text-muted hover:text-brand-primary"
                  >
                    License
                  </a>
                </div>
              </div>
            </div>

            {/* Copyright */}
            <div className="mt-6 pt-4 border-t border-surface-border text-center">
              <p className="caption text-text-muted">
                &copy; {new Date().getFullYear()} Mustard Seed Networks. All rights reserved.
              </p>
            </div>
          </div>
        </div>
      </footer>

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
        rfc2544Config={rfc2544Config}
        setRFC2544Config={setRFC2544Config}
        rfc2889Config={rfc2889Config}
        setRFC2889Config={setRFC2889Config}
        rfc6349Config={rfc6349Config}
        setRFC6349Config={setRFC6349Config}
        y1564Config={y1564Config}
        setY1564Config={setY1564Config}
        y1731Config={y1731Config}
        setY1731Config={setY1731Config}
        tsnConfig={tsnConfig}
        setTSNConfig={setTSNConfig}
        trafficGenConfig={trafficGenConfig}
        setTrafficGenConfig={setTrafficGenConfig}
      />

      {/* Help Drawer */}
      <HelpDrawer isOpen={helpOpen} onClose={() => setHelpOpen(false)} />

      {/* Result History Drawer */}
      <ResultHistory
        isOpen={historyOpen}
        onClose={() => setHistoryOpen(false)}
        currentResult={testResult}
      />
      {!token && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-3xl border border-[var(--color-surface-border)] bg-[var(--color-surface-raised)] p-6 shadow-2xl">
            <div className="flex items-center gap-2 text-lg font-semibold text-[var(--color-text-primary)]">
              <Lock className="w-5 h-5 text-[var(--color-brand-primary)]" />
              Sign in to continue
            </div>
            <p className="text-sm text-[var(--color-text-muted)]">
              Authenticate with your Stem credentials. Default dev account: admin / admin.
            </p>
            <form className="mt-6 space-y-4" onSubmit={handleLogin}>
              <div>
                <label
                  htmlFor="stem-login-username"
                  className="text-xs font-semibold text-[var(--color-text-muted)]"
                >
                  Username
                </label>
                <input
                  id="stem-login-username"
                  type="text"
                  autoComplete="username"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  className="mt-1 w-full rounded-xl border border-[var(--color-surface-border)] bg-[var(--color-surface-base)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30"
                />
              </div>
              <div>
                <label
                  htmlFor="stem-login-password"
                  className="text-xs font-semibold text-[var(--color-text-muted)]"
                >
                  Password
                </label>
                <input
                  id="stem-login-password"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  className="mt-1 w-full rounded-xl border border-[var(--color-surface-border)] bg-[var(--color-surface-base)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-brand-primary)]/30"
                />
              </div>
              {loginError && (
                <p className="text-xs text-[var(--color-status-error)]">{loginError}</p>
              )}
              <button
                type="submit"
                className="btn btn-primary w-full justify-center"
                disabled={loginLoading}
              >
                {loginLoading ? 'Signing in...' : 'Sign In'}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

// Wrapper component that provides context
function App(): ReactElement {
  return (
    <ModuleSettingsProvider>
      <AppContent />
    </ModuleSettingsProvider>
  );
}

export default App;

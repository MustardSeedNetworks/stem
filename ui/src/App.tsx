/**
 * @fileoverview The Stem - Main Application Component
 * @description The primary React component that renders the test suite dashboard.
 *              Handles connection state, test execution, and real-time statistics display.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { valibotResolver } from '@hookform/resolvers/valibot';
import {
  Activity,
  AlertTriangle,
  Lock,
  LogOut,
  Moon,
  Play,
  RefreshCw,
  Square,
  Sun,
  Wifi,
  WifiOff,
} from 'lucide-react';
import type { ReactElement } from 'react';
import { Suspense, useCallback, useEffect, useRef, useState } from 'react';
import { type SubmitHandler, useForm } from 'react-hook-form';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { HelpDrawer } from './components/HelpDrawer';
import { ResultHistory } from './components/ResultHistory';
import { defaultRFC2544Config, type RFC2544Config } from './components/RFC2544ConfigForm';
import { defaultRFC2889Config, type RFC2889Config } from './components/RFC2889ConfigForm';
import { defaultRFC6349Config, type RFC6349Config } from './components/RFC6349ConfigForm';
import { RoleChip } from './components/RoleChip';
import { RecoveryForm } from './components/recovery/RecoveryForm';
import { SettingsDrawer } from './components/SettingsDrawer';
import type { ReflectorProfile } from './components/settings/types';
import { SetupWizard } from './components/setup/SetupWizard';
import { TestProgressBar, useTestProgress } from './components/TestProgressBar';
import { defaultTrafficGenConfig, type TrafficGenConfig } from './components/TrafficGenConfigForm';
import { defaultTSNConfig, type TSNConfig } from './components/TSNConfigForm';
import { CommandPalette } from './components/ui/CommandPalette';
import { defaultY1564Config, type Y1564Config } from './components/Y1564ConfigForm';
import { defaultY1731Config, type Y1731Config } from './components/Y1731ConfigForm';
import { AppContext, type AppContextValue } from './contexts/AppContext';
import { ModuleSettingsProvider, useModuleSettings } from './contexts/ModuleSettingsContext';
import { RoleProvider, useRole } from './contexts/RoleContext';
import { useBuildVersion } from './hooks/useBuildVersion';
import { useFocusTrap } from './hooks/useFocusTrap';
import { useTheme } from './hooks/useTheme';
import { navGroups } from './navGroups';
import { pages } from './pageRegistry';
import { LoginSchema, MfaVerifySchema } from './schemas/auth';
import {
  type InterfaceInfo,
  initialStats,
  isValidAuthResponse,
  isValidInterfaceArray,
  isValidStats,
  type Stats,
  type TestResult,
} from './types/api';
import { PageLoader } from './ui/PageLoader';
import { SidebarLayout } from './ui/Sidebar';
import { logError, logWarn } from './utils/logger';

// Note: Tokens are now stored in httpOnly cookies set by the backend.
// localStorage is no longer used for token storage (security improvement).
// We only track if the user is authenticated via a simple boolean flag.
const AUTH_FLAG_KEY = 'stem-authenticated';

/** Setup status response from /api/v1/setup/status */
interface SetupStatus {
  needsSetup: boolean;
  username?: string;
  suggestedPassword?: string;
  setupToken?: string;
}

/** Recovery status response from /api/v1/recovery/status */
interface RecoveryStatus {
  active: boolean;
  remainingTime?: number;
  instructions?: string;
}

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

// Helper: check if test just completed (status transition to completed/error)
function isTestCompleted(prev: string, curr: string): boolean {
  return (curr === 'completed' || curr === 'error') && prev !== 'completed' && prev !== 'error';
}

// Helper: check if new test is starting
function isTestStarting(prev: string, curr: string): boolean {
  return curr === 'starting' && prev !== 'starting';
}

interface TestResultsProps {
  testStatus: Stats['testStatus'];
  result: TestResult | null;
}

function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
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
        <div className="text-center py-centered text-text-muted">
          <p>{message}</p>
        </div>
      </div>
    );
  }

  // Show actual test results
  const statusColor = result.success ? 'text-status-success' : 'text-status-error';

  return (
    <div className="card">
      <div className="card-header">
        <Activity className="w-4 h-4" />
        Test Results
      </div>

      {/* Test Header */}
      <div className="flex-between mb-content pb-4 border-b border-surface-border">
        <div>
          <div className="heading-3 text-text-primary">{result.testType}</div>
          <div className="text-sm text-text-muted">Module: {result.module}</div>
        </div>
        <div className="text-right">
          <div className={`heading-3 ${statusColor}`}>{result.success ? 'PASSED' : 'FAILED'}</div>
          {result.duration !== undefined && (
            <div className="text-sm text-text-muted">
              Duration: {formatDuration(result.duration)}
            </div>
          )}
        </div>
      </div>

      {/* Error Message */}
      {result.error ? (
        <div className="mb-content pad-sm rounded-lg bg-status-error/10 border border-status-error/20">
          <div className="text-sm font-medium text-status-error">Error</div>
          <div className="text-sm text-text-primary">{result.error}</div>
        </div>
      ) : null}

      {/* Metrics Grid */}
      {result.metrics && Object.keys(result.metrics).length > 0 && (
        <div className="mb-content">
          <div className="text-sm font-semibold text-text-muted mb-2">Metrics</div>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-default">
            {Object.entries(result.metrics).map(([key, value]) => (
              <div
                key={key}
                className="pad-sm rounded-lg bg-surface-base border border-surface-border"
              >
                <div className="text-xs text-text-muted capitalize">{key.replace(/_/g, ' ')}</div>
                <div className="heading-3 text-text-primary">
                  {typeof value === 'number' ? formatNumber(value) : String(value)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Timestamps */}
      <div className="text-xs text-text-muted flex gap-comfortable">
        {result.startedAt ? (
          <span>Started: {new Date(result.startedAt).toLocaleString()}</span>
        ) : null}
        {result.completedAt ? (
          <span>Completed: {new Date(result.completedAt).toLocaleString()}</span>
        ) : null}
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

/** Extract error message from response JSON, or return default */
async function extractResponseError(response: Response, defaultMessage: string): Promise<string> {
  try {
    const errorData = await (response.json() as Promise<{ error?: string }>);
    return errorData?.error || defaultMessage;
  } catch {
    return defaultMessage;
  }
}

/** Build test configuration based on test type prefix */
function buildTestConfig(
  testType: string,
  configs: {
    rfc2544: RFC2544Config;
    rfc2889: RFC2889Config;
    rfc6349: RFC6349Config;
    y1564: Y1564Config;
    y1731: Y1731Config;
    tsn: TSNConfig;
    trafficGen: TrafficGenConfig;
  },
): Record<string, unknown> | undefined {
  const prefixToConfig: Record<string, Record<string, unknown>> = {
    rfc2544: { rfc2544: configs.rfc2544 },
    rfc2889: { rfc2889: configs.rfc2889 },
    rfc6349: { rfc6349: configs.rfc6349 },
    y1564: { y1564: configs.y1564 },
    y1731: { y1731: configs.y1731 },
    tsn: { tsn: configs.tsn },
  };

  if (testType === 'custom_stream') {
    return { trafficGen: configs.trafficGen };
  }

  for (const [prefix, config] of Object.entries(prefixToConfig)) {
    if (testType.startsWith(prefix)) {
      return config;
    }
  }

  return;
}

// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: AppContent orchestrates auth, test state, dark mode, and the routed shell — the topBar JSX adds one branch over the existing 40 threshold; planned to extract into a TopBar component in the follow-up Phase A.1 commit.
function AppContent(): ReactElement {
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [helpOpen, setHelpOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const { isDark, toggleTheme } = useTheme();
  const buildVersion = useBuildVersion();
  // Track authentication state (tokens are in httpOnly cookies, inaccessible to JS)
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem(AUTH_FLAG_KEY) === 'true';
    }
    return false;
  });
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);
  // Wave 3 (#85): when login returns mfaRequired, we hold the
  // mfa_token and prompt for the second-factor code before flipping
  // isAuthenticated.
  const [mfaPending, setMfaPending] = useState<{ mfaToken: string; factor: string } | null>(null);

  // react-hook-form instances for login + MFA verify (#332).
  const loginForm = useForm<{ username: string; password: string }>({
    resolver: valibotResolver(LoginSchema),
    defaultValues: { username: '', password: '' },
    mode: 'onBlur',
  });
  const mfaForm = useForm<{ code: string }>({
    resolver: valibotResolver(MfaVerifySchema),
    defaultValues: { code: '' },
    mode: 'onBlur',
  });
  const [setupStatus, setSetupStatus] = useState<SetupStatus | null>(null);
  const [setupChecked, setSetupChecked] = useState(false);
  const [recoveryStatus, setRecoveryStatus] = useState<RecoveryStatus | null>(null);
  const [showRecoveryForm, setShowRecoveryForm] = useState(false);
  const [isStoppingTest, setIsStoppingTest] = useState(false);
  const [testStartError, setTestStartError] = useState<string | null>(null);
  const statsIntervalRef = useRef<number | null>(null);
  const [connected, setConnected] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem(AUTH_FLAG_KEY) === 'true';
    }
    return false;
  });
  const [stats, setStats] = useState<Stats>(initialStats);
  const [testResult, setTestResult] = useState<TestResult | null>(null);
  const [interfaces, setInterfaces] = useState<InterfaceInfo[]>([]);
  const [selectedInterface, setSelectedInterface] = useState<string>('');
  // The Stem instance role drives the legacy `mode` state. RoleContext
  // persists the choice to localStorage and is mutated by the header
  // RoleChip and per-page RoleGuard.
  const { role: mode } = useRole();
  const [selectedTests, setSelectedTests] = useState<string[]>([
    'rfc2544_throughput',
    'rfc2544_latency',
    'rfc2544_frame_loss',
    'rfc2544_back_to_back',
  ]);
  const [reflectorProfile, setReflectorProfile] = useState<ReflectorProfile>('all');
  const [isStartingTest, setIsStartingTest] = useState(false);
  const [rfc2544Config, setRFC2544Config] = useState<RFC2544Config>(defaultRFC2544Config);
  const [rfc2889Config, setRFC2889Config] = useState<RFC2889Config>(defaultRFC2889Config);
  const [rfc6349Config, setRFC6349Config] = useState<RFC6349Config>(defaultRFC6349Config);
  const [y1564Config, setY1564Config] = useState<Y1564Config>(defaultY1564Config);
  const [y1731Config, setY1731Config] = useState<Y1731Config>(defaultY1731Config);
  const [tsnConfig, setTSNConfig] = useState<TSNConfig>(defaultTSNConfig);
  const [trafficGenConfig, setTrafficGenConfig] =
    useState<TrafficGenConfig>(defaultTrafficGenConfig);

  // Focus trap for login modal (no onEscape - user must authenticate)
  const loginModalRef = useFocusTrap<HTMLDivElement>({
    isActive: !isAuthenticated,
    autoFocus: true,
    restoreFocus: false, // No element to restore to after login
  });

  // Module settings context — used by per-module test pages (consumed
  // via ModuleSettingsProvider mounted around App).
  useModuleSettings();

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
    setIsAuthenticated(false);
    setConnected(false);
    setLoginError(message);
    // Clear auth flag from storage (cookies cleared by server on logout)
    window.localStorage.removeItem(AUTH_FLAG_KEY);
  }, []);

  // Token refresh function - attempts to get a new access token using refresh cookie
  const refreshAccessToken = useCallback(async (): Promise<boolean> => {
    try {
      // Refresh token is in httpOnly cookie, sent automatically with credentials
      const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include', // Include cookies in request
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}), // Empty body, refresh token is in cookie
      });

      if (!response.ok) {
        return false;
      }

      // New access token is set in httpOnly cookie by server
      return true;
    } catch {
      return false;
    }
  }, []);

  // Helper: Make an authenticated request with headers
  const makeAuthRequest = useCallback(
    async (input: RequestInfo, init: RequestInit, headers: Headers): Promise<Response> =>
      fetch(input, { ...init, headers, credentials: 'include' }),
    [],
  );

  // Helper: Handle 401 unauthorized response with retry
  const handle401Retry = useCallback(
    async (input: RequestInfo, init: RequestInit, headers: Headers): Promise<Response | null> => {
      const refreshed = await refreshAccessToken();
      if (!refreshed) {
        return null;
      }
      const retryResponse = await makeAuthRequest(input, init, headers);
      return retryResponse.ok ? retryResponse : null;
    },
    [makeAuthRequest, refreshAccessToken],
  );

  const authFetch = useCallback(
    async (input: RequestInfo, init: RequestInit = {}) => {
      if (!isAuthenticated) {
        throw new Error('Not authenticated');
      }
      const headers = new Headers(init.headers || {});
      if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json');
      }

      const response = await makeAuthRequest(input, init, headers);

      if (response.status === 401) {
        const retryResponse = await handle401Retry(input, init, headers);
        if (retryResponse) {
          return retryResponse;
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
    [expireSession, handle401Retry, isAuthenticated, makeAuthRequest],
  );

  // Helper: Select best interface by score or keep current
  const selectBestInterface = useCallback((interfaceData: InterfaceInfo[]): void => {
    if (interfaceData.length === 0) {
      return;
    }
    setSelectedInterface((prev) => {
      if (prev) {
        return prev;
      }
      const best = interfaceData.reduce((a, b) => (a.score > b.score ? a : b));
      return best.name;
    });
  }, []);

  const fetchInterfaces = useCallback(async () => {
    if (!isAuthenticated) {
      return;
    }
    try {
      const response = await authFetch('/api/v1/interfaces');
      if (!response.ok) {
        throw new Error('Failed to load interfaces');
      }
      const data = await (response.json() as Promise<unknown>);
      if (!isValidInterfaceArray(data)) {
        throw new Error('Invalid interface data received from server');
      }
      setInterfaces(data);
      selectBestInterface(data);
      setConnected(true);
    } catch (error) {
      const err = error instanceof Error ? error : new Error('Unknown error');
      if (err.message === 'Unauthorized') {
        return;
      }
      setConnected(false);
    }
  }, [authFetch, isAuthenticated, selectBestInterface]);

  // Fetch test result when test completes
  const fetchTestResult = useCallback(async () => {
    try {
      const response = await authFetch('/api/v1/test/result');
      if (!response.ok) {
        return;
      }
      const data = await (response.json() as Promise<TestResult>);
      if (data.status === 'completed' || data.status === 'error') {
        setTestResult(data);
      }
    } catch (error) {
      // Log for debugging but don't disrupt UX for result fetching
      logWarn('Failed to fetch test result', {
        component: 'App',
        action: 'fetchTestResult',
        additionalData: {
          error: error instanceof Error ? error.message : String(error),
        },
      });
    }
  }, [authFetch]);

  // Track previous test status to detect transitions
  const prevTestStatus = useRef<string>('idle');

  // Handle test status transitions - extracted to reduce cognitive complexity
  const handleStatusTransition = useCallback(
    (prevStatus: string, newStatus: string): void => {
      if (isTestCompleted(prevStatus, newStatus)) {
        fetchTestResult().catch(() => {
          // Silent fail - result fetch is non-critical
        });
      }
      if (isTestStarting(prevStatus, newStatus)) {
        setTestResult(null);
      }
      prevTestStatus.current = newStatus;
    },
    [fetchTestResult],
  );

  const fetchStats = useCallback(async () => {
    try {
      const response = await authFetch('/api/v1/stats');
      if (!response.ok) {
        throw new Error('Failed to refresh stats');
      }
      const data = await (response.json() as Promise<unknown>);
      if (!isValidStats(data)) {
        throw new Error('Invalid stats data received from server');
      }
      const newStats = mapStatsPayload(data as Partial<Stats>);
      setStats(newStats);
      handleStatusTransition(prevTestStatus.current, newStats.testStatus);
    } catch (error) {
      if ((error as Error).message === 'Unauthorized') {
        return;
      }
      logWarn('Stats polling failed', {
        component: 'App',
        action: 'fetchStats',
        additionalData: {
          error: error instanceof Error ? error.message : String(error),
        },
      });
    }
  }, [authFetch, handleStatusTransition]);

  const handleLogin: SubmitHandler<{ username: string; password: string }> = useCallback(
    async ({ username, password }) => {
      setLoginLoading(true);
      setLoginError(null);
      try {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          credentials: 'include', // Allow server to set cookies
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password }),
        });
        if (!response.ok) {
          const text = (await (response.text() as Promise<string>)) || 'Authentication failed';
          setLoginError(text);
          setConnected(false);
          return;
        }
        const data = (await response.json()) as
          | { mfaRequired: true; mfaToken: string; factor: string }
          | unknown;
        // Wave 3 (#85): if MFA is required, hold the mfa_token and
        // show the code prompt. The user will POST to
        // /api/v1/auth/login/totp to complete.
        if (
          typeof data === 'object' &&
          data !== null &&
          (data as { mfaRequired?: unknown }).mfaRequired === true
        ) {
          const mfaData = data as { mfaRequired: true; mfaToken: string; factor: string };
          setMfaPending({ mfaToken: mfaData.mfaToken, factor: mfaData.factor });
          return;
        }
        if (!isValidAuthResponse(data)) {
          setLoginError('Authentication failed');
          setConnected(false);
          return;
        }
        // Tokens are now in httpOnly cookies set by server
        // Just mark as authenticated
        setIsAuthenticated(true);
        window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
        setLoginError(null);
        setConnected(true);
      } catch {
        setLoginError('Unable to reach authentication server.');
        setConnected(false);
      } finally {
        setLoginLoading(false);
      }
    },
    [],
  );

  // Wave 3 (#85): submit the MFA code captured by the MFA prompt.
  const handleMFAVerify: SubmitHandler<{ code: string }> = useCallback(
    async ({ code }) => {
      if (!mfaPending) {
        return;
      }
      setLoginLoading(true);
      setLoginError(null);
      try {
        const response = await fetch('/api/v1/auth/login/totp', {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ mfaToken: mfaPending.mfaToken, code }),
        });
        if (!response.ok) {
          const text = (await response.text()) || 'Verification failed';
          setLoginError(text);
          return;
        }
        const data = (await response.json()) as unknown;
        if (!isValidAuthResponse(data)) {
          setLoginError('Verification failed');
          return;
        }
        setIsAuthenticated(true);
        window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
        setMfaPending(null);
        mfaForm.reset();
        setLoginError(null);
        setConnected(true);
      } catch {
        setLoginError('Unable to reach verification endpoint.');
      } finally {
        setLoginLoading(false);
      }
    },
    [mfaPending, mfaForm],
  );

  const handleStartTest = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      return;
    }
    setIsStartingTest(true);
    setTestStartError(null);

    try {
      // Determine test type based on mode
      const testType =
        mode === 'reflector' ? 'reflect' : (selectedTests[0] ?? 'rfc2544_throughput');

      // Build test configuration using helper
      const config = buildTestConfig(testType, {
        rfc2544: rfc2544Config,
        rfc2889: rfc2889Config,
        rfc6349: rfc6349Config,
        y1564: y1564Config,
        y1731: y1731Config,
        tsn: tsnConfig,
        trafficGen: trafficGenConfig,
      });

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
        const errorMessage = await extractResponseError(response, 'Failed to start test');
        setTestStartError(errorMessage);
        return;
      }

      // Status updates will come from polling - don't update optimistically
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to start test';
      setTestStartError(message);
    } finally {
      setIsStartingTest(false);
    }
  }, [
    authFetch,
    mode,
    reflectorProfile,
    isAuthenticated,
    rfc2544Config,
    rfc2889Config,
    rfc6349Config,
    selectedInterface,
    selectedTests,
    trafficGenConfig,
    tsnConfig,
    y1564Config,
    y1731Config,
  ]);

  const handleStopTest = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      return;
    }
    setIsStoppingTest(true);
    try {
      await authFetch('/api/v1/test/stop', { method: 'POST' });
      // Status update will come from polling
    } catch (error) {
      // Log the error but don't disrupt UX - test may already be stopped
      // or the stop request may have actually succeeded
      logError(error, {
        component: 'App',
        action: 'handleStopTest',
      });
    } finally {
      setIsStoppingTest(false);
    }
  }, [authFetch, isAuthenticated]);

  // Frame-size result initialization (previously used by per-module Start
  // buttons) returns alongside the Phase A.1 follow-up that wires each
  // test page's Start button. Helper kept available via setModuleResults
  // when that work lands.

  // Module-level start/stop/configure handlers were previously invoked by
  // the in-page Module Cards grid. After the Phase A router refactor the
  // module cards live on their dedicated /tests/* pages and use the
  // global start/stop button at the top of the shell. The per-module
  // helpers will return in a follow-up commit when each test page wires
  // its own Start/Stop button.

  // Logout handler - clears auth state and calls server to clear cookies
  const handleLogout = useCallback(async () => {
    try {
      // Call server to clear cookies and revoke token
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        credentials: 'include',
      });
    } catch (error) {
      // Log but proceed with local cleanup - user is logging out regardless
      logWarn('Logout API call failed', {
        component: 'App',
        action: 'handleLogout',
        additionalData: {
          error: error instanceof Error ? error.message : String(error),
        },
      });
    }
    // Clear auth state
    setIsAuthenticated(false);
    setConnected(false);
    setLoginError(null);
    window.localStorage.removeItem(AUTH_FLAG_KEY);
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

  // Check setup and recovery status on mount (before authentication check)
  useEffect(() => {
    const checkStatuses = async (): Promise<void> => {
      try {
        // Check setup status
        const setupResponse = await fetch('/api/v1/setup/status', {
          method: 'GET',
          credentials: 'include',
        });
        if (setupResponse.ok) {
          const data = await (setupResponse.json() as Promise<SetupStatus>);
          setSetupStatus(data);
        }

        // Check recovery status
        const recoveryResponse = await fetch('/api/v1/recovery/status', {
          method: 'GET',
          credentials: 'include',
        });
        if (recoveryResponse.ok) {
          const data = await (recoveryResponse.json() as Promise<RecoveryStatus>);
          setRecoveryStatus(data);
        }
      } catch (error) {
        // Log but continue - status check failure shouldn't block the app
        logWarn('Failed to check status', {
          component: 'App',
          action: 'checkStatuses',
          additionalData: {
            error: error instanceof Error ? error.message : String(error),
          },
        });
      } finally {
        setSetupChecked(true);
      }
    };
    checkStatuses().catch(() => {
      // Errors already logged inside checkStatuses
    });
  }, []);

  // Login helper function (shared by login form and setup wizard)
  const performLogin = useCallback(
    async (loginUsername: string, loginPassword: string): Promise<boolean> => {
      try {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            username: loginUsername,
            password: loginPassword,
          }),
        });
        if (!response.ok) {
          return false;
        }
        const data = await (response.json() as Promise<unknown>);
        if (!isValidAuthResponse(data)) {
          return false;
        }
        setIsAuthenticated(true);
        window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
        setConnected(true);
        return true;
      } catch {
        return false;
      }
    },
    [],
  );

  // Handle setup completion
  const handleSetupComplete = useCallback(() => {
    // Clear setup status so we don't show wizard again
    setSetupStatus(null);
  }, []);

  // Handle recovery completion
  const handleRecoveryComplete = useCallback(() => {
    // Clear recovery status and form
    setRecoveryStatus(null);
    setShowRecoveryForm(false);
    setLoginError('Password has been reset. Please sign in with your new password.');
  }, []);

  // Handle back to login from recovery
  const handleBackToLogin = useCallback(() => {
    setShowRecoveryForm(false);
  }, []);

  // Sync auth flag with localStorage
  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    if (isAuthenticated) {
      window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
    } else {
      window.localStorage.removeItem(AUTH_FLAG_KEY);
    }
  }, [isAuthenticated]);

  // Dark mode managed by useTheme(); persists to localStorage.

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
    fetchInterfaces().catch(() => {
      // Errors already handled inside fetchInterfaces
    });
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

    const triggerFetchStats = (): void => {
      fetchStats().catch(() => {
        // Errors already logged inside fetchStats
      });
    };
    statsIntervalRef.current = window.setInterval(triggerFetchStats, 1000);
    triggerFetchStats();

    return () => {
      if (statsIntervalRef.current !== null) {
        clearInterval(statsIntervalRef.current);
        statsIntervalRef.current = null;
      }
    };
  }, [connected, fetchStats]);

  const openHelp = (): void => {
    setHelpOpen(true);
  };

  const openSettings = (): void => {
    setSettingsOpen(true);
  };

  const openHistory = (): void => {
    setHistoryOpen(true);
  };

  const appContextValue: AppContextValue = {
    rfc2544Config,
    setRFC2544Config,
    rfc2889Config,
    setRFC2889Config,
    rfc6349Config,
    setRFC6349Config,
    y1564Config,
    setY1564Config,
    y1731Config,
    setY1731Config,
    tsnConfig,
    setTSNConfig,
    trafficGenConfig,
    setTrafficGenConfig,
    selectedTests,
    testResult,
    interfaces,
    selectedInterface,
    setSelectedInterface,
    stats,
    reflectorProfile,
    setReflectorProfile,
    onStartReflector: () => {
      handleStartTest().catch(() => {
        // Errors surface via testStartError state.
      });
    },
    onStopReflector: () => {
      handleStopTest().catch(() => {
        // Errors are already logged inside handleStopTest.
      });
    },
    isStartingReflector: isStartingTest,
    isStoppingReflector: isStoppingTest,
    reflectorStartError: testStartError,
  };

  const topBar = (
    <div className="px-4 sm:px-6 lg:px-8 pt-6 pb-inline stack-lg">
      {/* Top strip: connection status + role chip + theme/refresh/logout */}
      <div className="flex flex-wrap items-center justify-between gap-default">
        <div className="flex items-center gap-default">
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
        <div className="flex items-center gap-compact">
          <RoleChip />
          <button
            type="button"
            data-testid="header-theme-toggle"
            onClick={toggleTheme}
            className="pad-xs rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover"
            title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {isDark ? (
              <Sun className="h-5 w-5" aria-hidden="true" />
            ) : (
              <Moon className="h-5 w-5" aria-hidden="true" />
            )}
          </button>
          <button
            type="button"
            onClick={fetchInterfaces}
            className="pad-xs rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover"
            title="Refresh interfaces"
            aria-label="Refresh interfaces"
          >
            <RefreshCw className="h-5 w-5" aria-hidden="true" />
          </button>
          <button
            type="button"
            onClick={handleLogout}
            className="pad-xs rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-hover"
            title="Logout"
            aria-label="Logout"
            data-testid="logout-button"
          >
            <LogOut className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>
      </div>

      {/* Test-Master control row — only when role is test_master. The
          Reflector role drives Start/Stop from the Reflector page. The
          per-test-page Start/Stop is coming in Phase A.1 (#64). */}
      {mode === 'test_master' ? (
        <div className="flex flex-wrap items-center gap-default">
          <select
            value={selectedInterface}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
              setSelectedInterface(e.target.value)
            }
            className="w-48"
            aria-label="Select network interface"
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
              aria-busy={isStoppingTest}
            >
              {isStoppingTest ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                  Stopping...
                </>
              ) : (
                <>
                  <Square className="w-4 h-4" aria-hidden="true" />
                  Stop Test
                </>
              )}
            </button>
          ) : (
            <button
              type="button"
              onClick={handleStartTest}
              className="btn btn-primary"
              disabled={!selectedInterface || isStartingTest}
              aria-busy={isStartingTest}
            >
              {isStartingTest ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" aria-hidden="true" />
                  Starting...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" aria-hidden="true" />
                  Start Test
                </>
              )}
            </button>
          )}

          {testStartError ? (
            <div
              className="text-sm text-status-error flex items-center gap-compact"
              role="alert"
              aria-live="assertive"
            >
              <AlertTriangle className="w-4 h-4" aria-hidden="true" />
              {testStartError}
            </div>
          ) : null}

          <div
            className="flex items-center gap-default ml-auto"
            aria-live="polite"
            aria-atomic="true"
          >
            {stats.testStatus === 'running' || stats.testStatus === 'starting' ? (
              <output className="status-badge success flex items-center gap-compact">
                <span
                  className="w-2 h-2 rounded-full bg-status-success animate-pulse"
                  aria-hidden="true"
                />
                {stats.testStatus === 'starting' ? 'Starting' : 'Running'}:{' '}
                {stats.currentTest || mode}
              </output>
            ) : null}
            {stats.testStatus === 'completed' ? (
              <output className="status-badge info">Completed: {stats.currentTest}</output>
            ) : null}
            {stats.testStatus === 'error' ? (
              <output className="status-badge error" role="alert">
                Error: {stats.currentTest || 'Test failed'}
              </output>
            ) : null}
            {stats.testStatus === 'cancelled' ? (
              <output className="status-badge warning">Stopped: {stats.currentTest}</output>
            ) : null}
          </div>
        </div>
      ) : null}

      <TestProgressBar progress={testProgress} />
    </div>
  );

  return (
    <BrowserRouter>
      <AppContext.Provider value={appContextValue}>
        {/* Only mount the authenticated shell once signed in. Rendering the
            full SidebarLayout + lazy routes + live TestResults *behind* the
            login modal was the dominant CLS source (Suspense fallback→page
            swap and WebSocket-driven TestResults height changes shifting the
            background). Unauthenticated users get a stable gradient backdrop
            under the auth overlays — also avoids briefly exposing the app
            shell and loading routes they can't access. */}
        {isAuthenticated ? (
          <>
            <SidebarLayout
              groups={navGroups}
              version={buildVersion.version}
              onOpenHelp={openHelp}
              onOpenSettings={openSettings}
              onOpenHistory={openHistory}
              topBar={topBar}
            >
              <Suspense fallback={<PageLoader />}>
                <Routes>
                  <Route path="/" element={<Navigate to="/reflector" replace={true} />} />
                  {pages.map((page) => (
                    <Route key={page.path} path={page.path} element={<page.component />} />
                  ))}
                  <Route path="*" element={<Navigate to="/reflector" replace={true} />} />
                </Routes>
              </Suspense>

              {/* Pinned below the routed page so test outcomes stay visible no
              matter which page is active. */}
              <div className="mt-6">
                <TestResults testStatus={stats.testStatus} result={testResult} />
              </div>
            </SidebarLayout>

            <SettingsDrawer
              isOpen={settingsOpen}
              onClose={() => setSettingsOpen(false)}
              selectedTests={selectedTests}
              setSelectedTests={setSelectedTests}
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

            <HelpDrawer isOpen={helpOpen} onClose={() => setHelpOpen(false)} />

            <ResultHistory
              isOpen={historyOpen}
              onClose={() => setHistoryOpen(false)}
              currentResult={testResult}
            />
          </>
        ) : (
          <div className="min-h-screen bg-gradient-to-br from-surface-base via-surface-raised to-surface-deep" />
        )}

        {/* Setup Wizard - shown before login if initial setup required */}
        {setupChecked && setupStatus?.needsSetup ? (
          <SetupWizard
            onComplete={handleSetupComplete}
            onLogin={performLogin}
            suggestedPassword={setupStatus.suggestedPassword}
            username={setupStatus.username}
            setupToken={setupStatus.setupToken}
          />
        ) : null}

        {/* Recovery Form - shown when user clicks "Forgot Password" and recovery is available */}
        {showRecoveryForm && recoveryStatus?.active ? (
          <RecoveryForm
            onRecoveryComplete={handleRecoveryComplete}
            onBackToLogin={handleBackToLogin}
            remainingTime={recoveryStatus.remainingTime}
            tokenFilePath={recoveryStatus.instructions}
          />
        ) : null}

        {/* Login Modal - shown after setup complete or if setup not needed */}
        {!isAuthenticated && setupChecked && !setupStatus?.needsSetup && !showRecoveryForm ? (
          <div className="fixed inset-0 z-50 flex-center bg-scrim/60 backdrop-blur-sm">
            <div
              ref={loginModalRef}
              role="dialog"
              aria-modal="true"
              aria-labelledby="login-dialog-title"
              className="w-full max-w-md rounded-3xl border border-surface-border bg-surface-raised pad-lg shadow-2xl"
            >
              <h2
                id="login-dialog-title"
                className="flex items-center gap-compact heading-3 text-text-primary"
              >
                <Lock className="w-5 h-5 text-brand-primary" />
                Sign in to continue
              </h2>
              <p className="text-sm text-text-muted">Authenticate with your Stem credentials.</p>
              {mfaPending ? (
                <form className="mt-6 stack-lg" onSubmit={mfaForm.handleSubmit(handleMFAVerify)}>
                  <p className="text-sm text-text-muted">
                    Enter the code from your authenticator app to continue.
                  </p>
                  <div>
                    <label
                      htmlFor="stem-login-mfa"
                      className="text-xs font-semibold text-text-muted"
                    >
                      Verification code
                    </label>
                    <input
                      id="stem-login-mfa"
                      type="text"
                      inputMode="numeric"
                      pattern="[0-9]{6}"
                      autoComplete="one-time-code"
                      {...mfaForm.register('code')}
                      className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm font-mono tracking-widest text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                    />
                    {mfaForm.formState.errors.code ? (
                      <p className="mt-tight text-xs text-status-error">
                        {mfaForm.formState.errors.code.message}
                      </p>
                    ) : null}
                  </div>
                  {loginError ? <p className="text-xs text-status-error">{loginError}</p> : null}
                  <button
                    type="submit"
                    className="btn btn-primary w-full justify-center"
                    disabled={loginLoading}
                  >
                    {loginLoading ? 'Verifying...' : 'Verify'}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setMfaPending(null);
                      mfaForm.reset();
                      setLoginError(null);
                    }}
                    className="w-full mt-inline text-sm text-text-muted hover:text-brand-primary"
                  >
                    Use different account
                  </button>
                </form>
              ) : (
                <form className="mt-6 stack-lg" onSubmit={loginForm.handleSubmit(handleLogin)}>
                  <div>
                    <label
                      htmlFor="stem-login-username"
                      className="text-xs font-semibold text-text-muted"
                    >
                      Username
                    </label>
                    <input
                      id="stem-login-username"
                      type="text"
                      autoComplete="username"
                      {...loginForm.register('username')}
                      className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                    />
                    {loginForm.formState.errors.username ? (
                      <p className="mt-tight text-xs text-status-error">
                        {loginForm.formState.errors.username.message}
                      </p>
                    ) : null}
                  </div>
                  <div>
                    <label
                      htmlFor="stem-login-password"
                      className="text-xs font-semibold text-text-muted"
                    >
                      Password
                    </label>
                    <input
                      id="stem-login-password"
                      type="password"
                      autoComplete="current-password"
                      {...loginForm.register('password')}
                      className="mt-tight w-full rounded-xl border border-surface-border bg-surface-base px-3 py-row text-sm text-text-primary focus:border-brand-primary focus:outline-none focus:ring-2 focus:ring-brand-primary/30"
                    />
                    {loginForm.formState.errors.password ? (
                      <p className="mt-tight text-xs text-status-error">
                        {loginForm.formState.errors.password.message}
                      </p>
                    ) : null}
                  </div>
                  {loginError ? <p className="text-xs text-status-error">{loginError}</p> : null}
                  <button
                    type="submit"
                    className="btn btn-primary w-full justify-center"
                    disabled={loginLoading}
                  >
                    {loginLoading ? 'Signing in...' : 'Sign In'}
                  </button>

                  {/* Forgot Password link - only shown when recovery is available */}
                  {recoveryStatus?.active ? (
                    <button
                      type="button"
                      onClick={() => setShowRecoveryForm(true)}
                      className="w-full mt-content text-sm text-text-muted hover:text-brand-primary"
                    >
                      Forgot password?
                    </button>
                  ) : null}
                </form>
              )}
            </div>
          </div>
        ) : null}

        {/* Command palette (⌘K / Ctrl+K) — authenticated feature only */}
        {isAuthenticated ? (
          <CommandPalette
            groups={navGroups}
            open={paletteOpen}
            onOpenChange={setPaletteOpen}
            onOpenSettings={openSettings}
            onOpenHelp={openHelp}
            onToggleTheme={toggleTheme}
            isDark={isDark}
          />
        ) : null}
      </AppContext.Provider>
    </BrowserRouter>
  );
}

// Wrapper component that provides context
function App(): ReactElement {
  return (
    <RoleProvider>
      <ModuleSettingsProvider>
        <AppContent />
      </ModuleSettingsProvider>
    </RoleProvider>
  );
}

export default App;

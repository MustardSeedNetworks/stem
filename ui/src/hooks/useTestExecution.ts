/**
 * @fileoverview useTestExecution — test/interface/stats orchestration hook.
 * @description Owns the interfaces + stats React Query reads, the connection
 *              indicator, the test start/stop handlers, and the test-status
 *              state machine that fetches a result on completion. Extracted
 *              verbatim from App.tsx during the W5.5 providers+routing
 *              decomposition — behavior is unchanged.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { RFC2544Config } from '../components/RFC2544ConfigForm';
import type { RFC2889Config } from '../components/RFC2889ConfigForm';
import type { RFC6349Config } from '../components/RFC6349ConfigForm';
import { useTestProgress } from '../components/TestProgressBar';
import type { TrafficGenConfig } from '../components/TrafficGenConfigForm';
import type { TSNConfig } from '../components/TSNConfigForm';
import type { Y1564Config } from '../components/Y1564ConfigForm';
import type { Y1731Config } from '../components/Y1731ConfigForm';
import { useRole } from '../contexts/RoleContext';
import { authFetch, useAuthStore } from '../stores/auth-store';
import { type StopOutcome, useTestStore } from '../stores/test-store';
import {
  type InterfaceInfo,
  initialStats,
  isValidInterfaceArray,
  isValidStats,
  type Stats,
  type TestResult,
} from '../types/api';
import { logError, logWarn } from '../utils/logger';

// Helper: check if test just completed (status transition to completed/error)
function isTestCompleted(prev: string, curr: string): boolean {
  const terminal = curr === 'completed' || curr === 'error' || curr === 'cancelled';
  return terminal && prev !== curr;
}

// Helper: check if new test is starting
function isTestStarting(prev: string, curr: string): boolean {
  return curr === 'starting' && prev !== 'starting';
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
    errorMessage: payload.errorMessage,
    suiteId: payload.suiteId ?? '',
    steps: payload.steps ?? [],
    currentStep: Number(payload.currentStep ?? 0),
    stepsComplete: Number(payload.stepsComplete ?? 0),
    stepsTotal: Number(payload.stepsTotal ?? 0),
    phase: payload.phase ?? '',
    elapsedSeconds: Number(payload.elapsedSeconds ?? 0),
    estimatedRemainingSeconds: payload.estimatedRemainingSeconds ?? null,
  };
}

/** The daemon's error envelope, as `internal/api/errors.go` writes it. */
interface ApiErrorBody {
  error?: string;
  code?: string;
  message?: string;
  requiredFeature?: string;
}

/**
 * Why a request was refused, in the shape the render path needs.
 *
 * The 402 entitlement answer is not a generic failure: it carries the feature
 * the licence is missing, so the operator is told what to buy rather than that
 * something went wrong (#1070). The server's own `upgradeMessage` is CLI prose
 * and is deliberately not rendered — the message is built from the translation
 * catalog at the call site. Its `currentTier` is not rendered either: a fresh
 * install reports "Invalid" (license.TierInvalid), which is a state name, not
 * a tier an operator has heard of (#1095).
 */
export type RequestFailure =
  | { kind: 'message'; message?: string }
  | { kind: 'featureGate'; feature?: string };

/**
 * Classify a refused request from its body.
 *
 * The daemon's envelope (`HTTPErrorResponse`) carries the error *type* in
 * `error` ("Bad Request") and the sentence written for the operator in
 * `message`, so `message` is what gets rendered when it is present.
 */
export async function classifyFailure(response: Response): Promise<RequestFailure> {
  try {
    const body = await (response.json() as Promise<ApiErrorBody>);
    if (response.status === 402 && body?.code === 'TIER_TOO_LOW') {
      return { kind: 'featureGate', feature: body.requiredFeature };
    }
    return { kind: 'message', message: body?.message ?? body?.error };
  } catch {
    return { kind: 'message' };
  }
}

/**
 * Resolve a stop response into the outcome the UI renders.
 *
 * A stop the daemon refuses — "No test is currently running" — is an answer,
 * not an error: it is reported to the operator, not logged and dropped
 * (#1080).
 */
export async function resolveStopOutcome(
  response: Response,
  fallbackMessage: string,
): Promise<StopOutcome> {
  if (response.ok) {
    return { kind: 'stopped' };
  }
  const failure = await classifyFailure(response);
  return {
    kind: 'stopRejected',
    message: failure.kind === 'message' ? (failure.message ?? fallbackMessage) : fallbackMessage,
  };
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

export interface UseTestExecution {
  connected: boolean;
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  setSelectedInterface: (name: string) => void;
  stats: Stats;
  testResult: TestResult | null;
  testProgress: ReturnType<typeof useTestProgress>;
  isStartingTest: boolean;
  stopOutcome: StopOutcome;
  testStartError: string | null;
  handleStartTest: () => Promise<void>;
  handleStopTest: () => Promise<void>;
  refetchInterfaces: () => void;
}

/**
 * Orchestrates test execution and the live interface/stats reads for the
 * Stem shell. State that the routed pages consume is surfaced through the
 * returned object (and threaded into AppContext by the caller).
 */
export function useTestExecution(): UseTestExecution {
  const { t } = useTranslation(['common', 'errors']);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  // The Stem instance role drives the legacy `mode` state. RoleContext
  // persists the choice to localStorage and is mutated by the header
  // RoleChip and per-page RoleGuard.
  const { role: mode } = useRole();
  // Test-execution config + run state lives in the test-store. Its setters are
  // useState-compatible (accept a value or an updater), so every call site
  // below — config forms and the selected-tests toggles — is unchanged.
  const {
    selectedTests,
    setSelectedTests,
    reflectorProfile,
    isStartingTest,
    setIsStartingTest,
    stopOutcome,
    setStopOutcome,
    testStartError,
    setTestStartError,
    rfc2544Config,
    rfc2889Config,
    rfc6349Config,
    y1564Config,
    y1731Config,
    tsnConfig,
    trafficGenConfig,
  } = useTestStore();

  // `connected` (interfaces reachable) is App-local: seeded from the persisted
  // auth flag to avoid a "Disconnected" flash on reload, then reconciled off
  // the interfaces query (success → true) and the auth flag (signed out → false).
  const [connected, setConnected] = useState<boolean>(
    () => useAuthStore.getState().isAuthenticated,
  );
  const [testResult, setTestResult] = useState<TestResult | null>(null);
  const [selectedInterface, setSelectedInterface] = useState<string>('');

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

  // Interfaces are a React Query read. Connection + best-interface selection
  // are driven off the query result below (see the interface effects). The
  // queryFn closes over the module-level authFetch; retry is disabled to match
  // the previous single-attempt fetch semantics.
  const {
    data: interfacesData,
    refetch: refetchInterfaces,
    error: interfacesError,
  } = useQuery({
    queryKey: ['interfaces'],
    enabled: isAuthenticated,
    retry: false,
    queryFn: async ({ signal }): Promise<InterfaceInfo[]> => {
      const response = await authFetch('/api/v1/interfaces', { signal });
      if (!response.ok) {
        throw new Error('Failed to load interfaces');
      }
      const data = await (response.json() as Promise<unknown>);
      if (!isValidInterfaceArray(data)) {
        throw new Error('Invalid interface data received from server');
      }
      return data;
    },
  });
  const interfaces = interfacesData ?? [];

  // Fetch test result when test completes
  const fetchTestResult = useCallback(async () => {
    try {
      const response = await authFetch('/api/v1/test/result');
      if (!response.ok) {
        return;
      }
      const data = await (response.json() as Promise<TestResult>);
      if (data.status === 'completed' || data.status === 'error' || data.status === 'cancelled') {
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
  }, []);

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

  // Stats are polled via React Query while connected (1s interval, in the
  // background too — matching the previous setInterval). retry:false +
  // staleTime:0 keep the cadence single-shot-per-tick and always fresh; the
  // status-transition side-effect runs off the data below.
  const { data: statsData } = useQuery({
    queryKey: ['stats'],
    enabled: connected,
    refetchInterval: 1000,
    refetchIntervalInBackground: true,
    staleTime: 0,
    retry: false,
    queryFn: async ({ signal }): Promise<Stats> => {
      const response = await authFetch('/api/v1/stats', { signal });
      if (!response.ok) {
        throw new Error('Failed to refresh stats');
      }
      const data = await (response.json() as Promise<unknown>);
      if (!isValidStats(data)) {
        throw new Error('Invalid stats data received from server');
      }
      return mapStatsPayload(data as Partial<Stats>);
    },
  });
  const stats = statsData ?? initialStats;

  const testProgress = useTestProgress(stats);

  const handleStartTest = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      return;
    }
    setIsStartingTest(true);
    setTestStartError(null);
    setStopOutcome({ kind: 'idle' });

    try {
      const configs = {
        rfc2544: rfc2544Config,
        rfc2889: rfc2889Config,
        rfc6349: rfc6349Config,
        y1564: y1564Config,
        y1731: y1731Config,
        tsn: tsnConfig,
        trafficGen: trafficGenConfig,
      };
      const tests = selectedTests.map((testType) => ({
        testType,
        config: buildTestConfig(testType, configs),
      }));

      const response = await authFetch('/api/v1/test/start', {
        method: 'POST',
        body: JSON.stringify({
          interface: selectedInterface,
          profile: tests.some((step) => step.testType === 'reflect') ? reflectorProfile : undefined,
          tests,
        }),
      });

      // Check for validation errors in response
      if (!response.ok) {
        const failure = await classifyFailure(response);
        setTestStartError(
          failure.kind === 'featureGate'
            ? t('errors:test.featureGate', { feature: failure.feature })
            : (failure.message ?? t('errors:test.failedToStart')),
        );
        return;
      }

      // Status updates will come from polling - don't update optimistically
    } catch (error) {
      const message = error instanceof Error ? error.message : t('errors:test.failedToStart');
      setTestStartError(message);
    } finally {
      setIsStartingTest(false);
    }
  }, [
    mode,
    reflectorProfile,
    isAuthenticated,
    t,
    rfc2544Config,
    rfc2889Config,
    rfc6349Config,
    selectedInterface,
    selectedTests,
    trafficGenConfig,
    tsnConfig,
    y1564Config,
    y1731Config,
    setIsStartingTest,
    setTestStartError,
    setStopOutcome,
  ]);

  const handleStopTest = useCallback(async (): Promise<void> => {
    if (!isAuthenticated) {
      return;
    }
    setStopOutcome({ kind: 'stopping' });
    try {
      const response = await authFetch('/api/v1/test/stop', { method: 'POST' });
      setStopOutcome(await resolveStopOutcome(response, t('errors:test.failedToStop')));
    } catch (error) {
      // The request never reached the daemon; a refusal is not this branch.
      logError(error, {
        component: 'App',
        action: 'handleStopTest',
      });
      setStopOutcome({ kind: 'stopFailed', message: t('errors:test.failedToStop') });
    }
  }, [isAuthenticated, t, setStopOutcome]);

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
  }, [mode, setSelectedTests]);

  // Signing out (store flips isAuthenticated false on logout/expiry) drops the
  // connection indicator; the interfaces query owns the reconnect (→ true) side.
  useEffect(() => {
    if (!isAuthenticated) {
      setConnected(false);
    }
  }, [isAuthenticated]);

  // Drive interface auto-selection + connection state off the interfaces query.
  // (The query itself fetches on mount and whenever isAuthenticated flips true.)
  useEffect(() => {
    if (interfacesData) {
      selectBestInterface(interfacesData);
      setConnected(true);
    }
  }, [interfacesData, selectBestInterface]);

  // A genuine load failure drops the connection; an auth lapse does not (the
  // auth flow owns that transition), matching the previous fetch's behavior.
  useEffect(() => {
    if (interfacesError && interfacesError.message !== 'Unauthorized') {
      setConnected(false);
    }
  }, [interfacesError]);

  // Each new stats payload drives the test-status state machine (detect
  // start/complete transitions, fetch the result on completion). The polling
  // itself is owned by the stats query's refetchInterval above.
  useEffect(() => {
    if (statsData) {
      handleStatusTransition(prevTestStatus.current, statsData.testStatus);
    }
  }, [statsData, handleStatusTransition]);

  return {
    connected,
    interfaces,
    selectedInterface,
    setSelectedInterface,
    stats,
    testResult,
    testProgress,
    isStartingTest,
    stopOutcome,
    testStartError,
    handleStartTest,
    handleStopTest,
    refetchInterfaces: () => {
      refetchInterfaces().catch(() => {
        // Connection state is reconciled by the interface effects.
      });
    },
  };
}

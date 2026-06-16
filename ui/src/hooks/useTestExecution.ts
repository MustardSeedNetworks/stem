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
import { useTestStore } from '../stores/test-store';
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
  return (curr === 'completed' || curr === 'error') && prev !== 'completed' && prev !== 'error';
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

export interface UseTestExecution {
  connected: boolean;
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  setSelectedInterface: (name: string) => void;
  stats: Stats;
  testResult: TestResult | null;
  testProgress: ReturnType<typeof useTestProgress>;
  isStartingTest: boolean;
  isStoppingTest: boolean;
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
    isStoppingTest,
    setIsStoppingTest,
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

  // Calculate expected test duration based on config.
  const expectedDuration =
    (rfc2544Config.duration + rfc2544Config.warmup) *
    rfc2544Config.trials *
    rfc2544Config.frameSizes.length *
    selectedTests.filter((t) => t.startsWith('rfc2544')).length;

  // Track test progress.
  const testProgress = useTestProgress(stats.testStatus, stats.currentTest, expectedDuration);

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
    setIsStartingTest,
    setTestStartError,
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
  }, [isAuthenticated, setIsStoppingTest]);

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
    isStoppingTest,
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

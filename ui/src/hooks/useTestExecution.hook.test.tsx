/**
 * @fileoverview useTestExecution — the hook itself, mounted against a fake
 *               daemon.
 * @description The pure helpers have their own suites; these drive the hook's
 *              wiring: which interface is picked, what Start actually POSTs,
 *              how a refused start or stop reaches the operator, and the
 *              status machine that pins a finished run's result (#824).
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { StemRole } from '../contexts/RoleContext';
import { authFetch, useAuthStore } from '../stores/auth-store';
import { useTestStore } from '../stores/test-store';
import { useTestExecution } from './useTestExecution';

vi.mock('../stores/auth-store', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../stores/auth-store')>()),
  authFetch: vi.fn(),
}));

let role: StemRole = 'test_master';
vi.mock('../contexts/RoleContext', () => ({
  useRole: () => ({ role }),
}));

vi.mock('../utils/logger', () => ({
  logError: vi.fn(),
  logWarn: vi.fn(),
}));

type Route = () => Response | Promise<Response>;

function json(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

const IFACES = [
  { name: 'eth0', mac: '00:00:00:00:00:01', speed: 1000, score: 10 },
  { name: 'eth1', mac: '00:00:00:00:00:02', speed: 10000, score: 90 },
  { name: 'eth2', mac: '00:00:00:00:00:03', speed: 100, score: 40 },
];

let routes: Record<string, Route>;
let statsStatus: string;

function stats(): Response {
  return json(200, { uptime: 5, testStatus: statsStatus });
}

function fetched(path: string): RequestInit[] {
  return vi
    .mocked(authFetch)
    .mock.calls.filter(([input]) => input === path)
    .map(([, init]) => init ?? {});
}

function postedBody(path: string): Record<string, unknown> {
  const calls = fetched(path);
  expect(calls).toHaveLength(1);
  return JSON.parse(String(calls[0].body)) as Record<string, unknown>;
}

let client: QueryClient;

function mount() {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return renderHook(() => useTestExecution(), { wrapper });
}

async function mountConnected() {
  const view = mount();
  await waitFor(() => expect(view.result.current.selectedInterface).toBe('eth1'));
  return view;
}

/**
 * Run async hook work inside act and wait for it to finish. Awaiting act itself
 * trips Biome's useAwaitThenable, which cannot see act's thenable overload.
 */
async function settle(work: () => Promise<unknown>): Promise<void> {
  let done = false;
  act(() => {
    void work().finally(() => {
      done = true;
    });
  });
  await waitFor(() => expect(done).toBe(true));
}

async function pollStats(status: string) {
  statsStatus = status;
  await settle(() => client.refetchQueries({ queryKey: ['stats'] }));
  await waitFor(() =>
    expect(client.getQueryData<{ testStatus: string }>(['stats'])?.testStatus).toBe(status),
  );
}

beforeEach(() => {
  role = 'test_master';
  statsStatus = 'idle';
  client = new QueryClient();
  routes = {
    '/api/v1/interfaces': () => json(200, IFACES),
    '/api/v1/stats': stats,
    '/api/v1/test/start': () => json(202, { status: 'starting' }),
    '/api/v1/test/stop': () => json(200, { status: 'stopped' }),
  };
  vi.mocked(authFetch).mockImplementation(async (input) => {
    const route = routes[String(input)];
    if (!route) {
      throw new Error(`unexpected request ${String(input)}`);
    }
    return route();
  });
  useAuthStore.setState({ isAuthenticated: true });
  useTestStore.setState(useTestStore.getInitialState());
});

afterEach(() => {
  client.clear();
  vi.mocked(authFetch).mockReset();
});

describe('interfaces and connection', () => {
  it('connects and picks the highest-scoring interface', async () => {
    const { result } = await mountConnected();

    expect(result.current.connected).toBe(true);
    expect(result.current.interfaces.map((i) => i.name)).toEqual(['eth0', 'eth1', 'eth2']);
  });

  it('keeps an interface the operator already chose across a refetch', async () => {
    const { result } = await mountConnected();

    act(() => result.current.setSelectedInterface('eth2'));
    act(() => result.current.refetchInterfaces());

    await waitFor(() => expect(fetched('/api/v1/interfaces')).toHaveLength(2));
    expect(result.current.selectedInterface).toBe('eth2');
  });

  it('drops the connection when the interface list cannot load', async () => {
    routes['/api/v1/interfaces'] = () => json(500, { error: 'boom' });

    const { result } = mount();

    await waitFor(() => expect(result.current.connected).toBe(false));
    expect(result.current.interfaces).toEqual([]);
  });

  it('drops the connection on a malformed interface list', async () => {
    routes['/api/v1/interfaces'] = () => json(200, [{ name: 'eth0' }]);

    const { result } = mount();

    await waitFor(() => expect(result.current.connected).toBe(false));
  });

  it('disconnects when the session signs out', async () => {
    const { result } = await mountConnected();

    act(() => useAuthStore.setState({ isAuthenticated: false }));

    expect(result.current.connected).toBe(false);
  });
});

describe('role drives the selected tests', () => {
  it('runs only the reflector in reflector mode', () => {
    role = 'reflector';
    mount();

    expect(useTestStore.getState().selectedTests).toEqual(['reflect']);
  });

  it('restores the RFC 2544 defaults when leaving reflector mode', () => {
    useTestStore.setState({ selectedTests: ['reflect'] });
    mount();

    expect(useTestStore.getState().selectedTests).toEqual([
      'rfc2544_throughput',
      'rfc2544_latency',
      'rfc2544_frame_loss',
      'rfc2544_back_to_back',
    ]);
  });

  it('keeps a real test-master selection', () => {
    useTestStore.setState({ selectedTests: ['y1564_config'] });
    mount();

    expect(useTestStore.getState().selectedTests).toEqual(['y1564_config']);
  });
});

describe('handleStartTest', () => {
  it('posts the selection with each step’s config block and the trimmed peer', async () => {
    useTestStore.setState({ selectedTests: ['rfc2544_throughput', 'custom_stream', 'unknown'] });
    const { result } = await mountConnected();
    act(() => {
      result.current.setPeer('  10.0.0.9 ');
      result.current.setPeerPort(4000);
    });

    await settle(() => result.current.handleStartTest());

    const { rfc2544Config, trafficGenConfig } = useTestStore.getState();
    expect(postedBody('/api/v1/test/start')).toEqual({
      interface: 'eth1',
      peer: '10.0.0.9',
      peerPort: 4000,
      tests: [
        { testType: 'rfc2544_throughput', config: { rfc2544: rfc2544Config } },
        { testType: 'custom_stream', config: { trafficGen: trafficGenConfig } },
        { testType: 'unknown' },
      ],
    });
    expect(fetched('/api/v1/test/start')[0].method).toBe('POST');
    expect(result.current.isStartingTest).toBe(false);
    expect(result.current.testStartError).toBeNull();
  });

  it('sends the reflector profile, and no peer, for an explicit reflect run', async () => {
    role = 'reflector';
    useTestStore.setState({ reflectorProfile: 'msn' });
    const { result } = await mountConnected();
    act(() => result.current.setPeer('10.0.0.9'));

    await settle(() => result.current.handleStartTest(['reflect']));

    expect(postedBody('/api/v1/test/start')).toEqual({
      interface: 'eth1',
      profile: 'msn',
      tests: [{ testType: 'reflect' }],
    });
  });

  it('names the missing feature when the licence refuses the run', async () => {
    routes['/api/v1/test/start'] = () =>
      json(402, { code: 'TIER_TOO_LOW', requiredFeature: 'rfc2889' });
    const { result } = await mountConnected();

    await settle(() => result.current.handleStartTest());

    expect(result.current.testStartError).toContain('rfc2889');
    expect(result.current.isStartingTest).toBe(false);
  });

  it('shows the daemon’s sentence for a refused start', async () => {
    routes['/api/v1/test/start'] = () =>
      json(400, { error: 'Bad Request', message: 'Interface eth1 is down' });
    const { result } = await mountConnected();

    await settle(() => result.current.handleStartTest());

    expect(result.current.testStartError).toBe('Interface eth1 is down');
  });

  it('falls back to the catalog message when the refusal has no body', async () => {
    routes['/api/v1/test/start'] = () => new Response('', { status: 500 });
    const { result } = await mountConnected();

    await settle(() => result.current.handleStartTest());

    expect(result.current.testStartError).toBeTruthy();
    expect(result.current.testStartError).not.toContain('500');
  });

  it('reports a start that never reached the daemon', async () => {
    routes['/api/v1/test/start'] = () => {
      throw new Error('network down');
    };
    const { result } = await mountConnected();

    await settle(() => result.current.handleStartTest());

    expect(result.current.testStartError).toBe('network down');
    expect(result.current.isStartingTest).toBe(false);
  });

  it('clears a previous stop outcome and error when a new run starts', async () => {
    useTestStore.setState({
      stopOutcome: { kind: 'stopped' },
      testStartError: 'old failure',
    });
    const { result } = await mountConnected();

    await settle(() => result.current.handleStartTest());

    expect(result.current.stopOutcome).toEqual({ kind: 'idle' });
    expect(result.current.testStartError).toBeNull();
  });

  it('does nothing while signed out', async () => {
    const { result } = await mountConnected();
    act(() => useAuthStore.setState({ isAuthenticated: false }));

    await settle(() => result.current.handleStartTest());
    await settle(() => result.current.handleStopTest());

    expect(fetched('/api/v1/test/start')).toEqual([]);
    expect(fetched('/api/v1/test/stop')).toEqual([]);
  });
});

describe('handleStopTest', () => {
  it('records a stop the daemon accepted', async () => {
    const { result } = await mountConnected();

    await settle(() => result.current.handleStopTest());

    expect(fetched('/api/v1/test/stop')[0].method).toBe('POST');
    expect(result.current.stopOutcome).toEqual({ kind: 'stopped' });
  });

  it('shows the daemon’s refusal of a stop', async () => {
    routes['/api/v1/test/stop'] = () =>
      json(400, { error: 'Bad Request', message: 'No test is currently running' });
    const { result } = await mountConnected();

    await settle(() => result.current.handleStopTest());

    expect(result.current.stopOutcome).toEqual({
      kind: 'stopRejected',
      message: 'No test is currently running',
    });
  });

  it('separates a stop that never arrived from a refusal', async () => {
    routes['/api/v1/test/stop'] = () => {
      throw new Error('network down');
    };
    const { result } = await mountConnected();

    await settle(() => result.current.handleStopTest());

    expect(result.current.stopOutcome.kind).toBe('stopFailed');
  });
});

describe('status transitions', () => {
  it('fetches and pins the result when a run completes', async () => {
    routes['/api/v1/test/result'] = () =>
      json(200, { testType: 'rfc2544_throughput', status: 'completed', success: true });
    const { result } = await mountConnected();

    await pollStats('running');
    expect(result.current.stats.testStatus).toBe('running');
    expect(fetched('/api/v1/test/result')).toEqual([]);

    await pollStats('completed');

    await waitFor(() =>
      expect(result.current.testResult).toMatchObject({
        testType: 'rfc2544_throughput',
        success: true,
      }),
    );
    expect(result.current.testProgress.status).toBe('completed');
  });

  it('clears the pinned result when the next run starts', async () => {
    routes['/api/v1/test/result'] = () =>
      json(200, { testType: 'rfc2544_throughput', status: 'completed', success: true });
    const { result } = await mountConnected();
    await pollStats('running');
    await pollStats('completed');
    await waitFor(() => expect(result.current.testResult).not.toBeNull());

    await pollStats('starting');

    await waitFor(() => expect(result.current.testResult).toBeNull());
  });

  it('pins an error result when the result request fails', async () => {
    routes['/api/v1/test/result'] = () => {
      throw new Error('socket closed');
    };
    const { result } = await mountConnected();
    await pollStats('running');

    await pollStats('error');

    await waitFor(() =>
      expect(result.current.testResult).toMatchObject({
        success: false,
        error: 'Result unavailable: socket closed',
      }),
    );
  });

  it('keeps the last stats when a poll answers with garbage', async () => {
    const { result } = await mountConnected();
    await pollStats('running');

    routes['/api/v1/stats'] = () => json(200, { nothing: true });
    await settle(() => client.refetchQueries({ queryKey: ['stats'] }));

    expect(result.current.stats.testStatus).toBe('running');
  });
});

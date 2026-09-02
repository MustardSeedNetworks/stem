/**
 * ReflectorPage tests.
 *
 * The page's job is to tell an operator, at a glance, whether the reflector
 * can run here and whether it is running. Two states are easy to get wrong and
 * expensive when wrong: a platform that has no reflector dataplane at all
 * (macOS / Windows builds), where zeroed counters would read as "running fine
 * with no traffic"; and a run that ended in error, which must not read as idle.
 * These assert what the operator sees, not which helpers were called.
 */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { AppContextValue } from '../contexts/AppContext';
import type { RoleContextValue } from '../contexts/RoleContext';
import type { Capabilities } from '../hooks/useCapabilities';
import type { InterfaceInfo, Stats } from '../types/api';
import { ReflectorPage } from './ReflectorPage';

const { appContext, roleContext, capabilities } = vi.hoisted(() => ({
  appContext: { current: null as unknown as AppContextValue },
  roleContext: { current: null as unknown as RoleContextValue },
  capabilities: { current: null as unknown as Capabilities },
}));

vi.mock('../contexts/AppContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../contexts/AppContext')>()),
  useAppContext: () => appContext.current,
}));

vi.mock('../contexts/RoleContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../contexts/RoleContext')>()),
  useRole: () => roleContext.current,
}));

vi.mock('../hooks/useCapabilities', () => ({
  useCapabilities: () => capabilities.current,
}));

const eth0: InterfaceInfo = {
  name: 'eth0',
  mac: '00:11:22:33:44:55',
  speed: 10000,
  duplex: 'full',
  driver: 'ixgbe',
  state: 'up',
  xdp: true,
  score: 92,
} as InterfaceInfo;

function makeStats(overrides: Partial<Stats> = {}): Stats {
  return {
    packetsReceived: 0,
    packetsSent: 0,
    bytesReceived: 0,
    bytesSent: 0,
    currentPps: 0,
    currentMbps: 0,
    uptime: 0,
    testStatus: 'idle',
    ...overrides,
  } as Stats;
}

interface Options {
  stats?: Partial<Stats>;
  reflectorSupported?: boolean;
  reason?: string;
  selectedInterface?: string;
  reflectorStartError?: string | null;
  isStartingReflector?: boolean;
  isStoppingReflector?: boolean;
}

function renderPage(options: Options = {}) {
  const setRole = vi.fn();
  const onStartReflector = vi.fn();
  const onStopReflector = vi.fn();

  capabilities.current = {
    reflector:
      options.reflectorSupported === false
        ? { supported: false, reason: options.reason }
        : { supported: true },
    testMaster: { supported: true },
  };

  roleContext.current = {
    role: 'reflector',
    setRole,
    isSwitchingRole: false,
    roleSwitchError: null,
    clearRoleSwitchError: vi.fn(),
  };

  appContext.current = {
    interfaces: [eth0],
    selectedInterface: options.selectedInterface ?? 'eth0',
    setSelectedInterface: vi.fn(),
    stats: makeStats(options.stats),
    reflectorProfile: 'netally',
    setReflectorProfile: vi.fn(),
    onStartReflector,
    onStopReflector,
    isStartingReflector: options.isStartingReflector ?? false,
    isStoppingReflector: options.isStoppingReflector ?? false,
    reflectorStartError: options.reflectorStartError ?? null,
  } as unknown as AppContextValue;

  render(<ReflectorPage />);
  return { setRole, onStartReflector, onStopReflector };
}

describe('ReflectorPage — platform support', () => {
  afterEach(cleanup);

  it('reports counters as unavailable rather than as zeroes on an unsupported platform', () => {
    renderPage({ reflectorSupported: false, reason: 'CGO + Linux required' });

    expect(
      screen.getByText('Reflector counters are not available on this platform'),
    ).toBeInTheDocument();
    // The reason appears twice by design — in the rollup body and again in
    // the banner. Scope to the banner so this asserts one place, not both.
    expect(screen.getByTestId('reflector-platform-banner')).toHaveTextContent(
      'CGO + Linux required',
    );
    expect(screen.queryByText('Reflector is idle')).toBeNull();
  });

  it('disables Start and says why when the dataplane is missing', () => {
    renderPage({ reflectorSupported: false, reason: 'CGO + Linux required' });

    const start = screen.getByTestId('reflector-start-button');
    expect(start).toBeDisabled();
    expect(start).toHaveAttribute('aria-disabled', 'true');
    expect(start).toHaveAttribute('title');
    expect(start.getAttribute('title')).not.toBe('');
  });

  it('offers the Test Master switch from the platform banner', async () => {
    const { setRole } = renderPage({ reflectorSupported: false, reason: 'no dataplane' });

    await userEvent.click(screen.getByRole('button', { name: /test master/i }));

    expect(setRole).toHaveBeenCalledWith('test_master');
  });

  it('shows no platform banner when the reflector is supported', () => {
    renderPage();

    expect(screen.queryByTestId('reflector-platform-banner')).toBeNull();
    expect(screen.getByTestId('reflector-start-button')).toBeEnabled();
  });
});

describe('ReflectorPage — run state', () => {
  afterEach(cleanup);

  it('surfaces the reflector error message instead of showing the page as idle', () => {
    renderPage({ stats: { testStatus: 'error', errorMessage: 'bind: address in use' } });

    expect(screen.getByText('bind: address in use')).toBeInTheDocument();
    expect(screen.queryByText('Reflector is idle')).toBeNull();
  });

  it('falls back to a generic headline when an errored run recorded no message', () => {
    renderPage({ stats: { testStatus: 'error' } });

    expect(screen.getByText('The reflector stopped with an error')).toBeInTheDocument();
  });

  it('names the interface it is reflecting on while running', () => {
    renderPage({ stats: { testStatus: 'running' }, selectedInterface: 'eth0' });

    expect(screen.getByText('Reflecting on eth0')).toBeInTheDocument();
  });

  it('distinguishes a cancelled run from an idle one', () => {
    renderPage({ stats: { testStatus: 'cancelled' } });

    expect(screen.getByText('The last reflector run was cancelled')).toBeInTheDocument();
  });

  it('offers Stop, not Start, while the reflector is running', async () => {
    const { onStopReflector } = renderPage({ stats: { testStatus: 'running' } });

    expect(screen.queryByTestId('reflector-start-button')).toBeNull();
    await userEvent.click(screen.getByRole('button', { name: /stop/i }));

    expect(onStopReflector).toHaveBeenCalledTimes(1);
  });

  it('keeps Start disabled until an interface is chosen', () => {
    renderPage({ selectedInterface: '' });

    expect(screen.getByTestId('reflector-start-button')).toBeDisabled();
  });

  it('announces a failed start to assistive technology', () => {
    renderPage({ reflectorStartError: 'permission denied opening eth0' });

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('permission denied opening eth0');
    expect(alert).toHaveAttribute('aria-live', 'assertive');
  });
});

describe('ReflectorPage — counters', () => {
  afterEach(cleanup);

  it.each([
    [999, '999'],
    [1500, '1.50K'],
    [2_500_000, '2.50M'],
    [3_200_000_000, '3.20B'],
  ])('abbreviates %d packets as %s', (received, expected) => {
    renderPage({ stats: { packetsReceived: received } });

    expect(screen.getByTestId('stats-packets-received')).toHaveTextContent(expected);
  });

  it('renders uptime zero-padded so the column does not jitter', () => {
    renderPage({ stats: { uptime: 3661 } });

    expect(screen.getByTestId('stats-uptime')).toHaveTextContent('01:01:01');
  });

  it('shows the selected interface details', () => {
    renderPage();

    expect(screen.getByText('00:11:22:33:44:55')).toBeInTheDocument();
    expect(screen.getByText('ixgbe')).toBeInTheDocument();
    expect(screen.getByText('92')).toBeInTheDocument();
  });

  it('omits the interface card when the selection matches no known interface', () => {
    renderPage({ selectedInterface: 'eth9' });

    expect(screen.queryByText('00:11:22:33:44:55')).toBeNull();
  });
});

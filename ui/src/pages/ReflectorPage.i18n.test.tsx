/**
 * ReflectorPage.i18n.test.tsx — the reflector page renders real locale copy.
 *
 * #779: the page called `t()` for its buttons and status words, but most of
 * what an operator actually reads on it was hardcoded English — the rollup
 * headline and body, the three rollup figure labels, the four stat-card
 * titles, the unit suffixes, and the raw `testStatus` value. `check-source.py`
 * reads JSX text nodes, and none of those are text nodes: they are object
 * literals, component props and template-literal fragments. The detector is
 * documented as false-negatives-only, and this is what that costs.
 *
 * These assertions are deliberately negative as well as positive. Asserting
 * only that Spanish appears would still pass if English were rendered beside
 * it, which is the actual failure mode when a page is half-translated.
 */
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { AppContextValue } from '../contexts/AppContext';
import type { RoleContextValue } from '../contexts/RoleContext';
import type { Capabilities } from '../hooks/useCapabilities';
import i18n from '../i18n';
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

const eth0 = {
  name: 'eth0',
  mac: '00:11:22:33:44:55',
  speed: 10000,
  duplex: 'full',
  driver: 'ixgbe',
  state: 'up',
  xdp: true,
  dpdk: false,
  score: 92,
} as InterfaceInfo;

function renderPage(stats: Partial<Stats> = {}, reflectorSupported = true): void {
  capabilities.current = {
    reflector: reflectorSupported ? { supported: true } : { supported: false, reason: 'sin CGO' },
    testMaster: { supported: true },
  };
  roleContext.current = {
    role: 'reflector',
    setRole: vi.fn(),
    isSwitchingRole: false,
    roleSwitchError: null,
    clearRoleSwitchError: vi.fn(),
  };
  appContext.current = {
    interfaces: [eth0],
    selectedInterface: 'eth0',
    setSelectedInterface: vi.fn(),
    stats: {
      packetsReceived: 0,
      packetsSent: 0,
      bytesReceived: 0,
      bytesSent: 0,
      currentPps: 0,
      currentMbps: 0,
      uptime: 0,
      testStatus: 'idle',
      ...stats,
    } as Stats,
    reflectorProfile: 'netally',
    setReflectorProfile: vi.fn(),
    onStartReflector: vi.fn(),
    onStopReflector: vi.fn(),
    isStartingReflector: false,
    isStoppingReflector: false,
    reflectorStartError: null,
  } as unknown as AppContextValue;

  render(<ReflectorPage />);
}

afterEach(async () => {
  cleanup();
  await i18n.changeLanguage('en');
});

describe('ReflectorPage — real locale copy', () => {
  it('renders the English idle copy, figure labels and stat titles', async () => {
    await i18n.changeLanguage('en');
    renderPage();

    expect(screen.getByText('Reflector is idle')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Pick an interface and start the reflector for a test master to measure against.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByText('Received')).toBeInTheDocument();
    expect(screen.getByText('Packets Received')).toBeInTheDocument();
    expect(screen.getByText('Current Rate')).toBeInTheDocument();
    expect(screen.getByText('Uptime')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    renderPage();

    expect(screen.getByText('El Reflector está inactivo')).toBeInTheDocument();
    expect(screen.getByText('Recibidos')).toBeInTheDocument();
    expect(screen.getByText('Enviados')).toBeInTheDocument();
    expect(screen.getByText('Paquetes Recibidos')).toBeInTheDocument();
    expect(screen.getByText('Paquetes Enviados')).toBeInTheDocument();
    expect(screen.getByText('Tasa Actual')).toBeInTheDocument();
    expect(screen.getByText('Tiempo Activo')).toBeInTheDocument();

    // The exact strings #779 reported as English under es.
    for (const english of [
      'Reflector is idle',
      'Pick an interface and start the reflector for a test master to measure against.',
      'Received',
      'Sent',
      'Packets Received',
      'Packets Sent',
      'Current Rate',
      'Uptime',
    ]) {
      expect(screen.queryByText(english)).toBeNull();
    }
  });

  it('translates the rollup state label, not just the headline under it', async () => {
    await i18n.changeLanguage('es');
    renderPage();

    // StatusRollup's own kicker. It was a module-level literal, so it stayed
    // English while the headline beside it translated.
    expect(screen.getByText('Todo correcto')).toBeInTheDocument();
    expect(screen.queryByText('All clear')).toBeNull();
  });

  it('translates the status value, which was rendered straight from the API', async () => {
    await i18n.changeLanguage('es');
    renderPage({ testStatus: 'cancelled' });

    expect(screen.getByTestId('stats-uptime')).toHaveTextContent('Estado:');
    expect(screen.getByTestId('stats-uptime')).toHaveTextContent('Cancelado');
    expect(screen.getByTestId('stats-uptime')).not.toHaveTextContent('cancelled');
  });

  it('translates the platform-unsupported headline', async () => {
    await i18n.changeLanguage('es');
    renderPage({}, false);

    expect(
      screen.getByText('Los contadores del Reflector no están disponibles en esta plataforma'),
    ).toBeInTheDocument();
    expect(screen.queryByText('Reflector counters are not available on this platform')).toBeNull();
  });

  it('keeps the start control one translatable phrase, not a verb plus English noun', async () => {
    await i18n.changeLanguage('es');
    renderPage();

    // Was `{t('buttons.start')} Reflector`, which pins the word order to
    // English no matter how many locales are added.
    expect(screen.getByTestId('reflector-start-button')).toHaveTextContent('Iniciar Reflector');
  });
});

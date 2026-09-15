/**
 * Every module page must say something useful on a first visit.
 *
 * `selectedTests` defaults to the four RFC 2544 ids, so ServiceTest,
 * TrafficGen, Measure and Certify open with every ConfigForm returning null
 * and the operator facing a title and nothing else (#1257). These assert the
 * page body, not the header AppShell renders for it: with no test of that
 * module selected the page must offer an empty state that opens Settings, and
 * with one selected it must render the form instead.
 */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { defaultRFC2544Config } from '../components/RFC2544ConfigForm';
import { defaultRFC2889Config } from '../components/RFC2889ConfigForm';
import { defaultRFC6349Config } from '../components/RFC6349ConfigForm';
import { defaultTrafficGenConfig } from '../components/TrafficGenConfigForm';
import { defaultTSNConfig } from '../components/TSNConfigForm';
import { defaultY1564Config } from '../components/Y1564ConfigForm';
import { defaultY1731Config } from '../components/Y1731ConfigForm';
import type { AppContextValue } from '../contexts/AppContext';
import type { RoleContextValue } from '../contexts/RoleContext';
import { useShellStore } from '../stores/shell-store';
import { BenchmarkPage } from './BenchmarkPage';
import { CertifyPage } from './CertifyPage';
import { MeasurePage } from './MeasurePage';
import { ServiceTestPage } from './ServiceTestPage';
import { TrafficGenPage } from './TrafficGenPage';

const { appContext, roleContext } = vi.hoisted(() => ({
  appContext: { current: null as unknown as AppContextValue },
  roleContext: { current: null as unknown as RoleContextValue },
}));

vi.mock('../contexts/AppContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../contexts/AppContext')>()),
  useAppContext: () => appContext.current,
}));

vi.mock('../contexts/RoleContext', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../contexts/RoleContext')>()),
  useRole: () => roleContext.current,
}));

function setContext(selectedTests: string[]): void {
  appContext.current = {
    rfc2544Config: defaultRFC2544Config,
    setRFC2544Config: () => {},
    rfc2889Config: defaultRFC2889Config,
    setRFC2889Config: () => {},
    rfc6349Config: defaultRFC6349Config,
    setRFC6349Config: () => {},
    y1564Config: defaultY1564Config,
    setY1564Config: () => {},
    y1731Config: defaultY1731Config,
    setY1731Config: () => {},
    tsnConfig: defaultTSNConfig,
    setTSNConfig: () => {},
    trafficGenConfig: defaultTrafficGenConfig,
    setTrafficGenConfig: () => {},
    selectedTests,
  } as unknown as AppContextValue;
  roleContext.current = {
    role: 'test_master',
    setRole: () => {},
  } as unknown as RoleContextValue;
}

afterEach(() => {
  cleanup();
  useShellStore.getState().setSettingsOpen(false);
});

const pages = [
  { name: 'Benchmark', Page: BenchmarkPage, selected: 'rfc2544_throughput' },
  { name: 'ServiceTest', Page: ServiceTestPage, selected: 'y1564_full' },
  { name: 'TrafficGen', Page: TrafficGenPage, selected: 'custom_stream' },
  { name: 'Measure', Page: MeasurePage, selected: 'y1731_delay' },
  { name: 'Certify', Page: CertifyPage, selected: 'rfc2889_forwarding' },
] as const;

describe.each(pages)('$name page with nothing of its own selected', ({ Page, selected }) => {
  it('renders an empty state instead of an empty body', () => {
    setContext(['reflect']);
    render(<Page />);
    expect(screen.getByTestId('module-empty-state')).toBeInTheDocument();
  });

  it('offers an action that opens Settings', async () => {
    setContext(['reflect']);
    render(<Page />);
    await userEvent.click(screen.getByTestId('module-empty-state-open-settings'));
    expect(useShellStore.getState().settingsOpen).toBe(true);
  });

  it('renders the form, and no empty state, once one of its tests is selected', () => {
    setContext([selected]);
    render(<Page />);
    expect(screen.queryByTestId('module-empty-state')).toBeNull();
  });
});

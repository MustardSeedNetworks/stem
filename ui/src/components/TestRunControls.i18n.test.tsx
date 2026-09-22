/**
 * TestRunControls.i18n.test.tsx — the run controls render real locale copy.
 *
 * Inherits TopBar.i18n.test.tsx's reason, and its assertions, from before the
 * controls moved out of the retired top bar (UI-STEM-9): the rest of the suite
 * renders components without initialising i18n, so a `t()` that resolves to
 * nothing still "passes" — the raw key is a non-empty string and testid
 * assertions never look at it (#654). These import the real i18n instance, the
 * same `internal/i18n/locales` JSON the browser loads, and assert the visible
 * strings in both locales.
 */
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { AppContext, type AppContextValue } from '../contexts/AppContext';
import i18n from '../i18n';
import type { StopOutcome } from '../stores/test-store';
import { initialStats } from '../types/api';
import { TestRunControls } from './TestRunControls';

const role = { current: 'test_master' };

vi.mock('../contexts/RoleContext', async (importOriginal) => ({
  ...(await importOriginal<Record<string, unknown>>()),
  useRole: () => ({ role: role.current }),
}));

const context = {
  interfaces: [],
  selectedInterface: '',
  setSelectedInterface: (): void => undefined,
  peer: '',
  setPeer: (): void => undefined,
  peerPort: 3842,
  setPeerPort: (): void => undefined,
  stats: initialStats,
  isStartingTest: false,
  stopOutcome: { kind: 'idle' } as StopOutcome,
  testStartError: null,
  onStartTest: (): void => undefined,
  onStopTest: (): void => undefined,
  testProgress: {
    status: 'idle' as const,
    currentTest: null,
    expectedDuration: 0,
    startedAt: null,
  },
} as unknown as AppContextValue;

const renderControls = (overrides: Partial<AppContextValue> = {}): void => {
  render(
    <AppContext.Provider value={{ ...context, ...overrides }}>
      <TestRunControls />
    </AppContext.Provider>,
  );
};

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('TestRunControls — renders real locale copy', () => {
  it('labels every control in English', async () => {
    await i18n.changeLanguage('en');
    renderControls();

    expect(screen.getByLabelText('Select network interface')).toBeInTheDocument();
    expect(screen.getByLabelText('Reflector host or IP')).toBeInTheDocument();
    expect(screen.getByLabelText('Reflector UDP port')).toBeInTheDocument();
    expect(screen.getByTestId('start-test-button')).toBeInTheDocument();
  });

  it('labels every control in Spanish, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    renderControls();

    expect(screen.queryByLabelText('Select network interface')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Seleccionar interfaz de red')).toBeInTheDocument();
    expect(screen.getByLabelText('Host o IP del Reflector')).toBeInTheDocument();
    expect(screen.getByLabelText('Puerto UDP del Reflector')).toBeInTheDocument();
  });

  it('interpolates the test name instead of concatenating it', async () => {
    await i18n.changeLanguage('es');
    renderControls({
      stats: { ...initialStats, testStatus: 'running', currentTest: 'RFC 2544' },
    });

    // Interpolated, not "Ejecutando" + ": " + name glued together in JSX —
    // word order round the value is the locale's to decide.
    expect(screen.getByText('Ejecutando: RFC 2544')).toBeInTheDocument();
  });

  it('renders nothing for a Reflector-role stem', async () => {
    await i18n.changeLanguage('en');
    role.current = 'reflector';
    renderControls();

    // RoleGuard on the page already says the role is wrong; a disabled run row
    // underneath would say it a second time.
    expect(screen.queryByTestId('test-run-controls')).not.toBeInTheDocument();
    role.current = 'test_master';
  });
});

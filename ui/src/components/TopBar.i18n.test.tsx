/**
 * TopBar.i18n.test.tsx — asserts the top strip renders real locale copy.
 *
 * The rest of the suite renders components without initialising i18n, so a
 * `t()` call that resolves to nothing still "passes": the raw key is a
 * non-empty string and testid assertions never look at it (#654). These tests
 * import the real i18n instance — the same `internal/i18n/locales` JSON the
 * browser loads — and assert on the visible strings in both locales, so a key
 * that is missing, misspelled or absent from `es` fails here rather than
 * shipping as an English word on a Spanish screen.
 */
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { RoleProvider } from '../contexts/RoleContext';
import i18n from '../i18n';
import { initialStats } from '../types/api';
import { TopBar } from './TopBar';

/** TopBar renders a RoleChip, which requires the role context. */
const renderTopBar = (extra: Partial<typeof props> = {}): void => {
  render(
    <RoleProvider>
      <TopBar {...props} {...extra} />
    </RoleProvider>,
  );
};

const props = {
  connected: true,
  isDark: false,
  onToggleTheme: (): void => undefined,
  onRefresh: (): void => undefined,
  onLogout: (): void => undefined,
  mode: 'test_master' as const,
  selectedInterface: '',
  setSelectedInterface: (): void => undefined,
  interfaces: [],
  stats: initialStats,
  isStartingTest: false,
  isStoppingTest: false,
  testStartError: null,
  onStartTest: (): void => undefined,
  onStopTest: (): void => undefined,
  testProgress: {
    status: 'idle' as const,
    currentTest: null,
    expectedDuration: 0,
    startedAt: null,
  },
};

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('TopBar — renders real locale copy', () => {
  it('shows the English connection state and control labels', async () => {
    await i18n.changeLanguage('en');
    renderTopBar();

    expect(screen.getByText('Connected')).toBeInTheDocument();
    expect(screen.getByLabelText('Refresh interfaces')).toBeInTheDocument();
    expect(screen.getByLabelText('Logout')).toBeInTheDocument();
    expect(screen.getByLabelText('Switch to dark mode')).toBeInTheDocument();
  });

  it('shows Spanish copy under es — no English left on the strip', async () => {
    await i18n.changeLanguage('es');
    renderTopBar({ connected: false });

    expect(screen.getByText('Desconectado')).toBeInTheDocument();
    expect(screen.getByLabelText('Actualizar interfaces')).toBeInTheDocument();
    expect(screen.getByLabelText('Cerrar Sesión')).toBeInTheDocument();
    expect(screen.queryByText('Disconnected')).not.toBeInTheDocument();
  });

  it('interpolates the test name instead of concatenating it', async () => {
    await i18n.changeLanguage('es');
    renderTopBar({
      stats: { ...initialStats, testStatus: 'running', currentTest: 'RFC 2544' },
    });

    // Interpolated, not "Ejecutando" + ": " + name glued together in JSX —
    // word order round the value is the locale's to decide.
    expect(screen.getByText('Ejecutando: RFC 2544')).toBeInTheDocument();
  });

  it('gives the refresh control a descriptive tooltip distinct from its label', async () => {
    await i18n.changeLanguage('en');
    renderTopBar();

    const refresh = screen.getByLabelText('Refresh interfaces');
    expect(refresh).toHaveAttribute(
      'title',
      'Rescan available network interfaces and reload current status',
    );
  });
});

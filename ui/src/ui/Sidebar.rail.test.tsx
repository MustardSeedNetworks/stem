/**
 * Sidebar.rail.test.tsx — the rail carries the shell's chrome, in both locales.
 *
 * The other half of TopBar.i18n.test.tsx's coverage after UI-STEM-9 moved the
 * connection state and the theme / refresh / logout controls into the rail.
 * Same reason it existed: the rest of the suite renders without initialising
 * i18n, so a `t()` resolving to nothing still "passes" (#654).
 *
 * The connection assertions are the substance rather than the copy — the dot
 * was a hard-coded success green, so the rail reported "connected" on a dead
 * socket, and the state has to reach a screen reader through a name because
 * colour alone is not a status.
 */
import { fireEvent, render, screen, within } from '@testing-library/react';
import { Activity, Gauge } from 'lucide-react';
import { MemoryRouter } from 'react-router';
import { afterEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { SidebarLayout } from './Sidebar';

const noop = (): void => undefined;

function renderRail(overrides: Partial<Parameters<typeof SidebarLayout>[0]> = {}, path = '/') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SidebarLayout
        groups={[]}
        status={{ state: 'connected', label: i18n.t('common:status.connected') }}
        onToggleTheme={noop}
        isDark={false}
        onRefresh={noop}
        onLogout={noop}
        {...overrides}
      >
        <div>page</div>
      </SidebarLayout>
    </MemoryRouter>,
  );
}

afterEach(async () => {
  await i18n.changeLanguage('en');
});

describe('the rail', () => {
  it('reports the live connection state, not a hard-coded one', async () => {
    await i18n.changeLanguage('en');
    const { rerender } = renderRail();

    expect(screen.getByTestId('rail-status')).toHaveAttribute('data-status', 'connected');
    // Colour alone is not a status, so the dot carries the name too.
    expect(screen.getByTestId('rail-status')).toHaveAccessibleName('Connected');

    rerender(
      <MemoryRouter>
        <SidebarLayout
          groups={[]}
          status={{ state: 'disconnected', label: i18n.t('common:status.disconnected') }}
          onToggleTheme={noop}
          isDark={false}
          onRefresh={noop}
          onLogout={noop}
        >
          <div>page</div>
        </SidebarLayout>
      </MemoryRouter>,
    );

    expect(screen.getByTestId('rail-status')).toHaveAttribute('data-status', 'disconnected');
    expect(screen.getByTestId('rail-status')).toHaveAccessibleName('Disconnected');
  });

  it('names the theme, refresh and logout controls in English', async () => {
    await i18n.changeLanguage('en');
    renderRail();

    expect(screen.getByTestId('rail-theme-toggle')).toHaveAccessibleName('Switch to dark mode');
    // The control's name and the description of what it does are different
    // strings: "Refresh interfaces" vs the tooltip's fuller sentence.
    expect(screen.getByTestId('rail-refresh')).toHaveAccessibleName('Refresh interfaces');
    expect(screen.getByTestId('rail-refresh')).toHaveAccessibleDescription(
      'Rescan available network interfaces and reload current status',
    );
    expect(screen.getByTestId('rail-logout')).toHaveAccessibleName('Logout');
  });

  it('names them in Spanish, with no English left on the rail', async () => {
    await i18n.changeLanguage('es');
    renderRail({ isDark: true });

    expect(screen.getByTestId('rail-theme-toggle')).toHaveAccessibleName('Cambiar a modo claro');
    expect(screen.getByTestId('rail-refresh')).toHaveAccessibleName('Actualizar interfaces');
    expect(screen.getByTestId('rail-logout')).toHaveAccessibleName('Cerrar Sesión');
    expect(screen.queryByText('Logout')).not.toBeInTheDocument();
  });

  it('renders the role control in the rail footer', async () => {
    await i18n.changeLanguage('en');
    renderRail({ roleControl: <span>role control</span> });

    // Both asides are always mounted — only CSS decides which is shown — so
    // the assertion says which surface it means rather than picking an index.
    const desktop = screen.getByTestId('desktop-sidebar');
    expect(within(desktop).getByText('role control')).toBeInTheDocument();
  });

  it('omits a control whose callback is not supplied', async () => {
    await i18n.changeLanguage('en');
    renderRail({ onLogout: undefined, onRefresh: undefined });

    expect(screen.queryByTestId('rail-logout')).not.toBeInTheDocument();
    expect(screen.queryByTestId('rail-refresh')).not.toBeInTheDocument();
    expect(screen.getByTestId('rail-theme-toggle')).toBeInTheDocument();
  });

  it('moves the current-page mark to the route a rail item opens', () => {
    renderRail(
      {
        groups: [
          {
            label: 'Tests',
            items: [
              { path: '/rfc2544', label: 'RFC 2544', icon: Gauge },
              { path: '/y1564', label: 'Y.1564', icon: Activity, badge: 'Beta' },
            ],
          },
        ],
      },
      '/rfc2544',
    );
    // The desktop and mobile rails both mount; the first is the desktop one.
    const [rfc2544] = screen.getAllByRole('button', { name: 'RFC 2544' });
    const [y1564] = screen.getAllByRole('button', { name: 'Y.1564' });
    expect(rfc2544).toHaveAttribute('aria-current', 'page');
    expect(y1564).not.toHaveAttribute('aria-current');
    expect(within(y1564).getByText('Beta')).toBeInTheDocument();

    fireEvent.click(y1564);

    expect(screen.getAllByRole('button', { name: 'Y.1564' })[0]).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(screen.getAllByRole('button', { name: 'RFC 2544' })[0]).not.toHaveAttribute(
      'aria-current',
    );
  });
});

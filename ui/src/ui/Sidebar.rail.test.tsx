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
const desktopRail = (): HTMLElement => screen.getByTestId('desktop-sidebar');

function renderRail(overrides: Partial<Parameters<typeof SidebarLayout>[0]> = {}, path = '/') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <SidebarLayout
        groups={[]}
        status={{
          state: 'connected',
          label: i18n.t('common:status.connected'),
          detail: i18n.t('common:status.connectedDetail'),
        }}
        onToggleTheme={noop}
        isDark={false}
        onOpenHelp={noop}
        onOpenSettings={noop}
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

    expect(within(desktopRail()).getByTestId('rail-status')).toHaveTextContent('Connected');
    // Colour alone is not a status: the pill says it in words.

    rerender(
      <MemoryRouter>
        <SidebarLayout
          groups={[]}
          status={{
            state: 'disconnected',
            label: i18n.t('common:status.disconnected'),
            detail: i18n.t('common:status.disconnectedDetail'),
          }}
          onToggleTheme={noop}
          isDark={false}
          onOpenHelp={noop}
          onOpenSettings={noop}
          onLogout={noop}
        >
          <div>page</div>
        </SidebarLayout>
      </MemoryRouter>,
    );

    expect(within(desktopRail()).getByTestId('rail-status')).toHaveTextContent('Disconnected');
  });

  it('names the theme and sign-out controls in English', async () => {
    await i18n.changeLanguage('en');
    renderRail();

    const rail = within(desktopRail());
    expect(rail.getByTestId('rail-theme')).toHaveAccessibleName('Switch to dark mode');
    expect(rail.getByTestId('rail-sign-out')).toHaveAccessibleName('Logout');
    // Refresh is a page action now, not a rail control.
    expect(rail.queryByTestId('rail-refresh')).not.toBeInTheDocument();
  });

  it('names them in Spanish, with no English left on the rail', async () => {
    await i18n.changeLanguage('es');
    renderRail({ isDark: true });

    const rail = within(desktopRail());
    expect(rail.getByTestId('rail-theme')).toHaveAccessibleName('Cambiar a modo claro');
    expect(rail.getByTestId('rail-sign-out')).toHaveAccessibleName('Cerrar Sesión');
    expect(rail.queryByText('Logout')).not.toBeInTheDocument();
  });

  it('renders the role control as the footer product slot', async () => {
    await i18n.changeLanguage('en');
    renderRail({
      productSlot: {
        label: 'Role',
        value: 'Reflector',
        icon: <span />,
        options: [{ value: 'reflector', label: 'Reflector' }],
        selected: 'reflector',
        onSelect: noop,
      },
    });

    // Both asides are always mounted — only CSS decides which is shown — so
    // the assertion says which surface it means rather than picking an index.
    expect(within(desktopRail()).getByTestId('rail-product-slot')).toHaveTextContent('Reflector');
  });

  it('omits sign-out when no handler is supplied', async () => {
    await i18n.changeLanguage('en');
    renderRail({ onLogout: undefined });

    expect(within(desktopRail()).queryByTestId('rail-sign-out')).not.toBeInTheDocument();
    expect(within(desktopRail()).getByTestId('rail-theme')).toBeInTheDocument();
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
    const [y1564] = screen.getAllByRole('button', { name: 'Y.1564 Beta' });
    expect(rfc2544).toHaveAttribute('aria-current', 'page');
    expect(y1564).not.toHaveAttribute('aria-current');
    expect(within(y1564).getByText('Beta')).toBeInTheDocument();

    fireEvent.click(y1564);

    expect(screen.getAllByRole('button', { name: 'Y.1564 Beta' })[0]).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(screen.getAllByRole('button', { name: 'RFC 2544' })[0]).not.toHaveAttribute(
      'aria-current',
    );
  });
});

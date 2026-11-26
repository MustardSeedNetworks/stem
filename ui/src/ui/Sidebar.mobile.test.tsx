/**
 * Sidebar.mobile.test.tsx — below lg the rail is replaced by a top bar carrying
 * the product mark, the product name and the menu button, as in seed and niac.
 * The menu button opens the navigation drawer.
 */
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Gauge } from 'lucide-react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';
import { SidebarLayout } from './Sidebar';

function renderLayout() {
  return render(
    <MemoryRouter>
      <SidebarLayout
        groups={[{ label: '', items: [{ path: '/reflector', label: 'Reflector', icon: Gauge }] }]}
        onOpenHelp={() => undefined}
        onOpenSettings={() => undefined}
        onToggleTheme={() => undefined}
        isDark={false}
      >
        <div>page</div>
      </SidebarLayout>
    </MemoryRouter>,
  );
}

describe('the mobile top bar', () => {
  it('carries the mark, the product name and the menu button', () => {
    renderLayout();
    const bar = screen.getByRole('banner');
    expect(within(bar).getByText('ST')).toBeInTheDocument();
    expect(within(bar).getByText('Stem')).toBeInTheDocument();
    expect(within(bar).getByTestId('mobile-nav-toggle')).toHaveAccessibleName('Open menu');
  });

  it('opens and closes the navigation drawer from the menu button', async () => {
    const user = userEvent.setup();
    renderLayout();
    const drawer = screen.getByTestId('mobile-sidebar');
    expect(drawer).toHaveClass('invisible');

    await user.click(screen.getByTestId('mobile-nav-toggle'));
    expect(drawer).not.toHaveClass('invisible');
    expect(screen.getByTestId('mobile-nav-toggle')).toHaveAccessibleName('Close menu');

    await user.click(screen.getByTestId('mobile-nav-toggle'));
    expect(drawer).toHaveClass('invisible');
  });
});

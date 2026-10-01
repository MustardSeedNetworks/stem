/**
 * CommandPalette tests — every entry does what it names and closes the palette.
 *
 * The palette is the keyboard route to every page and to the shell actions.
 * An entry that runs but leaves the dialog open, or closes without running,
 * looks fine in a screenshot, so these click each kind and check both halves.
 */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Gauge } from 'lucide-react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { SidebarNavGroup } from '../../ui/Sidebar';
import { CommandPalette, type CommandPaletteProps } from './CommandPalette';

const groups: SidebarNavGroup[] = [
  { label: 'Test', items: [{ path: '/benchmark', label: 'RFC 2544', icon: Gauge }] },
];

function Location(): React.JSX.Element {
  return <output data-testid="location">{useLocation().pathname}</output>;
}

function renderPalette(props: Partial<CommandPaletteProps> = {}) {
  const onOpenChange = vi.fn();
  render(
    <MemoryRouter initialEntries={['/']}>
      <CommandPalette groups={groups} open={true} onOpenChange={onOpenChange} {...props} />
      <Routes>
        <Route path="*" element={<Location />} />
      </Routes>
    </MemoryRouter>,
  );
  return { onOpenChange };
}

afterEach(cleanup);

describe('CommandPalette', () => {
  it('jumps to a page and closes', async () => {
    const { onOpenChange } = renderPalette();

    await userEvent.click(screen.getByRole('option', { name: /RFC 2544/ }));

    expect(screen.getByTestId('location')).toHaveTextContent('/benchmark');
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it.each([
    ['onOpenSettings', 'Open settings'],
    ['onOpenHelp', 'Open help'],
    ['onToggleTheme', 'Switch to dark mode'],
  ] as const)('runs %s and closes', async (prop, name) => {
    const perform = vi.fn();
    const { onOpenChange } = renderPalette({ [prop]: perform });

    await userEvent.click(screen.getByRole('option', { name }));

    expect(perform).toHaveBeenCalledTimes(1);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('offers light mode when the theme is dark', () => {
    renderPalette({ onToggleTheme: vi.fn(), isDark: true });

    expect(screen.getByRole('option', { name: 'Switch to light mode' })).toBeInTheDocument();
    expect(screen.queryByRole('option', { name: 'Switch to dark mode' })).toBeNull();
  });

  it('runs an extra action and shows its hint', async () => {
    const perform = vi.fn();
    const { onOpenChange } = renderPalette({
      extraActions: [{ id: 'stop', label: 'Stop test', hint: 'Esc', perform }],
    });

    const option = screen.getByRole('option', { name: /Stop test/ });
    expect(option).toHaveTextContent('Esc');
    await userEvent.click(option);

    expect(perform).toHaveBeenCalledTimes(1);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('closes from the backdrop without running anything', async () => {
    const perform = vi.fn();
    const { onOpenChange } = renderPalette({ onOpenSettings: perform });

    await userEvent.click(screen.getByRole('button', { name: 'Close command palette' }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(perform).not.toHaveBeenCalled();
  });
});

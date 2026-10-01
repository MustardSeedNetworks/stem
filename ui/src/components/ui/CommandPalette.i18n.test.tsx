/**
 * CommandPalette.i18n.test.tsx — the palette's own name and group heading
 * come from the locale.
 *
 * Both were props on cmdk components (`label`, `heading`), which the JSX-text
 * check cannot see, so a Spanish operator opening Ctrl+K got an English
 * dialog name and an English "Actions" heading (#1498).
 */
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { afterEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { CommandPalette } from './CommandPalette';

afterEach(async () => {
  cleanup();
  await i18n.changeLanguage('en');
});

describe('CommandPalette — real locale copy', () => {
  it('names the dialog and the actions group in Spanish, with no English left', async () => {
    await i18n.changeLanguage('es');
    render(
      <MemoryRouter>
        <CommandPalette groups={[]} open={true} onOpenChange={vi.fn()} onOpenSettings={vi.fn()} />
      </MemoryRouter>,
    );

    expect(screen.getByRole('dialog', { name: 'Paleta de comandos' })).toBeInTheDocument();
    expect(screen.getByText('Acciones')).toBeInTheDocument();
    expect(screen.queryByRole('dialog', { name: 'Command palette' })).toBeNull();
    expect(screen.queryByText('Actions')).toBeNull();
  });
});

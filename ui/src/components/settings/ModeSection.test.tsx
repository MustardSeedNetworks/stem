/**
 * ModeSection tests.
 *
 * The section that decides whether this stem reflects packets or runs tests.
 * The locale assertions are the load-bearing ones: every option here used to
 * render English under a translated heading, because the keys were addressed
 * with a dot instead of a namespace colon and the hardcoded fallbacks made
 * that silent.
 */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { ModeSection } from './ModeSection';

describe('ModeSection', () => {
  afterEach(async () => {
    cleanup();
    await i18n.changeLanguage('en');
  });

  it('marks the active mode and only that one', () => {
    render(<ModeSection mode="reflector" onModeChange={vi.fn()} />);

    const radios = screen.getAllByRole('radio') as HTMLInputElement[];
    expect(radios).toHaveLength(2);
    expect(radios.filter((r) => r.checked)).toHaveLength(1);
    expect(screen.getByRole('radio', { name: /reflector/i })).toBeChecked();
  });

  it('reports the mode the operator picked', async () => {
    const onModeChange = vi.fn();
    render(<ModeSection mode="reflector" onModeChange={onModeChange} />);

    await userEvent.click(screen.getByRole('radio', { name: /test master/i }));

    expect(onModeChange).toHaveBeenCalledWith('test_master');
  });

  it('does not report a change when the current mode is re-selected', async () => {
    const onModeChange = vi.fn();
    render(<ModeSection mode="reflector" onModeChange={onModeChange} />);

    await userEvent.click(screen.getByRole('radio', { name: /reflector/i }));

    // A radio that is already checked fires no change event; asserting it
    // keeps a future refactor from turning re-selection into a real switch.
    expect(onModeChange).not.toHaveBeenCalled();
  });

  it('renders every label from the locale, not from a hardcoded default', async () => {
    await i18n.changeLanguage('es');
    render(<ModeSection mode="reflector" onModeChange={vi.fn()} />);

    // These are the Spanish strings that already existed in the locale file
    // and were never reached. Before the namespace fix this section rendered
    // "Reflector Mode" and "Packet reflection (Tier 1)" under a Spanish
    // heading — English options in a translated UI.
    expect(screen.getByText('Modo Reflector')).toBeInTheDocument();
    expect(screen.getByText('Modo Test Master')).toBeInTheDocument();
    expect(screen.queryByText('Reflector Mode')).toBeNull();
    expect(screen.queryByText(/Packet reflection/)).toBeNull();
  });

  it('resolves its keys rather than echoing them', () => {
    render(<ModeSection mode="reflector" onModeChange={vi.fn()} />);

    // An unresolved i18next key renders as the key itself, which looks like
    // copy to a screenshot and like nothing to a reader.
    expect(screen.queryByText(/^settings[:.]mode\./)).toBeNull();
  });
});

import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { ToastProvider } from './Toast';
import { useToast } from './useToast';

function Trigger({
  message,
  type,
  duration,
}: {
  message: string;
  type?: 'error';
  duration?: number;
}) {
  const { addToast } = useToast();
  return (
    <button type="button" onClick={() => addToast(message, type, duration)}>
      fire
    </button>
  );
}

beforeEach(async () => {
  await i18n.changeLanguage('en');
});

afterEach(() => {
  vi.useRealTimers();
});

describe('Toast', () => {
  it('shows a toast in a labelled live region and stacks several', async () => {
    const user = userEvent.setup();
    render(
      <ToastProvider>
        <Trigger message="Start failed" type="error" />
      </ToastProvider>,
    );

    await user.click(screen.getByRole('button', { name: 'fire' }));
    await user.click(screen.getByRole('button', { name: 'fire' }));

    expect(screen.getByRole('region', { name: 'Notifications' })).toHaveAttribute(
      'aria-live',
      'polite',
    );
    expect(screen.getAllByTestId('toast')).toHaveLength(2);
    expect(screen.getAllByText('Start failed')).toHaveLength(2);
  });

  it('labels the region in Spanish under es', async () => {
    await i18n.changeLanguage('es');
    render(
      <ToastProvider>
        <Trigger message="x" />
      </ToastProvider>,
    );
    expect(screen.getByRole('region', { name: 'Notificaciones' })).toBeInTheDocument();
  });

  it('dismisses a toast from its close button', async () => {
    const user = userEvent.setup();
    render(
      <ToastProvider>
        <Trigger message="Saved" />
      </ToastProvider>,
    );
    await user.click(screen.getByRole('button', { name: 'fire' }));
    await user.click(screen.getByRole('button', { name: 'Dismiss alert' }));
    expect(screen.queryByTestId('toast')).not.toBeInTheDocument();
  });

  it('clears itself after its duration, and a zero duration stays', () => {
    vi.useFakeTimers();
    render(
      <ToastProvider>
        <Trigger message="short" duration={1000} />
        <Trigger message="sticky" duration={0} />
      </ToastProvider>,
    );
    const [short, sticky] = screen.getAllByRole('button', { name: 'fire' });
    act(() => {
      short?.click();
      sticky?.click();
    });
    expect(screen.getAllByTestId('toast')).toHaveLength(2);

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.queryByText('short')).not.toBeInTheDocument();
    expect(screen.getByText('sticky')).toBeInTheDocument();
  });

  it('refuses to be used without a provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    expect(() => render(<Trigger message="x" />)).toThrow(
      'useToast must be used within a ToastProvider',
    );
    spy.mockRestore();
  });
});

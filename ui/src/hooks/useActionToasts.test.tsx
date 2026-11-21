import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ToastProvider } from '../components/ui/Toast';
import type { StopOutcome } from '../stores/test-store';
import { useActionToasts } from './useActionToasts';

function Harness({ startError, stop }: { startError: string | null; stop: StopOutcome }) {
  useActionToasts(startError, stop);
  return null;
}

function renderHarness(startError: string | null, stop: StopOutcome) {
  return render(
    <ToastProvider>
      <Harness startError={startError} stop={stop} />
    </ToastProvider>,
  );
}

describe('useActionToasts', () => {
  it('stays quiet while nothing has failed', () => {
    renderHarness(null, { kind: 'idle' });
    expect(screen.queryByTestId('toast')).not.toBeInTheDocument();
  });

  it('toasts a failed start', () => {
    renderHarness('Daemon refused the run', { kind: 'idle' });
    expect(screen.getByTestId('toast')).toHaveTextContent('Daemon refused the run');
  });

  it.each([
    ['stopRejected', 'No test is currently running'],
    ['stopFailed', 'Failed to stop the test'],
  ] as const)('toasts a failed stop (%s)', (kind, message) => {
    renderHarness(null, { kind, message });
    expect(screen.getByTestId('toast')).toHaveTextContent(message);
  });

  it('does not toast a clean stop', () => {
    renderHarness(null, { kind: 'stopped' });
    expect(screen.queryByTestId('toast')).not.toBeInTheDocument();
  });
});

/**
 * TestProgressBar tests.
 *
 * This is what an operator watches while a test runs, so the interesting
 * questions are whether it tells the truth when the backend is quiet and
 * whether it disappears when it should. A progress bar that keeps advancing
 * on its own is worse than none: it reports progress the run is not making.
 */
import { act, cleanup, render, renderHook, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { type TestProgress, TestProgressBar, useTestProgress } from './TestProgressBar';

const START = 1_700_000_000_000;

function progress(overrides: Partial<TestProgress> = {}): TestProgress {
  return {
    status: 'running',
    currentTest: 'RFC 2544 throughput',
    expectedDuration: 120,
    startedAt: START,
    ...overrides,
  };
}

describe('TestProgressBar — when it shows at all', () => {
  afterEach(cleanup);

  it('renders nothing while idle', () => {
    const { container } = render(<TestProgressBar progress={progress({ status: 'idle' })} />);

    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing when no test is named, whatever the status says', () => {
    // A running status with no test is a state the backend should not produce;
    // rendering a nameless bar for it would be a progress display for nothing.
    const { container } = render(<TestProgressBar progress={progress({ currentTest: null })} />);

    expect(container).toBeEmptyDOMElement();
  });

  it('names the running test and its status', () => {
    render(<TestProgressBar progress={progress()} />);

    expect(screen.getByText('RFC 2544 throughput')).toBeInTheDocument();
    expect(screen.getByText('(Running)')).toBeInTheDocument();
  });

  it.each([
    ['starting', 'Starting...'],
    ['completed', 'Completed'],
    ['cancelled', 'Cancelled'],
    ['error', 'Error'],
  ] as const)('labels the %s status as %s', (status, label) => {
    render(<TestProgressBar progress={progress({ status })} />);

    expect(screen.getByText(`(${label})`)).toBeInTheDocument();
  });
});

describe('TestProgressBar — elapsed time and progress', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(START);
  });

  afterEach(() => {
    vi.useRealTimers();
    cleanup();
  });

  it('counts elapsed time up while the test runs', () => {
    render(<TestProgressBar progress={progress()} />);
    expect(screen.getByText('Elapsed: 0:00')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(65_000);
    });

    expect(screen.getByText('Elapsed: 1:05')).toBeInTheDocument();
  });

  it('stops counting once the test is no longer running', () => {
    const { rerender } = render(<TestProgressBar progress={progress()} />);
    act(() => {
      vi.advanceTimersByTime(30_000);
    });
    expect(screen.getByText('Elapsed: 0:30')).toBeInTheDocument();

    rerender(<TestProgressBar progress={progress({ status: 'completed' })} />);
    act(() => {
      vi.advanceTimersByTime(30_000);
    });

    // A bar that keeps ticking after the run ended reports time the test did
    // not take.
    expect(screen.getByText('Elapsed: 0:00')).toBeInTheDocument();
  });

  it('prefers the percentage the backend reported over its own estimate', () => {
    // The clock says 50% of the expected duration has passed; the backend says
    // 10%. The backend knows what it has actually done.
    render(<TestProgressBar progress={progress({ progressPercent: 10 })} />);
    act(() => {
      vi.advanceTimersByTime(60_000);
    });

    expect(screen.getByText('10%')).toBeInTheDocument();
    expect(screen.queryByText('50%')).toBeNull();
  });

  it('never claims more than 100% on an overrunning test', () => {
    render(<TestProgressBar progress={progress({ expectedDuration: 10 })} />);

    act(() => {
      vi.advanceTimersByTime(60_000);
    });

    expect(screen.getByText('100%')).toBeInTheDocument();
  });

  it('reports no ETA rather than a negative one once the estimate is passed', () => {
    render(<TestProgressBar progress={progress({ expectedDuration: 10 })} />);

    act(() => {
      vi.advanceTimersByTime(60_000);
    });

    expect(screen.getByText('ETA: Complete')).toBeInTheDocument();
  });

  // The switch is on time REMAINING, not elapsed: with a 120s estimate, 30s in
  // leaves 90s (minutes) and 90s in leaves 30s (seconds).
  it.each([
    [30, 'ETA: ~2m'],
    [90, 'ETA: ~30s'],
  ])('shows minutes while far out and seconds near the end (+%ds)', (advanceSeconds, expected) => {
    render(<TestProgressBar progress={progress()} />);

    act(() => {
      vi.advanceTimersByTime(advanceSeconds * 1000);
    });

    expect(screen.getByText(expected)).toBeInTheDocument();
  });

  it('shows no ETA for a finished test', () => {
    render(<TestProgressBar progress={progress({ status: 'completed' })} />);

    expect(screen.queryByText(/ETA:/)).toBeNull();
  });

  it('shows the step the backend reported', () => {
    render(<TestProgressBar progress={progress({ currentStep: '3 of 7 frame sizes' })} />);

    expect(screen.getByText('3 of 7 frame sizes')).toBeInTheDocument();
  });
});

describe('useTestProgress', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(START);
  });

  afterEach(() => {
    vi.useRealTimers();
    cleanup();
  });

  it('stamps a start time when the test starts', () => {
    const { result } = renderHook(() => useTestProgress('running', 'RFC 2544', 120));

    expect(result.current.startedAt).toBe(START);
  });

  it('keeps the original start time across the starting to running transition', () => {
    const { result, rerender } = renderHook(
      ({ status }: { status: 'idle' | 'starting' | 'running' }) =>
        useTestProgress(status, 'RFC 2544', 120),
      { initialProps: { status: 'starting' as const } },
    );
    const first = result.current.startedAt;

    vi.setSystemTime(START + 5_000);
    rerender({ status: 'running' });

    // Re-stamping here would reset elapsed time to zero mid-run, exactly when
    // the operator is watching it.
    expect(result.current.startedAt).toBe(first);
  });

  it('clears the start time when the test is no longer active', () => {
    const { result, rerender } = renderHook(
      ({ status }: { status: 'running' | 'completed' }) => useTestProgress(status, 'RFC 2544', 120),
      { initialProps: { status: 'running' as const } },
    );
    expect(result.current.startedAt).toBe(START);

    rerender({ status: 'completed' });

    expect(result.current.startedAt).toBeNull();
  });
});

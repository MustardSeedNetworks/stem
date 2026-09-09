import { cleanup, render, renderHook, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { initialStats } from '../types/api';
import { type TestProgress, TestProgressBar, useTestProgress } from './TestProgressBar';

function progress(overrides: Partial<TestProgress> = {}): TestProgress {
  return {
    status: 'running',
    currentTest: 'Y.1731 delay',
    currentStep: 1,
    stepsTotal: 2,
    phase: 'Executing y1731_delay',
    elapsedSeconds: 15,
    estimatedRemainingSeconds: null,
    steps: [
      { testType: 'rfc2544_throughput', module: 'benchmark', status: 'passed' },
      { testType: 'y1731_delay', module: 'measure', status: 'running' },
    ],
    ...overrides,
  };
}

describe('TestProgressBar', () => {
  afterEach(cleanup);

  it('renders nothing while idle', () => {
    const { container } = render(<TestProgressBar progress={progress({ status: 'idle' })} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('shows server-owned step and phase text', () => {
    render(<TestProgressBar progress={progress({ currentStep: 2 })} />);
    expect(screen.getByText(/Step 2 of 2/)).toHaveTextContent(
      'Step 2 of 2 · Executing y1731_delay',
    );
    expect(screen.getByText('Elapsed: 0:15')).toBeInTheDocument();
    expect(screen.getByText('rfc2544_throughput')).toBeInTheDocument();
    expect(screen.getByText('passed')).toBeInTheDocument();
  });

  // The bar carries role="progressbar", and axe fails a progressbar with no
  // accessible name (aria-progressbar-name). Both bar shapes are checked
  // because only one of them renders at a time.
  it.each([[null], [90]])('names the progressbar for estimate %s', (estimate) => {
    render(<TestProgressBar progress={progress({ estimatedRemainingSeconds: estimate })} />);
    expect(screen.getByRole('progressbar')).toHaveAccessibleName('Run plan progress');
  });

  it('uses an indeterminate bar when the module provides no estimate', () => {
    render(<TestProgressBar progress={progress()} />);
    expect(screen.getByTestId('indeterminate-progress')).toBeInTheDocument();
    expect(screen.queryByText(/ETA:/)).toBeNull();
  });

  it('uses a determinate bar when the server provides an estimate', () => {
    render(
      <TestProgressBar
        progress={progress({ elapsedSeconds: 30, estimatedRemainingSeconds: 90 })}
      />,
    );
    expect(screen.getByTestId('determinate-progress')).toHaveStyle({ width: '25%' });
    expect(screen.getByText('ETA: ~2m')).toBeInTheDocument();
    expect(screen.getByText('25%')).toBeInTheDocument();
  });

  it.each([
    ['starting', 'Starting...'],
    ['completed', 'Completed'],
    ['cancelled', 'Cancelled'],
    ['error', 'Error'],
  ] as const)('labels the %s state', (status, label) => {
    render(<TestProgressBar progress={progress({ status })} />);
    expect(screen.getByText(`(${label})`)).toBeInTheDocument();
  });
});

describe('useTestProgress', () => {
  it('maps the run-plan fields without deriving client-side progress', () => {
    const stats = {
      ...initialStats,
      testStatus: 'running' as const,
      currentTest: 'rfc2544_throughput',
      currentStep: 2,
      stepsTotal: 3,
      phase: 'Executing rfc2544_throughput',
      elapsedSeconds: 10,
      estimatedRemainingSeconds: 20,
      steps: [],
    };
    const { result } = renderHook(() => useTestProgress(stats));
    expect(result.current).toEqual({
      status: 'running',
      currentTest: 'rfc2544_throughput',
      currentStep: 2,
      stepsTotal: 3,
      phase: 'Executing rfc2544_throughput',
      elapsedSeconds: 10,
      estimatedRemainingSeconds: 20,
      steps: [],
    });
  });
});

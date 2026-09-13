import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import type { TestResult } from '../types/api';
import { TestResults } from './TestResults';

/**
 * TestResults — the panel that has to say why a run produced nothing (#1081).
 *
 * `fetchTestResult` used to return on a non-2xx and log the failure as
 * "non-critical", so a daemon that could not produce a result left the panel
 * showing the same placeholder as a machine with no run at all. It now builds
 * a visible failed result — `Result unavailable (HTTP 500)` — and a stats-side
 * daemon error is carried into `result.error` when the result body omits it.
 *
 * Nothing rendered that. This component had no test file, so the last step of
 * the path — the sentence reaching the screen — was the untested one.
 */
function result(overrides: Partial<TestResult> = {}): TestResult {
  return {
    testType: 'rfc2544_throughput',
    module: 'benchmark',
    status: 'completed',
    success: true,
    ...overrides,
  };
}

describe('TestResults', () => {
  afterEach(cleanup);

  it('renders a result the daemon could not return, with its status', () => {
    render(
      <TestResults
        testStatus="error"
        result={result({
          testType: 'run_plan',
          module: 'orchestrator',
          status: 'error',
          success: false,
          error: 'Result unavailable (HTTP 500)',
        })}
      />,
    );

    expect(screen.getByText('Result unavailable (HTTP 500)')).toBeInTheDocument();
    expect(screen.getByText('FAILED')).toBeInTheDocument();
    // Not the "no result yet" placeholder, which is what a swallowed failure
    // used to leave behind.
    expect(screen.queryByText(/An error occurred during the test/)).toBeNull();
  });

  it('renders the daemon error carried over from stats', () => {
    render(
      <TestResults
        testStatus="error"
        result={result({ status: 'error', success: false, error: 'Latency measurement failed' })}
      />,
    );

    expect(screen.getByText('Latency measurement failed')).toBeInTheDocument();
  });

  it('renders a per-step error inside the run plan', () => {
    render(
      <TestResults
        testStatus="error"
        result={result({
          testType: 'run_plan',
          module: 'orchestrator',
          status: 'error',
          success: false,
          steps: [
            { testType: 'rfc2544_throughput', module: 'benchmark', status: 'passed' },
            {
              testType: 'rfc2544_latency',
              module: 'benchmark',
              status: 'failed',
              error: 'no reply from peer',
            },
          ],
        })}
      />,
    );

    expect(screen.getByText('no reply from peer')).toBeInTheDocument();
  });

  it('renders metrics, formatting large numbers and leaving text alone', () => {
    render(
      <TestResults
        testStatus="completed"
        result={result({
          duration: 92_000,
          startedAt: '2026-09-12T23:00:00Z',
          completedAt: '2026-09-12T23:01:32Z',
          metrics: {
            frames_sent: 1_500_000_000,
            throughput_mbps: 9410,
            verdict: 'pass',
          },
        })}
      />,
    );

    expect(screen.getByText('1.50B')).toBeInTheDocument();
    expect(screen.getByText('9.41K')).toBeInTheDocument();
    expect(screen.getByText('pass')).toBeInTheDocument();
    // The key is the operator-facing label, so underscores are not.
    expect(screen.getByText('frames sent')).toBeInTheDocument();
    expect(screen.getByText(/1m 32s/)).toBeInTheDocument();
    expect(screen.getByText('PASSED')).toBeInTheDocument();
    expect(screen.getByText(/^Started:/)).toBeInTheDocument();
    expect(screen.getByText(/^Completed:/)).toBeInTheDocument();
  });

  it('formats a sub-minute duration in seconds', () => {
    render(<TestResults testStatus="completed" result={result({ duration: 4300 })} />);

    expect(screen.getByText(/4\.3s/)).toBeInTheDocument();
  });

  // Each waiting state says something different, because "nothing here yet"
  // and "the run failed" are not the same message to an operator.
  it.each([
    ['idle', /No tests running/],
    ['starting', /Test is starting/],
    ['running', /Test in progress/],
    ['cancelled', /Test cancelled/],
    ['error', /An error occurred during the test/],
  ] as const)('says what is happening while %s and no result exists', (status, expected) => {
    render(<TestResults testStatus={status} result={null} />);

    expect(screen.getByText(expected)).toBeInTheDocument();
  });
});

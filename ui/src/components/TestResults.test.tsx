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

  it('says nothing has run rather than showing an empty panel', () => {
    render(<TestResults testStatus="idle" result={null} />);

    expect(screen.getByText(/No tests running/)).toBeInTheDocument();
  });
});

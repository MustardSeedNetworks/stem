import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import type { TestResult } from '../types/api';
import { hostLimitedStallMs, TestResults } from './TestResults';

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

  // A reflector's stored result carries the start acknowledgement in
  // `success`, so reading it printed PASSED over a run the operator had
  // stopped (#1248). A stopped run has no verdict.
  it('says a stopped run was stopped rather than claiming it passed', () => {
    render(
      <TestResults
        testStatus="stopped"
        result={result({
          testType: 'reflect',
          module: 'reflector',
          status: 'stopped',
          success: true,
        })}
      />,
    );

    expect(screen.getByText('STOPPED')).toBeInTheDocument();
    expect(screen.queryByText('PASSED')).toBeNull();
    expect(screen.queryByText(/No tests running/)).toBeNull();
  });

  // Each waiting state says something different, because "nothing here yet"
  // and "the run failed" are not the same message to an operator.
  it.each([
    ['idle', /No tests running/],
    ['starting', /Test is starting/],
    ['running', /Test in progress/],
    ['cancelled', /Test cancelled/],
    ['stopped', /Stopped by the operator/],
    ['error', /An error occurred during the test/],
  ] as const)('says what is happening while %s and no result exists', (status, expected) => {
    render(<TestResults testStatus={status} result={null} />);

    expect(screen.getByText(expected)).toBeInTheDocument();
  });

  // A Y.1564 reply is timestamped when it is polled, so a stalled test host
  // adds delay the path never did (#1717). The daemon flags it; the card has
  // to say so, or FAILED over a busy host reads as a bad circuit.
  describe('test-host stall', () => {
    const step = (hostLimited: boolean, stallMs: number) => ({
      Step: 1,
      FDAvgMs: 4.1,
      HostStallMaxMs: stallMs,
      HostLimited: hostLimited,
    });

    it.each([
      [
        'a config result',
        { ServiceID: 1, Steps: [step(false, 0.4), step(true, 17.25), step(true, 9)] },
        17.25,
      ],
      ['a performance result', { ServiceID: 1, HostStallMaxMs: 31, HostLimited: true }, 31],
      [
        'a combined result',
        {
          config: { Steps: [step(true, 12)] },
          performance: { HostStallMaxMs: 20.5, HostLimited: true },
        },
        20.5,
      ],
      [
        'a clean result',
        {
          config: { Steps: [step(false, 2)] },
          performance: { HostStallMaxMs: 3, HostLimited: false },
        },
        null,
      ],
      ['a result without the fields', { throughputMbps: 940 }, null],
      ['no data', undefined, null],
    ] as const)('finds the longest host-limited stall in %s', (_name, data, expected) => {
      expect(hostLimitedStallMs(data)).toBe(expected);
    });

    it('warns on a host-limited result and leaves the verdict alone', () => {
      render(
        <TestResults
          testStatus="completed"
          result={result({
            testType: 'y1564_perf',
            module: 'certify',
            success: false,
            data: { HostStallMaxMs: 31, HostLimited: true },
          })}
        />,
      );

      expect(screen.getByTestId('host-limited-warning')).toHaveTextContent(
        'The test host stalled for up to 31.0 ms',
      );
      expect(screen.getByText('FAILED')).toBeInTheDocument();
    });

    it('warns inside the run-plan step that was host-limited', () => {
      render(
        <TestResults
          testStatus="completed"
          result={result({
            testType: 'run_plan',
            module: 'orchestrator',
            steps: [
              { testType: 'rfc2544_throughput', module: 'benchmark', status: 'passed' },
              {
                testType: 'y1564_config',
                module: 'certify',
                status: 'failed',
                result: { status: 'completed', data: { Steps: [step(true, 8)] } },
              },
            ],
          })}
        />,
      );

      const warnings = screen.getAllByTestId('host-limited-warning');
      expect(warnings).toHaveLength(1);
      expect(screen.getByTestId('run-plan-result-1')).toContainElement(warnings[0] ?? null);
    });

    it('says nothing when the host kept up', () => {
      render(
        <TestResults
          testStatus="completed"
          result={result({
            testType: 'y1564_perf',
            data: { HostStallMaxMs: 3, HostLimited: false },
          })}
        />,
      );

      expect(screen.queryByTestId('host-limited-warning')).toBeNull();
    });
  });
});

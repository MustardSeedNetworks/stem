import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import type { TestResult } from '../types/api';
import { useHistoryStore, useRecordTestResult } from './history-store';

/**
 * A run reaches the History page because it ended (#1333).
 *
 * The recorder this replaces gated on `result.completedAt`, a field the daemon
 * never sent, so nothing was ever recorded and the page was permanently empty.
 * The daemon now stamps the run's own clock, and recording is driven by the
 * run reaching a terminal state — not by a browser noticing a timestamp.
 */

const completed: TestResult = {
  status: 'completed',
  testType: 'rfc2544_throughput',
  module: 'benchmark',
  success: true,
  startedAt: '2026-09-18T10:00:00.000Z',
  completedAt: '2026-09-18T10:02:30.000Z',
  duration: 150000,
  data: { throughputMbps: 942.4, frameSize: 1518, verdict: 'pass' },
};

describe('useRecordTestResult', () => {
  beforeEach(() => {
    window.localStorage.clear();
    useHistoryStore.setState({ results: [], lastRecorded: null });
  });

  it('records a finished run with the daemon’s own timing', () => {
    renderHook(() => useRecordTestResult(completed, 'completed'));

    const [recorded] = useHistoryStore.getState().results;
    expect(recorded).toMatchObject({
      testType: 'rfc2544_throughput',
      module: 'benchmark',
      status: 'completed',
      success: true,
      startedAt: '2026-09-18T10:00:00.000Z',
      completedAt: '2026-09-18T10:02:30.000Z',
      duration: 150000,
    });
  });

  it('carries the run’s scalar measurements into the metrics panel', () => {
    renderHook(() => useRecordTestResult(completed, 'completed'));

    expect(useHistoryStore.getState().results[0]?.metrics).toEqual({
      throughputMbps: 942.4,
      frameSize: 1518,
      verdict: 'pass',
    });
  });

  it('records a stopped run: the operator ended it, so it happened', () => {
    renderHook(() =>
      useRecordTestResult(
        { ...completed, status: 'stopped', success: undefined },
        'stopped',
      ),
    );

    expect(useHistoryStore.getState().results).toHaveLength(1);
    expect(useHistoryStore.getState().results[0]?.status).toBe('stopped');
  });

  it('does not record a run still in flight', () => {
    renderHook(() =>
      useRecordTestResult(
        { status: 'running', testType: 'rfc2544_throughput', startedAt: completed.startedAt },
        'running',
      ),
    );

    expect(useHistoryStore.getState().results).toHaveLength(0);
  });

  it('records the finished run once however often the shell re-renders', () => {
    const { rerender } = renderHook(() => useRecordTestResult(completed, 'completed'));
    act(() => {
      rerender();
      rerender();
    });

    expect(useHistoryStore.getState().results).toHaveLength(1);
  });
});

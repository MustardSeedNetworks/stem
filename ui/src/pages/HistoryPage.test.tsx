/**
 * The archetype's honest-state rule on stem's runs: `success` is optional on
 * the wire, and a run that recorded no verdict must not be shown as a pass.
 */
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { useHistoryStore } from '../stores/history-store';
import { HistoryPage } from './HistoryPage';

const runs = [
  {
    id: 'passed',
    testType: 'RFC 2544 throughput',
    module: 'benchmark',
    status: 'completed',
    success: true,
    completedAt: '2026-08-17T10:00:00Z',
    metrics: { throughput_mbps: 940 },
  },
  {
    id: 'no-verdict',
    testType: 'Y.1564 service',
    module: 'servicetest',
    status: 'completed',
    completedAt: '2026-08-17T11:00:00Z',
  },
];

describe('HistoryPage — list + detail', () => {
  beforeEach(() => {
    useHistoryStore.setState({ results: runs, lastRecorded: null });
  });

  afterEach(cleanup);

  it('prints an em dash rather than a verdict for a run that recorded none', () => {
    render(<HistoryPage />);

    expect(within(screen.getByTestId('history-row-no-verdict')).getByText('—')).toBeTruthy();
  });

  it('says the verdict is missing rather than showing the run as passed', async () => {
    render(<HistoryPage />);

    await userEvent.click(screen.getByTestId('history-row-no-verdict'));

    expect(screen.getByText('No verdict recorded')).toBeTruthy();
    expect(screen.queryByText('Passed')).toBeNull();
  });

  it('shows the selected run metrics in the detail pane', async () => {
    render(<HistoryPage />);

    await userEvent.click(screen.getByTestId('history-row-passed'));

    expect(screen.getByText('Selected run')).toBeTruthy();
    expect(screen.getByText('940')).toBeTruthy();
  });
});

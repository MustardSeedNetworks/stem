/**
 * HistoryPage.i18n.test.tsx — asserts the page discloses where its history
 * lives and how much of it is kept, in real locale copy.
 *
 * History is `localStorage`, so it is per-browser and capped. The page used
 * to say "on this machine", which is wrong in the direction that costs an
 * operator data: a second browser on the same machine has a different
 * history, and the oldest run falls off the end without a word. These tests
 * import the real i18n instance and the real cap constant, so copy that
 * drifts from `HISTORY_MAX_ITEMS`, or that is missing from `es`, fails here
 * rather than shipping.
 */
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { usePages } from '../pageRegistry';
import { HISTORY_MAX_ITEMS, useHistoryStore } from '../stores/history-store';
import { HistoryPage } from './HistoryPage';

const oneRun = [
  {
    id: 'passed',
    testType: 'RFC 2544 throughput',
    module: 'benchmark',
    status: 'completed',
    success: true,
    completedAt: '2026-08-17T10:00:00Z',
  },
];

/** The page heading comes from the route table, not from HistoryPage. */
function historyPageTitle(): string {
  let title = '';
  function Probe(): null {
    title = usePages().find((p) => p.path === '/history')?.title ?? '';
    return null;
  }
  render(<Probe />);
  return title;
}

beforeEach(() => {
  useHistoryStore.setState({ results: oneRun, lastRecorded: null });
});

afterEach(async () => {
  cleanup();
  await i18n.changeLanguage('en');
});

describe('HistoryPage — discloses that history is per-browser and capped', () => {
  it('titles the page after the browser the results are kept in (en)', async () => {
    await i18n.changeLanguage('en');

    expect(historyPageTitle()).toBe('Recent results on this browser');
  });

  it('titles the page after the browser the results are kept in (es)', async () => {
    await i18n.changeLanguage('es');

    expect(historyPageTitle()).toBe('Resultados recientes en este navegador');
  });

  it('shows the cap, taken from the store constant, in English', async () => {
    await i18n.changeLanguage('en');
    render(<HistoryPage />);

    const cap = screen.getByTestId('history-scope');
    expect(cap.textContent).toContain(String(HISTORY_MAX_ITEMS));
    expect(cap.textContent).toContain('browser');
  });

  it('shows the cap, taken from the store constant, in Spanish', async () => {
    await i18n.changeLanguage('es');
    render(<HistoryPage />);

    const cap = screen.getByTestId('history-scope');
    expect(cap.textContent).toContain(String(HISTORY_MAX_ITEMS));
    expect(cap.textContent).toContain('navegador');
  });

  it('does not claim the history belongs to the machine, in either locale', async () => {
    for (const [lng, wrong] of [
      ['en', 'machine'],
      ['es', 'máquina'],
    ] as const) {
      await i18n.changeLanguage(lng);
      const { unmount } = render(<HistoryPage />);

      expect(screen.getByTestId('history-scope').textContent).not.toContain(wrong);
      expect(document.body.textContent).not.toContain(wrong);

      unmount();
    }
  });
});

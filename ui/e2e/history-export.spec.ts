import { readFile } from 'node:fs/promises';
import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * History export (#1262, plan row UI-STEM-6).
 *
 * History is per-browser localStorage, so export is the only way a result
 * leaves the browser. The archive is seeded under the store's own key rather
 * than produced by a run: a terminal run needs a dataplane or a licence the
 * E2E daemon deliberately lacks (see `history-run.spec.ts`), and what this
 * asserts is the file an operator gets from what is stored.
 */

const stored = [
  {
    id: '2026-08-17T10:00:00Z-RFC 2544',
    testType: 'RFC 2544',
    module: 'benchmark',
    status: 'completed',
    startedAt: '2026-08-17T09:58:00Z',
    completedAt: '2026-08-17T10:00:00Z',
    duration: 120000,
    success: true,
    metrics: { throughput_mbps: 940 },
  },
  {
    id: '2026-08-17T11:00:00Z-Y.1564',
    testType: 'Y.1564',
    module: 'servicetest',
    status: 'error',
    completedAt: '2026-08-17T11:00:00Z',
    success: false,
    error: 'peer unreachable',
  },
];

test.describe('History export', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.addInitScript((runs) => {
      window.localStorage.setItem('stem-result-history', JSON.stringify(runs));
    }, stored);
    await page.goto('/history');
    await expect(page.getByTestId('history-row-2026-08-17T11:00:00Z-Y.1564')).toBeVisible();
  });

  test('exports one CSV row per stored run with its summary fields', async ({ page }) => {
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.getByTestId('history-export-csv').click(),
    ]);

    expect(download.suggestedFilename()).toMatch(/^stem-history-\d{8}-\d{6}\.csv$/);
    const csv = await readFile(await download.path(), 'utf8');
    expect(csv.trimEnd().split('\r\n')).toEqual([
      'id,testType,module,status,verdict,startedAt,completedAt,durationMs,error,throughput_mbps',
      '2026-08-17T10:00:00Z-RFC 2544,RFC 2544,benchmark,completed,PASS,2026-08-17T09:58:00Z,2026-08-17T10:00:00Z,120000,,940',
      '2026-08-17T11:00:00Z-Y.1564,Y.1564,servicetest,error,FAIL,,2026-08-17T11:00:00Z,,peer unreachable,',
    ]);
  });

  test('exports the stored runs as JSON', async ({ page }) => {
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.getByTestId('history-export-json').click(),
    ]);

    expect(download.suggestedFilename()).toMatch(/^stem-history-\d{8}-\d{6}\.json$/);
    expect(JSON.parse(await readFile(await download.path(), 'utf8'))).toEqual(stored);
  });
});

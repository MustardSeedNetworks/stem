/**
 * History export
 *
 * History lives in this browser's localStorage, so without an export a result
 * cannot be handed to anyone. CSV is for a spreadsheet: one row per run, the
 * run's summary fields, then one column per metric any run recorded. JSON is
 * the stored records as they are.
 */

import type { HistoricalResult } from './history-store';

const SUMMARY_COLUMNS = [
  'id',
  'testType',
  'module',
  'status',
  'verdict',
  'startedAt',
  'completedAt',
  'durationMs',
  'error',
] as const;

/**
 * A run with no recorded verdict exports an empty cell, not PASS — the same
 * rule the History page applies on screen.
 */
function verdictOf(result: HistoricalResult): string {
  if (result.success === undefined) {
    return '';
  }
  return result.success ? 'PASS' : 'FAIL';
}

/**
 * A spreadsheet runs a cell that starts with `=`, `+`, `-` or `@` as a
 * formula, and the error text and metric strings are daemon data. A leading
 * quote keeps them text (OWASP CSV injection). A string that is just a
 * number, such as "-3.5", is left alone so it still sorts as one.
 */
function neutralize(value: string): string {
  if (!/^[=+\-@\t\r]/.test(value)) {
    return value;
  }
  return value.trim() !== '' && Number.isFinite(Number(value)) ? value : `'${value}`;
}

/** RFC 4180 quoting: a cell with a comma, quote or line break is quoted. */
function cell(value: string | number | undefined): string {
  if (value === undefined) {
    return '';
  }
  const text = typeof value === 'number' ? String(value) : neutralize(value);
  return /[",\r\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

export function historyToCsv(results: readonly HistoricalResult[]): string {
  const metricColumns = [...new Set(results.flatMap((r) => Object.keys(r.metrics ?? {})))].sort();
  const rows = results.map((r) =>
    [
      r.id,
      r.testType,
      r.module,
      r.status,
      verdictOf(r),
      r.startedAt,
      r.completedAt,
      r.duration,
      r.error,
      ...metricColumns.map((key) => r.metrics?.[key]),
    ]
      .map(cell)
      .join(','),
  );
  return [[...SUMMARY_COLUMNS, ...metricColumns].map(cell).join(','), ...rows]
    .map((line) => `${line}\r\n`)
    .join('');
}

export function historyToJson(results: readonly HistoricalResult[]): string {
  return `${JSON.stringify(results, null, 2)}\n`;
}

/** `stem-history-20260925-081700.csv`, in the operator's local time. */
export function exportFileName(now: Date, extension: 'csv' | 'json'): string {
  const pad = (n: number): string => String(n).padStart(2, '0');
  const date = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}`;
  const time = `${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
  return `stem-history-${date}-${time}.${extension}`;
}

export function downloadFile(fileName: string, mimeType: string, content: string): void {
  const url = URL.createObjectURL(new Blob([content], { type: mimeType }));
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName;
  link.click();
  URL.revokeObjectURL(url);
}

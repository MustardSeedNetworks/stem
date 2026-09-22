/**
 * History Store
 *
 * Zustand store for completed test results. Persisted to localStorage under
 * the key the previous drawer used, so an operator's existing history
 * survives the move onto the History page.
 *
 * Recording lives here rather than in a component so that it does not depend
 * on any particular view being mounted — a result is recorded because a test
 * finished, not because a drawer happened to be open.
 */

import { useEffect } from 'react';
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { Stats, TestResult } from '../types/api';

/** Test result record stored in history. */
export interface HistoricalResult {
  id: string;
  testType: string;
  module: string;
  status: string;
  startedAt?: string;
  completedAt?: string;
  duration?: number;
  success?: boolean;
  error?: string;
  metrics?: Record<string, number | string>;
  data?: Record<string, unknown>;
}

const STORAGE_KEY = 'stem-result-history';

/**
 * How many results this browser keeps. Exported because the History page
 * discloses the cap to the operator: the number in the copy is this
 * constant, not a literal that can drift away from it.
 */
export const HISTORY_MAX_ITEMS = 50;

function load(): HistoricalResult[] {
  if (typeof window === 'undefined') {
    return [];
  }
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      return [];
    }
    const parsed = JSON.parse(stored) as HistoricalResult[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    // A corrupt entry must not take the app down; an empty history is recoverable.
    return [];
  }
}

function save(results: HistoricalResult[]): void {
  if (typeof window === 'undefined') {
    return;
  }
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(results.slice(0, HISTORY_MAX_ITEMS)));
  } catch {
    // Storage full or blocked — the in-memory history still works this session.
  }
}

interface HistoryState {
  results: HistoricalResult[];
  /** completedAt of the last result recorded, so one run is not stored twice. */
  lastRecorded: string | null;
}

interface HistoryActions {
  record: (result: Omit<HistoricalResult, 'id'>) => void;
  remove: (id: string) => void;
  clear: () => void;
}

export type HistoryStore = HistoryState & HistoryActions;

export const useHistoryStore = create<HistoryStore>()(
  devtools(
    (set, get) => ({
      results: load(),
      lastRecorded: null,
      record: (result) => {
        if (!result.completedAt || result.completedAt === get().lastRecorded) {
          return;
        }
        const id = `${result.completedAt}-${result.testType}`;
        // A reload re-reads the daemon's still-terminal result, and
        // `lastRecorded` does not survive the reload — the persisted runs do.
        // Without this the same run comes back as a second row sharing its id.
        if (get().results.some((existing) => existing.id === id)) {
          return;
        }
        const results = [{ id, ...result }, ...get().results].slice(0, HISTORY_MAX_ITEMS);
        save(results);
        set({ results, lastRecorded: result.completedAt }, false, 'record');
      },
      remove: (id) => {
        const results = get().results.filter((r) => r.id !== id);
        save(results);
        set({ results }, false, 'remove');
      },
      clear: () => {
        save([]);
        set({ results: [] }, false, 'clear');
      },
    }),
    { name: 'history-store' },
  ),
);

/**
 * The states in which a run has ended, mirroring the daemon's own
 * `runHasEnded` (internal/api/run_timing.go). A run is recorded because it
 * reached one of these, not because a timestamp happened to be present: the
 * recorder this replaces gated on `completedAt`, which the daemon never sent,
 * so the History page could never show a row (#1333).
 */
const TERMINAL_STATUSES: ReadonlySet<string> = new Set([
  'completed',
  'error',
  'stopped',
  'cancelled',
]);

/**
 * The run's measurements, as the detail pane's metrics panel reads them.
 *
 * `data` is whatever the module computed, so only its scalar entries are
 * measurements a two-column grid can show; a nested object is the module's
 * own structure and stays in `data`.
 */
function metricsOf(data: unknown): Record<string, number | string> | undefined {
  if (typeof data !== 'object' || data === null || Array.isArray(data)) {
    return undefined;
  }
  const metrics: Record<string, number | string> = {};
  for (const [key, value] of Object.entries(data)) {
    if (typeof value === 'number' || typeof value === 'string') {
      metrics[key] = value;
    }
  }
  return Object.keys(metrics).length > 0 ? metrics : undefined;
}

/**
 * Records a run as it ends. Called once, high in the tree, so recording does
 * not depend on the History page being open.
 *
 * The timing is the daemon's: stem is a measurement instrument, and a browser
 * clock stamping the end of a run would report the moment the answer arrived
 * rather than the moment the run finished.
 */
export function useRecordTestResult(result: TestResult | null, status: Stats['testStatus']): void {
  const record = useHistoryStore((s) => s.record);
  useEffect(() => {
    if (!result || !TERMINAL_STATUSES.has(status) || !result.completedAt) {
      return;
    }
    record({
      testType: result.testType ?? status,
      module: result.module ?? '',
      status: result.status,
      startedAt: result.startedAt,
      completedAt: result.completedAt,
      duration: result.duration,
      success: result.success,
      error: result.error,
      metrics: metricsOf(result.data),
    });
  }, [result, status, record]);
}

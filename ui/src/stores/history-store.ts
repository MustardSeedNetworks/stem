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
const MAX_ITEMS = 50;

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
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(results.slice(0, MAX_ITEMS)));
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
        const results = [
          { id: `${result.completedAt}-${result.testType}`, ...result },
          ...get().results,
        ].slice(0, MAX_ITEMS);
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
 * Records a run as it completes. Called once, high in the tree, so recording
 * does not depend on the History page being open.
 */
export function useRecordTestResult(result: Omit<HistoricalResult, 'id'> | null): void {
  const record = useHistoryStore((s) => s.record);
  useEffect(() => {
    if (result?.completedAt) {
      record(result);
    }
  }, [result, record]);
}

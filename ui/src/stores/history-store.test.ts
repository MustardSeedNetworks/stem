/**
 * History-store tests.
 *
 * The store is the only thing standing between a finished run and the
 * operator's record of it, and it persists across sessions — so the cases that
 * matter are the ones about not losing or duplicating results: the
 * same-completedAt guard, the 50-item cap, and surviving a corrupt or
 * unwritable localStorage rather than taking the app down.
 */

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { type HistoricalResult, useHistoryStore } from './history-store';

const STORAGE_KEY = 'stem-result-history';
const MAX_ITEMS = 50;

/** A completed run, with a distinct completedAt so it is not deduplicated. */
function result(overrides: Partial<HistoricalResult> = {}): Omit<HistoricalResult, 'id'> {
  return {
    testType: 'rfc2544',
    module: 'benchmark',
    status: 'completed',
    completedAt: '2026-08-30T10:00:00.000Z',
    success: true,
    ...overrides,
  };
}

function stored(): HistoricalResult[] {
  const raw = window.localStorage.getItem(STORAGE_KEY);
  return raw ? (JSON.parse(raw) as HistoricalResult[]) : [];
}

beforeEach(() => {
  window.localStorage.clear();
  vi.restoreAllMocks();
  useHistoryStore.setState({ results: [], lastRecorded: null });
});

describe('record', () => {
  it('stores the run and remembers when it completed', () => {
    useHistoryStore.getState().record(result());

    const { results, lastRecorded } = useHistoryStore.getState();
    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({ testType: 'rfc2544', module: 'benchmark' });
    expect(lastRecorded).toBe('2026-08-30T10:00:00.000Z');
  });

  it('derives the id from completedAt and testType', () => {
    useHistoryStore.getState().record(result());

    expect(useHistoryStore.getState().results[0].id).toBe('2026-08-30T10:00:00.000Z-rfc2544');
  });

  it('persists to localStorage, not only to memory', () => {
    useHistoryStore.getState().record(result());

    expect(stored()).toHaveLength(1);
  });

  it('ignores a run with no completedAt', () => {
    // An in-progress run has no completion time; recording it would put a
    // half-finished entry in the operator's history.
    useHistoryStore.getState().record(result({ completedAt: undefined }));

    expect(useHistoryStore.getState().results).toHaveLength(0);
  });

  it('ignores a repeat of the run it just recorded', () => {
    // useRecordTestResult fires on every render while a result object is in
    // scope, so the same finished run arrives many times.
    const finished = result();
    useHistoryStore.getState().record(finished);
    useHistoryStore.getState().record(finished);
    useHistoryStore.getState().record(finished);

    expect(useHistoryStore.getState().results).toHaveLength(1);
  });

  it('records a later run with a different completion time', () => {
    useHistoryStore.getState().record(result());
    useHistoryStore.getState().record(result({ completedAt: '2026-08-30T11:00:00.000Z' }));

    expect(useHistoryStore.getState().results).toHaveLength(2);
  });

  it('puts the newest run first', () => {
    useHistoryStore.getState().record(result({ testType: 'older' }));
    useHistoryStore
      .getState()
      .record(result({ testType: 'newer', completedAt: '2026-08-30T11:00:00.000Z' }));

    expect(useHistoryStore.getState().results.map((r) => r.testType)).toEqual(['newer', 'older']);
  });

  it('keeps only the newest 50 runs', () => {
    for (let i = 0; i < MAX_ITEMS + 10; i++) {
      const second = String(i).padStart(2, '0');
      const run = result({ testType: `run-${i}`, completedAt: `2026-08-30T10:00:${second}.000Z` });
      useHistoryStore.getState().record(run);
    }

    const { results } = useHistoryStore.getState();
    expect(results).toHaveLength(MAX_ITEMS);
    expect(results[0].testType).toBe(`run-${MAX_ITEMS + 9}`);
    expect(stored()).toHaveLength(MAX_ITEMS);
  });

  it('keeps the run in memory when localStorage refuses the write', () => {
    // A full or blocked quota must not cost the operator the result they are
    // currently looking at.
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('QuotaExceededError');
    });

    expect(() => useHistoryStore.getState().record(result())).not.toThrow();
    expect(useHistoryStore.getState().results).toHaveLength(1);
  });
});

describe('remove', () => {
  it('removes only the run with the given id, and persists that', () => {
    useHistoryStore.getState().record(result({ testType: 'keep' }));
    useHistoryStore
      .getState()
      .record(result({ testType: 'drop', completedAt: '2026-08-30T11:00:00.000Z' }));

    const dropId = useHistoryStore.getState().results.find((r) => r.testType === 'drop')?.id;
    expect(dropId).toBeDefined();

    useHistoryStore.getState().remove(dropId as string);

    expect(useHistoryStore.getState().results.map((r) => r.testType)).toEqual(['keep']);
    expect(stored().map((r) => r.testType)).toEqual(['keep']);
  });

  it('is a no-op for an id that is not present', () => {
    useHistoryStore.getState().record(result());

    useHistoryStore.getState().remove('no-such-id');

    expect(useHistoryStore.getState().results).toHaveLength(1);
  });
});

describe('clear', () => {
  it('empties both the store and localStorage', () => {
    useHistoryStore.getState().record(result());

    useHistoryStore.getState().clear();

    expect(useHistoryStore.getState().results).toEqual([]);
    expect(stored()).toEqual([]);
  });
});

describe('hydration from localStorage', () => {
  /**
   * load() runs once at module evaluation, so each case has to seed storage
   * and then re-import the module. The corrupt-entry branches are the point:
   * history is operator data that has survived across releases, so a value
   * written by an older version must degrade to an empty list rather than
   * throwing during app start-up.
   */
  async function freshStore(raw: string | null) {
    window.localStorage.clear();
    if (raw !== null) {
      window.localStorage.setItem(STORAGE_KEY, raw);
    }
    vi.resetModules();
    const mod = await import('./history-store');

    return mod.useHistoryStore.getState().results;
  }

  it('restores previously stored results', async () => {
    const saved = [{ id: 'x', testType: 'rfc2544', module: 'benchmark', status: 'completed' }];

    await expect(freshStore(JSON.stringify(saved))).resolves.toEqual(saved);
  });

  it('starts empty when nothing is stored', async () => {
    await expect(freshStore(null)).resolves.toEqual([]);
  });

  it('starts empty when the stored value is not JSON', async () => {
    await expect(freshStore('{ this is not json')).resolves.toEqual([]);
  });

  it('starts empty when the stored value is JSON but not an array', async () => {
    // An older release could have written an object here; JSON.parse succeeds,
    // so only the Array.isArray check catches it.
    await expect(freshStore('{"results":[]}')).resolves.toEqual([]);
  });
});

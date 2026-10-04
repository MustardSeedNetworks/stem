/**
 * Test store tests.
 *
 * Every setter promises React's `useState` dispatch contract — a value or an
 * updater over the previous value — and must write only its own field. The
 * config forms and selected-test toggles rely on both halves (#824).
 */

import { beforeEach, describe, expect, it } from 'vitest';
import { type TestStore, useTestStore } from './test-store';

type Setter = {
  [K in keyof TestStore]: TestStore[K] extends (update: never) => void ? K : never;
}[keyof TestStore];

const cases: { setter: Setter; field: keyof TestStore; next: unknown }[] = [
  { setter: 'setSelectedTests', field: 'selectedTests', next: ['y1564_config'] },
  { setter: 'setReflectorProfile', field: 'reflectorProfile', next: 'netally' },
  { setter: 'setIsStartingTest', field: 'isStartingTest', next: true },
  { setter: 'setStopOutcome', field: 'stopOutcome', next: { kind: 'stopping' } },
  { setter: 'setTestStartError', field: 'testStartError', next: 'refused' },
  { setter: 'setRFC2544Config', field: 'rfc2544Config', next: { marker: 2544 } },
  { setter: 'setRFC2889Config', field: 'rfc2889Config', next: { marker: 2889 } },
  { setter: 'setRFC6349Config', field: 'rfc6349Config', next: { marker: 6349 } },
  { setter: 'setY1564Config', field: 'y1564Config', next: { marker: 1564 } },
  { setter: 'setY1731Config', field: 'y1731Config', next: { marker: 1731 } },
  { setter: 'setTSNConfig', field: 'tsnConfig', next: { marker: 'tsn' } },
  { setter: 'setTrafficGenConfig', field: 'trafficGenConfig', next: { marker: 'gen' } },
];

function dataFields(state: TestStore): Partial<TestStore> {
  return Object.fromEntries(
    Object.entries(state).filter(([, value]) => typeof value !== 'function'),
  ) as Partial<TestStore>;
}

function call(setter: Setter, update: unknown): void {
  (useTestStore.getState()[setter] as (u: unknown) => void)(update);
}

describe('useTestStore setters', () => {
  beforeEach(() => {
    useTestStore.setState(useTestStore.getInitialState());
  });

  it('covers every setter the store exposes', () => {
    const setters = Object.keys(useTestStore.getState()).filter((k) => k.startsWith('set'));
    expect(setters.sort()).toEqual(cases.map((c) => c.setter).sort());
  });

  it.each(cases)('$setter writes a value to $field and nothing else', ({ setter, field, next }) => {
    const before = dataFields(useTestStore.getState());

    call(setter, next);

    expect(dataFields(useTestStore.getState())).toEqual({ ...before, [field]: next });
  });

  it.each(cases)('$setter hands an updater the previous $field', ({ setter, field, next }) => {
    const previous = useTestStore.getState()[field];
    let seen: unknown;

    call(setter, (prev: unknown) => {
      seen = prev;
      return next;
    });

    expect(seen).toBe(previous);
    expect(useTestStore.getState()[field]).toEqual(next);
  });
});

/**
 * Test Store
 *
 * Zustand store for the test-execution configuration and run state that
 * previously lived as local `useState` in App.tsx (stage 4 of the App.tsx
 * decomposition): the selected tests, reflector profile, per-protocol config
 * objects, and the start/stop/error flags.
 *
 * Every setter mirrors React's `useState` dispatch — it accepts either a value
 * or an updater `(prev) => next` — so call sites (the config forms, the
 * selected-tests toggles) are exact drop-ins and behaviour is preserved.
 *
 * Not persisted: matches the previous useState behaviour (test config resets on
 * reload). The module *catalog* persistence lives separately in
 * ModuleSettingsContext and is intentionally untouched here.
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { defaultRFC2544Config, type RFC2544Config } from '../components/RFC2544ConfigForm';
import { defaultRFC2889Config, type RFC2889Config } from '../components/RFC2889ConfigForm';
import { defaultRFC6349Config, type RFC6349Config } from '../components/RFC6349ConfigForm';
import type { ReflectorProfile } from '../components/settings/types';
import { defaultTrafficGenConfig, type TrafficGenConfig } from '../components/TrafficGenConfigForm';
import { defaultTSNConfig, type TSNConfig } from '../components/TSNConfigForm';
import { defaultY1564Config, type Y1564Config } from '../components/Y1564ConfigForm';
import { defaultY1731Config, type Y1731Config } from '../components/Y1731ConfigForm';

/** React-useState-style update: a value or an updater function. */
type SetState<T> = T | ((prev: T) => T);

function resolve<T>(update: SetState<T>, prev: T): T {
  return typeof update === 'function' ? (update as (p: T) => T)(prev) : update;
}

interface TestState {
  selectedTests: string[];
  reflectorProfile: ReflectorProfile;
  isStartingTest: boolean;
  isStoppingTest: boolean;
  testStartError: string | null;
  rfc2544Config: RFC2544Config;
  rfc2889Config: RFC2889Config;
  rfc6349Config: RFC6349Config;
  y1564Config: Y1564Config;
  y1731Config: Y1731Config;
  tsnConfig: TSNConfig;
  trafficGenConfig: TrafficGenConfig;
}

interface TestActions {
  setSelectedTests: (update: SetState<string[]>) => void;
  setReflectorProfile: (update: SetState<ReflectorProfile>) => void;
  setIsStartingTest: (update: SetState<boolean>) => void;
  setIsStoppingTest: (update: SetState<boolean>) => void;
  setTestStartError: (update: SetState<string | null>) => void;
  setRFC2544Config: (update: SetState<RFC2544Config>) => void;
  setRFC2889Config: (update: SetState<RFC2889Config>) => void;
  setRFC6349Config: (update: SetState<RFC6349Config>) => void;
  setY1564Config: (update: SetState<Y1564Config>) => void;
  setY1731Config: (update: SetState<Y1731Config>) => void;
  setTSNConfig: (update: SetState<TSNConfig>) => void;
  setTrafficGenConfig: (update: SetState<TrafficGenConfig>) => void;
}

export type TestStore = TestState & TestActions;

export const useTestStore = create<TestStore>()(
  devtools(
    (set) => ({
      selectedTests: [
        'rfc2544_throughput',
        'rfc2544_latency',
        'rfc2544_frame_loss',
        'rfc2544_back_to_back',
      ],
      reflectorProfile: 'all',
      isStartingTest: false,
      isStoppingTest: false,
      testStartError: null,
      rfc2544Config: defaultRFC2544Config,
      rfc2889Config: defaultRFC2889Config,
      rfc6349Config: defaultRFC6349Config,
      y1564Config: defaultY1564Config,
      y1731Config: defaultY1731Config,
      tsnConfig: defaultTSNConfig,
      trafficGenConfig: defaultTrafficGenConfig,
      setSelectedTests: (u) =>
        set((s) => ({ selectedTests: resolve(u, s.selectedTests) }), false, 'setSelectedTests'),
      setReflectorProfile: (u) =>
        set(
          (s) => ({ reflectorProfile: resolve(u, s.reflectorProfile) }),
          false,
          'setReflectorProfile',
        ),
      setIsStartingTest: (u) =>
        set((s) => ({ isStartingTest: resolve(u, s.isStartingTest) }), false, 'setIsStartingTest'),
      setIsStoppingTest: (u) =>
        set((s) => ({ isStoppingTest: resolve(u, s.isStoppingTest) }), false, 'setIsStoppingTest'),
      setTestStartError: (u) =>
        set((s) => ({ testStartError: resolve(u, s.testStartError) }), false, 'setTestStartError'),
      setRFC2544Config: (u) =>
        set((s) => ({ rfc2544Config: resolve(u, s.rfc2544Config) }), false, 'setRFC2544Config'),
      setRFC2889Config: (u) =>
        set((s) => ({ rfc2889Config: resolve(u, s.rfc2889Config) }), false, 'setRFC2889Config'),
      setRFC6349Config: (u) =>
        set((s) => ({ rfc6349Config: resolve(u, s.rfc6349Config) }), false, 'setRFC6349Config'),
      setY1564Config: (u) =>
        set((s) => ({ y1564Config: resolve(u, s.y1564Config) }), false, 'setY1564Config'),
      setY1731Config: (u) =>
        set((s) => ({ y1731Config: resolve(u, s.y1731Config) }), false, 'setY1731Config'),
      setTSNConfig: (u) =>
        set((s) => ({ tsnConfig: resolve(u, s.tsnConfig) }), false, 'setTSNConfig'),
      setTrafficGenConfig: (u) =>
        set(
          (s) => ({ trafficGenConfig: resolve(u, s.trafficGenConfig) }),
          false,
          'setTrafficGenConfig',
        ),
    }),
    { name: 'test-store' },
  ),
);

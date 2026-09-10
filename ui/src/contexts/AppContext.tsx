/**
 * AppContext — shared state surface for the routed Stem pages.
 *
 * The Stem App owns all the test/auth/state hooks; pages read from
 * this context to render their slice (Reflector view, Benchmark form,
 * etc.). This keeps state ownership in one place during the Phase A
 * router refactor — pages don't fetch anything themselves.
 */
import { createContext, type Dispatch, type SetStateAction, useContext } from 'react';
import type { RFC2544Config } from '../components/RFC2544ConfigForm';
import type { RFC2889Config } from '../components/RFC2889ConfigForm';
import type { RFC6349Config } from '../components/RFC6349ConfigForm';
import type { ReflectorProfile } from '../components/settings/types';
import type { TrafficGenConfig } from '../components/TrafficGenConfigForm';
import type { TSNConfig } from '../components/TSNConfigForm';
import type { Y1564Config } from '../components/Y1564ConfigForm';
import type { Y1731Config } from '../components/Y1731ConfigForm';
import type { StopOutcome } from '../stores/test-store';
import type { InterfaceInfo, Stats, TestResult } from '../types/api';

export interface AppContextValue {
  // Test configs
  rfc2544Config: RFC2544Config;
  setRFC2544Config: Dispatch<SetStateAction<RFC2544Config>>;
  rfc2889Config: RFC2889Config;
  setRFC2889Config: Dispatch<SetStateAction<RFC2889Config>>;
  rfc6349Config: RFC6349Config;
  setRFC6349Config: Dispatch<SetStateAction<RFC6349Config>>;
  y1564Config: Y1564Config;
  setY1564Config: Dispatch<SetStateAction<Y1564Config>>;
  y1731Config: Y1731Config;
  setY1731Config: Dispatch<SetStateAction<Y1731Config>>;
  tsnConfig: TSNConfig;
  setTSNConfig: Dispatch<SetStateAction<TSNConfig>>;
  trafficGenConfig: TrafficGenConfig;
  setTrafficGenConfig: Dispatch<SetStateAction<TrafficGenConfig>>;

  // Test selection state
  selectedTests: string[];

  // Test result for History page
  testResult: TestResult | null;

  // Reflector-page surface (filled when the Reflector page mounts).
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  setSelectedInterface: (name: string) => void;
  stats: Stats;
  reflectorProfile: ReflectorProfile;
  setReflectorProfile: Dispatch<SetStateAction<ReflectorProfile>>;
  onStartReflector: () => void;
  onStopReflector: () => void;
  isStartingReflector: boolean;
  reflectorStopOutcome: StopOutcome;
  reflectorStartError: string | null;
}

export const AppContext = createContext<AppContextValue | null>(null);

export function useAppContext(): AppContextValue {
  const ctx = useContext(AppContext);
  if (!ctx) {
    throw new Error('useAppContext must be used inside <AppContext.Provider>');
  }
  return ctx;
}

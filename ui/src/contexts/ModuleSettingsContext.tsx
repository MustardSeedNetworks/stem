/**
 * @fileoverview The Stem - Module Settings Context
 * @description Manages per-module configuration state including enabled status,
 *              autostart options, and individual test toggles.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import {
  createContext,
  type ReactElement,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import type {
  ModuleConfig,
  ModuleStatus,
  ModuleTest,
  ModuleTestResults,
} from '../components/ModuleCard';

const STORAGE_KEY = 'stem-module-settings';

// Default module configurations based on Stem architecture
const defaultModules: ModuleConfig[] = [
  {
    name: 'reflector',
    displayName: 'Reflector',
    description: 'Packet reflection and loopback testing',
    color: '#0891b2', // Cyan
    standard: 'Loopback',
    enabled: true,
    autoStart: false,
    tests: [
      {
        id: 'reflect',
        name: 'Reflect',
        description: 'Basic packet reflection',
        enabled: true,
      },
      {
        id: 'echo',
        name: 'Echo',
        description: 'ICMP echo response',
        enabled: true,
      },
    ],
  },
  {
    name: 'benchmark',
    displayName: 'Benchmark',
    description: 'RFC 2544 network benchmarking tests',
    color: '#dc2626', // Red
    standard: 'RFC 2544',
    enabled: true,
    autoStart: false,
    tests: [
      {
        id: 'throughput',
        name: 'Throughput',
        description: 'Maximum frame rate without loss',
        enabled: true,
      },
      {
        id: 'latency',
        name: 'Latency',
        description: 'Round-trip delay measurement',
        enabled: true,
      },
      {
        id: 'frame_loss',
        name: 'Frame Loss',
        description: 'Frame loss rate determination',
        enabled: true,
      },
      {
        id: 'back_to_back',
        name: 'Back-to-Back',
        description: 'Burst handling capability',
        enabled: false,
      },
      {
        id: 'system_recovery',
        name: 'System Recovery',
        description: 'Recovery time after overload',
        enabled: false,
      },
      {
        id: 'reset',
        name: 'Reset',
        description: 'Reset recovery time',
        enabled: false,
      },
    ],
  },
  {
    name: 'servicetest',
    displayName: 'Service Test',
    description: 'ITU-T Y.1564 service activation testing',
    color: '#ea580c', // Orange
    standard: 'Y.1564 / MEF',
    enabled: true,
    autoStart: false,
    tests: [
      {
        id: 'y1564_config',
        name: 'Y.1564 Config',
        description: 'Service configuration test',
        enabled: true,
      },
      {
        id: 'y1564_perf',
        name: 'Y.1564 Performance',
        description: 'Service performance test',
        enabled: true,
      },
      {
        id: 'y1564',
        name: 'Y.1564 Full',
        description: 'Complete Y.1564 test suite',
        enabled: false,
      },
      {
        id: 'mef_config',
        name: 'MEF Config',
        description: 'MEF configuration test',
        enabled: false,
      },
      {
        id: 'mef_perf',
        name: 'MEF Performance',
        description: 'MEF performance test',
        enabled: false,
      },
    ],
  },
  {
    name: 'trafficgen',
    displayName: 'Traffic Gen',
    description: 'Custom traffic generation',
    color: '#ca8a04', // Yellow
    standard: 'Custom',
    enabled: false,
    autoStart: false,
    tests: [
      {
        id: 'custom_stream',
        name: 'Custom Stream',
        description: 'User-defined traffic pattern',
        enabled: true,
      },
    ],
  },
  {
    name: 'measure',
    displayName: 'Measure',
    description: 'ITU-T Y.1731 OAM measurements',
    color: '#2563eb', // Blue
    standard: 'Y.1731',
    enabled: false,
    autoStart: false,
    tests: [
      {
        id: 'y1731_delay',
        name: 'Delay',
        description: 'Frame delay measurement',
        enabled: true,
      },
      {
        id: 'y1731_loss',
        name: 'Loss',
        description: 'Frame loss measurement',
        enabled: true,
      },
      {
        id: 'y1731_slm',
        name: 'SLM',
        description: 'Synthetic loss measurement',
        enabled: false,
      },
      {
        id: 'y1731_loopback',
        name: 'Loopback',
        description: 'ETH-LB test',
        enabled: false,
      },
    ],
  },
  {
    name: 'certify',
    displayName: 'Certify',
    description: 'Network certification tests',
    color: '#16a34a', // Green
    standard: 'RFC 2889/6349',
    enabled: false,
    autoStart: false,
    tests: [
      {
        id: 'rfc2889_fwd',
        name: 'RFC 2889 Forwarding',
        description: 'LAN switch forwarding',
        enabled: true,
      },
      {
        id: 'rfc2889_congestion',
        name: 'RFC 2889 Congestion',
        description: 'Congestion control',
        enabled: false,
      },
      {
        id: 'rfc6349_tcp',
        name: 'RFC 6349 TCP',
        description: 'TCP throughput testing',
        enabled: true,
      },
    ],
  },
];

interface ModuleSettingsContextValue {
  modules: ModuleConfig[];
  moduleStatuses: Record<string, ModuleStatus>;
  moduleResults: Record<string, ModuleTestResults | null>;
  toggleModule: (moduleName: string, enabled: boolean) => void;
  toggleAutoStart: (moduleName: string, enabled: boolean) => void;
  toggleTest: (moduleName: string, testId: string, enabled: boolean) => void;
  setModuleStatus: (moduleName: string, status: ModuleStatus) => void;
  setModuleResults: (moduleName: string, results: ModuleTestResults | null) => void;
  updateModuleResults: (
    moduleName: string,
    updater: (prev: ModuleTestResults | null) => ModuleTestResults | null,
  ) => void;
  clearModuleResults: (moduleName: string) => void;
  getEnabledTests: (moduleName: string) => ModuleTest[];
  getAllEnabledTests: () => Array<{ module: string; test: ModuleTest }>;
  resetToDefaults: () => void;
}

const ModuleSettingsContext: React.Context<ModuleSettingsContextValue | null> =
  createContext<ModuleSettingsContextValue | null>(null);

function loadFromStorage(): ModuleConfig[] | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored) as ModuleConfig[];
    }
  } catch {
    // Ignore parse errors
  }
  return null;
}

function saveToStorage(modules: ModuleConfig[]): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(modules));
  } catch {
    // Ignore storage errors
  }
}

interface ModuleSettingsProviderProps {
  children: ReactNode;
}

export function ModuleSettingsProvider({ children }: ModuleSettingsProviderProps): ReactElement {
  const [modules, setModules] = useState<ModuleConfig[]>(() => {
    const stored = loadFromStorage();
    return stored ?? defaultModules;
  });

  const [moduleStatuses, setModuleStatuses] = useState<Record<string, ModuleStatus>>(() => {
    const statuses: Record<string, ModuleStatus> = {};
    for (const mod of defaultModules) {
      statuses[mod.name] = { status: 'idle', currentTest: null };
    }
    return statuses;
  });

  const [moduleResults, setModuleResultsState] = useState<Record<string, ModuleTestResults | null>>(
    () => {
      const results: Record<string, ModuleTestResults | null> = {};
      for (const mod of defaultModules) {
        results[mod.name] = null;
      }
      return results;
    },
  );

  // Save to storage when modules change
  useEffect(() => {
    saveToStorage(modules);
  }, [modules]);

  const toggleModule = useCallback((moduleName: string, enabled: boolean) => {
    setModules((prev) => prev.map((mod) => (mod.name === moduleName ? { ...mod, enabled } : mod)));
  }, []);

  const toggleAutoStart = useCallback((moduleName: string, enabled: boolean) => {
    setModules((prev) =>
      prev.map((mod) => (mod.name === moduleName ? { ...mod, autoStart: enabled } : mod)),
    );
  }, []);

  const toggleTest = useCallback((moduleName: string, testId: string, enabled: boolean) => {
    setModules((prev) =>
      prev.map((mod) =>
        mod.name === moduleName
          ? {
              ...mod,
              tests: mod.tests.map((t) => (t.id === testId ? { ...t, enabled } : t)),
            }
          : mod,
      ),
    );
  }, []);

  const setModuleStatus = useCallback((moduleName: string, status: ModuleStatus) => {
    setModuleStatuses((prev) => ({ ...prev, [moduleName]: status }));
  }, []);

  const setModuleResults = useCallback((moduleName: string, results: ModuleTestResults | null) => {
    setModuleResultsState((prev) => ({ ...prev, [moduleName]: results }));
  }, []);

  const updateModuleResults = useCallback(
    (moduleName: string, updater: (prev: ModuleTestResults | null) => ModuleTestResults | null) => {
      setModuleResultsState((prev) => ({
        ...prev,
        [moduleName]: updater(prev[moduleName] ?? null),
      }));
    },
    [],
  );

  const clearModuleResults = useCallback((moduleName: string) => {
    setModuleResultsState((prev) => ({ ...prev, [moduleName]: null }));
  }, []);

  const getEnabledTests = useCallback(
    (moduleName: string): ModuleTest[] => {
      const mod = modules.find((m) => m.name === moduleName);
      return mod?.tests.filter((t) => t.enabled) ?? [];
    },
    [modules],
  );

  const getAllEnabledTests = useCallback((): Array<{
    module: string;
    test: ModuleTest;
  }> => {
    const result: Array<{ module: string; test: ModuleTest }> = [];
    for (const mod of modules) {
      if (mod.enabled) {
        for (const test of mod.tests) {
          if (test.enabled) {
            result.push({ module: mod.name, test });
          }
        }
      }
    }
    return result;
  }, [modules]);

  const resetToDefaults = useCallback(() => {
    setModules(defaultModules);
    localStorage.removeItem(STORAGE_KEY);
  }, []);

  const value = useMemo(
    () => ({
      modules,
      moduleStatuses,
      moduleResults,
      toggleModule,
      toggleAutoStart,
      toggleTest,
      setModuleStatus,
      setModuleResults,
      updateModuleResults,
      clearModuleResults,
      getEnabledTests,
      getAllEnabledTests,
      resetToDefaults,
    }),
    [
      modules,
      moduleStatuses,
      moduleResults,
      toggleModule,
      toggleAutoStart,
      toggleTest,
      setModuleStatus,
      setModuleResults,
      updateModuleResults,
      clearModuleResults,
      getEnabledTests,
      getAllEnabledTests,
      resetToDefaults,
    ],
  );

  return <ModuleSettingsContext.Provider value={value}>{children}</ModuleSettingsContext.Provider>;
}

export function useModuleSettings(): ModuleSettingsContextValue {
  const context = useContext(ModuleSettingsContext);
  if (!context) {
    throw new Error('useModuleSettings must be used within a ModuleSettingsProvider');
  }
  return context;
}

export default ModuleSettingsContext;

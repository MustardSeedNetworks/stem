/**
 * @fileoverview The Stem - Module Selector Component
 * @description A component that displays tests organized by module (Reflector, Benchmark,
 *              ServiceTest, TrafficGen, Measure, Certify) with module-specific colors.
 *              Allows users to select tests using the module-oriented architecture.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { type ReactElement, useEffect, useState } from 'react';
import { HelpIcon } from './HelpIcon';

// Module definitions matching the Go backend
interface Module {
  name: string;
  displayName: string;
  description: string;
  color: string;
  standard: string;
  tests: string[];
}

// Test descriptions for tooltips
const testDescriptions: Record<string, { name: string; desc: string; tooltip: string }> = {
  // Benchmark (RFC 2544)
  throughput: {
    name: 'Throughput',
    desc: 'Max rate with 0% loss',
    tooltip: 'Find the maximum rate at which the DUT can forward frames with zero packet loss.',
  },
  latency: {
    name: 'Latency',
    desc: 'Round-trip time',
    tooltip: 'Measure round-trip packet delay at various loads and frame sizes.',
  },
  frame_loss: {
    name: 'Frame Loss',
    desc: 'Loss vs offered load',
    tooltip: 'Measure packet loss percentage across different offered load levels.',
  },
  back_to_back: {
    name: 'Back-to-Back',
    desc: 'Burst capacity',
    tooltip: 'Test maximum burst capacity - how many frames at line rate before drops.',
  },
  system_recovery: {
    name: 'System Recovery',
    desc: 'Recovery after overload',
    tooltip: 'Measure time to recover normal forwarding after sustained overload.',
  },
  reset: {
    name: 'Reset',
    desc: 'Device reset recovery',
    tooltip: 'Measure time from device restart to when it resumes forwarding.',
  },

  // ServiceTest (Y.1564 / MEF)
  y1564_config: {
    name: 'Y.1564 Config',
    desc: 'Service config validation',
    tooltip: 'Validate service at step loads (25%, 50%, 75%, 100% of CIR).',
  },
  y1564_perf: {
    name: 'Y.1564 Performance',
    desc: 'Sustained 15+ min test',
    tooltip: 'Extended duration test at full CIR to verify SLA compliance.',
  },
  y1564: {
    name: 'Y.1564 Full',
    desc: 'Both config and perf',
    tooltip: 'Complete Service Activation Test combining both phases.',
  },
  mef_config: {
    name: 'MEF Config',
    desc: 'MEF service step test',
    tooltip: 'MEF service configuration test - validates service at step loads.',
  },
  mef_perf: {
    name: 'MEF Performance',
    desc: 'MEF sustained test',
    tooltip: 'MEF service performance test - verifies SLA compliance.',
  },
  mef: {
    name: 'MEF Full',
    desc: 'Both MEF phases',
    tooltip: 'Complete MEF validation including configuration and performance.',
  },

  // Reflector (Tier 1 - standalone operational mode)
  reflect: {
    name: 'Reflect',
    desc: 'Packet reflection mode',
    tooltip: 'Echo received packets back to sender for remote device testing.',
  },

  // TrafficGen
  custom_stream: {
    name: 'Custom Stream',
    desc: 'Custom traffic generation',
    tooltip: 'Generate custom traffic patterns for specific testing needs.',
  },

  // Measure (Y.1731)
  y1731_delay: {
    name: 'Delay (DMM/DMR)',
    desc: 'Frame delay measurement',
    tooltip: 'Measure one-way and two-way frame delay using DMM/DMR OAM.',
  },
  y1731_loss: {
    name: 'Loss (LMM/LMR)',
    desc: 'Frame loss measurement',
    tooltip: 'Measure frame loss ratio using LMM/LMR OAM messages.',
  },
  y1731_slm: {
    name: 'Synthetic Loss',
    desc: 'SLM measurement',
    tooltip: 'Synthetic loss measurement using SLM/SLR frames.',
  },
  y1731_loopback: {
    name: 'Loopback',
    desc: 'LBM/LBR test',
    tooltip: 'Verify connectivity using OAM loopback messages.',
  },

  // Certify (RFC 2889, RFC 6349, TSN)
  rfc2889_forwarding: {
    name: 'Forwarding Rate',
    desc: 'Switch forwarding capacity',
    tooltip: 'Measure aggregate forwarding rate across all switch ports.',
  },
  rfc2889_caching: {
    name: 'Address Caching',
    desc: 'MAC table capacity',
    tooltip: 'Determine maximum MAC addresses the switch can learn.',
  },
  rfc2889_learning: {
    name: 'Address Learning',
    desc: 'Learning rate',
    tooltip: 'Measure how quickly the switch learns new MAC addresses.',
  },
  rfc2889_broadcast: {
    name: 'Broadcast',
    desc: 'Broadcast forwarding',
    tooltip: 'Test how the switch handles broadcast traffic flooding.',
  },
  rfc2889_congestion: {
    name: 'Congestion Control',
    desc: 'Backpressure handling',
    tooltip: 'Verify backpressure and flow control under congestion.',
  },
  rfc6349_throughput: {
    name: 'TCP Throughput',
    desc: 'BDP analysis',
    tooltip: 'Measure real TCP performance with Bandwidth-Delay Product.',
  },
  rfc6349_path: {
    name: 'Path Analysis',
    desc: 'RTT/Bandwidth',
    tooltip: 'Characterize network path properties including RTT and loss.',
  },
  tsn_timing: {
    name: 'Gate Timing',
    desc: 'GCL accuracy',
    tooltip: 'Verify Time-Aware Shaper gate control list timing accuracy.',
  },
  tsn_isolation: {
    name: 'Traffic Isolation',
    desc: 'Class isolation',
    tooltip: 'Verify traffic class separation and priority enforcement.',
  },
  tsn_latency: {
    name: 'Scheduled Latency',
    desc: 'Deterministic delay',
    tooltip: 'Measure deterministic latency for scheduled traffic.',
  },
  tsn: {
    name: 'TSN Full Suite',
    desc: 'All TSN tests',
    tooltip: 'Complete TSN validation including timing and isolation.',
  },
};

interface ModuleSelectorProps {
  selectedTests: string[];
  setSelectedTests: (tests: string[]) => void;
}

export function ModuleSelector({
  selectedTests,
  setSelectedTests,
}: ModuleSelectorProps): ReactElement {
  const [modules, setModules] = useState<Module[]>([]);
  const [expandedModule, setExpandedModule] = useState<string | null>('benchmark');
  const [loading, setLoading] = useState(true);

  // Fetch modules from API
  useEffect(() => {
    const fetchModules = async () => {
      try {
        const response = await fetch('/api/modules');
        if (response.ok) {
          const data = await response.json();
          setModules(data.modules || []);
        } else {
          // Use fallback static data if API unavailable
          setModules(getStaticModules());
        }
      } catch (_error) {
        // Use fallback static data
        setModules(getStaticModules());
      } finally {
        setLoading(false);
      }
    };
    fetchModules();
  }, []);

  const toggleTest = (test: string) => {
    if (selectedTests.includes(test)) {
      setSelectedTests(selectedTests.filter((t) => t !== test));
    } else {
      setSelectedTests([...selectedTests, test]);
    }
  };

  const toggleModule = (moduleName: string) => {
    setExpandedModule(expandedModule === moduleName ? null : moduleName);
  };

  const getSelectedCount = (mod: Module): number => {
    return mod.tests.filter((t) => selectedTests.includes(t)).length;
  };

  const selectAllInModule = (mod: Module) => {
    const newTests = [...selectedTests];
    for (const test of mod.tests) {
      if (!newTests.includes(test)) {
        newTests.push(test);
      }
    }
    setSelectedTests(newTests);
  };

  const deselectAllInModule = (mod: Module) => {
    setSelectedTests(selectedTests.filter((t) => !mod.tests.includes(t)));
  };

  if (loading) {
    return (
      <div className="text-center py-8 text-[var(--color-text-muted)]">Loading modules...</div>
    );
  }

  return (
    <div className="space-y-2">
      {modules.map((mod) => (
        <div
          key={mod.name}
          className="border border-[var(--color-surface-border)] rounded-lg overflow-hidden"
        >
          {/* Module Header */}
          <button
            type="button"
            onClick={() => toggleModule(mod.name)}
            className="w-full flex items-center justify-between p-3 hover:bg-[var(--color-surface-hover)] transition-colors"
            style={{ borderLeft: `4px solid ${mod.color}` }}
          >
            <div className="flex items-center gap-3">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: mod.color }}
                title={mod.displayName}
              />
              <div className="text-left">
                <div className="font-medium text-sm">{mod.displayName}</div>
                <div className="text-xs text-[var(--color-text-muted)]">{mod.standard}</div>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-[var(--color-text-muted)]">
                {getSelectedCount(mod)}/{mod.tests.length}
              </span>
              <svg
                className={`w-4 h-4 transition-transform ${expandedModule === mod.name ? 'rotate-180' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </div>
          </button>

          {/* Module Tests */}
          {expandedModule === mod.name && (
            <div className="border-t border-[var(--color-surface-border)] bg-[var(--color-surface-base)]">
              {/* Select All / Deselect All */}
              <div className="flex justify-end gap-2 px-3 py-2 border-b border-[var(--color-surface-border)]">
                <button
                  type="button"
                  onClick={() => selectAllInModule(mod)}
                  className="text-xs text-[var(--color-status-info)] hover:underline"
                >
                  Select All
                </button>
                <span className="text-[var(--color-text-muted)]">|</span>
                <button
                  type="button"
                  onClick={() => deselectAllInModule(mod)}
                  className="text-xs text-[var(--color-text-muted)] hover:underline"
                >
                  Deselect All
                </button>
              </div>

              {/* Test List */}
              <div className="p-2 space-y-1">
                {mod.tests.map((testId) => {
                  const testInfo = testDescriptions[testId] || {
                    name: testId,
                    desc: '',
                    tooltip: `Run ${testId} test`,
                  };
                  return (
                    <label
                      key={testId}
                      className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                    >
                      <input
                        type="checkbox"
                        checked={selectedTests.includes(testId)}
                        onChange={() => toggleTest(testId)}
                        className="mt-0.5 w-4 h-4"
                        style={{ accentColor: mod.color }}
                      />
                      <div className="flex-1">
                        <div className="font-medium text-sm flex items-center gap-1">
                          {testInfo.name}
                          <HelpIcon tooltip={testInfo.tooltip} />
                        </div>
                        {testInfo.desc && (
                          <div className="text-xs text-[var(--color-text-muted)]">
                            {testInfo.desc}
                          </div>
                        )}
                      </div>
                    </label>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

// Fallback static module data when API is unavailable
function getStaticModules(): Module[] {
  return [
    {
      name: 'reflector',
      displayName: 'Reflector',
      description: 'Packet reflection/loopback for remote device testing (Tier 1 mode)',
      color: '#0891b2',
      standard: 'Loopback/Echo',
      tests: ['reflect'],
    },
    {
      name: 'benchmark',
      displayName: 'Benchmark',
      description: 'RFC 2544 device benchmarking',
      color: '#dc2626',
      standard: 'RFC 2544',
      tests: ['throughput', 'latency', 'frame_loss', 'back_to_back', 'system_recovery', 'reset'],
    },
    {
      name: 'servicetest',
      displayName: 'ServiceTest',
      description: 'Y.1564 and MEF service activation',
      color: '#ea580c',
      standard: 'ITU-T Y.1564 / MEF',
      tests: ['y1564_config', 'y1564_perf', 'y1564', 'mef_config', 'mef_perf', 'mef'],
    },
    {
      name: 'trafficgen',
      displayName: 'TrafficGen',
      description: 'Custom traffic stream generation with configurable patterns',
      color: '#ca8a04',
      standard: 'Custom Traffic',
      tests: ['custom_stream'],
    },
    {
      name: 'measure',
      displayName: 'Measure',
      description: 'Y.1731 OAM performance measurement',
      color: '#2563eb',
      standard: 'ITU-T Y.1731',
      tests: ['y1731_delay', 'y1731_loss', 'y1731_slm', 'y1731_loopback'],
    },
    {
      name: 'certify',
      displayName: 'Certify',
      description: 'Compliance certification testing',
      color: '#16a34a',
      standard: 'RFC 2889 / RFC 6349 / TSN',
      tests: [
        'rfc2889_forwarding',
        'rfc2889_caching',
        'rfc2889_learning',
        'rfc2889_broadcast',
        'rfc2889_congestion',
        'rfc6349_throughput',
        'rfc6349_path',
        'tsn_timing',
        'tsn_isolation',
        'tsn_latency',
        'tsn',
      ],
    },
  ];
}

export default ModuleSelector;

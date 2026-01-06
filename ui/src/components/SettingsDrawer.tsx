/**
 * @fileoverview The Stem - Settings Drawer Component
 * @description The main configuration panel for the test suite. Allows users to:
 *              - Select operating mode (Reflector vs Test Master)
 *              - Configure network interfaces
 *              - Select and configure test suites (RFC 2544, Y.1564, RFC 2889, etc.)
 *              - Manage license activation
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import {
  Activity,
  Cpu,
  Grid,
  List,
  Monitor,
  Network,
  Radio,
  Settings2,
  X,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';
import { LicenseSection } from './LicenseSection';
import { ModuleSelector } from './ModuleSelector';
import { type RFC2544Config, RFC2544ConfigForm } from './RFC2544ConfigForm';
import { type RFC2889Config, RFC2889ConfigForm } from './RFC2889ConfigForm';
import { type RFC6349Config, RFC6349ConfigForm } from './RFC6349ConfigForm';
import { type TrafficGenConfig, TrafficGenConfigForm } from './TrafficGenConfigForm';
import { type TSNConfig, TSNConfigForm } from './TSNConfigForm';
import { type Y1564Config, Y1564ConfigForm } from './Y1564ConfigForm';
import { type Y1731Config, Y1731ConfigForm } from './Y1731ConfigForm';

interface InterfaceInfo {
  name: string;
  mac: string;
  speed: number;
  state: string;
  driver: string;
  physical: boolean;
  xdp: boolean;
  score: number;
}

interface SettingsDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  mode: 'reflector' | 'test_master';
  setMode: (mode: 'reflector' | 'test_master') => void;
  interfaces: InterfaceInfo[];
  selectedInterface: string;
  setSelectedInterface: (iface: string) => void;
  selectedTests: string[];
  setSelectedTests: (tests: string[]) => void;
  reflectorProfile: string;
  setReflectorProfile: (profile: string) => void;
  rfc2544Config: RFC2544Config;
  setRFC2544Config: (config: RFC2544Config) => void;
  rfc2889Config: RFC2889Config;
  setRFC2889Config: (config: RFC2889Config) => void;
  rfc6349Config: RFC6349Config;
  setRFC6349Config: (config: RFC6349Config) => void;
  y1564Config: Y1564Config;
  setY1564Config: (config: Y1564Config) => void;
  y1731Config: Y1731Config;
  setY1731Config: (config: Y1731Config) => void;
  tsnConfig: TSNConfig;
  setTSNConfig: (config: TSNConfig) => void;
  trafficGenConfig: TrafficGenConfig;
  setTrafficGenConfig: (config: TrafficGenConfig) => void;
}

export function SettingsDrawer({
  isOpen,
  onClose,
  mode,
  setMode,
  interfaces,
  selectedInterface,
  setSelectedInterface,
  selectedTests,
  setSelectedTests,
  reflectorProfile,
  setReflectorProfile,
  rfc2544Config,
  setRFC2544Config,
  rfc2889Config,
  setRFC2889Config,
  rfc6349Config,
  setRFC6349Config,
  y1564Config,
  setY1564Config,
  y1731Config,
  setY1731Config,
  tsnConfig,
  setTSNConfig,
  trafficGenConfig,
  setTrafficGenConfig,
}: SettingsDrawerProps) {
  const [viewMode, setViewMode] = useState<'standard' | 'module'>('standard');

  if (!isOpen) return null;

  const toggleTest = (test: string) => {
    if (selectedTests.includes(test)) {
      setSelectedTests(selectedTests.filter((t) => t !== test));
    } else {
      setSelectedTests([...selectedTests, test]);
    }
  };

  const maxScore = Math.max(...interfaces.map((i) => i.score), 0);

  return (
    <>
      {/* Backdrop */}
      <button
        type="button"
        className="fixed inset-0 bg-black/50 z-40 cursor-default"
        onClick={onClose}
        aria-label="Close settings drawer"
      />

      {/* Drawer */}
      <div className="fixed right-0 top-0 h-full w-96 max-w-full bg-[var(--color-surface-raised)] border-l border-[var(--color-surface-border)] z-50 overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-[var(--color-surface-raised)] border-b border-[var(--color-surface-border)] px-4 py-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-[var(--color-text-primary)]">Settings</h2>
          <button
            type="button"
            onClick={onClose}
            className="p-2 hover:bg-[var(--color-surface-hover)] rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-[var(--color-text-muted)]" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          {/* License Section */}
          <LicenseSection />

          {/* Mode Selection */}
          <CollapsibleSection
            title={
              <div className="flex items-center gap-2">
                <Monitor className="w-4 h-4" />
                <span>Mode</span>
              </div>
            }
            defaultOpen={true}
          >
            <div className="space-y-2">
              <label className="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]">
                <input
                  type="radio"
                  name="mode"
                  checked={mode === 'reflector'}
                  onChange={() => setMode('reflector')}
                  className="w-4 h-4 accent-[var(--color-stem-green)]"
                />
                <div>
                  <div className="font-medium text-sm">Reflector Mode</div>
                  <div className="text-xs text-[var(--color-text-muted)]">
                    Packet reflection (Tier 1)
                  </div>
                </div>
              </label>
              <label className="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]">
                <input
                  type="radio"
                  name="mode"
                  checked={mode === 'test_master'}
                  onChange={() => setMode('test_master')}
                  className="w-4 h-4 accent-[var(--color-stem-green)]"
                />
                <div>
                  <div className="font-medium text-sm">Test Master Mode</div>
                  <div className="text-xs text-[var(--color-text-muted)]">
                    Network testing (Tier 2)
                  </div>
                </div>
              </label>
            </div>
          </CollapsibleSection>

          {/* Interface Selection */}
          <CollapsibleSection
            title={
              <div className="flex items-center gap-2">
                <Network className="w-4 h-4" />
                <span>Interface</span>
              </div>
            }
            defaultOpen={true}
          >
            <div className="space-y-3">
              <select
                value={selectedInterface}
                onChange={(e) => setSelectedInterface(e.target.value)}
                className="w-full"
              >
                {interfaces.map((iface) => (
                  <option key={iface.name} value={iface.name}>
                    {iface.name} ({iface.speed}Mbps)
                    {iface.score === maxScore && ' ★'}
                  </option>
                ))}
              </select>
              {interfaces.find((i) => i.name === selectedInterface) && (
                <div className="text-xs text-[var(--color-text-muted)] space-y-1">
                  <div>Driver: {interfaces.find((i) => i.name === selectedInterface)?.driver}</div>
                  <div className="flex gap-2">
                    {interfaces.find((i) => i.name === selectedInterface)?.xdp && (
                      <span className="status-badge success">XDP</span>
                    )}
                    {interfaces.find((i) => i.name === selectedInterface)?.physical && (
                      <span className="status-badge success">Physical</span>
                    )}
                  </div>
                </div>
              )}
            </div>
          </CollapsibleSection>

          {/* Test Master Tests */}
          {mode === 'test_master' && (
            <>
              {/* View Toggle */}
              <div className="flex items-center justify-between p-2 bg-[var(--color-surface-base)] rounded-lg mb-2">
                <span className="text-sm text-[var(--color-text-muted)]">View by:</span>
                <div className="flex rounded-lg overflow-hidden border border-[var(--color-surface-border)]">
                  <button
                    type="button"
                    onClick={() => setViewMode('standard')}
                    className={`flex items-center gap-1 px-3 py-1.5 text-xs ${
                      viewMode === 'standard'
                        ? 'bg-[var(--color-stem-green)] text-white'
                        : 'bg-[var(--color-surface-raised)] text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)]'
                    }`}
                  >
                    <List className="w-3 h-3" />
                    Standard
                  </button>
                  <button
                    type="button"
                    onClick={() => setViewMode('module')}
                    className={`flex items-center gap-1 px-3 py-1.5 text-xs ${
                      viewMode === 'module'
                        ? 'bg-[var(--color-stem-green)] text-white'
                        : 'bg-[var(--color-surface-raised)] text-[var(--color-text-muted)] hover:bg-[var(--color-surface-hover)]'
                    }`}
                  >
                    <Grid className="w-3 h-3" />
                    Module
                  </button>
                </div>
              </div>

              {/* Module View */}
              {viewMode === 'module' && (
                <ModuleSelector selectedTests={selectedTests} setSelectedTests={setSelectedTests} />
              )}

              {/* Standard View */}
              {viewMode === 'standard' && (
                <>
                  {/* RFC 2544 Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Zap className="w-4 h-4" />
                        <span>RFC 2544 Tests</span>
                        <span className="text-xs text-[var(--color-text-muted)]">
                          ({selectedTests.filter((t) => t.startsWith('rfc2544')).length}/6)
                        </span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'rfc2544_throughput',
                          name: 'Throughput',
                          desc: 'Max rate with 0% loss',
                          tooltip:
                            'Find the maximum rate at which the DUT can forward frames with zero packet loss using binary search.',
                        },
                        {
                          id: 'rfc2544_latency',
                          name: 'Latency',
                          desc: 'Round-trip time',
                          tooltip:
                            'Measure round-trip packet delay at various loads and frame sizes.',
                        },
                        {
                          id: 'rfc2544_frame_loss',
                          name: 'Frame Loss',
                          desc: 'Loss vs offered load',
                          tooltip:
                            'Measure packet loss percentage across different offered load levels.',
                        },
                        {
                          id: 'rfc2544_back_to_back',
                          name: 'Back-to-Back',
                          desc: 'Burst capacity',
                          tooltip:
                            'Test maximum burst capacity - how many frames at line rate before drops occur.',
                        },
                        {
                          id: 'rfc2544_system_recovery',
                          name: 'System Recovery',
                          desc: 'Recovery after overload',
                          tooltip:
                            'Measure time to recover normal forwarding after sustained overload condition.',
                        },
                        {
                          id: 'rfc2544_reset',
                          name: 'Reset',
                          desc: 'Device reset recovery',
                          tooltip:
                            'Measure time from device restart to when it resumes forwarding traffic.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* Y.1564 Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Activity className="w-4 h-4" />
                        <span>Y.1564 / EtherSAM</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'y1564_config',
                          name: 'Configuration Test',
                          desc: 'Service config validation',
                          tooltip:
                            'Validate service at step loads (25%, 50%, 75%, 100% of CIR) with quick pass/fail.',
                        },
                        {
                          id: 'y1564_perf',
                          name: 'Performance Test',
                          desc: 'Sustained 15+ min test',
                          tooltip:
                            'Extended duration test at full CIR to verify SLA compliance over time.',
                        },
                        {
                          id: 'y1564_full',
                          name: 'Full Test',
                          desc: 'Both config and perf',
                          tooltip:
                            'Complete Service Activation Test combining both configuration and performance phases.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* RFC 2889 Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Cpu className="w-4 h-4" />
                        <span>RFC 2889 LAN Switch</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'rfc2889_forwarding',
                          name: 'Forwarding Rate',
                          desc: 'Switch forwarding capacity',
                          tooltip:
                            'Measure aggregate forwarding rate across all ports of a LAN switch.',
                        },
                        {
                          id: 'rfc2889_caching',
                          name: 'Address Caching',
                          desc: 'MAC table capacity',
                          tooltip:
                            'Determine maximum number of MAC addresses the switch can learn and forward.',
                        },
                        {
                          id: 'rfc2889_learning',
                          name: 'Address Learning',
                          desc: 'Learning rate',
                          tooltip: 'Measure how quickly the switch learns new MAC addresses.',
                        },
                        {
                          id: 'rfc2889_broadcast',
                          name: 'Broadcast',
                          desc: 'Broadcast forwarding',
                          tooltip: 'Test how the switch handles broadcast traffic flooding.',
                        },
                        {
                          id: 'rfc2889_congestion',
                          name: 'Congestion Control',
                          desc: 'Backpressure handling',
                          tooltip: 'Verify backpressure and flow control under congestion.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* RFC 6349 Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Activity className="w-4 h-4" />
                        <span>RFC 6349 TCP</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'rfc6349_throughput',
                          name: 'TCP Throughput',
                          desc: 'BDP analysis',
                          tooltip:
                            'Measure real TCP performance with Bandwidth-Delay Product optimization.',
                        },
                        {
                          id: 'rfc6349_path',
                          name: 'Path Analysis',
                          desc: 'RTT/Bandwidth',
                          tooltip:
                            'Characterize network path properties including RTT, loss, and capacity.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* Y.1731 Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Radio className="w-4 h-4" />
                        <span>Y.1731 OAM</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'y1731_delay',
                          name: 'Delay (DMM/DMR)',
                          desc: 'Frame delay measurement',
                          tooltip:
                            'Measure one-way and two-way frame delay using DMM/DMR OAM messages.',
                        },
                        {
                          id: 'y1731_loss',
                          name: 'Loss (LMM/LMR)',
                          desc: 'Frame loss measurement',
                          tooltip: 'Measure frame loss ratio using LMM/LMR OAM messages.',
                        },
                        {
                          id: 'y1731_slm',
                          name: 'Synthetic Loss',
                          desc: 'SLM measurement',
                          tooltip: 'Synthetic loss measurement using SLM/SLR frames.',
                        },
                        {
                          id: 'y1731_loopback',
                          name: 'Loopback',
                          desc: 'LBM/LBR test',
                          tooltip: 'Verify connectivity using OAM loopback messages (LBM/LBR).',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* MEF Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Settings2 className="w-4 h-4" />
                        <span>MEF Service</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'mef_config',
                          name: 'Configuration',
                          desc: 'Service step test',
                          tooltip:
                            'MEF service configuration test - validates service at step loads.',
                        },
                        {
                          id: 'mef_perf',
                          name: 'Performance',
                          desc: 'Sustained test',
                          tooltip:
                            'MEF service performance test - verifies SLA compliance over time.',
                        },
                        {
                          id: 'mef_full',
                          name: 'Full Test',
                          desc: 'Both phases',
                          tooltip:
                            'Complete MEF validation including both configuration and performance.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* TSN Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Cpu className="w-4 h-4" />
                        <span>TSN 802.1Qbv</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'tsn_timing',
                          name: 'Gate Timing',
                          desc: 'GCL accuracy',
                          tooltip:
                            'Verify Time-Aware Shaper (TAS) gate control list timing accuracy.',
                        },
                        {
                          id: 'tsn_isolation',
                          name: 'Traffic Isolation',
                          desc: 'Class isolation',
                          tooltip: 'Verify traffic class separation and priority enforcement.',
                        },
                        {
                          id: 'tsn_latency',
                          name: 'Scheduled Latency',
                          desc: 'Deterministic delay',
                          tooltip: 'Measure deterministic latency for scheduled traffic flows.',
                        },
                        {
                          id: 'tsn_full',
                          name: 'Full Suite',
                          desc: 'All TSN tests',
                          tooltip:
                            'Complete TSN validation including timing, isolation, and latency.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>

                  {/* Traffic Generator Tests */}
                  <CollapsibleSection
                    title={
                      <div className="flex items-center gap-2">
                        <Radio className="w-4 h-4" />
                        <span>Traffic Generator</span>
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      {[
                        {
                          id: 'custom_stream',
                          name: 'Custom Stream',
                          desc: 'Configurable traffic',
                          tooltip:
                            'Generate custom traffic patterns with configurable frame size, rate, and duration.',
                        },
                        {
                          id: 'trafficgen_burst',
                          name: 'Burst Mode',
                          desc: 'Burst traffic generation',
                          tooltip: 'Generate burst traffic with configurable burst size and gap.',
                        },
                        {
                          id: 'trafficgen_multistream',
                          name: 'Multi-Stream',
                          desc: 'Multiple traffic streams',
                          tooltip:
                            'Generate multiple concurrent traffic streams with different parameters.',
                        },
                      ].map((test) => (
                        <label
                          key={test.id}
                          className="flex items-start gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                        >
                          <input
                            type="checkbox"
                            checked={selectedTests.includes(test.id)}
                            onChange={() => toggleTest(test.id)}
                            className="mt-0.5 w-4 h-4 accent-[var(--color-stem-green)]"
                          />
                          <div className="flex-1">
                            <div className="font-medium text-sm flex items-center gap-1">
                              {test.name}
                              <HelpIcon tooltip={test.tooltip} />
                            </div>
                            <div className="text-xs text-[var(--color-text-muted)]">
                              {test.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </CollapsibleSection>
                </>
              )}

              {/* RFC 2544 Advanced Configuration */}
              <RFC2544ConfigForm
                config={rfc2544Config}
                setConfig={setRFC2544Config}
                selectedTests={selectedTests}
              />

              {/* Y.1564/MEF Advanced Configuration */}
              <Y1564ConfigForm
                config={y1564Config}
                setConfig={setY1564Config}
                selectedTests={selectedTests}
              />

              {/* RFC 2889 LAN Switch Configuration */}
              <RFC2889ConfigForm
                config={rfc2889Config}
                setConfig={setRFC2889Config}
                selectedTests={selectedTests}
              />

              {/* RFC 6349 TCP Configuration */}
              <RFC6349ConfigForm
                config={rfc6349Config}
                setConfig={setRFC6349Config}
                selectedTests={selectedTests}
              />

              {/* Y.1731 OAM Configuration */}
              <Y1731ConfigForm
                config={y1731Config}
                setConfig={setY1731Config}
                selectedTests={selectedTests}
              />

              {/* TSN Configuration */}
              <TSNConfigForm
                config={tsnConfig}
                setConfig={setTSNConfig}
                selectedTests={selectedTests}
              />

              {/* Traffic Generator Configuration */}
              <TrafficGenConfigForm
                config={trafficGenConfig}
                setConfig={setTrafficGenConfig}
                selectedTests={selectedTests}
              />
            </>
          )}

          {/* Reflector Config */}
          {mode === 'reflector' && (
            <CollapsibleSection
              title={
                <div className="flex items-center gap-2">
                  <Settings2 className="w-4 h-4" />
                  <span>Reflector Profile</span>
                </div>
              }
              defaultOpen={true}
            >
              <div className="space-y-2">
                {[
                  { id: 'netally', name: 'NetAlly', desc: 'ITO signatures only' },
                  { id: 'msn', name: 'MSN', desc: 'Mustard Seed signatures' },
                  { id: 'all', name: 'All', desc: 'All signature types' },
                  { id: 'custom', name: 'Custom', desc: 'Manual configuration' },
                ].map((profile) => (
                  <label
                    key={profile.id}
                    className="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
                  >
                    <input
                      type="radio"
                      name="profile"
                      checked={reflectorProfile === profile.id}
                      onChange={() => setReflectorProfile(profile.id)}
                      className="w-4 h-4 accent-[var(--color-stem-green)]"
                    />
                    <div>
                      <div className="font-medium text-sm">{profile.name}</div>
                      <div className="text-xs text-[var(--color-text-muted)]">{profile.desc}</div>
                    </div>
                  </label>
                ))}
              </div>
            </CollapsibleSection>
          )}
        </div>
      </div>
    </>
  );
}

export default SettingsDrawer;

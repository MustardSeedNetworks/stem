/**
 * @fileoverview The Stem - TSN Test Configuration
 * @description Configuration form for IEEE 802.1 Time-Sensitive Networking Testing.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { Clock, Info } from 'lucide-react';
import type { ReactElement } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** TSN test configuration parameters */
export interface TSNConfig {
  /** Test duration in seconds */
  duration: number;
  /** Warmup duration in seconds */
  warmup: number;
  /** Frame size in bytes */
  frameSize: number;
  /** Maximum acceptable latency in nanoseconds */
  maxLatencyNs: number;
  /** Maximum acceptable jitter in nanoseconds */
  maxJitterNs: number;
  /** Require PTP synchronization */
  requirePTPSync: boolean;
  /** Maximum acceptable PTP sync offset in nanoseconds */
  maxSyncOffsetNs: number;
  /** Enable PTP timestamping */
  ptpEnabled: boolean;
  /** Enable frame preemption (802.1Qbu) */
  preemptionEnabled: boolean;
  /** Number of traffic classes */
  numTrafficClasses: number;
  /** Base time for scheduling (nanoseconds since epoch) */
  baseTimeNs: number;
  /** Cycle time in nanoseconds */
  cycleTimeNs: number;
  /** Traffic class for test frames (0-7) */
  trafficClass: number;
}

/** Default TSN configuration */
export const defaultTSNConfig: TSNConfig = {
  duration: 60,
  warmup: 5,
  frameSize: 64,
  maxLatencyNs: 1000000, // 1ms
  maxJitterNs: 100000, // 100us
  requirePTPSync: true,
  maxSyncOffsetNs: 1000, // 1us
  ptpEnabled: true,
  preemptionEnabled: false,
  numTrafficClasses: 8,
  baseTimeNs: 0,
  cycleTimeNs: 1000000, // 1ms cycle
  trafficClass: 7,
};

/** Frame size options */
const FRAME_SIZE_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 64, label: '64 B (min)' },
  { value: 128, label: '128 B' },
  { value: 256, label: '256 B' },
  { value: 512, label: '512 B' },
  { value: 1024, label: '1024 B' },
  { value: 1518, label: '1518 B (max)' },
];

/** Common cycle times */
const CYCLE_TIME_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 125000, label: '125 us' },
  { value: 250000, label: '250 us' },
  { value: 500000, label: '500 us' },
  { value: 1000000, label: '1 ms' },
  { value: 2000000, label: '2 ms' },
  { value: 4000000, label: '4 ms' },
];

interface TSNConfigFormProps {
  config: TSNConfig;
  setConfig: (config: TSNConfig) => void;
  selectedTests: string[];
}

// Format nanoseconds for display
function formatNs(ns: number): string {
  if (ns >= 1000000000) {
    return `${(ns / 1000000000).toFixed(1)} s`;
  }
  if (ns >= 1000000) {
    return `${(ns / 1000000).toFixed(1)} ms`;
  }
  if (ns >= 1000) {
    return `${(ns / 1000).toFixed(1)} us`;
  }
  return `${ns} ns`;
}

/** Props for TestParametersSection */
interface TestParametersSectionProps {
  config: TSNConfig;
  updateConfig: (updates: Partial<TSNConfig>) => void;
}

/** Test parameters section (duration, warmup, frame size) */
function TestParametersSection({ config, updateConfig }: TestParametersSectionProps): ReactElement {
  return (
    <div className="space-y-3">
      <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
        Test Parameters
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div>
          <label
            htmlFor="tsn-duration"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Duration (s)
            <HelpIcon tooltip="Test duration in seconds." />
          </label>
          <input
            id="tsn-duration"
            type="number"
            min={10}
            max={3600}
            step={1}
            value={config.duration}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ duration: Number(e.target.value) })
            }
            className="mt-1 w-full"
          />
        </div>

        <div>
          <label
            htmlFor="tsn-warmup"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Warmup (s)
            <HelpIcon tooltip="Warmup period for synchronization." />
          </label>
          <input
            id="tsn-warmup"
            type="number"
            min={0}
            max={60}
            step={1}
            value={config.warmup}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ warmup: Number(e.target.value) })
            }
            className="mt-1 w-full"
          />
        </div>

        <div>
          <label
            htmlFor="tsn-framesize"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Frame Size
            <HelpIcon tooltip="Ethernet frame size for testing." />
          </label>
          <select
            id="tsn-framesize"
            value={config.frameSize}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
              updateConfig({ frameSize: Number(e.target.value) })
            }
            className="mt-1 w-full"
          >
            {FRAME_SIZE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
}

/** Props for TimingRequirementsSection */
interface TimingRequirementsSectionProps {
  config: TSNConfig;
  updateConfig: (updates: Partial<TSNConfig>) => void;
}

/** Timing requirements section (max latency, max jitter) */
function TimingRequirementsSection({
  config,
  updateConfig,
}: TimingRequirementsSectionProps): ReactElement {
  return (
    <div className="space-y-3">
      <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
        Timing Requirements
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div>
          <label
            htmlFor="tsn-maxlatency"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Max Latency (us)
            <HelpIcon tooltip="Maximum acceptable end-to-end latency." />
          </label>
          <input
            id="tsn-maxlatency"
            type="number"
            min={1}
            max={100000}
            step={1}
            value={config.maxLatencyNs / 1000}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ maxLatencyNs: Number(e.target.value) * 1000 })
            }
            className="mt-1 w-full"
          />
        </div>

        <div>
          <label
            htmlFor="tsn-maxjitter"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Max Jitter (us)
            <HelpIcon tooltip="Maximum acceptable jitter (PDV)." />
          </label>
          <input
            id="tsn-maxjitter"
            type="number"
            min={1}
            max={10000}
            step={1}
            value={config.maxJitterNs / 1000}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ maxJitterNs: Number(e.target.value) * 1000 })
            }
            className="mt-1 w-full"
          />
        </div>
      </div>
    </div>
  );
}

/** Props for PTPConfigSection */
interface PTPConfigSectionProps {
  config: TSNConfig;
  updateConfig: (updates: Partial<TSNConfig>) => void;
}

/** PTP synchronization configuration section */
function PTPConfigSection({ config, updateConfig }: PTPConfigSectionProps): ReactElement {
  return (
    <div className="space-y-3">
      <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
        PTP Synchronization
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <input
            id="tsn-ptpenabled"
            type="checkbox"
            checked={config.ptpEnabled}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ ptpEnabled: e.target.checked })
            }
            className="rounded border-[var(--color-surface-border)]"
          />
          <label htmlFor="tsn-ptpenabled" className="text-sm text-[var(--color-text-primary)]">
            Enable PTP timestamping (IEEE 1588)
          </label>
        </div>

        <div className="flex items-center gap-2">
          <input
            id="tsn-requiresync"
            type="checkbox"
            checked={config.requirePTPSync}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ requirePTPSync: e.target.checked })
            }
            className="rounded border-[var(--color-surface-border)]"
          />
          <label htmlFor="tsn-requiresync" className="text-sm text-[var(--color-text-primary)]">
            Require PTP synchronization before test
          </label>
        </div>
      </div>

      {config.ptpEnabled ? (
        <div>
          <label
            htmlFor="tsn-syncoffset"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Max Sync Offset (ns)
            <HelpIcon tooltip="Maximum acceptable PTP clock offset." />
          </label>
          <input
            id="tsn-syncoffset"
            type="number"
            min={1}
            max={1000000}
            step={1}
            value={config.maxSyncOffsetNs}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ maxSyncOffsetNs: Number(e.target.value) })
            }
            className="mt-1 w-full"
          />
        </div>
      ) : null}
    </div>
  );
}

/** Props for SchedulingConfigSection */
interface SchedulingConfigSectionProps {
  config: TSNConfig;
  updateConfig: (updates: Partial<TSNConfig>) => void;
}

/** Traffic scheduling configuration section (802.1Qbv) */
function SchedulingConfigSection({
  config,
  updateConfig,
}: SchedulingConfigSectionProps): ReactElement {
  return (
    <div className="space-y-3">
      <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
        Traffic Scheduling (802.1Qbv)
      </div>

      <div className="flex items-center gap-2">
        <input
          id="tsn-preemption"
          type="checkbox"
          checked={config.preemptionEnabled}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
            updateConfig({ preemptionEnabled: e.target.checked })
          }
          className="rounded border-[var(--color-surface-border)]"
        />
        <label htmlFor="tsn-preemption" className="text-sm text-[var(--color-text-primary)]">
          Enable frame preemption (802.1Qbu)
        </label>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div>
          <label
            htmlFor="tsn-cycletime"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Cycle Time
            <HelpIcon tooltip="Time-Aware Shaper cycle duration." />
          </label>
          <select
            id="tsn-cycletime"
            value={config.cycleTimeNs}
            onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
              updateConfig({ cycleTimeNs: Number(e.target.value) })
            }
            className="mt-1 w-full"
          >
            {CYCLE_TIME_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label
            htmlFor="tsn-trafficclass"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Traffic Class
            <HelpIcon tooltip="Traffic class for test frames (0-7)." />
          </label>
          <input
            id="tsn-trafficclass"
            type="number"
            min={0}
            max={7}
            step={1}
            value={config.trafficClass}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ trafficClass: Number(e.target.value) })
            }
            className="mt-1 w-full"
          />
        </div>
      </div>

      <div>
        <label
          htmlFor="tsn-numclasses"
          className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
        >
          Number of Traffic Classes
          <HelpIcon tooltip="Total traffic classes configured (1-8)." />
        </label>
        <input
          id="tsn-numclasses"
          type="number"
          min={1}
          max={8}
          step={1}
          value={config.numTrafficClasses}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
            updateConfig({ numTrafficClasses: Number(e.target.value) })
          }
          className="mt-1 w-full"
        />
      </div>

      <div>
        <label
          htmlFor="tsn-basetime"
          className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
        >
          Base Time (ns)
          <HelpIcon tooltip="Base time for gate schedule (ns since epoch). 0 = use current time." />
        </label>
        <input
          id="tsn-basetime"
          type="number"
          min={0}
          step={1}
          value={config.baseTimeNs}
          onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
            updateConfig({ baseTimeNs: Number(e.target.value) })
          }
          className="mt-1 w-full"
        />
      </div>
    </div>
  );
}

/** Props for TestSummarySection */
interface TestSummarySectionProps {
  config: TSNConfig;
  hasLatency: boolean;
  hasJitter: boolean;
  hasSync: boolean;
  hasPreemption: boolean;
  hasScheduling: boolean;
}

/** Test summary section displaying configured values */
function TestSummarySection({
  config,
  hasLatency,
  hasJitter,
  hasSync,
  hasPreemption,
  hasScheduling,
}: TestSummarySectionProps): ReactElement {
  const selectedTestNames = [
    hasLatency && 'Latency',
    hasJitter && 'Jitter',
    hasSync && 'Sync Verification',
    hasPreemption && 'Preemption',
    hasScheduling && 'Scheduling',
  ].filter(Boolean);

  return (
    <div className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
      <div className="flex items-center gap-2 text-sm font-medium text-[var(--color-text-primary)] mb-2">
        <Info className="w-4 h-4" />
        Test Summary
      </div>
      <div className="text-xs text-[var(--color-text-muted)] space-y-1">
        <div>Selected tests: {selectedTestNames.join(', ')}</div>
        <div>Frame size: {config.frameSize} bytes</div>
        <div>
          Timing: &le;{formatNs(config.maxLatencyNs)} latency | &le;{formatNs(config.maxJitterNs)}{' '}
          jitter
        </div>
        <div>
          PTP: {config.ptpEnabled ? 'Enabled' : 'Disabled'}
          {config.ptpEnabled && config.requirePTPSync ? ' (required)' : ''}
          {config.ptpEnabled ? ` | Max offset: ${formatNs(config.maxSyncOffsetNs)}` : ''}
        </div>
        {hasScheduling || hasPreemption ? (
          <div>
            Scheduling: Cycle {formatNs(config.cycleTimeNs)} | TC {config.trafficClass}
            {config.preemptionEnabled ? ' | Preemption enabled' : ''}
          </div>
        ) : null}
        <div>
          Duration: {config.duration}s + {config.warmup}s warmup
        </div>
      </div>
    </div>
  );
}

export function TSNConfigForm({
  config,
  setConfig,
  selectedTests,
}: TSNConfigFormProps): ReactElement | null {
  const hasTSNTests = selectedTests.some((t) => t.startsWith('tsn_'));

  if (!hasTSNTests) {
    return null;
  }

  const updateConfig = (updates: Partial<TSNConfig>): void => {
    setConfig({ ...config, ...updates });
  };

  const hasLatency = selectedTests.includes('tsn_latency');
  const hasJitter = selectedTests.includes('tsn_jitter');
  const hasSync = selectedTests.includes('tsn_sync');
  const hasPreemption = selectedTests.includes('tsn_preemption');
  const hasScheduling = selectedTests.includes('tsn_scheduling');

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4" />
          <span>TSN Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        <TestParametersSection config={config} updateConfig={updateConfig} />
        <TimingRequirementsSection config={config} updateConfig={updateConfig} />
        <PTPConfigSection config={config} updateConfig={updateConfig} />

        {hasScheduling || hasPreemption ? (
          <SchedulingConfigSection config={config} updateConfig={updateConfig} />
        ) : null}

        <TestSummarySection
          config={config}
          hasLatency={hasLatency}
          hasJitter={hasJitter}
          hasSync={hasSync}
          hasPreemption={hasPreemption}
          hasScheduling={hasScheduling}
        />
      </div>
    </CollapsibleSection>
  );
}

export default TSNConfigForm;

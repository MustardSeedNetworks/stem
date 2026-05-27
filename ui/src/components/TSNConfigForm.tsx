/**
 * @fileoverview The Stem - TSN Test Configuration
 * @description Migrated to react-hook-form + valibot per #325. Uses
 *              FormProvider/useFormContext so the sub-component
 *              decomposition (test params / timing / PTP / scheduling /
 *              summary) doesn't have to thread the form instance
 *              through props.
 */

import { AlertTriangle, Clock, Info } from 'lucide-react';
import type { ReactElement } from 'react';
import { FormProvider, useFormContext } from 'react-hook-form';
import { useConfigForm } from '../forms/useConfigForm';
import { TSNConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** TSN test configuration parameters */
export interface TSNConfig {
  duration: number;
  warmup: number;
  frameSize: number;
  maxLatencyNs: number;
  maxJitterNs: number;
  requirePTPSync: boolean;
  maxSyncOffsetNs: number;
  ptpEnabled: boolean;
  preemptionEnabled: boolean;
  numTrafficClasses: number;
  baseTimeNs: number;
  cycleTimeNs: number;
  trafficClass: number;
}

/** Default TSN configuration */
export const defaultTSNConfig: TSNConfig = {
  duration: 60,
  warmup: 5,
  frameSize: 64,
  maxLatencyNs: 1000000,
  maxJitterNs: 100000,
  requirePTPSync: true,
  maxSyncOffsetNs: 1000,
  ptpEnabled: true,
  preemptionEnabled: false,
  numTrafficClasses: 8,
  baseTimeNs: 0,
  cycleTimeNs: 1000000,
  trafficClass: 7,
};

const FRAME_SIZE_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 64, label: '64 B (min)' },
  { value: 128, label: '128 B' },
  { value: 256, label: '256 B' },
  { value: 512, label: '512 B' },
  { value: 1024, label: '1024 B' },
  { value: 1518, label: '1518 B (max)' },
];

const CYCLE_TIME_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 125000, label: '125 us' },
  { value: 250000, label: '250 us' },
  { value: 500000, label: '500 us' },
  { value: 1000000, label: '1 ms' },
  { value: 2000000, label: '2 ms' },
  { value: 4000000, label: '4 ms' },
];

function formatNs(ns: number): string {
  if (ns >= 1000000000) return `${(ns / 1000000000).toFixed(1)} s`;
  if (ns >= 1000000) return `${(ns / 1000000).toFixed(1)} ms`;
  if (ns >= 1000) return `${(ns / 1000).toFixed(1)} us`;
  return `${ns} ns`;
}

function FieldError({ message }: { message?: string }): ReactElement | null {
  if (!message) return null;
  return (
    <div className="mt-tight text-xs text-status-danger flex items-center gap-tight">
      <AlertTriangle className="w-3 h-3" />
      {message}
    </div>
  );
}

function TestParametersSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  return (
    <div className="stack">
      <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
        Test Parameters
      </div>
      <div className="grid grid-cols-3 gap-default">
        <div>
          <label htmlFor="tsn-duration" className="flex items-center gap-tight label">
            Duration (s)
            <HelpIcon tooltip="Test duration in seconds." />
          </label>
          <input
            id="tsn-duration"
            type="number"
            step={1}
            {...register('duration', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.duration?.message} />
        </div>
        <div>
          <label htmlFor="tsn-warmup" className="flex items-center gap-tight label">
            Warmup (s)
            <HelpIcon tooltip="Warmup period for synchronization." />
          </label>
          <input
            id="tsn-warmup"
            type="number"
            step={1}
            {...register('warmup', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.warmup?.message} />
        </div>
        <div>
          <label htmlFor="tsn-framesize" className="flex items-center gap-tight label">
            Frame Size
            <HelpIcon tooltip="Ethernet frame size for testing." />
          </label>
          <select
            id="tsn-framesize"
            {...register('frameSize', { valueAsNumber: true })}
            className="mt-tight w-full"
          >
            {FRAME_SIZE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <FieldError message={errors.frameSize?.message} />
        </div>
      </div>
    </div>
  );
}

function TimingRequirementsSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  return (
    <div className="stack">
      <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
        Timing Requirements
      </div>
      <div className="grid grid-cols-2 gap-default">
        <div>
          <label htmlFor="tsn-maxlatency" className="flex items-center gap-tight label">
            Max Latency (ns)
            <HelpIcon tooltip="Maximum acceptable end-to-end latency in nanoseconds." />
          </label>
          <input
            id="tsn-maxlatency"
            type="number"
            step={1000}
            {...register('maxLatencyNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.maxLatencyNs?.message} />
        </div>
        <div>
          <label htmlFor="tsn-maxjitter" className="flex items-center gap-tight label">
            Max Jitter (ns)
            <HelpIcon tooltip="Maximum acceptable jitter (PDV) in nanoseconds." />
          </label>
          <input
            id="tsn-maxjitter"
            type="number"
            step={1000}
            {...register('maxJitterNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.maxJitterNs?.message} />
        </div>
      </div>
    </div>
  );
}

function PTPConfigSection(): ReactElement {
  const {
    register,
    watch,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const ptpEnabled = watch('ptpEnabled');
  return (
    <div className="stack">
      <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
        PTP Synchronization
      </div>
      <div className="stack-sm">
        <div className="flex items-center gap-compact">
          <input
            id="tsn-ptpenabled"
            type="checkbox"
            {...register('ptpEnabled')}
            aria-label="Enable IEEE 1588 PTP hardware timestamping"
            className="rounded border-surface-border"
          />
          <label
            htmlFor="tsn-ptpenabled"
            title="Use IEEE 1588 PTP hardware timestamps for sub-microsecond delay measurement; requires PTP-capable NIC"
            className="text-sm text-text-primary"
          >
            Enable PTP timestamping (IEEE 1588)
          </label>
        </div>
        <div className="flex items-center gap-compact">
          <input
            id="tsn-requiresync"
            type="checkbox"
            {...register('requirePTPSync')}
            aria-label="Require PTP synchronization before starting test"
            className="rounded border-surface-border"
          />
          <label
            htmlFor="tsn-requiresync"
            title="Block the test from starting until the local PTP clock has locked to the grandmaster within the configured tolerance"
            className="text-sm text-text-primary"
          >
            Require PTP synchronization before test
          </label>
        </div>
      </div>
      {ptpEnabled ? (
        <div>
          <label htmlFor="tsn-syncoffset" className="flex items-center gap-tight label">
            Max Sync Offset (ns)
            <HelpIcon tooltip="Maximum acceptable PTP clock offset." />
          </label>
          <input
            id="tsn-syncoffset"
            type="number"
            step={1}
            {...register('maxSyncOffsetNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.maxSyncOffsetNs?.message} />
        </div>
      ) : null}
    </div>
  );
}

function SchedulingConfigSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  return (
    <div className="stack">
      <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
        Traffic Scheduling (802.1Qbv)
      </div>
      <div className="flex items-center gap-compact">
        <input
          id="tsn-preemption"
          type="checkbox"
          {...register('preemptionEnabled')}
          aria-label="Enable IEEE 802.1Qbu frame preemption"
          className="rounded border-surface-border"
        />
        <label
          htmlFor="tsn-preemption"
          title="Allow express traffic to preempt in-flight preemptable frames per IEEE 802.1Qbu; reduces latency for critical traffic classes"
          className="text-sm text-text-primary"
        >
          Enable frame preemption (802.1Qbu)
        </label>
      </div>
      <div className="grid grid-cols-2 gap-default">
        <div>
          <label htmlFor="tsn-cycletime" className="flex items-center gap-tight label">
            Cycle Time
            <HelpIcon tooltip="Time-Aware Shaper cycle duration." />
          </label>
          <select
            id="tsn-cycletime"
            {...register('cycleTimeNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          >
            {CYCLE_TIME_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <FieldError message={errors.cycleTimeNs?.message} />
        </div>
        <div>
          <label htmlFor="tsn-trafficclass" className="flex items-center gap-tight label">
            Traffic Class
            <HelpIcon tooltip="Traffic class for test frames (0-7)." />
          </label>
          <input
            id="tsn-trafficclass"
            type="number"
            step={1}
            {...register('trafficClass', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.trafficClass?.message} />
        </div>
      </div>
      <div>
        <label htmlFor="tsn-numclasses" className="flex items-center gap-tight label">
          Number of Traffic Classes
          <HelpIcon tooltip="Total traffic classes configured (1-8)." />
        </label>
        <input
          id="tsn-numclasses"
          type="number"
          step={1}
          {...register('numTrafficClasses', { valueAsNumber: true })}
          className="mt-tight w-full"
        />
        <FieldError message={errors.numTrafficClasses?.message} />
      </div>
      <div>
        <label htmlFor="tsn-basetime" className="flex items-center gap-tight label">
          Base Time (ns)
          <HelpIcon tooltip="Base time for gate schedule (ns since epoch). 0 = use current time." />
        </label>
        <input
          id="tsn-basetime"
          type="number"
          step={1}
          {...register('baseTimeNs', { valueAsNumber: true })}
          className="mt-tight w-full"
        />
        <FieldError message={errors.baseTimeNs?.message} />
      </div>
    </div>
  );
}

interface TestSummarySectionProps {
  hasLatency: boolean;
  hasJitter: boolean;
  hasSync: boolean;
  hasPreemption: boolean;
  hasScheduling: boolean;
}

function TestSummarySection({
  hasLatency,
  hasJitter,
  hasSync,
  hasPreemption,
  hasScheduling,
}: TestSummarySectionProps): ReactElement {
  const { watch } = useFormContext<TSNConfig>();
  const v = watch();
  const selectedTestNames = [
    hasLatency && 'Latency',
    hasJitter && 'Jitter',
    hasSync && 'Sync Verification',
    hasPreemption && 'Preemption',
    hasScheduling && 'Scheduling',
  ].filter(Boolean);
  return (
    <div className="pad-sm rounded-lg bg-surface-base border border-surface-border">
      <div className="flex items-center gap-compact label mb-2">
        <Info className="w-4 h-4" />
        Test Summary
      </div>
      <div className="text-xs text-text-muted stack-xs">
        <div>Selected tests: {selectedTestNames.join(', ')}</div>
        <div>Frame size: {v.frameSize} bytes</div>
        <div>
          Timing: &le;{formatNs(v.maxLatencyNs)} latency | &le;{formatNs(v.maxJitterNs)} jitter
        </div>
        <div>
          PTP: {v.ptpEnabled ? 'Enabled' : 'Disabled'}
          {v.ptpEnabled && v.requirePTPSync ? ' (required)' : ''}
          {v.ptpEnabled ? ` | Max offset: ${formatNs(v.maxSyncOffsetNs)}` : ''}
        </div>
        {hasScheduling || hasPreemption ? (
          <div>
            Scheduling: Cycle {formatNs(v.cycleTimeNs)} | TC {v.trafficClass}
            {v.preemptionEnabled ? ' | Preemption enabled' : ''}
          </div>
        ) : null}
        <div>
          Duration: {v.duration}s + {v.warmup}s warmup
        </div>
      </div>
    </div>
  );
}

interface TSNConfigFormProps {
  config: TSNConfig;
  setConfig: (config: TSNConfig) => void;
  selectedTests: string[];
}

export function TSNConfigForm({
  config,
  setConfig,
  selectedTests,
}: TSNConfigFormProps): ReactElement | null {
  const hasTSNTests = selectedTests.some((t) => t.startsWith('tsn_'));

  const form = useConfigForm<TSNConfig>({
    schema: TSNConfigSchema,
    config,
    setConfig,
  });

  if (!hasTSNTests) {
    return null;
  }

  const hasLatency = selectedTests.includes('tsn_latency');
  const hasJitter = selectedTests.includes('tsn_jitter');
  const hasSync = selectedTests.includes('tsn_sync');
  const hasPreemption = selectedTests.includes('tsn_preemption');
  const hasScheduling = selectedTests.includes('tsn_scheduling');

  const rootErrors = form.formState.errors.root;
  const crossFieldError = rootErrors
    ? Object.values(rootErrors).find(
        (e): e is { message: string } =>
          typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string',
      )
    : undefined;

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Clock className="w-4 h-4" />
          <span>TSN Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <FormProvider {...form}>
        <div className="stack-lg">
          <TestParametersSection />
          <TimingRequirementsSection />
          <PTPConfigSection />

          {hasScheduling || hasPreemption ? <SchedulingConfigSection /> : null}

          {crossFieldError && (
            <div className="pad-xs rounded-lg bg-status-danger-subtle text-status-danger text-sm flex items-center gap-compact">
              <AlertTriangle className="w-4 h-4" />
              {crossFieldError.message}
            </div>
          )}

          <TestSummarySection
            hasLatency={hasLatency}
            hasJitter={hasJitter}
            hasSync={hasSync}
            hasPreemption={hasPreemption}
            hasScheduling={hasScheduling}
          />
        </div>
      </FormProvider>
    </CollapsibleSection>
  );
}

export default TSNConfigForm;

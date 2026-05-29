/**
 * @fileoverview The Stem - Traffic Generator Configuration
 * @description Migrated to react-hook-form + valibot per #325. MAC fields
 *              now validate format (or accept empty string for "auto").
 */

import { AlertTriangle, Info, Radio } from 'lucide-react';
import type { ReactElement } from 'react';
import { useConfigForm } from '../forms/useConfigForm';
import { TrafficGenConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** Traffic generator configuration parameters */
export interface TrafficGenConfig {
  frameSize: number;
  ratePct: number;
  duration: number;
  warmup: number;
  streamId: number;
  burstMode: boolean;
  burstSize: number;
  interBurstGapUs: number;
  srcMac: string;
  dstMac: string;
  vlanId: number;
  vlanPriority: number;
}

/** Default traffic generator configuration */
export const defaultTrafficGenConfig: TrafficGenConfig = {
  frameSize: 64,
  ratePct: 100,
  duration: 60,
  warmup: 2,
  streamId: 1,
  burstMode: false,
  burstSize: 100,
  interBurstGapUs: 1000,
  srcMac: '',
  dstMac: '',
  vlanId: 0,
  vlanPriority: 0,
};

const FRAME_SIZE_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 64, label: '64 B (min)' },
  { value: 128, label: '128 B' },
  { value: 256, label: '256 B' },
  { value: 512, label: '512 B' },
  { value: 1024, label: '1024 B' },
  { value: 1280, label: '1280 B' },
  { value: 1518, label: '1518 B (max)' },
  { value: 9000, label: '9000 B (jumbo)' },
];

const RATE_PRESETS: Array<{ value: number; label: string }> = [
  { value: 10, label: '10%' },
  { value: 25, label: '25%' },
  { value: 50, label: '50%' },
  { value: 75, label: '75%' },
  { value: 90, label: '90%' },
  { value: 100, label: '100%' },
];

interface TrafficGenConfigFormProps {
  config: TrafficGenConfig;
  setConfig: (config: TrafficGenConfig) => void;
  selectedTests: string[];
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

export function TrafficGenConfigForm({
  config,
  setConfig,
  selectedTests,
}: TrafficGenConfigFormProps): ReactElement | null {
  const hasTrafficGenTests = selectedTests.some(
    (t) => t.startsWith('trafficgen_') || t === 'custom_stream',
  );

  const form = useConfigForm<TrafficGenConfig>({
    schema: TrafficGenConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;

  if (!hasTrafficGenTests) {
    return null;
  }

  const frameSize = watch('frameSize') ?? 0;
  const ratePct = watch('ratePct') ?? 0;
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;
  const burstMode = watch('burstMode') ?? false;
  const burstSize = watch('burstSize') ?? 0;
  const interBurstGapUs = watch('interBurstGapUs') ?? 0;
  const vlanId = watch('vlanId') ?? 0;
  const vlanPriority = watch('vlanPriority') ?? 0;

  const hasCustomStream = selectedTests.includes('custom_stream');
  const hasBurst = selectedTests.includes('trafficgen_burst');
  const hasMultiStream = selectedTests.includes('trafficgen_multistream');

  const calculateThroughput = (): string => {
    const lineRateMbps = 10000;
    const throughputMbps = (lineRateMbps * ratePct) / 100;
    if (throughputMbps >= 1000) return `${(throughputMbps / 1000).toFixed(1)} Gbps`;
    return `${throughputMbps.toFixed(0)} Mbps`;
  };

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Radio className="w-4 h-4" />
          <span>Traffic Generator Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Traffic Parameters
          </div>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="tgen-framesize" className="flex items-center gap-tight label">
                Frame Size
                <HelpIcon tooltip="Ethernet frame size including FCS." />
              </label>
              <select
                id="tgen-framesize"
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

            <div>
              <label htmlFor="tgen-rate" className="flex items-center gap-tight label">
                Rate (% of line rate)
                <HelpIcon tooltip="Traffic rate as percentage of interface line rate." />
              </label>
              <div className="mt-tight flex gap-compact">
                <input
                  id="tgen-rate"
                  type="number"
                  step={0.01}
                  {...register('ratePct', { valueAsNumber: true })}
                  className="w-full"
                />
              </div>
              <FieldError message={errors.ratePct?.message} />
              <div className="mt-tight flex gap-tight flex-wrap">
                {RATE_PRESETS.map((preset) => (
                  <button
                    key={preset.value}
                    type="button"
                    onClick={() =>
                      setValue('ratePct', preset.value, {
                        shouldValidate: true,
                        shouldDirty: true,
                      })
                    }
                    className={`text-xs px-cell py-0.5 rounded border ${
                      ratePct === preset.value
                        ? 'bg-brand-primary text-on-brand border-brand-primary'
                        : 'bg-surface-base border-surface-border text-text-muted'
                    }`}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="tgen-duration" className="flex items-center gap-tight label">
                Duration (s)
                <HelpIcon tooltip="Traffic generation duration." />
              </label>
              <input
                id="tgen-duration"
                type="number"
                step={1}
                {...register('duration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.duration?.message} />
            </div>

            <div>
              <label htmlFor="tgen-warmup" className="flex items-center gap-tight label">
                Warmup (s)
                <HelpIcon tooltip="Warmup period before measurement." />
              </label>
              <input
                id="tgen-warmup"
                type="number"
                step={1}
                {...register('warmup', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.warmup?.message} />
            </div>
          </div>
        </div>

        {hasMultiStream ? (
          <div className="stack">
            <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
              Stream Configuration
            </div>
            <div>
              <label htmlFor="tgen-streamid" className="flex items-center gap-tight label">
                Stream ID
                <HelpIcon tooltip="Unique identifier for this traffic stream." />
              </label>
              <input
                id="tgen-streamid"
                type="number"
                step={1}
                {...register('streamId', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.streamId?.message} />
            </div>
          </div>
        ) : null}

        {hasBurst ? (
          <div className="stack">
            <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
              Burst Mode
            </div>
            <div className="flex items-center gap-compact">
              <input
                id="tgen-burstmode"
                type="checkbox"
                {...register('burstMode')}
                aria-label="Enable burst mode traffic generation"
                className="rounded border-surface-border"
              />
              <label
                htmlFor="tgen-burstmode"
                title="Send frames in short bursts separated by idle gaps rather than at a continuous rate; useful for testing buffer behavior"
                className="text-sm text-text-primary"
              >
                Enable burst mode
              </label>
            </div>

            {burstMode ? (
              <div className="grid grid-cols-2 gap-default">
                <div>
                  <label htmlFor="tgen-burstsize" className="flex items-center gap-tight label">
                    Burst Size (frames)
                    <HelpIcon tooltip="Number of frames per burst." />
                  </label>
                  <input
                    id="tgen-burstsize"
                    type="number"
                    step={1}
                    {...register('burstSize', { valueAsNumber: true })}
                    className="mt-tight w-full"
                  />
                  <FieldError message={errors.burstSize?.message} />
                </div>
                <div>
                  <label htmlFor="tgen-ibg" className="flex items-center gap-tight label">
                    Inter-Burst Gap (µs)
                    <HelpIcon tooltip="Gap between bursts in microseconds." />
                  </label>
                  <input
                    id="tgen-ibg"
                    type="number"
                    step={1}
                    {...register('interBurstGapUs', { valueAsNumber: true })}
                    className="mt-tight w-full"
                  />
                  <FieldError message={errors.interBurstGapUs?.message} />
                </div>
              </div>
            ) : null}
          </div>
        ) : null}

        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            VLAN Configuration
          </div>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="tgen-vlanid" className="flex items-center gap-tight label">
                VLAN ID
                <HelpIcon tooltip="VLAN ID (0 = untagged)." />
              </label>
              <input
                id="tgen-vlanid"
                type="number"
                step={1}
                {...register('vlanId', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.vlanId?.message} />
            </div>
            <div>
              <label htmlFor="tgen-vlanpri" className="flex items-center gap-tight label">
                VLAN Priority
                <HelpIcon tooltip="802.1p priority (0-7)." />
              </label>
              <input
                id="tgen-vlanpri"
                type="number"
                step={1}
                disabled={vlanId === 0}
                {...register('vlanPriority', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.vlanPriority?.message} />
            </div>
          </div>
        </div>

        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            MAC Addresses (Optional)
          </div>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="tgen-srcmac" className="flex items-center gap-tight label">
                Source MAC
                <HelpIcon tooltip="Source MAC address (leave empty for auto). Format: AA:BB:CC:DD:EE:FF." />
              </label>
              <input
                id="tgen-srcmac"
                type="text"
                placeholder="aa:bb:cc:dd:ee:ff"
                {...register('srcMac')}
                className="mt-tight w-full"
              />
              <FieldError message={errors.srcMac?.message} />
            </div>
            <div>
              <label htmlFor="tgen-dstmac" className="flex items-center gap-tight label">
                Destination MAC
                <HelpIcon tooltip="Destination MAC address (leave empty for broadcast). Format: AA:BB:CC:DD:EE:FF." />
              </label>
              <input
                id="tgen-dstmac"
                type="text"
                placeholder="aa:bb:cc:dd:ee:ff"
                {...register('dstMac')}
                className="mt-tight w-full"
              />
              <FieldError message={errors.dstMac?.message} />
            </div>
          </div>
        </div>

        <div className="pad-sm rounded-lg bg-surface-base border border-surface-border">
          <div className="flex items-center gap-compact label mb-2">
            <Info className="w-4 h-4" />
            Traffic Summary
          </div>
          <div className="text-xs text-text-muted stack-xs">
            <div>
              Selected tests:{' '}
              {[
                hasCustomStream && 'Custom Stream',
                hasBurst && 'Burst Mode',
                hasMultiStream && 'Multi-Stream',
              ]
                .filter(Boolean)
                .join(', ')}
            </div>
            <div>
              Frame: {frameSize}B @ {ratePct}% line rate (~{calculateThroughput()})
            </div>
            {vlanId > 0 ? (
              <div>
                VLAN: {vlanId} (priority {vlanPriority})
              </div>
            ) : null}
            {burstMode ? (
              <div>
                Burst: {burstSize} frames, {interBurstGapUs}µs gap
              </div>
            ) : null}
            <div>
              Duration: {duration}s + {warmup}s warmup
            </div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default TrafficGenConfigForm;

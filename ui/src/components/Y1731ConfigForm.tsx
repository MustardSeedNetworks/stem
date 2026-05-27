/**
 * @fileoverview The Stem - Y.1731 OAM Test Configuration
 * @description Migrated to react-hook-form + valibot per #325.
 */

import { AlertTriangle, Gauge, Info } from 'lucide-react';
import type { ReactElement } from 'react';
import { useConfigForm } from '../forms/useConfigForm';
import { Y1731ConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** Y.1731 test configuration parameters */
export interface Y1731Config {
  mepId: number;
  megLevel: number;
  megId: string;
  ccmInterval: number;
  priority: number;
  duration: number;
  intervalMs: number;
  count: number;
  frameSize: number;
  priorityTagged: boolean;
}

/** Default Y.1731 configuration */
export const defaultY1731Config: Y1731Config = {
  mepId: 1,
  megLevel: 4,
  megId: 'MSN-MEG-01',
  ccmInterval: 1000,
  priority: 6,
  duration: 60,
  intervalMs: 100,
  count: 10,
  frameSize: 64,
  priorityTagged: true,
};

const CCM_INTERVAL_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 3, label: '3.33 ms' },
  { value: 10, label: '10 ms' },
  { value: 100, label: '100 ms' },
  { value: 1000, label: '1 s' },
  { value: 10000, label: '10 s' },
  { value: 60000, label: '1 min' },
  { value: 600000, label: '10 min' },
];

const FRAME_SIZE_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 64, label: '64 B (min)' },
  { value: 128, label: '128 B' },
  { value: 256, label: '256 B' },
  { value: 512, label: '512 B' },
  { value: 1024, label: '1024 B' },
  { value: 1518, label: '1518 B (max)' },
];

interface Y1731ConfigFormProps {
  config: Y1731Config;
  setConfig: (config: Y1731Config) => void;
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

export function Y1731ConfigForm({
  config,
  setConfig,
  selectedTests,
}: Y1731ConfigFormProps): ReactElement | null {
  const hasY1731Tests = selectedTests.some((t) => t.startsWith('y1731'));

  const form = useConfigForm<Y1731Config>({
    schema: Y1731ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    formState: { errors },
  } = form;

  if (!hasY1731Tests) {
    return null;
  }

  const mepId = watch('mepId') ?? 0;
  const megLevel = watch('megLevel') ?? 0;
  const megId = watch('megId') ?? '';
  const ccmInterval = watch('ccmInterval') ?? 0;
  const priority = watch('priority') ?? 0;
  const priorityTagged = watch('priorityTagged') ?? false;
  const frameSize = watch('frameSize') ?? 0;
  const intervalMs = watch('intervalMs') ?? 0;
  const count = watch('count') ?? 0;
  const duration = watch('duration') ?? 0;

  const hasDelay = selectedTests.includes('y1731_delay');
  const hasLoss = selectedTests.includes('y1731_loss');
  const hasSLM = selectedTests.includes('y1731_slm');
  const hasLoopback = selectedTests.includes('y1731_loopback');

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Gauge className="w-4 h-4" />
          <span>Y.1731 OAM Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            MEP/MEG Configuration
          </div>

          <div className="grid grid-cols-3 gap-default">
            <div>
              <label htmlFor="y1731-mepid" className="flex items-center gap-tight label">
                MEP ID
                <HelpIcon tooltip="Maintenance End Point identifier (1-8191)." />
              </label>
              <input
                id="y1731-mepid"
                type="number"
                step={1}
                {...register('mepId', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.mepId?.message} />
            </div>

            <div>
              <label htmlFor="y1731-meglevel" className="flex items-center gap-tight label">
                MEG Level
                <HelpIcon tooltip="Maintenance Entity Group level (0-7). Higher = wider domain." />
              </label>
              <input
                id="y1731-meglevel"
                type="number"
                step={1}
                {...register('megLevel', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.megLevel?.message} />
            </div>

            <div>
              <label htmlFor="y1731-megid" className="flex items-center gap-tight label">
                MEG ID
                <HelpIcon tooltip="Maintenance Entity Group identifier string." />
              </label>
              <input
                id="y1731-megid"
                type="text"
                maxLength={45}
                {...register('megId')}
                className="mt-tight w-full"
              />
              <FieldError message={errors.megId?.message} />
            </div>
          </div>
        </div>

        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            OAM Parameters
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="y1731-ccm" className="flex items-center gap-tight label">
                CCM Interval
                <HelpIcon tooltip="Continuity Check Message interval per Y.1731." />
              </label>
              <select
                id="y1731-ccm"
                {...register('ccmInterval', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {CCM_INTERVAL_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <FieldError message={errors.ccmInterval?.message} />
            </div>

            <div>
              <label htmlFor="y1731-priority" className="flex items-center gap-tight label">
                Priority
                <HelpIcon tooltip="802.1p priority value (0-7). 6-7 typically for OAM." />
              </label>
              <input
                id="y1731-priority"
                type="number"
                step={1}
                {...register('priority', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.priority?.message} />
            </div>
          </div>

          <div className="flex items-center gap-compact">
            <input
              id="y1731-tagged"
              type="checkbox"
              {...register('priorityTagged')}
              aria-label="Use 802.1Q priority tagging on OAM frames"
              className="rounded border-surface-border"
            />
            <label
              htmlFor="y1731-tagged"
              title="Add an IEEE 802.1Q VLAN tag with the selected priority to all Y.1731 OAM frames"
              className="text-sm text-text-primary"
            >
              Use priority tagging (802.1Q)
            </label>
          </div>
        </div>

        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Measurement Parameters
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="y1731-framesize" className="flex items-center gap-tight label">
                Frame Size
                <HelpIcon tooltip="OAM frame size for measurements." />
              </label>
              <select
                id="y1731-framesize"
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
              <label htmlFor="y1731-interval" className="flex items-center gap-tight label">
                Measurement Interval (ms)
                <HelpIcon tooltip="Interval between measurement probes." />
              </label>
              <input
                id="y1731-interval"
                type="number"
                step={10}
                {...register('intervalMs', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.intervalMs?.message} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="y1731-count" className="flex items-center gap-tight label">
                Frames per Interval
                <HelpIcon tooltip="Number of measurement frames per interval." />
              </label>
              <input
                id="y1731-count"
                type="number"
                step={1}
                {...register('count', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.count?.message} />
            </div>

            <div>
              <label htmlFor="y1731-duration" className="flex items-center gap-tight label">
                Duration (s)
                <HelpIcon tooltip="Total measurement duration." />
              </label>
              <input
                id="y1731-duration"
                type="number"
                step={1}
                {...register('duration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.duration?.message} />
            </div>
          </div>
        </div>

        <div className="pad-sm rounded-lg bg-surface-base border border-surface-border">
          <div className="flex items-center gap-compact label mb-2">
            <Info className="w-4 h-4" />
            Test Summary
          </div>
          <div className="text-xs text-text-muted stack-xs">
            <div>
              Selected tests:{' '}
              {[
                hasDelay && 'Delay Measurement',
                hasLoss && 'Frame Loss',
                hasSLM && 'Synthetic Loss',
                hasLoopback && 'Loopback',
              ]
                .filter(Boolean)
                .join(', ')}
            </div>
            <div>
              MEP: {mepId} | Level: {megLevel} | MEG: {megId}
            </div>
            <div>
              CCM:{' '}
              {CCM_INTERVAL_OPTIONS.find((c) => c.value === ccmInterval)?.label ||
                `${ccmInterval}ms`}{' '}
              | Priority: {priority}
              {priorityTagged ? ' (tagged)' : ''}
            </div>
            <div>
              Frame: {frameSize}B | Interval: {intervalMs}ms | Count: {count}
            </div>
            <div>Duration: {duration}s</div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default Y1731ConfigForm;

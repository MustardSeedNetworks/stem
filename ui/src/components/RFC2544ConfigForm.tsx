/**
 * @fileoverview The Stem - RFC 2544 Benchmark Test Configuration
 * @description Advanced configuration form for RFC 2544 Benchmarking Tests.
 *              Migrated to react-hook-form + valibot per #325. The schema
 *              lives at src/schemas/configs.ts (RFC2544ConfigSchema).
 */

import { AlertTriangle, Info } from 'lucide-react';
import type { ReactElement } from 'react';
import { useConfigForm } from '../forms/useConfigForm';
import { RFC2544ConfigSchema } from '../schemas/configs';
import { HelpIcon } from './HelpIcon';

/** RFC 2544 test configuration parameters */
export interface RFC2544Config {
  /** Test duration in seconds */
  duration: number;
  /** Frame sizes to test (bytes) */
  frameSizes: number[];
  /** Resolution for binary search (percentage) */
  resolution: number;
  /** Maximum acceptable frame loss (percentage) */
  maxLoss: number;
  /** Warmup duration before measurement (seconds) */
  warmup: number;
  /** Number of trials per test point */
  trials: number;
  /** Step size for frame loss rate test (percentage) */
  stepSize: number;
  /** Enable bidirectional testing */
  bidirectional: boolean;
}

/** Default RFC 2544 configuration */
export const defaultRFC2544Config: RFC2544Config = {
  duration: 60,
  frameSizes: [64, 128, 256, 512, 1024, 1280, 1518],
  resolution: 0.1,
  maxLoss: 0.0,
  warmup: 2,
  trials: 3,
  stepSize: 10,
  bidirectional: false,
};

interface RFC2544ConfigFormProps {
  config: RFC2544Config;
  setConfig: (config: RFC2544Config) => void;
  selectedTests: string[];
}

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

function FieldError({ message }: { message?: string }): ReactElement | null {
  if (!message) return null;
  return (
    <div className="mt-tight text-xs text-status-danger flex items-center gap-tight">
      <AlertTriangle className="w-3 h-3" />
      {message}
    </div>
  );
}

export function RFC2544ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC2544ConfigFormProps): ReactElement | null {
  const hasRFC2544Tests = selectedTests.some((t) => t.startsWith('rfc2544'));

  const form = useConfigForm<RFC2544Config>({
    schema: RFC2544ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;

  if (!hasRFC2544Tests) {
    return null;
  }

  const frameSizes = watch('frameSizes') ?? [];
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;
  const trials = watch('trials') ?? 0;
  const resolution = watch('resolution') ?? 0;
  const maxLoss = watch('maxLoss') ?? 0;
  const stepSize = watch('stepSize') ?? 0;
  const bidirectional = watch('bidirectional') ?? false;

  const toggleFrameSize = (size: number): void => {
    if (frameSizes.includes(size)) {
      setValue(
        'frameSizes',
        frameSizes.filter((s) => s !== size),
        { shouldValidate: true, shouldDirty: true },
      );
    } else {
      setValue(
        'frameSizes',
        [...frameSizes, size].sort((a, b) => a - b),
        {
          shouldValidate: true,
          shouldDirty: true,
        },
      );
    }
  };

  const hasThroughput = selectedTests.includes('rfc2544_throughput');
  const hasLatency = selectedTests.includes('rfc2544_latency');
  const hasFrameLoss = selectedTests.includes('rfc2544_frame_loss');
  const hasBackToBack = selectedTests.includes('rfc2544_back_to_back');

  return (
    <div className="stack-lg">
      {/* Test Duration */}
      <div className="stack">
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
          Test Duration
        </div>

        <div>
          <label htmlFor="rfc2544-duration" className="flex items-center gap-tight label">
            Duration per Test (s)
            <HelpIcon tooltip="Duration for each test iteration. Longer durations provide more accurate results but take more time." />
          </label>
          <input
            id="rfc2544-duration"
            type="number"
            step={1}
            {...register('duration', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.duration?.message} />
        </div>

        <div>
          <label htmlFor="rfc2544-warmup" className="flex items-center gap-tight label">
            Warmup Duration (s)
            <HelpIcon tooltip="Time to stabilize traffic flow before starting measurements." />
          </label>
          <input
            id="rfc2544-warmup"
            type="number"
            step={1}
            {...register('warmup', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.warmup?.message} />
        </div>

        <div>
          <label htmlFor="rfc2544-trials" className="flex items-center gap-tight label">
            Number of Trials
            <HelpIcon tooltip="Number of times to repeat each test point. More trials improve statistical accuracy." />
          </label>
          <input
            id="rfc2544-trials"
            type="number"
            step={1}
            {...register('trials', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.trials?.message} />
        </div>
      </div>

      {/* Throughput Test Parameters */}
      {hasThroughput ? (
        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Throughput Test
          </div>

          <div>
            <label htmlFor="rfc2544-resolution" className="flex items-center gap-tight label">
              Resolution (%)
              <HelpIcon tooltip="Binary search resolution. Default: 0.1%." />
            </label>
            <input
              id="rfc2544-resolution"
              type="number"
              step={0.01}
              {...register('resolution', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.resolution?.message} />
          </div>

          <div>
            <label htmlFor="rfc2544-maxloss" className="flex items-center gap-tight label">
              Max Acceptable Loss (%)
              <HelpIcon tooltip="Maximum frame loss considered acceptable. RFC 2544 specifies 0%." />
            </label>
            <input
              id="rfc2544-maxloss"
              type="number"
              step={0.001}
              {...register('maxLoss', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.maxLoss?.message} />
          </div>
        </div>
      ) : null}

      {/* Frame Loss Test Parameters */}
      {hasFrameLoss ? (
        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Frame Loss Test
          </div>

          <div>
            <label htmlFor="rfc2544-stepsize" className="flex items-center gap-tight label">
              Step Size (%)
              <HelpIcon tooltip="Load increment step for frame loss rate measurement." />
            </label>
            <input
              id="rfc2544-stepsize"
              type="number"
              step={1}
              {...register('stepSize', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.stepSize?.message} />
            <div className="text-xs text-text-muted mt-tight">
              Tests at:{' '}
              {Array.from(
                { length: Math.floor(100 / Math.max(1, stepSize)) + 1 },
                (_, i) => `${i * stepSize}%`,
              ).join(', ')}
            </div>
          </div>
        </div>
      ) : null}

      {/* Frame Sizes */}
      <div className="stack-sm">
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide flex items-center gap-tight">
          Frame Sizes
          <HelpIcon tooltip="Select frame sizes to test. RFC 2544 specifies: 64, 128, 256, 512, 1024, 1280, 1518 bytes." />
        </div>
        <div className="grid grid-cols-2 gap-compact">
          {FRAME_SIZE_OPTIONS.map((option) => (
            <label
              key={option.value}
              title={`Include ${option.value}-byte frames in the RFC 2544 sweep`}
              className="flex items-center gap-compact pad-xs rounded-lg cursor-pointer hover:bg-surface-hover text-sm"
            >
              <input
                type="checkbox"
                checked={frameSizes.includes(option.value)}
                onChange={() => toggleFrameSize(option.value)}
                aria-label={`Test ${option.value}-byte frames`}
                className="w-4 h-4 accent-brand-primary"
              />
              <span className="text-text-primary">{option.label}</span>
            </label>
          ))}
        </div>
        <FieldError message={errors.frameSizes?.message} />
      </div>

      {/* Advanced Options */}
      <div className="stack">
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
          Advanced Options
        </div>

        <label
          title="Send and measure traffic in both directions simultaneously"
          className="flex items-center gap-default pad-xs rounded-lg cursor-pointer hover:bg-surface-hover"
        >
          <input
            type="checkbox"
            {...register('bidirectional')}
            aria-label="Enable bidirectional testing"
            className="w-4 h-4 accent-brand-primary"
          />
          <div>
            <div className="font-medium text-sm flex items-center gap-tight">
              Bidirectional Testing
              <HelpIcon tooltip="Run tests in both directions simultaneously." />
            </div>
            <div className="text-xs text-text-muted">Test both TX and RX paths</div>
          </div>
        </label>
      </div>

      {/* Test Summary */}
      <div className="pad-sm rounded-lg bg-surface-base border border-surface-border">
        <div className="flex items-center gap-compact label mb-2">
          <Info className="w-4 h-4" />
          Test Summary
        </div>
        <div className="text-xs text-text-muted stack-xs">
          <div>
            Selected tests:{' '}
            {[
              hasThroughput && 'Throughput',
              hasLatency && 'Latency',
              hasFrameLoss && 'Frame Loss',
              hasBackToBack && 'Back-to-Back',
              selectedTests.includes('rfc2544_system_recovery') && 'System Recovery',
              selectedTests.includes('rfc2544_reset') && 'Reset',
            ]
              .filter(Boolean)
              .join(', ')}
          </div>
          <div>Frame sizes: {frameSizes.join(', ')} bytes</div>
          <div>
            Duration: {duration}s × {trials} trials
            {warmup > 0 && ` + ${warmup}s warmup`}
          </div>
          {hasThroughput ? (
            <div>
              Throughput: {resolution}% resolution, ≤{maxLoss}% loss
            </div>
          ) : null}
          {bidirectional ? <div>Mode: Bidirectional</div> : null}
          <div className="pt-tight border-t border-surface-border mt-tight">
            Estimated time: ~
            {Math.ceil(
              ((duration + warmup) *
                trials *
                frameSizes.length *
                selectedTests.filter((t) => t.startsWith('rfc2544')).length) /
                60,
            )}{' '}
            minutes
          </div>
        </div>
      </div>
    </div>
  );
}

export default RFC2544ConfigForm;

/**
 * @fileoverview The Stem - RFC 2544 Benchmark Test Configuration
 * @description Advanced configuration form for RFC 2544 Benchmarking Tests.
 *              Allows users to configure test parameters including duration,
 *              frame sizes, resolution, loss threshold, warmup, and trials.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { AlertTriangle, Info } from 'lucide-react';
import type { ReactElement } from 'react';
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

/** Standard Ethernet frame sizes per RFC 2544 */
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

export function RFC2544ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC2544ConfigFormProps): ReactElement | null {
  const hasRFC2544Tests = selectedTests.some((t) => t.startsWith('rfc2544'));

  if (!hasRFC2544Tests) {
    return null;
  }

  const updateConfig = (updates: Partial<RFC2544Config>): void => {
    setConfig({ ...config, ...updates });
  };

  const toggleFrameSize = (size: number): void => {
    if (config.frameSizes.includes(size)) {
      updateConfig({ frameSizes: config.frameSizes.filter((s) => s !== size) });
    } else {
      updateConfig({
        frameSizes: [...config.frameSizes, size].sort((a, b) => a - b),
      });
    }
  };

  const hasThroughput = selectedTests.includes('rfc2544_throughput');
  const hasLatency = selectedTests.includes('rfc2544_latency');
  const hasFrameLoss = selectedTests.includes('rfc2544_frame_loss');
  const hasBackToBack = selectedTests.includes('rfc2544_back_to_back');

  return (
    <div className="space-y-4">
      {/* Test Duration */}
      <div className="space-y-3">
        <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
          Test Duration
        </div>

        <div>
          <label
            htmlFor="rfc2544-duration"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Duration per Test (s)
            <HelpIcon tooltip="Duration for each test iteration. Longer durations provide more accurate results but take more time." />
          </label>
          <input
            id="rfc2544-duration"
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
            htmlFor="rfc2544-warmup"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Warmup Duration (s)
            <HelpIcon tooltip="Time to stabilize traffic flow before starting measurements. Helps ensure accurate results." />
          </label>
          <input
            id="rfc2544-warmup"
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
            htmlFor="rfc2544-trials"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Number of Trials
            <HelpIcon tooltip="Number of times to repeat each test point. More trials improve statistical accuracy." />
          </label>
          <input
            id="rfc2544-trials"
            type="number"
            min={1}
            max={10}
            step={1}
            value={config.trials}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ trials: Number(e.target.value) })
            }
            className="mt-1 w-full"
          />
        </div>
      </div>

      {/* Throughput Test Parameters */}
      {hasThroughput ? (
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Throughput Test
          </div>

          <div>
            <label
              htmlFor="rfc2544-resolution"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Resolution (%)
              <HelpIcon tooltip="Binary search resolution. Lower values give more precise throughput results but take longer. Default: 0.1%" />
            </label>
            <input
              id="rfc2544-resolution"
              type="number"
              min={0.01}
              max={10}
              step={0.01}
              value={config.resolution}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ resolution: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>

          <div>
            <label
              htmlFor="rfc2544-maxloss"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Max Acceptable Loss (%)
              <HelpIcon tooltip="Maximum frame loss considered acceptable for throughput calculation. RFC 2544 specifies 0% but some implementations allow small tolerance." />
            </label>
            <input
              id="rfc2544-maxloss"
              type="number"
              min={0}
              max={1}
              step={0.001}
              value={config.maxLoss}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ maxLoss: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>
        </div>
      ) : null}

      {/* Frame Loss Test Parameters */}
      {hasFrameLoss ? (
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Frame Loss Test
          </div>

          <div>
            <label
              htmlFor="rfc2544-stepsize"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Step Size (%)
              <HelpIcon tooltip="Load increment step for frame loss rate measurement. Smaller steps give finer granularity." />
            </label>
            <input
              id="rfc2544-stepsize"
              type="number"
              min={1}
              max={25}
              step={1}
              value={config.stepSize}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ stepSize: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
            <div className="text-xs text-[var(--color-text-muted)] mt-1">
              Tests at:{' '}
              {Array.from(
                { length: Math.floor(100 / config.stepSize) + 1 },
                (_, i) => `${i * config.stepSize}%`,
              ).join(', ')}
            </div>
          </div>
        </div>
      ) : null}

      {/* Frame Sizes */}
      <div className="space-y-2">
        <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide flex items-center gap-1">
          Frame Sizes
          <HelpIcon tooltip="Select frame sizes to test. RFC 2544 specifies: 64, 128, 256, 512, 1024, 1280, 1518 bytes." />
        </div>
        <div className="grid grid-cols-2 gap-2">
          {FRAME_SIZE_OPTIONS.map((option) => (
            <label
              key={option.value}
              className="flex items-center gap-2 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)] text-sm"
            >
              <input
                type="checkbox"
                checked={config.frameSizes.includes(option.value)}
                onChange={() => toggleFrameSize(option.value)}
                className="w-4 h-4 accent-[var(--color-brand-primary)]"
              />
              <span className="text-[var(--color-text-primary)]">{option.label}</span>
            </label>
          ))}
        </div>
        {config.frameSizes.length === 0 ? (
          <div className="flex items-center gap-2 text-xs text-[var(--color-status-warning)]">
            <AlertTriangle className="w-3 h-3" />
            Select at least one frame size
          </div>
        ) : null}
      </div>

      {/* Advanced Options */}
      <div className="space-y-3">
        <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
          Advanced Options
        </div>

        <label className="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]">
          <input
            type="checkbox"
            checked={config.bidirectional}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ bidirectional: e.target.checked })
            }
            className="w-4 h-4 accent-[var(--color-brand-primary)]"
          />
          <div>
            <div className="font-medium text-sm flex items-center gap-1">
              Bidirectional Testing
              <HelpIcon tooltip="Run tests in both directions simultaneously. Doubles traffic load but tests full-duplex capacity." />
            </div>
            <div className="text-xs text-[var(--color-text-muted)]">Test both TX and RX paths</div>
          </div>
        </label>
      </div>

      {/* Test Summary */}
      <div className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
        <div className="flex items-center gap-2 text-sm font-medium text-[var(--color-text-primary)] mb-2">
          <Info className="w-4 h-4" />
          Test Summary
        </div>
        <div className="text-xs text-[var(--color-text-muted)] space-y-1">
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
          <div>Frame sizes: {config.frameSizes.join(', ')} bytes</div>
          <div>
            Duration: {config.duration}s × {config.trials} trials
            {config.warmup > 0 && ` + ${config.warmup}s warmup`}
          </div>
          {hasThroughput ? (
            <div>
              Throughput: {config.resolution}% resolution, ≤{config.maxLoss}% loss
            </div>
          ) : null}
          {config.bidirectional ? <div>Mode: Bidirectional</div> : null}
          <div className="pt-1 border-t border-[var(--color-surface-border)] mt-1">
            Estimated time: ~
            {Math.ceil(
              ((config.duration + config.warmup) *
                config.trials *
                config.frameSizes.length *
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

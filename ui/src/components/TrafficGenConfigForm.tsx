/**
 * @fileoverview The Stem - Traffic Generator Configuration
 * @description Configuration form for Custom Traffic Generation.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { Info, Radio } from 'lucide-react';
import type { ReactElement } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** Traffic generator configuration parameters */
export interface TrafficGenConfig {
  /** Frame size in bytes */
  frameSize: number;
  /** Traffic rate as percentage of line rate */
  ratePct: number;
  /** Traffic generation duration in seconds */
  duration: number;
  /** Warmup duration in seconds */
  warmup: number;
  /** Stream identifier for multi-stream tests */
  streamId: number;
  /** Enable burst mode */
  burstMode: boolean;
  /** Burst size in frames */
  burstSize: number;
  /** Inter-burst gap in microseconds */
  interBurstGapUs: number;
  /** Source MAC address (optional) */
  srcMac: string;
  /** Destination MAC address (optional) */
  dstMac: string;
  /** VLAN ID (0 = untagged) */
  vlanId: number;
  /** VLAN priority (0-7) */
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

/** Frame size options */
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

/** Rate presets */
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

export function TrafficGenConfigForm({
  config,
  setConfig,
  selectedTests,
}: TrafficGenConfigFormProps): ReactElement | null {
  const hasTrafficGenTests = selectedTests.some(
    (t) => t.startsWith('trafficgen_') || t === 'custom_stream',
  );

  if (!hasTrafficGenTests) {
    return null;
  }

  const updateConfig = (updates: Partial<TrafficGenConfig>): void => {
    setConfig({ ...config, ...updates });
  };

  const hasCustomStream = selectedTests.includes('custom_stream');
  const hasBurst = selectedTests.includes('trafficgen_burst');
  const hasMultiStream = selectedTests.includes('trafficgen_multistream');

  // Calculate approximate throughput
  const calculateThroughput = (): string => {
    // Assume 10 Gbps line rate for calculation
    const lineRateMbps = 10000;
    const throughputMbps = (lineRateMbps * config.ratePct) / 100;
    if (throughputMbps >= 1000) {
      return `${(throughputMbps / 1000).toFixed(1)} Gbps`;
    }
    return `${throughputMbps.toFixed(0)} Mbps`;
  };

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Radio className="w-4 h-4" />
          <span>Traffic Generator Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        {/* Basic Parameters */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Traffic Parameters
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="tgen-framesize"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Frame Size
                <HelpIcon tooltip="Ethernet frame size including FCS." />
              </label>
              <select
                id="tgen-framesize"
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

            <div>
              <label
                htmlFor="tgen-rate"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Rate (% of line rate)
                <HelpIcon tooltip="Traffic rate as percentage of interface line rate." />
              </label>
              <div className="mt-1 flex gap-2">
                <input
                  id="tgen-rate"
                  type="number"
                  min={0.01}
                  max={100}
                  step={0.01}
                  value={config.ratePct}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                    updateConfig({ ratePct: Number(e.target.value) })
                  }
                  className="w-full"
                />
              </div>
              <div className="mt-1 flex gap-1 flex-wrap">
                {RATE_PRESETS.map((preset) => (
                  <button
                    key={preset.value}
                    type="button"
                    onClick={() => updateConfig({ ratePct: preset.value })}
                    className={`text-xs px-2 py-0.5 rounded border ${
                      config.ratePct === preset.value
                        ? 'bg-[var(--color-brand-primary)] text-white border-[var(--color-brand-primary)]'
                        : 'bg-[var(--color-surface-base)] border-[var(--color-surface-border)] text-[var(--color-text-muted)]'
                    }`}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="tgen-duration"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Duration (s)
                <HelpIcon tooltip="Traffic generation duration." />
              </label>
              <input
                id="tgen-duration"
                type="number"
                min={1}
                max={86400}
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
                htmlFor="tgen-warmup"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Warmup (s)
                <HelpIcon tooltip="Warmup period before measurement." />
              </label>
              <input
                id="tgen-warmup"
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
          </div>
        </div>

        {/* Stream Configuration */}
        {hasMultiStream ? (
          <div className="space-y-3">
            <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
              Stream Configuration
            </div>

            <div>
              <label
                htmlFor="tgen-streamid"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Stream ID
                <HelpIcon tooltip="Unique identifier for this traffic stream." />
              </label>
              <input
                id="tgen-streamid"
                type="number"
                min={1}
                max={65535}
                step={1}
                value={config.streamId}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ streamId: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>
          </div>
        ) : null}

        {/* Burst Mode */}
        {hasBurst ? (
          <div className="space-y-3">
            <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
              Burst Mode
            </div>

            <div className="flex items-center gap-2">
              <input
                id="tgen-burstmode"
                type="checkbox"
                checked={config.burstMode}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ burstMode: e.target.checked })
                }
                aria-label="Enable burst mode traffic generation"
                className="rounded border-[var(--color-surface-border)]"
              />
              <label
                htmlFor="tgen-burstmode"
                title="Send frames in short bursts separated by idle gaps rather than at a continuous rate; useful for testing buffer behavior"
                className="text-sm text-[var(--color-text-primary)]"
              >
                Enable burst mode
              </label>
            </div>

            {config.burstMode ? (
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label
                    htmlFor="tgen-burstsize"
                    className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
                  >
                    Burst Size (frames)
                    <HelpIcon tooltip="Number of frames per burst." />
                  </label>
                  <input
                    id="tgen-burstsize"
                    type="number"
                    min={1}
                    max={10000}
                    step={1}
                    value={config.burstSize}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                      updateConfig({ burstSize: Number(e.target.value) })
                    }
                    className="mt-1 w-full"
                  />
                </div>

                <div>
                  <label
                    htmlFor="tgen-ibg"
                    className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
                  >
                    Inter-Burst Gap (µs)
                    <HelpIcon tooltip="Gap between bursts in microseconds." />
                  </label>
                  <input
                    id="tgen-ibg"
                    type="number"
                    min={0}
                    max={1000000}
                    step={1}
                    value={config.interBurstGapUs}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                      updateConfig({ interBurstGapUs: Number(e.target.value) })
                    }
                    className="mt-1 w-full"
                  />
                </div>
              </div>
            ) : null}
          </div>
        ) : null}

        {/* VLAN Configuration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            VLAN Configuration
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="tgen-vlanid"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                VLAN ID
                <HelpIcon tooltip="VLAN ID (0 = untagged)." />
              </label>
              <input
                id="tgen-vlanid"
                type="number"
                min={0}
                max={4094}
                step={1}
                value={config.vlanId}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ vlanId: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="tgen-vlanpri"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                VLAN Priority
                <HelpIcon tooltip="802.1p priority (0-7)." />
              </label>
              <input
                id="tgen-vlanpri"
                type="number"
                min={0}
                max={7}
                step={1}
                value={config.vlanPriority}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ vlanPriority: Number(e.target.value) })
                }
                className="mt-1 w-full"
                disabled={config.vlanId === 0}
              />
            </div>
          </div>
        </div>

        {/* MAC Address Configuration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            MAC Addresses (Optional)
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="tgen-srcmac"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Source MAC
                <HelpIcon tooltip="Source MAC address (leave empty for auto)." />
              </label>
              <input
                id="tgen-srcmac"
                type="text"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={config.srcMac}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ srcMac: e.target.value })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="tgen-dstmac"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Destination MAC
                <HelpIcon tooltip="Destination MAC address (leave empty for broadcast)." />
              </label>
              <input
                id="tgen-dstmac"
                type="text"
                placeholder="aa:bb:cc:dd:ee:ff"
                value={config.dstMac}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ dstMac: e.target.value })
                }
                className="mt-1 w-full"
              />
            </div>
          </div>
        </div>

        {/* Test Summary */}
        <div className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
          <div className="flex items-center gap-2 text-sm font-medium text-[var(--color-text-primary)] mb-2">
            <Info className="w-4 h-4" />
            Traffic Summary
          </div>
          <div className="text-xs text-[var(--color-text-muted)] space-y-1">
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
              Frame: {config.frameSize}B @ {config.ratePct}% line rate (~
              {calculateThroughput()})
            </div>
            {config.vlanId > 0 ? (
              <div>
                VLAN: {config.vlanId} (priority {config.vlanPriority})
              </div>
            ) : null}
            {config.burstMode ? (
              <div>
                Burst: {config.burstSize} frames, {config.interBurstGapUs}µs gap
              </div>
            ) : null}
            <div>
              Duration: {config.duration}s + {config.warmup}s warmup
            </div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default TrafficGenConfigForm;

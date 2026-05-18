/**
 * @fileoverview The Stem - Y.1731 OAM Test Configuration
 * @description Configuration form for ITU-T Y.1731 Ethernet OAM Testing.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { Gauge, Info } from 'lucide-react';
import type React from 'react';
import type { ReactElement } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** Y.1731 test configuration parameters */
export interface Y1731Config {
  /** Maintenance End Point ID */
  mepId: number;
  /** Maintenance Entity Group Level (0-7) */
  megLevel: number;
  /** Maintenance Entity Group ID */
  megId: string;
  /** CCM interval in milliseconds */
  ccmInterval: number;
  /** Frame priority (0-7) */
  priority: number;
  /** Test duration in seconds */
  duration: number;
  /** Measurement interval in milliseconds */
  intervalMs: number;
  /** Number of test frames per interval */
  count: number;
  /** Frame size in bytes */
  frameSize: number;
  /** Whether to use priority tagging */
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

/** Standard CCM intervals per Y.1731 */
const CCM_INTERVAL_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 3, label: '3.33 ms' },
  { value: 10, label: '10 ms' },
  { value: 100, label: '100 ms' },
  { value: 1000, label: '1 s' },
  { value: 10000, label: '10 s' },
  { value: 60000, label: '1 min' },
  { value: 600000, label: '10 min' },
];

/** Frame size options */
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

export function Y1731ConfigForm({
  config,
  setConfig,
  selectedTests,
}: Y1731ConfigFormProps): ReactElement | null {
  const hasY1731Tests = selectedTests.some((t) => t.startsWith('y1731'));

  if (!hasY1731Tests) {
    return null;
  }

  const updateConfig = (updates: Partial<Y1731Config>): void => {
    setConfig({ ...config, ...updates });
  };

  const hasDelay: boolean = selectedTests.includes('y1731_delay');
  const hasLoss: boolean = selectedTests.includes('y1731_loss');
  const hasSLM: boolean = selectedTests.includes('y1731_slm');
  const hasLoopback: boolean = selectedTests.includes('y1731_loopback');

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Gauge className="w-4 h-4" />
          <span>Y.1731 OAM Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        {/* MEP/MEG Configuration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            MEP/MEG Configuration
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label
                htmlFor="y1731-mepid"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                MEP ID
                <HelpIcon tooltip="Maintenance End Point identifier (1-8191)." />
              </label>
              <input
                id="y1731-mepid"
                type="number"
                min={1}
                max={8191}
                step={1}
                value={config.mepId}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ mepId: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="y1731-meglevel"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                MEG Level
                <HelpIcon tooltip="Maintenance Entity Group level (0-7). Higher = wider domain." />
              </label>
              <input
                id="y1731-meglevel"
                type="number"
                min={0}
                max={7}
                step={1}
                value={config.megLevel}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ megLevel: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="y1731-megid"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                MEG ID
                <HelpIcon tooltip="Maintenance Entity Group identifier string." />
              </label>
              <input
                id="y1731-megid"
                type="text"
                maxLength={48}
                value={config.megId}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ megId: e.target.value })
                }
                className="mt-1 w-full"
              />
            </div>
          </div>
        </div>

        {/* OAM Parameters */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            OAM Parameters
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="y1731-ccm"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                CCM Interval
                <HelpIcon tooltip="Continuity Check Message interval per Y.1731." />
              </label>
              <select
                id="y1731-ccm"
                value={config.ccmInterval}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
                  updateConfig({ ccmInterval: Number(e.target.value) })
                }
                className="mt-1 w-full"
              >
                {CCM_INTERVAL_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label
                htmlFor="y1731-priority"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Priority
                <HelpIcon tooltip="802.1p priority value (0-7). 6-7 typically for OAM." />
              </label>
              <input
                id="y1731-priority"
                type="number"
                min={0}
                max={7}
                step={1}
                value={config.priority}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ priority: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>
          </div>

          <div className="flex items-center gap-2">
            <input
              id="y1731-tagged"
              type="checkbox"
              checked={config.priorityTagged}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ priorityTagged: e.target.checked })
              }
              aria-label="Use 802.1Q priority tagging on OAM frames"
              className="rounded border-[var(--color-surface-border)]"
            />
            <label
              htmlFor="y1731-tagged"
              title="Add an IEEE 802.1Q VLAN tag with the selected priority to all Y.1731 OAM frames"
              className="text-sm text-[var(--color-text-primary)]"
            >
              Use priority tagging (802.1Q)
            </label>
          </div>
        </div>

        {/* Measurement Parameters */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Measurement Parameters
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="y1731-framesize"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Frame Size
                <HelpIcon tooltip="OAM frame size for measurements." />
              </label>
              <select
                id="y1731-framesize"
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
                htmlFor="y1731-interval"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Measurement Interval (ms)
                <HelpIcon tooltip="Interval between measurement probes." />
              </label>
              <input
                id="y1731-interval"
                type="number"
                min={10}
                max={10000}
                step={10}
                value={config.intervalMs}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ intervalMs: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="y1731-count"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Frames per Interval
                <HelpIcon tooltip="Number of measurement frames per interval." />
              </label>
              <input
                id="y1731-count"
                type="number"
                min={1}
                max={1000}
                step={1}
                value={config.count}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ count: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="y1731-duration"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Duration (s)
                <HelpIcon tooltip="Total measurement duration." />
              </label>
              <input
                id="y1731-duration"
                type="number"
                min={10}
                max={86400}
                step={1}
                value={config.duration}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ duration: Number(e.target.value) })
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
            Test Summary
          </div>
          <div className="text-xs text-[var(--color-text-muted)] space-y-1">
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
              MEP: {config.mepId} | Level: {config.megLevel} | MEG: {config.megId}
            </div>
            <div>
              CCM:{' '}
              {CCM_INTERVAL_OPTIONS.find((c) => c.value === config.ccmInterval)?.label ||
                `${config.ccmInterval}ms`}{' '}
              | Priority: {config.priority}
              {config.priorityTagged ? ' (tagged)' : ''}
            </div>
            <div>
              Frame: {config.frameSize}B | Interval: {config.intervalMs}ms | Count: {config.count}
            </div>
            <div>Duration: {config.duration}s</div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default Y1731ConfigForm;

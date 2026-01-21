/**
 * @fileoverview The Stem - Y.1564 Service Activation Test Configuration
 * @description Advanced configuration form for ITU-T Y.1564 / MEF Service Activation Testing.
 *              Allows users to configure service parameters including CIR, EIR, CBS, EBS,
 *              frame sizes, test duration, and VLAN settings.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { AlertTriangle, Info, Settings2 } from 'lucide-react';
import type React from 'react';
import type { ReactElement } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** Y.1564 service configuration parameters */
export interface Y1564Config {
  /** Committed Information Rate in Mbps */
  cir: number;
  /** Excess Information Rate in Mbps */
  eir: number;
  /** Committed Burst Size in KB */
  cbs: number;
  /** Excess Burst Size in KB */
  ebs: number;
  /** Frame sizes to test */
  frameSizes: number[];
  /** Configuration test step duration in seconds */
  configStepDuration: number;
  /** Performance test duration in seconds */
  perfTestDuration: number;
  /** VLAN ID (0 = untagged) */
  vlanId: number;
  /** Priority Code Point (0-7) */
  pcp: number;
  /** Color-aware mode enabled */
  colorAware: boolean;
  /** Frame Loss Ratio threshold (percentage) */
  flrThreshold: number;
  /** Frame Delay threshold (ms) */
  fdThreshold: number;
  /** Frame Delay Variation threshold (ms) */
  fdvThreshold: number;
}

/** Default Y.1564 configuration */
export const defaultY1564Config: Y1564Config = {
  cir: 100,
  eir: 0,
  cbs: 12,
  ebs: 0,
  frameSizes: [64, 128, 256, 512, 1024, 1280, 1518],
  configStepDuration: 15,
  perfTestDuration: 900,
  vlanId: 0,
  pcp: 0,
  colorAware: false,
  flrThreshold: 0.01,
  fdThreshold: 10,
  fdvThreshold: 5,
};

interface Y1564ConfigFormProps {
  config: Y1564Config;
  setConfig: (config: Y1564Config) => void;
  selectedTests: string[];
}

/** Standard Ethernet frame sizes */
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

export function Y1564ConfigForm({
  config,
  setConfig,
  selectedTests,
}: Y1564ConfigFormProps): ReactElement | null {
  const hasY1564Tests = selectedTests.some((t) => t.startsWith('y1564') || t.startsWith('mef'));

  if (!hasY1564Tests) {
    return null;
  }

  const updateConfig = (updates: Partial<Y1564Config>): void => {
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

  const isConfigTest: boolean = selectedTests.some((t) => t.includes('config'));
  const isPerfTest: boolean = selectedTests.some((t) => t.includes('perf'));
  const isFullTest: boolean = selectedTests.some((t) => t.includes('full'));

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Settings2 className="w-4 h-4" />
          <span>Y.1564 / MEF Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        {/* Bandwidth Configuration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Bandwidth Parameters
          </div>

          {/* CIR */}
          <div>
            <label
              htmlFor="y1564-cir"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              CIR (Mbps)
              <HelpIcon tooltip="Committed Information Rate - the guaranteed bandwidth that the service will always provide." />
            </label>
            <input
              id="y1564-cir"
              type="number"
              min={1}
              max={10000}
              step={1}
              value={config.cir}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ cir: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>

          {/* EIR */}
          <div>
            <label
              htmlFor="y1564-eir"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              EIR (Mbps)
              <HelpIcon tooltip="Excess Information Rate - additional bandwidth available when network capacity permits (best-effort)." />
            </label>
            <input
              id="y1564-eir"
              type="number"
              min={0}
              max={10000}
              step={1}
              value={config.eir}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ eir: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>

          {/* CBS */}
          <div>
            <label
              htmlFor="y1564-cbs"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              CBS (KB)
              <HelpIcon tooltip="Committed Burst Size - maximum burst of traffic allowed at CIR before excess marking." />
            </label>
            <input
              id="y1564-cbs"
              type="number"
              min={1}
              max={1024}
              step={1}
              value={config.cbs}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ cbs: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>

          {/* EBS */}
          <div>
            <label
              htmlFor="y1564-ebs"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              EBS (KB)
              <HelpIcon tooltip="Excess Burst Size - maximum burst of excess traffic (beyond CIR) that may be allowed." />
            </label>
            <input
              id="y1564-ebs"
              type="number"
              min={0}
              max={1024}
              step={1}
              value={config.ebs}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ ebs: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>
        </div>

        {/* SLA Thresholds */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            SLA Thresholds
          </div>

          {/* Frame Loss Ratio */}
          <div>
            <label
              htmlFor="y1564-flr"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Max Frame Loss (%)
              <HelpIcon tooltip="Maximum acceptable Frame Loss Ratio (FLR). Typical SLA values: 0.01% to 0.1%." />
            </label>
            <input
              id="y1564-flr"
              type="number"
              min={0}
              max={100}
              step={0.001}
              value={config.flrThreshold}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ flrThreshold: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>

          {/* Frame Delay */}
          <div>
            <label
              htmlFor="y1564-fd"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Max Frame Delay (ms)
              <HelpIcon tooltip="Maximum acceptable one-way Frame Delay (FD). Typical values: 5-50ms depending on service class." />
            </label>
            <input
              id="y1564-fd"
              type="number"
              min={1}
              max={1000}
              step={1}
              value={config.fdThreshold}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ fdThreshold: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>

          {/* Frame Delay Variation */}
          <div>
            <label
              htmlFor="y1564-fdv"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Max Jitter (ms)
              <HelpIcon tooltip="Maximum acceptable Frame Delay Variation (FDV/jitter). Typical values: 2-10ms." />
            </label>
            <input
              id="y1564-fdv"
              type="number"
              min={1}
              max={100}
              step={1}
              value={config.fdvThreshold}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ fdvThreshold: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>
        </div>

        {/* Frame Sizes */}
        <div className="space-y-2">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide flex items-center gap-1">
            Frame Sizes
            <HelpIcon tooltip="Select frame sizes to test. RFC 2544 recommends: 64, 128, 256, 512, 1024, 1280, 1518 bytes." />
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
          {config.frameSizes.length === 0 && (
            <div className="flex items-center gap-2 text-xs text-[var(--color-status-warning)]">
              <AlertTriangle className="w-3 h-3" />
              Select at least one frame size
            </div>
          )}
        </div>

        {/* Test Duration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Test Duration
          </div>

          {isConfigTest || isFullTest ? (
            <div>
              <label
                htmlFor="y1564-config-duration"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Config Step Duration (s)
                <HelpIcon tooltip="Duration for each step load (25%, 50%, 75%, 100% of CIR) during configuration test." />
              </label>
              <input
                id="y1564-config-duration"
                type="number"
                min={5}
                max={300}
                step={1}
                value={config.configStepDuration}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ configStepDuration: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
              <div className="text-xs text-[var(--color-text-muted)] mt-1">
                Total config test: ~{config.configStepDuration * 4 * config.frameSizes.length}s
              </div>
            </div>
          ) : null}

          {isPerfTest || isFullTest ? (
            <div>
              <label
                htmlFor="y1564-perf-duration"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Performance Duration (s)
                <HelpIcon tooltip="Duration of the sustained performance test at 100% CIR. Y.1564 recommends minimum 15 minutes (900s)." />
              </label>
              <input
                id="y1564-perf-duration"
                type="number"
                min={60}
                max={86400}
                step={60}
                value={config.perfTestDuration}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ perfTestDuration: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
              <div className="text-xs text-[var(--color-text-muted)] mt-1">
                = {Math.floor(config.perfTestDuration / 60)} minutes
              </div>
            </div>
          ) : null}
        </div>

        {/* VLAN Configuration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            VLAN Settings
          </div>

          <div>
            <label
              htmlFor="y1564-vlan"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              VLAN ID
              <HelpIcon tooltip="VLAN tag for test traffic. Set to 0 for untagged traffic." />
            </label>
            <input
              id="y1564-vlan"
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
            {config.vlanId === 0 && (
              <div className="text-xs text-[var(--color-text-muted)] mt-1">Untagged traffic</div>
            )}
          </div>

          {config.vlanId > 0 && (
            <div>
              <label
                htmlFor="y1564-pcp"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Priority (PCP)
                <HelpIcon tooltip="Priority Code Point (802.1p). 0=Best Effort, 7=Highest Priority." />
              </label>
              <select
                id="y1564-pcp"
                value={config.pcp}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
                  updateConfig({ pcp: Number(e.target.value) })
                }
                className="mt-1 w-full"
              >
                <option value={0}>0 - Best Effort (BE)</option>
                <option value={1}>1 - Background (BK)</option>
                <option value={2}>2 - Excellent Effort (EE)</option>
                <option value={3}>3 - Critical (CA)</option>
                <option value={4}>4 - Video (VI)</option>
                <option value={5}>5 - Voice (VO)</option>
                <option value={6}>6 - Internetwork Control (IC)</option>
                <option value={7}>7 - Network Control (NC)</option>
              </select>
            </div>
          )}

          {/* Color-Aware Mode */}
          <label className="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]">
            <input
              type="checkbox"
              checked={config.colorAware}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ colorAware: e.target.checked })
              }
              className="w-4 h-4 accent-[var(--color-brand-primary)]"
            />
            <div>
              <div className="font-medium text-sm flex items-center gap-1">
                Color-Aware Mode
                <HelpIcon tooltip="Enable dual-rate testing with green (CIR) and yellow (EIR) traffic classes." />
              </div>
              <div className="text-xs text-[var(--color-text-muted)]">
                Test green/yellow traffic separation
              </div>
            </div>
          </label>
        </div>

        {/* Summary */}
        <div className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
          <div className="flex items-center gap-2 text-sm font-medium text-[var(--color-text-primary)] mb-2">
            <Info className="w-4 h-4" />
            Test Summary
          </div>
          <div className="text-xs text-[var(--color-text-muted)] space-y-1">
            <div>
              Service: {config.cir} Mbps CIR
              {config.eir > 0 && ` + ${config.eir} Mbps EIR`}
            </div>
            <div>Frame sizes: {config.frameSizes.join(', ')} bytes</div>
            <div>
              SLA: FLR≤{config.flrThreshold}%, FD≤{config.fdThreshold}ms, FDV≤
              {config.fdvThreshold}
              ms
            </div>
            {config.vlanId > 0 && (
              <div>
                VLAN {config.vlanId} / PCP {config.pcp}
              </div>
            )}
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default Y1564ConfigForm;

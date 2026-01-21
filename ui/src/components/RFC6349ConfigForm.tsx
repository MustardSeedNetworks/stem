/**
 * @fileoverview The Stem - RFC 6349 TCP Throughput Test Configuration
 * @description Configuration form for RFC 6349 TCP Throughput Testing.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { Activity, Info } from 'lucide-react';
import type { ReactElement } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** RFC 6349 test configuration parameters */
export interface RFC6349Config {
  /** Target rate in Mbps */
  targetRateMbps: number;
  /** Minimum RTT in milliseconds */
  minRTTMs: number;
  /** Maximum RTT in milliseconds */
  maxRTTMs: number;
  /** Receive window size in bytes */
  rwndSize: number;
  /** Test duration in seconds */
  duration: number;
  /** Number of parallel TCP streams */
  parallelStreams: number;
  /** Maximum Segment Size */
  mss: number;
  /** Test mode: 0=bidirectional, 1=upstream, 2=downstream */
  mode: number;
}

/** Default RFC 6349 configuration */
export const defaultRFC6349Config: RFC6349Config = {
  targetRateMbps: 100,
  minRTTMs: 1,
  maxRTTMs: 100,
  rwndSize: 65535,
  duration: 30,
  parallelStreams: 1,
  mss: 1460,
  mode: 0,
};

/** Test mode options */
const MODE_OPTIONS: Array<{ value: number; label: string; description: string }> = [
  { value: 0, label: 'Bidirectional', description: 'Test both directions' },
  { value: 1, label: 'Upstream', description: 'Client to server only' },
  { value: 2, label: 'Downstream', description: 'Server to client only' },
];

/** Common MSS values */
const MSS_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 536, label: '536 B (min)' },
  { value: 1220, label: '1220 B (IPv6)' },
  { value: 1460, label: '1460 B (standard)' },
  { value: 8960, label: '8960 B (jumbo)' },
];

/** Format BDP value in appropriate unit */
function formatBDP(bdpBytes: number): string {
  if (bdpBytes >= 1048576) {
    return `${(bdpBytes / 1048576).toFixed(2)} MB`;
  }
  if (bdpBytes >= 1024) {
    return `${(bdpBytes / 1024).toFixed(2)} KB`;
  }
  return `${bdpBytes.toFixed(0)} B`;
}

interface RFC6349ConfigFormProps {
  config: RFC6349Config;
  setConfig: (config: RFC6349Config) => void;
  selectedTests: string[];
}

export function RFC6349ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC6349ConfigFormProps): ReactElement | null {
  const hasRFC6349Tests = selectedTests.some((t) => t.startsWith('rfc6349'));

  if (!hasRFC6349Tests) {
    return null;
  }

  const updateConfig = (updates: Partial<RFC6349Config>): void => {
    setConfig({ ...config, ...updates });
  };

  const hasThroughput = selectedTests.includes('rfc6349_throughput');
  const hasBDP = selectedTests.includes('rfc6349_bdp');
  const hasEfficiency = selectedTests.includes('rfc6349_efficiency');

  // Calculate Bandwidth-Delay Product
  const bdp = (config.targetRateMbps * 1000000 * config.maxRTTMs) / 8000;
  const bdpFormatted = formatBDP(bdp);

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Activity className="w-4 h-4" />
          <span>RFC 6349 Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        {/* Target Rate and RTT */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Network Parameters
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label
                htmlFor="rfc6349-rate"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Target Rate (Mbps)
                <HelpIcon tooltip="Target throughput rate for TCP testing." />
              </label>
              <input
                id="rfc6349-rate"
                type="number"
                min={1}
                max={100000}
                step={1}
                value={config.targetRateMbps}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ targetRateMbps: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="rfc6349-minrtt"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Min RTT (ms)
                <HelpIcon tooltip="Minimum expected round-trip time." />
              </label>
              <input
                id="rfc6349-minrtt"
                type="number"
                min={0.1}
                max={1000}
                step={0.1}
                value={config.minRTTMs}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ minRTTMs: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="rfc6349-maxrtt"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Max RTT (ms)
                <HelpIcon tooltip="Maximum expected round-trip time." />
              </label>
              <input
                id="rfc6349-maxrtt"
                type="number"
                min={0.1}
                max={5000}
                step={0.1}
                value={config.maxRTTMs}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ maxRTTMs: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>
          </div>
        </div>

        {/* TCP Parameters */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            TCP Parameters
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="rfc6349-rwnd"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                RWND Size (bytes)
                <HelpIcon tooltip="TCP Receive Window size. Should be >= BDP for optimal throughput." />
              </label>
              <input
                id="rfc6349-rwnd"
                type="number"
                min={4096}
                max={16777216}
                step={1024}
                value={config.rwndSize}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ rwndSize: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="rfc6349-mss"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                MSS
                <HelpIcon tooltip="Maximum Segment Size for TCP." />
              </label>
              <select
                id="rfc6349-mss"
                value={config.mss}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
                  updateConfig({ mss: Number(e.target.value) })
                }
                className="mt-1 w-full"
              >
                {MSS_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="rfc6349-streams"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Parallel Streams
                <HelpIcon tooltip="Number of parallel TCP connections." />
              </label>
              <input
                id="rfc6349-streams"
                type="number"
                min={1}
                max={128}
                step={1}
                value={config.parallelStreams}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ parallelStreams: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="rfc6349-mode"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Test Mode
                <HelpIcon tooltip="Direction of throughput testing." />
              </label>
              <select
                id="rfc6349-mode"
                value={config.mode}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
                  updateConfig({ mode: Number(e.target.value) })
                }
                className="mt-1 w-full"
              >
                {MODE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Duration */}
        <div>
          <label
            htmlFor="rfc6349-duration"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Duration (s)
            <HelpIcon tooltip="Test duration in seconds." />
          </label>
          <input
            id="rfc6349-duration"
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

        {/* Test Summary */}
        <div className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
          <div className="flex items-center gap-2 text-sm font-medium text-[var(--color-text-primary)] mb-2">
            <Info className="w-4 h-4" />
            Test Summary
          </div>
          <div className="text-xs text-[var(--color-text-muted)] space-y-1">
            <div>
              Selected tests:{' '}
              {[hasThroughput && 'Throughput', hasBDP && 'BDP', hasEfficiency && 'Efficiency']
                .filter(Boolean)
                .join(', ')}
            </div>
            <div>Target: {config.targetRateMbps} Mbps</div>
            <div>
              RTT Range: {config.minRTTMs} - {config.maxRTTMs} ms
            </div>
            <div>
              BDP (calculated): {bdpFormatted}
              {config.rwndSize < bdp ? (
                <span className="text-[var(--color-status-warning)] ml-2">RWND &lt; BDP</span>
              ) : null}
            </div>
            <div>
              Mode: {MODE_OPTIONS.find((m) => m.value === config.mode)?.label} | Streams:{' '}
              {config.parallelStreams}
            </div>
            <div>Duration: {config.duration}s</div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default RFC6349ConfigForm;

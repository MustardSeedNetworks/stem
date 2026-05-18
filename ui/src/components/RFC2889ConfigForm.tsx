/**
 * @fileoverview The Stem - RFC 2889 LAN Switch Test Configuration
 * @description Configuration form for RFC 2889 LAN Switch Benchmarking Tests.
 */

import { Info, Network } from 'lucide-react';
import type { ReactElement } from 'react';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** RFC 2889 test configuration parameters */
export interface RFC2889Config {
  /** Frame size in bytes */
  frameSize: number;
  /** Test duration in seconds */
  duration: number;
  /** Warmup duration before measurement (seconds) */
  warmup: number;
  /** Number of MAC addresses for learning/caching tests */
  addressCount: number;
  /** Maximum acceptable frame loss (percentage) */
  acceptableLoss: number;
  /** Number of ports to test */
  portCount: number;
  /** Traffic pattern: 0=mesh, 1=pair, 2=broadcast */
  pattern: number;
}

/** Default RFC 2889 configuration */
export const defaultRFC2889Config: RFC2889Config = {
  frameSize: 64,
  duration: 60,
  warmup: 2,
  addressCount: 8192,
  acceptableLoss: 0.0,
  portCount: 2,
  pattern: 0,
};

/** Standard frame sizes for RFC 2889 */
const FRAME_SIZE_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 64, label: '64 B (min)' },
  { value: 128, label: '128 B' },
  { value: 256, label: '256 B' },
  { value: 512, label: '512 B' },
  { value: 1024, label: '1024 B' },
  { value: 1280, label: '1280 B' },
  { value: 1518, label: '1518 B (max)' },
];

/** Traffic pattern options */
const PATTERN_OPTIONS: Array<{ value: number; label: string; description: string }> = [
  { value: 0, label: 'Full Mesh', description: 'All ports to all ports' },
  { value: 1, label: 'Pair', description: 'Port pairs (1→2, 3→4, etc.)' },
  { value: 2, label: 'Broadcast', description: 'One port to all others' },
];

interface RFC2889ConfigFormProps {
  config: RFC2889Config;
  setConfig: (config: RFC2889Config) => void;
  selectedTests: string[];
}

export function RFC2889ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC2889ConfigFormProps): ReactElement | null {
  const hasRFC2889Tests = selectedTests.some((t) => t.startsWith('rfc2889'));

  if (!hasRFC2889Tests) {
    return null;
  }

  const updateConfig = (updates: Partial<RFC2889Config>): void => {
    setConfig({ ...config, ...updates });
  };

  const hasForwarding = selectedTests.includes('rfc2889_forwarding');
  const hasCaching = selectedTests.includes('rfc2889_caching');
  const hasLearning = selectedTests.includes('rfc2889_learning');
  const hasBroadcast = selectedTests.includes('rfc2889_broadcast');
  const hasCongestion = selectedTests.includes('rfc2889_congestion');

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Network className="w-4 h-4" />
          <span>RFC 2889 Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="space-y-4">
        {/* Test Duration */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Test Parameters
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="rfc2889-duration"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Duration (s)
                <HelpIcon tooltip="Duration for each test iteration in seconds." />
              </label>
              <input
                id="rfc2889-duration"
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
                htmlFor="rfc2889-warmup"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Warmup (s)
                <HelpIcon tooltip="Warmup period before measurement begins." />
              </label>
              <input
                id="rfc2889-warmup"
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

        {/* Frame Size */}
        <div>
          <label
            htmlFor="rfc2889-framesize"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Frame Size
            <HelpIcon tooltip="Ethernet frame size for testing." />
          </label>
          <select
            id="rfc2889-framesize"
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

        {/* Port and Pattern Config */}
        <div className="space-y-3">
          <div className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-wide">
            Switch Configuration
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="rfc2889-portcount"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Port Count
                <HelpIcon tooltip="Number of switch ports to test." />
              </label>
              <input
                id="rfc2889-portcount"
                type="number"
                min={2}
                max={48}
                step={1}
                value={config.portCount}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  updateConfig({ portCount: Number(e.target.value) })
                }
                className="mt-1 w-full"
              />
            </div>

            <div>
              <label
                htmlFor="rfc2889-pattern"
                className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
              >
                Traffic Pattern
                <HelpIcon tooltip="How traffic is distributed across ports." />
              </label>
              <select
                id="rfc2889-pattern"
                value={config.pattern}
                onChange={(e: React.ChangeEvent<HTMLSelectElement>): void =>
                  updateConfig({ pattern: Number(e.target.value) })
                }
                className="mt-1 w-full"
              >
                {PATTERN_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Learning/Caching Config */}
        {hasCaching || hasLearning ? (
          <div>
            <label
              htmlFor="rfc2889-addresscount"
              className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
            >
              Address Count
              <HelpIcon tooltip="Number of MAC addresses for learning/caching tests. RFC 2889 recommends testing at 1, 10, 100, 1000, 10000 addresses." />
            </label>
            <input
              id="rfc2889-addresscount"
              type="number"
              min={1}
              max={100000}
              step={1}
              value={config.addressCount}
              onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                updateConfig({ addressCount: Number(e.target.value) })
              }
              className="mt-1 w-full"
            />
          </div>
        ) : null}

        {/* Acceptable Loss */}
        <div>
          <label
            htmlFor="rfc2889-loss"
            className="flex items-center gap-1 text-sm font-medium text-[var(--color-text-primary)]"
          >
            Acceptable Loss (%)
            <HelpIcon tooltip="Maximum acceptable frame loss percentage. RFC 2889 specifies 0%." />
          </label>
          <input
            id="rfc2889-loss"
            type="number"
            min={0}
            max={1}
            step={0.001}
            value={config.acceptableLoss}
            onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
              updateConfig({ acceptableLoss: Number(e.target.value) })
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
              {[
                hasForwarding && 'Forwarding',
                hasCaching && 'Caching',
                hasLearning && 'Learning',
                hasBroadcast && 'Broadcast',
                hasCongestion && 'Congestion',
              ]
                .filter(Boolean)
                .join(', ')}
            </div>
            <div>Frame size: {config.frameSize} bytes</div>
            <div>
              Ports: {config.portCount} | Pattern:{' '}
              {PATTERN_OPTIONS.find((p) => p.value === config.pattern)?.label}
            </div>
            <div>
              Duration: {config.duration}s + {config.warmup}s warmup
            </div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default RFC2889ConfigForm;

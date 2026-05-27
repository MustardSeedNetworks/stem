/**
 * @fileoverview The Stem - RFC 2889 LAN Switch Test Configuration
 * @description Configuration form for RFC 2889 LAN Switch Benchmarking Tests.
 *              Migrated to react-hook-form + valibot per #325.
 */

import { AlertTriangle, Info, Network } from 'lucide-react';
import type { ReactElement } from 'react';
import { useConfigForm } from '../forms/useConfigForm';
import { RFC2889ConfigSchema } from '../schemas/configs';
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

const FRAME_SIZE_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 64, label: '64 B (min)' },
  { value: 128, label: '128 B' },
  { value: 256, label: '256 B' },
  { value: 512, label: '512 B' },
  { value: 1024, label: '1024 B' },
  { value: 1280, label: '1280 B' },
  { value: 1518, label: '1518 B (max)' },
];

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

function FieldError({ message }: { message?: string }): ReactElement | null {
  if (!message) return null;
  return (
    <div className="mt-tight text-xs text-status-danger flex items-center gap-tight">
      <AlertTriangle className="w-3 h-3" />
      {message}
    </div>
  );
}

export function RFC2889ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC2889ConfigFormProps): ReactElement | null {
  const hasRFC2889Tests = selectedTests.some((t) => t.startsWith('rfc2889'));

  const form = useConfigForm<RFC2889Config>({
    schema: RFC2889ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    formState: { errors },
  } = form;

  if (!hasRFC2889Tests) {
    return null;
  }

  const frameSize = watch('frameSize') ?? 0;
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;
  const portCount = watch('portCount') ?? 0;
  const pattern = watch('pattern') ?? 0;

  const hasForwarding = selectedTests.includes('rfc2889_forwarding');
  const hasCaching = selectedTests.includes('rfc2889_caching');
  const hasLearning = selectedTests.includes('rfc2889_learning');
  const hasBroadcast = selectedTests.includes('rfc2889_broadcast');
  const hasCongestion = selectedTests.includes('rfc2889_congestion');

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Network className="w-4 h-4" />
          <span>RFC 2889 Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Test Parameters
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc2889-duration" className="flex items-center gap-tight label">
                Duration (s)
                <HelpIcon tooltip="Duration for each test iteration in seconds." />
              </label>
              <input
                id="rfc2889-duration"
                type="number"
                step={1}
                {...register('duration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.duration?.message} />
            </div>

            <div>
              <label htmlFor="rfc2889-warmup" className="flex items-center gap-tight label">
                Warmup (s)
                <HelpIcon tooltip="Warmup period before measurement begins." />
              </label>
              <input
                id="rfc2889-warmup"
                type="number"
                step={1}
                {...register('warmup', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.warmup?.message} />
            </div>
          </div>
        </div>

        <div>
          <label htmlFor="rfc2889-framesize" className="flex items-center gap-tight label">
            Frame Size
            <HelpIcon tooltip="Ethernet frame size for testing." />
          </label>
          <select
            id="rfc2889-framesize"
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

        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Switch Configuration
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc2889-portcount" className="flex items-center gap-tight label">
                Port Count
                <HelpIcon tooltip="Number of switch ports to test." />
              </label>
              <input
                id="rfc2889-portcount"
                type="number"
                step={1}
                {...register('portCount', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.portCount?.message} />
            </div>

            <div>
              <label htmlFor="rfc2889-pattern" className="flex items-center gap-tight label">
                Traffic Pattern
                <HelpIcon tooltip="How traffic is distributed across ports." />
              </label>
              <select
                id="rfc2889-pattern"
                {...register('pattern', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {PATTERN_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <FieldError message={errors.pattern?.message} />
            </div>
          </div>
        </div>

        {hasCaching || hasLearning ? (
          <div>
            <label htmlFor="rfc2889-addresscount" className="flex items-center gap-tight label">
              Address Count
              <HelpIcon tooltip="Number of MAC addresses for learning/caching tests. RFC 2889 recommends testing at 1, 10, 100, 1000, 10000 addresses." />
            </label>
            <input
              id="rfc2889-addresscount"
              type="number"
              step={1}
              {...register('addressCount', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.addressCount?.message} />
          </div>
        ) : null}

        <div>
          <label htmlFor="rfc2889-loss" className="flex items-center gap-tight label">
            Acceptable Loss (%)
            <HelpIcon tooltip="Maximum acceptable frame loss percentage. RFC 2889 specifies 0%." />
          </label>
          <input
            id="rfc2889-loss"
            type="number"
            step={0.001}
            {...register('acceptableLoss', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.acceptableLoss?.message} />
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
                hasForwarding && 'Forwarding',
                hasCaching && 'Caching',
                hasLearning && 'Learning',
                hasBroadcast && 'Broadcast',
                hasCongestion && 'Congestion',
              ]
                .filter(Boolean)
                .join(', ')}
            </div>
            <div>Frame size: {frameSize} bytes</div>
            <div>
              Ports: {portCount} | Pattern:{' '}
              {PATTERN_OPTIONS.find((p) => p.value === pattern)?.label}
            </div>
            <div>
              Duration: {duration}s + {warmup}s warmup
            </div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default RFC2889ConfigForm;

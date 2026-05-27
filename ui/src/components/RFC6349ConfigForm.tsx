/**
 * @fileoverview The Stem - RFC 6349 TCP Throughput Test Configuration
 * @description Configuration form for RFC 6349 TCP Throughput Testing.
 *              Migrated to react-hook-form + valibot per #325; the
 *              cross-field rule (minRTT ≤ maxRTT) is enforced by the
 *              schema and surfaced via the form footer.
 */

import { Activity, AlertTriangle, Info } from 'lucide-react';
import type { ReactElement } from 'react';
import { useConfigForm } from '../forms/useConfigForm';
import { RFC6349ConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { HelpIcon } from './HelpIcon';

/** RFC 6349 test configuration parameters */
export interface RFC6349Config {
  targetRateMbps: number;
  minRTTMs: number;
  maxRTTMs: number;
  rwndSize: number;
  duration: number;
  parallelStreams: number;
  mss: number;
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

const MODE_OPTIONS: Array<{ value: number; label: string; description: string }> = [
  { value: 0, label: 'Bidirectional', description: 'Test both directions' },
  { value: 1, label: 'Upstream', description: 'Client to server only' },
  { value: 2, label: 'Downstream', description: 'Server to client only' },
];

const MSS_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 536, label: '536 B (min)' },
  { value: 1220, label: '1220 B (IPv6)' },
  { value: 1460, label: '1460 B (standard)' },
  { value: 8960, label: '8960 B (jumbo)' },
];

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

function FieldError({ message }: { message?: string }): ReactElement | null {
  if (!message) return null;
  return (
    <div className="mt-tight text-xs text-status-danger flex items-center gap-tight">
      <AlertTriangle className="w-3 h-3" />
      {message}
    </div>
  );
}

export function RFC6349ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC6349ConfigFormProps): ReactElement | null {
  const hasRFC6349Tests = selectedTests.some((t) => t.startsWith('rfc6349'));

  const form = useConfigForm<RFC6349Config>({
    schema: RFC6349ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    formState: { errors },
  } = form;

  if (!hasRFC6349Tests) {
    return null;
  }

  const targetRateMbps = watch('targetRateMbps') ?? 0;
  const minRTTMs = watch('minRTTMs') ?? 0;
  const maxRTTMs = watch('maxRTTMs') ?? 0;
  const rwndSize = watch('rwndSize') ?? 0;
  const duration = watch('duration') ?? 0;
  const parallelStreams = watch('parallelStreams') ?? 0;
  const mode = watch('mode') ?? 0;

  const hasThroughput = selectedTests.includes('rfc6349_throughput');
  const hasBDP = selectedTests.includes('rfc6349_bdp');
  const hasEfficiency = selectedTests.includes('rfc6349_efficiency');

  const bdp = (targetRateMbps * 1000000 * maxRTTMs) / 8000;
  const bdpFormatted = formatBDP(bdp);

  // Cross-field error (minRTT > maxRTT) from valibot v.check().
  const rootErrors = errors.root;
  const crossFieldError = rootErrors
    ? Object.values(rootErrors).find(
        (e): e is { message: string } =>
          typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string',
      )
    : undefined;

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Activity className="w-4 h-4" />
          <span>RFC 6349 Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            Network Parameters
          </div>

          <div className="grid grid-cols-3 gap-default">
            <div>
              <label htmlFor="rfc6349-rate" className="flex items-center gap-tight label">
                Target Rate (Mbps)
                <HelpIcon tooltip="Target throughput rate for TCP testing." />
              </label>
              <input
                id="rfc6349-rate"
                type="number"
                step={1}
                {...register('targetRateMbps', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.targetRateMbps?.message} />
            </div>

            <div>
              <label htmlFor="rfc6349-minrtt" className="flex items-center gap-tight label">
                Min RTT (ms)
                <HelpIcon tooltip="Minimum expected round-trip time." />
              </label>
              <input
                id="rfc6349-minrtt"
                type="number"
                step={0.1}
                {...register('minRTTMs', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.minRTTMs?.message} />
            </div>

            <div>
              <label htmlFor="rfc6349-maxrtt" className="flex items-center gap-tight label">
                Max RTT (ms)
                <HelpIcon tooltip="Maximum expected round-trip time." />
              </label>
              <input
                id="rfc6349-maxrtt"
                type="number"
                step={0.1}
                {...register('maxRTTMs', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.maxRTTMs?.message} />
            </div>
          </div>
        </div>

        <div className="stack">
          <div className="text-xs font-semibold text-text-muted uppercase tracking-wide">
            TCP Parameters
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc6349-rwnd" className="flex items-center gap-tight label">
                RWND Size (bytes)
                <HelpIcon tooltip="TCP Receive Window size. Should be >= BDP for optimal throughput." />
              </label>
              <input
                id="rfc6349-rwnd"
                type="number"
                step={1024}
                {...register('rwndSize', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.rwndSize?.message} />
            </div>

            <div>
              <label htmlFor="rfc6349-mss" className="flex items-center gap-tight label">
                MSS
                <HelpIcon tooltip="Maximum Segment Size for TCP." />
              </label>
              <select
                id="rfc6349-mss"
                {...register('mss', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {MSS_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <FieldError message={errors.mss?.message} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc6349-streams" className="flex items-center gap-tight label">
                Parallel Streams
                <HelpIcon tooltip="Number of parallel TCP connections." />
              </label>
              <input
                id="rfc6349-streams"
                type="number"
                step={1}
                {...register('parallelStreams', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.parallelStreams?.message} />
            </div>

            <div>
              <label htmlFor="rfc6349-mode" className="flex items-center gap-tight label">
                Test Mode
                <HelpIcon tooltip="Direction of throughput testing." />
              </label>
              <select
                id="rfc6349-mode"
                {...register('mode', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {MODE_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <FieldError message={errors.mode?.message} />
            </div>
          </div>
        </div>

        <div>
          <label htmlFor="rfc6349-duration" className="flex items-center gap-tight label">
            Duration (s)
            <HelpIcon tooltip="Test duration in seconds." />
          </label>
          <input
            id="rfc6349-duration"
            type="number"
            step={1}
            {...register('duration', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.duration?.message} />
        </div>

        {crossFieldError && (
          <div className="pad-xs rounded-lg bg-status-danger-subtle text-status-danger text-sm flex items-center gap-compact">
            <AlertTriangle className="w-4 h-4" />
            {crossFieldError.message}
          </div>
        )}

        <div className="pad-sm rounded-lg bg-surface-base border border-surface-border">
          <div className="flex items-center gap-compact label mb-2">
            <Info className="w-4 h-4" />
            Test Summary
          </div>
          <div className="text-xs text-text-muted stack-xs">
            <div>
              Selected tests:{' '}
              {[hasThroughput && 'Throughput', hasBDP && 'BDP', hasEfficiency && 'Efficiency']
                .filter(Boolean)
                .join(', ')}
            </div>
            <div>Target: {targetRateMbps} Mbps</div>
            <div>
              RTT Range: {minRTTMs} - {maxRTTMs} ms
            </div>
            <div>
              BDP (calculated): {bdpFormatted}
              {rwndSize < bdp ? (
                <span className="text-status-warning ml-inline">RWND &lt; BDP</span>
              ) : null}
            </div>
            <div>
              Mode: {MODE_OPTIONS.find((m) => m.value === mode)?.label} | Streams: {parallelStreams}
            </div>
            <div>Duration: {duration}s</div>
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default RFC6349ConfigForm;

/**
 * @fileoverview The Stem - RFC 6349 TCP Throughput Test Configuration
 * @description Configuration form for RFC 6349 TCP Throughput Testing.
 *              Migrated to react-hook-form + valibot per #325; the
 *              cross-field rule (minRTT ≤ maxRTT) is enforced by the
 *              schema and surfaced via the form footer.
 */

import { Activity, AlertTriangle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { useConfigForm } from '../forms/useConfigForm';
import { RFC6349ConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
import { TestSummary } from './TestSummary';

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

/** Wire values, with the key that names each for the operator. */
const MODE_OPTIONS: Array<{ value: number; key: string }> = [
  { value: 0, key: 'modeBidirectional' },
  { value: 1, key: 'modeUpstream' },
  { value: 2, key: 'modeDownstream' },
];

/** Segment sizes, with the qualifier as a key so only the qualifier moves. */
const MSS_OPTIONS: Array<{ value: number; key: string }> = [
  { value: 536, key: 'mssMin' },
  { value: 1220, key: 'mssIpv6' },
  { value: 1460, key: 'mssStandard' },
  { value: 8960, key: 'mssJumbo' },
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
  const { t } = useTranslation('settings');

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
      testId="rfc6349-config-form"
      title={
        <div className="flex items-center gap-compact">
          <Activity className="w-4 h-4" />
          <span>{t('testConfig.rfc6349.title')}</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <FormSection title={t('testConfig.rfc6349.network.title')}>
          <div className="grid grid-cols-3 gap-default">
            <div>
              <label htmlFor="rfc6349-rate" className="flex items-center gap-tight label">
                {t('testConfig.rfc6349.network.targetRate')}
                <HelpIcon tooltip={t('testConfig.rfc6349.network.targetRateHelp')} />
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
                {t('testConfig.rfc6349.network.minRtt')}
                <HelpIcon tooltip={t('testConfig.rfc6349.network.minRttHelp')} />
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
                {t('testConfig.rfc6349.network.maxRtt')}
                <HelpIcon tooltip={t('testConfig.rfc6349.network.maxRttHelp')} />
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
        </FormSection>

        <FormSection title={t('testConfig.rfc6349.tcp.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc6349-rwnd" className="flex items-center gap-tight label">
                {t('testConfig.rfc6349.tcp.rwnd')}
                <HelpIcon tooltip={t('testConfig.rfc6349.tcp.rwndHelp')} />
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
                {t('testConfig.rfc6349.tcp.mss')}
                <HelpIcon tooltip={t('testConfig.rfc6349.tcp.mssHelp')} />
              </label>
              <select
                id="rfc6349-mss"
                {...register('mss', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {MSS_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {t(`testConfig.rfc6349.tcp.${option.key}` as never, { size: option.value })}
                  </option>
                ))}
              </select>
              <FieldError message={errors.mss?.message} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc6349-streams" className="flex items-center gap-tight label">
                {t('testConfig.rfc6349.tcp.streams')}
                <HelpIcon tooltip={t('testConfig.rfc6349.tcp.streamsHelp')} />
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
                {t('testConfig.rfc6349.tcp.mode')}
                <HelpIcon tooltip={t('testConfig.rfc6349.tcp.modeHelp')} />
              </label>
              <select
                id="rfc6349-mode"
                {...register('mode', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {MODE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {t(`testConfig.rfc6349.tcp.${option.key}` as never)}
                  </option>
                ))}
              </select>
              <FieldError message={errors.mode?.message} />
            </div>
          </div>
        </FormSection>

        <div>
          <label htmlFor="rfc6349-duration" className="flex items-center gap-tight label">
            {t('testConfig.rfc6349.tcp.duration')}
            <HelpIcon tooltip={t('testConfig.rfc6349.tcp.durationHelp')} />
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
          <div className="pad-xs rounded-lg bg-status-error/10 text-status-error text-sm flex items-center gap-compact">
            <AlertTriangle className="w-4 h-4" />
            {crossFieldError.message}
          </div>
        )}

        <TestSummary>
          <div>
            {t('testConfig.common.selectedTests')}:{' '}
            {[
              hasThroughput && t('testConfig.rfc6349.tests.throughput'),
              hasBDP && t('testConfig.rfc6349.tests.bdp'),
              hasEfficiency && t('testConfig.rfc6349.tests.efficiency'),
            ]
              .filter(Boolean)
              .join(', ')}
          </div>
          <div>{t('testConfig.rfc6349.summary.target', { rate: targetRateMbps })}</div>
          <div>{t('testConfig.rfc6349.summary.rtt', { min: minRTTMs, max: maxRTTMs })}</div>
          <div>
            {t('testConfig.rfc6349.summary.bdp', { bdp: bdpFormatted })}
            {rwndSize < bdp ? (
              <span className="text-status-warning ml-inline">
                {t('testConfig.rfc6349.summary.rwndBelowBdp')}
              </span>
            ) : null}
          </div>
          <div>
            {t('testConfig.rfc6349.summary.mode', {
              mode: t(
                `testConfig.rfc6349.tcp.${
                  MODE_OPTIONS.find((option) => option.value === mode)?.key ?? 'modeBidirectional'
                }` as never,
              ),
              streams: parallelStreams,
            })}
          </div>
          <div>{t('testConfig.rfc6349.summary.duration', { duration })}</div>
        </TestSummary>
      </div>
    </CollapsibleSection>
  );
}

export default RFC6349ConfigForm;

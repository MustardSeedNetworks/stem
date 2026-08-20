/**
 * @fileoverview The Stem - Y.1731 OAM Test Configuration
 * @description Migrated to react-hook-form + valibot per #325.
 */

import { Gauge } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { FrameSizeOption } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { Y1731ConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
import { TestSummary } from './TestSummary';

/** Y.1731 test configuration parameters */
export interface Y1731Config {
  mepId: number;
  megLevel: number;
  megId: string;
  ccmInterval: number;
  priority: number;
  duration: number;
  intervalMs: number;
  count: number;
  frameSize: number;
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

const CCM_INTERVAL_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 3, label: '3.33 ms' },
  { value: 10, label: '10 ms' },
  { value: 100, label: '100 ms' },
  { value: 1000, label: '1 s' },
  { value: 10000, label: '10 s' },
  { value: 60000, label: '1 min' },
  { value: 600000, label: '10 min' },
];

/**
 * OAM measurement frames are deliberately a subset of the sweep sizes: no
 * 1280 and no jumbo. Kept local rather than filtered out of the shared list,
 * so the set is stated rather than inferred from an exclusion.
 */
const FRAME_SIZE_OPTIONS: FrameSizeOption[] = [
  { value: 64, qualifier: 'frameSizeMin' },
  { value: 128, qualifier: 'frameSize' },
  { value: 256, qualifier: 'frameSize' },
  { value: 512, qualifier: 'frameSize' },
  { value: 1024, qualifier: 'frameSize' },
  { value: 1518, qualifier: 'frameSizeMax' },
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

  const form = useConfigForm<Y1731Config>({
    schema: Y1731ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    formState: { errors },
  } = form;
  const { t } = useTranslation('settings');

  if (!hasY1731Tests) {
    return null;
  }

  const mepId = watch('mepId') ?? 0;
  const megLevel = watch('megLevel') ?? 0;
  const megId = watch('megId') ?? '';
  const ccmInterval = watch('ccmInterval') ?? 0;
  const priority = watch('priority') ?? 0;
  const priorityTagged = watch('priorityTagged') ?? false;
  const frameSize = watch('frameSize') ?? 0;
  const intervalMs = watch('intervalMs') ?? 0;
  const count = watch('count') ?? 0;
  const duration = watch('duration') ?? 0;

  const hasDelay = selectedTests.includes('y1731_delay');
  const hasLoss = selectedTests.includes('y1731_loss');
  const hasSLM = selectedTests.includes('y1731_slm');
  const hasLoopback = selectedTests.includes('y1731_loopback');

  return (
    <CollapsibleSection
      testId="y1731-config-form"
      title={
        <div className="flex items-center gap-compact">
          <Gauge className="w-4 h-4" />
          <span>{t('testConfig.y1731.title')}</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <FormSection title={t('testConfig.y1731.mep.title')}>
          <div className="grid grid-cols-3 gap-default">
            <div>
              <label htmlFor="y1731-mepid" className="flex items-center gap-tight label">
                {t('testConfig.y1731.mep.mepId')}
                <HelpIcon tooltip={t('testConfig.y1731.mep.mepIdHelp')} />
              </label>
              <input
                id="y1731-mepid"
                type="number"
                step={1}
                {...register('mepId', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.mepId?.message} />
            </div>

            <div>
              <label htmlFor="y1731-meglevel" className="flex items-center gap-tight label">
                {t('testConfig.y1731.mep.megLevel')}
                <HelpIcon tooltip={t('testConfig.y1731.mep.megLevelHelp')} />
              </label>
              <input
                id="y1731-meglevel"
                type="number"
                step={1}
                {...register('megLevel', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.megLevel?.message} />
            </div>

            <div>
              <label htmlFor="y1731-megid" className="flex items-center gap-tight label">
                {t('testConfig.y1731.mep.megId')}
                <HelpIcon tooltip={t('testConfig.y1731.mep.megIdHelp')} />
              </label>
              <input
                id="y1731-megid"
                type="text"
                maxLength={45}
                {...register('megId')}
                className="mt-tight w-full"
              />
              <FieldError message={errors.megId?.message} />
            </div>
          </div>
        </FormSection>

        <FormSection title={t('testConfig.y1731.oam.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="y1731-ccm" className="flex items-center gap-tight label">
                {t('testConfig.y1731.oam.ccm')}
                <HelpIcon tooltip={t('testConfig.y1731.oam.ccmHelp')} />
              </label>
              <select
                id="y1731-ccm"
                {...register('ccmInterval', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {CCM_INTERVAL_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <FieldError message={errors.ccmInterval?.message} />
            </div>

            <div>
              <label htmlFor="y1731-priority" className="flex items-center gap-tight label">
                {t('testConfig.y1731.oam.priority')}
                <HelpIcon tooltip={t('testConfig.y1731.oam.priorityHelp')} />
              </label>
              <input
                id="y1731-priority"
                type="number"
                step={1}
                {...register('priority', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.priority?.message} />
            </div>
          </div>

          <div className="flex items-center gap-compact">
            <input
              id="y1731-tagged"
              type="checkbox"
              {...register('priorityTagged')}
              aria-label={t('testConfig.y1731.oam.taggedAria')}
              className="rounded border-surface-border"
            />
            <label
              htmlFor="y1731-tagged"
              title={t('testConfig.y1731.oam.taggedTitle')}
              className="text-sm text-text-primary"
            >
              {t('testConfig.y1731.oam.tagged')}
            </label>
          </div>
        </FormSection>

        <FormSection title={t('testConfig.y1731.measurement.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="y1731-framesize" className="flex items-center gap-tight label">
                {t('testConfig.y1731.measurement.frameSize')}
                <HelpIcon tooltip={t('testConfig.y1731.measurement.frameSizeHelp')} />
              </label>
              <select
                id="y1731-framesize"
                {...register('frameSize', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {FRAME_SIZE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {t(`testConfig.common.${option.qualifier}`, { size: option.value })}
                  </option>
                ))}
              </select>
              <FieldError message={errors.frameSize?.message} />
            </div>

            <div>
              <label htmlFor="y1731-interval" className="flex items-center gap-tight label">
                {t('testConfig.y1731.measurement.interval')}
                <HelpIcon tooltip={t('testConfig.y1731.measurement.intervalHelp')} />
              </label>
              <input
                id="y1731-interval"
                type="number"
                step={10}
                {...register('intervalMs', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.intervalMs?.message} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="y1731-count" className="flex items-center gap-tight label">
                {t('testConfig.y1731.measurement.count')}
                <HelpIcon tooltip={t('testConfig.y1731.measurement.countHelp')} />
              </label>
              <input
                id="y1731-count"
                type="number"
                step={1}
                {...register('count', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.count?.message} />
            </div>

            <div>
              <label htmlFor="y1731-duration" className="flex items-center gap-tight label">
                {t('testConfig.y1731.measurement.duration')}
                <HelpIcon tooltip={t('testConfig.y1731.measurement.durationHelp')} />
              </label>
              <input
                id="y1731-duration"
                type="number"
                step={1}
                {...register('duration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.duration?.message} />
            </div>
          </div>
        </FormSection>

        <TestSummary>
          <div>
            {t('testConfig.common.selectedTests')}:{' '}
            {[
              hasDelay && t('testConfig.y1731.tests.delay'),
              hasLoss && t('testConfig.y1731.tests.loss'),
              hasSLM && t('testConfig.y1731.tests.slm'),
              hasLoopback && t('testConfig.y1731.tests.loopback'),
            ]
              .filter(Boolean)
              .join(', ')}
          </div>
          <div>
            {t('testConfig.y1731.summary.mep', { mep: mepId, level: megLevel, meg: megId })}
          </div>
          <div>
            {t(
              priorityTagged
                ? 'testConfig.y1731.summary.ccmTagged'
                : 'testConfig.y1731.summary.ccm',
              {
                ccm:
                  CCM_INTERVAL_OPTIONS.find((option) => option.value === ccmInterval)?.label ??
                  `${ccmInterval}ms`,
                priority,
              },
            )}
          </div>
          <div>
            {t('testConfig.y1731.summary.frame', {
              size: frameSize,
              interval: intervalMs,
              count,
            })}
          </div>
          <div>{t('testConfig.y1731.summary.duration', { duration })}</div>
        </TestSummary>
      </div>
    </CollapsibleSection>
  );
}

export default Y1731ConfigForm;

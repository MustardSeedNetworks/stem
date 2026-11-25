/**
 * @fileoverview Stem - Y.1731 OAM Test Configuration
 * @description Migrated to react-hook-form + valibot per #325.
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { FrameSizeOption } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { hasGroupTests } from '../lib/testGroups';
import { Y1731ConfigSchema } from '../schemas/configs';
import { FormField } from '../shared/FormField';
import { FormSection } from '../shared/FormSection';
import { Input } from '../shared/Input';
import { CheckboxField, NumberSelectField } from './ConfigFormFields';
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
  frameSize: 128,
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
 * 1280, no jumbo, and no 64 — the extended payload does not fit one
 * (MIN_EXTENDED_FRAME_SIZE). Kept local rather than filtered out of the shared
 * list, so the set is stated rather than inferred from an exclusion.
 */
export const Y1731_FRAME_SIZE_OPTIONS: FrameSizeOption[] = [
  { value: 128, qualifier: 'frameSizeMin' },
  { value: 256, qualifier: 'frameSize' },
  { value: 512, qualifier: 'frameSize' },
  { value: 1024, qualifier: 'frameSize' },
  { value: 1518, qualifier: 'frameSizeMax' },
];

interface Y1731ConfigFormProps {
  config: Y1731Config;
  setConfig: (config: Y1731Config) => void;
  selectedTests: string[];
  /** Overrides the root test id: the drawer and the page can both render this form at once. */
  testId?: string;
}

export function Y1731ConfigForm({
  config,
  setConfig,
  selectedTests,
  testId = 'y1731-config-form',
}: Y1731ConfigFormProps): ReactElement | null {
  const hasY1731Tests = hasGroupTests('y1731', selectedTests);

  const form = useConfigForm<Y1731Config>({
    schema: Y1731ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    control,
    watch,
    formState: { errors },
  } = form;
  const { t } = useTranslation(['settings', 'help']);

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
    <div data-testid={testId}>
      <div className="stack-lg">
        <FormSection title={t('testConfig.y1731.mep.title')}>
          <div className="grid grid-cols-3 gap-default">
            <FormField
              label={t('testConfig.y1731.mep.mepId')}
              hint={t('help:tooltips.testConfig.y1731.mep.mepIdHelp')}
              error={errors.mepId?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('mepId', { valueAsNumber: true })}
                />
              )}
            </FormField>

            <FormField
              label={t('testConfig.y1731.mep.megLevel')}
              hint={t('help:tooltips.testConfig.y1731.mep.megLevelHelp')}
              error={errors.megLevel?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('megLevel', { valueAsNumber: true })}
                />
              )}
            </FormField>

            <FormField
              label={t('testConfig.y1731.mep.megId')}
              hint={t('help:tooltips.testConfig.y1731.mep.megIdHelp')}
              error={errors.megId?.message}
            >
              {(control) => (
                <Input {...control} type="text" maxLength={45} {...register('megId')} />
              )}
            </FormField>
          </div>
        </FormSection>

        <FormSection title={t('testConfig.y1731.oam.title')}>
          <div className="grid grid-cols-2 gap-default">
            <NumberSelectField
              control={control}
              name="ccmInterval"
              label={t('testConfig.y1731.oam.ccm')}
              hint={t('help:tooltips.testConfig.y1731.oam.ccmHelp')}
              error={errors.ccmInterval?.message}
              options={CCM_INTERVAL_OPTIONS}
            />

            <FormField
              label={t('testConfig.y1731.oam.priority')}
              hint={t('help:tooltips.testConfig.y1731.oam.priorityHelp')}
              error={errors.priority?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('priority', { valueAsNumber: true })}
                />
              )}
            </FormField>
          </div>

          <CheckboxField
            control={control}
            name="priorityTagged"
            label={t('testConfig.y1731.oam.tagged')}
            help={t('help:tooltips.testConfig.y1731.oam.taggedTitle')}
          />
        </FormSection>

        <FormSection title={t('testConfig.y1731.measurement.title')}>
          <div className="grid grid-cols-2 gap-default">
            <NumberSelectField
              control={control}
              name="frameSize"
              label={t('testConfig.y1731.measurement.frameSize')}
              hint={t('help:tooltips.testConfig.y1731.measurement.frameSizeHelp')}
              error={errors.frameSize?.message}
              options={Y1731_FRAME_SIZE_OPTIONS.map((option) => ({
                value: option.value,
                label: t(`testConfig.common.${option.qualifier}`, { size: option.value }),
              }))}
            />

            <FormField
              label={t('testConfig.y1731.measurement.interval')}
              hint={t('help:tooltips.testConfig.y1731.measurement.intervalHelp')}
              error={errors.intervalMs?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={10}
                  {...register('intervalMs', { valueAsNumber: true })}
                />
              )}
            </FormField>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <FormField
              label={t('testConfig.y1731.measurement.count')}
              hint={t('help:tooltips.testConfig.y1731.measurement.countHelp')}
              error={errors.count?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('count', { valueAsNumber: true })}
                />
              )}
            </FormField>

            <FormField
              label={t('testConfig.y1731.measurement.duration')}
              hint={t('help:tooltips.testConfig.y1731.measurement.durationHelp')}
              error={errors.duration?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('duration', { valueAsNumber: true })}
                />
              )}
            </FormField>
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
    </div>
  );
}

export default Y1731ConfigForm;

/**
 * @fileoverview The Stem - Y.1564 Service Activation Test Configuration
 * @description Advanced configuration form for ITU-T Y.1564 / MEF Service Activation Testing.
 *              Allows users to configure service parameters including CIR, EIR, CBS, EBS,
 *              frame sizes, test duration, and VLAN settings.
 *
 * Forms-stack pilot (#325): this form is the first to migrate to
 * react-hook-form + valibot. The schema lives at `src/schemas/configs.ts`
 * and is plumbed in via the `useConfigForm` helper. Field-level errors
 * render inline below each input; the cross-field rule (FDV ≤ FD) is
 * shown at the form footer. The 6 remaining ConfigForms follow the
 * same pattern — see issue #325 for the sweep.
 */

import { AlertTriangle, Info, Settings2 } from 'lucide-react';
import type { ReactElement } from 'react';
import { useConfigForm } from '../forms/useConfigForm';
import { Y1564ConfigSchema } from '../schemas/configs';
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

/** Field-level error display. Keeps JSX terse and consistent. */
function FieldError({ message }: { message?: string }): ReactElement | null {
  if (!message) return null;
  return (
    <div className="mt-1 text-xs text-[var(--color-status-danger)] flex items-center gap-1">
      <AlertTriangle className="w-3 h-3" />
      {message}
    </div>
  );
}

export function Y1564ConfigForm({
  config,
  setConfig,
  selectedTests,
}: Y1564ConfigFormProps): ReactElement | null {
  const hasY1564Tests = selectedTests.some((t) => t.startsWith('y1564') || t.startsWith('mef'));

  const form = useConfigForm<Y1564Config>({
    schema: Y1564ConfigSchema,
    config,
    setConfig,
  });

  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;

  if (!hasY1564Tests) {
    return null;
  }

  // Watched values for derived displays (frame-size checkboxes,
  // summary panel, VLAN PCP conditional render). react-hook-form's
  // watch keeps these in sync with the form's internal state.
  const frameSizes = watch('frameSizes') ?? [];
  const vlanId = watch('vlanId') ?? 0;
  const cir = watch('cir') ?? 0;
  const eir = watch('eir') ?? 0;
  const flrThreshold = watch('flrThreshold') ?? 0;
  const fdThreshold = watch('fdThreshold') ?? 0;
  const fdvThreshold = watch('fdvThreshold') ?? 0;
  const pcp = watch('pcp') ?? 0;
  const configStepDuration = watch('configStepDuration') ?? 0;
  const perfTestDuration = watch('perfTestDuration') ?? 0;

  const toggleFrameSize = (size: number): void => {
    if (frameSizes.includes(size)) {
      setValue(
        'frameSizes',
        frameSizes.filter((s) => s !== size),
        { shouldValidate: true, shouldDirty: true },
      );
    } else {
      setValue(
        'frameSizes',
        [...frameSizes, size].sort((a, b) => a - b),
        {
          shouldValidate: true,
          shouldDirty: true,
        },
      );
    }
  };

  const isConfigTest = selectedTests.some((t) => t.includes('config'));
  const isPerfTest = selectedTests.some((t) => t.includes('perf'));
  const isFullTest = selectedTests.some((t) => t.includes('full'));

  // Cross-field error (fdv > fd). valibot's v.check() surfaces under
  // formState.errors.root.<unique-key>; we render the first one found.
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
              step={1}
              {...register('cir', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.cir?.message} />
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
              step={1}
              {...register('eir', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.eir?.message} />
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
              step={1}
              {...register('cbs', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.cbs?.message} />
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
              step={1}
              {...register('ebs', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.ebs?.message} />
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
              step={0.001}
              {...register('flrThreshold', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.flrThreshold?.message} />
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
              step={1}
              {...register('fdThreshold', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.fdThreshold?.message} />
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
              step={1}
              {...register('fdvThreshold', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.fdvThreshold?.message} />
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
                title={`Include ${option.value}-byte frames in the Y.1564 service activation sweep`}
                className="flex items-center gap-2 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)] text-sm"
              >
                <input
                  type="checkbox"
                  checked={frameSizes.includes(option.value)}
                  onChange={() => toggleFrameSize(option.value)}
                  aria-label={`Test ${option.value}-byte frames`}
                  className="w-4 h-4 accent-[var(--color-brand-primary)]"
                />
                <span className="text-[var(--color-text-primary)]">{option.label}</span>
              </label>
            ))}
          </div>
          <FieldError message={errors.frameSizes?.message} />
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
                step={1}
                {...register('configStepDuration', { valueAsNumber: true })}
                className="mt-1 w-full"
              />
              <FieldError message={errors.configStepDuration?.message} />
              <div className="text-xs text-[var(--color-text-muted)] mt-1">
                Total config test: ~{configStepDuration * 4 * frameSizes.length}s
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
                step={60}
                {...register('perfTestDuration', { valueAsNumber: true })}
                className="mt-1 w-full"
              />
              <FieldError message={errors.perfTestDuration?.message} />
              <div className="text-xs text-[var(--color-text-muted)] mt-1">
                = {Math.floor(perfTestDuration / 60)} minutes
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
              step={1}
              {...register('vlanId', { valueAsNumber: true })}
              className="mt-1 w-full"
            />
            <FieldError message={errors.vlanId?.message} />
            {vlanId === 0 && (
              <div className="text-xs text-[var(--color-text-muted)] mt-1">Untagged traffic</div>
            )}
          </div>

          {vlanId > 0 && (
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
                {...register('pcp', { valueAsNumber: true })}
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
              <FieldError message={errors.pcp?.message} />
            </div>
          )}

          {/* Color-Aware Mode */}
          <label
            title="Send traffic as green (in-profile, conforms to CIR) and yellow (out-of-profile, conforms to EIR) and verify each color is treated correctly"
            className="flex items-center gap-3 p-2 rounded-lg cursor-pointer hover:bg-[var(--color-surface-hover)]"
          >
            <input
              type="checkbox"
              {...register('colorAware')}
              aria-label="Enable color-aware Y.1564 testing"
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

        {/* Cross-field error footer */}
        {crossFieldError && (
          <div className="p-2 rounded-lg bg-[var(--color-status-danger-subtle)] text-[var(--color-status-danger)] text-sm flex items-center gap-2">
            <AlertTriangle className="w-4 h-4" />
            {crossFieldError.message}
          </div>
        )}

        {/* Summary */}
        <div className="p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
          <div className="flex items-center gap-2 text-sm font-medium text-[var(--color-text-primary)] mb-2">
            <Info className="w-4 h-4" />
            Test Summary
          </div>
          <div className="text-xs text-[var(--color-text-muted)] space-y-1">
            <div>
              Service: {cir} Mbps CIR
              {eir > 0 && ` + ${eir} Mbps EIR`}
            </div>
            <div>Frame sizes: {frameSizes.join(', ')} bytes</div>
            <div>
              SLA: FLR≤{flrThreshold}%, FD≤{fdThreshold}ms, FDV≤{fdvThreshold}
              ms
            </div>
            {vlanId > 0 && (
              <div>
                VLAN {vlanId} / PCP {pcp}
              </div>
            )}
          </div>
        </div>
      </div>
    </CollapsibleSection>
  );
}

export default Y1564ConfigForm;

/**
 * @fileoverview The Stem - Module Card Component
 * @description Card component for each test module (Benchmark, ServiceTest, etc.)
 *              with enable/disable toggles, autostart options, and test execution.
 */

import {
  Check,
  ChevronDown,
  ChevronUp,
  Clock,
  Play,
  Power,
  RefreshCw,
  Settings2,
  Square,
  XCircle,
} from 'lucide-react';
import type { ReactElement } from 'react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { cn, icon as iconTokens, spacing, status as statusColor } from '../styles/theme';

export interface ModuleTest {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
}

export interface ModuleConfig {
  name: string;
  displayName: string;
  description: string;
  color: string;
  standard: string;
  enabled: boolean;
  autoStart: boolean;
  tests: ModuleTest[];
}

export interface ModuleStatus {
  status: 'idle' | 'starting' | 'running' | 'completed' | 'error' | 'cancelled';
  currentTest: string | null;
  progress?: number;
  message?: string;
}

/** Per-frame-size result for RFC 2544 style tests */
export interface FrameSizeResult {
  frameSize: number;
  status: 'pending' | 'running' | 'completed' | 'error';
  progress?: number; // 0-100 for running state
  txPackets?: number;
  rxPackets?: number;
  txBytes?: number;
  rxBytes?: number;
  lossPercent?: number;
  throughputPps?: number;
  throughputMbps?: number;
  latencyUs?: number; // microseconds
  jitterUs?: number;
}

/** Service flow result for Y.1564 style tests */
export interface ServiceFlowResult {
  flowId: string;
  flowName: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  cir?: number; // Committed Information Rate
  cirAchieved?: number;
  eir?: number; // Excess Information Rate
  eirAchieved?: number;
  frameDelay?: number;
  frameDelayVariation?: number;
  frameLoss?: number;
}

/** OAM measurement result for Y.1731 style tests */
export interface OamMeasurementResult {
  measurementType: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  delayMin?: number;
  delayAvg?: number;
  delayMax?: number;
  jitter?: number;
  lossNear?: number;
  lossFar?: number;
}

/** Combined test results that can hold different result types */
export interface ModuleTestResults {
  testType: string;
  startedAt?: string;
  completedAt?: string;
  duration?: number;
  success?: boolean;
  error?: string;
  // Different result types based on module
  frameSizeResults?: FrameSizeResult[];
  serviceFlowResults?: ServiceFlowResult[];
  oamResults?: OamMeasurementResult[];
}

interface ModuleCardProps {
  config: ModuleConfig;
  status: ModuleStatus;
  results?: ModuleTestResults | null;
  onToggleModule: (enabled: boolean) => void;
  onToggleAutoStart: (enabled: boolean) => void;
  onToggleTest: (testId: string, enabled: boolean) => void;
  onStart: () => void;
  onStop: () => void;
  onConfigure: () => void;
}

function formatNumber(num: number): string {
  if (num >= 1e9) {
    return `${(num / 1e9).toFixed(1)}G`;
  }
  if (num >= 1e6) {
    return `${(num / 1e6).toFixed(1)}M`;
  }
  if (num >= 1e3) {
    return `${(num / 1e3).toFixed(1)}K`;
  }
  return num.toString();
}

function formatRate(pps: number): string {
  if (pps >= 1e6) {
    return `${(pps / 1e6).toFixed(2)}Mpps`;
  }
  if (pps >= 1e3) {
    return `${(pps / 1e3).toFixed(1)}Kpps`;
  }
  return `${pps}pps`;
}

/** Get color class for loss percentage based on threshold */
function getLossColorClass(lossPercent: number, isPending: boolean): string {
  if (isPending) {
    return 'text-text-muted';
  }
  if (lossPercent === 0) {
    return statusColor.text.success;
  }
  if (lossPercent < 1) {
    return statusColor.text.warning;
  }
  return statusColor.text.error;
}

/** Render rate cell content based on result status */
function RateCellContent({ result }: { result: FrameSizeResult }): ReactElement {
  if (result.status === 'pending') {
    return <>—</>;
  }
  if (result.status === 'running') {
    return <span className="text-text-muted">measuring</span>;
  }
  return <>{formatRate(result.throughputPps ?? 0)}</>;
}

/** Renders the frame size results table for RFC 2544 style tests */
function FrameSizeResultsTable({
  results,
  color,
}: {
  results: FrameSizeResult[];
  color: string;
}): ReactElement {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-xs">
        <thead>
          <tr className="text-text-muted border-b border-surface-border">
            <th className="text-left py-row pr-2 font-medium">Frame</th>
            <th className="text-right py-row px-cell font-medium">TX</th>
            <th className="text-right py-row px-cell font-medium">RX</th>
            <th className="text-right py-row px-cell font-medium">Loss</th>
            <th className="text-right py-row px-cell font-medium">Rate</th>
            <th className="text-center py-row pl-2 font-medium w-8">Status</th>
          </tr>
        </thead>
        <tbody>
          {results.map((result) => (
            <tr
              key={result.frameSize}
              className={cn(
                'border-b border-surface-border/50',
                result.status === 'running' && statusColor.bg.successSubtle,
              )}
            >
              <td className="py-row pr-2 font-mono font-medium text-text-primary">
                {result.frameSize}B
              </td>
              <td className="py-row px-cell text-right font-mono text-text-secondary">
                {result.status === 'pending' ? '—' : formatNumber(result.txPackets ?? 0)}
              </td>
              <td className="py-row px-cell text-right font-mono text-text-secondary">
                {result.status === 'pending' ? '—' : formatNumber(result.rxPackets ?? 0)}
              </td>
              <td
                className={cn(
                  'py-row px-cell text-right font-mono',
                  getLossColorClass(result.lossPercent ?? 0, result.status === 'pending'),
                )}
              >
                {result.status === 'pending'
                  ? '\u2014'
                  : `${(result.lossPercent ?? 0).toFixed(2)}%`}
              </td>
              <td className="py-row px-cell text-right font-mono text-text-secondary">
                <RateCellContent result={result} />
              </td>
              <td className="py-row pl-2 text-center">
                {result.status === 'completed' && (
                  <Check className={cn('w-4 h-4 inline', statusColor.text.success)} />
                )}
                {result.status === 'running' && (
                  <div
                    className="w-4 h-4 rounded-full border-2 border-t-transparent animate-spin inline-block"
                    style={{
                      borderColor: color,
                      borderTopColor: 'transparent',
                    }}
                  />
                )}
                {result.status === 'error' && (
                  <XCircle className={cn('w-4 h-4 inline', statusColor.text.error)} />
                )}
                {result.status === 'pending' && (
                  <Clock className="w-4 h-4 text-text-muted inline" />
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Renders service flow results for Y.1564 style tests */
function ServiceFlowResultsTable({ results }: { results: ServiceFlowResult[] }): ReactElement {
  return (
    <div className="stack-sm">
      {results.map((flow) => (
        <div
          key={flow.flowId}
          className={cn(
            'pad-xs rounded-lg border border-surface-border',
            flow.status === 'running' && statusColor.bg.successSubtle,
          )}
        >
          <div className="flex-between mb-tight">
            <span className="label">{flow.flowName}</span>
            <span
              className={cn(
                'text-xs px-cell py-0.5 rounded-full',
                flow.status === 'completed' && statusColor.badge.success,
                flow.status === 'running' && statusColor.badge.info,
                flow.status === 'pending' && 'bg-surface-base text-text-muted',
                flow.status === 'error' && statusColor.badge.error,
              )}
            >
              {flow.status}
            </span>
          </div>
          {flow.status !== 'pending' && (
            <div className="grid grid-cols-4 gap-compact text-xs">
              <div>
                <div className="text-text-muted">CIR</div>
                <div className="font-mono">
                  {flow.cirAchieved ?? '—'}/{flow.cir ?? '—'}
                </div>
              </div>
              <div>
                <div className="text-text-muted">Delay</div>
                <div className="font-mono">
                  {flow.frameDelay !== null ? `${flow.frameDelay}ms` : '—'}
                </div>
              </div>
              <div>
                <div className="text-text-muted">Jitter</div>
                <div className="font-mono">
                  {flow.frameDelayVariation !== null ? `${flow.frameDelayVariation}ms` : '—'}
                </div>
              </div>
              <div>
                <div className="text-text-muted">Loss</div>
                <div className="font-mono">
                  {flow.frameLoss !== null ? `${flow.frameLoss}%` : '—'}
                </div>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

/** Renders OAM measurement results for Y.1731 style tests */
function OamResultsTable({ results }: { results: OamMeasurementResult[] }): ReactElement {
  return (
    <div className="stack-sm">
      {results.map((measurement) => (
        <div
          key={measurement.measurementType}
          className={cn(
            'pad-xs rounded-lg border border-surface-border',
            measurement.status === 'running' && statusColor.bg.successSubtle,
          )}
        >
          <div className="flex-between mb-tight">
            <span className="label">{measurement.measurementType}</span>
            <span
              className={cn(
                'text-xs px-cell py-0.5 rounded-full',
                measurement.status === 'completed' && statusColor.badge.success,
                measurement.status === 'running' && statusColor.badge.info,
                measurement.status === 'pending' && 'bg-surface-base text-text-muted',
                measurement.status === 'error' && statusColor.badge.error,
              )}
            >
              {measurement.status}
            </span>
          </div>
          {measurement.status !== 'pending' && (
            <div className="grid grid-cols-3 gap-compact text-xs">
              <div>
                <div className="text-text-muted">Delay (min/avg/max)</div>
                <div className="font-mono">
                  {measurement.delayMin ?? '—'}/{measurement.delayAvg ?? '—'}/
                  {measurement.delayMax ?? '—'}μs
                </div>
              </div>
              <div>
                <div className="text-text-muted">Jitter</div>
                <div className="font-mono">
                  {measurement.jitter !== null ? `${measurement.jitter}μs` : '—'}
                </div>
              </div>
              <div>
                <div className="text-text-muted">Loss (near/far)</div>
                <div className="font-mono">
                  {measurement.lossNear ?? '—'}%/{measurement.lossFar ?? '—'}%
                </div>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

/** Status indicator badge for module card */
function ModuleStatusIndicator({
  status,
  isRunning,
}: {
  status: ModuleStatus;
  isRunning: boolean;
}): ReactElement | null {
  if (isRunning) {
    return (
      <div
        className={cn(
          'flex items-center gap-compact px-3 py-compact-md rounded-full',
          statusColor.bg.successSoft,
        )}
      >
        <span className={cn('w-2 h-2 rounded-full animate-pulse', statusColor.bg.success)} />
        <span className={cn('text-xs font-medium', statusColor.text.success)}>
          {status.status === 'starting' ? 'Starting...' : status.currentTest || 'Running'}
        </span>
      </div>
    );
  }
  if (status.status === 'completed') {
    return (
      <div
        className={cn(
          'flex items-center gap-compact px-3 py-compact-md rounded-full',
          statusColor.bg.infoSoft,
        )}
      >
        <span className={cn('text-xs font-medium', statusColor.text.info)}>Completed</span>
      </div>
    );
  }
  if (status.status === 'error') {
    return (
      <div
        className={cn(
          'flex items-center gap-compact px-3 py-compact-md rounded-full',
          statusColor.bg.errorSoft,
        )}
      >
        <span className={cn('text-xs font-medium', statusColor.text.error)}>
          Error{status.message ? `: ${status.message}` : ''}
        </span>
      </div>
    );
  }
  return null;
}

/** Start/Stop button for module card */
function ModuleActionButton({
  config,
  isRunning,
  enabledTestCount,
  onStart,
  onStop,
}: {
  config: ModuleConfig;
  isRunning: boolean;
  enabledTestCount: number;
  onStart: () => void;
  onStop: () => void;
}): ReactElement | null {
  const { t } = useTranslation('modules');
  if (!config.enabled) {
    return null;
  }
  if (isRunning) {
    return (
      <button
        type="button"
        onClick={onStop}
        title={t('card.stop.title', { name: config.displayName })}
        aria-label={t('card.stop.ariaLabel', { name: config.displayName })}
        className={cn(
          'px-4 py-row rounded-lg flex items-center gap-compact transition-colors',
          statusColor.badge.error,
          statusColor.hover.errorStrong,
        )}
      >
        <Square className="w-4 h-4" />
        <span className="text-sm font-medium">{t('card.action.stopLabel')}</span>
      </button>
    );
  }
  return (
    <button
      type="button"
      onClick={onStart}
      disabled={enabledTestCount === 0}
      title={
        enabledTestCount === 0
          ? t('card.start.titleEmpty', { name: config.displayName })
          : t('card.start.titleEnabled', {
              count: enabledTestCount,
              name: config.displayName,
            })
      }
      aria-label={t('card.start.ariaLabel', { name: config.displayName })}
      className={cn(
        'px-4 py-row rounded-lg flex items-center gap-compact transition-colors',
        enabledTestCount > 0
          ? 'bg-brand-primary text-white hover:bg-brand-primary'
          : 'bg-surface-base text-text-muted cursor-not-allowed',
      )}
    >
      <Play className="w-4 h-4" />
      <span className="text-sm font-medium">{t('card.action.startLabel')}</span>
    </button>
  );
}

/** Results section for module card */
function ModuleResultsSection({
  results,
  config,
}: {
  results: ModuleTestResults | null | undefined;
  config: ModuleConfig;
}): ReactElement {
  const { t } = useTranslation('modules');
  return (
    <div className="border-t border-surface-border bg-surface-base/50">
      <div className={spacing.pad.sm}>
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-2">
          {results?.testType
            ? `${results.testType} ${t('card.section.results')}`
            : t('card.section.results')}
        </div>

        {/* Frame Size Results (RFC 2544 style) */}
        {results?.frameSizeResults && results.frameSizeResults.length > 0 && (
          <FrameSizeResultsTable results={results.frameSizeResults} color={config.color} />
        )}

        {/* Service Flow Results (Y.1564 style) */}
        {results?.serviceFlowResults && results.serviceFlowResults.length > 0 && (
          <ServiceFlowResultsTable results={results.serviceFlowResults} />
        )}

        {/* OAM Results (Y.1731 style) */}
        {results?.oamResults && results.oamResults.length > 0 && (
          <OamResultsTable results={results.oamResults} />
        )}

        {/* Error message */}
        {results?.error ? (
          <div
            className={cn(
              'mt-inline pad-xs rounded-lg border',
              statusColor.bg.errorSoft,
              statusColor.border.errorSoft,
            )}
          >
            <span className={cn('text-xs', statusColor.text.error)}>{results.error}</span>
          </div>
        ) : null}

        {/* Duration */}
        {results?.duration !== undefined ? (
          <div className="mt-inline text-xs text-text-muted">
            Duration: {(results.duration / 1000).toFixed(1)}s
          </div>
        ) : null}
      </div>
    </div>
  );
}

/** Expanded test list section for module card */
function ModuleExpandedContent({
  config,
  status,
  isRunning,
  onToggleAutoStart,
  onToggleTest,
}: {
  config: ModuleConfig;
  status: ModuleStatus;
  isRunning: boolean;
  onToggleAutoStart: (enabled: boolean) => void;
  onToggleTest: (testId: string, enabled: boolean) => void;
}): ReactElement {
  const { t } = useTranslation('modules');
  return (
    <div className="border-t border-surface-border">
      {/* Auto-start Toggle */}
      <div className={cn(spacing.pad.sm, 'flex-between bg-surface-base')}>
        <div className="flex items-center gap-compact">
          <RefreshCw className={cn(iconTokens.size.sm, 'text-text-muted')} />
          <span className="text-sm text-text-secondary">{t('card.section.autoStartLabel')}</span>
        </div>
        <button
          type="button"
          onClick={(): void => onToggleAutoStart(!config.autoStart)}
          title={
            config.autoStart ? t('card.autoStart.titleEnabled') : t('card.autoStart.titleDisabled')
          }
          aria-label={
            config.autoStart
              ? t('card.autoStart.ariaLabelEnabled')
              : t('card.autoStart.ariaLabelDisabled')
          }
          className={cn(
            'w-10 h-6 rounded-full relative transition-colors',
            config.autoStart ? 'bg-brand-primary' : 'bg-surface-border',
          )}
        >
          <span
            className={cn(
              'absolute top-1 w-4 h-4 rounded-full bg-white transition-transform',
              config.autoStart ? 'translate-x-5' : 'translate-x-1',
            )}
          />
        </button>
      </div>

      {/* Test List */}
      <div className={spacing.pad.sm}>
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-2">
          {t('card.section.tests')}
        </div>
        <div className="stack-xs">
          {config.tests.map((test) => (
            <label
              key={test.id}
              title={t('card.test.title', { description: test.description, name: test.name })}
              className={cn(
                'flex items-center gap-default pad-xs rounded-lg cursor-pointer transition-colors',
                'hover:bg-surface-hover',
                test.enabled ? '' : 'opacity-60',
              )}
            >
              <input
                type="checkbox"
                checked={test.enabled}
                onChange={(e: React.ChangeEvent<HTMLInputElement>): void =>
                  onToggleTest(test.id, e.target.checked)
                }
                aria-label={t('card.test.ariaLabel', { name: test.name })}
                className="w-4 h-4"
                style={{ accentColor: config.color }}
              />
              <div className="flex-1 min-w-0">
                <div className="label">{test.name}</div>
                <div className="text-xs text-text-muted truncate">{test.description}</div>
              </div>
              {isRunning && status.currentTest === test.id && (
                <span className={cn(statusColor.dot, statusColor.bg.success, 'animate-pulse')} />
              )}
            </label>
          ))}
        </div>
      </div>
    </div>
  );
}

export function ModuleCard({
  config,
  status,
  results,
  onToggleModule,
  onToggleAutoStart,
  onToggleTest,
  onStart,
  onStop,
  onConfigure,
}: ModuleCardProps): ReactElement {
  const { t } = useTranslation('modules');
  const [expanded, setExpanded] = useState(false);
  const enabledTestCount = config.tests.filter((t2) => t2.enabled).length;
  const isRunning = status.status === 'running' || status.status === 'starting';
  const hasResults = checkHasResults(results);
  const showResults = isRunning || status.status === 'completed' || status.status === 'error';

  return (
    <div
      className={cn(
        'border rounded-xl overflow-hidden transition-all',
        config.enabled ? 'border-surface-border' : 'border-transparent opacity-60',
        'bg-surface-raised',
      )}
      style={{
        borderLeftWidth: '4px',
        borderLeftColor: config.enabled ? config.color : 'transparent',
      }}
    >
      {/* Module Header */}
      <div className={cn(spacing.pad.default, 'flex-between')}>
        <div className="flex items-center gap-default flex-1">
          {/* Enable Toggle */}
          <button
            type="button"
            onClick={(): void => onToggleModule(!config.enabled)}
            className={cn(
              'w-8 h-8 rounded-lg flex-center transition-colors',
              config.enabled ? statusColor.badge.successStrong : 'bg-surface-base text-text-muted',
            )}
            title={
              config.enabled
                ? t('card.module.titleEnabled', {
                    name: config.displayName,
                    standard: config.standard,
                  })
                : t('card.module.titleDisabled', {
                    name: config.displayName,
                    standard: config.standard,
                  })
            }
            aria-label={
              config.enabled
                ? t('card.module.ariaLabelEnabled', { name: config.displayName })
                : t('card.module.ariaLabelDisabled', { name: config.displayName })
            }
          >
            <Power className="w-4 h-4" />
          </button>

          {/* Module Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-compact">
              <span
                className="w-3 h-3 rounded-full flex-shrink-0"
                style={{ backgroundColor: config.color }}
              />
              <h3 className="font-semibold text-text-primary truncate">{config.displayName}</h3>
              <span className="text-xs px-cell py-0.5 rounded-full bg-surface-base text-text-muted">
                {config.standard}
              </span>
            </div>
            <p className="text-xs text-text-muted mt-0.5 truncate">
              {t('card.module.testsEnabledCount', {
                enabled: enabledTestCount,
                total: config.tests.length,
              })}
            </p>
          </div>

          {/* Status Indicator */}
          <ModuleStatusIndicator status={status} isRunning={isRunning} />
        </div>

        {/* Actions */}
        <div className="flex items-center gap-compact">
          {/* Configure Button */}
          <button
            type="button"
            onClick={onConfigure}
            className={cn(
              'pad-xs rounded-lg transition-colors',
              'text-text-muted hover:text-text-primary',
              'hover:bg-surface-hover',
            )}
            title={t('card.configure.title', { name: config.displayName })}
            aria-label={t('card.configure.ariaLabel', { name: config.displayName })}
          >
            <Settings2 className="w-4 h-4" />
          </button>

          {/* Start/Stop Button */}
          <ModuleActionButton
            config={config}
            isRunning={isRunning}
            enabledTestCount={enabledTestCount}
            onStart={onStart}
            onStop={onStop}
          />

          {/* Expand Toggle */}
          <button
            type="button"
            onClick={(): void => setExpanded(!expanded)}
            className={cn(
              'pad-xs rounded-lg transition-colors',
              'text-text-muted hover:text-text-primary',
              'hover:bg-surface-hover',
            )}
            title={
              expanded
                ? t('card.expand.titleExpanded', { name: config.displayName })
                : t('card.expand.titleCollapsed', { name: config.displayName })
            }
            aria-label={
              expanded ? t('card.expand.ariaLabelExpanded') : t('card.expand.ariaLabelCollapsed')
            }
          >
            {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>
      </div>

      {/* Test Results Section - Always visible when running or has results */}
      {config.enabled && showResults && hasResults ? (
        <ModuleResultsSection results={results} config={config} />
      ) : null}

      {/* Expanded Content - Settings and Test Selection */}
      {expanded && config.enabled ? (
        <ModuleExpandedContent
          config={config}
          status={status}
          isRunning={isRunning}
          onToggleAutoStart={onToggleAutoStart}
          onToggleTest={onToggleTest}
        />
      ) : null}
    </div>
  );
}

/** Helper to check if results have data */
function checkHasResults(results: ModuleTestResults | null | undefined): boolean {
  if (!results) {
    return false;
  }
  const hasFrameSizeResults =
    results.frameSizeResults !== null &&
    results.frameSizeResults !== undefined &&
    results.frameSizeResults.length > 0;
  const hasServiceFlowResults =
    results.serviceFlowResults !== null &&
    results.serviceFlowResults !== undefined &&
    results.serviceFlowResults.length > 0;
  const hasOamResults =
    results.oamResults !== null &&
    results.oamResults !== undefined &&
    results.oamResults.length > 0;
  return hasFrameSizeResults || hasServiceFlowResults || hasOamResults;
}

export default ModuleCard;

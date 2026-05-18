/**
 * @fileoverview The Stem - Module Card Component
 * @description Card component for each test module (Benchmark, ServiceTest, etc.)
 *              with enable/disable toggles, autostart options, and test execution.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
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
            <th className="text-left py-2 pr-2 font-medium">Frame</th>
            <th className="text-right py-2 px-2 font-medium">TX</th>
            <th className="text-right py-2 px-2 font-medium">RX</th>
            <th className="text-right py-2 px-2 font-medium">Loss</th>
            <th className="text-right py-2 px-2 font-medium">Rate</th>
            <th className="text-center py-2 pl-2 font-medium w-8">Status</th>
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
              <td className="py-2 pr-2 font-mono font-medium text-text-primary">
                {result.frameSize}B
              </td>
              <td className="py-2 px-2 text-right font-mono text-text-secondary">
                {result.status === 'pending' ? '—' : formatNumber(result.txPackets ?? 0)}
              </td>
              <td className="py-2 px-2 text-right font-mono text-text-secondary">
                {result.status === 'pending' ? '—' : formatNumber(result.rxPackets ?? 0)}
              </td>
              <td
                className={cn(
                  'py-2 px-2 text-right font-mono',
                  getLossColorClass(result.lossPercent ?? 0, result.status === 'pending'),
                )}
              >
                {result.status === 'pending'
                  ? '\u2014'
                  : `${(result.lossPercent ?? 0).toFixed(2)}%`}
              </td>
              <td className="py-2 px-2 text-right font-mono text-text-secondary">
                <RateCellContent result={result} />
              </td>
              <td className="py-2 pl-2 text-center">
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
    <div className="space-y-2">
      {results.map((flow) => (
        <div
          key={flow.flowId}
          className={cn(
            'p-2 rounded-lg border border-surface-border',
            flow.status === 'running' && statusColor.bg.successSubtle,
          )}
        >
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-text-primary">{flow.flowName}</span>
            <span
              className={cn(
                'text-xs px-2 py-0.5 rounded-full',
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
            <div className="grid grid-cols-4 gap-2 text-xs">
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
    <div className="space-y-2">
      {results.map((measurement) => (
        <div
          key={measurement.measurementType}
          className={cn(
            'p-2 rounded-lg border border-surface-border',
            measurement.status === 'running' && statusColor.bg.successSubtle,
          )}
        >
          <div className="flex items-center justify-between mb-1">
            <span className="text-sm font-medium text-text-primary">
              {measurement.measurementType}
            </span>
            <span
              className={cn(
                'text-xs px-2 py-0.5 rounded-full',
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
            <div className="grid grid-cols-3 gap-2 text-xs">
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
          'flex items-center gap-2 px-3 py-1.5 rounded-full',
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
        className={cn('flex items-center gap-2 px-3 py-1.5 rounded-full', statusColor.bg.infoSoft)}
      >
        <span className={cn('text-xs font-medium', statusColor.text.info)}>Completed</span>
      </div>
    );
  }
  if (status.status === 'error') {
    return (
      <div
        className={cn('flex items-center gap-2 px-3 py-1.5 rounded-full', statusColor.bg.errorSoft)}
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
  if (!config.enabled) {
    return null;
  }
  if (isRunning) {
    return (
      <button
        type="button"
        onClick={onStop}
        title={`Stop the running ${config.displayName} test and discard incomplete results`}
        aria-label={`Stop ${config.displayName} test`}
        className={cn(
          'px-4 py-2 rounded-lg flex items-center gap-2 transition-colors',
          statusColor.badge.error,
          statusColor.hover.errorStrong,
        )}
      >
        <Square className="w-4 h-4" />
        <span className="text-sm font-medium">Stop</span>
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
          ? `Enable at least one ${config.displayName} test below to start`
          : `Start the ${enabledTestCount} enabled ${config.displayName} test${enabledTestCount === 1 ? '' : 's'} using the current configuration`
      }
      aria-label={`Start ${config.displayName} tests`}
      className={cn(
        'px-4 py-2 rounded-lg flex items-center gap-2 transition-colors',
        enabledTestCount > 0
          ? 'bg-brand-primary text-white hover:bg-brand-primary'
          : 'bg-surface-base text-text-muted cursor-not-allowed',
      )}
    >
      <Play className="w-4 h-4" />
      <span className="text-sm font-medium">Start</span>
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
  return (
    <div className="border-t border-surface-border bg-surface-base/50">
      <div className={spacing.pad.sm}>
        <div className="text-xs font-semibold text-text-muted uppercase tracking-wide mb-2">
          {results?.testType || 'Test'} Results
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
              'mt-2 p-2 rounded-lg border',
              statusColor.bg.errorSoft,
              statusColor.border.errorSoft,
            )}
          >
            <span className={cn('text-xs', statusColor.text.error)}>{results.error}</span>
          </div>
        ) : null}

        {/* Duration */}
        {results?.duration !== undefined ? (
          <div className="mt-2 text-xs text-text-muted">
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
  return (
    <div className="border-t border-surface-border">
      {/* Auto-start Toggle */}
      <div className={cn(spacing.pad.sm, 'flex items-center justify-between bg-surface-base')}>
        <div className="flex items-center gap-2">
          <RefreshCw className={cn(iconTokens.size.sm, 'text-text-muted')} />
          <span className="text-sm text-text-secondary">Auto-start on link</span>
        </div>
        <button
          type="button"
          onClick={(): void => onToggleAutoStart(!config.autoStart)}
          title={
            config.autoStart
              ? 'Disable auto-start; tests will only run when manually started'
              : 'Automatically run the enabled tests in this module whenever a link comes up on the test interface'
          }
          aria-label={config.autoStart ? 'Disable auto-start on link' : 'Enable auto-start on link'}
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
          Tests
        </div>
        <div className="space-y-1">
          {config.tests.map((test) => (
            <label
              key={test.id}
              title={`${test.description} — toggle whether to include "${test.name}" when starting the module`}
              className={cn(
                'flex items-center gap-3 p-2 rounded-lg cursor-pointer transition-colors',
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
                aria-label={`Include ${test.name} when running module`}
                className="w-4 h-4"
                style={{ accentColor: config.color }}
              />
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-text-primary">{test.name}</div>
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
  const [expanded, setExpanded] = useState(false);
  const enabledTestCount = config.tests.filter((t) => t.enabled).length;
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
      <div className={cn(spacing.pad.default, 'flex items-center justify-between')}>
        <div className="flex items-center gap-3 flex-1">
          {/* Enable Toggle */}
          <button
            type="button"
            onClick={(): void => onToggleModule(!config.enabled)}
            className={cn(
              'w-8 h-8 rounded-lg flex items-center justify-center transition-colors',
              config.enabled ? statusColor.badge.successStrong : 'bg-surface-base text-text-muted',
            )}
            title={
              config.enabled
                ? `Disable ${config.displayName} (${config.standard}); tests in this module will be skipped`
                : `Enable ${config.displayName} (${config.standard}); allows running its tests`
            }
            aria-label={
              config.enabled
                ? `Disable ${config.displayName} module`
                : `Enable ${config.displayName} module`
            }
          >
            <Power className="w-4 h-4" />
          </button>

          {/* Module Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span
                className="w-3 h-3 rounded-full flex-shrink-0"
                style={{ backgroundColor: config.color }}
              />
              <h3 className="font-semibold text-text-primary truncate">{config.displayName}</h3>
              <span className="text-xs px-2 py-0.5 rounded-full bg-surface-base text-text-muted">
                {config.standard}
              </span>
            </div>
            <p className="text-xs text-text-muted mt-0.5 truncate">
              {enabledTestCount}/{config.tests.length} tests enabled
            </p>
          </div>

          {/* Status Indicator */}
          <ModuleStatusIndicator status={status} isRunning={isRunning} />
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {/* Configure Button */}
          <button
            type="button"
            onClick={onConfigure}
            className={cn(
              'p-2 rounded-lg transition-colors',
              'text-text-muted hover:text-text-primary',
              'hover:bg-surface-hover',
            )}
            title={`Open the configuration drawer for ${config.displayName} (frame sizes, durations, thresholds)`}
            aria-label={`Configure ${config.displayName}`}
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
              'p-2 rounded-lg transition-colors',
              'text-text-muted hover:text-text-primary',
              'hover:bg-surface-hover',
            )}
            title={
              expanded
                ? `Collapse the ${config.displayName} card and hide test selection`
                : `Expand the ${config.displayName} card to choose which tests to run and toggle auto-start`
            }
            aria-label={expanded ? 'Collapse module card' : 'Expand module card'}
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

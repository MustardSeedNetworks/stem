/**
 * @fileoverview The Stem - Module result tables
 * @description The three read-only result tables a module card renders, and the
 *              number/colour formatting they share.
 *
 * Split out of ModuleCard.tsx, which had reached 799 lines against an 800-line
 * red-flag boundary — five lines of headroom meant an ordinary change turned
 * into a red CI job on a file nobody was refactoring (#946).
 *
 * These are the natural seam: each table takes its results array and renders,
 * with no dependency on the card's state or callbacks. Nothing here is a
 * redesign; the components are unchanged from where they sat before.
 */

import { Check, Clock, XCircle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { cn, status as statusColor } from '../styles/theme';

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
export function FrameSizeResultsTable({
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
export function ServiceFlowResultsTable({
  results,
}: {
  results: ServiceFlowResult[];
}): ReactElement {
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
export function OamResultsTable({ results }: { results: OamMeasurementResult[] }): ReactElement {
  const { t } = useTranslation('modules');
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
                <div className="text-text-muted">{t('labels.delayMinAvgMax')}</div>
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
                <div className="text-text-muted">{t('labels.lossNearFar')}</div>
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

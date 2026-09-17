/** Displays progress reported by the server-owned run plan. */

import type { TFunction } from 'i18next';
import { Clock, Loader2 } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { RunPlanStep, Stats } from '../types/api';

export interface TestProgress {
  status: Stats['testStatus'];
  currentTest: string | null;
  currentStep: number;
  stepsTotal: number;
  phase: string;
  elapsedSeconds: number;
  estimatedRemainingSeconds: number | null;
  steps: NonNullable<Stats['steps']>;
}

interface TestProgressBarProps {
  progress: TestProgress;
}

function formatTime(seconds: number): string {
  const bounded = Math.max(0, seconds);
  const mins = Math.floor(bounded / 60);
  const secs = Math.floor(bounded % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function formatETA(t: TFunction<'common'>, seconds: number): string {
  if (seconds <= 0) return t('status.complete');
  if (seconds < 60) return t('progress.etaSeconds', { count: Math.ceil(seconds) });
  return t('progress.etaMinutes', { count: Math.ceil(seconds / 60) });
}

function statusPresentation(
  t: TFunction<'common'>,
  status: TestProgress['status'],
): [string, string, string] {
  switch (status) {
    case 'starting':
      return [t('status.starting'), 'text-status-info', 'bg-status-info'];
    case 'running':
      return [t('status.running'), 'text-status-success', 'bg-brand-primary'];
    case 'completed':
      return [t('status.completed'), 'text-status-success', 'bg-status-success'];
    case 'cancelled':
      return [t('status.cancelled'), 'text-status-warning', 'bg-status-warning'];
    case 'error':
      return [t('status.error'), 'text-status-error', 'bg-status-error'];
    default:
      return [t('status.idle'), 'text-text-muted', 'bg-text-muted'];
  }
}

// The server sends the step status as a bare string, so each arm spells its
// key out literally: a `t(`status.${step.status}`)` would be invisible to the
// extractor and to the fleet key gate, and the string would be dropped.
function stepStatusLabel(t: TFunction<'common'>, status: RunPlanStep['status']): string {
  switch (status) {
    case 'pending':
      return t('status.pending');
    case 'running':
      return t('status.running');
    case 'passed':
      return t('status.passed');
    case 'failed':
      return t('status.failed');
    case 'skipped':
      return t('status.skipped');
    case 'cancelled':
      return t('status.cancelled');
  }
}

export function TestProgressBar({ progress }: TestProgressBarProps): ReactElement | null {
  const { t } = useTranslation('common');

  if (progress.status === 'idle' || !progress.currentTest) return null;

  const active = progress.status === 'running' || progress.status === 'starting';
  const determinate = progress.estimatedRemainingSeconds !== null;
  const totalEstimate = progress.elapsedSeconds + (progress.estimatedRemainingSeconds ?? 0);
  const percent =
    progress.status === 'completed'
      ? 100
      : determinate && totalEstimate > 0
        ? Math.min(100, (progress.elapsedSeconds / totalEstimate) * 100)
        : 0;
  const [statusText, statusColor, barColor] = statusPresentation(t, progress.status);

  return (
    <div className="card mb-section">
      <div className="flex-between mb-heading">
        <div className="flex items-center gap-compact">
          {active ? <Loader2 className="w-4 h-4 animate-spin text-brand-primary" /> : null}
          <span className="font-medium text-text-primary">{progress.currentTest}</span>
          <span className={`text-sm ${statusColor}`}>({statusText})</span>
        </div>
        <div className="flex items-center gap-default text-sm text-text-muted">
          <span className="flex items-center gap-tight">
            <Clock className="w-3 h-3" />{' '}
            {t('progress.elapsed', { time: formatTime(progress.elapsedSeconds) })}
          </span>
          {active && determinate ? (
            <span>
              {t('progress.eta', { eta: formatETA(t, progress.estimatedRemainingSeconds ?? 0) })}
            </span>
          ) : null}
        </div>
      </div>

      <div
        className="relative h-3 rounded-full bg-surface-base overflow-hidden"
        role="progressbar"
        aria-label={t('accessibility.runPlanProgress')}
        aria-valuenow={determinate ? Math.round(percent) : undefined}
        aria-valuemin={determinate ? 0 : undefined}
        aria-valuemax={determinate ? 100 : undefined}
      >
        <div
          data-testid={determinate ? 'determinate-progress' : 'indeterminate-progress'}
          className={`absolute inset-y-0 left-0 rounded-full ${barColor} ${
            determinate ? 'transition-all duration-300' : 'w-1/3 animate-shimmer'
          }`}
          style={determinate ? { width: `${percent}%` } : undefined}
        />
      </div>

      <div className="flex-between mt-inline text-xs text-text-muted">
        <span>
          {progress.currentStep > 0 && progress.stepsTotal > 0
            ? t('progress.stepOf', {
                current: progress.currentStep,
                total: progress.stepsTotal,
              })
            : null}
          {progress.phase ? ` · ${progress.phase}` : null}
        </span>
        {determinate ? <span className="font-medium">{Math.round(percent)}%</span> : null}
      </div>
      {progress.steps.length > 0 ? (
        <ol
          className="mt-inline grid gap-tight text-xs"
          aria-label={t('accessibility.runPlanSteps')}
        >
          {progress.steps.map((step, index) => (
            <li key={`${step.testType}-${index}`} className="flex-between">
              <span>{step.testType}</span>
              <span>{stepStatusLabel(t, step.status)}</span>
            </li>
          ))}
        </ol>
      ) : null}
    </div>
  );
}

// The counters and the step list carry `omitempty` on the daemon side, so
// an idle or single-step run simply omits them. This is where absent
// becomes the zero the bar renders; the view model stays total.
export function useTestProgress(stats: Stats): TestProgress {
  return {
    status: stats.testStatus,
    currentTest: stats.currentTest,
    currentStep: stats.currentStep ?? 0,
    stepsTotal: stats.stepsTotal ?? 0,
    phase: stats.phase ?? '',
    elapsedSeconds: stats.elapsedSeconds ?? 0,
    estimatedRemainingSeconds: stats.estimatedRemainingSeconds,
    steps: stats.steps ?? [],
  };
}

export default TestProgressBar;

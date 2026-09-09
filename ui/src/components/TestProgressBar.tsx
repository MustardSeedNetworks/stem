/** Displays progress reported by the server-owned run plan. */
import { Clock, Loader2 } from 'lucide-react';
import type { ReactElement } from 'react';
import type { Stats } from '../types/api';

export interface TestProgress {
  status: Stats['testStatus'];
  currentTest: string | null;
  currentStep: number;
  stepsTotal: number;
  phase: string;
  elapsedSeconds: number;
  estimatedRemainingSeconds: number | null;
  steps: Stats['steps'];
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

function formatETA(seconds: number): string {
  if (seconds <= 0) return 'Complete';
  if (seconds < 60) return `~${Math.ceil(seconds)}s`;
  return `~${Math.ceil(seconds / 60)}m`;
}

function statusPresentation(status: TestProgress['status']): [string, string, string] {
  switch (status) {
    case 'starting':
      return ['Starting...', 'text-status-info', 'bg-status-info'];
    case 'running':
      return ['Running', 'text-status-success', 'bg-brand-primary'];
    case 'completed':
      return ['Completed', 'text-status-success', 'bg-status-success'];
    case 'cancelled':
      return ['Cancelled', 'text-status-warning', 'bg-status-warning'];
    case 'error':
      return ['Error', 'text-status-error', 'bg-status-error'];
    default:
      return ['Idle', 'text-text-muted', 'bg-text-muted'];
  }
}

export function TestProgressBar({ progress }: TestProgressBarProps): ReactElement | null {
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
  const [statusText, statusColor, barColor] = statusPresentation(progress.status);

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
            <Clock className="w-3 h-3" /> Elapsed: {formatTime(progress.elapsedSeconds)}
          </span>
          {active && determinate ? (
            <span>ETA: {formatETA(progress.estimatedRemainingSeconds ?? 0)}</span>
          ) : null}
        </div>
      </div>

      <div
        className="relative h-3 rounded-full bg-surface-base overflow-hidden"
        role="progressbar"
        aria-label="Run plan progress"
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
            ? `Step ${progress.currentStep} of ${progress.stepsTotal}`
            : null}
          {progress.phase ? ` · ${progress.phase}` : null}
        </span>
        {determinate ? <span className="font-medium">{Math.round(percent)}%</span> : null}
      </div>
      {progress.steps.length > 0 ? (
        <ol className="mt-inline grid gap-tight text-xs" aria-label="Run plan steps">
          {progress.steps.map((step, index) => (
            <li key={`${step.testType}-${index}`} className="flex-between">
              <span>{step.testType}</span>
              <span>{step.status}</span>
            </li>
          ))}
        </ol>
      ) : null}
    </div>
  );
}

export function useTestProgress(stats: Stats): TestProgress {
  return {
    status: stats.testStatus,
    currentTest: stats.currentTest,
    currentStep: stats.currentStep,
    stepsTotal: stats.stepsTotal,
    phase: stats.phase,
    elapsedSeconds: stats.elapsedSeconds,
    estimatedRemainingSeconds: stats.estimatedRemainingSeconds,
    steps: stats.steps,
  };
}

export default TestProgressBar;

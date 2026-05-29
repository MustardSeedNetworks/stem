/**
 * @fileoverview The Stem - Test Progress Bar Component
 * @description Displays test execution progress with elapsed time, ETA, and visual progress bar.
 */

import { Clock, Loader2 } from 'lucide-react';
import type { ReactElement } from 'react';
import { useEffect, useState } from 'react';

/** Test progress information */
export interface TestProgress {
  /** Current test status */
  status: 'idle' | 'starting' | 'running' | 'completed' | 'cancelled' | 'error';
  /** Current test name */
  currentTest: string | null;
  /** Expected total duration in seconds */
  expectedDuration: number;
  /** Test start timestamp */
  startedAt: number | null;
  /** Current test step (e.g., "1 of 7 frame sizes") */
  currentStep?: string;
  /** Progress percentage (0-100) if provided by backend */
  progressPercent?: number;
}

interface TestProgressBarProps {
  progress: TestProgress;
}

function formatTime(seconds: number): string {
  if (seconds < 0) {
    return '0:00';
  }
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function formatETA(seconds: number): string {
  if (seconds <= 0) {
    return 'Complete';
  }
  if (seconds < 60) {
    return `~${Math.ceil(seconds)}s`;
  }
  const mins = Math.ceil(seconds / 60);
  return `~${mins}m`;
}

export function TestProgressBar({ progress }: TestProgressBarProps): ReactElement | null {
  const [elapsedSeconds, setElapsedSeconds] = useState(0);

  // Update elapsed time every second when test is running
  useEffect(() => {
    if (progress.status !== 'running' || !progress.startedAt) {
      setElapsedSeconds(0);
      return;
    }

    // Calculate initial elapsed time
    const initialElapsed = (Date.now() - progress.startedAt) / 1000;
    setElapsedSeconds(initialElapsed);

    const interval = setInterval(() => {
      if (progress.startedAt) {
        const elapsed = (Date.now() - progress.startedAt) / 1000;
        setElapsedSeconds(elapsed);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [progress.status, progress.startedAt]);

  // Don't show progress bar when idle or no test running
  if (progress.status === 'idle' || !progress.currentTest) {
    return null;
  }

  // Calculate progress percentage
  const calculatedPercent =
    progress.progressPercent ??
    (progress.expectedDuration > 0
      ? Math.min(100, (elapsedSeconds / progress.expectedDuration) * 100)
      : 0);

  // Calculate remaining time
  const remainingSeconds = Math.max(0, progress.expectedDuration - elapsedSeconds);

  // Determine status text and colors
  let statusText: string;
  let statusColor: string;
  let barColor: string;

  switch (progress.status) {
    case 'starting':
      statusText = 'Starting...';
      statusColor = 'text-status-info';
      barColor = 'bg-status-info';
      break;
    case 'running':
      statusText = 'Running';
      statusColor = 'text-status-success';
      barColor = 'bg-brand-primary';
      break;
    case 'completed':
      statusText = 'Completed';
      statusColor = 'text-status-success';
      barColor = 'bg-status-success';
      break;
    case 'cancelled':
      statusText = 'Cancelled';
      statusColor = 'text-status-warning';
      barColor = 'bg-status-warning';
      break;
    case 'error':
      statusText = 'Error';
      statusColor = 'text-status-error';
      barColor = 'bg-status-error';
      break;
    default:
      statusText = 'Unknown';
      statusColor = 'text-text-muted';
      barColor = 'bg-text-muted';
  }

  const isActive = progress.status === 'running' || progress.status === 'starting';

  return (
    <div className="card mb-section">
      {/* Header */}
      <div className="flex-between mb-heading">
        <div className="flex items-center gap-compact">
          {isActive ? <Loader2 className="w-4 h-4 animate-spin text-brand-primary" /> : null}
          <span className="font-medium text-text-primary">{progress.currentTest}</span>
          <span className={`text-sm ${statusColor}`}>({statusText})</span>
        </div>
        <div className="flex items-center gap-default text-sm text-text-muted">
          <div className="flex items-center gap-tight">
            <Clock className="w-3 h-3" />
            <span>Elapsed: {formatTime(elapsedSeconds)}</span>
          </div>
          {isActive && progress.expectedDuration > 0 && (
            <span>ETA: {formatETA(remainingSeconds)}</span>
          )}
        </div>
      </div>

      {/* Progress Bar */}
      <div className="relative h-3 rounded-full bg-surface-base overflow-hidden">
        <div
          className={`absolute inset-y-0 left-0 rounded-full transition-all duration-300 ${barColor}`}
          style={{ width: `${Math.min(100, calculatedPercent)}%` }}
        />
        {/* Animated shine effect for active tests */}
        {isActive ? (
          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-knob/20 to-transparent animate-shimmer" />
        ) : null}
      </div>

      {/* Footer */}
      <div className="flex-between mt-inline text-xs text-text-muted">
        <div>{progress.currentStep ? <span>{progress.currentStep}</span> : null}</div>
        <div className="font-medium">{Math.round(calculatedPercent)}%</div>
      </div>
    </div>
  );
}

/** Hook to manage test progress state */
export function useTestProgress(
  testStatus: 'idle' | 'starting' | 'running' | 'completed' | 'cancelled' | 'error',
  currentTest: string | null,
  expectedDuration: number,
): TestProgress {
  const [startedAt, setStartedAt] = useState<number | null>(null);

  // Track when test starts
  useEffect(() => {
    if (testStatus === 'starting' || testStatus === 'running') {
      setStartedAt((prev) => prev ?? Date.now());
    } else {
      setStartedAt(null);
    }
  }, [testStatus]);

  return {
    status: testStatus,
    currentTest,
    expectedDuration,
    startedAt,
  };
}

export default TestProgressBar;

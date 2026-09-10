/**
 * StopOutcomeMessage — what came back from the last stop request.
 *
 * Stop used to be fire-and-forget: `handleStopTest` ignored `response.ok`, so
 * the daemon refusing ("No test is currently running") and the daemon actually
 * stopping the run looked the same to the operator — a spinner that ended
 * (#1080). Both stop controls render this, so the two surfaces cannot drift.
 *
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */
import { AlertTriangle, Square } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { StopOutcome } from '../stores/test-store';

interface StopOutcomeMessageProps {
  outcome: StopOutcome;
}

export function StopOutcomeMessage({ outcome }: StopOutcomeMessageProps): ReactElement | null {
  const { t } = useTranslation('common');

  if (outcome.kind === 'idle' || outcome.kind === 'stopping') {
    return null;
  }

  if (outcome.kind === 'stopped') {
    return (
      <div
        className="text-sm text-text-muted flex items-center gap-compact"
        role="status"
        aria-live="polite"
        data-testid="test-stop-message"
      >
        <Square className="w-4 h-4" aria-hidden="true" />
        {t('status.stopped')}
      </div>
    );
  }

  // stopRejected carries the daemon's own sentence; stopFailed carries ours,
  // because in that branch the request never got an answer.
  return (
    <div
      className="text-sm text-status-error flex items-center gap-compact"
      role="alert"
      aria-live="assertive"
      data-testid="test-stop-message"
    >
      <AlertTriangle className="w-4 h-4" aria-hidden="true" />
      {outcome.message}
    </div>
  );
}

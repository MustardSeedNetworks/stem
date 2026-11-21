/**
 * useActionToasts — raises a toast when a test Start or Stop fails.
 *
 * The inline messages beside the run controls stay; they are only visible on
 * the pages that carry the controls, so a run started from one page and
 * refused after the operator navigated away went unseen.
 */
import { useEffect } from 'react';
import { useToast } from '../components/ui/useToast';
import type { StopOutcome } from '../stores/test-store';

export function useActionToasts(testStartError: string | null, stopOutcome: StopOutcome): void {
  const { addToast } = useToast();

  useEffect(() => {
    if (testStartError) {
      addToast(testStartError, 'error');
    }
  }, [testStartError, addToast]);

  useEffect(() => {
    if (stopOutcome.kind === 'stopRejected' || stopOutcome.kind === 'stopFailed') {
      addToast(stopOutcome.message, 'error');
    }
  }, [stopOutcome, addToast]);
}

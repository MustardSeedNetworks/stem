/**
 * Deadlines for requests the operator is waiting on.
 *
 * fetch has no default timeout. A blocking submit — first-run setup, login —
 * that never settles leaves its form disabled with no error and no retry, so
 * the operator's only way forward is a page reload they have no reason to
 * think will help. Every such request carries a deadline.
 */

/** Long enough for a loaded box to answer, short enough to stay a UI. */
export const REQUEST_TIMEOUT_MS = 15_000;

/** Abort signal for a request an operator is blocked on. */
export function requestDeadline(): AbortSignal {
  return AbortSignal.timeout(REQUEST_TIMEOUT_MS);
}

/**
 * True when a rejected fetch was our own deadline firing rather than a
 * transport failure — the two need different copy, since a timeout means the
 * server was reachable and did not answer.
 */
export function isDeadlineExceeded(err: unknown): boolean {
  // Checked by name rather than by instanceof: AbortSignal.timeout rejects
  // with a DOMException, which is not an instance of Error in every realm the
  // app and its tests run in.
  return (
    typeof err === 'object' && err !== null && (err as { name?: unknown }).name === 'TimeoutError'
  );
}

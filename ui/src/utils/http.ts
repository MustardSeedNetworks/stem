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
 * True when `signal` is the reason a request failed — our own deadline firing
 * rather than a transport failure. The two need different copy, since a
 * timeout means the server was reachable and did not answer.
 *
 * Asked of the signal, not of the rejection. Engines disagree on the error:
 * Chromium rejects an aborted fetch with `TimeoutError`, WebKit with
 * `AbortError`. Matching on the name meant Safari users saw WebKit's internal
 * "Fetch is aborted" text instead of the sentence written for them — the same
 * defect as seed#2256. Nothing else holds these signals, so `aborted` means
 * the deadline expired.
 */
export function deadlineExpired(signal: AbortSignal): boolean {
  return signal.aborted;
}

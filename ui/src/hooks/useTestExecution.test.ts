/**
 * @fileoverview classifyFailure and resolveStopOutcome — how a refused request
 *               becomes something an operator can read.
 * @description The 402 entitlement answer must be distinguishable from a
 *              generic failure, or the render path cannot tell an operator
 *              what to buy (#1070). A refused *stop* must be an outcome the
 *              UI models rather than a swallowed log line (#1080).
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { describe, expect, it } from 'vitest';
import { classifyFailure, resolveStopOutcome } from './useTestExecution';

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('classifyFailure', () => {
  it('reads the missing feature out of a 402 TIER_TOO_LOW body', async () => {
    const failure = await classifyFailure(
      jsonResponse(402, {
        error: 'Feature requires a higher tier',
        code: 'TIER_TOO_LOW',
        requiredFeature: 'rfc2544',
        currentTier: 'Reflector',
      }),
    );

    expect(failure).toEqual({ kind: 'featureGate', feature: 'rfc2544' });
  });

  it('treats a 402 without the code as an ordinary failure', async () => {
    const failure = await classifyFailure(jsonResponse(402, { error: 'payment plumbing' }));

    expect(failure).toEqual({ kind: 'message', message: 'payment plumbing' });
  });

  // The daemon's error envelope (internal/api/errors.go HTTPErrorResponse) puts
  // the *type* in `error` ("Bad Request") and the operator-facing text in
  // `message`. Reading `error` — which this classifier did — showed every
  // refusal as "Bad Request" and threw away the only sentence that says what
  // went wrong. Body copied verbatim from a live daemon.
  it('prefers the daemon message over the error type on a 400', async () => {
    const failure = await classifyFailure(
      jsonResponse(400, {
        error: 'Bad Request',
        code: 'INVALID_REQUEST',
        message: 'No test is currently running',
      }),
    );

    expect(failure).toEqual({ kind: 'message', message: 'No test is currently running' });
  });

  it('falls back to the error type when the body carries no message', async () => {
    const failure = await classifyFailure(
      jsonResponse(400, { error: 'Unknown or unsupported test type', code: 'INVALID_REQUEST' }),
    );

    expect(failure).toEqual({ kind: 'message', message: 'Unknown or unsupported test type' });
  });

  it('falls back to no message when the body is not JSON', async () => {
    const failure = await classifyFailure(new Response('gateway timeout', { status: 504 }));

    expect(failure).toEqual({ kind: 'message' });
  });
});

describe('resolveStopOutcome', () => {
  const fallback = 'Failed to stop test';

  it('reports a 2xx as stopped', async () => {
    const outcome = await resolveStopOutcome(jsonResponse(200, { status: 'stopped' }), fallback);

    expect(outcome).toEqual({ kind: 'stopped' });
  });

  // The row's first acceptance clause: the daemon refuses a stop when nothing
  // is running, and the operator must be shown that sentence, not a spinner
  // that silently ends.
  it('reports the daemon refusal verbatim as stopRejected', async () => {
    const outcome = await resolveStopOutcome(
      jsonResponse(400, {
        error: 'Bad Request',
        code: 'INVALID_REQUEST',
        message: 'No test is currently running',
      }),
      fallback,
    );

    expect(outcome).toEqual({ kind: 'stopRejected', message: 'No test is currently running' });
  });

  it('uses the fallback text when the refusal carries no readable body', async () => {
    const outcome = await resolveStopOutcome(new Response('', { status: 502 }), fallback);

    expect(outcome).toEqual({ kind: 'stopRejected', message: fallback });
  });
});

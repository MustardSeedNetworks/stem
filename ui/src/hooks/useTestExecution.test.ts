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
import {
  classifyFailure,
  isTerminalTestStatus,
  normalizeTestStatus,
  resolveStopOutcome,
  resolveTestResult,
} from './useTestExecution';

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

describe('resolveTestResult', () => {
  it('turns a failed fetch into a visible result with the HTTP status', async () => {
    const result = await resolveTestResult(new Response('', { status: 500 }));

    expect(result).toMatchObject({
      status: 'error',
      success: false,
      error: 'Result unavailable (HTTP 500)',
    });
  });

  it('preserves the daemon error reported by stats when the result omits it', async () => {
    const result = await resolveTestResult(
      jsonResponse(200, { testType: 'rfc2544_latency', module: 'benchmark', status: 'error' }),
      'Latency measurement failed',
    );

    expect(result).toMatchObject({ success: false, error: 'Latency measurement failed' });
  });

  // The endpoint answers mid-run too. Pinning a running run as "the result"
  // would freeze a snapshot the card then never replaces.
  it.each(['idle', 'starting', 'running'] as const)(
    'has nothing to pin while %s',
    async (status) => {
      const result = await resolveTestResult(
        jsonResponse(200, { testType: 'reflect', module: 'reflector', status }),
      );

      expect(result).toBeNull();
    },
  );

  // A reflector the operator stopped is finished, and its result is the only
  // record of what the run measured. Dropping it is what left the card saying
  // no test had run (#1248).
  it('keeps the result of a run that was stopped', async () => {
    const result = await resolveTestResult(
      jsonResponse(200, {
        testType: 'reflect',
        module: 'reflector',
        status: 'stopped',
        success: true,
      }),
    );

    expect(result).toMatchObject({ status: 'stopped', testType: 'reflect' });
  });
});

/**
 * The daemon reports `testStatus: "stopped"` when the reflector is stopped
 * (internal/api/types.go `statusStopped`). The union had no such member and
 * the normaliser mapped anything unknown to 'idle', so a just-stopped run read
 * to the rest of the UI as "nothing ever ran": no terminal transition, so no
 * result fetch, so the Results card showed the idle placeholder straight after
 * a real run (#1248).
 */
describe('normalizeTestStatus', () => {
  it.each(['idle', 'starting', 'running', 'completed', 'cancelled', 'stopped', 'error'] as const)(
    'passes %s through unchanged',
    (status) => {
      expect(normalizeTestStatus(status)).toBe(status);
    },
  );

  // The fallback is deliberate, and it is also what hid this defect: a status
  // the daemon really sends and the switch does not list disappears silently.
  it.each([undefined, '', 'paused'])('falls back to idle for %s', (status) => {
    expect(normalizeTestStatus(status)).toBe('idle');
  });
});

describe('isTerminalTestStatus', () => {
  // A stop is as final as a completion. This is the single list both the
  // transition detector and the result filter read, so they cannot drift.
  it.each(['completed', 'error', 'cancelled', 'stopped'] as const)(
    'treats %s as the end of a run',
    (status) => {
      expect(isTerminalTestStatus(status)).toBe(true);
    },
  );

  it.each(['idle', 'starting', 'running'] as const)('leaves %s running', (status) => {
    expect(isTerminalTestStatus(status)).toBe(false);
  });
});

/**
 * @fileoverview classifyStartFailure — the failed-start classifier.
 * @description The 402 entitlement answer must be distinguishable from a
 *              generic failure, or the render path cannot tell an operator
 *              what to buy (#1070).
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { describe, expect, it } from 'vitest';
import { classifyStartFailure } from './useTestExecution';

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('classifyStartFailure', () => {
  it('reads the feature and tier out of a 402 TIER_TOO_LOW body', async () => {
    const failure = await classifyStartFailure(
      jsonResponse(402, {
        error: 'Feature requires a higher tier',
        code: 'TIER_TOO_LOW',
        requiredFeature: 'rfc2544',
        currentTier: 'Reflector',
      }),
    );

    expect(failure).toEqual({ kind: 'featureGate', feature: 'rfc2544', tier: 'Reflector' });
  });

  it('treats a 402 without the code as an ordinary failure', async () => {
    const failure = await classifyStartFailure(jsonResponse(402, { error: 'payment plumbing' }));

    expect(failure).toEqual({ kind: 'message', message: 'payment plumbing' });
  });

  it('passes a 400 through with its message', async () => {
    const failure = await classifyStartFailure(
      jsonResponse(400, { error: 'Unknown or unsupported test type', code: 'INVALID_REQUEST' }),
    );

    expect(failure).toEqual({ kind: 'message', message: 'Unknown or unsupported test type' });
  });

  it('falls back to no message when the body is not JSON', async () => {
    const failure = await classifyStartFailure(new Response('gateway timeout', { status: 504 }));

    expect(failure).toEqual({ kind: 'message' });
  });
});

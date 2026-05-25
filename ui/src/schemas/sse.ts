/**
 * Valibot schemas for the stem SSE channel (#296).
 *
 * The backend at /api/v1/events streams JSON-encoded frames whose
 * `type` discriminator narrows the `payload` shape. Schemas here let
 * the consumer hook safeParse each frame before dispatching to
 * subscribers — invalid frames are dropped + logged rather than
 * propagating into render code that assumes a valid shape.
 *
 * Mirrors seed's ui/src/schemas/sse.ts pattern.
 */
import * as v from 'valibot';

import { ModeUpdateResponseSchema } from './role';

/**
 * SseEnvelope — outer shape every frame carries. The narrower
 * per-type schemas (below) parse the inner `payload`.
 */
export const SseEnvelopeSchema = v.object({
  type: v.string(),
  payload: v.unknown(),
});

export type SseEnvelope = v.InferOutput<typeof SseEnvelopeSchema>;

/**
 * Frame: mode_changed — published by POST /api/v1/mode. Payload
 * matches the response shape of the mode-switch endpoint so the
 * consumer can reuse RoleContext's existing handler.
 */
export const SseModeChangedSchema = v.object({
  type: v.literal('mode_changed'),
  payload: ModeUpdateResponseSchema,
});

export type SseModeChanged = v.InferOutput<typeof SseModeChangedSchema>;

/**
 * Frame: reflector_stats — published by the server's reflector-stats
 * publisher every ~1s when reflector mode is active. Field set
 * mirrors the REST /api/v1/reflector/stats response. We declare the
 * fields as optional because the publisher may grow the response
 * over time and we'd rather drop a known frame than reject the
 * entire envelope.
 */
export const SseReflectorStatsSchema = v.object({
  type: v.literal('reflector_stats'),
  payload: v.object({
    running: v.optional(v.boolean()),
    packetsReceived: v.optional(v.number()),
    packetsReflected: v.optional(v.number()),
    bytesReceived: v.optional(v.number()),
    bytesReflected: v.optional(v.number()),
    txErrors: v.optional(v.number()),
    rxInvalid: v.optional(v.number()),
    ratePps: v.optional(v.number()),
    rateMbps: v.optional(v.number()),
    uptime: v.optional(v.number()),
  }),
});

export type SseReflectorStats = v.InferOutput<typeof SseReflectorStatsSchema>;

/**
 * Frame: test_progress — emitted by test runners during a long
 * sweep (RFC 2544, Y.1564, etc.). The payload shape is intentionally
 * loose because each runner emits its own progress structure.
 */
export const SseTestProgressSchema = v.object({
  type: v.literal('test_progress'),
  payload: v.object({
    testId: v.string(),
    progress: v.unknown(),
  }),
});

export type SseTestProgress = v.InferOutput<typeof SseTestProgressSchema>;

/**
 * parseSseEnvelope — runs the envelope schema on raw JSON.parse
 * output and returns the typed envelope or null on shape mismatch.
 */
export function parseSseEnvelope(value: unknown): SseEnvelope | null {
  const result = v.safeParse(SseEnvelopeSchema, value);
  return result.success ? result.output : null;
}

/**
 * parseSseModeChanged — narrow an envelope to a mode_changed frame.
 * Returns null if the type discriminator or payload shape mismatches.
 */
export function parseSseModeChanged(envelope: SseEnvelope): SseModeChanged | null {
  if (envelope.type !== 'mode_changed') {
    return null;
  }
  const result = v.safeParse(SseModeChangedSchema, envelope);
  return result.success ? result.output : null;
}

/**
 * parseSseReflectorStats — narrow to a reflector_stats frame.
 */
export function parseSseReflectorStats(envelope: SseEnvelope): SseReflectorStats | null {
  if (envelope.type !== 'reflector_stats') {
    return null;
  }
  const result = v.safeParse(SseReflectorStatsSchema, envelope);
  return result.success ? result.output : null;
}

/**
 * parseSseTestProgress — narrow to a test_progress frame.
 */
export function parseSseTestProgress(envelope: SseEnvelope): SseTestProgress | null {
  if (envelope.type !== 'test_progress') {
    return null;
  }
  const result = v.safeParse(SseTestProgressSchema, envelope);
  return result.success ? result.output : null;
}

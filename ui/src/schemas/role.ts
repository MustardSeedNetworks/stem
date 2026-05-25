/**
 * Valibot schemas for the stem mode-switch API surface.
 *
 * Before this layer, RoleContext.tsx carried hand-rolled type guards
 * (isStemRole, isModeUpdateResponse, ad-hoc extractErrorMessage). Each
 * was tied to a single field shape and drifted from the Go side via
 * normal code rot. Centralising the shapes here:
 *
 *   - keeps one source of truth (the schema) for both the runtime
 *     check and the static type (via v.InferOutput),
 *   - makes adding a new mode (e.g. "passthrough") a one-place change,
 *   - drops the boilerplate of writing a custom predicate per shape.
 *
 * The Go side lives at internal/api/types.go (ModeUpdateResponse,
 * modeReflector / modeTestMaster constants). When that changes, this
 * schema needs to follow — until the JSON-Schema generation pipeline
 * lands (stem#269/#271), the link is manual.
 */
import * as v from 'valibot';

/**
 * StemRole — the two valid operating modes. Mirrors modeReflector /
 * modeTestMaster constants in internal/api/types.go.
 */
export const StemRoleSchema = v.picklist(['reflector', 'test_master']);

export type StemRole = v.InferOutput<typeof StemRoleSchema>;

export const DEFAULT_ROLE: StemRole = 'reflector';

/**
 * ModeUpdateResponse — body of POST /api/v1/mode. Mirrors the Go
 * ModeUpdateResponse struct (internal/api/types.go). `status` is
 * "updated" when the mode changed and "unchanged" when the server was
 * already in the requested mode; `previous` is always populated and
 * equals `mode` when status == "unchanged".
 */
export const ModeUpdateResponseSchema = v.object({
  status: v.string(),
  mode: StemRoleSchema,
  previous: StemRoleSchema,
});

export type ModeUpdateResponse = v.InferOutput<typeof ModeUpdateResponseSchema>;

/**
 * ApiErrorBody — generic 4xx/5xx envelope (HTTPErrorResponse in
 * internal/api/errors.go). Both fields are optional because some
 * fallback paths only set one or the other; the consumer picks
 * whichever is non-empty.
 */
export const ApiErrorBodySchema = v.object({
  message: v.optional(v.string()),
  error: v.optional(v.string()),
});

export type ApiErrorBody = v.InferOutput<typeof ApiErrorBodySchema>;

/**
 * parseModeUpdateResponse — runs the schema check on an unknown value
 * (typically `await response.json()`). Returns `{ok: true, value}` on
 * success or `{ok: false}` so the caller can surface a generic
 * "unexpected server response" message without leaking parser details.
 */
export function parseModeUpdateResponse(
  body: unknown,
): { ok: true; value: ModeUpdateResponse } | { ok: false } {
  const result = v.safeParse(ModeUpdateResponseSchema, body);
  return result.success ? { ok: true, value: result.output } : { ok: false };
}

/**
 * parseApiErrorBody — runs the schema check on a value expected to be
 * an HTTPErrorResponse. Returns the parsed object on success, or null
 * on failure so the caller can fall back to a generic message.
 */
export function parseApiErrorBody(body: unknown): ApiErrorBody | null {
  const result = v.safeParse(ApiErrorBodySchema, body);
  return result.success ? result.output : null;
}

/**
 * isStemRole — narrow type guard for inline branches that don't need
 * a full schema parse (e.g., reading a localStorage value).
 */
export function isStemRole(value: unknown): value is StemRole {
  return v.safeParse(StemRoleSchema, value).success;
}

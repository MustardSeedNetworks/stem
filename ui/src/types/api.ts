/**
 * @fileoverview Shared API types for Stem
 * @description The wire DTOs are GENERATED from the Go structs (see
 * ./generated, produced by `make schema && npm run gen-types` and gated by
 * scripts/check-schema-drift.sh + scripts/check-types-drift.sh). This file
 * re-exports them under the names the UI already imports and adds the
 * things a schema cannot carry: the seed value for `Stats` and the runtime
 * guards used where a response is parsed from `unknown`.
 *
 * Do not hand-write a DTO interface here. These were transcribed by hand
 * once and drifted from the daemon — a TypeScript-only `Stats.errorMessage`
 * (#1251), an `InterfaceInfo` missing mtu/ipv4/ipv6, a `TestResult` with
 * four fields the daemon never sent, and an `AuthResponse` claiming
 * `expiresIn` where the wire carries `expiresAt`.
 */

import type { AuthLoginResponse } from './generated/auth-login-response';
import type { InterfaceInfo } from './generated/interface-info';
import type { RunPlanStep, Stats, TestResultResponse } from './generated/stats';

export type { AuthLoginResponse, InterfaceInfo, RunPlanStep, Stats, TestResultResponse };

/** Test status values, narrowed by the enum tag on the Go field. */
export type TestStatus = Stats['testStatus'];

/**
 * The daemon's result payload. Named `TestResult` at the call sites; the
 * wire DTO is `TestResultResponse` in internal/api.
 */
export type TestResult = TestResultResponse;

/** Initial stats state */
export const initialStats: Stats = {
  packetsReceived: 0,
  packetsSent: 0,
  bytesReceived: 0,
  bytesSent: 0,
  currentPps: 0,
  currentMbps: 0,
  uptime: 0,
  testStatus: 'idle',
  currentTest: null,
  suiteId: '',
  steps: [],
  currentStep: 0,
  stepsComplete: 0,
  stepsTotal: 0,
  phase: '',
  elapsedSeconds: 0,
  estimatedRemainingSeconds: null,
};

/** Validate InterfaceInfo array response */
export function isValidInterfaceArray(data: unknown): data is InterfaceInfo[] {
  if (!Array.isArray(data)) {
    return false;
  }
  return data.every(
    (item) =>
      typeof item === 'object' &&
      item !== null &&
      typeof item.name === 'string' &&
      typeof item.mac === 'string' &&
      typeof item.speed === 'number',
  );
}

/** Validate Stats response */
export function isValidStats(data: unknown): data is Partial<Stats> {
  if (typeof data !== 'object' || data === null) {
    return false;
  }
  const obj = data as Record<string, unknown>;
  // At minimum, uptime should be a number
  return typeof obj.uptime === 'number' || typeof obj.packetsReceived === 'number';
}

/** Validate the login/MFA-verify response */
export function isValidAuthResponse(data: unknown): data is AuthLoginResponse {
  if (typeof data !== 'object' || data === null) {
    return false;
  }
  const obj = data as Record<string, unknown>;
  return typeof obj.token === 'string';
}

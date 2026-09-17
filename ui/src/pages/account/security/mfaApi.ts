/**
 * MFA API client.
 *
 * Wraps the /api/v1/auth/totp/* and /api/v1/auth/webauthn/* endpoints
 * introduced in Wave 3 (#85).
 *
 * Every post-login call goes through `authFetch`, which is the one client that
 * carries BOTH halves of the daemon's session contract: the CSRF header with a
 * re-fetch on 403, and the 401 -> refresh -> retry that this module used to
 * skip. Without it an expired access token turned every MFA action into a raw
 * "token expired" body on the Security page, and the operator was stuck until
 * a reload (#1253). `loginTotp` is the exception on purpose: it runs before a
 * session exists, so it is CSRF-exempt and `authFetch` would reject it.
 */

import { authFetch } from '../../../stores/auth-store';

const API_BASE = '/api/v1';

export class MFAError extends Error {
  public readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'MFAError';
    this.status = status;
  }
}

export interface MFAStatusResponse {
  totpEnabled: boolean;
  webauthnRegistered: boolean;
  webauthnCredentialCount: number;
}

export interface TotpSetupResponse {
  secret: string;
  provisioningUri: string;
  qrCodePngBase64: string;
}

export interface MFARequiredResponse {
  mfaRequired: true;
  mfaToken: string;
  factor: string;
}

export interface AuthLoginResponse {
  token: string;
  refreshToken?: string;
  expiresAt: number;
}

export type LoginResponse = MFARequiredResponse | AuthLoginResponse;

/**
 * Type-guard: did the login endpoint return an MFA challenge?
 */
export function isMFARequired(value: LoginResponse): value is MFARequiredResponse {
  return (value as MFARequiredResponse).mfaRequired === true;
}

// authFetch returns the Response rather than throwing on a non-2xx, so
// MFAError's status mapping stays here where the MFA surface's error semantics
// live. It does throw when the session is gone (401 after a failed refresh,
// or a CSRF 403 that a fresh token could not clear) — that is a dead session,
// not an MFA outcome, and it belongs to the auth store.
async function request<T>(path: string, body?: unknown): Promise<T> {
  const init: RequestInit =
    body === undefined
      ? {}
      : {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        };
  const response = await authFetch(`${API_BASE}${path}`, init);
  if (!response.ok) {
    const text = await response.text();
    throw new MFAError(response.status, text || `HTTP ${response.status}`);
  }
  return (await response.json()) as T;
}

export const mfaApi = {
  status: (): Promise<MFAStatusResponse> => request<MFAStatusResponse>('/auth/mfa/status'),

  totpSetup: (): Promise<TotpSetupResponse> => request<TotpSetupResponse>('/auth/totp/setup', {}),

  totpVerify: (code: string): Promise<{ success: boolean; totpEnabled: boolean }> =>
    request<{ success: boolean; totpEnabled: boolean }>('/auth/totp/verify', { code }),

  totpDisable: (
    password: string,
    code: string,
  ): Promise<{ success: boolean; totpEnabled: boolean }> =>
    request<{ success: boolean; totpEnabled: boolean }>('/auth/totp/disable', {
      password,
      code,
    }),

  loginTotp: (mfaToken: string, code: string): Promise<AuthLoginResponse> => {
    // CSRF-exempt path — same exemption rationale as /auth/login.
    return fetch(`${API_BASE}/auth/login/totp`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ mfaToken, code }),
    }).then(async (response) => {
      if (!response.ok) {
        const text = await response.text();
        throw new MFAError(response.status, text || `HTTP ${response.status}`);
      }
      return (await response.json()) as AuthLoginResponse;
    });
  },
};

/**
 * MFA API client.
 *
 * Wraps the /api/v1/auth/totp/* and /api/v1/auth/webauthn/* endpoints
 * introduced in Wave 3 (#85). The client mirrors the established
 * ApiError-based pattern from src/api/profiles.ts.
 *
 * CSRF: state-changing POSTs go through `fetchWithCsrf` from lib/csrf, which
 * attaches X-Csrf-Token and re-fetches once on a 403. This module used to carry
 * its own `fetchCsrfToken` and thread the token through every method signature.
 * Two implementations of one thing is how the two get to disagree, and they did:
 * the shared helper caches the token and retries on 403 because the daemon
 * rotates it on login, and the local copy did neither — so MFA enrolment
 * attempted after a session rotation failed where a role switch recovered (#953).
 */

import { fetchWithCsrf } from '../../../lib/csrf';

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

// fetchWithCsrf returns the Response rather than throwing, so MFAError's status
// mapping stays here where the MFA surface's error semantics live.
async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const response = await fetchWithCsrf(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body ?? {}),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new MFAError(response.status, text || `HTTP ${response.status}`);
  }
  return (await response.json()) as T;
}

async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
  });
  if (!response.ok) {
    const text = await response.text();
    throw new MFAError(response.status, text || `HTTP ${response.status}`);
  }
  return (await response.json()) as T;
}

export const mfaApi = {
  status: (): Promise<MFAStatusResponse> => getJSON<MFAStatusResponse>('/auth/mfa/status'),

  totpSetup: (): Promise<TotpSetupResponse> => postJSON<TotpSetupResponse>('/auth/totp/setup', {}),

  totpVerify: (code: string): Promise<{ success: boolean; totpEnabled: boolean }> =>
    postJSON<{ success: boolean; totpEnabled: boolean }>('/auth/totp/verify', { code }),

  totpDisable: (
    password: string,
    code: string,
  ): Promise<{ success: boolean; totpEnabled: boolean }> =>
    postJSON<{ success: boolean; totpEnabled: boolean }>('/auth/totp/disable', {
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

  webauthnRegisterBegin: (): Promise<PublicKeyCredentialCreationOptions> =>
    postJSON<PublicKeyCredentialCreationOptions>('/auth/webauthn/register/begin', {}),

  webauthnRegisterFinish: (
    credential: PublicKeyCredential,
  ): Promise<{ success: boolean; credentialId: string }> =>
    postJSON<{ success: boolean; credentialId: string }>(
      '/auth/webauthn/register/finish',
      credential,
    ),
};

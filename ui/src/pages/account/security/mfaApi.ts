/**
 * MFA API client.
 *
 * Wraps the /api/v1/auth/totp/* and /api/v1/auth/webauthn/* endpoints
 * introduced in Wave 3 (#85). The client mirrors the established
 * ApiError-based pattern from src/api/profiles.ts.
 *
 * CSRF: state-changing POSTs MUST carry the X-Csrf-Token header. The
 * helper `fetchCsrfToken` is called once per UI session and cached;
 * the caller threads it through the request headers.
 */

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

/** Fetch the current session's CSRF token. */
export async function fetchCsrfToken(): Promise<string> {
  const response = await fetch(`${API_BASE}/auth/csrf-token`, {
    credentials: 'include',
  });
  if (!response.ok) {
    throw new MFAError(response.status, `Failed to fetch CSRF token: ${response.status}`);
  }
  const data = (await response.json()) as { token: string };
  return data.token;
}

async function postJSON<T>(path: string, body: unknown, csrf: string): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      'X-Csrf-Token': csrf,
    },
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

  totpSetup: (csrf: string): Promise<TotpSetupResponse> =>
    postJSON<TotpSetupResponse>('/auth/totp/setup', {}, csrf),

  totpVerify: (code: string, csrf: string): Promise<{ success: boolean; totpEnabled: boolean }> =>
    postJSON<{ success: boolean; totpEnabled: boolean }>('/auth/totp/verify', { code }, csrf),

  totpDisable: (
    password: string,
    code: string,
    csrf: string,
  ): Promise<{ success: boolean; totpEnabled: boolean }> =>
    postJSON<{ success: boolean; totpEnabled: boolean }>(
      '/auth/totp/disable',
      { password, code },
      csrf,
    ),

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

  webauthnRegisterBegin: (csrf: string): Promise<PublicKeyCredentialCreationOptions> =>
    postJSON<PublicKeyCredentialCreationOptions>('/auth/webauthn/register/begin', {}, csrf),

  webauthnRegisterFinish: (
    credential: PublicKeyCredential,
    csrf: string,
  ): Promise<{ success: boolean; credentialId: string }> =>
    postJSON<{ success: boolean; credentialId: string }>(
      '/auth/webauthn/register/finish',
      credential,
      csrf,
    ),
};

/**
 * Auth Store
 *
 * Zustand store for authentication state and the authenticated-fetch primitive
 * (stage 3 of the App.tsx decomposition). Security-sensitive; cross-model
 * reviewed. Key invariants:
 *
 * - Tokens live in httpOnly cookies — JS never reads or stores them. We only
 *   track a boolean `isAuthenticated` flag, mirrored to localStorage so a
 *   reload restores the signed-in shell. Every request uses
 *   `credentials: 'include'`.
 * - NOT persisted via Zustand (the localStorage flag is the only persisted bit).
 * - `refreshAccessToken` is single-flight: concurrent 401s share one in-flight
 *   refresh, so one failed refresh can't expire a session another request just
 *   renewed.
 * - On logout/expiry we `cancelQueries()` THEN `clear()` the React Query cache,
 *   so a request in flight at logout cannot resolve afterwards and repopulate
 *   cross-session data.
 * - 401 → refresh+retry once; only a failed refresh or a retry that is *still*
 *   401 expires the session (a retried 5xx/4xx is returned to the caller).
 * - 403 → if the body is `{code:"PERMISSION_DENIED"}` it's an authorization
 *   denial for a still-valid session (returned to the caller); any other 403
 *   (CSRF-invalid / unknown, which the server sends as plain text) is treated
 *   as session-invalid and expires the session.
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { getQueryClient } from '../lib/queryClient';
import { isValidAuthResponse } from '../types/api';

const AUTH_FLAG_KEY = 'stem-authenticated';

/** Server response for GET /api/v1/setup/status. */
export interface SetupStatus {
  needsSetup: boolean;
  username?: string;
  suggestedPassword?: string;
  setupToken?: string;
}

/** Server response for GET /api/v1/recovery/status. */
export interface RecoveryStatus {
  active: boolean;
  remainingTime?: number;
  instructions?: string;
}

/** Outcome of a login / MFA-verify attempt. */
export type LoginResult =
  | { status: 'ok' }
  | { status: 'mfa'; mfaToken: string; factor: string }
  | { status: 'error'; message: string };

function readAuthFlag(): boolean {
  return typeof window !== 'undefined' && window.localStorage.getItem(AUTH_FLAG_KEY) === 'true';
}

function writeAuthFlag(value: boolean): void {
  if (typeof window === 'undefined') {
    return;
  }
  if (value) {
    window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
  } else {
    window.localStorage.removeItem(AUTH_FLAG_KEY);
  }
}

// Single-flight refresh: all concurrent 401s await the same in-flight promise.
let refreshPromise: Promise<boolean> | null = null;

function refreshAccessToken(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = (async (): Promise<boolean> => {
      try {
        // Refresh token is in an httpOnly cookie, sent automatically.
        const response = await fetch('/api/v1/auth/refresh', {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({}),
        });
        return response.ok;
      } catch {
        return false;
      }
    })().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}

/** Cancel in-flight queries, then clear the cache (order matters — see header). */
async function purgeQueryCache(): Promise<void> {
  const queryClient = getQueryClient();
  await queryClient.cancelQueries();
  queryClient.clear();
}

interface AuthState {
  isAuthenticated: boolean;
  loginLoading: boolean;
  loginError: string | null;
  mfaPending: { mfaToken: string; factor: string } | null;
  setupStatus: SetupStatus | null;
  setupChecked: boolean;
  recoveryStatus: RecoveryStatus | null;
  showRecoveryForm: boolean;
}

interface AuthActions {
  /** Username/password login. Resolves to ok | mfa-required | error. */
  login: (username: string, password: string) => Promise<LoginResult>;
  /** Complete an MFA challenge with a TOTP code. */
  verifyMfa: (code: string) => Promise<LoginResult>;
  /**
   * Silent login used by the setup wizard once it has provisioned an admin.
   * Returns true on success. Does not surface loginError/loginLoading.
   */
  performLogin: (username: string, password: string) => Promise<boolean>;
  /** Explicit user logout: server cookie clear + local teardown. */
  logout: () => Promise<void>;
  /** Tear down the session locally (called by authFetch on 401/invalid 403). */
  expireSession: (message?: string) => void;
  setLoginError: (message: string | null) => void;
  setSetupStatus: (status: SetupStatus | null) => void;
  setRecoveryStatus: (status: RecoveryStatus | null) => void;
  setShowRecoveryForm: (show: boolean) => void;
  /** One-shot check of setup + recovery status on app load. */
  checkStatuses: () => Promise<void>;
}

export type AuthStore = AuthState & AuthActions;

/** Local teardown shared by logout + expireSession. */
function tearDownSession(set: (partial: Partial<AuthState>) => void, message: string | null): void {
  writeAuthFlag(false);
  set({ isAuthenticated: false, mfaPending: null, loginError: message });
  // Cancel-then-clear so a late in-flight request can't repopulate the cache.
  void purgeQueryCache();
}

export const useAuthStore = create<AuthStore>()(
  devtools(
    (set, get) => ({
      isAuthenticated: readAuthFlag(),
      loginLoading: false,
      loginError: null,
      mfaPending: null,
      setupStatus: null,
      setupChecked: false,
      recoveryStatus: null,
      showRecoveryForm: false,

      login: async (username, password): Promise<LoginResult> => {
        set({ loginLoading: true, loginError: null }, false, 'login/start');
        try {
          const response = await fetch('/api/v1/auth/login', {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
          });
          if (!response.ok) {
            const message = (await response.text()) || 'Authentication failed';
            set({ loginError: message }, false, 'login/error');
            return { status: 'error', message };
          }
          const data = (await response.json()) as unknown;
          if (
            typeof data === 'object' &&
            data !== null &&
            (data as { mfaRequired?: unknown }).mfaRequired === true
          ) {
            const mfa = data as { mfaToken: string; factor: string };
            set({ mfaPending: { mfaToken: mfa.mfaToken, factor: mfa.factor } }, false, 'login/mfa');
            return { status: 'mfa', mfaToken: mfa.mfaToken, factor: mfa.factor };
          }
          if (!isValidAuthResponse(data)) {
            set({ loginError: 'Authentication failed' }, false, 'login/invalid');
            return { status: 'error', message: 'Authentication failed' };
          }
          writeAuthFlag(true);
          set({ isAuthenticated: true, loginError: null }, false, 'login/ok');
          return { status: 'ok' };
        } catch {
          const message = 'Unable to reach authentication server.';
          set({ loginError: message }, false, 'login/network-error');
          return { status: 'error', message };
        } finally {
          set({ loginLoading: false }, false, 'login/done');
        }
      },

      verifyMfa: async (code): Promise<LoginResult> => {
        const pending = get().mfaPending;
        if (!pending) {
          return { status: 'error', message: 'No MFA challenge in progress.' };
        }
        set({ loginLoading: true, loginError: null }, false, 'verifyMfa/start');
        try {
          const response = await fetch('/api/v1/auth/login/totp', {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ mfaToken: pending.mfaToken, code }),
          });
          if (!response.ok) {
            const message = (await response.text()) || 'Verification failed';
            set({ loginError: message }, false, 'verifyMfa/error');
            return { status: 'error', message };
          }
          const data = (await response.json()) as unknown;
          if (!isValidAuthResponse(data)) {
            set({ loginError: 'Verification failed' }, false, 'verifyMfa/invalid');
            return { status: 'error', message: 'Verification failed' };
          }
          writeAuthFlag(true);
          set({ isAuthenticated: true, mfaPending: null, loginError: null }, false, 'verifyMfa/ok');
          return { status: 'ok' };
        } catch {
          const message = 'Unable to reach verification endpoint.';
          set({ loginError: message }, false, 'verifyMfa/network-error');
          return { status: 'error', message };
        } finally {
          set({ loginLoading: false }, false, 'verifyMfa/done');
        }
      },

      performLogin: async (username, password): Promise<boolean> => {
        try {
          const response = await fetch('/api/v1/auth/login', {
            method: 'POST',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password }),
          });
          if (!response.ok) {
            return false;
          }
          const data = (await response.json()) as unknown;
          if (!isValidAuthResponse(data)) {
            return false;
          }
          writeAuthFlag(true);
          set({ isAuthenticated: true, loginError: null }, false, 'performLogin/ok');
          return true;
        } catch {
          return false;
        }
      },

      logout: async (): Promise<void> => {
        try {
          await fetch('/api/v1/auth/logout', { method: 'POST', credentials: 'include' });
        } catch {
          // Proceed with local teardown regardless — the user is logging out.
        }
        tearDownSession(set, null);
      },

      expireSession: (message = 'Session expired. Please sign in again.'): void => {
        tearDownSession(set, message);
      },

      setLoginError: (message) => set({ loginError: message }, false, 'setLoginError'),
      setSetupStatus: (status) => set({ setupStatus: status }, false, 'setSetupStatus'),
      setRecoveryStatus: (status) => set({ recoveryStatus: status }, false, 'setRecoveryStatus'),
      setShowRecoveryForm: (show) => set({ showRecoveryForm: show }, false, 'setShowRecoveryForm'),

      checkStatuses: async (): Promise<void> => {
        try {
          const setupResponse = await fetch('/api/v1/setup/status', {
            method: 'GET',
            credentials: 'include',
          });
          if (setupResponse.ok) {
            set(
              { setupStatus: (await setupResponse.json()) as SetupStatus },
              false,
              'checkStatuses/setup',
            );
          }
          const recoveryResponse = await fetch('/api/v1/recovery/status', {
            method: 'GET',
            credentials: 'include',
          });
          if (recoveryResponse.ok) {
            set(
              { recoveryStatus: (await recoveryResponse.json()) as RecoveryStatus },
              false,
              'checkStatuses/recovery',
            );
          }
        } finally {
          set({ setupChecked: true }, false, 'checkStatuses/done');
        }
      },
    }),
    { name: 'auth-store' },
  ),
);

/**
 * Authenticated fetch — module-level so React Query queryFns (and mutations)
 * can call it without closing over a component callback. Reads/writes auth
 * state through the store. Always sends cookies; never touches tokens.
 *
 * Pass React Query's `signal` through `init` so requests abort on unmount /
 * query cancellation (and on logout's cancelQueries()).
 */
export async function authFetch(input: RequestInfo, init: RequestInit = {}): Promise<Response> {
  const { isAuthenticated, expireSession } = useAuthStore.getState();
  if (!isAuthenticated) {
    throw new Error('Not authenticated');
  }

  const headers = new Headers(init.headers ?? {});
  if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  const send = (): Promise<Response> => fetch(input, { ...init, headers, credentials: 'include' });

  const response = await send();

  if (response.status === 401) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      const retry = await send();
      // Only a retry that is *still* 401 means the session is truly gone; any
      // other status (incl. 5xx/4xx) is the caller's to handle.
      if (retry.status !== 401) {
        return retry;
      }
    }
    expireSession();
    throw new Error('Unauthorized');
  }

  if (response.status === 403) {
    // A still-valid session that simply lacks permission must NOT be logged
    // out. The server sends authorization denials as JSON {code:"PERMISSION_
    // DENIED"}; CSRF/unknown 403s arrive as plain text and are session-invalid.
    let code: string | undefined;
    try {
      code = ((await response.clone().json()) as { code?: string }).code;
    } catch {
      // Non-JSON body (e.g. CSRF plain-text rejection) — treat as invalid.
    }
    if (code === 'PERMISSION_DENIED') {
      return response;
    }
    expireSession('Access forbidden. Please sign in again.');
    throw new Error('Forbidden');
  }

  return response;
}

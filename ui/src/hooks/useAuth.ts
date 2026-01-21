/**
 * @fileoverview Authentication Hook
 * @description Manages authentication state, login/logout, token refresh, and CSRF protection.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { isValidAuthResponse } from '../types/api';
import { logWarn } from '../utils/logger';

const AUTH_FLAG_KEY = 'stem-authenticated';

/** CSRF token header name - must match backend auth.CSRFHeaderName */
const CSRF_HEADER_NAME = 'X-Csrf-Token';

/** HTTP methods that require CSRF token */
const STATE_CHANGING_METHODS: readonly string[] = ['POST', 'PUT', 'DELETE', 'PATCH'];

interface UseAuthResult {
  /** Whether the user is authenticated */
  isAuthenticated: boolean;
  /** Whether a login request is in progress */
  loginLoading: boolean;
  /** Error message from last login attempt */
  loginError: string | null;
  /** Whether connected to the backend */
  connected: boolean;
  /** Handle login form submission */
  handleLogin: (username: string, password: string) => Promise<void>;
  /** Handle logout */
  handleLogout: () => Promise<void>;
  /** Expire the session with an optional message */
  expireSession: (message?: string) => void;
  /** Authenticated fetch wrapper with token refresh and CSRF protection */
  authFetch: (input: RequestInfo, init?: RequestInit) => Promise<Response>;
  /** Set connected state */
  setConnected: (connected: boolean) => void;
  /** Clear login error */
  clearLoginError: () => void;
  /** Stats polling interval ref (for cleanup on session expire) */
  statsIntervalRef: React.MutableRefObject<number | null>;
}

/**
 * Hook for managing authentication state and operations.
 * Handles login, logout, token refresh, CSRF protection, and session expiration.
 */
export function useAuth(): UseAuthResult {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem(AUTH_FLAG_KEY) === 'true';
    }
    return false;
  });
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);
  const [connected, setConnected] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem(AUTH_FLAG_KEY) === 'true';
    }
    return false;
  });
  const statsIntervalRef = useRef<number | null>(null);

  // CSRF token cache - fetched after login, cleared on logout
  const csrfTokenRef = useRef<string | null>(null);

  // Expire session handler
  const expireSession = useCallback((message = 'Session expired. Please sign in again.') => {
    if (statsIntervalRef.current !== null) {
      clearInterval(statsIntervalRef.current);
      statsIntervalRef.current = null;
    }
    csrfTokenRef.current = null;
    setIsAuthenticated(false);
    setConnected(false);
    setLoginError(message);
    window.localStorage.removeItem(AUTH_FLAG_KEY);
  }, []);

  // Fetch CSRF token from backend
  const fetchCsrfToken = useCallback(async (): Promise<string | null> => {
    try {
      const response = await fetch('/api/v1/auth/csrf', {
        method: 'GET',
        credentials: 'include',
      });
      if (response.ok) {
        const data = await (response.json() as Promise<{ token: string }>);
        csrfTokenRef.current = data.token;
        return data.token;
      }
      return null;
    } catch {
      return null;
    }
  }, []);

  // Get current CSRF token, fetching if needed
  const getCsrfToken = useCallback((): Promise<string | null> => {
    if (csrfTokenRef.current) {
      return Promise.resolve(csrfTokenRef.current);
    }
    return fetchCsrfToken();
  }, [fetchCsrfToken]);

  // Token refresh function
  const refreshAccessToken = useCallback(async (): Promise<boolean> => {
    try {
      const response = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      });
      if (response.ok) {
        // Refresh CSRF token after access token refresh
        await fetchCsrfToken();
        return true;
      }
      return false;
    } catch {
      return false;
    }
  }, [fetchCsrfToken]);

  // Helper: Add CSRF token to headers if needed
  const addCsrfTokenIfNeeded = useCallback(
    async (headers: Headers, method: string): Promise<void> => {
      if (!STATE_CHANGING_METHODS.includes(method)) {
        return;
      }
      const token = await getCsrfToken();
      if (token) {
        headers.set(CSRF_HEADER_NAME, token);
      }
    },
    [getCsrfToken],
  );

  // Helper: Make authenticated request
  const makeRequest = useCallback(
    async (input: RequestInfo, init: RequestInit, headers: Headers): Promise<Response> =>
      fetch(input, { ...init, headers, credentials: 'include' }),
    [],
  );

  // Helper: Handle 401 unauthorized - attempt token refresh and retry
  const handle401 = useCallback(
    async (
      input: RequestInfo,
      init: RequestInit,
      headers: Headers,
      method: string,
    ): Promise<Response | null> => {
      const refreshed = await refreshAccessToken();
      if (!refreshed) {
        return null;
      }

      await addCsrfTokenIfNeeded(headers, method);
      const retryResponse = await makeRequest(input, init, headers);
      return retryResponse.ok ? retryResponse : null;
    },
    [addCsrfTokenIfNeeded, makeRequest, refreshAccessToken],
  );

  // Helper: Handle 403 forbidden - check for CSRF error and retry
  const handle403 = useCallback(
    async (
      response: Response,
      input: RequestInfo,
      init: RequestInit,
      headers: Headers,
    ): Promise<Response | null> => {
      const text = await (response.text() as Promise<string>);
      if (!text.includes('CSRF')) {
        return null;
      }

      csrfTokenRef.current = null;
      const newToken = await getCsrfToken();
      if (!newToken) {
        return null;
      }

      headers.set(CSRF_HEADER_NAME, newToken);
      const retryResponse = await makeRequest(input, init, headers);
      return retryResponse.ok ? retryResponse : null;
    },
    [getCsrfToken, makeRequest],
  );

  // Helper: Handle response status and retry if needed
  const handleResponseStatus = useCallback(
    async (
      response: Response,
      input: RequestInfo,
      init: RequestInit,
      headers: Headers,
      method: string,
    ): Promise<Response> => {
      if (response.status === 401) {
        const retryResponse = await handle401(input, init, headers, method);
        if (retryResponse) {
          return retryResponse;
        }
        expireSession();
        throw new Error('Unauthorized');
      }

      if (response.status === 403) {
        const retryResponse = await handle403(response, input, init, headers);
        if (retryResponse) {
          return retryResponse;
        }
        expireSession('Access forbidden. Please sign in again.');
        throw new Error('Forbidden');
      }

      return response;
    },
    [expireSession, handle401, handle403],
  );

  // Authenticated fetch wrapper with CSRF protection
  const authFetch = useCallback(
    async (input: RequestInfo, init: RequestInit = {}): Promise<Response> => {
      if (!isAuthenticated) {
        throw new Error('Not authenticated');
      }

      const method = init.method?.toUpperCase() ?? 'GET';
      const headers = new Headers(init.headers || {});
      const needsContentType =
        init.body && !(init.body instanceof FormData) && !headers.has('Content-Type');

      if (needsContentType) {
        headers.set('Content-Type', 'application/json');
      }

      await addCsrfTokenIfNeeded(headers, method);
      const response = await makeRequest(input, init, headers);

      return handleResponseStatus(response, input, init, headers, method);
    },
    [addCsrfTokenIfNeeded, handleResponseStatus, isAuthenticated, makeRequest],
  );

  // Login handler
  const handleLogin = useCallback(
    async (username: string, password: string): Promise<void> => {
      setLoginLoading(true);
      setLoginError(null);
      try {
        const response = await fetch('/api/v1/auth/login', {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password }),
        });
        if (!response.ok) {
          const text = (await (response.text() as Promise<string>)) || 'Authentication failed';
          setLoginError(text);
          setConnected(false);
          return;
        }
        const data: unknown = await (response.json() as Promise<unknown>);
        if (!isValidAuthResponse(data)) {
          setLoginError('Authentication failed');
          setConnected(false);
          return;
        }
        setIsAuthenticated(true);
        window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
        setLoginError(null);
        setConnected(true);

        // Fetch CSRF token immediately after login
        await fetchCsrfToken();
      } catch {
        setLoginError('Unable to reach authentication server.');
        setConnected(false);
      } finally {
        setLoginLoading(false);
      }
    },
    [fetchCsrfToken],
  );

  // Logout handler
  const handleLogout = useCallback(async (): Promise<void> => {
    try {
      await fetch('/api/v1/auth/logout', {
        method: 'POST',
        credentials: 'include',
      });
    } catch (error) {
      logWarn('Logout API call failed', {
        component: 'useAuth',
        action: 'handleLogout',
        additionalData: {
          error: error instanceof Error ? error.message : String(error),
        },
      });
    }
    csrfTokenRef.current = null;
    setIsAuthenticated(false);
    setConnected(false);
    setLoginError(null);
    window.localStorage.removeItem(AUTH_FLAG_KEY);
  }, []);

  // Clear login error
  const clearLoginError = useCallback(() => {
    setLoginError(null);
  }, []);

  // Sync auth flag with localStorage
  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    if (isAuthenticated) {
      window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
    } else {
      window.localStorage.removeItem(AUTH_FLAG_KEY);
    }
  }, [isAuthenticated]);

  // Fetch CSRF token on mount if already authenticated
  useEffect(() => {
    if (isAuthenticated && !csrfTokenRef.current) {
      fetchCsrfToken().catch(() => {
        // Silent fail - CSRF token will be fetched on next request
      });
    }
  }, [fetchCsrfToken, isAuthenticated]);

  return {
    isAuthenticated,
    loginLoading,
    loginError,
    connected,
    handleLogin,
    handleLogout,
    expireSession,
    authFetch,
    setConnected,
    clearLoginError,
    statsIntervalRef,
  };
}

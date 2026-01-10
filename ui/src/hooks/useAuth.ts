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
const STATE_CHANGING_METHODS = ['POST', 'PUT', 'DELETE', 'PATCH'];

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
        const data = (await response.json()) as { token: string };
        csrfTokenRef.current = data.token;
        return data.token;
      }
      return null;
    } catch {
      return null;
    }
  }, []);

  // Get current CSRF token, fetching if needed
  const getCsrfToken = useCallback(async (): Promise<string | null> => {
    if (csrfTokenRef.current) {
      return csrfTokenRef.current;
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

  // Authenticated fetch wrapper with CSRF protection
  const authFetch = useCallback(
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Auth logic with retry and CSRF requires branching
    async (input: RequestInfo, init: RequestInit = {}): Promise<Response> => {
      if (!isAuthenticated) {
        throw new Error('Not authenticated');
      }

      const method = init.method?.toUpperCase() ?? 'GET';
      const headers = new Headers(init.headers || {});

      if (init.body && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json');
      }

      // Add CSRF token for state-changing requests
      if (STATE_CHANGING_METHODS.includes(method)) {
        const csrfToken = await getCsrfToken();
        if (csrfToken) {
          headers.set(CSRF_HEADER_NAME, csrfToken);
        }
      }

      let response = await fetch(input, { ...init, headers, credentials: 'include' });

      if (response.status === 401) {
        const refreshed = await refreshAccessToken();
        if (refreshed) {
          // Retry with new CSRF token
          if (STATE_CHANGING_METHODS.includes(method)) {
            const newCsrfToken = await getCsrfToken();
            if (newCsrfToken) {
              headers.set(CSRF_HEADER_NAME, newCsrfToken);
            }
          }
          response = await fetch(input, { ...init, headers, credentials: 'include' });
          if (response.ok) {
            return response;
          }
        }
        expireSession();
        throw new Error('Unauthorized');
      }

      if (response.status === 403) {
        // Check if it's a CSRF error - try to refresh token and retry
        const text = await response.text();
        if (text.includes('CSRF')) {
          csrfTokenRef.current = null; // Clear cached token
          const newCsrfToken = await getCsrfToken();
          if (newCsrfToken) {
            headers.set(CSRF_HEADER_NAME, newCsrfToken);
            response = await fetch(input, { ...init, headers, credentials: 'include' });
            if (response.ok) {
              return response;
            }
          }
        }
        expireSession('Access forbidden. Please sign in again.');
        throw new Error('Forbidden');
      }
      return response;
    },
    [expireSession, getCsrfToken, isAuthenticated, refreshAccessToken],
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
          const text = (await response.text()) || 'Authentication failed';
          setLoginError(text);
          setConnected(false);
          return;
        }
        const data: unknown = await response.json();
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
        additionalData: { error: error instanceof Error ? error.message : String(error) },
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
    if (typeof window === 'undefined') return;
    if (isAuthenticated) {
      window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
    } else {
      window.localStorage.removeItem(AUTH_FLAG_KEY);
    }
  }, [isAuthenticated]);

  // Fetch CSRF token on mount if already authenticated
  useEffect(() => {
    if (isAuthenticated && !csrfTokenRef.current) {
      void fetchCsrfToken();
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

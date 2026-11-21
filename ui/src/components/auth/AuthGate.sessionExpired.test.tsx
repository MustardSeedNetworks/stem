/**
 * AuthGate.sessionExpired.test.tsx — a dead session lands on the login screen
 * with the reason, not on a bare form.
 *
 * Drives the real path: authFetch gets a 401, the refresh is refused, and the
 * auth-store expires the session. AuthGate must then show the sign-in form
 * carrying the "session expired" message, in the active locale.
 */
import { act, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { authFetch, useAuthStore } from '../../stores/auth-store';
import { AuthGate } from './AuthGate';

beforeEach(() => {
  // Installed before AuthGate mounts: its on-mount status check also uses fetch.
  vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
    if (String(input).includes('/setup/status') || String(input).includes('/recovery/status')) {
      return new Response(JSON.stringify({ needsSetup: false, active: false }), { status: 200 });
    }
    return new Response(null, { status: 401 });
  });
  useAuthStore.setState({
    isAuthenticated: true,
    mfaPending: null,
    loginError: null,
    setupStatus: null,
    setupChecked: true,
    recoveryStatus: null,
  });
});

afterEach(async () => {
  vi.restoreAllMocks();
  await i18n.changeLanguage('en');
});

async function expireViaRefreshFailure(): Promise<void> {
  // The store update re-renders AuthGate, so it has to happen inside act().
  const expired: Promise<void> = act(async () => {
    await expect(authFetch('/api/v1/stats')).rejects.toThrow('Unauthorized');
  });
  await expired;
}

describe('AuthGate after the session expires', () => {
  it('shows the sign-in form with the session-expired message', async () => {
    await i18n.changeLanguage('en');
    render(<AuthGate />);
    expect(screen.queryByTestId('login-username')).not.toBeInTheDocument();

    await expireViaRefreshFailure();

    await waitFor(() => {
      expect(screen.getByTestId('login-username')).toBeInTheDocument();
    });
    expect(screen.getByRole('alert')).toHaveTextContent('Your session expired, sign in again');
  });

  it('shows the message in Spanish under es', async () => {
    await i18n.changeLanguage('es');
    render(<AuthGate />);

    await expireViaRefreshFailure();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Su sesión expiró, inicie sesión de nuevo',
      );
    });
  });
});

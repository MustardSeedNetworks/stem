/**
 * AuthGate.i18n.test.tsx — the login screen renders real locale copy.
 *
 * #759: wrecking every EN locale file failed only 14 of 201 tests, because the
 * suite asserts on testids rather than text. A key can go missing, or ship
 * untranslated, and CI stays green. These tests assert the visible strings in
 * both locales against the real `internal/i18n/locales` JSON.
 *
 * The login surface is first because it is the one screen every user sees, and
 * until #762 it had no `t()` calls at all.
 */
import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { useAuthStore } from '../../stores/auth-store';
import { AuthGate } from './AuthGate';

beforeEach(() => {
  // AuthGate probes /auth/setup-status on mount; without a stub the promise
  // rejects and the login form never settles.
  vi.spyOn(globalThis, 'fetch').mockImplementation((async () => ({
    ok: true,
    status: 200,
    json: async () => ({ needsSetup: false }),
  })) as unknown as typeof fetch);
  useAuthStore.setState({ isAuthenticated: false, mfaPending: null, loginError: null });
});

afterEach(async () => {
  vi.restoreAllMocks();
  await i18n.changeLanguage('en');
});

describe('AuthGate — real locale copy', () => {
  it('renders the English sign-in prompt', async () => {
    await i18n.changeLanguage('en');
    render(<AuthGate />);

    await waitFor(() => {
      expect(screen.getByText('Sign in to continue')).toBeInTheDocument();
    });
    expect(screen.getByText('Authenticate with your Stem credentials.')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    render(<AuthGate />);

    await waitFor(() => {
      expect(screen.getByText('Inicie sesión para continuar')).toBeInTheDocument();
    });
    expect(screen.queryByText('Sign in to continue')).not.toBeInTheDocument();
  });
});

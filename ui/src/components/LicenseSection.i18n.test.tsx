/**
 * LicenseSection.i18n.test.tsx — the licence panel renders real locale copy.
 *
 * #759: wrecking every EN locale file failed only 19 of 206 tests, because the
 * suite asserts on testids and on English hardcoded in components. These assert
 * visible strings in both locales against the real `internal/i18n/locales`
 * JSON, so a key that goes missing or ships untranslated fails here.
 *
 * The error paths are covered too: five of this component's user-visible
 * messages were English string literals passed to setError, which the
 * hardcoded-text detector cannot see — it reads JSX text nodes, not arguments.
 */
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../i18n';
import { invalidateCsrfToken } from '../lib/csrf';
import { useAuthStore } from '../stores/auth-store';
import { LicenseSection } from './LicenseSection';

/**
 * Answers the license route with `body` and the CSRF route with a token.
 *
 * Routing by URL is not tidiness: the panel's POSTs go through `authFetch`,
 * which fetches `/api/v1/auth/csrf-token` first, and a mock that hands the
 * license payload to that call makes `getCsrfToken` throw — every assertion
 * below would then be reading the network-error copy (#1247).
 */
function stubLicense(body: unknown, ok = true) {
  vi.spyOn(globalThis, 'fetch').mockImplementation((async (url: string) => ({
    ok: url === CSRF_ENDPOINT ? true : ok,
    status: url === CSRF_ENDPOINT || ok ? 200 : 500,
    headers: new Headers({ 'Content-Type': 'application/json' }),
    json: async () => (url === CSRF_ENDPOINT ? { token: 'test-csrf-token' } : body),
  })) as unknown as typeof fetch);
}

const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';

beforeEach(() => {
  invalidateCsrfToken();
  // The panel only renders behind a session, and `authFetch` refuses to send
  // anything without one.
  useAuthStore.setState({ isAuthenticated: true });
  stubLicense({ activated: false, tier: 'free', deviceId: 'dev-1', features: [] });
});

afterEach(async () => {
  vi.restoreAllMocks();
  invalidateCsrfToken();
  useAuthStore.setState({ isAuthenticated: false });
  await i18n.changeLanguage('en');
});

describe('LicenseSection — real locale copy', () => {
  it('renders the English panel, title included', async () => {
    await i18n.changeLanguage('en');
    render(<LicenseSection />);

    // 'Activate License' is both the form heading and the button label.
    await waitFor(() => {
      expect(screen.getAllByText('Activate License').length).toBeGreaterThan(0);
    });
    expect(screen.getByText('License')).toBeInTheDocument();
  });

  it('renders Spanish under es, with no English left behind', async () => {
    await i18n.changeLanguage('es');
    render(<LicenseSection />);

    await waitFor(() => {
      expect(screen.getAllByText('Activar Licencia').length).toBeGreaterThan(0);
    });
    expect(screen.getByText('Licencia')).toBeInTheDocument();
    expect(screen.queryByText('Activate License')).toBeNull();
    expect(screen.queryByText('License')).toBeNull();
  });

  it('localizes the activation failure, which is a string literal not JSX', async () => {
    await i18n.changeLanguage('es');
    render(<LicenseSection />);

    await waitFor(() => {
      expect(screen.getAllByText('Activar Licencia').length).toBeGreaterThan(0);
    });

    // The activate button is disabled until a key is present, so the failure
    // path is only reachable with one typed in.
    await userEvent.type(
      screen.getByPlaceholderText('MSN1.<payload>.<signature>'),
      'MSN1.eyJ2IjoxfQ.c2ln',
    );
    stubLicense({ success: false, message: '' });
    const buttons = screen.getAllByRole('button', { name: /Activar Licencia/ });
    const activate = buttons[buttons.length - 1];
    if (activate) {
      await userEvent.click(activate);
    }

    await waitFor(() => {
      expect(screen.getByText('Error en la activacion de licencia')).toBeInTheDocument();
    });
  });
});

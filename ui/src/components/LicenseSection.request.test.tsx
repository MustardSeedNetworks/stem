/**
 * LicenseSection.request.test.tsx — the panel's requests match the daemon.
 *
 * #1247: all three calls were wrong at once — `/api/license*` without `/v1`
 * (the SPA handler answers those with index.html and HTTP 200, so nothing
 * 404s), no `X-Csrf-Token` on the two POSTs (403 from the CSRF manager), and
 * `{key}` where `LicenseActivateRequest` declares `licenseKey` and
 * `decodeJSONStrict` refuses unknown fields (400). Activation could not
 * succeed from the UI, and every failure surfaced as "connection failed".
 *
 * The suite that existed mocked `fetch` wholesale and asserted only rendered
 * copy, which is why three independent breaks shipped together. These assert
 * the wire: exact URL, the CSRF header, and the request body against the Go
 * type — the `check-request-fields.sh` idea applied to the browser side.
 */

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../i18n';
import { invalidateCsrfToken } from '../lib/csrf';
import { useAuthStore } from '../stores/auth-store';
import { LicenseSection } from './LicenseSection';

const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';
const CSRF_TOKEN = 'test-csrf-token';
const STATUS_ENDPOINT = '/api/v1/license';
const ACTIVATE_ENDPOINT = '/api/v1/license/activate';
const TRIAL_ENDPOINT = '/api/v1/license/trial';

const KEY = 'MSN1.eyJ2IjoxfQ.c2ln';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

/** An unlicensed daemon: the panel shows the activation form and trial button. */
const UNLICENSED = {
  activated: false,
  isTrialMode: false,
  tier: 0,
  tierName: '',
  daysRemaining: 0,
  features: [],
  deviceHash: 'abc123',
  message: 'No license. Start a trial or enter a license key.',
};

let fetchMock: ReturnType<typeof vi.fn>;

/**
 * Routes by URL rather than answering everything with one body: the POSTs are
 * preceded by a CSRF token fetch, and a mock that returns the license payload
 * for that call makes `getCsrfToken` throw — which the component would then
 * report as a network error, hiding whatever the test meant to check.
 */
function respondWith(post: Response): void {
  fetchMock.mockImplementation(async (url: string) => {
    if (url === CSRF_ENDPOINT) {
      return jsonResponse({ token: CSRF_TOKEN });
    }
    if (url === STATUS_ENDPOINT) {
      return jsonResponse(UNLICENSED);
    }
    return post;
  });
}

/** The arguments of the request to `url`, whichever call index it landed on. */
function callTo(url: string): [string, RequestInit] {
  const call = fetchMock.mock.calls.find(([called]) => called === url);
  if (call === undefined) {
    throw new Error(
      `no request to ${url}; saw ${JSON.stringify(fetchMock.mock.calls.map(([u]) => u))}`,
    );
  }
  return call as [string, RequestInit];
}

async function typeKeyAndActivate(): Promise<void> {
  await userEvent.type(screen.getByPlaceholderText('MSN1.<payload>.<signature>'), KEY);
  const buttons = screen.getAllByRole('button', { name: /Activate License/ });
  const activate = buttons[buttons.length - 1];
  if (activate === undefined) {
    throw new Error('activate button not rendered');
  }
  await userEvent.click(activate);
}

beforeEach(async () => {
  await i18n.changeLanguage('en');
  invalidateCsrfToken();
  useAuthStore.setState({ isAuthenticated: true });
  fetchMock = vi.fn();
  vi.spyOn(globalThis, 'fetch').mockImplementation(fetchMock as unknown as typeof fetch);
  respondWith(jsonResponse({ success: true, message: 'Licensed: Professional' }));
});

afterEach(() => {
  vi.restoreAllMocks();
  invalidateCsrfToken();
  useAuthStore.setState({ isAuthenticated: false });
});

describe('LicenseSection — requests the daemon actually serves', () => {
  it('reads status from the versioned route', async () => {
    render(<LicenseSection />);

    await waitFor(() => {
      expect(callTo(STATUS_ENDPOINT)).toBeDefined();
    });
  });

  it('activates with the versioned route, the CSRF header and {licenseKey}', async () => {
    render(<LicenseSection />);
    await screen.findByPlaceholderText('MSN1.<payload>.<signature>');

    await typeKeyAndActivate();

    await waitFor(() => {
      expect(callTo(ACTIVATE_ENDPOINT)).toBeDefined();
    });
    const [, init] = callTo(ACTIVATE_ENDPOINT);
    expect(init.method).toBe('POST');
    expect(new Headers(init.headers).get('X-Csrf-Token')).toBe(CSRF_TOKEN);
    expect(JSON.parse(String(init.body))).toEqual({ licenseKey: KEY });
  });

  it('starts a trial with the versioned route and the CSRF header', async () => {
    respondWith(jsonResponse({ success: true, message: 'Trial started', isTrialMode: true }));
    render(<LicenseSection />);
    const trial = await screen.findByRole('button', { name: /Start 14-Day Trial/ });

    await userEvent.click(trial);

    await waitFor(() => {
      expect(callTo(TRIAL_ENDPOINT)).toBeDefined();
    });
    const [, init] = callTo(TRIAL_ENDPOINT);
    expect(init.method).toBe('POST');
    expect(new Headers(init.headers).get('X-Csrf-Token')).toBe(CSRF_TOKEN);
  });

  it('shows the daemon rejection, not "connection failed", on a 400', async () => {
    respondWith(
      jsonResponse(
        { error: 'Bad Request', code: 'INVALID_REQUEST', message: 'Invalid license key format' },
        400,
      ),
    );
    render(<LicenseSection />);
    await screen.findByPlaceholderText('MSN1.<payload>.<signature>');

    await typeKeyAndActivate();

    await waitFor(() => {
      expect(screen.getByText('Invalid license key format')).toBeInTheDocument();
    });
  });
});

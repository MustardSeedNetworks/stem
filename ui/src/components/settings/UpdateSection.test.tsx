/**
 * UpdateSection.test.tsx — the automatic update toggle speaks the daemon's
 * contract: GET on open, POST `{automatic}` with the CSRF header, and the
 * switch shows what the daemon saved rather than what was clicked.
 */

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { invalidateCsrfToken } from '../../lib/csrf';
import { useAuthStore } from '../../stores/auth-store';
import { UPDATE_SETTINGS_ENDPOINT, UpdateSection } from './UpdateSection';

const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';
const CSRF_TOKEN = 'test-csrf-token';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

let fetchMock: ReturnType<typeof vi.fn>;

function respondWith(post: () => Response): void {
  fetchMock.mockImplementation(async (url: string, init?: RequestInit) => {
    if (url === CSRF_ENDPOINT) {
      return jsonResponse({ token: CSRF_TOKEN });
    }
    if (url === UPDATE_SETTINGS_ENDPOINT && init?.method === 'POST') {
      return post();
    }
    return jsonResponse({ automatic: false, source: '/srv/stem-mirror' });
  });
}

function postInit(): RequestInit {
  const call = fetchMock.mock.calls.find(
    ([url, init]) => url === UPDATE_SETTINGS_ENDPOINT && init?.method === 'POST',
  );
  if (call === undefined) {
    throw new Error('no POST to the update settings');
  }
  return call[1] as RequestInit;
}

beforeEach(async () => {
  await i18n.changeLanguage('en');
  invalidateCsrfToken();
  useAuthStore.setState({ isAuthenticated: true });
  fetchMock = vi.fn();
  vi.spyOn(globalThis, 'fetch').mockImplementation(fetchMock as unknown as typeof fetch);
});

afterEach(() => {
  vi.restoreAllMocks();
  invalidateCsrfToken();
  useAuthStore.setState({ isAuthenticated: false });
});

describe('UpdateSection', () => {
  it('shows the saved opt-in and its source', async () => {
    respondWith(() => jsonResponse({}));
    render(<UpdateSection />);

    const toggle = await screen.findByRole('switch', { name: 'Automatic updates' });
    expect(toggle).toHaveAttribute('aria-checked', 'false');
    expect(screen.getByTestId('settings-update-source')).toHaveTextContent('/srv/stem-mirror');
  });

  it('turns it on with {automatic: true} and the CSRF header', async () => {
    respondWith(() => jsonResponse({ automatic: true, source: '/srv/stem-mirror' }));
    render(<UpdateSection />);
    const toggle = await screen.findByRole('switch', { name: 'Automatic updates' });

    await userEvent.click(toggle);

    await waitFor(() => {
      expect(toggle).toHaveAttribute('aria-checked', 'true');
    });
    const init = postInit();
    expect(new Headers(init.headers).get('X-Csrf-Token')).toBe(CSRF_TOKEN);
    expect(JSON.parse(String(init.body))).toEqual({ automatic: true });
  });

  it('keeps the saved state and says so when the daemon refuses', async () => {
    respondWith(() => jsonResponse({ error: 'Internal Server Error' }, 500));
    render(<UpdateSection />);
    const toggle = await screen.findByRole('switch', { name: 'Automatic updates' });

    await userEvent.click(toggle);

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not save the automatic update setting.',
    );
    expect(toggle).toHaveAttribute('aria-checked', 'false');
  });
});

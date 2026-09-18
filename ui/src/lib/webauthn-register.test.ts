import { afterEach, describe, expect, it, vi } from 'vitest';
import { useAuthStore } from '../stores/auth-store';
import { registerPasskey } from './webauthn-register';

const challenge = 'AQID';

afterEach(() => {
  vi.unstubAllGlobals();
  useAuthStore.setState({ isAuthenticated: false });
});

describe('passkey registration', () => {
  it('runs the browser ceremony with decoded options and posts serialized output', async () => {
    class MockPublicKeyCredential {}
    const credential = Object.assign(new MockPublicKeyCredential(), {
      id: 'credential',
      rawId: new Uint8Array([1]).buffer,
      type: 'public-key',
      response: {
        clientDataJSON: new Uint8Array([2]).buffer,
        attestationObject: new Uint8Array([3]).buffer,
        getTransports: () => [],
      },
    });
    const create = vi.fn().mockResolvedValue(credential);
    vi.stubGlobal('PublicKeyCredential', MockPublicKeyCredential);
    vi.stubGlobal('navigator', { credentials: { create } });
    const begin = vi.fn().mockResolvedValue({
      publicKey: {
        rp: { name: 'Stem' },
        user: { id: 'AQ', name: 'admin', displayName: 'Administrator' },
        challenge,
        pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
      },
    });
    const finish = vi.fn().mockResolvedValue({ success: true, credentialId: 'credential' });

    await registerPasskey({ begin, finish });

    expect(create).toHaveBeenCalledWith({
      publicKey: expect.objectContaining({ challenge: expect.any(ArrayBuffer) }),
    });
    expect(finish).toHaveBeenCalledWith(
      expect.objectContaining({
        rawId: 'AQ',
        response: expect.objectContaining({ clientDataJSON: 'Ag' }),
      }),
    );
  });

  it('fails before contacting the daemon when WebAuthn is unavailable', async () => {
    vi.stubGlobal('PublicKeyCredential', undefined);
    const begin = vi.fn();

    await expect(registerPasskey({ begin, finish: vi.fn() })).rejects.toThrow(
      'This browser cannot create a passkey.',
    );
    expect(begin).not.toHaveBeenCalled();
  });
});

describe('passkey registration over the wire', () => {
  /** Stubs the browser side so only the two daemon calls are under test. */
  function stubCredentialCeremony(): void {
    class MockPublicKeyCredential {}
    const credential = Object.assign(new MockPublicKeyCredential(), {
      id: 'credential',
      rawId: new Uint8Array([1]).buffer,
      type: 'public-key',
      response: {
        clientDataJSON: new Uint8Array([2]).buffer,
        attestationObject: new Uint8Array([3]).buffer,
        getTransports: () => [] as AuthenticatorTransport[],
      },
    });
    vi.stubGlobal('PublicKeyCredential', MockPublicKeyCredential);
    vi.stubGlobal('navigator', { credentials: { create: vi.fn().mockResolvedValue(credential) } });
  }

  it('refreshes and retries when the access token has expired', async () => {
    // The same defect as the role switch (#1318): enrolling a passkey posts to
    // /auth/webauthn/register/{begin,finish}, both registered `auth: true`, and
    // fetchWithCsrf has no 401 branch — so an expired access token surfaced to
    // the operator as a failed enrolment instead of refreshing and succeeding.
    useAuthStore.setState({ isAuthenticated: true });
    stubCredentialCeremony();
    let accessTokenValid = false;
    const csrfTokens = new Map<string, string | null>();
    const fetchMock = vi.fn(async (input: RequestInfo, init: RequestInit = {}) => {
      const url = String(input);
      if (url.includes('/auth/csrf-token')) {
        return new Response(JSON.stringify({ token: 'csrf' }), { status: 200 });
      }
      if (url.includes('/auth/refresh')) {
        accessTokenValid = true;
        return new Response(null, { status: 200 });
      }
      csrfTokens.set(url, new Headers(init.headers ?? {}).get('X-Csrf-Token'));
      if (!accessTokenValid) {
        return new Response(null, { status: 401 });
      }
      if (url.endsWith('/register/begin')) {
        return new Response(
          JSON.stringify({
            publicKey: {
              rp: { name: 'Stem' },
              user: { id: 'AQ', name: 'admin', displayName: 'Administrator' },
              challenge,
              pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
            },
          }),
          { status: 200 },
        );
      }
      return new Response(JSON.stringify({ success: true }), { status: 200 });
    });
    vi.stubGlobal('fetch', fetchMock);

    await expect(registerPasskey()).resolves.toBeUndefined();

    const urls = fetchMock.mock.calls.map(([input]) => String(input));
    expect(urls).toContain('/api/v1/auth/refresh');
    expect(urls.filter((url) => url.endsWith('/register/finish'))).toHaveLength(1);
    // Both posts, not just the one the 401 lands on: a call left on bare fetch
    // carries no CSRF header and the daemon answers it 403.
    expect([...csrfTokens]).toEqual([
      ['/api/v1/auth/webauthn/register/begin', 'csrf'],
      ['/api/v1/auth/webauthn/register/finish', 'csrf'],
    ]);
  });
});

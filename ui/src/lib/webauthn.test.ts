import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  decodeCreationOptions,
  decodeRequestOptions,
  loginWithPasskey,
  registerPasskey,
  serializeAuthenticationCredential,
  serializeRegistrationCredential,
} from './webauthn';

const challenge = 'AQID';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('WebAuthn wire conversion', () => {
  it('decodes creation challenge, user id, and excluded credential ids', () => {
    const options = decodeCreationOptions({
      publicKey: {
        rp: { name: 'Stem' },
        user: { id: 'BAUG', name: 'admin', displayName: 'Administrator' },
        challenge,
        pubKeyCredParams: [{ type: 'public-key', alg: -7 }],
        excludeCredentials: [{ type: 'public-key', id: 'BwgJ' }],
      },
    });

    expect([...new Uint8Array(options.challenge)]).toEqual([1, 2, 3]);
    expect([...new Uint8Array(options.user.id)]).toEqual([4, 5, 6]);
    expect([...new Uint8Array(options.excludeCredentials?.[0]?.id ?? new ArrayBuffer(0))]).toEqual([
      7, 8, 9,
    ]);
  });

  it('decodes assertion challenge and allowed credential ids', () => {
    const options = decodeRequestOptions({
      publicKey: {
        challenge,
        allowCredentials: [{ type: 'public-key', id: 'BAUG' }],
      },
    });

    expect([...new Uint8Array(options.challenge)]).toEqual([1, 2, 3]);
    expect([...new Uint8Array(options.allowCredentials?.[0]?.id ?? new ArrayBuffer(0))]).toEqual([
      4, 5, 6,
    ]);
  });

  it('serializes the native registration credential for the daemon', () => {
    const credential = {
      id: 'credential',
      rawId: new Uint8Array([1, 2, 3]).buffer,
      type: 'public-key',
      response: {
        clientDataJSON: new Uint8Array([4, 5]).buffer,
        attestationObject: new Uint8Array([6, 7]).buffer,
        getTransports: () => ['internal'] as AuthenticatorTransport[],
      },
    } as PublicKeyCredential;

    expect(serializeRegistrationCredential(credential)).toEqual({
      id: 'credential',
      rawId: 'AQID',
      type: 'public-key',
      response: {
        clientDataJSON: 'BAU',
        attestationObject: 'Bgc',
        transports: ['internal'],
      },
    });
  });

  it('serializes the native assertion credential for the daemon', () => {
    const credential = {
      id: 'credential',
      rawId: new Uint8Array([1, 2, 3]).buffer,
      type: 'public-key',
      response: {
        clientDataJSON: new Uint8Array([4]).buffer,
        authenticatorData: new Uint8Array([5]).buffer,
        signature: new Uint8Array([6]).buffer,
        userHandle: new Uint8Array([7]).buffer,
      },
    } as PublicKeyCredential;

    expect(serializeAuthenticationCredential(credential)).toEqual({
      id: 'credential',
      rawId: 'AQID',
      type: 'public-key',
      response: {
        clientDataJSON: 'BA',
        authenticatorData: 'BQ',
        signature: 'Bg',
        userHandle: 'Bw',
      },
    });
  });
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

describe('passkey sign-in', () => {
  it('runs the assertion ceremony and posts the serialized credential', async () => {
    class MockPublicKeyCredential {}
    const credential = Object.assign(new MockPublicKeyCredential(), {
      id: 'credential',
      rawId: new Uint8Array([1]).buffer,
      type: 'public-key',
      response: {
        clientDataJSON: new Uint8Array([2]).buffer,
        authenticatorData: new Uint8Array([3]).buffer,
        signature: new Uint8Array([4]).buffer,
        userHandle: null,
      },
    });
    const get = vi.fn().mockResolvedValue(credential);
    vi.stubGlobal('PublicKeyCredential', MockPublicKeyCredential);
    vi.stubGlobal('navigator', { credentials: { get } });
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ publicKey: { challenge } }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ token: 'token', expiresAt: 1 }), { status: 200 }),
      );
    vi.stubGlobal('fetch', fetchMock);

    await expect(loginWithPasskey()).resolves.toEqual({ token: 'token', expiresAt: 1 });

    expect(get).toHaveBeenCalledWith({
      publicKey: expect.objectContaining({ challenge: expect.any(ArrayBuffer) }),
    });
    expect(fetchMock.mock.calls[0]?.[1]?.signal).toBeInstanceOf(AbortSignal);
    expect(JSON.parse(String(fetchMock.mock.calls[1]?.[1]?.body))).toEqual({
      id: 'credential',
      rawId: 'AQ',
      type: 'public-key',
      response: {
        clientDataJSON: 'Ag',
        authenticatorData: 'Aw',
        signature: 'BA',
        userHandle: null,
      },
    });
  });
});

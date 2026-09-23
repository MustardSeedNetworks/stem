import { deadlineExpired, requestDeadline } from '../utils/http';
import type {
  AuthenticationResponse,
  PasskeyAuthResponse,
  RegistrationResponse,
  WireCreationOptions,
  WireCredentialDescriptor,
  WireRequestOptions,
} from './webauthn-wire';

export type { WireCreationOptions, WireRequestOptions } from './webauthn-wire';

const API_BASE = '/api/v1/auth/webauthn';

/**
 * The transport a passkey post runs on — narrow enough that both `fetch` and
 * `authFetch` satisfy it, which `typeof fetch` does not (authFetch takes a
 * `RequestInfo`, not a `URL`).
 */
export type PasskeyTransport = (input: string, init: RequestInit) => Promise<Response>;

function base64UrlToBuffer(value: string): ArrayBuffer {
  const base64 = value.replace(/-/g, '+').replace(/_/g, '/');
  const binary = atob(base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '='));
  return Uint8Array.from(binary, (character) => character.charCodeAt(0)).buffer;
}

function bufferToBase64Url(buffer: ArrayBuffer): string {
  const binary = String.fromCharCode(...new Uint8Array(buffer));
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function isPasskeySupported(): boolean {
  return (
    typeof PublicKeyCredential !== 'undefined' &&
    typeof navigator !== 'undefined' &&
    navigator.credentials !== undefined
  );
}

function decodeDescriptors(
  descriptors: WireCredentialDescriptor[] | undefined,
): PublicKeyCredentialDescriptor[] | undefined {
  return descriptors?.map(({ id, type, transports }) => ({
    id: base64UrlToBuffer(id),
    type,
    transports,
  }));
}

export function decodeCreationOptions(
  options: WireCreationOptions,
): PublicKeyCredentialCreationOptions {
  const { user, challenge, excludeCredentials, ...rest } = options.publicKey;
  return {
    ...rest,
    challenge: base64UrlToBuffer(challenge),
    user: { ...user, id: base64UrlToBuffer(user.id) },
    excludeCredentials: decodeDescriptors(excludeCredentials),
  };
}

export function decodeRequestOptions(
  options: WireRequestOptions,
): PublicKeyCredentialRequestOptions {
  const { challenge, allowCredentials, ...rest } = options.publicKey;
  return {
    ...rest,
    challenge: base64UrlToBuffer(challenge),
    allowCredentials: decodeDescriptors(allowCredentials),
  };
}

export function serializeRegistrationCredential(
  credential: PublicKeyCredential,
): RegistrationResponse {
  const response = credential.response as AuthenticatorAttestationResponse;
  return {
    id: credential.id,
    rawId: bufferToBase64Url(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64Url(response.clientDataJSON),
      attestationObject: bufferToBase64Url(response.attestationObject),
      transports: response.getTransports(),
    },
  };
}

export function serializeAuthenticationCredential(
  credential: PublicKeyCredential,
): AuthenticationResponse {
  const response = credential.response as AuthenticatorAssertionResponse;
  return {
    id: credential.id,
    rawId: bufferToBase64Url(credential.rawId),
    type: credential.type,
    response: {
      clientDataJSON: bufferToBase64Url(response.clientDataJSON),
      authenticatorData: bufferToBase64Url(response.authenticatorData),
      signature: bufferToBase64Url(response.signature),
      userHandle: response.userHandle ? bufferToBase64Url(response.userHandle) : null,
    },
  };
}

/**
 * Posts to one passkey endpoint through `request`, which is the caller's choice
 * of transport: bare `fetch` for the pre-session sign-in endpoints here, and
 * `authFetch` for the `auth: true` enrolment endpoints in `webauthn-register`
 * — which is also why that flow lives in its own module, so this one need not
 * import the auth store the store itself imports.
 */
async function sendJSON(path: string, body: unknown, request: PasskeyTransport): Promise<Response> {
  const deadline = requestDeadline();
  try {
    return await request(API_BASE + path, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: deadline,
    });
  } catch (error) {
    if (deadlineExpired(deadline)) {
      throw new Error('The passkey endpoint did not respond. Check the connection and try again.');
    }
    throw error;
  }
}

export async function postPasskeyJSON<T>(
  path: string,
  body: unknown,
  request: PasskeyTransport,
): Promise<T> {
  const response = await sendJSON(path, body, request);
  if (!response.ok) {
    throw new Error((await response.text()) || `Passkey request failed (HTTP ${response.status})`);
  }
  return (await response.json()) as T;
}

export async function loginWithPasskey(): Promise<PasskeyAuthResponse> {
  if (!isPasskeySupported()) {
    throw new Error('This browser cannot use a passkey.');
  }
  const options = await postPasskeyJSON<WireRequestOptions>('/login/begin', {}, fetch);
  const credential = await navigator.credentials.get({
    publicKey: decodeRequestOptions(options),
  });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error('Passkey sign-in was cancelled.');
  }
  return postPasskeyJSON<PasskeyAuthResponse>(
    '/login/finish',
    serializeAuthenticationCredential(credential),
    fetch,
  );
}

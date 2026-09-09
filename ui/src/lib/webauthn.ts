import { deadlineExpired, requestDeadline } from '../utils/http';
import { fetchWithCsrf } from './csrf';
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

interface RegistrationAPI {
  begin: () => Promise<WireCreationOptions>;
  finish: (credential: RegistrationResponse) => Promise<unknown>;
}

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

async function sendJSON(path: string, body: unknown, csrf: boolean): Promise<Response> {
  const deadline = requestDeadline();
  try {
    const request = csrf ? fetchWithCsrf : fetch;
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

async function postJSON<T>(path: string, body: unknown, csrf: boolean): Promise<T> {
  const response = await sendJSON(path, body, csrf);
  if (!response.ok) {
    throw new Error((await response.text()) || `Passkey request failed (HTTP ${response.status})`);
  }
  return (await response.json()) as T;
}

const registrationAPI: RegistrationAPI = {
  begin: () => postJSON<WireCreationOptions>('/register/begin', {}, true),
  finish: (credential) => postJSON('/register/finish', credential, true),
};

export async function registerPasskey(api: RegistrationAPI = registrationAPI): Promise<void> {
  if (!isPasskeySupported()) {
    throw new Error('This browser cannot create a passkey.');
  }
  const options = await api.begin();
  const credential = await navigator.credentials.create({
    publicKey: decodeCreationOptions(options),
  });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error('Passkey registration was cancelled.');
  }
  await api.finish(serializeRegistrationCredential(credential));
}

export async function loginWithPasskey(): Promise<PasskeyAuthResponse> {
  if (!isPasskeySupported()) {
    throw new Error('This browser cannot use a passkey.');
  }
  const options = await postJSON<WireRequestOptions>('/login/begin', {}, false);
  const credential = await navigator.credentials.get({
    publicKey: decodeRequestOptions(options),
  });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error('Passkey sign-in was cancelled.');
  }
  return postJSON<PasskeyAuthResponse>(
    '/login/finish',
    serializeAuthenticationCredential(credential),
    false,
  );
}

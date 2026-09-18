/**
 * Passkey enrolment — the authenticated half of the WebAuthn flow.
 *
 * `/api/v1/auth/webauthn/register/{begin,finish}` are registered `auth: true`,
 * so both posts go through `authFetch`: it attaches the CSRF header and, on a
 * 401, refreshes the access token and retries. Enrolment previously ran on
 * `fetchWithCsrf`, which has no 401 branch at all, so an operator whose access
 * token had expired was told the passkey could not be created (#1318).
 *
 * It sits apart from `webauthn.ts` because the auth store imports that module
 * for the pre-session sign-in flow; importing the store back from it would
 * close an import cycle.
 */
import { authFetch } from '../stores/auth-store';
import {
  decodeCreationOptions,
  isPasskeySupported,
  postPasskeyJSON,
  serializeRegistrationCredential,
} from './webauthn';
import type { RegistrationResponse, WireCreationOptions } from './webauthn-wire';

interface RegistrationAPI {
  begin: () => Promise<WireCreationOptions>;
  finish: (credential: RegistrationResponse) => Promise<unknown>;
}

const registrationAPI: RegistrationAPI = {
  begin: () => postPasskeyJSON<WireCreationOptions>('/register/begin', {}, authFetch),
  finish: (credential) => postPasskeyJSON('/register/finish', credential, authFetch),
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

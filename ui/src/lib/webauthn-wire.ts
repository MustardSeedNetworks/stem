export interface WireCredentialDescriptor {
  id: string;
  type: PublicKeyCredentialType;
  transports?: AuthenticatorTransport[];
}

export interface WireCreationOptions {
  publicKey: {
    rp: PublicKeyCredentialRpEntity;
    user: Omit<PublicKeyCredentialUserEntity, 'id'> & { id: string };
    challenge: string;
    pubKeyCredParams: PublicKeyCredentialParameters[];
    timeout?: number;
    excludeCredentials?: WireCredentialDescriptor[];
    authenticatorSelection?: AuthenticatorSelectionCriteria;
    attestation?: AttestationConveyancePreference;
    extensions?: AuthenticationExtensionsClientInputs;
  };
}

export interface WireRequestOptions {
  publicKey: {
    challenge: string;
    timeout?: number;
    rpId?: string;
    allowCredentials?: WireCredentialDescriptor[];
    userVerification?: UserVerificationRequirement;
    extensions?: AuthenticationExtensionsClientInputs;
  };
}

export interface RegistrationResponse {
  id: string;
  rawId: string;
  type: string;
  response: {
    clientDataJSON: string;
    attestationObject: string;
    transports: string[];
  };
}

export interface AuthenticationResponse {
  id: string;
  rawId: string;
  type: string;
  response: {
    clientDataJSON: string;
    authenticatorData: string;
    signature: string;
    userHandle: string | null;
  };
}

export interface PasskeyAuthResponse {
  token: string;
  refreshToken?: string;
  expiresAt: number;
}

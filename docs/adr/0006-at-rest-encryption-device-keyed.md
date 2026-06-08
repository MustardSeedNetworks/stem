# ADR 0006: At-rest secret encryption is device-keyed; DEK/JWT separation is N/A

**Status:** Accepted (2026-06-08)

## Context

seed's ADR-0015 establishes a hard rule: the **data-encryption key (DEK)** that
protects at-rest secrets must be separate from the **JWT signing secret**, so
that rotating or leaking one does not compromise the other. The stem/niac
harmonization plan asks whether stem has the same coupling and, if so, to add a
dedicated DEK (a 32-byte CSPRNG key at `<datadir>/credential.key`, 0600, with
HKDF-SHA256) mirroring seed.

An audit of stem's at-rest state found:

- **JWT secret** (`internal/auth`, `STEM_JWT_SECRET`): used **only** to sign and
  verify JWTs inside `internal/auth`. `grep` confirms no use of `jwtSecret` /
  `JWTSecret` anywhere outside that package. When unset it is generated per
  process and is ephemeral.
- **The only encrypted at-rest artifact** is the license activation file
  (`internal/license/activation.go`), written with AES-256-GCM at mode 0600. Its
  key is derived from the **device fingerprint**, not the JWT secret:
  `key = SHA-256(fingerprint.Hash() + "MSN-SEED-2024-LICENSE")` (`deriveKey`).
- Recovery-token files (`internal/auth/recovery_token.go`) store a token value at
  0600; they are not encrypted with — and do not derive from — the JWT secret.

So the coupling ADR-0015 exists to break **does not exist in stem**: no
data-encryption key is derived from, or shared with, the JWT signing secret.

## Decision

No `credential.key` DEK is introduced for stem. The existing separation is
already correct and is recorded here as the deliberate design:

- The JWT signing secret stays confined to `internal/auth` for token signing.
- License at-rest encryption stays **device-fingerprint-keyed**. This is the
  intentional model for an offline, air-gapped-capable product (clinical,
  industrial, government): the key is reproducible from the hardware with no key
  file to provision, manage, or exfiltrate, and no phone-home. Binding the
  ciphertext to the device is a feature (a copied license file will not decrypt
  on another machine), not a gap.

If stem ever persists a secret that must survive a hardware change or be portable
across devices, that secret gets its own CSPRNG DEK per seed ADR-0015, and this
ADR is amended.

## Consequences

- stem reaches DEK/JWT-separation parity with seed by construction; no code
  change is required, and no new key file is added to the install footprint.
- The device-keyed model is documented so a future reader does not "fix" it by
  switching the license cipher to a stored key — that would break air-gapped
  re-provisioning and weaken the copy-resistance property.
- niac's token-file model is the analogous deliberate equivalent (it has no
  password store and no JWT-derived at-rest secret); see the niac ADRs.

## Related issues and PRs

- seed ADR-0015 (DEK separation); the DEK-parity item in the stem/niac
  remediation plan.

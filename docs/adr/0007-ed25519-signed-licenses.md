# ADR 0007: Ed25519-signed license tokens

**Status:** Accepted (2026-06-08)

## Context

Stem's license keys were 16-character strings validated offline by a rotor
("Enigma-style") substitution cipher plus a 2-character polynomial self-checksum
(`internal/license/cipher.go`, `validator.go`). The scheme was **symmetric and
self-describing**: the algorithm, rotor tables, and `GenerateLicenseKey` shipped
inside every Stem binary, so anyone with a copy could mint unlimited valid
Professional keys. seed and niac shipped the same forgeable scheme.

We need offline validation (air-gapped industrial/government customers,
no phone-home) **and** un-forgeability. Asymmetric signatures reconcile the two:
the binary needs only a public key to verify; the secret needed to mint a key
never ships.

This is pre-launch, so there are no issued customer keys to honor — we can
replace the format outright rather than carry a compatibility shim.

## Decision

License keys become **Ed25519-signed tokens**. The format (`signing.go`) is
shared by convention across seed/stem/niac and the keygen tool — each repo owns
its own copy (no master module), matching the harmonization rule:

```
MSN1.<base64url(payload)>.<base64url(signature)>
```

- `MSN1` is the scheme + version tag.
- `payload` is canonical JSON: `{v, product, code, serial, tier, maxDevices,
  iat, exp}`. `product` binds a token to one product; `exp=0` means perpetual.
- `signature` is the 64-byte Ed25519 signature over the exact payload bytes.

`Verifier.Validate` checks the signature **before** interpreting any field, then
enforces scheme, payload version, `product == "stem"`, tier, product-code/tier
pairing (1001→Reflector, 2001→Professional), and expiry. Tier→feature mapping
stays **in-binary**, so a signed token only grants features this build defines.

- The binary embeds only the base64 Ed25519 **public** key
  (`licensePublicKeyB64`). The private key lives solely in the keygen tool.
- `GenerateLicenseKey` and the rotor cipher are **deleted** from the product.
- `Manager` gains a `verifier` field (production key by default) plus
  `NewManagerWithVerifier`, so tests can activate tokens minted with an ephemeral
  key without the production private key. Tests pin the cross-tool contract with
  production-signed vectors for each tier.

## Consequences

- A Stem binary can no longer mint a valid license; forging one requires the
  Ed25519 private key. Forge/tamper rejection is covered by tests.
- Validation stays fully offline — no network, no phone-home.
- Tokens are longer (~200 chars) than the old 16-char key; they are copy/paste
  artifacts, and `FormatKey` no longer strips characters (base64url uses `-`/`_`).
- Valid product codes are 1001 (Reflector) and 2001 (Professional). The wire
  values are preserved across the signature migration.

### Amendment 2026-06-16 — Enterprise tier retired

The `TierEnterprise` SKU (tier 3 / product code 3001) and the deprecated
`TierTestSuite` alias are **removed entirely**, per LICENSE_STRATEGY (Enterprise
is not a SKU; it was folded into Professional) and the pre-v1 no-grandfathering
rule. A token presenting tier 3 or product code 3001 is now **rejected** (it
falls through to "Invalid license tier" / product-code mismatch) — previously
such tokens were grandfathered to Professional features. The keygen tool must no
longer mint 3001 (cross-repo follow-up in msn-internal-tools/keygen).
- The embedded public key is a **pre-launch** key generated for this change; it
  must be rotated via keygen before GA (regenerate the key + the contract vectors
  together).
- seed and niac adopt the identical format so one keygen serves all three.

## Related issues and PRs

- The Ed25519 license item in the stem/niac remediation plan; the parallel niac
  (#802) and seed adoptions and the keygen signing change.

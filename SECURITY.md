# Security Policy

## Supported versions

Until The Stem reaches 1.0, only the **latest released version** receives
security fixes. Older 0.x versions are kept on the repo for reference but
are not patched — upgrade to the current minor.

| Version          | Supported           |
| ---------------- | ------------------- |
| Latest (`main`)  | :white_check_mark:  |
| Older 0.x        | :x:                 |
| Future 1.x       | :white_check_mark:  |

## Reporting a vulnerability

**Please do not open a public issue for a security vulnerability.**

Use one of these private channels:

1. **GitHub Security Advisories (preferred):**
   <https://github.com/krisarmstrong/stem/security/advisories/new>.
   Creates a private advisory visible only to maintainers and you, with
   a built-in audit trail and CVE coordination workflow.
2. **Email:** `kris.armstrong@icloud.com` with subject
   `[STEM SECURITY]`.

Include in your report:

- A description of the vulnerability and the affected component(s).
- Steps to reproduce, ideally with a minimal proof-of-concept.
- The version / commit you tested against.
- The potential impact (e.g. unauthenticated RCE, info disclosure, DoS).
- A suggested fix or mitigation, if you have one.

## What to expect

- **Acknowledgment** within 2 business days.
- **Triage** with a severity assessment within 7 business days.
- **Fix or mitigation** released within the target window for the
  severity tier (see table below). We coordinate disclosure timing
  with you for high-impact issues.
- **Credit** in the resulting GitHub Security Advisory and release
  notes, if you'd like it.

### Severity levels

| Level    | Description                         | Target Resolution |
| -------- | ----------------------------------- | ----------------- |
| Critical | Remote code execution, auth bypass  | 24-48 hours       |
| High     | Data exposure, privilege escalation | 7 days            |
| Medium   | Limited impact vulnerabilities      | 30 days           |
| Low      | Minor issues, hardening             | Next release      |

## Scope

In scope:

- Code in this repository (Go backend, embedded React UI, CI workflows,
  release pipeline).
- Built artifacts published as part of a tagged GitHub release
  (verifiable via the included `cosign` signatures and SBOM).

Out of scope:

- Vulnerabilities in third-party dependencies — please report those
  upstream. We track them via Dependabot and `govulncheck` and patch
  on the next release.
- Denial of service requiring sustained external traffic.
- Social engineering or physical access attacks.
- Self-inflicted misconfigurations (e.g. exposing the daemon to a
  public network without an API token — the daemon explicitly warns
  against this).

## Hardening notes for operators

- Use valid TLS certificates in production. Self-signed certs are
  acceptable for isolated lab networks; `./stem install-ca` trusts the
  generated cert system-wide for dev.
- Set strong API tokens via env (`STEM_API_TOKEN`) when exposing the
  web UI off-loopback. The daemon explicitly warns when bound to a
  non-loopback address without a token.
- Restrict the reflector / traffic-gen modules to controlled test
  networks — they generate high-rate traffic that can saturate
  production links.
- Verify release artifacts with `cosign verify-blob` against the
  `<file>.cosign.bundle` shipped with each release; each archive also
  ships a CycloneDX SBOM.


## Acknowledgments

We appreciate security researchers who help keep The Stem secure.
Contributors are credited in the resulting GitHub Security Advisory /
release notes unless they prefer to remain anonymous.

# ADR 0010: JSON wire-casing convention (camelCase API, no exceptions)

**Status:** Accepted (2026-06-14)

## Context

The Mustard Seed Networks fleet (seed / stem / niac) standardizes on a single
JSON wire-casing rule. Seed codified this rule in its revised ADR-0010
(2026-06-14); stem adopts the same rule per the harmonized-by-convention,
no-master principle. A shared reference is also available at
`msn-docs-internal/05-Engineering/JSON_WIRE_CASING.md`.

Stem's API serves RFC 2544 / Y.1564 / Y.1731 / 2889 / 6349 / MEF / TSN
performance-test results and reflector control over HTTPS. All fields the
API currently emits or accepts are already camelCase:

```
grep -rnoE 'json:"[a-z][a-z0-9]*_[a-z0-9_]+[^"]*"' \
  internal/api internal/reflector --include='*.go' | grep -v _test | wc -l
# → 0
```

This ADR formalizes the rule and adds the CI gate that locks compliance in
for all future changes.

## Decision

1. **Every field stem's API emits or accepts on the wire is camelCase.**
   There are no snake_case exceptions and no allow-list or baseline entries.

2. **snake_case is permitted only off the wire:**
   - SQL column names and migration files.
   - Configuration files read from disk (e.g. `stem.yaml`).
   - Internal adapters that parse output from external tools — for example,
     iperf3's `-json` output is snake_case inside `internal/protocols` and is
     mapped to camelCase before crossing the API boundary. External casing is
     never re-emitted verbatim on our wire.

3. **Source-level casing is unchanged:** Go source files use snake_case for
   identifiers where idiomatic (e.g. multi-word local variables); TypeScript
   identifiers are camelCase. Neither is affected by this ADR.

4. **Enforcement:** `scripts/check-json-casing.sh` runs in CI with an **empty**
   baseline file. Any new `json:"…"` tag that contains an underscore and
   matches the wire-field pattern fails the build immediately, with no
   grandfather exceptions. This gate is introduced in PR #421.

## Consequences

- Stem is already 100% compliant at the time of adoption; no field renames
  are required.
- The CI gate makes non-compliance a build error rather than a review
  finding, eliminating the need for per-PR manual checks.
- Reflector control payloads, test-result frames (RFC 2544 / Y.1564 / Y.1731
  / MEF 45.1 / TSN), and all other wire types share a single consistent
  casing model, making client libraries and cross-fleet tooling simpler to
  write.
- External tool output (iperf3, etc.) that is snake_case on arrival is
  translated at the adapter boundary; this boundary is the single place
  where snake_case → camelCase conversion is allowed.

## Alternatives considered

- **Allow snake_case where the upstream standard uses it** (e.g. match
  iperf3's native field names). Rejected: it creates a mixed API surface
  and forces every consumer to handle two naming conventions.
- **Allow-list specific fields.** Rejected: an allow-list grows over time
  and becomes a maintenance burden; a zero-entry baseline is cleaner and
  clearer.
- **Defer to individual PR review.** Rejected: inconsistent enforcement; a
  CI gate is more reliable and removes reviewer burden.

## Related

- Seed ADR-0010 (JSON wire-casing convention, revised 2026-06-14) — origin
  of this rule
- `msn-docs-internal/05-Engineering/JSON_WIRE_CASING.md` — fleet-wide
  reference
- ADR-0001 (capability registry), ADR-0003 (depguard) — related CI
  enforcement patterns
- PR #421 — introduces `scripts/check-json-casing.sh` with empty baseline

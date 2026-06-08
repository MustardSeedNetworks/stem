# Architecture Decision Records

Durable record of significant architectural decisions for The Stem. Each ADR
captures the Context, the Decision, and its Consequences so the reasoning
survives the people and the diffs. Format mirrors the sibling repos (seed/niac).

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-capability-registry.md) | Capability registry for route policy | Accepted |
| [0002](0002-module-registry-single-source.md) | Single source of truth for module metadata + executor | Accepted |
| [0003](0003-dependency-direction-depguard.md) | Dependency direction enforced by depguard | Accepted |
| [0004](0004-auth-and-cors-posture.md) | Single-admin auth model + CORS posture | Accepted |
| [0005](0005-background-component-lifecycle.md) | Background-component lifecycle holder | Accepted |
| [0006](0006-at-rest-encryption-device-keyed.md) | At-rest encryption device-keyed; DEK/JWT separation N/A | Accepted |
| [0007](0007-ed25519-signed-licenses.md) | Ed25519-signed license tokens | Accepted |
| [0008](0008-dataplane-parser-memory-safety.md) | Memory-safety gate for the C dataplane packet parser | Accepted |

Status values: Proposed · Accepted · Amended · Superseded.

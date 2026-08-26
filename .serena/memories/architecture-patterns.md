# Architecture Patterns - The Stem

## Module System

Modules own workflows and delegate to subsystems:
- **Benchmark** (Red) - RFC 2544 tests
- **ServiceTest** (Orange) - Y.1564/MEF tests
- **TrafficGen** (Yellow) - Custom traffic generation
- **Measure** (Blue) - Y.1731 OAM
- **Certify** (Green) - RFC 2889/6349/TSN

## Key Subsystems

| Subsystem | Path | Purpose |
|-----------|------|---------|
| testmaster | `internal/testmaster/` | Test execution engine |
| reflector | `internal/reflector/` | Packet reflection |
| server | `internal/server/` | REST API handlers |
| license | `internal/license/` | License management |

## Frontend Patterns

- **i18n**: Use `useTranslation('namespace')` - namespaces in `ui/locales/`
- **Theme**: Import tokens from `ui/src/styles/theme.ts`
- **State**: Zustand stores in `ui/src/stores/`

## API Conventions

- Base path: `/api/v1/`
- Auth endpoints: `/api/v1/auth/login`, `/api/v1/auth/logout`, `/api/v1/auth/csrf`
- CSRF header: `X-Csrf-Token`

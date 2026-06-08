# Architecture Overview

## Product

**The Stem** — Network Performance Testing Platform.

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26.4 |
| Frontend | React 19, TypeScript 6 |
| Styling | Tailwind CSS v4 |
| Data Plane | C23 (DPDK / AF_PACKET / AF_XDP), Linux |
| Testing | Vitest, Playwright, Go test |
| Linting | golangci-lint (golden config), Biome, clang-tidy |

## High-Level Architecture

```
┌───────────────────────────────────────────────────────────────────────┐
│  MODULE LAYER  (internal/services)                                     │
│  Benchmark · ServiceTest · TrafficGen · Measure · Certify · Reflector  │
│  Each module owns metadata + (for executable modules) an executor       │
│  factory, registered together in services.buildDefaultRegistry.         │
│  The API layer resolves executors via services.Factory (ADR-0002).      │
└──────────────────────────────┬────────────────────────────────────────┘
                               │ (dependencies point inward; enforced by
                               │  depguard — ADR-0003)
┌──────────────────────────────┴────────────────────────────────────────┐
│  SUBSYSTEMS                                                            │
│  orchestrator (test execution + dataplane) · reflector (packet         │
│  reflection) · api (HTTPS REST + embedded UI) · auth · license ·       │
│  database · metrics                                                    │
└───────────────────────────────────────────────────────────────────────┘
```

The web layer registers every route through a capability registry
(`internal/api/route.go`, ADR-0001): routes are declared as data and a single
`register()` composes auth + rate limiting in one canonical order, so a route
cannot ship without its policy. `GET /__capabilities` exposes the manifest.

## Directory Structure

```
stem/
├── cmd/
│   ├── stem/                       # CLI entry point
│   └── stem-schema/                # code-first DTO -> JSON schema generator
├── internal/
│   ├── api/                        # HTTPS REST API + embedded UI + capability registry
│   ├── auth/                       # authentication, sessions, CSRF, MFA
│   ├── database/                   # persistence
│   ├── license/                    # offline license validation
│   ├── reflector/                  # packet reflector subsystem
│   ├── services/                   # MODULE LAYER
│   │   ├── benchmark/  servicetest/  trafficgen/  measure/  certify/  reflector/
│   │   ├── modtypes/               # shared executor types (cycle-free)
│   │   └── orchestrator/           # test execution engine
│   │       ├── dataplane/          # Go bindings to the C dataplane (cgo on Linux)
│   │       ├── config/             # test config
│   │       └── tui/                # terminal UI
│   ├── netif/  metrics/  logging/  version/  truststore/  oauth/  backup/  …
│   └── web/                        # embedded UI build output (web/dist; no Go source)
├── src/                            # C source (C23): dataplane/ + reflector/
├── include/                        # C headers
├── ui/                             # React frontend source
└── docs/adr/                       # Architecture Decision Records
```

## Modules

| Module | Standard | Purpose |
|--------|----------|---------|
| Benchmark | RFC 2544 | Throughput, latency, frame loss |
| ServiceTest | Y.1564 / MEF | Service activation testing |
| TrafficGen | Custom | Traffic generation |
| Measure | Y.1731 | OAM measurements |
| Certify | RFC 2889 / 6349 / TSN | Compliance certification |
| Reflector | Loopback | Packet reflection (distinct lifecycle) |

## API

HTTPS-only REST API on port **8444** (TLS 1.2+, self-signed for local dev; the
HTTP→HTTPS redirector that previously bound 8043 was removed — there is no
plaintext listener). `GET /__version` (build metadata) and `GET /__capabilities`
(route-policy manifest) are unauthenticated introspection endpoints.

## Decisions

Significant architectural decisions are recorded in [`docs/adr/`](adr/README.md).

# Code Landmarks

Quick reference for important code locations. Update when discovering key areas.

---

## Entry Points

| Location | Purpose |
|----------|---------|
| `cmd/stem/main.go` | CLI entry point |
| `internal/api/` | HTTP handlers |
| `ui/src/main.tsx` | Frontend entry |

## Core Modules

| Module | Path | Purpose |
|--------|------|---------|
| Benchmark | `internal/modules/benchmark/` | RFC 2544 tests |
| ServiceTest | `internal/modules/servicetest/` | Y.1564/MEF tests |
| TrafficGen | `internal/modules/trafficgen/` | Traffic generation |
| Measure | `internal/modules/measure/` | Y.1731 OAM |
| Certify | `internal/modules/certify/` | RFC 2889/6349/TSN |
| Reflector | `internal/reflector/` | Packet reflection |

## Key Interfaces

| Interface | Location | Implementers |
|-----------|----------|--------------|
| Module | `internal/modules/module.go` | All test modules |

## Configuration

| File | Purpose |
|------|---------|
| `configs/` | Configuration files |
| `.golangci.yml` | Go linter config |
| `ui/biome.json` | Frontend linter |

## Patterns to Follow

- Module registration: See `internal/modules/registry.go`
- API handlers: See existing handlers in `internal/api/`
- Frontend components: Check `ui/src/components/`

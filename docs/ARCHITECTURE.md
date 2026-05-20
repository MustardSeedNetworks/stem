# Architecture Overview

## Product

**The Stem** - Network Performance Testing Platform

## Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.25.5 |
| Frontend | React 19, TypeScript 5.9 |
| Styling | Tailwind CSS v4 |
| Data Plane | C23 (DPDK/AF_PACKET/AF_XDP) |
| Testing | Vitest, Playwright |
| Linting | golangci-lint, Biome, clang-tidy |

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         MODULE LAYER                                     │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐ │
│  │ Benchmark │ │ServiceTest│ │ TrafficGen│ │  Measure  │ │  Certify  │ │
│  │  (Red)    │ │ (Orange)  │ │ (Yellow)  │ │  (Blue)   │ │  (Green)  │ │
│  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ │
└────────┼─────────────┼─────────────┼─────────────┼─────────────┼───────┘
         │             │             │             │             │
┌────────┴─────────────┴─────────────┴─────────────┴─────────────┴───────┐
│                     SUBSYSTEM LAYER                                     │
│           testmaster  │  reflector  │  web  │  license                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
stem/
├── cmd/stem/              # CLI entry point
├── internal/
│   ├── modules/           # Module layer
│   ├── reflector/         # Packet reflector
│   ├── testmaster/        # Test execution
│   └── api/               # REST API
├── src/                   # C source (C23)
├── include/               # C headers
├── ui/                    # React frontend
├── docs/                  # Documentation
└── Makefile
```

## Modules

| Module | Standard | Purpose |
|--------|----------|---------|
| Benchmark | RFC 2544 | Throughput, latency, frame loss |
| ServiceTest | Y.1564/MEF | Service activation testing |
| TrafficGen | Custom | Traffic generation |
| Measure | Y.1731 | OAM measurements |
| Certify | RFC 2889/6349/TSN | Compliance certification |

## API

REST API served on port 8444 (HTTPS, default) with a 308 redirector on
8043. The `--http` flag (or `STEM_HTTP_ONLY=1`) serves plaintext HTTP on
8444 for legacy deployments.

See [API Reference](../../../msn-docs-internal/) for detailed API documentation.

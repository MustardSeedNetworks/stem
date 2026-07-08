# The Stem

> RFC-compliant network performance testing — reflector, traffic generator, and certifier in one binary.

[![CI](https://github.com/MustardSeedNetworks/stem/actions/workflows/ci.yml/badge.svg)](https://github.com/MustardSeedNetworks/stem/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/MustardSeedNetworks/stem?logo=github)](https://github.com/MustardSeedNetworks/stem/releases/latest)
[![CodeQL](https://github.com/MustardSeedNetworks/stem/actions/workflows/codeql.yml/badge.svg)](https://github.com/MustardSeedNetworks/stem/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/MustardSeedNetworks/stem/badge)](https://scorecard.dev/viewer/?uri=github.com/MustardSeedNetworks/stem)
[![Go Reference](https://pkg.go.dev/badge/github.com/MustardSeedNetworks/stem.svg)](https://pkg.go.dev/github.com/MustardSeedNetworks/stem)
[![Go Report Card](https://goreportcard.com/badge/github.com/MustardSeedNetworks/stem)](https://goreportcard.com/report/github.com/MustardSeedNetworks/stem)
[![License: BSL 1.1](https://img.shields.io/badge/License-BSL%201.1-blue.svg)](LICENSE)

The Stem is a network performance testing tool from **Mustard Seed Networks**.
It packages a high-performance reflector and a full suite of RFC-compliant
testing modules into a single Go binary with a CLI, TUI, and React web UI.

Run it as a service-level loopback target, generate traffic against another
endpoint, or drive a full RFC 2544 / Y.1564 certification suite — all from
the same install.

## Features

### Reflector (always available)
- High-performance packet reflection on AF_PACKET, AF_XDP, or DPDK
- Signature detection for NetAlly, RFC 2544/Y.1564 testers, MSN
- Profile presets: NetAlly, MSN, All, Custom
- Filter by signature, OUI, or UDP/TCP port

### Test modules

| Module | Standard | Test Types |
|--------|----------|------------|
| **Benchmark** | RFC 2544 | throughput, latency, frame loss, back-to-back |
| **ServiceTest** | ITU-T Y.1564, MEF 48/49 | config + performance test, full service test |
| **TrafficGen** | custom | scriptable stream generation |
| **Measure** | ITU-T Y.1731 | delay, loss, synthetic loss measurement, loopback (OAM) |
| **Certify** | RFC 2889 / RFC 6349 / IEEE 802.1Qbv | LAN switch certification, TCP throughput, TSN gate-timing |

### Interfaces
- **CLI** — scriptable `stem <cmd>` for CI integration
- **TUI** — single-screen Bubbletea dashboard for ad-hoc use
- **Web UI** — React/TypeScript control plane on port 8444 (HTTPS by default; 8043 plaintext redirector)
- **REST + SSE** — `/api/v1/events` streams live test results

## Quick Start

```bash
# Install (Linux/macOS, requires Go 1.26+)
git clone https://github.com/MustardSeedNetworks/stem
cd stem
make build

# Run as a reflector on eth0
sudo ./bin/stem reflect -i eth0

# Run a throughput test
sudo ./bin/stem test -t throughput -i eth0 --target 192.0.2.10

# List tests by standard, or by module
./bin/stem list-tests
./bin/stem list-tests --by-module

# Start the web UI (HTTPS by default)
sudo ./bin/stem web -p 8444
# → open https://localhost:8444 (self-signed cert)
# → run `sudo ./bin/stem install-ca` once to trust the cert system-wide

# Or the TUI
sudo ./bin/stem tui
```

## Commands

| Command | Purpose |
|---------|---------|
| `stem version` | Show version + build metadata |
| `stem reflect -i <iface>` | Start the reflector |
| `stem test -t <type> -i <iface>` | Run a single test |
| `stem web -p <port>` | Start the web UI + REST API |
| `stem tui` | Launch the TUI dashboard |
| `stem license --status` | Show license tier + activation state |
| `stem list-tests [--by-module]` | Catalogue all supported tests |
| `stem help modules` | Module + test type reference |

Run `stem <cmd> --help` for flags.

## Architecture

```
ui/src/             → React/TypeScript control plane (Vite)
                          ↓ npm run build
internal/api/ui/    → Built assets (embedded via go:embed)
                          ↓
cmd/stem/           → CLI entry point
internal/
├── services/       → Test module implementations
│   ├── reflector/
│   ├── benchmark/  (RFC 2544)
│   ├── servicetest/(Y.1564, MEF)
│   ├── trafficgen/
│   ├── measure/    (Y.1731)
│   ├── certify/    (RFC 2889/6349, TSN)
│   └── orchestrator/  test execution + lifecycle
├── api/            → HTTP/SSE handlers + WebUI embed
├── auth/           → JWT authentication
├── license/        → Tiered licensing
├── netif/          → Interface discovery
├── reflector/      → Reflector kernel-bypass plumbing
└── version/        → Build metadata (injected via ldflags)
src/dataplane/      → C dataplane (C23, Linux-only): DPDK / AF_XDP / AF_PACKET
include/            → C headers
```

The Go binary is pure-Go (`CGO_ENABLED=0`); the C dataplane is built
separately on Linux for kernel-bypass workloads.

## Licensing

Tiered model. License keys are 16-character alphanumeric
(`XXXX-XXXX-XXXX-XXXX`), validated offline, bound to up to three devices
via hardware fingerprint.

| Tier | Features | Key Prefix |
|------|----------|------------|
| Trial | Full features, 14 days | (no key needed) |
| Tier 1 | Reflector only | `1001-*` |
| Tier 2 | Reflector + full test suite | `2001-*` |
| Tier 3 | Enterprise (planned) | `3001-*` |

Start a trial via web UI Settings → License → Start Trial, or:
```bash
stem license trial
```

## Build

| Command | Purpose |
|---------|---------|
| `make build` | Full build (frontend + Go backend; C dataplane on Linux) |
| `make test` | Go tests |
| `make lint` | golangci-lint + Biome + clang-tidy + cppcheck |
| `make lint-go` | Go only |
| `make lint-c` | C only (Linux) |
| `make fmt` | Format all (Go + TS + C) |
| `make packages` | `.deb` + `.rpm` via GoReleaser |
| `make pkg` | macOS `.pkg` |
| `make quick` | Backend-only dev iteration (do **not** ship) |

Verified versions: **Go 1.26.5**, Node.js 26.4.0, golangci-lint v2.12.2,
DPDK 25.11.2 LTS (optional, Linux).

## REST API (selected)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/__version` | GET | Build metadata (no auth) |
| `/api/health` | GET | Server liveness |
| `/api/interfaces` | GET | List network interfaces |
| `/api/modules` | GET | List test modules |
| `/api/modules/{name}` | GET | Module details + supported test types |
| `/api/test/start` | POST | Start a test |
| `/api/test/stop` | POST | Stop the running test |
| `/api/test/result` | GET | Latest result |
| `/api/v1/events` | GET (SSE) | Live test events |
| `/api/auth/login` | POST | Issue JWT |
| `/api/license` | GET | License status |
| `/api/license/activate` | POST | Activate a license key |
| `/api/license/trial` | POST | Start trial |
| `/api/reflector/config` | GET / POST | Reflector configuration |
| `/api/reflector/stats` | GET | Reflector counters |

Most write endpoints require a JWT issued by `/api/auth/login`.

## Versioning & Releases

Conventional commits drive [release-please](https://github.com/googleapis/release-please).
Release tags trigger `release.yml` which cross-builds binaries
(linux/macOS/windows × amd64/arm64), `.deb`, `.rpm`, macOS `.pkg`, and Windows
`.zip` — all signed via cosign keyless OIDC and shipped with SLSA-3
provenance + Syft SBOM.

## License

[Business Source License 1.1](LICENSE). The Stem converts to Apache-2.0 on
the change date stated in the LICENSE file.

## Security

See [SECURITY.md](SECURITY.md) for the vulnerability-disclosure policy.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Related projects

The Stem is the performance-testing tool. Two sibling projects round out
the Mustard Seed Networks toolkit:

- **[seed](https://github.com/MustardSeedNetworks/seed)** — portable network diagnostic appliance
- **[niac-go](https://github.com/MustardSeedNetworks/niac-go)** — network device simulator

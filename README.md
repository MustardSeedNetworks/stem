# Seed Test Suite

**Unified Network Testing Tool by Mustard Seed Networks**

Part of The Seed ecosystem - combining packet reflection and RFC-compliant network testing in a single tool.

## Quick Start

```bash
# Build everything
make

# Run reflector (Tier 1)
./bin/seedtest reflect -i eth0

# Run network tests (Tier 2)
./bin/seedtest test -t throughput -i eth0

# Start WebUI
./bin/seedtest web -p 8080

# Start TUI dashboard
./bin/seedtest tui
```

## Features

### Tier 1: Seed Reflector
- High-performance packet reflection
- ITO/RFC2544/Y.1564/MSN signature detection
- Profile presets (NetAlly, MSN, All, Custom)
- AF_XDP and DPDK support

### Tier 2: Seed Test Suite (includes Reflector)
- **RFC 2544**: Throughput, Latency, Frame Loss, Back-to-Back, System Recovery, Reset
- **ITU-T Y.1564**: Configuration Test, Performance Test, Full Test
- **RFC 2889**: Forwarding, Caching, Learning, Broadcast, Congestion
- **RFC 6349**: TCP Throughput, Path Analysis
- **ITU-T Y.1731**: Delay, Loss, SLM, Loopback (OAM)
- **MEF 48/49**: Service Configuration, Performance Test
- **IEEE 802.1Qbv TSN**: Gate Timing, Traffic Isolation, Scheduled Latency

## Interfaces

| Interface | Command | Description |
|-----------|---------|-------------|
| CLI | `seedtest <command>` | Command-line interface |
| TUI | `seedtest tui` | Terminal dashboard |
| WebUI | `seedtest web -p 8080` | http://localhost:8080 |

## Realtime Updates

The current WebUI uses REST requests for status and results. WebSocket streaming is planned but not implemented yet.

## Build

### Prerequisites
- Go 1.21+
- Node.js 18+
- GCC (for C dataplane)

### Build Commands

```bash
# Build CLI binary
make build

# Build WebUI
cd ui && npm install && npm run build

# Build everything
make

# Run tests
make test
```

## Licensing

Seed Test Suite uses a tiered licensing model:

| Tier | Features | License Key Prefix |
|------|----------|-------------------|
| Tier 1 | Reflector only | `1001-*` |
| Tier 2 | Full test suite + Reflector | `2001-*` |
| Tier 3 | Enterprise (future) | `3001-*` |

### Trial Mode
- 14-day full-featured trial (no license required)
- Start via WebUI Settings > License > Start Trial

### Activation
- License keys are 16-character alphanumeric (XXXX-XXXX-XXXX-XXXX)
- Offline validation (no internet required after activation)
- 3 device activations per license
- Device binding via hardware fingerprint

## Project Structure

```
seed-test-suite/
├── cmd/seedtest/       # CLI entrypoint
├── pkg/
│   ├── license/        # Licensing system
│   ├── interfaces/     # Network interface detection
│   ├── reflector/      # Reflector package
│   ├── testmaster/     # Test suite package
│   └── web/            # Unified web server
├── src/dataplane/      # C dataplane code
├── include/            # C headers
└── ui/                 # React WebUI
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/interfaces` | GET | List network interfaces |
| `/api/stats` | GET | Current statistics |
| `/api/test/start` | POST | Start test |
| `/api/test/stop` | POST | Stop test |
| `/api/license` | GET | License status |
| `/api/license/activate` | POST | Activate license |
| `/api/license/trial` | POST | Start trial |
| `/api/reflector/mode` | GET/POST | Reflector mode |
| `/api/reflector/config` | GET/POST | Reflector config |
| `/api/reflector/stats` | GET | Reflector stats |

## Support

For licensing inquiries and support, contact Mustard Seed Networks.

## License

Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

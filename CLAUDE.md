# The Stem - Claude Code Instructions

**Product:** The Stem - Network Performance Testing
**Version:** v0.1.0
**Repo:** stem
**Last Updated:** 2025-12-30

---

## Documentation Home

All strategic and engineering documentation lives at:
```
/Users/krisarmstrong/Documents/Mustard Seed Networks/
```

### Key References (READ BEFORE CODING)

| Document | Path | Purpose |
|----------|------|---------|
| **Coding Standards** | `05-Engineering/CODING_STANDARDS.md` | Go/TS/C patterns, tooling |
| **API Design** | `05-Engineering/API_DESIGN_GUIDELINES.md` | REST conventions |
| **Error Handling** | `05-Engineering/ERROR_HANDLING.md` | Error patterns, logging |
| **Testing Strategy** | `05-Engineering/TESTING_STRATEGY.md` | Unit, E2E, coverage |
| **Docker Standards** | `05-Engineering/DOCKER_STANDARDS.md` | Container best practices |
| **CI/CD Config** | `05-Engineering/CI_CD_CONFIGURATION.md` | GitHub Actions pipeline |
| **Database Migrations** | `05-Engineering/DATABASE_MIGRATIONS.md` | SQLite migration patterns |
| **Performance** | `05-Engineering/PERFORMANCE_GUIDELINES.md` | Profiling, optimization |
| **Accessibility** | `05-Engineering/ACCESSIBILITY_STANDARDS.md` | WCAG 2.1 AA compliance |

### Product-Specific Docs

| Document | Path |
|----------|------|
| Handoff Guide | `03-The-Stem/THE_STEM_HANDOFF_GUIDE.md` |
| Implementation Spec | `03-The-Stem/THE_STEM_IMPLEMENTATION_SPEC.md` |
| Backend Architecture | `03-The-Stem/THE_STEM_BACKEND_ARCHITECTURE.md` |
| UI Architecture | `03-The-Stem/THE_STEM_UI_ARCHITECTURE.md` |
| API Reference | `03-The-Stem/THE_STEM_API_REFERENCE.md` |
| Data Model | `03-The-Stem/THE_STEM_DATA_MODEL.md` |
| Hardware Specs | `03-The-Stem/THE_STEM_HARDWARE_SPECS.md` |
| DPDK Setup | `03-The-Stem/THE_STEM_DEVELOPMENT_SETUP.md` |

---

## Required Versions (NO EXCEPTIONS)

| Technology | Version | Notes |
|------------|---------|-------|
| **Go** | 1.25.5 | API server, orchestration |
| **C** | C23 | Packet engine (GCC/Clang 7.3.0+) |
| **Node.js** | 25.2.1 | Frontend build |
| **React** | 19 | UI framework |
| **TypeScript** | 5.9.3 | NO JavaScript allowed |
| **DPDK** | 23.11 LTS | Packet processing (when applicable) |

**Check versions before any work:**
```bash
go version      # Must show 1.25.x
gcc --version   # Must show 7.3.0+
node --version  # Must show v25.x
```

---

## Network Environment

### Development Hardware

| Device | Role | Notes |
|--------|------|-------|
| MacBook Air M2 16GB | Primary development | macOS |
| Mac Mini 2018 64GB | Build server, heavy testing | Intel |
| Dell Mini PC 8GB | Deployment target | Linux |

### Common Test Targets

| Target | IP/Host | Purpose |
|--------|---------|---------|
| RFC 2544 DUT | (configure per test) | Device under test |
| Y.1564 endpoint | (configure per test) | Service activation |
| Local loopback | - | Self-test mode |
| Google DNS | 8.8.8.8 | Latency baseline |

### Default Ports

| Service | Port | Protocol |
|---------|------|----------|
| Stem HTTPS | 8443 | TLS 1.2+ |
| Stem HTTP (dev) | 8080 | Development only |
| WebSocket | 8443 | /ws endpoint |

---

## Tooling Stack (2025)

### C
- **clang-format** - Code formatting (config in `.clang-format`)
- **clang-tidy** - Static analysis (config in `.clang-tidy`)
- **Cppcheck** - Additional static analysis

### Go
- **golangci-lint** - Comprehensive linting (config in `.golangci.yml`)
- **gosec** - Security scanning

### TypeScript/Frontend (when added)
- **Biome** - Linting AND formatting (NOT ESLint/Prettier)
- **Playwright** - E2E testing
- **Storybook** - Component development

### Security (All)
- **Semgrep** - SAST scanning
- **Snyk/Trivy** - Dependency scanning
- **TruffleHog/gitleaks** - Secret detection

---

## Hard Rules (NEVER VIOLATE)

### Code Quality
- [ ] **NO JavaScript** - TypeScript only, always
- [ ] **NO `any` type** - Use proper typing or `unknown`
- [ ] **NO ESLint/Prettier** - Use Biome only (when frontend added)
- [ ] **NO console.log in production** - Use structured logging
- [ ] **NO hardcoded secrets** - Use environment variables
- [ ] **C23 standard only** - No legacy C

### C-Specific
- [ ] **Memory safety** - No buffer overflows, use bounds checking
- [ ] **Thread safety** - Proper locking when needed
- [ ] **clang-format on save** - Consistent formatting
- [ ] **clang-tidy clean** - No warnings

### Patterns
- [ ] **Check existing patterns first** - Don't reinvent
- [ ] **Follow module structure** - Benchmark/ServiceTest/TrafficGen/Measure/Certify
- [ ] **Match API conventions** - Check API_DESIGN_GUIDELINES.md

### Testing
- [ ] **Test files treated same as production** - Lint, format, security scan
- [ ] **Standards compliance tests** - RFC 2544, Y.1564

### Git
- [ ] **Conventional commits** - `feat:`, `fix:`, `docs:`, etc.
- [ ] **No secrets in commits** - gitleaks will catch
- [ ] **No force push to main** - Ever

---

## Before Any Code Change

1. **Read the relevant doc** - Don't guess patterns
2. **Check existing code** - Find similar implementations
3. **Run linting** - `make lint`
4. **Run tests** - `make test`

---

## Project Structure

```
stem/
├── cmd/stem/              # Main CLI entry point
├── internal/              # Private Go packages
│   ├── modules/           # Module layer (NEW)
│   │   ├── module.go      # Module interface
│   │   ├── registry.go    # Module registry
│   │   ├── init.go        # Module initialization
│   │   ├── benchmark/     # RFC 2544 module
│   │   ├── servicetest/   # Y.1564/MEF module
│   │   ├── trafficgen/    # Traffic generation module
│   │   ├── measure/       # Y.1731 OAM module
│   │   └── certify/       # RFC 2889/6349/TSN module
│   ├── reflector/         # Packet reflector subsystem
│   ├── testmaster/        # Test execution subsystem
│   ├── license/           # License management
│   ├── help/              # Built-in help system
│   ├── version/           # Version info
│   └── web/               # Web server with module API
├── src/                   # C source files (C23)
│   └── dataplane/         # DPDK/AF_PACKET/AF_XDP backends
├── include/               # C headers
├── ui/                    # React frontend (TypeScript)
│   └── src/components/    # React components
│       └── ModuleSelector.tsx  # Module-aware test selector
├── tests/                 # Test files
├── bin/                   # Built binaries
├── docs/                  # Local docs
├── .clang-format          # C formatting config
├── .clang-tidy            # C static analysis config
├── .golangci.yml          # Go linting config
├── biome.json             # TypeScript linting config
├── go.mod                 # Go modules
└── Makefile               # Build commands
```

---

## Quick Commands

```bash
# Build
make build           # Build all
make clean           # Clean build artifacts

# Test
make test            # Run all tests

# Lint
make lint            # Run all linters (Go + C)
make lint-go         # Go only
make lint-c          # C only (clang-tidy + cppcheck)

# Format
make format          # Format all code
```

---

## Module Architecture

The Stem uses a module-oriented architecture where modules own workflows and delegate to subsystems.

### Module Layer

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

### Modules

| Module | Color | Standard | Test Types |
|--------|-------|----------|------------|
| **Reflector** | #0891b2 (Cyan) | Loopback/Echo | reflect (Tier 1 operational mode) |
| **Benchmark** | #dc2626 (Red) | RFC 2544 | throughput, latency, frame_loss, back_to_back, system_recovery, reset |
| **ServiceTest** | #ea580c (Orange) | ITU-T Y.1564 / MEF | y1564_config, y1564_perf, y1564, mef_config, mef_perf, mef |
| **TrafficGen** | #ca8a04 (Yellow) | Custom Traffic | custom_stream |
| **Measure** | #2563eb (Blue) | ITU-T Y.1731 | y1731_delay, y1731_loss, y1731_slm, y1731_loopback |
| **Certify** | #16a34a (Green) | RFC 2889/6349/TSN | rfc2889_*, rfc6349_*, tsn_* |

### Module Interface

```go
type Module interface {
    Name() string           // "benchmark"
    DisplayName() string    // "Benchmark"
    Description() string    // Human-readable description
    Color() string          // "#dc2626"
    Standard() string       // "RFC 2544"
    TestTypes() []string    // ["throughput", "latency", ...]
    CanRun(testType string) bool
}
```

### Module API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/modules` | GET | List all modules with metadata |
| `/api/modules/:name` | GET | Get specific module info |
| `/api/modules/:name/tests` | GET | List test types for module |

### CLI Module Commands

```bash
stem list-tests              # List tests by standard
stem list-tests --by-module  # List tests grouped by module
stem list-tests --json       # JSON output
stem help modules            # Show tests by module in help
```

---

## Supported Standards

| Standard | Description |
|----------|-------------|
| RFC 2544 | Throughput, Latency, Frame Loss, Back-to-Back |
| ITU-T Y.1564 | Service Activation Testing |
| ITU-T Y.1731 | Ethernet OAM |
| RFC 6349 | TCP Throughput |
| RFC 2889 | LAN Switch Testing |

---

## When Unsure

1. **Check documentation first** - It's comprehensive
2. **Look at existing code** - Find similar patterns
3. **Ask** - Don't guess on architectural decisions
4. **Read the error** - Error messages are detailed

---

## Binary Usage

The binary is named `stem`. Common usage:
```bash
stem version              # Show version
stem reflect -i eth0      # Start reflector
stem test -t throughput   # Run tests
stem web -p 8080          # Start WebUI
stem license --status     # Check license
```

---

**This file is the source of truth for development consistency.**

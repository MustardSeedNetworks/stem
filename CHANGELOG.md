# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.10] - 2026-01-04

### Added

- Document current API with a Target API vNext section.

### Fixed

- Avoid inline error handling in writeJSON to satisfy lint rules.

### Changed

- Allow golangci-lint parallel runners in Makefile.

## [0.1.0] - 2025-12-30

### Added

- Initial project structure
- Module-oriented architecture (Benchmark, ServiceTest, TrafficGen, Measure, Certify)
- Reflector mode (Tier 1)
- RFC 2544 test support (throughput, latency, frame loss, back-to-back)
- ITU-T Y.1564 service activation testing
- CLI interface with `stem` binary
- WebUI with React/TypeScript
- TUI dashboard
- License management system (Tier 1/2/3)
- Go 1.25+ backend
- C23 dataplane with AF_PACKET support
- Biome linting for TypeScript
- golangci-lint for Go

### Infrastructure

- Makefile build system
- Development documentation
- CLAUDE.md for AI-assisted development

---

For detailed commit history, see: https://github.com/krisarmstrong/stem/commits/main

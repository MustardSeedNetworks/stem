# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.12](https://github.com/krisarmstrong/stem/compare/v0.9.11...v0.9.12) (2026-05-18)


### Bug Fixes

* **api:** update fs.Sub subdir to "ui" to match embed glob ([058d44f](https://github.com/krisarmstrong/stem/commit/058d44fdf297cb15b689eb3c5329260b98526460))
* **ci:** auto-trigger release-please on CI completion (was manual-only) ([5334db2](https://github.com/krisarmstrong/stem/commit/5334db21fa76875e2a7ded4a24e14a8a52f31147))
* **ci:** bump Dockerfile go-build to golang:1.26-bookworm ([032a37e](https://github.com/krisarmstrong/stem/commit/032a37e2d50e3d774469132756532ee783eaae38))
* **ci:** correct artifact path + Docker [@locales](https://github.com/locales) copy ([b4902e4](https://github.com/krisarmstrong/stem/commit/b4902e4ac2ae194aa06925c48fab173c33f74804))
* **metrics:** serialize tests that share Prometheus counter labels ([3e413bc](https://github.com/krisarmstrong/stem/commit/3e413bc196564221a31f5a4ced920cc446623e15))

## [0.9.11](https://github.com/krisarmstrong/stem/compare/v0.9.10...v0.9.11) (2026-05-14)


### Bug Fixes

* **build:** expose linux feature APIs for c23 ([ef93e2a](https://github.com/krisarmstrong/stem/commit/ef93e2ad74b7080d8a30e0e334c776bb7e0593d6))
* **ci:** align container and license validation ([655c917](https://github.com/krisarmstrong/stem/commit/655c9171e8194e45c76d2a499a07353c638942e7))
* **ci:** allow gitleaks to inspect pull requests ([cd5728a](https://github.com/krisarmstrong/stem/commit/cd5728a6ccf84af1c460a518186e8df59f1c15cd))
* **ci:** allow MPL npm dependencies ([5b03f31](https://github.com/krisarmstrong/stem/commit/5b03f3139d72c6a18b6dd8efe202221c9c07821f))
* **ci:** build browser test server without cgo ([46d3a3b](https://github.com/krisarmstrong/stem/commit/46d3a3ba31a1bdd77d1fbc434f42f6b9f4767242))
* **ci:** build stem native library with clang ([59f46a0](https://github.com/krisarmstrong/stem/commit/59f46a0fa7d6bef2a24e6f5558b27fd03b2c15ca))
* **ci:** build stem native test dependencies ([dfb6d45](https://github.com/krisarmstrong/stem/commit/dfb6d45d0128dfc2f31aa38347dd4fddeb0e2818))
* **ci:** fetch full history for security scans ([655c135](https://github.com/krisarmstrong/stem/commit/655c135c05b9d7c025cc1138bbd1f3826932acb9))
* **ci:** handle missing dataplane contexts ([8736134](https://github.com/krisarmstrong/stem/commit/8736134b10b1a8a23a23d9b2007bad41ed7dac2f))
* **ci:** keep stem analysis advisory ([74f779e](https://github.com/krisarmstrong/stem/commit/74f779e0de00fa7bd4c2fef92f0bed0cce4347ac))
* **ci:** link native dataplane tests ([b6da226](https://github.com/krisarmstrong/stem/commit/b6da22688638460abb5b2279024cfcf1b00793b8))
* **ci:** repair buildpacks project metadata ([cdcb63f](https://github.com/krisarmstrong/stem/commit/cdcb63f4965cc080cae68daa7b9be0fd7d0033f0))
* **ci:** repair label sync workflow ([7acb464](https://github.com/krisarmstrong/stem/commit/7acb4647a4eb80d138f01a10a5a3b113bebaae40))
* **ci:** report stem analyzer findings ([d726b50](https://github.com/krisarmstrong/stem/commit/d726b501d973ee8fbf1bda2975d9ed13ff7feb48))
* **ci:** resolve stem workflow blockers ([314785d](https://github.com/krisarmstrong/stem/commit/314785d6c3f3a0f763e3758b3ba64fffdddf50c5))
* **ci:** restore stem validation pipeline ([c1a26b2](https://github.com/krisarmstrong/stem/commit/c1a26b20afce1f59e5a0b694d263d62860b1c41f))
* **ci:** run stub unit tests without race ([6272714](https://github.com/krisarmstrong/stem/commit/62727147bada8993d1ce1682e64925c09aee02b6))
* **ci:** run stem intel macos release on current runner ([7f9d234](https://github.com/krisarmstrong/stem/commit/7f9d23427a7a4466b8626f6b6d8ee76179df6f10))
* **ci:** satisfy servicetest lint ([ec275df](https://github.com/krisarmstrong/stem/commit/ec275df79aa63360ee069f492469d13c6633fc70))
* **ci:** scope stem container and license checks ([d267154](https://github.com/krisarmstrong/stem/commit/d2671547ae280830d09777768d5635d58721dfd6))
* **ci:** scope stem e2e smoke suite ([4ce2153](https://github.com/krisarmstrong/stem/commit/4ce2153966bff419ad4fb47f75edbd336db2c9a9))
* **ci:** skip stem docker publish without dockerfile ([a5a9deb](https://github.com/krisarmstrong/stem/commit/a5a9deb1064f7ee462c400b3e3138918940e2a20))
* **ci:** split native compile from unit tests ([f1f8c82](https://github.com/krisarmstrong/stem/commit/f1f8c82c6be3026e969a8917cffc075841eafeba))
* **ci:** stabilize automated validation ([76209fa](https://github.com/krisarmstrong/stem/commit/76209faef490df7baa09d161222ec7fc5da838e8))
* **ci:** stabilize stem browser smoke gate ([7dc7655](https://github.com/krisarmstrong/stem/commit/7dc765542a92fcd6465aa1f483e19aadea440ab1))
* **ci:** start stem web server in browser jobs ([2c9f44b](https://github.com/krisarmstrong/stem/commit/2c9f44b0c29dc60aaf97345a781a9748355defac))
* **ci:** use compatible labeler action ([99c9c57](https://github.com/krisarmstrong/stem/commit/99c9c57eab8ee28c0a69d6a1570046cd6b49c596))
* **ci:** use hosted node setup in container workflow ([9023b15](https://github.com/krisarmstrong/stem/commit/9023b15e74f79c4d929145aaca4dd1067da8b718))
* **ci:** use labeler yaml format ([8d68517](https://github.com/krisarmstrong/stem/commit/8d6851793528dd8862dd6c5bd9fde29866b485b2))
* **security:** scope generated TLS certificate writes ([83f6cef](https://github.com/krisarmstrong/stem/commit/83f6cef51e216a8c2a9b7c6e713fc064541de697))

## [Unreleased]

## [0.1.13] - 2026-01-04

### Changed

- Standardize branding to use "The Stem" in CLI and documentation headings.

## [0.1.12] - 2026-01-04

### Added

- Wire RFC 2889, RFC 6349, Y.1731, MEF, TSN, and custom stream configs into the dataplane wrapper.

### Changed

- Route Measure, TrafficGen, ServiceTest, and Certify executors through the dataplane API.
- Update module status documentation to reflect implemented test execution.

## [0.1.11] - 2026-01-04

### Changed

- Document the SSE-based UI transport for real-time updates.

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

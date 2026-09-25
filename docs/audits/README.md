# Audit snapshots

Point-in-time audit reports, kept unchanged as a record of what was true on the
date each was taken. **None of them describes the product today** — read them as
history, not as current state. Current architecture lives in
[../ARCHITECTURE.md](../ARCHITECTURE.md) and the decision records in
[../adr/](../adr/README.md); the current security posture is the CI gates
(`golangci-lint` with `gosec`, `govulncheck`, Trivy, gitleaks) plus
[../security/AUTH_AUDIT_2026-05-19.md](../security/AUTH_AUDIT_2026-05-19.md).

| Snapshot | Date | Scope |
| ---------- | ------ | ------- |
| [ARCHITECTURE_AUDIT_2025-01-19.md](ARCHITECTURE_AUDIT_2025-01-19.md) | 2025-01-19 | Cross-repo architecture review of Seed, Stem and NIAC |
| [SECURITY_AUDIT_2026-01-06.md](SECURITY_AUDIT_2026-01-06.md) | 2026-01-06 | v0.2.2 static security scan: gosec, govulncheck, OWASP API Top 10 |
| [AUDIT_SUITE_2026-01-26.md](AUDIT_SUITE_2026-01-26.md) | 2026-01-26 | Combined lint, security and quality sweep |
| [GITHUB_ISSUES_2026-01-26.md](GITHUB_ISSUES_2026-01-26.md) | 2026-01-26 | Issue backlog written out of the 2026-01-26 sweep |
| [LINT_BIOME_MAKE_AUDIT_2026-01-26.md](LINT_BIOME_MAKE_AUDIT_2026-01-26.md) | 2026-01-26 | Go lint, Biome and Makefile target review |
| [STM-11_ETHERSCOPE_2026-09-09.md](STM-11_ETHERSCOPE_2026-09-09.md) | 2026-09-09 | v1 plan STM-11: why the EtherScope nXG comparison could not be run |
| [STM-13_CONTAINERS_2026-09-14.md](STM-13_CONTAINERS_2026-09-14.md) | 2026-09-14 | v1 plan STM-13: whether Apple containers substitute for the lab hosts |
| [STM-13_WIRE_2026-09-14.md](STM-13_WIRE_2026-09-14.md) | 2026-09-14 | v1 plan STM-13: #1217 proven fixed on the wire |
| [STM-13_FALLBACK_2026-09-15.md](STM-13_FALLBACK_2026-09-15.md) | 2026-09-15 | v1 plan STM-13: #1232, the test master runs where AF_XDP does not |
| [STM-13_MEASURED_2026-09-15.md](STM-13_MEASURED_2026-09-15.md) | 2026-09-15 | v1 plan STM-13: #1233, a throughput result is a measurement |
| [D-STEM-3_CRASHLOOP_2026-09-15.md](D-STEM-3_CRASHLOOP_2026-09-15.md) | 2026-09-15 | v1 plan D-STEM-3: #1249, a fresh `.deb` crash-loops while the installer reports success |
| [D-STEM-24_RFC2889_MEASURED_2026-09-24.md](D-STEM-24_RFC2889_MEASURED_2026-09-24.md) | 2026-09-24 | v1 plan D-STEM-24: #1242, an RFC 2889 forwarding rate is a measurement |
| [STM-23_UPGRADE_2026-09-25.md](STM-23_UPGRADE_2026-09-25.md) | 2026-09-25 | v1 plan STM-23: in-place upgrades on Ubuntu and Fedora; #1445, an RPM upgrade leaves stem stopped |

New audits produced by the v1 plan land here under the same
`<TOPIC>_<YYYY-MM-DD>.md` convention and get a row in this table.

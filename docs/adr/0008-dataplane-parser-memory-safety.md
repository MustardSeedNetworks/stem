# ADR 0008: Memory-safety gate for the C dataplane packet parser

**Status:** Accepted (2026-06-08)

## Context

The Go services are memory-safe, but the performance dataplane is C
(`src/dataplane/common/packet.c`, plus AF_PACKET/AF_XDP platform layers).
The RFC2544 / Y.1564 frame **validators** (`*_is_valid_response`) and the
extractors built on them are the one place attacker-controlled bytes meet C: a
received frame's length and contents are fully untrusted.

An audit found a latent out-of-bounds read. The validators guarded on
`RFC2544_MIN_FRAME` / `Y1564_MIN_FRAME` = **64** (the Ethernet minimum), but the
24-byte `rfc2544_payload_t` they vouch for sits at frame offset 42 and runs to
offset **66**. So a 64- or 65-byte frame passed validation, and any consumer
that then read the full payload ran 1–2 bytes past the buffer. Reproduced under
AddressSanitizer: a 64-byte heap frame was `accepted=1`, and the payload read
tripped `heap-buffer-overflow READ of size 1`.

The C lint that existed (clang-format, clang-tidy, cppcheck) ran **advisory**
(`continue-on-error`), so nothing actually gated merge on a memory-safety
finding, and there was no sanitizer or fuzz coverage of the parser.

## Decision

1. **Fix the guard.** Validators require their full header+payload length.
   Y.1564 and extended custom packets require 66 bytes. The standard RFC 2544
   measurement format uses an 18-byte compact payload, so its complete packet
   buffer is 60 bytes; the NIC appends the four-byte FCS to form the standard
   64-byte Ethernet frame. `ETH_P_IP` gained a non-Linux fallback so the parser
   compiles and is testable off Linux.

2. **Add a BLOCKING `dataplane-safety` CI job** (separate from advisory c-lint),
   required via the `CI Complete` aggregator. It fails the build on:
   - any ASAN/UBSan finding in the parser unit tests (`make c-test-asan`,
     `tests/c/test_packet_parse.c` — asserts each format's exact minimum and
     reads an accepted frame's full payload so a regression trips ASAN);
   - a cppcheck `warning`/`performance`/`portability` issue **scoped to
     packet.c** (tree-wide `style` findings stay advisory in c-lint to avoid
     blocking on unrelated pre-existing noise);
   - any crash from a bounded libFuzzer run of the parser
     (`make c-fuzz`, `tests/c/fuzz_packet.c`, 120s in CI).

## Consequences

- A short-frame OOB regression now fails CI three independent ways (assertion,
  ASAN, fuzz), not just a silently-passing advisory warning.
- The parser is continuously fuzzed; the harness reads the full payload of any
  accepted frame, so "accepted a frame too short to hold its payload" is caught.
- The blocking cppcheck is deliberately scoped to packet.c; widening it
  tree-wide is future work gated on clearing pre-existing `style`/`portability`
  findings (void-pointer arithmetic in the reflector platform code, etc.).
- The fuzz/ASAN targets use clang (libFuzzer); normal release builds are
  unchanged.

## Related issues and PRs

- The C dataplane hardening item in the stem/niac remediation plan.

# Stem Module Implementation Status

This document tracks which modules and tests are fully implemented vs stubbed.

**Last Updated:** 2025-12-30
**Version:** v0.1.5

---

## Implementation Summary

| Module | Status | Executable Tests | Notes |
|--------|--------|------------------|-------|
| **Benchmark** | ✅ Implemented | 6/6 | Full RFC 2544 support |
| **ServiceTest** | ⚠️ Partial | 3/6 | Y.1564 works, MEF stubbed |
| **TrafficGen** | ❌ Stub | 0/1 | Requires C dataplane work |
| **Measure** | ❌ Stub | 0/4 | Requires Y.1731 C implementation |
| **Certify** | ❌ Stub | 0/12 | Requires RFC 2889/6349/TSN C implementation |

---

## Module Details

### Benchmark (RFC 2544) ✅ IMPLEMENTED

**Location:** `internal/modules/benchmark/`
**Dataplane:** Uses `internal/testmaster/dataplane`
**Status:** Fully functional with CGO-enabled Linux builds

| Test Type | Status | Command |
|-----------|--------|---------|
| `throughput` | ✅ Works | `stem test -t throughput` |
| `latency` | ✅ Works | `stem test -t latency` |
| `frame_loss` | ✅ Works | `stem test -t frame_loss` |
| `back_to_back` | ✅ Works | `stem test -t back_to_back` |
| `system_recovery` | ✅ Works | `stem test -t system_recovery` |
| `reset` | ✅ Works | `stem test -t reset` |

**Notes:**
- Requires CGO and Linux for actual test execution
- Non-CGO builds return `ErrNotSupported`
- Full RFC 2544 methodology implemented in C dataplane

---

### ServiceTest (Y.1564 / MEF) ⚠️ PARTIAL

**Location:** `internal/modules/servicetest/`
**Dataplane:** Uses `internal/testmaster/dataplane`
**Status:** Y.1564 tests work; MEF tests return `ErrTestNotImplemented`

| Test Type | Status | Command |
|-----------|--------|---------|
| `y1564_config` | ✅ Works | `stem test -t y1564_config --cir 100` |
| `y1564_perf` | ✅ Works | `stem test -t y1564_perf --cir 100` |
| `y1564` | ✅ Works | `stem test -t y1564 --cir 100` |
| `mef_config` | ❌ Stub | Returns `ErrTestNotImplemented` |
| `mef_perf` | ❌ Stub | Returns `ErrTestNotImplemented` |
| `mef` | ❌ Stub | Returns `ErrTestNotImplemented` |

**Notes:**
- Y.1564 (ITU-T) tests are fully implemented
- MEF tests require additional C dataplane implementation
- MEF tests share similar structure to Y.1564 but need MEF-specific validation logic

**Required Work for MEF:**
- Implement MEF 14/48 bandwidth profile validation in C
- Add CoS (Class of Service) identification support
- Integrate MEF-specific SLA parameters

---

### TrafficGen (Custom Traffic) ❌ STUB

**Location:** `internal/modules/trafficgen/`
**Dataplane:** Would use `internal/reflector/dataplane`
**Status:** All tests return `ErrTestNotImplemented`

| Test Type | Status | Command |
|-----------|--------|---------|
| `custom_stream` | ❌ Stub | Returns `ErrTestNotImplemented` |

**Notes:**
- Custom traffic generation is planned but not implemented
- The reflector mode works for packet reflection but not arbitrary traffic generation
- Would require new C code in `src/dataplane/` for:
  - Programmable packet templates
  - Rate-controlled transmission
  - Pattern generation

---

### Measure (Y.1731 OAM) ❌ STUB

**Location:** `internal/modules/measure/`
**Dataplane:** Would use `internal/testmaster/dataplane`
**Status:** All tests return `ErrTestNotImplemented`

| Test Type | Status | Command |
|-----------|--------|---------|
| `y1731_delay` | ❌ Stub | Returns `ErrTestNotImplemented` |
| `y1731_loss` | ❌ Stub | Returns `ErrTestNotImplemented` |
| `y1731_slm` | ❌ Stub | Returns `ErrTestNotImplemented` |
| `y1731_loopback` | ❌ Stub | Returns `ErrTestNotImplemented` |

**Notes:**
- Y.1731 OAM requires carrier network support on the path
- Would require C implementation of:
  - DMM/DMR (Delay Measurement Message/Reply)
  - LMM/LMR (Loss Measurement Message/Reply)
  - SLM/SLR (Synthetic Loss Measurement)
  - LBM/LBR (Loopback Message/Reply)
- Needs MEP (Maintenance Entity Group End Point) support

---

### Certify (RFC 2889 / RFC 6349 / TSN) ❌ STUB

**Location:** `internal/modules/certify/`
**Dataplane:** Would use `internal/testmaster/dataplane`
**Status:** All tests return `ErrTestNotImplemented`

#### RFC 2889 (LAN Switching)

| Test Type | Status | Notes |
|-----------|--------|-------|
| `rfc2889_forwarding` | ❌ Stub | Multi-port forwarding rate |
| `rfc2889_caching` | ❌ Stub | MAC address table capacity |
| `rfc2889_learning` | ❌ Stub | Address learning rate |
| `rfc2889_broadcast` | ❌ Stub | Broadcast frame handling |
| `rfc2889_congestion` | ❌ Stub | Congestion control behavior |

#### RFC 6349 (TCP Throughput)

| Test Type | Status | Notes |
|-----------|--------|-------|
| `rfc6349_throughput` | ❌ Stub | TCP throughput measurement |
| `rfc6349_path` | ❌ Stub | Path analysis (RTT, BDP) |

#### TSN (IEEE 802.1Qbv)

| Test Type | Status | Notes |
|-----------|--------|-------|
| `tsn_timing` | ❌ Stub | Gate timing accuracy |
| `tsn_isolation` | ❌ Stub | Traffic class isolation |
| `tsn_latency` | ❌ Stub | Scheduled latency |
| `tsn` | ❌ Stub | Full TSN validation |

**Required Work:**
- RFC 2889: Multi-port traffic generation, MAC table manipulation
- RFC 6349: TCP stack integration, BDP calculation
- TSN: IEEE 1588 PTP synchronization, time-aware shaping support

---

## Platform Requirements

### Full Functionality (CGO + Linux)

For full test execution capability, build with:

```bash
CGO_ENABLED=1 GOOS=linux go build ./cmd/stem
```

Requirements:
- Linux kernel 4.15+ (for AF_PACKET)
- Optional: AF_XDP support (kernel 4.18+, libbpf)
- Optional: DPDK 23.11 LTS
- GCC 7.3+ or Clang 7+

### Stub Mode (Non-CGO or Non-Linux)

On macOS or with `CGO_ENABLED=0`:

- All dataplane operations return `ErrNotSupported`
- API endpoints return HTTP 503 with status "unavailable"
- Useful for development and testing of non-dataplane code

---

## Reflector Mode

The reflector mode (`stem reflect`) is separate from the module system:

| Feature | Status | Notes |
|---------|--------|-------|
| MAC reflection | ✅ Works | Swaps src/dst MAC |
| MAC+IP reflection | ✅ Works | Swaps MAC and IP |
| Full reflection | ✅ Works | All-layer swap |
| OUI filtering | ✅ Works | Filter by vendor OUI |
| Signature detection | ✅ Works | ITO, RFC2544, Y.1564, etc. |
| Statistics | ✅ Works | Real-time counters |
| Runtime config | ✅ Works | Hot-reconfigurable |

**Location:** `internal/reflector/`
**Dataplane:** Uses `internal/reflector/dataplane/`

---

## API Behavior

### When Tests Are Available

```json
POST /api/test/start
{
  "testType": "throughput",
  "interface": "eth0"
}

Response 200:
{
  "status": "running",
  "testType": "throughput"
}
```

### When Tests Are Stubbed

```json
POST /api/test/start
{
  "testType": "y1731_delay",
  "interface": "eth0"
}

Response 503:
{
  "status": "unavailable",
  "error": "Y.1731 OAM tests require additional dataplane implementation"
}
```

### When Platform Doesn't Support

```json
POST /api/test/start
{
  "testType": "throughput",
  "interface": "eth0"
}

Response 503:
{
  "status": "unavailable",
  "error": "CGO dataplane not available on this platform"
}
```

---

## Roadmap

### v0.2.0 (Planned)
- [ ] MEF test implementation
- [ ] Reflector API wiring to dataplane

### v0.3.0 (Planned)
- [ ] Y.1731 OAM basic support (loopback, delay)
- [ ] RFC 2889 forwarding rate

### Future
- [ ] RFC 6349 TCP throughput
- [ ] TSN support
- [ ] Custom traffic generation

---

## Contributing

To implement a stubbed module:

1. Add C implementation in `src/dataplane/`
2. Add CGO bindings in `internal/*/dataplane/`
3. Wire executor to call dataplane functions
4. Add tests in `*_test.go`
5. Update this document

See `internal/modules/benchmark/executor.go` for a reference implementation.

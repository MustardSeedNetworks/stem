# The Stem - Issue Tracker

Issues are now tracked via GitHub Issues: https://github.com/krisarmstrong/stem/issues

## Open Issues

### P1 (High)
(No open P1 issues)

### P2 (Medium)
- [#2 Replace interface{} with concrete types](https://github.com/krisarmstrong/stem/issues/2)
  - Web server uses typed response structs (ReflectorConfig, StatusResponse, etc.)
  - Remaining: UpdateConfig map[string]interface{}, test file JSON unmarshalling, executor Params
  - Note: Logging `...interface{}` is standard Go pattern and acceptable

---

## FIXED (v0.1.5)

### ~~Issue #9: /api/test/start reports running without execution~~
**Status**: FIXED in v0.1.5
- Implemented actual test execution via module executors (benchmark, servicetest)
- Returns 503 with "unavailable" status on platforms without CGO support

### ~~Issue #11: UpdateConfig silently ignores invalid OUI values~~
**Status**: FIXED in v0.1.5
- Now returns error with invalid OUI value and restores previous configuration

### ~~Issue #13: Consolidate 3 web servers into single production server~~
**Status**: FIXED in v0.1.5
- Removed unused reflector/web and testmaster/web packages
- Single web server at internal/web serves all functionality

### ~~Issue #4: Add interface validation to handleSettings~~
**Status**: FIXED in v0.1.5
- Validates interface exists before accepting settings update
- Logs interface selection for observability

### ~~Issue #3: Extract hardcoded values to constants~~
**Status**: FIXED in v0.1.5
- Added HTTP timeout constants (HTTPReadHeaderTimeout, etc.)

### ~~Issue #5: Add observability for config update failures~~
**Status**: FIXED in v0.1.5
- Added logging to handleReflectorConfig for failures and updates
- Added logging to handleMode for mode changes

### ~~Issue #7: Fix errcheck warnings in test files~~
**Status**: FIXED in v0.1.5
- Fixed all unchecked error returns in test files
- Used t.Setenv() instead of os.Setenv for proper cleanup

### ~~Issue #8: Fix golangci-lint warnings~~
**Status**: PARTIALLY FIXED in v0.1.5
- Fixed all errcheck warnings in production and test code
- Fixed exitAfterDefer warnings with explicit cleanup
- Fixed exhaustive switch warnings
- Extracted repeated strings as constants (goconst)
- Remaining: gocognit (high complexity), gosec (security), revive (style)

### ~~Issue #6: Document web server architecture~~
**Status**: FIXED in v0.1.5
- Added comprehensive package documentation to internal/web
- Documents API endpoints, security features, and architecture

### ~~Issue #10: Document interface capability detection~~
**Status**: FIXED in v0.1.5
- Added comprehensive package documentation to internal/interfaces
- Documents driver heuristic approach and its limitations
- Lists XDP-capable and DPDK-capable drivers

### ~~Issue #12: Document sysfs dependency~~
**Status**: FIXED in v0.1.5
- Documented sysfs paths used for interface metadata
- Noted platform limitations (Linux-only for full functionality)
- Added usage notes for operators

### ~~Issue #14: Add interface selection test coverage~~
**Status**: FIXED in v0.1.5
- Added edge case tests for score calculation
- Added score ordering verification test
- Added XDP/DPDK driver coverage tests
- Added loopback filtering test
- Added interface state detection test

---

## FIXED (v0.1.4)

### Module Architecture (v0.1.4)
- Created 6-module architecture (Reflector, Benchmark, ServiceTest, TrafficGen, Measure, Certify)
- Added module routing via handleTestStart
- Comprehensive module unit tests

### ~~Issue #1: JSON encode errors silently ignored in web handlers~~
**Status**: FIXED in v0.1.1, v0.1.2, v0.1.3
- Added `writeJSON()` helper with error logging to all web servers

### ~~Issue #5: Wildcard CORS headers allow any origin~~
**Status**: FIXED in v0.1.1, v0.1.2
- Added `setCORSHeaders()` function restricting to localhost origins

### ~~Issue #10: HTTP servers without timeouts~~
**Status**: FIXED in v0.1.3
- Added ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout

---

*Last Updated: 2025-12-30*
*Latest Release: v0.1.5*

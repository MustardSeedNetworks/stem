# The Stem - Issue Tracker

Issues are now tracked via GitHub Issues: https://github.com/krisarmstrong/stem/issues

## Open Issues

### P0 (Critical)
- **Issue #15**: Modules beyond benchmark/servicetest return 'executor not implemented'
  - trafficgen, measure, certify modules need C dataplane implementation

### P1 (High)
(No open P1 issues)

### P2 (Medium)
- **Issue #19**: Address enumeration ignores errors
- **Issue #20**: Params parsing relies on float64 type assertions
- **Issue #21**: No default interface selection applied
- **Issue #23**: Missing tests for test cancellation behavior
- **Issue #24**: Missing validation for executor parameter types
- **Issue #25**: Missing E2E coverage for reflector mode

---

## FIXED (v0.1.6)

### ~~Issue #17: Reflector API config/stats not wired to dataplane~~
**Status**: FIXED in v0.1.6
- Added reflectorExec field to Server struct for active reflector executor
- handleReflectorStats reads from actual dataplane stats
- handleReflectorConfig calls dataplane.UpdateConfig() for real-time updates
- Added Dataplane() accessor to reflector executor

### ~~Issue #18: /api/test/stop does not signal dataplane to stop~~
**Status**: FIXED in v0.1.6
- handleTestStop checks if reflector executor is running and stops it
- Added executeReflector() to start reflector via module system
- Properly updates testStatus and logs stop events

### ~~Issue #22: Document which modules/tests are executable vs stub~~
**Status**: FIXED in v0.1.6
- Created docs/MODULE_STATUS.md with comprehensive implementation status
- Documents all 5 modules with test-by-test availability
- Includes platform requirements and API behavior documentation

### ~~Issue #16: MEF tests always return ErrTestNotImplemented~~
**Status**: DOCUMENTED in v0.1.6
- MEF tests documented in MODULE_STATUS.md as requiring C dataplane work
- Clear API behavior documented for stubbed tests

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

### ~~Issue #2: Replace interface{} with concrete types~~
**Status**: FIXED in v0.1.5
- Created typed ConfigUpdate struct for UpdateConfig method
- Web server already uses typed response structs
- Note: Logging `...interface{}` and executor Params are acceptable Go patterns

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
*Latest Release: v0.1.6*

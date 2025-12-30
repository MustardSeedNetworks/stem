# The Stem - Issue Tracker

## FIXED (v0.1.3)

### ~~Issue #1: JSON encode errors silently ignored in web handlers~~
**Status**: FIXED in v0.1.1, v0.1.2, v0.1.3
- Added `writeJSON()` helper with error logging to all web servers
- `internal/web/server.go` - 16 instances fixed
- `internal/reflector/web/server.go` - 6 instances fixed
- `internal/testmaster/web/web.go` - 7 instances fixed

### ~~Issue #2: Missing port validation allows invalid values~~
**Status**: FIXED in v0.1.1
- Added port range validation (1-65535) in `cmd/stem/main.go`

### ~~Issue #3: Nil license manager not handled consistently~~
**Status**: FIXED in v0.1.1
- Added proper nil checks with logging warnings

### ~~Issue #4: Frame size validation silently drops invalid values~~
**Status**: FIXED in v0.1.1
- Added logging.Warn() for invalid frame sizes

### ~~Issue #5: Wildcard CORS headers allow any origin~~
**Status**: FIXED in v0.1.1, v0.1.2
- Added `setCORSHeaders()` function restricting to localhost origins

### ~~Issue #6: Signal handler makes blocking call in select~~
**Status**: FIXED in v0.1.1
- Changed to `go dp.Stop()` for non-blocking shutdown

### ~~Issue #9: Logging initialization error ignored~~
**Status**: FIXED in v0.1.1
- Added proper error handling with stderr warning

### ~~Issue #10: HTTP servers without timeouts~~
**Status**: FIXED in v0.1.3
- Added ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout
- Fixes gosec G114 security vulnerability

---

## REMAINING - LOW PRIORITY

### Issue #7: Replace interface{} with concrete types
**Severity**: LOW
**Files**: `internal/web/server.go`, `cmd/stem/main.go`
**Status**: Partially addressed - most critical spots use concrete types now

### Issue #8: Extract hardcoded values to constants
**Severity**: LOW
**Files**: Various
**Status**: OPEN - cosmetic improvement only

---

## KNOWN ISSUES (Test Files Only)

The following are test file issues flagged by golangci-lint. These are not production code issues:

- `cmd/stem/main_test.go` - Unchecked `w.Close()` and `buf.ReadFrom()` in test helpers
- `internal/help/help_test.go` - Unchecked `w.Close()` and `io.Copy()` in test helpers

These are acceptable in test code and do not affect production.

---

*Last Updated: 2025-12-29*
*Latest Release: v0.1.3*

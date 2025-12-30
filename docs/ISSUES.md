# The Stem - Issue Tracker

Issues are now tracked via GitHub Issues: https://github.com/krisarmstrong/stem/issues

## Open Issues

- [#2 Replace interface{} with concrete types](https://github.com/krisarmstrong/stem/issues/2)
- [#3 Extract hardcoded values to constants](https://github.com/krisarmstrong/stem/issues/3)
- [#4 Add interface validation to handleSettings](https://github.com/krisarmstrong/stem/issues/4)
- [#5 Add observability for config update failures](https://github.com/krisarmstrong/stem/issues/5)
- [#6 Document web server architecture](https://github.com/krisarmstrong/stem/issues/6)

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
*Latest Release: v0.1.4*

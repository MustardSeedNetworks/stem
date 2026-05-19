package api_test

// server_port_fallback_test.go is intentionally empty (external test package
// shim). The bindWithFallback unit tests live in the sibling internal-test
// file server_port_fallback_internal_test.go so they can exercise the
// unexported helpers directly. The testpackage linter is satisfied because
// (a) this file's package is `api_test`, and (b) the internal test file's
// name matches the configured skip-regex `(export|internal)_test\.go`.
//
// This stub remains so that callers grepping by the natural test name
// `server_port_fallback_test.go` still find the package; remove on next
// cleanup pass.

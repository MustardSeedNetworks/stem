// SPDX-License-Identifier: BUSL-1.1

package auth

// export_test.go exposes unexported HIBP test seams to the external
// `auth_test` package so hibp_test.go can run with the testpackage
// linter convention. These exports are only compiled into test
// binaries (the `_test.go` suffix ensures the symbols never leak into
// production builds).

// SetHibpEndpointForTest overrides the in-package hibpEndpoint variable
// so tests can redirect the HIBP API URL at an httptest.Server. It
// returns the previous value so the caller can restore it on cleanup.
// The write is serialised through hibpMu so parallel hibp_test.go
// cases (t.Parallel + withTestEndpoint) clear the race detector.
func SetHibpEndpointForTest(url string) string {
	hibpMu.Lock()
	defer hibpMu.Unlock()
	prev := hibpEndpoint
	hibpEndpoint = url
	return prev
}

// SetHibpLoggerForTest overrides the in-package hibpLogger variable so
// tests can capture soft-failure log calls. It returns the previous
// logger so the caller can restore it on cleanup. Serialised through
// hibpMu for the same reason as SetHibpEndpointForTest.
func SetHibpLoggerForTest(fn func(string, error)) func(string, error) {
	hibpMu.Lock()
	defer hibpMu.Unlock()
	prev := hibpLogger
	hibpLogger = fn
	return prev
}

// HibpPrefixLen re-exports hibpPrefixLen for tests that need to slice
// SHA-1 hex strings the same way CheckPasswordBreached does.
const HibpPrefixLen = hibpPrefixLen

// HibpUserAgent re-exports hibpUserAgent for tests that assert the
// outgoing User-Agent header.
const HibpUserAgent = hibpUserAgent

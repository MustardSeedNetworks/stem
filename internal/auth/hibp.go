// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"bufio"
	"context"
	"crypto/sha1" //nolint:gosec // SHA-1 is the only hash HIBP exposes (k-anonymity API).
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HIBP integration constants. See https://haveibeenpwned.com/API/v3#PwnedPasswords.
const (
	// hibpDefaultEndpoint is the public k-anonymity range endpoint.
	hibpDefaultEndpoint = "https://api.pwnedpasswords.com/range/"
	// hibpUserAgent identifies this client to HIBP per their TOS.
	hibpUserAgent = "stem/security-check"
	// hibpTimeout is the maximum time we'll wait for an HIBP response.
	// Set deliberately short — air-gapped deployments must not stall
	// password changes for 30 seconds while DNS times out.
	hibpTimeout = 5 * time.Second
	// hibpPrefixLen is the SHA-1 hex prefix length expected by the
	// k-anonymity API.
	hibpPrefixLen = 5
	// hibpSuffixLen is the remaining hex suffix length (40 total - 5 prefix).
	hibpSuffixLen = 35
	// hibpDisableEnv is the opt-out switch for air-gapped deployments.
	hibpDisableEnv = "STEM_DISABLE_HIBP"
	// hibpSuccessStatusClass is the integer-divided value of HTTP 2xx
	// status codes (status / 100 == 2). Named to avoid the magic-number
	// linter complaint at the comparison site.
	hibpSuccessStatusClass = 2
)

// hibpMu guards the test seams (hibpEndpoint, hibpLogger) so the
// race detector stays clean when parallel hibp_test.go cases mutate
// them via withTestEndpoint / SetHibpLoggerForTest. Production reads
// the same vars on every CheckPasswordBreached call, so the lock is
// required to make the test-write / production-read pairing safe.
//
//nolint:gochecknoglobals // sentinel for the package-level test seams
var hibpMu sync.RWMutex

// hibpEndpoint is overridable for tests; production code reads via
// resolveHIBPEndpoint() which falls back to hibpDefaultEndpoint.
//
//nolint:gochecknoglobals // test seam, guarded by hibpMu
var hibpEndpoint = ""

// hibpClient is a lazily-constructed HTTP client with a short timeout.
// Lazy init lets tests override hibpEndpoint without racing this var.
//
//nolint:gochecknoglobals // test seam
var hibpClient = &http.Client{Timeout: hibpTimeout}

// resolveHIBPEndpoint returns the active HIBP range URL prefix.
func resolveHIBPEndpoint() string {
	hibpMu.RLock()
	defer hibpMu.RUnlock()
	if hibpEndpoint != "" {
		return hibpEndpoint
	}
	return hibpDefaultEndpoint
}

// HIBPDisabled reports whether the operator has opted out of HIBP
// breach checks via [hibpDisableEnv]. This is the documented escape
// hatch for air-gapped or compliance-restricted deployments.
func HIBPDisabled() bool {
	return os.Getenv(hibpDisableEnv) == "1"
}

// ErrHIBPUnavailable indicates the HIBP API could not be reached. It
// is **deliberately not exported** as a hard rejection: the
// CheckPasswordBreached contract is that network failures degrade
// gracefully (see the func docs).
var ErrHIBPUnavailable = errors.New("hibp api unavailable")

// CheckPasswordBreached queries the HIBP Pwned Passwords k-anonymity
// API to see if `password` appears in a known breach corpus.
//
// Contract:
//   - Returns (breached, count, nil) when the API responded.
//   - Returns (false, 0, nil) — explicitly not blocking — when:
//     STEM_DISABLE_HIBP=1, the network is unreachable, the request
//     times out, or the server returns a non-2xx status. Operators on
//     air-gapped networks should not be locked out of password
//     changes by an unreachable third-party.
//   - Only returns a non-nil error for genuinely unexpected internal
//     failures (e.g. SHA-1 hashing returns short — impossible in
//     practice but kept defensively).
//
// The password leaves this function only as a SHA-1 hex prefix (5
// chars) per the k-anonymity protocol — the full hash is never sent.
func CheckPasswordBreached(ctx context.Context, password string) (bool, int, error) {
	if HIBPDisabled() {
		return false, 0, nil
	}

	hash := sha1.Sum([]byte(password)) //nolint:gosec // HIBP requires SHA-1.
	hex := strings.ToUpper(hex.EncodeToString(hash[:]))
	if len(hex) != hibpPrefixLen+hibpSuffixLen {
		return false, 0, fmt.Errorf("sha1 produced unexpected length %d", len(hex))
	}
	prefix := hex[:hibpPrefixLen]
	suffix := hex[hibpPrefixLen:]

	ctx, cancel := context.WithTimeout(ctx, hibpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolveHIBPEndpoint()+prefix, nil)
	if err != nil {
		// A malformed request is a real bug, not a network issue.
		return false, 0, fmt.Errorf("build hibp request: %w", err)
	}
	req.Header.Set("User-Agent", hibpUserAgent)
	// Per HIBP docs, opt out of the padding mechanism is not required;
	// the response is already padded for k-anonymity.

	resp, err := hibpClient.Do(req)
	if err != nil {
		// Network failure: log + return non-blocking false.
		logHIBPSoftFailure("hibp request failed", err)
		return false, 0, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != hibpSuccessStatusClass {
		logHIBPSoftFailure("hibp non-2xx", fmt.Errorf("status=%d", resp.StatusCode))
		return false, 0, nil
	}

	count, err := parseHIBPRangeBody(resp.Body, suffix)
	if err != nil {
		logHIBPSoftFailure("hibp parse failed", err)
		return false, 0, nil
	}

	return count > 0, count, nil
}

// parseHIBPRangeBody scans the `<suffix>:<count>` lines returned by
// the range endpoint and returns the count matching `wantSuffix`
// (case-insensitive). Returns 0 if not found.
func parseHIBPRangeBody(body io.Reader, wantSuffix string) (int, error) {
	scanner := bufio.NewScanner(body)
	wantUpper := strings.ToUpper(wantSuffix)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		gotSuffixRaw, countRaw, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.ToUpper(gotSuffixRaw) != wantUpper {
			continue
		}
		count, parseErr := strconv.Atoi(strings.TrimSpace(countRaw))
		if parseErr != nil {
			return 0, fmt.Errorf("bad count for suffix: %w", parseErr)
		}
		return count, nil
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan hibp body: %w", err)
	}
	return 0, nil
}

// logHIBPSoftFailure is wired by [SetHIBPLogger] so we don't pull
// logging into the auth package. By default it is a no-op; the api
// package registers a slog-backed logger on init.
//
//nolint:gochecknoglobals // singleton hook (no init cycle), guarded by hibpMu
var hibpLogger = func(_ string, _ error) {}

// SetHIBPLogger lets the api layer inject its slog instance for soft
// HIBP failures without creating an import cycle into internal/logging.
func SetHIBPLogger(fn func(msg string, err error)) {
	if fn == nil {
		return
	}
	hibpMu.Lock()
	defer hibpMu.Unlock()
	hibpLogger = fn
}

func logHIBPSoftFailure(msg string, err error) {
	hibpMu.RLock()
	fn := hibpLogger
	hibpMu.RUnlock()
	fn(msg, err)
}

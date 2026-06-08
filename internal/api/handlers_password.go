// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"net/http"
	"strings"

	"github.com/MustardSeedNetworks/stem/internal/auth"
	"github.com/MustardSeedNetworks/stem/internal/logging"
)

// detectHashAlgorithm returns a short tag identifying the hashing
// algorithm of an existing password hash, for inclusion in audit logs
// (`previous_algorithm`). The function never errors — anything we
// don't recognise (including empty) is reported as "none".
func detectHashAlgorithm(hash string) string {
	switch {
	case strings.HasPrefix(hash, "$argon2id$"):
		return "argon2id"
	case strings.HasPrefix(hash, "$2a$"),
		strings.HasPrefix(hash, "$2b$"),
		strings.HasPrefix(hash, "$2y$"):
		return "bcrypt"
	default:
		return "none"
	}
}

// validatePasswordOrReject runs the layered password-policy check used
// by both the initial-setup and recovery endpoints:
//
//  1. Length floor ([auth.ValidatePasswordStrength])
//  2. zxcvbn score floor ([auth.EvaluatePasswordStrength], min score 3)
//  3. HIBP breach corpus ([auth.CheckPasswordBreached])
//
// On rejection it writes an HTTP 400 with a short operator-friendly
// message, emits an audit log, and returns false. The caller must stop
// processing.
//
// HIBP is treated as opt-out: STEM_DISABLE_HIBP=1 skips it. Network
// errors against the HIBP API are non-blocking (see CheckPasswordBreached).
func validatePasswordOrReject(
	w http.ResponseWriter,
	r *http.Request,
	password, username, prevAlgorithm string,
) bool {
	// Layer 1: length / cheap validation.
	if err := auth.ValidatePasswordStrength(password); err != nil {
		logging.AuditPasswordChange(r.Context(), r, username,
			logging.PasswordChangeRejected, "too_short", prevAlgorithm, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Layer 2: zxcvbn strength meter.
	strength := auth.EvaluatePasswordStrength(password, []string{username, "stem", "thestem"})
	if !strength.Acceptable() {
		msg := strength.Warning
		if len(strength.Suggestions) > 0 {
			msg = strength.Warning + " " + strength.Suggestions[0]
		}
		if msg == "" {
			msg = "Password is too weak."
		}
		logging.AuditPasswordChange(r.Context(), r, username,
			logging.PasswordChangeRejected, "weak_score", prevAlgorithm, msg)
		http.Error(w, msg, http.StatusBadRequest)
		return false
	}

	// Layer 3: HIBP breach corpus.
	breached, count, err := auth.CheckPasswordBreached(r.Context(), password)
	if err != nil {
		// Internal failure (not a network soft-fail). Log + continue —
		// fail-open here mirrors the air-gapped contract.
		logging.Warn("HIBP check internal failure (continuing)", "error", err,
			"event", "auth.hibp.internal_error")
	}
	if breached {
		msg := formatBreachedMessage(count)
		logging.AuditPasswordChange(r.Context(), r, username,
			logging.PasswordChangeRejected, "breached", prevAlgorithm, msg)
		http.Error(w, msg, http.StatusBadRequest)
		return false
	}

	return true
}

// formatBreachedMessage returns an operator-friendly rejection message
// noting how many breaches contain this password.
func formatBreachedMessage(count int) string {
	if count <= 0 {
		return "This password has been found in known data breaches. " +
			"Choose a different password."
	}
	return formatBreachedWithCount(count)
}

// formatBreachedWithCount is split out so tests can pin the exact
// wording independent of plural handling logic.
func formatBreachedWithCount(count int) string {
	noun := "breach"
	if count != 1 {
		noun = "breaches"
	}
	return "This password has appeared in known data " + noun +
		". Choose a different password."
}

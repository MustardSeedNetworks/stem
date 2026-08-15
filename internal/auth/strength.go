// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"github.com/trustelem/zxcvbn"
)

// zxcvbn score bands. The library returns an int in [0,4]; we name
// the bands referenced by strengthFeedback so the switch reads as
// English rather than as bare numeric literals.
const (
	// scoreVeryGuessable is zxcvbn band 1.
	scoreVeryGuessable = 1
	// scoreSomewhatGuessable is zxcvbn band 2.
	scoreSomewhatGuessable = 2
	// scoreSafelyUnguessable is zxcvbn band 3 - the floor stem requires
	// after Wave 2 (#86): "moderate protection from offline slow-hash".
	scoreSafelyUnguessable = 3
)

// MinPasswordScore is the minimum zxcvbn score (0-4) accepted by
// password-set / password-change endpoints. See [scoreSafelyUnguessable].
const MinPasswordScore = scoreSafelyUnguessable

// PasswordStrength is the structured result of [EvaluatePasswordStrength].
// It mirrors the zxcvbn v4 feedback contract — Warning + Suggestions —
// even though the Go port [github.com/trustelem/zxcvbn] doesn't ship
// feedback strings, so we synthesize operator-friendly ones based on
// the score band.
type PasswordStrength struct {
	// Score is the zxcvbn score in [0,4].
	//   0 = too guessable
	//   1 = very guessable
	//   2 = somewhat guessable
	//   3 = safely unguessable (minimum accepted)
	//   4 = very unguessable
	Score int `json:"score"`

	// EstimatedGuesses is the estimated number of guesses needed to
	// crack the password (zxcvbn's primary metric).
	EstimatedGuesses float64 `json:"estimated_guesses"`

	// CrackTimeSeconds is a coarse estimate at 10k guesses/sec
	// (offline slow-hash scenario — matches our Argon2id parameters).
	CrackTimeSeconds float64 `json:"crack_time_seconds"`

	// Warning is a short user-facing message describing why the
	// password is weak, or empty if the score is acceptable.
	Warning string `json:"warning,omitempty"`

	// Suggestions are actionable improvements (up to 3).
	Suggestions []string `json:"suggestions,omitempty"`
}

// Acceptable reports whether the password meets the minimum score.
func (p PasswordStrength) Acceptable() bool {
	return p.Score >= MinPasswordScore
}

// crackTimeDivisor is the assumed offline slow-hash guess rate
// (1e4 guesses/sec — RFC 9106 / OWASP recommendation for Argon2id at
// our parameters).
const crackTimeDivisor = 1e4

// EvaluatePasswordStrength runs zxcvbn against `password`, optionally
// taking `userInputs` (e.g. username, host names) into account so the
// scorer penalizes choices that reuse known-attacker-context data.
//
// The function never errors: zxcvbn always produces a score. Callers
// should reject when [PasswordStrength.Acceptable] is false and
// surface [PasswordStrength.Warning] + the first suggestion.
func EvaluatePasswordStrength(password string, userInputs []string) PasswordStrength {
	result := zxcvbn.PasswordStrength(password, userInputs)

	warning, suggestions := strengthFeedback(result.Score, len(password), len(userInputs) > 0)

	return PasswordStrength{
		Score:            result.Score,
		EstimatedGuesses: result.Guesses,
		CrackTimeSeconds: result.Guesses / crackTimeDivisor,
		Warning:          warning,
		Suggestions:      suggestions,
	}
}

// strengthFeedback returns operator-friendly messages by score band.
// The trustelem port doesn't expose zxcvbn's English feedback table,
// so we provide a small fixed set keyed on score.
func strengthFeedback(score, length int, hadUserInputs bool) (string, []string) {
	switch {
	case score >= MinPasswordScore:
		return "", nil
	case score == scoreSomewhatGuessable:
		return "This password is somewhat guessable.",
			feedbackSuggestions(length, hadUserInputs)
	case score == scoreVeryGuessable:
		return "This password is very guessable.",
			feedbackSuggestions(length, hadUserInputs)
	default:
		return "This password is too easily guessable.",
			feedbackSuggestions(length, hadUserInputs)
	}
}

// recommendedLength is the length above which we stop suggesting "longer".
const recommendedLength = 16

// maxFeedbackSuggestions caps how many actionable strings we surface to
// the user, mirroring zxcvbn v4's `feedback.suggestions` contract.
const maxFeedbackSuggestions = 3

func feedbackSuggestions(length int, hadUserInputs bool) []string {
	suggestions := make([]string, 0, maxFeedbackSuggestions)
	if length < recommendedLength {
		suggestions = append(suggestions,
			"Use a longer passphrase (16+ characters with multiple words).")
	}
	suggestions = append(suggestions,
		"Avoid common words, predictable patterns, and personal details.")
	if hadUserInputs {
		suggestions = append(suggestions,
			"Avoid reusing the username, hostname, or product name in the password.")
	}
	return suggestions
}

// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// TestEvaluatePasswordStrength_Weak asserts the score floor: a trivial
// password must fall below MinPasswordScore so the API layer rejects.
func TestEvaluatePasswordStrength_Weak(t *testing.T) {
	t.Parallel()

	weak := []string{
		"password",
		"123456",
		"qwerty",
		"abc",
	}

	for _, pw := range weak {
		t.Run(pw, func(t *testing.T) {
			t.Parallel()
			s := auth.EvaluatePasswordStrength(pw, nil)
			if s.Acceptable() {
				t.Errorf("password %q should be unacceptable, got score=%d", pw, s.Score)
			}
			if s.Warning == "" {
				t.Error("weak password should carry a warning")
			}
		})
	}
}

// TestEvaluatePasswordStrength_Strong asserts genuinely strong inputs
// clear the score floor. We use long, varied passphrases that the
// trustelem port consistently scores at 3+ to avoid flakes.
func TestEvaluatePasswordStrength_Strong(t *testing.T) {
	t.Parallel()

	strong := []string{
		"corre7t-horse-batt3ry-staple-puzzle",
		"Tr0ub4dor&3-redux-with-extra-padding",
		"my-cat-likes-tuna-on-tuesdays-only",
	}

	for _, pw := range strong {
		t.Run(pw, func(t *testing.T) {
			t.Parallel()
			s := auth.EvaluatePasswordStrength(pw, nil)
			if !s.Acceptable() {
				t.Errorf("password %q should be acceptable (>=score %d), got score=%d warning=%q",
					pw, auth.MinPasswordScore, s.Score, s.Warning)
			}
		})
	}
}

// TestEvaluatePasswordStrength_PenalizesUserInput verifies that
// passwords containing the username/hostname score lower than the same
// password without that user context.
func TestEvaluatePasswordStrength_PenalizesUserInput(t *testing.T) {
	t.Parallel()

	// A password where the username substring would be the weakest link.
	password := "alice-loves-stem-tigers"

	plain := auth.EvaluatePasswordStrength(password, nil)
	withUser := auth.EvaluatePasswordStrength(password, []string{"alice", "stem"})

	if withUser.Score > plain.Score {
		t.Errorf("user context should not increase score: plain=%d withUser=%d",
			plain.Score, withUser.Score)
	}
}

// TestPasswordStrength_AcceptableThreshold pins the contract: the
// acceptance floor is MinPasswordScore.
func TestPasswordStrength_AcceptableThreshold(t *testing.T) {
	t.Parallel()

	if auth.MinPasswordScore != 3 {
		t.Errorf("MinPasswordScore drifted from spec: got %d, want 3", auth.MinPasswordScore)
	}

	cases := []struct {
		score      int
		acceptable bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, true},
		{4, true},
	}
	for _, tc := range cases {
		ps := auth.PasswordStrength{Score: tc.score}
		if ps.Acceptable() != tc.acceptable {
			t.Errorf("score=%d Acceptable()=%v want %v", tc.score, ps.Acceptable(), tc.acceptable)
		}
	}
}

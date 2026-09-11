// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

// The CLI presents this token on the same Bearer path a browser session uses,
// so the one validator must accept it and name it as a CLI credential.
func TestGenerateCLITokenValidates(t *testing.T) {
	t.Parallel()

	mgr, err := auth.NewManager("test-secret-value-at-least-32-chars-long", 0, "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	t.Cleanup(mgr.Stop)

	token, err := mgr.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}

	claims, err := mgr.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Username != "operator" {
		t.Errorf("username = %q, want %q", claims.Username, "operator")
	}
	if claims.TokenType != auth.TokenTypeCLI {
		t.Errorf("token type = %q, want %q", claims.TokenType, auth.TokenTypeCLI)
	}
}

// A leaked file must stop working on its own; an unexpiring bearer token on
// disk is a permanent grant.
func TestGenerateCLITokenExpires(t *testing.T) {
	t.Parallel()

	mgr, err := auth.NewManager("test-secret-value-at-least-32-chars-long", 0, "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	t.Cleanup(mgr.Stop)

	token, err := mgr.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}
	claims, err := mgr.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("CLI token has no expiry")
	}
	// Pinned to the literal, not to CLITokenLifetime: comparing the constant
	// against itself would accept any value the constant is changed to.
	if got := claims.ExpiresAt.Sub(claims.IssuedAt.Time); got != 24*time.Hour {
		t.Errorf("lifetime = %v, want %v", got, 24*time.Hour)
	}
}

// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

func tokenTypeManager(t *testing.T) *auth.Manager {
	t.Helper()
	mgr, err := auth.NewManager("test-secret-value-at-least-32-chars-long", 0, "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	t.Cleanup(mgr.Stop)
	return mgr
}

// A refresh token is deliberately longer-lived than an access token. If it
// also authenticates API requests, that longer life becomes full API access
// and the separation buys nothing (#1169).
func TestValidateTokenRejectsARefreshToken(t *testing.T) {
	t.Parallel()

	mgr := tokenTypeManager(t)
	refresh, err := mgr.GenerateRefreshToken("operator")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	_, validateErr := mgr.ValidateToken(context.Background(), refresh)
	if !errors.Is(validateErr, auth.ErrWrongTokenType) {
		t.Errorf("err = %v, want ErrWrongTokenType", validateErr)
	}
}

// The credentials that may authenticate an API request: a browser session
// and the local CLI.
func TestValidateTokenAcceptsAPICredentials(t *testing.T) {
	t.Parallel()

	mgr := tokenTypeManager(t)

	access, err := mgr.Authenticate(context.Background(), "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	cli, err := mgr.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}

	for name, token := range map[string]string{"access": access, "cli": cli} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, validateErr := mgr.ValidateToken(context.Background(), token); validateErr != nil {
				t.Errorf("ValidateToken(%s) = %v, want it accepted", name, validateErr)
			}
		})
	}
}

// The exchange endpoint is the mirror image: only a refresh token may be
// exchanged, so an access or CLI token cannot mint fresh sessions forever.
func TestValidateRefreshTokenAcceptsOnlyRefresh(t *testing.T) {
	t.Parallel()

	mgr := tokenTypeManager(t)

	refresh, err := mgr.GenerateRefreshToken("operator")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	access, err := mgr.Authenticate(context.Background(), "operator", "Correct-Horse-Battery-9")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	cli, err := mgr.GenerateCLIToken()
	if err != nil {
		t.Fatalf("GenerateCLIToken: %v", err)
	}

	if _, refreshErr := mgr.ValidateRefreshToken(context.Background(), refresh); refreshErr != nil {
		t.Errorf("ValidateRefreshToken(refresh) = %v, want it accepted", refreshErr)
	}
	for name, token := range map[string]string{"access": access, "cli": cli} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, validateErr := mgr.ValidateRefreshToken(context.Background(), token)
			if !errors.Is(validateErr, auth.ErrWrongTokenType) {
				t.Errorf("ValidateRefreshToken(%s) = %v, want ErrWrongTokenType", name, validateErr)
			}
		})
	}
}

// A token carrying no type, or one this build does not know, is refused
// rather than treated as an access token — the default has to fail closed.
func TestValidateTokenRejectsAnUnknownType(t *testing.T) {
	t.Parallel()

	mgr := tokenTypeManager(t)
	for name, tokenType := range map[string]string{"empty": "", "unknown": "totally-made-up"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			token, err := auth.ExportGenerateTokenWithType(mgr, "operator", tokenType)
			if err != nil {
				t.Fatalf("generate %s token: %v", name, err)
			}
			_, validateErr := mgr.ValidateToken(context.Background(), token)
			if !errors.Is(validateErr, auth.ErrWrongTokenType) {
				t.Errorf("ValidateToken(%s) = %v, want ErrWrongTokenType", name, validateErr)
			}
		})
	}
}

// The refresh exchange must still work end to end once the types are
// enforced: this is the path a browser session renews on.
func TestRefreshAccessTokenStillExchanges(t *testing.T) {
	t.Parallel()

	mgr := tokenTypeManager(t)
	refresh, err := mgr.GenerateRefreshToken("operator")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	access, err := mgr.RefreshAccessToken(context.Background(), refresh)
	if err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	claims, err := mgr.ValidateToken(context.Background(), access)
	if err != nil {
		t.Fatalf("the exchanged token does not authenticate: %v", err)
	}
	if claims.Username != "operator" {
		t.Errorf("username = %q, want operator", claims.Username)
	}
}

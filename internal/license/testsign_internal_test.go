// SPDX-License-Identifier: BUSL-1.1

package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

// TEST-ONLY Ed25519 private seed (distinct from the production key). Tokens
// minted with it validate through internalTestVerifier but are rejected by the
// embedded production key, so forgery tests remain meaningful.
const internalTestSigningSeedB64 = "XCx+b6yDNFoRanJhHeqX3pjXlhjNvXvAzojwaSq8lAs="

func internalTestSigningKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	seed, err := base64.StdEncoding.DecodeString(internalTestSigningSeedB64)
	if err != nil {
		t.Fatalf("decode test seed: %v", err)
	}
	return ed25519.NewKeyFromSeed(seed)
}

// internalTestVerifier returns a Verifier for the test signing key.
func internalTestVerifier(t *testing.T) *Verifier {
	t.Helper()
	return NewVerifier(internalTestSigningKey(t).Public().(ed25519.PublicKey))
}

// signTestKey is a drop-in replacement for the removed GenerateLicenseKey: it
// mints a production-shaped token signed by the TEST key and returns
// (token, nil). It validates through internalTestVerifier.
func signTestKey(t *testing.T, code, serial string, tier Tier) (string, error) {
	t.Helper()
	payload := map[string]any{
		"v":          1,
		"product":    "stem",
		"code":       code,
		"serial":     serial,
		"tier":       int(tier),
		"maxDevices": 3,
		"iat":        time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC).Unix(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	sig := ed25519.Sign(internalTestSigningKey(t), b)
	return "MSN1." + base64.RawURLEncoding.EncodeToString(b) +
		"." + base64.RawURLEncoding.EncodeToString(sig), nil
}

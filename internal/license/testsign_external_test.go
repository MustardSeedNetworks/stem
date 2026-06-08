// SPDX-License-Identifier: BUSL-1.1

package license_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// testSigningSeedB64 is a TEST-ONLY Ed25519 private seed, distinct from the
// production key. Tokens minted with it validate through a Verifier built on its
// public half (testVerifier) and through a Manager created with that verifier,
// but are correctly REJECTED by the embedded production key — which is what
// makes the forgery tests meaningful.
const testSigningSeedB64 = "XCx+b6yDNFoRanJhHeqX3pjXlhjNvXvAzojwaSq8lAs="

// Production-signed contract vectors (serial 1234567), produced by the keygen
// tool against the embedded production key. They MUST validate via the default
// verifier; regenerate together with the embedded key if it rotates.
const (
	prodStemReflectorVector    = "MSN1.eyJ2IjoxLCJwcm9kdWN0Ijoic3RlbSIsImNvZGUiOiIxMDAxIiwic2VyaWFsIjoiMTIzNDU2NyIsInRpZXIiOjEsIm1heERldmljZXMiOjMsImlhdCI6MTc4MDg3NjgwMH0.5SV0MhUNd9_em5B1_noWxJUbZpWjtbgf91BnTqi3uK2GBqeDiy0xdYWR3fDTq1HRRKa1TfDx6MNhufpjIbzhBg"
	prodStemProfessionalVector = "MSN1.eyJ2IjoxLCJwcm9kdWN0Ijoic3RlbSIsImNvZGUiOiIyMDAxIiwic2VyaWFsIjoiMTIzNDU2NyIsInRpZXIiOjIsIm1heERldmljZXMiOjMsImlhdCI6MTc4MDg3NjgwMH0.hPCX0LGKbRfOIRl6CQxUPcCgzlD0BnZloMFIWL_z7NaGmLOtvRoxIvOYjNxQ7uq1rpPMrWEY1dHDnwPtQlNFDA"
	prodStemEnterpriseVector   = "MSN1.eyJ2IjoxLCJwcm9kdWN0Ijoic3RlbSIsImNvZGUiOiIzMDAxIiwic2VyaWFsIjoiMTIzNDU2NyIsInRpZXIiOjMsIm1heERldmljZXMiOjMsImlhdCI6MTc4MDg3NjgwMH0.55Qx3dguyUocqoi_9YXufcIkphPB5kiSV28SueO2wGZ6UK3d_HaGcwpO2hon2k-qX9BBz0QC1mjNC1DeBkWIBQ"
)

func testSigningKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	seed, err := base64.StdEncoding.DecodeString(testSigningSeedB64)
	if err != nil {
		t.Fatalf("decode test seed: %v", err)
	}
	return ed25519.NewKeyFromSeed(seed)
}

// testVerifier returns a Verifier for the test signing key. Tokens from
// signTestKey validate against it.
func testVerifier(t *testing.T) *license.Verifier {
	t.Helper()
	pub := testSigningKey(t).Public().(ed25519.PublicKey)
	return license.NewVerifier(pub)
}

// signLicenseToken mints an MSN1 token signed by priv. It mirrors the keygen /
// verifier wire format so tests can produce arbitrary tokens (including ones
// signed by an attacker key for forgery tests) without the production private
// key.
func signLicenseToken(
	t *testing.T,
	priv ed25519.PrivateKey,
	code, serial string,
	tier license.Tier,
	exp int64,
) string {
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
	if exp > 0 {
		payload["exp"] = exp
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	sig := ed25519.Sign(priv, b)
	return "MSN1." + base64.RawURLEncoding.EncodeToString(b) +
		"." + base64.RawURLEncoding.EncodeToString(sig)
}

// signTestKey is a drop-in replacement for the removed license.GenerateLicenseKey:
// it mints a production-shaped token signed by the TEST key and returns
// (token, nil). It validates through testVerifier and through a Manager built
// with NewManagerWithVerifier(testVerifier(t)).
func signTestKey(t *testing.T, code, serial string, tier license.Tier) (string, error) {
	t.Helper()
	return signLicenseToken(t, testSigningKey(t), code, serial, tier, 0), nil
}

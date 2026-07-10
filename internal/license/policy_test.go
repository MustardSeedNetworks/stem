// SPDX-License-Identifier: BUSL-1.1

package license_test

import (
	"slices"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/license"
)

// Production-signed Stem tokens (serial 1234567), produced by the canonical
// keygen tool against the embedded production public key. They pin the
// cross-tool signing contract (product name "stem", codes 1001/2001, tier
// feature sets) through Stem's product policy, the way niac/seed pin theirs.
// The Enterprise vector (code 3001 / tier 3) is retired and MUST be rejected.
// If the production key rotates, regenerate these vectors and the embedded key
// together. Generic crypto properties (forgery, tampering, wrong-product,
// expiry, bad input) are covered in foundation's pkg/license tests; this file
// only exercises Stem's product-specific wiring.
const (
	prodStemReflectorVector    = "MSN1.eyJ2IjoxLCJwcm9kdWN0Ijoic3RlbSIsImNvZGUiOiIxMDAxIiwic2VyaWFsIjoiMTIzNDU2NyIsInRpZXIiOjEsIm1heERldmljZXMiOjMsImlhdCI6MTc4MDg3NjgwMH0.5SV0MhUNd9_em5B1_noWxJUbZpWjtbgf91BnTqi3uK2GBqeDiy0xdYWR3fDTq1HRRKa1TfDx6MNhufpjIbzhBg"
	prodStemProfessionalVector = "MSN1.eyJ2IjoxLCJwcm9kdWN0Ijoic3RlbSIsImNvZGUiOiIyMDAxIiwic2VyaWFsIjoiMTIzNDU2NyIsInRpZXIiOjIsIm1heERldmljZXMiOjMsImlhdCI6MTc4MDg3NjgwMH0.hPCX0LGKbRfOIRl6CQxUPcCgzlD0BnZloMFIWL_z7NaGmLOtvRoxIvOYjNxQ7uq1rpPMrWEY1dHDnwPtQlNFDA"
	prodStemEnterpriseVector   = "MSN1.eyJ2IjoxLCJwcm9kdWN0Ijoic3RlbSIsImNvZGUiOiIzMDAxIiwic2VyaWFsIjoiMTIzNDU2NyIsInRpZXIiOjMsIm1heERldmljZXMiOjMsImlhdCI6MTc4MDg3NjgwMH0.55Qx3dguyUocqoi_9YXufcIkphPB5kiSV28SueO2wGZ6UK3d_HaGcwpO2hon2k-qX9BBz0QC1mjNC1DeBkWIBQ"
)

func TestTierString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		tier license.Tier
		want string
	}{
		{license.TierReflector, "Reflector"},
		{license.TierProfessional, "Professional"},
		{license.TierInvalid, "Invalid"},
	}
	for _, c := range cases {
		if got := c.tier.String(); got != c.want {
			t.Errorf("Tier(%d).String() = %q, want %q", c.tier, got, c.want)
		}
	}
}

// TestKeygenContract pins the cross-tool signing contract end-to-end: each live
// production vector activates through Stem's policy at the right tier, and the
// retired Enterprise vector (tier 3 / code 3001) is rejected. This catches a
// wrong product code, salt, or embedded key in Stem's policy wiring, and guards
// the Enterprise retirement.
func TestKeygenContract(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		vector   string
		wantTier license.Tier
	}{
		{"reflector", prodStemReflectorVector, license.TierReflector},
		{"professional", prodStemProfessionalVector, license.TierProfessional},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			mgr, err := license.NewManagerWithDir(t.TempDir())
			if err != nil {
				t.Fatalf("NewManagerWithDir: %v", err)
			}
			res := mgr.Activate(c.vector)
			if !res.Success {
				t.Fatalf("production vector did not activate: %s", res.Message)
			}
			if res.Tier != int(c.wantTier) {
				t.Errorf("Tier = %d, want %s (%d)", res.Tier, c.wantTier, c.wantTier)
			}
		})
	}

	t.Run("enterprise-retired", func(t *testing.T) {
		t.Parallel()
		mgr, err := license.NewManagerWithDir(t.TempDir())
		if err != nil {
			t.Fatalf("NewManagerWithDir: %v", err)
		}
		if res := mgr.Activate(prodStemEnterpriseVector); res.Success {
			t.Error("Enterprise prod vector (tier 3) must be rejected after retirement")
		}
	})
}

// TestFeaturesForTier verifies Stem's product policy directly: Reflector and
// Professional map to their catalogs and codes; the unlicensed tier 0, the
// retired Enterprise tier 3, and any unknown tier are rejected so a signed
// token can't grant more than this build knows about.
func TestFeaturesForTier(t *testing.T) {
	t.Parallel()
	p := license.Policy()

	refl, reflCode, ok := p.FeaturesForTier(int(license.TierReflector))
	if !ok {
		t.Fatal("Reflector tier not recognized")
	}
	if reflCode != "1001" || !slices.Equal(refl, []string{"reflector"}) {
		t.Errorf("Reflector = %v / %q, want [reflector] / 1001", refl, reflCode)
	}

	pro, proCode, ok := p.FeaturesForTier(int(license.TierProfessional))
	if !ok {
		t.Fatal("Professional tier not recognized")
	}
	if proCode != "2001" {
		t.Errorf("Professional code = %q, want 2001", proCode)
	}
	if !slices.Contains(pro, "rfc2544") || !slices.Contains(pro, "reflector") {
		t.Errorf("Professional missing expected features: %v", pro)
	}

	for _, tier := range []int{int(license.TierInvalid), 3 /* retired Enterprise */, 99} {
		if _, _, recognized := p.FeaturesForTier(tier); recognized {
			t.Errorf("tier %d unexpectedly recognized", tier)
		}
	}
}

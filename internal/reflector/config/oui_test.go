// SPDX-License-Identifier: BUSL-1.1

package config_test

import (
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

// An unset OUI means "do not filter by OUI", which is the default the daemon
// builds its reflector with. Treating it as a parse failure made the daemon
// unable to create a dataplane at all, so a daemon-owned reflector could
// never start on real hardware.
func TestParseOUIAcceptsAnUnsetFilter(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{"empty": "", "spaces": "   "} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{}
			cfg.Filtering.OUI = value

			oui, err := cfg.ParseOUI()
			if err != nil {
				t.Fatalf("ParseOUI(%q) = %v, want no error for an unset filter", value, err)
			}
			if oui != [3]byte{} {
				t.Errorf("ParseOUI(%q) = %v, want the zero OUI", value, oui)
			}
		})
	}
}

func TestParseOUIReadsAConfiguredFilter(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	cfg.Filtering.OUI = "00:c0:17"

	oui, err := cfg.ParseOUI()
	if err != nil {
		t.Fatalf("ParseOUI: %v", err)
	}
	if want := [3]byte{0x00, 0xc0, 0x17}; oui != want {
		t.Errorf("ParseOUI = %v, want %v", oui, want)
	}
}

// A value the operator meant as a filter but got wrong must still be an
// error — silently reflecting everything would be worse than refusing.
func TestParseOUIRejectsAMalformedFilter(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"not hex":    "zz:zz:zz",
		"too short":  "00:c0",
		"no colons":  "00c017x",
		"trailing g": "00:c0:1g",
		// Even-length but too long: hex decodes fine, so only a length
		// check rejects it. Without one the extra digits are dropped and
		// the filter silently becomes a different OUI.
		"group too long": "0000:c0:17",
		"four groups":    "00:c0:17:aa",
		"two groups":     "00:c0",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{}
			cfg.Filtering.OUI = value
			if _, err := cfg.ParseOUI(); err == nil {
				t.Errorf("ParseOUI(%q) = nil error, want a rejection", value)
			}
		})
	}
}

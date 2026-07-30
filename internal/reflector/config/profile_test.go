// SPDX-License-Identifier: BUSL-1.1

package config_test

import (
	"testing"

	reflectorConfig "github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

func TestProfileSettings(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		mode      string
		port      uint16
	}{
		{
			name:      reflectorConfig.ProfileNetAlly,
			signature: reflectorConfig.ProfileITO,
			mode:      "mac-ip",
			port:      reflectorConfig.NetAllyPort,
		},
		{
			name:      reflectorConfig.ProfileITO,
			signature: reflectorConfig.ProfileITO,
			mode:      "mac-ip",
			port:      reflectorConfig.NetAllyPort,
		},
		{
			name: reflectorConfig.ProfileMSN, signature: reflectorConfig.ProfileMSN,
			mode: reflectorConfig.ProfileAll,
		},
		{
			name: reflectorConfig.ProfileCustom, signature: reflectorConfig.ProfileCustom,
			mode: reflectorConfig.ProfileAll,
		},
		{
			name: reflectorConfig.ProfileAll, signature: reflectorConfig.ProfileAll,
			mode: reflectorConfig.ProfileAll, port: reflectorConfig.NetAllyPort,
		},
	}
	for _, test := range tests {
		settings := reflectorConfig.SettingsForProfile(test.name)
		if settings.SignatureFilter != test.signature || settings.Mode != test.mode || settings.Port != test.port {
			t.Errorf("SettingsForProfile(%q) = %+v", test.name, settings)
		}
	}
}

func TestValidateSignatureFilter(t *testing.T) {
	for _, filter := range []string{"all", "ito", "probeot", "dataot", "latency", "rfc2544", "y1564", "custom", "msn"} {
		if err := reflectorConfig.ValidateSignatureFilter(filter); err != nil {
			t.Errorf("ValidateSignatureFilter(%q) = %v", filter, err)
		}
	}
	if err := reflectorConfig.ValidateSignatureFilter("invalid"); err == nil {
		t.Fatal("ValidateSignatureFilter(invalid) = nil")
	}
}

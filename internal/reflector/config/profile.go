// SPDX-License-Identifier: BUSL-1.1

package config

import "fmt"

// ProfileSettings defines packet handling required by a reflector profile.
type ProfileSettings struct {
	SignatureFilter string
	Mode            string
	Port            uint16
}

// Reflector profiles select the signature filter, MAC/IP echo mode, and
// listening port a peer reflector expects; SettingsForProfile maps each name
// to those settings. ProfileNetAlly and ProfileITO are aliases for the same
// behavior (interop with NetAlly/ITO test sets on the industry-standard port).
const (
	ProfileNetAlly = "netally"
	ProfileITO     = "ito"
	ProfileMSN     = "msn"
	ProfileAll     = filterAll
	ProfileCustom  = "custom"
	// NetAllyPort is the reflector port expected by NetAlly/ITO test sets.
	NetAllyPort = 3842
)

// SettingsForProfile returns consistent CLI and API behavior for a profile.
func SettingsForProfile(name string) ProfileSettings {
	switch name {
	case ProfileNetAlly, ProfileITO:
		return ProfileSettings{SignatureFilter: ProfileITO, Mode: "mac-ip", Port: NetAllyPort}
	case ProfileMSN:
		return ProfileSettings{SignatureFilter: ProfileMSN, Mode: filterAll}
	case ProfileCustom:
		return ProfileSettings{SignatureFilter: ProfileCustom, Mode: filterAll}
	default:
		return ProfileSettings{SignatureFilter: filterAll, Mode: filterAll, Port: NetAllyPort}
	}
}

// ValidateSignatureFilter checks whether the dataplane can represent a filter exactly.
func ValidateSignatureFilter(filter string) error {
	switch filter {
	case filterAll, ProfileITO, "probeot", "dataot", "latency", "rfc2544", "y1564", ProfileCustom, ProfileMSN:
		return nil
	default:
		return fmt.Errorf("invalid signature filter %q", filter)
	}
}

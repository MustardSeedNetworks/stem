// SPDX-License-Identifier: BUSL-1.1

// Package config provides YAML configuration support for the Reflector.
//
// Defines configuration structures for interface settings, signature filtering,
// web UI options, platform-specific settings, and statistics collection.
package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Filter and mode constants.
const (
	filterAll = "all"
	modeMAC   = "mac"
	modeMACIP = "mac-ip"
)

// Reflection mode integer values for C interop.
const (
	reflectModeMAC   = 0
	reflectModeMACIP = 1
	reflectModeAll   = 2
)

// Config holds all reflector configuration.
type Config struct {
	Interface       string         `yaml:"interface"`
	Verbose         bool           `yaml:"verbose"`
	SignatureFilter string         `yaml:"signature_filter"` // all, ito, rfc2544, y1564, msn, custom
	WebUI           WebUIConfig    `yaml:"web_ui"`
	TUI             TUIConfig      `yaml:"tui"`
	Filtering       FilterConfig   `yaml:"filtering"`
	Reflection      ReflectConfig  `yaml:"reflection"`
	Platform        PlatformConfig `yaml:"platform"`
	Stats           StatsConfig    `yaml:"stats"`
}

// WebUIConfig holds web UI settings.
type WebUIConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// TUIConfig holds TUI settings.
type TUIConfig struct {
	Enabled bool `yaml:"enabled"`
}

// FilterConfig holds packet filtering settings.
type FilterConfig struct {
	Port      uint16 `yaml:"port"`       // ITO UDP port (0 = any)
	FilterOUI bool   `yaml:"filter_oui"` // Enable OUI filtering
	OUI       string `yaml:"oui"`        // Source OUI (XX:XX:XX)
	FilterMAC bool   `yaml:"filter_mac"` // Enable destination MAC filtering
}

// ReflectConfig holds reflection mode settings.
type ReflectConfig struct {
	Mode string `yaml:"mode"` // mac, mac-ip, or all
}

// PlatformConfig holds platform-specific settings.
type PlatformConfig struct {
	UseAFXDP bool `yaml:"use_af_xdp"` // Use AF_XDP (default on Linux)
}

// StatsConfig holds statistics settings.
type StatsConfig struct {
	Format   string `yaml:"format"`   // text, json, csv
	Interval int    `yaml:"interval"` // seconds
}

// LoadFile loads configuration from a YAML file.
func LoadFile(path string) (*Config, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		Interface:       "",
		Verbose:         false,
		SignatureFilter: "",
		WebUI:           WebUIConfig{Enabled: false, Port: 0},
		TUI:             TUIConfig{Enabled: false},
		Filtering:       FilterConfig{Port: 0, FilterOUI: false, OUI: "", FilterMAC: false},
		Reflection:      ReflectConfig{Mode: ""},
		Platform:        PlatformConfig{UseAFXDP: true},
		Stats:           StatsConfig{Format: "", Interval: 0},
	}
	unmarshalErr := yaml.Unmarshal(data, cfg)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", unmarshalErr)
	}

	// Apply defaults.
	cfg.applyDefaults()

	return cfg, nil
}

// applyDefaults sets default values for unspecified fields.
func (c *Config) applyDefaults() {
	if c.WebUI.Port == 0 {
		c.WebUI.Port = 8444
	}
	if c.SignatureFilter == "" {
		c.SignatureFilter = filterAll
	}
	if c.Filtering.Port == 0 {
		c.Filtering.Port = 3842
	}
	if c.Filtering.OUI == "" {
		c.Filtering.OUI = "00:c0:17"
	}
	if c.Reflection.Mode == "" {
		c.Reflection.Mode = filterAll
	}
	if c.Stats.Format == "" {
		c.Stats.Format = "text"
	}
	if c.Stats.Interval == 0 {
		c.Stats.Interval = 10
	}
	// TUI enabled by default.
	if !c.TUI.Enabled && c.Interface != "" {
		c.TUI.Enabled = true
	}
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Interface == "" {
		return errors.New("interface is required")
	}

	// Validate OUI format (XX:XX:XX).
	ouiPattern := regexp.MustCompile(`^[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}$`)
	if c.Filtering.FilterOUI && !ouiPattern.MatchString(c.Filtering.OUI) {
		return fmt.Errorf("invalid OUI format: %s (expected XX:XX:XX)", c.Filtering.OUI)
	}

	// Validate reflection mode.
	switch c.Reflection.Mode {
	case modeMAC, modeMACIP, filterAll:
		// Valid.
	default:
		return fmt.Errorf("invalid reflection mode: %s (expected mac, mac-ip, or all)", c.Reflection.Mode)
	}

	// Validate stats format.
	switch c.Stats.Format {
	case "text", "json", "csv":
		// Valid.
	default:
		return fmt.Errorf("invalid stats format: %s (expected text, json, or csv)", c.Stats.Format)
	}

	if c.WebUI.Port < 1 || c.WebUI.Port > 65535 {
		return fmt.Errorf("invalid web port: %d", c.WebUI.Port)
	}

	return nil
}

// ouiGroups is the number of colon-separated bytes in an OUI, and
// ouiGroupDigits the hex digits in each.
const (
	ouiGroups      = 3
	ouiGroupDigits = 2
)

// errMalformedOUI reports a value meant as an OUI filter that is not one.
var errMalformedOUI = errors.New("malformed OUI filter")

// ParseOUI parses the OUI string into bytes. An unset value means "do not
// filter by OUI" and yields the zero OUI, not an error: that is the default
// the daemon builds its reflector with, and treating it as a parse failure
// stopped a daemon-owned reflector from ever creating a dataplane. A value
// that was meant as a filter but is malformed is still an error — silently
// reflecting everything would be worse than refusing to start.
func (c *Config) ParseOUI() ([3]byte, error) {
	var oui [3]byte
	if strings.TrimSpace(c.Filtering.OUI) == "" {
		return oui, nil
	}
	// Sscanf("%02x:%02x:%02x") accepts a single hex digit per group and
	// ignores trailing rubbish, so "00:c0:1g" parsed as a different filter
	// than the operator wrote. Decode the whole value instead.
	groups := strings.Split(c.Filtering.OUI, ":")
	if len(groups) != ouiGroups {
		return oui, fmt.Errorf("%w: %q", errMalformedOUI, c.Filtering.OUI)
	}
	for i, group := range groups {
		if len(group) != ouiGroupDigits {
			return oui, fmt.Errorf("%w: %q", errMalformedOUI, c.Filtering.OUI)
		}
		decoded, err := hex.DecodeString(group)
		if err != nil {
			return oui, fmt.Errorf("%w: %q", errMalformedOUI, c.Filtering.OUI)
		}
		oui[i] = decoded[0]
	}
	return oui, nil
}

// ReflectModeInt converts the mode string to an int for C.
func (c *Config) ReflectModeInt() int {
	switch c.Reflection.Mode {
	case modeMAC:
		return reflectModeMAC
	case modeMACIP:
		return reflectModeMACIP
	case filterAll:
		return reflectModeAll
	default:
		return reflectModeAll
	}
}

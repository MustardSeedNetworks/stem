// SPDX-License-Identifier: BUSL-1.1

// Package config provides YAML configuration support for the Reflector.
//
// Defines configuration structures for interface settings, signature filtering,
// web UI options, platform-specific settings, and statistics collection.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

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
	UseDPDK  bool   `yaml:"use_dpdk"`
	UseAFXDP bool   `yaml:"use_af_xdp"` // Use AF_XDP (default on Linux)
	DPDKArgs string `yaml:"dpdk_args"`
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
		Platform:        PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
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
		c.WebUI.Port = 8080
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

// ParseOUI parses the OUI string into bytes.
func (c *Config) ParseOUI() ([3]byte, error) {
	var oui [3]byte
	_, err := fmt.Sscanf(c.Filtering.OUI, "%02x:%02x:%02x", &oui[0], &oui[1], &oui[2])
	if err != nil {
		return oui, fmt.Errorf("failed to parse OUI: %w", err)
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

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Interface:       "eth0",
		Verbose:         true,
		SignatureFilter: "all",
	}

	if cfg.Interface != "eth0" {
		t.Errorf("Expected Interface 'eth0', got '%s'", cfg.Interface)
	}
	if !cfg.Verbose {
		t.Error("Expected Verbose true")
	}
	if cfg.SignatureFilter != "all" {
		t.Errorf("Expected SignatureFilter 'all', got '%s'", cfg.SignatureFilter)
	}
}

func TestWebUIConfig(t *testing.T) {
	cfg := WebUIConfig{
		Enabled: true,
		Port:    8080,
	}

	if !cfg.Enabled {
		t.Error("Expected Enabled true")
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", cfg.Port)
	}
}

func TestFilterConfig(t *testing.T) {
	cfg := FilterConfig{
		Port:      3842,
		FilterOUI: true,
		OUI:       "00:c0:17",
		FilterMAC: false,
	}

	if cfg.Port != 3842 {
		t.Errorf("Expected Port 3842, got %d", cfg.Port)
	}
	if !cfg.FilterOUI {
		t.Error("Expected FilterOUI true")
	}
	if cfg.OUI != "00:c0:17" {
		t.Errorf("Expected OUI '00:c0:17', got '%s'", cfg.OUI)
	}
}

func TestReflectConfig(t *testing.T) {
	modes := []string{"mac", "mac-ip", "all"}
	for _, mode := range modes {
		cfg := ReflectConfig{Mode: mode}
		if cfg.Mode != mode {
			t.Errorf("Expected Mode '%s', got '%s'", mode, cfg.Mode)
		}
	}
}

func TestPlatformConfig(t *testing.T) {
	cfg := PlatformConfig{
		UseDPDK:  true,
		UseAFXDP: false,
		DPDKArgs: "-l 0-3 -n 4",
	}

	if !cfg.UseDPDK {
		t.Error("Expected UseDPDK true")
	}
	if cfg.UseAFXDP {
		t.Error("Expected UseAFXDP false")
	}
	if cfg.DPDKArgs != "-l 0-3 -n 4" {
		t.Errorf("Expected DPDKArgs '-l 0-3 -n 4', got '%s'", cfg.DPDKArgs)
	}
}

func TestStatsConfig(t *testing.T) {
	formats := []string{"text", "json", "csv"}
	for _, format := range formats {
		cfg := StatsConfig{Format: format, Interval: 10}
		if cfg.Format != format {
			t.Errorf("Expected Format '%s', got '%s'", format, cfg.Format)
		}
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{Interface: "eth0"}
	cfg.applyDefaults()

	if cfg.WebUI.Port != 8080 {
		t.Errorf("Expected default WebUI.Port 8080, got %d", cfg.WebUI.Port)
	}
	if cfg.SignatureFilter != "all" {
		t.Errorf("Expected default SignatureFilter 'all', got '%s'", cfg.SignatureFilter)
	}
	if cfg.Filtering.Port != 3842 {
		t.Errorf("Expected default Filtering.Port 3842, got %d", cfg.Filtering.Port)
	}
	if cfg.Filtering.OUI != "00:c0:17" {
		t.Errorf("Expected default Filtering.OUI '00:c0:17', got '%s'", cfg.Filtering.OUI)
	}
	if cfg.Reflection.Mode != "all" {
		t.Errorf("Expected default Reflection.Mode 'all', got '%s'", cfg.Reflection.Mode)
	}
	if cfg.Stats.Format != "text" {
		t.Errorf("Expected default Stats.Format 'text', got '%s'", cfg.Stats.Format)
	}
	if cfg.Stats.Interval != 10 {
		t.Errorf("Expected default Stats.Interval 10, got %d", cfg.Stats.Interval)
	}
}

func TestApplyDefaultsDoesNotOverride(t *testing.T) {
	cfg := &Config{
		Interface:       "eth0",
		SignatureFilter: "custom",
		WebUI:           WebUIConfig{Port: 9090},
		Filtering:       FilterConfig{Port: 5000, OUI: "11:22:33"},
		Reflection:      ReflectConfig{Mode: "mac"},
		Stats:           StatsConfig{Format: "json", Interval: 30},
	}
	cfg.applyDefaults()

	if cfg.WebUI.Port != 9090 {
		t.Errorf("WebUI.Port should not be overwritten, expected 9090, got %d", cfg.WebUI.Port)
	}
	if cfg.SignatureFilter != "custom" {
		t.Errorf("SignatureFilter should not be overwritten, expected 'custom', got '%s'", cfg.SignatureFilter)
	}
	if cfg.Filtering.Port != 5000 {
		t.Errorf("Filtering.Port should not be overwritten, expected 5000, got %d", cfg.Filtering.Port)
	}
	if cfg.Reflection.Mode != "mac" {
		t.Errorf("Reflection.Mode should not be overwritten, expected 'mac', got '%s'", cfg.Reflection.Mode)
	}
	if cfg.Stats.Format != "json" {
		t.Errorf("Stats.Format should not be overwritten, expected 'json', got '%s'", cfg.Stats.Format)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty interface",
			cfg:     Config{},
			wantErr: true,
			errMsg:  "interface is required",
		},
		{
			name: "invalid OUI format",
			cfg: Config{
				Interface: "eth0",
				Filtering: FilterConfig{FilterOUI: true, OUI: "invalid"},
			},
			wantErr: true,
			errMsg:  "invalid OUI format",
		},
		{
			name: "invalid reflection mode",
			cfg: Config{
				Interface:  "eth0",
				Filtering:  FilterConfig{OUI: "00:c0:17"},
				Reflection: ReflectConfig{Mode: "invalid"},
				Stats:      StatsConfig{Format: "text"},
				WebUI:      WebUIConfig{Port: 8080},
			},
			wantErr: true,
			errMsg:  "invalid reflection mode",
		},
		{
			name: "invalid stats format",
			cfg: Config{
				Interface:  "eth0",
				Filtering:  FilterConfig{OUI: "00:c0:17"},
				Reflection: ReflectConfig{Mode: "all"},
				Stats:      StatsConfig{Format: "invalid"},
				WebUI:      WebUIConfig{Port: 8080},
			},
			wantErr: true,
			errMsg:  "invalid stats format",
		},
		{
			name: "invalid web port - zero",
			cfg: Config{
				Interface:  "eth0",
				Filtering:  FilterConfig{OUI: "00:c0:17"},
				Reflection: ReflectConfig{Mode: "all"},
				Stats:      StatsConfig{Format: "text"},
				WebUI:      WebUIConfig{Port: 0},
			},
			wantErr: true,
			errMsg:  "invalid web port",
		},
		{
			name: "invalid web port - too high",
			cfg: Config{
				Interface:  "eth0",
				Filtering:  FilterConfig{OUI: "00:c0:17"},
				Reflection: ReflectConfig{Mode: "all"},
				Stats:      StatsConfig{Format: "text"},
				WebUI:      WebUIConfig{Port: 70000},
			},
			wantErr: true,
			errMsg:  "invalid web port",
		},
		{
			name: "valid config",
			cfg: Config{
				Interface:  "eth0",
				Filtering:  FilterConfig{OUI: "00:c0:17"},
				Reflection: ReflectConfig{Mode: "all"},
				Stats:      StatsConfig{Format: "text"},
				WebUI:      WebUIConfig{Port: 8080},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseOUI(t *testing.T) {
	tests := []struct {
		name    string
		oui     string
		want    [3]byte
		wantErr bool
	}{
		{
			name:    "valid OUI lowercase",
			oui:     "00:c0:17",
			want:    [3]byte{0x00, 0xc0, 0x17},
			wantErr: false,
		},
		{
			name:    "valid OUI uppercase",
			oui:     "AA:BB:CC",
			want:    [3]byte{0xAA, 0xBB, 0xCC},
			wantErr: false,
		},
		{
			name:    "valid OUI mixed case",
			oui:     "Ff:Ee:Dd",
			want:    [3]byte{0xFF, 0xEE, 0xDD},
			wantErr: false,
		},
		{
			name:    "invalid OUI format",
			oui:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Filtering: FilterConfig{OUI: tt.oui}}
			got, err := cfg.ParseOUI()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOUI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseOUI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReflectModeInt(t *testing.T) {
	tests := []struct {
		mode string
		want int
	}{
		{"mac", 0},
		{"mac-ip", 1},
		{"all", 2},
		{"unknown", 2}, // Default to 2
		{"", 2},        // Default to 2
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			cfg := &Config{Reflection: ReflectConfig{Mode: tt.mode}}
			got := cfg.ReflectModeInt()
			if got != tt.want {
				t.Errorf("ReflectModeInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
interface: eth0
verbose: true
signature_filter: all
web_ui:
  enabled: true
  port: 8080
filtering:
  port: 3842
  filter_oui: true
  oui: "00:c0:17"
reflection:
  mode: all
stats:
  format: text
  interval: 10
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cfg, err := LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.Interface != "eth0" {
		t.Errorf("Expected Interface 'eth0', got '%s'", cfg.Interface)
	}
	if !cfg.Verbose {
		t.Error("Expected Verbose true")
	}
	if cfg.WebUI.Port != 8080 {
		t.Errorf("Expected WebUI.Port 8080, got %d", cfg.WebUI.Port)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(tmpFile, []byte("invalid: yaml: content: [}"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := LoadFile(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestTUIConfig(t *testing.T) {
	cfg := TUIConfig{Enabled: true}
	if !cfg.Enabled {
		t.Error("Expected TUI Enabled true")
	}
}

// Benchmark tests
func BenchmarkValidate(b *testing.B) {
	cfg := Config{
		Interface:  "eth0",
		Filtering:  FilterConfig{OUI: "00:c0:17"},
		Reflection: ReflectConfig{Mode: "all"},
		Stats:      StatsConfig{Format: "text"},
		WebUI:      WebUIConfig{Port: 8080},
	}

	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

func BenchmarkParseOUI(b *testing.B) {
	cfg := &Config{Filtering: FilterConfig{OUI: "00:c0:17"}}

	for i := 0; i < b.N; i++ {
		_, _ = cfg.ParseOUI()
	}
}

func BenchmarkApplyDefaults(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := &Config{Interface: "eth0"}
		cfg.applyDefaults()
	}
}

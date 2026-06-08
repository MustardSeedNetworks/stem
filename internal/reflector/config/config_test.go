// SPDX-License-Identifier: BUSL-1.1

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/reflector/config"
)

// Test constants for repeated strings.
const (
	testIfaceEth0       = "eth0"
	testFilterAll       = "all"
	testDefaultOUI      = "00:c0:17"
	testStatsFormatText = "text"
	testReflectModeMAC  = "mac"
)

func TestConfigStruct(t *testing.T) {
	cfg := config.Config{
		Interface:       testIfaceEth0,
		Verbose:         true,
		SignatureFilter: testFilterAll,
		WebUI:           config.WebUIConfig{Enabled: false, Port: 0},
		TUI:             config.TUIConfig{Enabled: false},
		Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "", FilterMAC: false},
		Reflection:      config.ReflectConfig{Mode: ""},
		Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:           config.StatsConfig{Format: "", Interval: 0},
	}

	if cfg.Interface != testIfaceEth0 {
		t.Errorf("Expected Interface '%s', got '%s'", testIfaceEth0, cfg.Interface)
	}
	if !cfg.Verbose {
		t.Error("Expected Verbose true")
	}
	if cfg.SignatureFilter != testFilterAll {
		t.Errorf("Expected SignatureFilter '%s', got '%s'", testFilterAll, cfg.SignatureFilter)
	}
}

func TestWebUIConfig(t *testing.T) {
	cfg := config.WebUIConfig{
		Enabled: true,
		Port:    8444,
	}

	if !cfg.Enabled {
		t.Error("Expected Enabled true")
	}
	if cfg.Port != 8444 {
		t.Errorf("Expected Port 8444, got %d", cfg.Port)
	}
}

func TestFilterConfig(t *testing.T) {
	cfg := config.FilterConfig{
		Port:      3842,
		FilterOUI: true,
		OUI:       testDefaultOUI,
		FilterMAC: false,
	}

	if cfg.Port != 3842 {
		t.Errorf("Expected Port 3842, got %d", cfg.Port)
	}
	if !cfg.FilterOUI {
		t.Error("Expected FilterOUI true")
	}
	if cfg.OUI != testDefaultOUI {
		t.Errorf("Expected OUI '%s', got '%s'", testDefaultOUI, cfg.OUI)
	}
	if cfg.FilterMAC {
		t.Error("Expected FilterMAC false")
	}
}

func TestReflectConfig(t *testing.T) {
	modes := []string{"mac", "mac-ip", "all"}
	for _, mode := range modes {
		cfg := config.ReflectConfig{Mode: mode}
		if cfg.Mode != mode {
			t.Errorf("Expected Mode '%s', got '%s'", mode, cfg.Mode)
		}
	}
}

func TestPlatformConfig(t *testing.T) {
	cfg := config.PlatformConfig{
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
		cfg := config.StatsConfig{Format: format, Interval: 10}
		if cfg.Format != format {
			t.Errorf("Expected Format '%s', got '%s'", format, cfg.Format)
		}
		if cfg.Interval != 10 {
			t.Errorf("Expected Interval 10, got %d", cfg.Interval)
		}
	}
}

// TestApplyDefaultsViaLoadFile tests defaults applied through LoadFile.
func TestApplyDefaultsViaLoadFile(t *testing.T) {
	// Create a minimal YAML file with only interface set.
	yamlContent := `interface: eth0
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "minimal.yaml")
	writeErr := os.WriteFile(tmpFile, []byte(yamlContent), 0o600)
	if writeErr != nil {
		t.Fatalf("Failed to create temp file: %v", writeErr)
	}

	cfg, err := config.LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.WebUI.Port != 8444 {
		t.Errorf("Expected default WebUI.Port 8444, got %d", cfg.WebUI.Port)
	}
	if cfg.SignatureFilter != testFilterAll {
		t.Errorf("Expected default SignatureFilter '%s', got '%s'", testFilterAll, cfg.SignatureFilter)
	}
	if cfg.Filtering.Port != 3842 {
		t.Errorf("Expected default Filtering.Port 3842, got %d", cfg.Filtering.Port)
	}
	if cfg.Filtering.OUI != testDefaultOUI {
		t.Errorf("Expected default Filtering.OUI '%s', got '%s'", testDefaultOUI, cfg.Filtering.OUI)
	}
	if cfg.Reflection.Mode != testFilterAll {
		t.Errorf("Expected default Reflection.Mode '%s', got '%s'", testFilterAll, cfg.Reflection.Mode)
	}
	if cfg.Stats.Format != testStatsFormatText {
		t.Errorf("Expected default Stats.Format '%s', got '%s'", testStatsFormatText, cfg.Stats.Format)
	}
	if cfg.Stats.Interval != 10 {
		t.Errorf("Expected default Stats.Interval 10, got %d", cfg.Stats.Interval)
	}
}

// TestApplyDefaultsDoesNotOverrideViaLoadFile tests that defaults don't override explicit values.
func TestApplyDefaultsDoesNotOverrideViaLoadFile(t *testing.T) {
	// Create a YAML file with explicit non-default values.
	yamlContent := `interface: eth0
signature_filter: custom
web_ui:
  port: 9090
filtering:
  port: 5000
  oui: "11:22:33"
reflection:
  mode: mac
stats:
  format: json
  interval: 30
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "custom.yaml")
	writeErr := os.WriteFile(tmpFile, []byte(yamlContent), 0o600)
	if writeErr != nil {
		t.Fatalf("Failed to create temp file: %v", writeErr)
	}

	cfg, err := config.LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.WebUI.Port != 9090 {
		t.Errorf("WebUI.Port should not be overwritten, expected 9090, got %d", cfg.WebUI.Port)
	}
	if cfg.SignatureFilter != "custom" {
		t.Errorf("SignatureFilter should not be overwritten, expected 'custom', got '%s'", cfg.SignatureFilter)
	}
	if cfg.Filtering.Port != 5000 {
		t.Errorf("Filtering.Port should not be overwritten, expected 5000, got %d", cfg.Filtering.Port)
	}
	if cfg.Reflection.Mode != testReflectModeMAC {
		t.Errorf("Reflection.Mode should not be overwritten, expected '%s', got '%s'",
			testReflectModeMAC, cfg.Reflection.Mode)
	}
	if cfg.Stats.Format != "json" {
		t.Errorf("Stats.Format should not be overwritten, expected 'json', got '%s'", cfg.Stats.Format)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty interface",
			cfg: config.Config{
				Interface:       "",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 0},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: ""},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "", Interval: 0},
			},
			wantErr: true,
			errMsg:  "interface is required",
		},
		{
			name: "invalid OUI format",
			cfg: config.Config{
				Interface:       "eth0",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 8444},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: true, OUI: "invalid", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: "all"},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "text", Interval: 0},
			},
			wantErr: true,
			errMsg:  "invalid OUI format",
		},
		{
			name: "invalid reflection mode",
			cfg: config.Config{
				Interface:       "eth0",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 8444},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: "invalid"},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "text", Interval: 0},
			},
			wantErr: true,
			errMsg:  "invalid reflection mode",
		},
		{
			name: "invalid stats format",
			cfg: config.Config{
				Interface:       "eth0",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 8444},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: "all"},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "invalid", Interval: 0},
			},
			wantErr: true,
			errMsg:  "invalid stats format",
		},
		{
			name: "invalid web port - zero",
			cfg: config.Config{
				Interface:       "eth0",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 0},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: "all"},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "text", Interval: 0},
			},
			wantErr: true,
			errMsg:  "invalid web port",
		},
		{
			name: "invalid web port - too high",
			cfg: config.Config{
				Interface:       "eth0",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 70000},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: "all"},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "text", Interval: 0},
			},
			wantErr: true,
			errMsg:  "invalid web port",
		},
		{
			name: "valid config",
			cfg: config.Config{
				Interface:       "eth0",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 8444},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: "all"},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "text", Interval: 0},
			},
			wantErr: false,
			errMsg:  "",
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
			want:    [3]byte{0x00, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Interface:       "",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 0},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: tt.oui, FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: ""},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "", Interval: 0},
			}
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
		{"unknown", 2}, // Default to 2.
		{"", 2},        // Default to 2.
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			cfg := &config.Config{
				Interface:       "",
				Verbose:         false,
				SignatureFilter: "",
				WebUI:           config.WebUIConfig{Enabled: false, Port: 0},
				TUI:             config.TUIConfig{Enabled: false},
				Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "", FilterMAC: false},
				Reflection:      config.ReflectConfig{Mode: tt.mode},
				Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
				Stats:           config.StatsConfig{Format: "", Interval: 0},
			}
			got := cfg.ReflectModeInt()
			if got != tt.want {
				t.Errorf("ReflectModeInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	// Create a temporary YAML file.
	yamlContent := `
interface: eth0
verbose: true
signature_filter: all
web_ui:
  enabled: true
  port: 8444
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
	writeErr := os.WriteFile(tmpFile, []byte(yamlContent), 0o600)
	if writeErr != nil {
		t.Fatalf("Failed to create temp file: %v", writeErr)
	}

	cfg, err := config.LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.Interface != "eth0" {
		t.Errorf("Expected Interface 'eth0', got '%s'", cfg.Interface)
	}
	if !cfg.Verbose {
		t.Error("Expected Verbose true")
	}
	if cfg.WebUI.Port != 8444 {
		t.Errorf("Expected WebUI.Port 8444, got %d", cfg.WebUI.Port)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := config.LoadFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	writeErr := os.WriteFile(tmpFile, []byte("invalid: yaml: content: [}"), 0o600)
	if writeErr != nil {
		t.Fatalf("Failed to create temp file: %v", writeErr)
	}

	_, err := config.LoadFile(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestTUIConfig(t *testing.T) {
	cfg := config.TUIConfig{Enabled: true}
	if !cfg.Enabled {
		t.Error("Expected TUI Enabled true")
	}
}

// Benchmark tests.
func BenchmarkValidate(b *testing.B) {
	cfg := config.Config{
		Interface:       "eth0",
		Verbose:         false,
		SignatureFilter: "",
		WebUI:           config.WebUIConfig{Enabled: false, Port: 8444},
		TUI:             config.TUIConfig{Enabled: false},
		Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
		Reflection:      config.ReflectConfig{Mode: "all"},
		Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:           config.StatsConfig{Format: "text", Interval: 0},
	}

	for b.Loop() {
		_ = cfg.Validate()
	}
}

func BenchmarkParseOUI(b *testing.B) {
	cfg := &config.Config{
		Interface:       "",
		Verbose:         false,
		SignatureFilter: "",
		WebUI:           config.WebUIConfig{Enabled: false, Port: 0},
		TUI:             config.TUIConfig{Enabled: false},
		Filtering:       config.FilterConfig{Port: 0, FilterOUI: false, OUI: "00:c0:17", FilterMAC: false},
		Reflection:      config.ReflectConfig{Mode: ""},
		Platform:        config.PlatformConfig{UseDPDK: false, UseAFXDP: false, DPDKArgs: ""},
		Stats:           config.StatsConfig{Format: "", Interval: 0},
	}

	for b.Loop() {
		_, _ = cfg.ParseOUI()
	}
}

// BenchmarkApplyDefaultsViaLoadFile benchmarks defaults via LoadFile.
func BenchmarkApplyDefaultsViaLoadFile(b *testing.B) {
	// Create a minimal YAML file.
	yamlContent := `interface: eth0
`
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "minimal.yaml")
	writeErr := os.WriteFile(tmpFile, []byte(yamlContent), 0o600)
	if writeErr != nil {
		b.Fatalf("Failed to create temp file: %v", writeErr)
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = config.LoadFile(tmpFile)
	}
}

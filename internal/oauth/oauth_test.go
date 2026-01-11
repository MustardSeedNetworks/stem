// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

package oauth_test

import (
	"strings"
	"testing"

	"github.com/krisarmstrong/stem/internal/oauth"
)

// TestNewManager tests OAuth manager creation.
func TestNewManager(t *testing.T) {
	manager := oauth.NewManager()
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	providers := manager.ListProviders()
	if len(providers) != 0 {
		t.Errorf("Expected empty providers, got %d", len(providers))
	}
}

// TestRegisterProvider tests provider registration.
func TestRegisterProvider(t *testing.T) {
	manager := oauth.NewManager()

	provider := oauth.NewGoogleProvider("client-id", "client-secret", "https://example.com/callback", nil)
	manager.RegisterProvider("google", provider)

	providers := manager.ListProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}

	retrieved, err := manager.GetProvider("google")
	if err != nil {
		t.Fatalf("GetProvider() error: %v", err)
	}

	if retrieved.Name != "google" {
		t.Errorf("Expected provider name 'google', got '%s'", retrieved.Name)
	}
}

// TestGetProviderCaseInsensitive tests case-insensitive provider lookup.
func TestGetProviderCaseInsensitive(t *testing.T) {
	manager := oauth.NewManager()

	provider := oauth.NewGoogleProvider("client-id", "client-secret", "https://example.com/callback", nil)
	manager.RegisterProvider("Google", provider)

	// Should find regardless of case
	tests := []string{"google", "Google", "GOOGLE", "gOoGlE"}
	for _, name := range tests {
		_, err := manager.GetProvider(name)
		if err != nil {
			t.Errorf("GetProvider(%s) error: %v", name, err)
		}
	}
}

// TestGetProviderNotFound tests error on unknown provider.
func TestGetProviderNotFound(t *testing.T) {
	manager := oauth.NewManager()

	_, err := manager.GetProvider("unknown")
	if err == nil {
		t.Error("Expected error for unknown provider")
	}
}

// TestNewGoogleProvider tests Google provider creation.
func TestNewGoogleProvider(t *testing.T) {
	provider := oauth.NewGoogleProvider(
		"client-id",
		"client-secret",
		"https://example.com/callback",
		nil,
	)

	if provider.Name != "google" {
		t.Errorf("Expected name 'google', got '%s'", provider.Name)
	}

	if provider.Config.ClientID != "client-id" {
		t.Errorf("Expected ClientID 'client-id', got '%s'", provider.Config.ClientID)
	}

	if provider.UserInfoURL != oauth.GoogleUserInfoURL {
		t.Errorf("Expected UserInfoURL '%s', got '%s'", oauth.GoogleUserInfoURL, provider.UserInfoURL)
	}

	// Should have default scopes
	if len(provider.Config.Scopes) != 3 {
		t.Errorf("Expected 3 default scopes, got %d", len(provider.Config.Scopes))
	}
}

// TestNewMicrosoftProvider tests Microsoft provider creation.
func TestNewMicrosoftProvider(t *testing.T) {
	provider := oauth.NewMicrosoftProvider(
		"client-id",
		"client-secret",
		"https://example.com/callback",
		oauth.MicrosoftTenantCommon,
		nil,
	)

	if provider.Name != "microsoft" {
		t.Errorf("Expected name 'microsoft', got '%s'", provider.Name)
	}

	if provider.UserInfoURL != oauth.MicrosoftUserInfoURL {
		t.Errorf("Expected UserInfoURL '%s', got '%s'", oauth.MicrosoftUserInfoURL, provider.UserInfoURL)
	}
}

// TestNewGitHubProvider tests GitHub provider creation.
func TestNewGitHubProvider(t *testing.T) {
	provider := oauth.NewGitHubProvider(
		"client-id",
		"client-secret",
		"https://example.com/callback",
		nil,
	)

	if provider.Name != "github" {
		t.Errorf("Expected name 'github', got '%s'", provider.Name)
	}

	if provider.UserInfoURL != oauth.GitHubUserInfoURL {
		t.Errorf("Expected UserInfoURL '%s', got '%s'", oauth.GitHubUserInfoURL, provider.UserInfoURL)
	}

	// Should have default scopes
	if len(provider.Config.Scopes) != 2 {
		t.Errorf("Expected 2 default scopes, got %d", len(provider.Config.Scopes))
	}
}

// TestGenerateState tests state generation.
func TestGenerateState(t *testing.T) {
	state1, err := oauth.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error: %v", err)
	}

	if len(state1) == 0 {
		t.Error("Expected non-empty state")
	}

	// Should generate unique states
	state2, err := oauth.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error: %v", err)
	}

	if state1 == state2 {
		t.Error("Expected unique states")
	}
}

// TestValidateState tests state validation.
func TestValidateState(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		wantErr  bool
	}{
		{"matching states", "abc123", "abc123", false},
		{"mismatched states", "abc123", "xyz789", true},
		{"empty expected", "", "abc123", true},
		{"empty actual", "abc123", "", true},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := oauth.ValidateState(tt.expected, tt.actual)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetAuthURL tests auth URL generation.
func TestGetAuthURL(t *testing.T) {
	provider := oauth.NewGoogleProvider(
		"client-id",
		"client-secret",
		"https://example.com/callback",
		nil,
	)

	state := "test-state"
	url := provider.GetAuthURL(state)

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// URL should contain state parameter
	if !strings.Contains(url, "state=test-state") {
		t.Errorf("Expected URL to contain state parameter: %s", url)
	}
}

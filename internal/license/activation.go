// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

// License activation and persistence for Seed Test Suite.
// Handles license key validation, device binding, trial mode, and secure storage.
package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	// License file locations
	licenseFileName = ".seed-license"

	// Grace periods
	TrialDays       = 14 // Days before license required
	OfflineMaxDays  = 90 // Days allowed offline after activation
	CheckInInterval = 30 // Days between optional check-ins

	// Encryption key derivation salt
	encryptionSalt = "MSN-SEED-2024-LICENSE"
)

// ActivationState represents the current license activation status
type ActivationState struct {
	LicenseKey      string    `json:"licenseKey"`
	DeviceHash      string    `json:"deviceHash"`
	Tier            Tier      `json:"tier"`
	ActivatedAt     time.Time `json:"activatedAt"`
	LastValidatedAt time.Time `json:"lastValidatedAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	TrialStartedAt  time.Time `json:"trialStartedAt,omitempty"`
	IsTrialMode     bool      `json:"isTrialMode"`
	Features        []string  `json:"features"`
}

// ActivationResult contains the result of an activation attempt
type ActivationResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Tier          Tier   `json:"tier,omitempty"`
	DaysRemaining int    `json:"daysRemaining,omitempty"`
	IsTrialMode   bool   `json:"isTrialMode"`
}

// Manager handles license activation and validation
type Manager struct {
	state       *ActivationState
	fingerprint *DeviceFingerprint
	configDir   string
}

// NewManager creates a new license manager
func NewManager() (*Manager, error) {
	m := &Manager{}

	// Get device fingerprint
	fp, err := GenerateFingerprint()
	if err != nil {
		return nil, fmt.Errorf("failed to generate fingerprint: %w", err)
	}
	m.fingerprint = fp

	// Determine config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	m.configDir = filepath.Join(homeDir, ".config", "seed-test-suite")

	// Load existing state
	m.loadState()

	return m, nil
}

// GetState returns the current activation state
func (m *Manager) GetState() *ActivationState {
	return m.state
}

// GetFingerprint returns the device fingerprint
func (m *Manager) GetFingerprint() *DeviceFingerprint {
	return m.fingerprint
}

// IsActivated returns true if a valid license is active
func (m *Manager) IsActivated() bool {
	if m.state == nil {
		return false
	}

	// Check if still valid
	if m.state.IsTrialMode {
		return m.IsTrialValid()
	}

	// Check expiration
	if !m.state.ExpiresAt.IsZero() && time.Now().After(m.state.ExpiresAt) {
		return false
	}

	// Check device binding
	if m.state.DeviceHash != m.fingerprint.Hash() {
		return false
	}

	return true
}

// IsTrialValid returns true if trial period is still active
func (m *Manager) IsTrialValid() bool {
	if m.state == nil || !m.state.IsTrialMode {
		return false
	}

	if m.state.TrialStartedAt.IsZero() {
		return false
	}

	trialEnd := m.state.TrialStartedAt.AddDate(0, 0, TrialDays)
	return time.Now().Before(trialEnd)
}

// TrialDaysRemaining returns days left in trial
func (m *Manager) TrialDaysRemaining() int {
	if m.state == nil || !m.state.IsTrialMode {
		return 0
	}

	if m.state.TrialStartedAt.IsZero() {
		return TrialDays
	}

	trialEnd := m.state.TrialStartedAt.AddDate(0, 0, TrialDays)
	remaining := int(time.Until(trialEnd).Hours() / 24)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// StartTrial begins the trial period
func (m *Manager) StartTrial() *ActivationResult {
	// Check if already activated
	if m.IsActivated() && !m.state.IsTrialMode {
		return &ActivationResult{
			Success: true,
			Message: "Already activated with full license",
			Tier:    m.state.Tier,
		}
	}

	// Check if trial already started
	if m.state != nil && !m.state.TrialStartedAt.IsZero() {
		remaining := m.TrialDaysRemaining()
		if remaining <= 0 {
			return &ActivationResult{
				Success:     false,
				Message:     "Trial period has expired. Please enter a license key.",
				IsTrialMode: true,
			}
		}
		return &ActivationResult{
			Success:       true,
			Message:       fmt.Sprintf("Trial active: %d days remaining", remaining),
			Tier:          TierTestSuite, // Full features during trial
			DaysRemaining: remaining,
			IsTrialMode:   true,
		}
	}

	// Start new trial
	m.state = &ActivationState{
		DeviceHash:     m.fingerprint.Hash(),
		Tier:           TierTestSuite, // Full features during trial
		TrialStartedAt: time.Now(),
		IsTrialMode:    true,
		Features:       []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"},
	}

	if err := m.saveState(); err != nil {
		return &ActivationResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save trial state: %v", err),
		}
	}

	return &ActivationResult{
		Success:       true,
		Message:       fmt.Sprintf("Trial started! %d days of full access.", TrialDays),
		Tier:          TierTestSuite,
		DaysRemaining: TrialDays,
		IsTrialMode:   true,
	}
}

// Activate attempts to activate a license key
func (m *Manager) Activate(licenseKey string) *ActivationResult {
	// Validate the license key offline
	info := ValidateLicenseKey(licenseKey)
	if !info.Valid {
		return &ActivationResult{
			Success: false,
			Message: info.ErrorMsg,
		}
	}

	// Create new activation state
	m.state = &ActivationState{
		LicenseKey:      info.Key,
		DeviceHash:      m.fingerprint.Hash(),
		Tier:            info.Tier,
		ActivatedAt:     time.Now(),
		LastValidatedAt: time.Now(),
		ExpiresAt:       time.Now().AddDate(1, 0, 0), // 1 year from activation
		IsTrialMode:     false,
		Features:        info.Features,
	}

	// Save state
	if err := m.saveState(); err != nil {
		return &ActivationResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save activation: %v", err),
		}
	}

	return &ActivationResult{
		Success:       true,
		Message:       fmt.Sprintf("License activated successfully! Tier: %s", info.Tier),
		Tier:          info.Tier,
		DaysRemaining: 365,
	}
}

// Deactivate removes the current license
func (m *Manager) Deactivate() error {
	licensePath := filepath.Join(m.configDir, licenseFileName)
	if err := os.Remove(licensePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	m.state = nil
	return nil
}

// CheckIn performs optional online validation (placeholder for future server)
func (m *Manager) CheckIn() *ActivationResult {
	if m.state == nil {
		return &ActivationResult{
			Success: false,
			Message: "No active license to validate",
		}
	}

	// Update last validated time
	m.state.LastValidatedAt = time.Now()
	m.saveState()

	return &ActivationResult{
		Success: true,
		Message: "License validated successfully",
		Tier:    m.state.Tier,
	}
}

// NeedsCheckIn returns true if optional check-in is recommended
func (m *Manager) NeedsCheckIn() bool {
	if m.state == nil || m.state.IsTrialMode {
		return false
	}

	daysSinceCheck := int(time.Since(m.state.LastValidatedAt).Hours() / 24)
	return daysSinceCheck >= CheckInInterval
}

// loadState loads activation state from disk
func (m *Manager) loadState() error {
	licensePath := filepath.Join(m.configDir, licenseFileName)

	data, err := os.ReadFile(licensePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No license file yet
		}
		return err
	}

	// Decrypt
	decrypted, err := m.decrypt(data)
	if err != nil {
		return fmt.Errorf("failed to decrypt license: %w", err)
	}

	// Parse
	state := &ActivationState{}
	if err := json.Unmarshal(decrypted, state); err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}

	m.state = state
	return nil
}

// saveState saves activation state to disk
func (m *Manager) saveState() error {
	if m.state == nil {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(m.configDir, 0700); err != nil {
		return err
	}

	// Serialize
	data, err := json.Marshal(m.state)
	if err != nil {
		return err
	}

	// Encrypt
	encrypted, err := m.encrypt(data)
	if err != nil {
		return err
	}

	// Write
	licensePath := filepath.Join(m.configDir, licenseFileName)
	return os.WriteFile(licensePath, encrypted, 0600)
}

// encrypt encrypts data using AES-GCM with device-derived key
func (m *Manager) encrypt(plaintext []byte) ([]byte, error) {
	key := m.deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// decrypt decrypts data using AES-GCM with device-derived key
func (m *Manager) decrypt(ciphertext []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, err
	}

	key := m.deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertextBytes, nil)
}

// deriveKey derives an encryption key from device fingerprint
func (m *Manager) deriveKey() []byte {
	data := m.fingerprint.Hash() + encryptionSalt
	hash := sha256.Sum256([]byte(data))
	return hash[:] // 32 bytes for AES-256
}

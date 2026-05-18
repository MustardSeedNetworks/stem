// SPDX-License-Identifier: BUSL-1.1

// Package license provides licensing functionality for Seed Test Suite.
// It handles license key validation, device binding, trial mode, and secure storage.
package license

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// License file locations.
const licenseFileName = ".seed-license"

// TrialDays is the number of days before license is required.
const TrialDays = 14

// OfflineMaxDays is the number of days allowed offline after activation.
const OfflineMaxDays = 90

// CheckInInterval is the number of days between optional check-ins.
const CheckInInterval = 30

// Encryption key derivation salt.
const encryptionSalt = "MSN-SEED-2024-LICENSE"

// Hours per day for time calculations.
const hoursPerDay = 24

// Days in a year for license expiration.
const daysPerYear = 365

// ActivationState represents the current license activation status.
type ActivationState struct {
	LicenseKey      string    `json:"licenseKey"`
	DeviceHash      string    `json:"deviceHash"`
	Tier            Tier      `json:"tier"`
	ActivatedAt     time.Time `json:"activatedAt"`
	LastValidatedAt time.Time `json:"lastValidatedAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	TrialStartedAt  time.Time `json:"trialStartedAt,omitzero"`
	IsTrialMode     bool      `json:"isTrialMode"`
	Features        []string  `json:"features"`
}

// ActivationResult contains the result of an activation attempt.
type ActivationResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Tier          Tier   `json:"tier,omitempty"`
	DaysRemaining int    `json:"daysRemaining,omitempty"`
	IsTrialMode   bool   `json:"isTrialMode"`
}

// Manager handles license activation and validation.
type Manager struct {
	state       *ActivationState
	fingerprint *DeviceFingerprint
	configDir   string
}

// NewManager creates a new license manager.
func NewManager() (*Manager, error) {
	m := &Manager{
		state:       nil,
		fingerprint: nil,
		configDir:   "",
	}

	// Get device fingerprint.
	fp, fpErr := GenerateFingerprint()
	if fpErr != nil {
		return nil, fmt.Errorf("failed to generate fingerprint: %w", fpErr)
	}
	m.fingerprint = fp

	// Determine config directory.
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		homeDir = "/tmp"
	}
	m.configDir = filepath.Join(homeDir, ".config", "seed-test-suite")

	// Load existing state (best-effort, non-fatal).
	loadErr := m.loadState()
	if loadErr != nil {
		// Log but don't fail - fresh state is acceptable.
		_ = loadErr // State will be initialized fresh.
	}

	return m, nil
}

// GetState returns the current activation state.
func (m *Manager) GetState() *ActivationState {
	return m.state
}

// GetFingerprint returns the device fingerprint.
func (m *Manager) GetFingerprint() *DeviceFingerprint {
	return m.fingerprint
}

// IsActivated returns true if a valid license is active.
func (m *Manager) IsActivated() bool {
	if m.state == nil {
		return false
	}

	// Check if still valid.
	if m.state.IsTrialMode {
		return m.IsTrialValid()
	}

	// Check expiration.
	if !m.state.ExpiresAt.IsZero() && time.Now().After(m.state.ExpiresAt) {
		return false
	}

	// Check device binding.
	if m.state.DeviceHash != m.fingerprint.Hash() {
		return false
	}

	return true
}

// IsTrialValid returns true if trial period is still active.
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

// TrialDaysRemaining returns days left in trial.
func (m *Manager) TrialDaysRemaining() int {
	if m.state == nil || !m.state.IsTrialMode {
		return 0
	}

	if m.state.TrialStartedAt.IsZero() {
		return TrialDays
	}

	trialEnd := m.state.TrialStartedAt.AddDate(0, 0, TrialDays)
	remaining := int(time.Until(trialEnd).Hours() / hoursPerDay)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// StartTrial begins the trial period.
func (m *Manager) StartTrial() *ActivationResult {
	// Check if already activated.
	if m.IsActivated() && !m.state.IsTrialMode {
		return &ActivationResult{
			Success:       true,
			Message:       "Already activated with full license",
			Tier:          m.state.Tier,
			DaysRemaining: 0,
			IsTrialMode:   false,
		}
	}

	// Check if trial already started.
	if m.state != nil && !m.state.TrialStartedAt.IsZero() {
		remaining := m.TrialDaysRemaining()
		if remaining <= 0 {
			return &ActivationResult{
				Success:       false,
				Message:       "Trial period has expired. Please enter a license key.",
				Tier:          TierInvalid,
				DaysRemaining: 0,
				IsTrialMode:   true,
			}
		}
		return &ActivationResult{
			Success:       true,
			Message:       fmt.Sprintf("Trial active: %d days remaining", remaining),
			Tier:          TierTestSuite, // Full features during trial.
			DaysRemaining: remaining,
			IsTrialMode:   true,
		}
	}

	// Start new trial.
	m.state = &ActivationState{
		LicenseKey:      "",
		DeviceHash:      m.fingerprint.Hash(),
		Tier:            TierTestSuite, // Full features during trial.
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Now(),
		IsTrialMode:     true,
		Features:        []string{"reflector", "rfc2544", "y1564", "rfc2889", "rfc6349", "y1731", "mef", "tsn"},
	}

	saveErr := m.saveState()
	if saveErr != nil {
		return &ActivationResult{
			Success:       false,
			Message:       fmt.Sprintf("Failed to save trial state: %v", saveErr),
			Tier:          TierInvalid,
			DaysRemaining: 0,
			IsTrialMode:   false,
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

// Activate attempts to activate a license key.
func (m *Manager) Activate(licenseKey string) *ActivationResult {
	// Validate the license key offline.
	info := ValidateLicenseKey(licenseKey)
	if !info.Valid {
		return &ActivationResult{
			Success:       false,
			Message:       info.ErrorMsg,
			Tier:          TierInvalid,
			DaysRemaining: 0,
			IsTrialMode:   false,
		}
	}

	// Create new activation state.
	m.state = &ActivationState{
		LicenseKey:      info.Key,
		DeviceHash:      m.fingerprint.Hash(),
		Tier:            info.Tier,
		ActivatedAt:     time.Now(),
		LastValidatedAt: time.Now(),
		ExpiresAt:       time.Now().AddDate(1, 0, 0), // 1 year from activation.
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        info.Features,
	}

	// Save state.
	saveErr := m.saveState()
	if saveErr != nil {
		return &ActivationResult{
			Success:       false,
			Message:       fmt.Sprintf("Failed to save activation: %v", saveErr),
			Tier:          TierInvalid,
			DaysRemaining: 0,
			IsTrialMode:   false,
		}
	}

	return &ActivationResult{
		Success:       true,
		Message:       fmt.Sprintf("License activated successfully! Tier: %s", info.Tier),
		Tier:          info.Tier,
		DaysRemaining: daysPerYear,
		IsTrialMode:   false,
	}
}

// Deactivate removes the current license.
func (m *Manager) Deactivate() error {
	licensePath := filepath.Join(m.configDir, licenseFileName)
	removeErr := os.Remove(licensePath)
	if removeErr != nil && !os.IsNotExist(removeErr) {
		return fmt.Errorf("failed to remove license file: %w", removeErr)
	}
	m.state = nil
	return nil
}

// CheckIn performs optional online validation (placeholder for future server).
func (m *Manager) CheckIn() *ActivationResult {
	if m.state == nil {
		return &ActivationResult{
			Success:       false,
			Message:       "No active license to validate",
			Tier:          TierInvalid,
			DaysRemaining: 0,
			IsTrialMode:   false,
		}
	}

	// Update last validated time.
	m.state.LastValidatedAt = time.Now()
	_ = m.saveState() // Best-effort state persistence.

	return &ActivationResult{
		Success:       true,
		Message:       "License validated successfully",
		Tier:          m.state.Tier,
		DaysRemaining: 0,
		IsTrialMode:   false,
	}
}

// NeedsCheckIn returns true if optional check-in is recommended.
func (m *Manager) NeedsCheckIn() bool {
	if m.state == nil || m.state.IsTrialMode {
		return false
	}

	daysSinceCheck := int(time.Since(m.state.LastValidatedAt).Hours() / hoursPerDay)
	return daysSinceCheck >= CheckInInterval
}

// loadState loads activation state from disk.
func (m *Manager) loadState() error {
	licensePath := filepath.Clean(filepath.Join(m.configDir, licenseFileName))

	f, openErr := os.Open(licensePath)
	if openErr != nil {
		if os.IsNotExist(openErr) {
			return nil // No license file yet.
		}
		return fmt.Errorf("open license file: %w", openErr)
	}
	defer func() { _ = f.Close() }()

	data, readErr := io.ReadAll(f)
	if readErr != nil {
		return fmt.Errorf("read license file: %w", readErr)
	}

	// Decrypt.
	decrypted, decryptErr := m.decrypt(data)
	if decryptErr != nil {
		return fmt.Errorf("failed to decrypt license: %w", decryptErr)
	}

	// Parse.
	state := &ActivationState{
		LicenseKey:      "",
		DeviceHash:      "",
		Tier:            TierInvalid,
		ActivatedAt:     time.Time{},
		LastValidatedAt: time.Time{},
		ExpiresAt:       time.Time{},
		TrialStartedAt:  time.Time{},
		IsTrialMode:     false,
		Features:        nil,
	}
	unmarshalErr := json.Unmarshal(decrypted, state)
	if unmarshalErr != nil {
		return fmt.Errorf("failed to parse license: %w", unmarshalErr)
	}

	m.state = state
	return nil
}

// saveState saves activation state to disk.
func (m *Manager) saveState() error {
	if m.state == nil {
		return nil
	}

	// Ensure directory exists.
	mkdirErr := os.MkdirAll(m.configDir, 0o700)
	if mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	// Serialize.
	data, marshalErr := json.Marshal(m.state)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal license state: %w", marshalErr)
	}

	// Encrypt.
	encrypted, encryptErr := m.encrypt(data)
	if encryptErr != nil {
		return encryptErr
	}

	// Write.
	licensePath := filepath.Join(m.configDir, licenseFileName)
	writeErr := os.WriteFile(licensePath, encrypted, 0o600)
	if writeErr != nil {
		return fmt.Errorf("failed to write license file: %w", writeErr)
	}
	return nil
}

// encrypt encrypts data using AES-GCM with device-derived key.
func (m *Manager) encrypt(plaintext []byte) ([]byte, error) {
	key := m.deriveKey()

	block, blockErr := aes.NewCipher(key)
	if blockErr != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", blockErr)
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", gcmErr)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, nonceErr := io.ReadFull(rand.Reader, nonce)
	if nonceErr != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", nonceErr)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// decrypt decrypts data using AES-GCM with device-derived key.
func (m *Manager) decrypt(ciphertext []byte) ([]byte, error) {
	data, decodeErr := base64.StdEncoding.DecodeString(string(ciphertext))
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", decodeErr)
	}

	key := m.deriveKey()

	block, blockErr := aes.NewCipher(key)
	if blockErr != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", blockErr)
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", gcmErr)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, openErr := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if openErr != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", openErr)
	}
	return plaintext, nil
}

// deriveKey derives an encryption key from device fingerprint.
func (m *Manager) deriveKey() []byte {
	data := m.fingerprint.Hash() + encryptionSalt
	hash := sha256.Sum256([]byte(data))
	return hash[:] // 32 bytes for AES-256.
}

// Package auth provides JWT authentication helpers.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidCredentials indicates the username or password was wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken signals the JWT could not be parsed.
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired indicates the token expired.
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenRevoked indicates the token has been revoked/logged out.
	ErrTokenRevoked = errors.New("token has been revoked")
	// ErrMissingCredentials indicates required credentials were not provided.
	ErrMissingCredentials = errors.New(
		"missing required credentials: set STEM_AUTH_USERNAME and STEM_AUTH_PASSWORD environment variables",
	)
	// ErrPasswordHashFailed indicates the password hash routine failed.
	ErrPasswordHashFailed = errors.New("failed to hash password")
	// ErrSecretGenerationFailed indicates random bytes could not be generated.
	ErrSecretGenerationFailed = errors.New("failed to generate JWT secret")
)

const (
	// AccessTokenDuration is the default lifetime for access tokens (15 minutes).
	AccessTokenDuration = 15 * time.Minute
	// RefreshTokenDuration is the lifetime for refresh tokens (7 days).
	RefreshTokenDuration = 7 * 24 * time.Hour

	jwtSecretLength = 32
	tokenIDLength   = 16

	// defaultIssuer is the JWT issuer string. Used by the access-token
	// generator and the Wave-3 mfa_pending token generator alike.
	defaultIssuer = "The Stem"
)

// Claims represents the custom portion of the JWT payload.
type Claims struct {
	jwt.RegisteredClaims

	Username  string `json:"username"`
	TokenType string `json:"token_type"`
}

// Manager issues and validates JWT tokens.
type Manager struct {
	mu             sync.RWMutex
	jwtSecret      []byte
	sessionTimeout time.Duration
	username       string
	passwordHash   []byte
	issuer         string
	blacklist      *TokenBlacklist
	randReader     io.Reader
	// totp holds the optional second-factor configuration. Wave 3 (#85)
	// introduced TOTP MFA; see totp_state.go for the storage contract.
	// The secret is in-memory only — operators who restart the binary
	// must re-enrol via /api/v1/auth/totp/setup.
	totp totpState
	// webauthnCredentials holds zero or more registered WebAuthn
	// credentials (passkeys). Each entry is one device. Wave 3 (#85)
	// stores these in-memory alongside totp; persistence is deferred to
	// a future wave that introduces a secret-management layer.
	webauthnCredentials []WebAuthnStoredCredential
	// webauthnSessions holds pending registration/login ceremonies
	// keyed by username. WebAuthn ceremonies are two-step: begin
	// returns a challenge, finish consumes it. The challenge MUST be
	// remembered server-side between the two calls.
	webauthnSessions map[string]*WebAuthnSessionState
}

// NewManager creates an auth manager that can sign tokens.
// Returns ErrMissingCredentials if username or password is empty.
func NewManager(jwtSecret string, sessionTimeout time.Duration, username, password string) (*Manager, error) {
	// Require explicit credentials - no defaults allowed.
	if username == "" || password == "" {
		return nil, ErrMissingCredentials
	}

	secret := jwtSecret
	if secret == "" {
		var err error
		secret, err = GenerateJWTSecret()
		if err != nil {
			return nil, err
		}
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPasswordHashFailed, err)
	}

	if sessionTimeout <= 0 {
		sessionTimeout = AccessTokenDuration
	}

	return &Manager{
		mu:                  sync.RWMutex{},
		jwtSecret:           []byte(secret),
		sessionTimeout:      sessionTimeout,
		username:            username,
		passwordHash:        []byte(hash),
		issuer:              defaultIssuer,
		blacklist:           NewTokenBlacklist(),
		randReader:          rand.Reader,
		totp:                totpState{secret: "", enabled: false},
		webauthnCredentials: nil,
		webauthnSessions:    make(map[string]*WebAuthnSessionState),
	}, nil
}

// Stop stops the manager's background goroutines.
// Should be called when the manager is no longer needed.
func (m *Manager) Stop() {
	if m.blacklist != nil {
		m.blacklist.Stop()
	}
}

// Authenticate validates credentials and emits a signed JWT token.
//
// As of Wave 2 (#84) the stored hash may be Argon2id (new format) or
// legacy bcrypt; [VerifyPassword] dispatches transparently. When a
// successful login is verified against a bcrypt hash the manager
// silently upgrades the stored hash to Argon2id — this is the
// just-in-time migration path for existing deployments and is the
// **only** way bcrypt hashes leave the system.
func (m *Manager) Authenticate(ctx context.Context, username, password string) (string, error) {
	m.mu.RLock()
	storedUsername := m.username
	storedHash := string(m.passwordHash)
	m.mu.RUnlock()

	usernameMatch := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(username)),
		[]byte(strings.ToLower(storedUsername)),
	) == 1

	matched, needsRehash, err := VerifyPassword(storedHash, password)
	if err != nil || !usernameMatch || !matched {
		return "", ErrInvalidCredentials
	}

	if needsRehash {
		m.upgradeHash(ctx, password)
	}

	return m.generateToken(username)
}

// upgradeHash re-hashes the verified password with Argon2id and
// persists it via UpdatePasswordHash. Errors are non-fatal — the user
// has already successfully authenticated. We log the failure rather
// than blocking the login, since a partial migration is preferable to
// locking the user out.
func (m *Manager) upgradeHash(ctx context.Context, password string) {
	newHash, err := HashPassword(password)
	if err != nil {
		// Migration is best-effort; the caller already verified.
		return
	}
	m.UpdatePasswordHash(ctx, newHash)
}

// ValidateToken parses and validates a JWT token.
// Returns ErrTokenRevoked if the token has been revoked via logout.
func (m *Manager) ValidateToken(_ context.Context, tokenString string) (*Claims, error) {
	claims := new(Claims)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("parse token: %w", ErrInvalidToken)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token has been revoked.
	if claims.ID != "" && m.blacklist.IsBlacklisted(claims.ID) {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

func (m *Manager) generateToken(username string) (string, error) {
	return m.generateTokenWithType(username, "access", m.sessionTimeout)
}

func (m *Manager) generateTokenWithType(username, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()

	// Generate unique token ID for revocation support.
	tokenID, err := m.generateTokenID()
	if err != nil {
		return "", fmt.Errorf("generate token ID: %w", err)
	}

	claims := &Claims{
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings(nil),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   username,
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// generateTokenID creates a unique identifier for tokens.
func (m *Manager) generateTokenID() (string, error) {
	bytes := make([]byte, tokenIDLength)
	_, err := io.ReadFull(m.randReader, bytes)
	if err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// SessionDuration returns the configured token lifetime.
func (m *Manager) SessionDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionTimeout
}

// SetRandReader sets the random reader for token ID generation.
// This is primarily useful for testing to inject error conditions.
func (m *Manager) SetRandReader(r io.Reader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.randReader = r
}

// RevokeToken adds a token to the blacklist, invalidating it for future use.
// This is used for logout functionality.
func (m *Manager) RevokeToken(claims *Claims) {
	if claims == nil || claims.ID == "" {
		return
	}
	expiresAt := time.Now().Add(m.sessionTimeout)
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	m.blacklist.Add(claims.ID, expiresAt)
}

// GenerateRefreshToken creates a long-lived refresh token for the user.
func (m *Manager) GenerateRefreshToken(username string) (string, error) {
	return m.generateTokenWithType(username, "refresh", RefreshTokenDuration)
}

// RefreshAccessToken validates a refresh token and issues a new access token.
// Returns ErrInvalidToken if the token is not a valid refresh token.
func (m *Manager) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := m.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	if claims.TokenType != "refresh" {
		return "", fmt.Errorf("%w: not a refresh token", ErrInvalidToken)
	}

	return m.generateToken(claims.Username)
}

// AuthenticateWithRefresh validates credentials and returns both access and refresh tokens.
//
// As of Wave 2 (#84) hash verification goes through [VerifyPassword],
// which transparently handles both Argon2id (preferred) and legacy
// bcrypt hashes. A successful login against a bcrypt hash triggers an
// in-place upgrade to Argon2id.
func (m *Manager) AuthenticateWithRefresh(
	ctx context.Context, username, password string,
) (string, string, error) {
	m.mu.RLock()
	storedUsername := m.username
	storedHash := string(m.passwordHash)
	m.mu.RUnlock()

	usernameMatch := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(username)),
		[]byte(strings.ToLower(storedUsername)),
	) == 1

	matched, needsRehash, err := VerifyPassword(storedHash, password)
	if err != nil || !usernameMatch || !matched {
		return "", "", ErrInvalidCredentials
	}

	if needsRehash {
		m.upgradeHash(ctx, password)
	}

	accessToken, err := m.generateToken(username)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := m.GenerateRefreshToken(username)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateJWTSecret returns a new 256-bit base64url JWT secret.
func GenerateJWTSecret() (string, error) {
	return GenerateJWTSecretFrom(rand.Reader)
}

// GenerateJWTSecretFrom returns a new 256-bit base64url JWT secret using the provided reader.
// This is useful for testing to inject error conditions.
func GenerateJWTSecretFrom(r io.Reader) (string, error) {
	bytes := make([]byte, jwtSecretLength)
	read, err := io.ReadFull(r, bytes)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSecretGenerationFailed, err)
	}
	if read != jwtSecretLength {
		return "", fmt.Errorf("%w: incomplete read", ErrSecretGenerationFailed)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

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
	"golang.org/x/crypto/bcrypt"
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
	// ErrPasswordHashFailed indicates bcrypt failed to hash the password.
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

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPasswordHashFailed, err)
	}

	if sessionTimeout <= 0 {
		sessionTimeout = AccessTokenDuration
	}

	return &Manager{
		mu:             sync.RWMutex{},
		jwtSecret:      []byte(secret),
		sessionTimeout: sessionTimeout,
		username:       username,
		passwordHash:   hash,
		issuer:         "The Stem",
		blacklist:      NewTokenBlacklist(),
		randReader:     rand.Reader,
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
func (m *Manager) Authenticate(_ context.Context, username, password string) (string, error) {
	m.mu.RLock()
	storedUsername := m.username
	storedHash := m.passwordHash
	m.mu.RUnlock()

	usernameMatch := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(username)),
		[]byte(strings.ToLower(storedUsername)),
	) == 1

	if !usernameMatch || bcrypt.CompareHashAndPassword(storedHash, []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}

	return m.generateToken(username)
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
func (m *Manager) AuthenticateWithRefresh(
	_ context.Context, username, password string,
) (string, string, error) {
	m.mu.RLock()
	storedUsername := m.username
	storedHash := m.passwordHash
	m.mu.RUnlock()

	usernameMatch := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(username)),
		[]byte(strings.ToLower(storedUsername)),
	) == 1

	if !usernameMatch || bcrypt.CompareHashAndPassword(storedHash, []byte(password)) != nil {
		return "", "", ErrInvalidCredentials
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

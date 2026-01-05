// Package auth provides JWT authentication helpers.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
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
)

// AccessTokenDuration is the default lifetime for access tokens.
const AccessTokenDuration = 30 * time.Minute

const jwtSecretLength = 32

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
}

// NewManager creates an auth manager that can sign tokens.
func NewManager(jwtSecret string, sessionTimeout time.Duration, username, password string) *Manager {
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin"
	}

	secret := jwtSecret
	if secret == "" {
		secret = GenerateJWTSecret()
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("failed to hash auth password: %v", err))
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

	return claims, nil
}

func (m *Manager) generateToken(username string) (string, error) {
	now := time.Now()
	claims := &Claims{
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings(nil),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.sessionTimeout)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   username,
			ID:        "",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// SessionDuration returns the configured token lifetime.
func (m *Manager) SessionDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionTimeout
}

// GenerateJWTSecret returns a new 256-bit base64url JWT secret.
func GenerateJWTSecret() string {
	bytes := make([]byte, jwtSecretLength)
	read, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("failed to generate JWT secret: %v", err))
	}
	if read != jwtSecretLength {
		panic("failed to generate complete JWT secret")
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

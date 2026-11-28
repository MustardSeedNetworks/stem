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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/credstore"
	"github.com/MustardSeedNetworks/foundation/pkg/password"
	"github.com/go-webauthn/webauthn/webauthn"
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

	// ErrWrongTokenType reports a token presented where its kind is not
	// accepted — a refresh token used to authenticate an API request, or an
	// access token offered for exchange. The kinds have deliberately
	// different lifetimes; accepting one for the other collapses that.
	ErrWrongTokenType = errors.New("token is not valid for this use")
	// ErrMissingCredentials indicates a credential given with one half missing.
	ErrMissingCredentials = errors.New(
		"incomplete credentials: set both STEM_AUTH_USERNAME and STEM_AUTH_PASSWORD, " +
			"or neither to set the password through first-run setup in the web UI",
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

	// FirstRunUsername is the administrator an unconfigured daemon offers
	// for claiming through first-run setup (#1282).
	FirstRunUsername = "admin"
)

// Claims represents the custom portion of the JWT payload.
type Claims struct {
	jwt.RegisteredClaims

	Username  string `json:"username"`
	TokenType string `json:"token_type"`
}

// Token kinds. Each has its own lifetime and its own valid use, and
// validateTokenOfType is what keeps them from being swapped (#1169).
// TokenTypeCLI is declared in clitoken.go with the credential it names.
const (
	// TokenTypeAccess authenticates API requests for a browser session.
	TokenTypeAccess = "access"

	// TokenTypeRefresh may only be exchanged for a fresh access token.
	TokenTypeRefresh = "refresh"
)

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
	// totp holds the optional second-factor configuration; see
	// totp_state.go. passkeyHandle is the account's opaque WebAuthn user
	// handle, and passkeys its complete credential records. store holds all
	// of it durably (store.go); these fields are the working copy.
	totp          totpState
	passkeyHandle []byte
	passkeys      []webauthn.Credential
	store         *credstore.Store
	// commitMu orders changes to the store and the refresh of the working
	// copy that follows each one.
	commitMu sync.Mutex
	// lock is the failed-login count and lock for the one account; now is
	// its clock, replaceable so tests need not wait out a lockout.
	lock lockout
	now  func() time.Time
	// verify checks a password against a stored hash; a field so tests can
	// observe that every login attempt pays for a hash.
	verify func(hash, plaintext string) (bool, error)
	// gate bounds concurrent password hashes; see verifyGate.
	gate verifyGate
	// dummyHash is a valid hash of a throwaway password. A refused attempt for
	// the real account verifies against it, so a locked account costs what an
	// unknown username or a wrong password costs.
	dummyHash func() string
}

// NewManager creates an auth manager for a credential given in plain text,
// as STEM_AUTH_USERNAME / STEM_AUTH_PASSWORD supply it.
// Returns ErrMissingCredentials if username or password is empty.
func NewManager(jwtSecret string, sessionTimeout time.Duration, username, plaintext string) (*Manager, error) {
	// Require explicit credentials - no defaults allowed.
	if username == "" || plaintext == "" {
		return nil, ErrMissingCredentials
	}

	hash, err := password.Hash(plaintext)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPasswordHashFailed, err)
	}
	return NewManagerFromHash(jwtSecret, sessionTimeout, username, hash)
}

// NewManagerFromHash creates an auth manager for an already-hashed
// credential. An empty passwordHash is an unclaimed daemon: every login is
// refused, since [password.Verify] accepts no empty record, until first-run
// setup stores a credential through [Manager.SetCredential].
func NewManagerFromHash(
	jwtSecret string,
	sessionTimeout time.Duration,
	username, passwordHash string,
) (*Manager, error) {
	secret := jwtSecret
	if secret == "" {
		var err error
		secret, err = GenerateJWTSecret()
		if err != nil {
			return nil, err
		}
	}

	if sessionTimeout <= 0 {
		sessionTimeout = AccessTokenDuration
	}

	return &Manager{
		mu:             sync.RWMutex{},
		jwtSecret:      []byte(secret),
		sessionTimeout: sessionTimeout,
		username:       username,
		passwordHash:   []byte(passwordHash),
		issuer:         defaultIssuer,
		blacklist:      NewTokenBlacklist(),
		randReader:     rand.Reader,
		totp:           totpState{secret: "", enabled: false},
		passkeyHandle:  newPasskeyHandle(),
		passkeys:       nil,
		store:          nil,
		commitMu:       sync.Mutex{},
		lock:           lockout{failures: 0, until: time.Time{}},
		now:            time.Now,
		verify:         password.Verify,
		gate:           newVerifyGate(),
		dummyHash: sync.OnceValue(func() string {
			hash, _ := password.Hash("not-a-credential-" + defaultIssuer)
			return hash
		}),
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
func (m *Manager) Authenticate(username, plaintext string) (string, error) {
	if err := m.checkPassword(username, plaintext); err != nil {
		return "", err
	}
	return m.generateToken(username)
}

// checkPassword is the one password check behind every login path. A locked
// account is refused before any hash work; a wrong password for the real
// account counts toward the lock, while a wrong username does not, so
// invented names can neither lock the administrator out nor grow state. An
// unclaimed daemon has nothing to guess, so it never locks.
func (m *Manager) checkPassword(username, plaintext string) error {
	m.mu.RLock()
	storedUsername := m.username
	storedHash := string(m.passwordHash)
	verify := m.verify
	m.mu.RUnlock()

	usernameMatch := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(username)),
		[]byte(strings.ToLower(storedUsername)),
	) == 1

	// Only the real, claimed account takes a lockout slot. Any other name
	// would let a flood of invented usernames fill the in-flight cap and lock
	// the administrator out; it is still hashed in full, so the response time
	// does not say which names exist.
	counted := usernameMatch && storedHash != ""
	release := func(bool) {}
	if counted {
		var ok bool
		if release, ok = m.TryBeginLoginAttempt(); !ok {
			_, _ = verify(m.dummyHash(), plaintext)
			return ErrAccountLocked
		}
	}

	matched, err := verify(storedHash, plaintext)
	if err != nil || !usernameMatch || !matched {
		release(false)
		return ErrInvalidCredentials
	}
	release(true)
	return nil
}

// ValidateToken parses and validates a JWT token.
// Returns ErrTokenRevoked if the token has been revoked via logout.
// ValidateToken validates a credential presented to authenticate an API
// request. Only the kinds that may do so are accepted — a browser session
// and the local CLI — so a refresh token, which lives far longer, cannot be
// replayed as an access token (#1169).
func (m *Manager) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	return m.validateTokenOfType(ctx, tokenString, TokenTypeAccess, TokenTypeCLI)
}

// ValidateRefreshToken validates a credential offered for exchange at the
// refresh endpoint. The mirror image of ValidateToken: only a refresh token
// qualifies, so an access or CLI token cannot mint fresh sessions forever.
func (m *Manager) ValidateRefreshToken(ctx context.Context, tokenString string) (*Claims, error) {
	return m.validateTokenOfType(ctx, tokenString, TokenTypeRefresh)
}

// validateTokenOfType verifies signature, expiry and revocation, then
// refuses any token whose kind is not in accepted. An unknown or absent
// kind matches nothing and is refused with it: the default fails closed.
func (m *Manager) validateTokenOfType(
	_ context.Context,
	tokenString string,
	accepted ...string,
) (*Claims, error) {
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

	if !slices.Contains(accepted, claims.TokenType) {
		return nil, fmt.Errorf("%w: %q", ErrWrongTokenType, claims.TokenType)
	}

	return claims, nil
}

func (m *Manager) generateToken(username string) (string, error) {
	return m.generateTokenWithType(username, TokenTypeAccess, m.sessionTimeout)
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
		Audience:  jwt.ClaimStrings(nil),
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    m.issuer,
		Subject:   username,
		ID:        tokenID,
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
	return m.generateTokenWithType(username, TokenTypeRefresh, RefreshTokenDuration)
}

// RefreshAccessToken validates a refresh token and issues a new access token.
// Returns ErrInvalidToken if the token is not a valid refresh token.
func (m *Manager) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := m.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	return m.generateToken(claims.Username)
}

// AuthenticateWithRefresh validates credentials and returns both access and refresh tokens.
// It does not clear the failed-login count: when a second factor follows, the
// login is not complete yet, and clearing here would let a caller who knows the
// password guess codes without limit.
func (m *Manager) AuthenticateWithRefresh(username, plaintext string) (string, string, error) {
	if err := m.checkPassword(username, plaintext); err != nil {
		return "", "", err
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

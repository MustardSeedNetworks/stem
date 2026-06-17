// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Argon2id RFC 9106 second-recommended parameters (interactive auth).
const (
	// argon2Memory is the memory cost in KiB (64 MiB).
	argon2Memory = 64 * 1024
	// argon2Time is the number of iterations.
	argon2Time = 3
	// argon2Threads is the degree of parallelism.
	argon2Threads = 4
	// argon2SaltLen is the salt length in bytes.
	argon2SaltLen = 16
	// argon2KeyLen is the derived key length in bytes.
	argon2KeyLen = 32

	// argon2idPrefix identifies a PHC-encoded Argon2id hash.
	argon2idPrefix = "$argon2id$"

	// argon2Version is the version field embedded in the PHC string. Value 19 = 0x13.
	argon2Version = argon2.Version

	// argon2TestMemoryKiB is the reduced Argon2id memory cost used when
	// STEM_TEST_MODE is set so the test suite can instantiate many auth
	// managers under the race detector. Production binaries never see this.
	argon2TestMemoryKiB = 64
)

// Password validation constants.
const (
	// MinPasswordLength is the minimum required password length.
	MinPasswordLength = 12
	// GeneratedPasswordLength is the length for auto-generated passwords.
	GeneratedPasswordLength = 24

	// defaultPasswordHashPrefix identifies the default/placeholder password.
	// If the stored hash starts with this, setup is required.
	defaultPasswordHashPrefix = "$2a$10$default"
)

// Password errors.
var (
	// ErrPasswordTooShort indicates the password doesn't meet minimum length.
	ErrPasswordTooShort = errors.New("password must be at least 12 characters")
	// ErrUnsupportedHash indicates the stored password hash uses an unknown algorithm.
	ErrUnsupportedHash = errors.New("unsupported password hash algorithm")
	// ErrInvalidArgon2Hash indicates the Argon2id PHC string is malformed.
	ErrInvalidArgon2Hash = errors.New("invalid argon2id hash format")
)

// argon2Params describes a tunable Argon2id parameter set. The defaults
// satisfy RFC 9106 §4.4; the test override (STEM_TEST_MODE) reduces
// memory + iterations so the suite finishes in milliseconds.
type argon2Params struct {
	memory  uint32
	time    uint32
	threads uint8
	saltLen uint32
	keyLen  uint32
}

// argon2Tunables returns the active Argon2id parameters. The test
// override (STEM_TEST_MODE) keeps the historical fast-mode behavior
// from the bcrypt era; production binaries always run with the
// RFC 9106 defaults.
func argon2Tunables() argon2Params {
	if os.Getenv("STEM_TEST_MODE") != "" {
		return argon2Params{
			memory:  argon2TestMemoryKiB,
			time:    1,
			threads: 1,
			saltLen: argon2SaltLen,
			keyLen:  argon2KeyLen,
		}
	}
	return argon2Params{
		memory:  argon2Memory,
		time:    argon2Time,
		threads: argon2Threads,
		saltLen: argon2SaltLen,
		keyLen:  argon2KeyLen,
	}
}

// IsDefaultPasswordHash checks if the given hash is the default placeholder.
// Returns true if initial setup is required.
func IsDefaultPasswordHash(hash string) bool {
	// If empty or matches the default prefix, setup is needed.
	if hash == "" {
		return true
	}
	// Check for the special default marker.
	if len(hash) >= len(defaultPasswordHashPrefix) {
		return hash[:len(defaultPasswordHashPrefix)] == defaultPasswordHashPrefix
	}
	return false
}

// GenerateSecurePassword creates a cryptographically random password.
// The password is base64-encoded for easy copying and includes special characters.
func GenerateSecurePassword(length int) (string, error) {
	if length <= 0 {
		length = GeneratedPasswordLength
	}

	// Generate random bytes (more than needed due to base64 expansion).
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	// Encode to base64 and trim to requested length.
	password := base64.URLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}

// ValidatePasswordStrength checks if a password meets security requirements.
// This is the cheap length-only check; the zxcvbn-backed strength meter
// lives in [EvaluatePasswordStrength].
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

// HashPassword creates an Argon2id (RFC 9106) hash of the given password
// and returns it in PHC format:
//
//	$argon2id$v=19$m=65536,t=3,p=4$<b64salt>$<b64hash>
//
// The signature is preserved from the bcrypt era so callers do not need
// to change. See [VerifyPassword] for verification and migration.
func HashPassword(password string) (string, error) {
	p := argon2Tunables()

	salt := make([]byte, p.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("%w: %w", ErrPasswordHashFailed, err)
	}

	key := argon2.IDKey([]byte(password), salt, p.time, p.memory, p.threads, p.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2Version, p.memory, p.time, p.threads, b64Salt, b64Key,
	)
	return encoded, nil
}

// VerifyPassword checks `password` against a stored hash.
//
// Returns:
//   - matched=true  if the password is correct
//   - needsRehash=true when the stored hash is a legacy bcrypt hash and
//     the caller should re-hash with Argon2id and persist on the next
//     successful login. (Always false for Argon2id hashes.)
//   - err is non-nil for malformed hashes (e.g. unsupported prefix) or
//     a non-mismatch bcrypt failure. A simple "wrong password" against
//     a bcrypt hash is reported as (false, false, nil); cryptographic
//     errors (e.g. truncated hash) propagate so callers can log them.
//
// Verification timing is constant-time within each algorithm path; the
// algorithm branch itself is leak-tolerant because the prefix is public.
func VerifyPassword(hash, password string) (bool, bool, error) {
	switch {
	case strings.HasPrefix(hash, argon2idPrefix):
		ok, vErr := verifyArgon2id(hash, password)
		if vErr != nil {
			return false, false, vErr
		}
		return ok, false, nil
	case strings.HasPrefix(hash, "$2a$"),
		strings.HasPrefix(hash, "$2b$"),
		strings.HasPrefix(hash, "$2y$"):
		bErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if bErr != nil {
			// A simple mismatch is "wrong password" — not an error to
			// the caller. Any other bcrypt failure (e.g. malformed
			// hash) is propagated so it can be logged.
			if errors.Is(bErr, bcrypt.ErrMismatchedHashAndPassword) {
				return false, false, nil
			}
			return false, false, fmt.Errorf("bcrypt compare: %w", bErr)
		}
		return true, true, nil
	default:
		return false, false, ErrUnsupportedHash
	}
}

// verifyArgon2id parses a PHC-encoded Argon2id hash and compares it
// against `password` in constant time.
func verifyArgon2id(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	// Expected layout: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<b64salt>", "<b64hash>"]
	const expectedParts = 6
	if len(parts) != expectedParts {
		return false, ErrInvalidArgon2Hash
	}
	if parts[1] != "argon2id" {
		return false, ErrInvalidArgon2Hash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("%w: bad version", ErrInvalidArgon2Hash)
	}
	if version != argon2Version {
		return false, fmt.Errorf("%w: unsupported argon2 version %d", ErrInvalidArgon2Hash, version)
	}

	memory, time, threads, err := parseArgon2Params(parts[3])
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("%w: bad salt encoding", ErrInvalidArgon2Hash)
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("%w: bad key encoding", ErrInvalidArgon2Hash)
	}
	// Argon2.IDKey takes a uint32 key length. len(want) is a non-negative
	// int; on 64-bit platforms it could theoretically exceed uint32, so
	// we bounds-check before narrowing to satisfy gosec G115 and prevent
	// silent truncation of a hypothetically absurd hash.
	wantLen := len(want)
	if wantLen > math.MaxUint32 {
		return false, fmt.Errorf("%w: key length out of range", ErrInvalidArgon2Hash)
	}
	keyLen := uint32(wantLen)

	got := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// parseArgon2Params parses the "m=<>,t=<>,p=<>" parameter segment and
// returns memory, time, and threads (in that order).
func parseArgon2Params(seg string) (uint32, uint32, uint8, error) {
	const expectedFields = 3
	fields := strings.Split(seg, ",")
	if len(fields) != expectedFields {
		return 0, 0, 0, fmt.Errorf("%w: bad params segment", ErrInvalidArgon2Hash)
	}
	var (
		memory  uint32
		time    uint32
		threads uint8
	)
	for _, f := range fields {
		key, val, ok := strings.Cut(f, "=")
		if !ok {
			return 0, 0, 0, fmt.Errorf("%w: malformed param %q", ErrInvalidArgon2Hash, f)
		}
		switch key {
		case "m":
			n, parseErr := strconv.ParseUint(val, 10, 32)
			if parseErr != nil {
				return 0, 0, 0, fmt.Errorf("%w: bad memory: %w", ErrInvalidArgon2Hash, parseErr)
			}
			memory = uint32(n)
		case "t":
			n, parseErr := strconv.ParseUint(val, 10, 32)
			if parseErr != nil {
				return 0, 0, 0, fmt.Errorf("%w: bad time: %w", ErrInvalidArgon2Hash, parseErr)
			}
			time = uint32(n)
		case "p":
			n, parseErr := strconv.ParseUint(val, 10, 8)
			if parseErr != nil {
				return 0, 0, 0, fmt.Errorf("%w: bad threads: %w", ErrInvalidArgon2Hash, parseErr)
			}
			threads = uint8(n)
		default:
			return 0, 0, 0, fmt.Errorf("%w: unknown param %q", ErrInvalidArgon2Hash, key)
		}
	}
	return memory, time, threads, nil
}

// UpdatePasswordHash updates the stored password hash in the auth manager.
// This is used during initial setup or password changes.
func (m *Manager) UpdatePasswordHash(_ context.Context, newHash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.passwordHash = []byte(newHash)
}

// GetPasswordHash returns the current password hash (for config saving).
func (m *Manager) GetPasswordHash() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return string(m.passwordHash)
}

// GetUsername returns the configured username.
func (m *Manager) GetUsername() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.username
}

// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

//nolint:testpackage // Internal tests needed for accessing private fields (mu, users)
package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestNewMemoryUserStore tests creating a new memory user store.
func TestNewMemoryUserStore(t *testing.T) {
	store := NewMemoryUserStore()
	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	if store.maxLoginAttempts != DefaultMaxLoginAttempts {
		t.Errorf("Expected maxLoginAttempts=%d, got=%d", DefaultMaxLoginAttempts, store.maxLoginAttempts)
	}

	if store.lockDuration != DefaultLockDuration {
		t.Errorf("Expected lockDuration=%v, got=%v", DefaultLockDuration, store.lockDuration)
	}
}

// TestNewMemoryUserStoreWithConfig tests creating a store with custom config.
func TestNewMemoryUserStoreWithConfig(t *testing.T) {
	maxAttempts := 3
	lockDuration := 5 * time.Minute

	store := NewMemoryUserStoreWithConfig(maxAttempts, lockDuration)

	if store.maxLoginAttempts != maxAttempts {
		t.Errorf("Expected maxLoginAttempts=%d, got=%d", maxAttempts, store.maxLoginAttempts)
	}

	if store.lockDuration != lockDuration {
		t.Errorf("Expected lockDuration=%v, got=%v", lockDuration, store.lockDuration)
	}
}

// TestAddUser tests adding a user to the store.
func TestAddUser(t *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	store.AddUser("testuser", "hashedpassword", "admin")

	hash, err := store.GetPasswordHash(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}

	if hash != "hashedpassword" {
		t.Errorf("Expected hash='hashedpassword', got='%s'", hash)
	}
}

// TestGetPasswordHash tests password hash retrieval.
func TestGetPasswordHash(t *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	// User not found
	_, err := store.GetPasswordHash(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Add user and retrieve hash
	store.AddUser("user1", "hash123", "user")
	hash, err := store.GetPasswordHash(ctx, "user1")
	if err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}

	if hash != "hash123" {
		t.Errorf("Expected hash='hash123', got='%s'", hash)
	}
}

// TestGetTokenVersion tests token version retrieval.
func TestGetTokenVersion(t *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	// User not found
	_, err := store.GetTokenVersion(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Add user - initial token version should be 1
	store.AddUser("user1", "hash123", "user")
	version, err := store.GetTokenVersion(ctx, "user1")
	if err != nil {
		t.Fatalf("GetTokenVersion() error: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected initial token version=1, got=%d", version)
	}
}

// TestUpdatePassword tests password update and token version increment.
func TestUpdatePassword(t *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	// User not found
	err := store.UpdatePassword(ctx, "nonexistent", "newhash")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Add user and update password
	store.AddUser("user1", "oldhash", "user")

	err = store.UpdatePassword(ctx, "user1", "newhash")
	if err != nil {
		t.Fatalf("UpdatePassword() error: %v", err)
	}

	// Verify hash changed
	hash, _ := store.GetPasswordHash(ctx, "user1")
	if hash != "newhash" {
		t.Errorf("Expected hash='newhash', got='%s'", hash)
	}

	// Verify token version incremented
	version, _ := store.GetTokenVersion(ctx, "user1")
	if version != 2 {
		t.Errorf("Expected token version=2 after password update, got=%d", version)
	}
}

// TestRecordLoginSuccess tests recording successful logins.
func TestRecordLoginSuccess(t *testing.T) {
	store := NewMemoryUserStoreWithConfig(3, time.Minute)
	ctx := context.Background()

	// User not found
	err := store.RecordLoginSuccess(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Add user and record some failures, then success
	store.AddUser("user1", "hash", "user")
	_ = store.RecordLoginFailure(ctx, "user1")
	_ = store.RecordLoginFailure(ctx, "user1")

	err = store.RecordLoginSuccess(ctx, "user1")
	if err != nil {
		t.Fatalf("RecordLoginSuccess() error: %v", err)
	}

	// Verify attempts reset
	store.mu.RLock()
	user := store.users["user1"]
	attempts := user.failedAttempts
	locked := user.lockedUntil
	store.mu.RUnlock()

	if attempts != 0 {
		t.Errorf("Expected failedAttempts=0 after success, got=%d", attempts)
	}

	if !locked.IsZero() {
		t.Errorf("Expected lockedUntil to be zero after success")
	}
}

// TestRecordLoginFailure tests recording failed logins.
func TestRecordLoginFailure(t *testing.T) {
	store := NewMemoryUserStoreWithConfig(3, time.Minute)
	ctx := context.Background()

	// User not found
	err := store.RecordLoginFailure(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Add user and record failures
	store.AddUser("user1", "hash", "user")

	for i := 1; i <= 3; i++ {
		err = store.RecordLoginFailure(ctx, "user1")
		if err != nil {
			t.Fatalf("RecordLoginFailure() error on attempt %d: %v", i, err)
		}
	}

	// Verify user is locked after max attempts
	locked, err := store.IsLocked(ctx, "user1")
	if err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	}

	if !locked {
		t.Error("Expected user to be locked after max failed attempts")
	}
}

// TestIsLocked tests account lock checking.
func TestIsLocked(t *testing.T) {
	store := NewMemoryUserStoreWithConfig(2, 100*time.Millisecond)
	ctx := context.Background()

	// User not found
	_, err := store.IsLocked(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}

	// Add user - should not be locked initially
	store.AddUser("user1", "hash", "user")
	locked, err := store.IsLocked(ctx, "user1")
	if err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	}

	if locked {
		t.Error("Expected user not to be locked initially")
	}

	// Trigger lockout
	_ = store.RecordLoginFailure(ctx, "user1")
	_ = store.RecordLoginFailure(ctx, "user1")

	locked, err = store.IsLocked(ctx, "user1")
	if err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	}

	if !locked {
		t.Error("Expected user to be locked after max attempts")
	}

	// Wait for lock to expire
	time.Sleep(150 * time.Millisecond)

	locked, err = store.IsLocked(ctx, "user1")
	if err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	}

	if locked {
		t.Error("Expected lock to expire after lock duration")
	}
}

// TestCreateUser tests user creation.
func TestCreateUser(t *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	// Create new user
	err := store.CreateUser(ctx, "newuser", "hash123", "admin")
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	// Verify user exists
	hash, err := store.GetPasswordHash(ctx, "newuser")
	if err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}

	if hash != "hash123" {
		t.Errorf("Expected hash='hash123', got='%s'", hash)
	}

	// Try to create duplicate user
	err = store.CreateUser(ctx, "newuser", "anotherhash", "user")
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("Expected ErrUserExists, got: %v", err)
	}
}

// TestGetUserCount tests user count retrieval.
func TestGetUserCount(t *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	// Initially empty
	count, err := store.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("GetUserCount() error: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count=0, got=%d", count)
	}

	// Add some users
	store.AddUser("user1", "hash1", "admin")
	store.AddUser("user2", "hash2", "user")
	_ = store.CreateUser(ctx, "user3", "hash3", "user")

	count, err = store.GetUserCount(ctx)
	if err != nil {
		t.Fatalf("GetUserCount() error: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count=3, got=%d", count)
	}
}

// TestMemoryUserStoreConcurrency tests thread-safety.
func TestMemoryUserStoreConcurrency(_ *testing.T) {
	store := NewMemoryUserStore()
	ctx := context.Background()

	// Create initial user
	store.AddUser("user1", "hash", "user")

	// Run concurrent operations
	var wg sync.WaitGroup
	const iterations = 100

	// Concurrent reads
	wg.Go(func() {
		for range iterations {
			_, _ = store.GetPasswordHash(ctx, "user1")
			_, _ = store.GetTokenVersion(ctx, "user1")
			_, _ = store.IsLocked(ctx, "user1")
			_, _ = store.GetUserCount(ctx)
		}
	})

	// Concurrent writes
	wg.Go(func() {
		for range iterations {
			_ = store.RecordLoginFailure(ctx, "user1")
			_ = store.RecordLoginSuccess(ctx, "user1")
		}
	})

	// Concurrent user creation
	wg.Go(func() {
		for i := range iterations {
			username := "concurrent" + string(rune('A'+i%26))
			_ = store.CreateUser(ctx, username, "hash", "user")
		}
	})

	wg.Wait()

	// If we got here without panicking, thread safety is working
}

// TestMemoryUserStoreImplementsInterface verifies interface compliance.
func TestMemoryUserStoreImplementsInterface(_ *testing.T) {
	var _ UserStore = (*MemoryUserStore)(nil)
}

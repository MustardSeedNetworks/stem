// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/krisarmstrong/stem/internal/auth"
)

func TestNewMemoryUserStore(t *testing.T) {
	store := auth.NewMemoryUserStore()
	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	if store.MaxLoginAttempts() != auth.DefaultMaxLoginAttempts {
		t.Fatalf("Expected maxLoginAttempts=%d, got=%d", auth.DefaultMaxLoginAttempts, store.MaxLoginAttempts())
	}

	if store.LockDuration() != auth.DefaultLockDuration {
		t.Fatalf("Expected lockDuration=%v, got=%v", auth.DefaultLockDuration, store.LockDuration())
	}
}

func TestNewMemoryUserStoreWithConfig(t *testing.T) {
	const (
		maxAttempts  = 3
		lockDuration = 5 * time.Minute
	)

	store := auth.NewMemoryUserStoreWithConfig(maxAttempts, lockDuration)

	if store.MaxLoginAttempts() != maxAttempts {
		t.Fatalf("Expected maxLoginAttempts=%d, got=%d", maxAttempts, store.MaxLoginAttempts())
	}

	if store.LockDuration() != lockDuration {
		t.Fatalf("Expected lockDuration=%v, got=%v", lockDuration, store.LockDuration())
	}
}

func TestAddUser(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStore()
	store.AddUser("testuser", "hashedpassword", "admin")

	hash, err := store.GetPasswordHash(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}

	if hash != "hashedpassword" {
		t.Fatalf("Expected hash='hashedpassword', got='%s'", hash)
	}
}

func TestGetPasswordHash(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStore()
	if _, err := store.GetPasswordHash(ctx, "missing"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}

	store.AddUser("alice", "hash123", "user")
	hash, err := store.GetPasswordHash(ctx, "alice")
	if err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}

	if hash != "hash123" {
		t.Fatalf("Expected hash='hash123', got='%s'", hash)
	}
}

func TestGetTokenVersion(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStore()
	if _, err := store.GetTokenVersion(ctx, "missing"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}

	store.AddUser("alice", "hash123", "user")
	version, err := store.GetTokenVersion(ctx, "alice")
	if err != nil {
		t.Fatalf("GetTokenVersion() error: %v", err)
	}

	if version != 1 {
		t.Fatalf("Expected initial token version=1, got=%d", version)
	}
}

func TestUpdatePassword(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStore()
	if err := store.UpdatePassword(ctx, "missing", "newhash"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}

	store.AddUser("alice", "oldhash", "user")
	if err := store.UpdatePassword(ctx, "alice", "newhash"); err != nil {
		t.Fatalf("UpdatePassword() error: %v", err)
	}

	hash, err := store.GetPasswordHash(ctx, "alice")
	if err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}
	if hash != "newhash" {
		t.Fatalf("Expected hash='newhash', got='%s'", hash)
	}

	version, err := store.GetTokenVersion(ctx, "alice")
	if err != nil {
		t.Fatalf("GetTokenVersion() error: %v", err)
	}
	if version != 2 {
		t.Fatalf("Expected token version=2 after update, got=%d", version)
	}
}

func TestRecordLoginSuccess(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStoreWithConfig(3, time.Minute)
	if err := store.RecordLoginSuccess(ctx, "missing"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}

	store.AddUser("alice", "hash", "user")
	store.RecordLoginFailure(ctx, "alice")
	store.RecordLoginFailure(ctx, "alice")

	if err := store.RecordLoginSuccess(ctx, "alice"); err != nil {
		t.Fatalf("RecordLoginSuccess() error: %v", err)
	}

	locked, err := store.IsLocked(ctx, "alice")
	if err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	}
	if locked {
		t.Fatal("Expected user not to be locked after successful login")
	}
}

func TestRecordLoginFailure(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStoreWithConfig(3, time.Minute)
	if err := store.RecordLoginFailure(ctx, "missing"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}

	store.AddUser("alice", "hash", "user")
	for attempt := range [3]struct{}{} {
		if err := store.RecordLoginFailure(ctx, "alice"); err != nil {
			t.Fatalf("RecordLoginFailure() error on attempt %d: %v", attempt+1, err)
		}
	}

	locked, err := store.IsLocked(ctx, "alice")
	if err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	}
	if !locked {
		t.Fatal("Expected user to be locked after max failed attempts")
	}
}

func TestIsLocked(t *testing.T) {
	ctx := context.Background()

	store := auth.NewMemoryUserStoreWithConfig(2, 100*time.Millisecond)
	if _, err := store.IsLocked(ctx, "missing"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("Expected ErrUserNotFound, got: %v", err)
	}

	store.AddUser("alice", "hash", "user")
	for attempt := range [2]struct{}{} {
		if err := store.RecordLoginFailure(ctx, "alice"); err != nil {
			t.Fatalf("RecordLoginFailure() error on attempt %d: %v", attempt+1, err)
		}
	}

	if locked, err := store.IsLocked(ctx, "alice"); err != nil {
		t.Fatalf("IsLocked() error: %v", err)
	} else if !locked {
		t.Fatal("Expected user to be locked after hitting limit")
	}

	time.Sleep(150 * time.Millisecond)

	if locked, err := store.IsLocked(ctx, "alice"); err != nil {
		t.Fatalf("IsLocked() error after sleep: %v", err)
	} else if locked {
		t.Fatal("Expected lock to expire after duration")
	}
}

func TestCreateUserAndGetCount(t *testing.T) {
	ctx := context.Background()
	store := auth.NewMemoryUserStore()

	if err := store.CreateUser(ctx, "newuser", "hash123", "admin"); err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	if _, err := store.GetPasswordHash(ctx, "newuser"); err != nil {
		t.Fatalf("GetPasswordHash() error: %v", err)
	}

	if err := store.CreateUser(ctx, "newuser", "hash123", "admin"); !errors.Is(err, auth.ErrUserExists) {
		t.Fatalf("Expected ErrUserExists on duplicate create, got: %v", err)
	}

	store.AddUser("user1", "hash1", "admin")
	store.AddUser("user2", "hash2", "user")
	if count, err := store.GetUserCount(ctx); err != nil {
		t.Fatalf("GetUserCount() error: %v", err)
	} else if count != 3 {
		t.Fatalf("Expected count=3, got=%d", count)
	}
}

func TestMemoryUserStoreConcurrency(t *testing.T) {
	ctx := context.Background()
	store := auth.NewMemoryUserStore()

	store.AddUser("user1", "hash", "user")
	t.Helper()

	var wg sync.WaitGroup
	const concurrencyIterations = 100
	wg.Add(3)

	go func() {
		defer wg.Done()
		for range [concurrencyIterations]struct{}{} {
			_, _ = store.GetPasswordHash(ctx, "user1")
			_, _ = store.GetTokenVersion(ctx, "user1")
			_, _ = store.IsLocked(ctx, "user1")
			_, _ = store.GetUserCount(ctx)
		}
	}()

	go func() {
		defer wg.Done()
		for range [concurrencyIterations]struct{}{} {
			_ = store.RecordLoginFailure(ctx, "user1")
			_ = store.RecordLoginSuccess(ctx, "user1")
		}
	}()

	go func() {
		defer wg.Done()
		for idx := range [concurrencyIterations]struct{}{} {
			username := "concurrent" + string(rune('A'+idx%26))
			_ = store.CreateUser(ctx, username, "hash", "user")
		}
	}()

	wg.Wait()
}

func TestMemoryUserStoreImplementsInterface(t *testing.T) {
	var _ auth.UserStore = (*auth.MemoryUserStore)(nil)
	_ = t
}

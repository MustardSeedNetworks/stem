// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/auth"
)

const lockoutPassword = "kettle-orbit-granite-velvet-42"

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newLockoutManager(t *testing.T) (*auth.Manager, *fakeClock) {
	t.Helper()
	m, err := auth.NewManager("secret", time.Minute, "admin", lockoutPassword)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	t.Cleanup(m.Stop)
	clock := &fakeClock{now: time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)}
	auth.SetClockForTest(m, clock.Now)
	return m, clock
}

func failLogins(t *testing.T, m *auth.Manager, username string, n int) {
	t.Helper()
	for range n {
		_, _, err := m.AuthenticateWithRefresh(username, "wrong-password-entirely")
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf("wrong password: got %v, want ErrInvalidCredentials", err)
		}
	}
}

func TestLockout_LocksAfterMaxFailuresThenRefusesCorrectPassword(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)

	failLogins(t, m, "admin", auth.MaxFailedLogins-1)
	if m.LoginLocked() {
		t.Fatal("locked before the failure threshold")
	}
	if _, _, err := m.AuthenticateWithRefresh("admin", lockoutPassword); err != nil {
		t.Fatalf("correct password before the threshold: %v", err)
	}

	failLogins(t, m, "admin", 1)
	if !m.LoginLocked() {
		t.Fatal("not locked after the threshold")
	}
	if _, _, err := m.AuthenticateWithRefresh("admin", lockoutPassword); !errors.Is(err, auth.ErrAccountLocked) {
		t.Fatalf("correct password while locked: got %v, want ErrAccountLocked", err)
	}
	if _, err := m.Authenticate("admin", lockoutPassword); !errors.Is(err, auth.ErrAccountLocked) {
		t.Fatalf("Authenticate while locked: got %v, want ErrAccountLocked", err)
	}
}

func TestLockout_ExpiresAfterDurationAndRelocksOnNextFailure(t *testing.T) {
	t.Parallel()
	m, clock := newLockoutManager(t)
	failLogins(t, m, "admin", auth.MaxFailedLogins)

	clock.Advance(auth.LockoutDuration - time.Second)
	if !m.LoginLocked() {
		t.Fatal("lock expired early")
	}
	clock.Advance(2 * time.Second)
	if m.LoginLocked() {
		t.Fatal("lock did not expire")
	}

	failLogins(t, m, "admin", 1)
	if !m.LoginLocked() {
		t.Fatal("one failure after expiry must re-lock")
	}
}

func TestLockout_ResetClearsLock(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	failLogins(t, m, "admin", auth.MaxFailedLogins)
	m.ResetLoginFailures()
	if m.LoginLocked() {
		t.Fatal("still locked after reset")
	}
	if _, _, err := m.AuthenticateWithRefresh("admin", lockoutPassword); err != nil {
		t.Fatalf("login after reset: %v", err)
	}
}

func TestLockout_SettingAPasswordClearsLock(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	failLogins(t, m, "admin", auth.MaxFailedLogins)
	m.UpdatePasswordHash(t.Context(), m.GetPasswordHash())
	if m.LoginLocked() {
		t.Fatal("recovery or setup must clear the lock")
	}
}

func TestLockout_UnknownUsernameNeverCounts(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	failLogins(t, m, "someone-else", auth.MaxFailedLogins*3)
	if m.LoginLocked() {
		t.Fatal("failures for an unknown username locked the real account")
	}
}

func TestLockout_SuccessfulPasswordDoesNotClearCount(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	for range auth.MaxFailedLogins - 1 {
		m.RecordLoginFailure()
	}
	if _, _, err := m.AuthenticateWithRefresh("admin", lockoutPassword); err != nil {
		t.Fatalf("login: %v", err)
	}
	if !m.RecordLoginFailure() {
		t.Fatal("a correct password alone must not reset the count: a second factor may still be pending")
	}
}

func TestLockout_UnclaimedDaemonNeverLocks(t *testing.T) {
	t.Parallel()
	m, err := auth.NewManagerFromHash("secret", time.Minute, "admin", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Stop)
	failLogins(t, m, "admin", auth.MaxFailedLogins*2)
	if m.LoginLocked() {
		t.Fatal("an unclaimed daemon locked")
	}
}

// TestLockout_ConcurrentWrongAttemptsNeverExceedLimit: the password check is
// slow, so a burst of parallel guesses must be charged before it runs, not
// after.
func TestLockout_ConcurrentWrongAttemptsNeverExceedLimit(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)

	const attempts = 64
	var checked, refused atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range attempts {
		wg.Go(func() {
			<-start
			_, _, err := m.AuthenticateWithRefresh("admin", "wrong-password-entirely")
			switch {
			case errors.Is(err, auth.ErrInvalidCredentials):
				checked.Add(1)
			case errors.Is(err, auth.ErrAccountLocked):
				refused.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	close(start)
	wg.Wait()

	if got := checked.Load(); got > auth.MaxFailedLogins {
		t.Fatalf("%d password checks ran, limit is %d", got, auth.MaxFailedLogins)
	}
	if got := checked.Load() + refused.Load(); got != attempts {
		t.Fatalf("accounted for %d of %d attempts", got, attempts)
	}
	if !m.LoginLocked() {
		t.Fatal("account not locked after the burst")
	}
}

func TestLockout_AttemptSlotIsReleasedOnce(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	release, ok := m.TryBeginLoginAttempt()
	if !ok {
		t.Fatal("gate refused a clean account")
	}
	release(true)
	release(false) // a second call must not record a failure
	for range auth.MaxFailedLogins {
		next, admitted := m.TryBeginLoginAttempt()
		if !admitted {
			t.Fatal("slot leaked: gate refused before the limit")
		}
		next(true)
	}
}

func TestMFAPendingToken_RevokedAfterUse(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	token, err := m.GenerateMFAPendingToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.ValidateMFAPendingToken(t.Context(), token); err != nil {
		t.Fatalf("fresh token: %v", err)
	}
	m.RevokeMFAPendingToken(token)
	if _, err = m.ValidateMFAPendingToken(t.Context(), token); !errors.Is(err, auth.ErrMFATokenInvalid) {
		t.Fatalf("revoked token: got %v, want ErrMFATokenInvalid", err)
	}
}

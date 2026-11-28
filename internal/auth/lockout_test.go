// SPDX-License-Identifier: BUSL-1.1

package auth_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/password"

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
			case errors.Is(err, auth.ErrAccountLocked), errors.Is(err, auth.ErrTooManyAttempts):
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

// TestLockout_BogusUsernameFloodDoesNotLockRealAccount: invented names must
// take no slot, or a flood of them would lock the administrator out.
func TestLockout_BogusUsernameFloodDoesNotLockRealAccount(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)

	var hashes atomic.Int32
	auth.SetVerifyForTest(m, func(hash, plaintext string) (bool, error) {
		hashes.Add(1)
		return password.Verify(hash, plaintext)
	})

	// Waves no wider than the verify cap, so the cap itself refuses nothing.
	const waves, flood = 10, 10 * auth.MaxConcurrentVerifies
	for wave := range waves {
		var wg sync.WaitGroup
		for i := range auth.MaxConcurrentVerifies {
			wg.Go(func() {
				name := fmt.Sprintf("nobody-%d-%d", wave, i)
				_, _, err := m.AuthenticateWithRefresh(name, "wrong-password-entirely")
				if !errors.Is(err, auth.ErrInvalidCredentials) {
					t.Errorf("bogus username: got %v, want ErrInvalidCredentials", err)
				}
			})
		}
		wg.Wait()
	}

	if got := hashes.Load(); got != flood {
		t.Errorf("bogus usernames ran %d password hashes, want %d: timing would reveal which names exist", got, flood)
	}
	if m.LoginLocked() {
		t.Fatal("bogus usernames locked the real account")
	}
	if _, _, err := m.AuthenticateWithRefresh("admin", lockoutPassword); err != nil {
		t.Fatalf("real login after a bogus flood: %v", err)
	}
}

// TestLockout_LockedAttemptStillPaysForAHash: a locked real account must cost
// what an unknown username costs, or response time names the real account.
func TestLockout_LockedAttemptStillPaysForAHash(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	failLogins(t, m, "admin", auth.MaxFailedLogins)

	var hashes atomic.Int32
	auth.SetVerifyForTest(m, func(string, string) (bool, error) {
		hashes.Add(1)
		return false, nil
	})
	if _, _, err := m.AuthenticateWithRefresh("admin", lockoutPassword); !errors.Is(err, auth.ErrAccountLocked) {
		t.Fatalf("got %v, want ErrAccountLocked", err)
	}
	if got := hashes.Load(); got != 1 {
		t.Fatalf("locked attempt ran %d verifies, want 1 (dummy)", got)
	}
}

func TestVerifyGate_OneClientCannotExceedItsCap(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)

	var releases []func()
	for range auth.MaxVerifiesPerClient {
		release, err := m.AcquireVerify(t.Context(), "198.51.100.1")
		if err != nil {
			t.Fatalf("within the per-client cap: %v", err)
		}
		releases = append(releases, release)
	}
	// Over the cap: refused at once, not queued.
	begin := time.Now()
	if _, err := m.AcquireVerify(t.Context(), "198.51.100.1"); !errors.Is(err, auth.ErrTooManyAttempts) {
		t.Fatalf("over the per-client cap: got %v, want ErrTooManyAttempts", err)
	}
	if time.Since(begin) > time.Second {
		t.Error("per-client refusal waited instead of refusing at once")
	}

	// A different client is unaffected while the flood holds its slots.
	other, err := m.AcquireVerify(t.Context(), "198.51.100.2")
	if err != nil {
		t.Fatalf("a different client was refused during a flood: %v", err)
	}
	other()

	releases[0]()
	releases[0]() // releasing twice frees one slot, not two
	again, err := m.AcquireVerify(t.Context(), "198.51.100.1")
	if err != nil {
		t.Fatalf("after a release: %v", err)
	}
	if _, err = m.AcquireVerify(t.Context(), "198.51.100.1"); !errors.Is(err, auth.ErrTooManyAttempts) {
		t.Fatalf("double release freed two slots: got %v", err)
	}
	again()
	releases[1]()
}

func TestVerifyGate_GlobalBackstopQueuesThenRefuses(t *testing.T) {
	t.Parallel()
	m, _ := newLockoutManager(t)
	auth.SetVerifyWaitForTest(m, 300*time.Millisecond)

	releases := make([]func(), 0, auth.MaxConcurrentVerifies)
	for i := range auth.MaxConcurrentVerifies {
		release, err := m.AcquireVerify(t.Context(), fmt.Sprintf("203.0.113.%d", i))
		if err != nil {
			t.Fatalf("filling the backstop: %v", err)
		}
		releases = append(releases, release)
	}

	// Saturated: a waiter is queued, and gets in when a slot frees in time.
	time.AfterFunc(50*time.Millisecond, releases[0])
	queued, err := m.AcquireVerify(t.Context(), "192.0.2.50")
	if err != nil {
		t.Fatalf("a queued request was refused although a slot freed in time: %v", err)
	}

	// Still saturated and nothing frees: refused after the bounded wait.
	begin := time.Now()
	if _, err = m.AcquireVerify(t.Context(), "192.0.2.51"); !errors.Is(err, auth.ErrTooManyAttempts) {
		t.Fatalf("saturated: got %v, want ErrTooManyAttempts", err)
	}
	if waited := time.Since(begin); waited < 250*time.Millisecond {
		t.Errorf("refused after %v, want the bounded wait first", waited)
	}

	// A cancelled request stops waiting and is refused.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err = m.AcquireVerify(ctx, "192.0.2.52"); !errors.Is(err, auth.ErrTooManyAttempts) {
		t.Fatalf("cancelled context: got %v", err)
	}

	queued()
	for _, release := range releases[1:] {
		release()
	}
}

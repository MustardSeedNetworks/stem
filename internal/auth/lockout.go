// SPDX-License-Identifier: BUSL-1.1

package auth

import (
	"errors"
	"time"
)

// Account lockout matches seed's: five consecutive failures lock the
// account for fifteen minutes. The per-client limit (a per-IP token bucket
// in internal/api/ratelimit) is separate and still applies in front of this;
// the lockout is what stops a guess spread across many addresses.
const (
	// MaxFailedLogins is the number of consecutive failed logins that lock
	// the administrator account.
	MaxFailedLogins = 5
	// LockoutDuration is how long a locked account refuses password logins.
	LockoutDuration = 15 * time.Minute
)

// ErrAccountLocked is returned in place of a credential check while the
// account is locked. The HTTP layer answers it exactly like a wrong
// password, so a client learns nothing it could not from the timing; the
// distinct error exists so the audit trail can tell the two apart.
var ErrAccountLocked = errors.New("account locked after repeated failed logins")

// lockout counts failures against the one real administrator account. It is
// never keyed on a submitted username, so a flood of invented names cannot
// grow it. A failure keeps counting while locked, so each attempt that gets
// through after the lock expires re-locks at once.
type lockout struct {
	failures int
	until    time.Time
}

// LoginLocked reports whether password logins are currently refused.
func (m *Manager) LoginLocked() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.now().Before(m.lock.until)
}

// RecordLoginFailure counts one failed login and reports whether the
// account is now locked.
func (m *Manager) RecordLoginFailure() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lock.failures++
	if m.lock.failures >= MaxFailedLogins {
		m.lock.until = m.now().Add(LockoutDuration)
	}
	return m.now().Before(m.lock.until)
}

// ResetLoginFailures clears the count and any lock. A completed login resets
// it, and so does a password set through setup or recovery, which already
// proved control of the host.
func (m *Manager) ResetLoginFailures() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lock = lockout{}
}

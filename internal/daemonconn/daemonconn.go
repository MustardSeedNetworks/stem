// SPDX-License-Identifier: BUSL-1.1

// Package daemonconn carries the descriptor a running Stem daemon publishes
// so the local command line can reach it (#1166).
//
// `stem test` and `stem reflect` are clients of the daemon rather than a
// second orchestration path, and a systemd unit or script has no operator to
// prompt for a password — or to tell where the daemon ended up. Both facts
// live in one file: the bearer token the daemon minted for its own identity,
// and the URL it actually bound, which port fallback (#69) makes unguessable.
// Filesystem permissions are the grant, so the file is owner-only on write
// and refused on read if anything has loosened it.
package daemonconn

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// fileName is what the daemon publishes inside its data directory.
	fileName = ".daemon.json"

	// fileMode is the only mode the descriptor is written or read at.
	fileMode os.FileMode = 0o600

	// forbiddenMode is the group and world half of the mode: any bit set
	// there means another account can reach the token.
	forbiddenMode os.FileMode = 0o077
)

var (
	// ErrNotFound reports that no descriptor is present, normally because
	// no daemon is running on this host.
	ErrNotFound = errors.New("no running Stem daemon found")

	// ErrPermissions reports a descriptor other accounts can read or write,
	// which makes the token in it unfit to present.
	ErrPermissions = errors.New("daemon descriptor is not owner-only")

	// ErrUnreadable reports a descriptor that exists but this account may
	// not read — normally an operator invoking the CLI as themselves while
	// the daemon runs as its own user.
	ErrUnreadable = errors.New("daemon descriptor is not readable by this account")

	errNoToken = errors.New("daemon descriptor carries no token")
	errBadURL  = errors.New("daemon descriptor needs an absolute https URL")
)

// Descriptor is how a client reaches the daemon that published it.
type Descriptor struct {
	// URL is the base the daemon actually bound, scheme included.
	URL string `json:"url"`

	// Token authenticates as the daemon's own operator identity.
	Token string `json:"token"`

	// CAFile is the certificate to trust when the daemon serves its
	// self-signed default. Empty when a real certificate is configured, in
	// which case the system roots apply.
	CAFile string `json:"caFile,omitempty"`
}

// validate rejects a descriptor no client could safely use.
func (d Descriptor) validate() error {
	if strings.TrimSpace(d.Token) == "" {
		return errNoToken
	}
	parsed, err := url.Parse(d.URL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("%w: %q", errBadURL, d.URL)
	}
	return nil
}

// Path returns the descriptor file inside dataDir.
func Path(dataDir string) string {
	return filepath.Join(dataDir, fileName)
}

// Publish writes the descriptor, replacing any previous one. The
// replacement is atomic, so a client reading concurrently sees the old
// descriptor or the new one, never a half-written file.
func Publish(dataDir string, d Descriptor) error {
	if err := d.validate(); err != nil {
		return err
	}
	encoded, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("encode daemon descriptor: %w", err)
	}

	tmp, err := os.CreateTemp(dataDir, fileName+"-*")
	if err != nil {
		return fmt.Errorf("create daemon descriptor: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeds

	// CreateTemp is already 0600, but the mode is the whole security
	// control here, so it is asserted rather than assumed.
	if chmodErr := tmp.Chmod(fileMode); chmodErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("restrict daemon descriptor: %w", chmodErr)
	}
	if _, writeErr := tmp.Write(encoded); writeErr != nil {
		_ = tmp.Close()
		return fmt.Errorf("write daemon descriptor: %w", writeErr)
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return fmt.Errorf("close daemon descriptor: %w", closeErr)
	}
	if renameErr := os.Rename(tmpName, Path(dataDir)); renameErr != nil {
		return fmt.Errorf("publish daemon descriptor: %w", renameErr)
	}
	return nil
}

// Read returns the descriptor a local daemon published, refusing a file any
// other account could have read or tampered with.
func Read(dataDir string) (Descriptor, error) {
	path := Path(dataDir)

	info, statErr := os.Stat(path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return Descriptor{}, ErrNotFound
		}
		return Descriptor{}, fmt.Errorf("stat daemon descriptor: %w", statErr)
	}
	// Windows carries no POSIX mode bits, so the check there would reject
	// every file; access control is the ACL on the data directory.
	if runtime.GOOS != "windows" && info.Mode().Perm()&forbiddenMode != 0 {
		return Descriptor{}, fmt.Errorf("%w: mode %#o", ErrPermissions, info.Mode().Perm())
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		if errors.Is(readErr, os.ErrPermission) {
			return Descriptor{}, fmt.Errorf("%w: %s", ErrUnreadable, path)
		}
		return Descriptor{}, fmt.Errorf("read daemon descriptor: %w", readErr)
	}
	var d Descriptor
	if err := json.Unmarshal(data, &d); err != nil {
		return Descriptor{}, fmt.Errorf("decode daemon descriptor: %w", err)
	}
	if err := d.validate(); err != nil {
		return Descriptor{}, err
	}
	return d, nil
}

// Withdraw removes the descriptor. A daemon that has stopped leaves no
// credential behind pointing at whatever binds its port next.
func Withdraw(dataDir string) error {
	err := os.Remove(Path(dataDir))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("withdraw daemon descriptor: %w", err)
	}
	return nil
}

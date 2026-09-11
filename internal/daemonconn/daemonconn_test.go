// SPDX-License-Identifier: BUSL-1.1

package daemonconn_test

import (
	"errors"
	"os"
	"runtime"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/daemonconn"
)

func sample() daemonconn.Descriptor {
	return daemonconn.Descriptor{
		URL:    "https://127.0.0.1:8444",
		Token:  "sentinel-token",
		CAFile: "/var/lib/stem/certs/server.crt",
	}
}

func TestPublishRoundTrips(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := daemonconn.Publish(dir, sample()); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	got, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got != sample() {
		t.Errorf("descriptor = %+v, want %+v", got, sample())
	}
}

// The file carries a bearer token, so it is only ever readable by the
// account the daemon runs as.
func TestPublishIsOwnerOnly(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	dir := t.TempDir()
	if err := daemonconn.Publish(dir, sample()); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	info, err := os.Stat(daemonconn.Path(dir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode = %#o, want %#o", perm, 0o600)
	}
}

// A file another account can reach is not a credential this code will use.
func TestReadRejectsLoosenedFile(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	for name, mode := range map[string]os.FileMode{
		"group readable": 0o640,
		"world readable": 0o604,
		"world writable": 0o622,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if err := daemonconn.Publish(dir, sample()); err != nil {
				t.Fatalf("Publish: %v", err)
			}
			if err := os.Chmod(daemonconn.Path(dir), mode); err != nil {
				t.Fatalf("chmod: %v", err)
			}
			if _, err := daemonconn.Read(dir); !errors.Is(err, daemonconn.ErrPermissions) {
				t.Errorf("err = %v, want ErrPermissions", err)
			}
		})
	}
}

func TestReadMissingFile(t *testing.T) {
	t.Parallel()

	if _, err := daemonconn.Read(t.TempDir()); !errors.Is(err, daemonconn.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// A half-written or hand-edited file must not read back as a usable
// connection: the CLI would present an empty token to an unknown host.
func TestReadRejectsIncompleteDescriptor(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"empty object": `{}`,
		"no token":     `{"url":"https://127.0.0.1:8444"}`,
		"no url":       `{"token":"sentinel-token"}`,
		"blank token":  `{"url":"https://127.0.0.1:8444","token":"   "}`,
		"not json":     `token=abc`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if err := os.WriteFile(daemonconn.Path(dir), []byte(body), 0o600); err != nil {
				t.Fatalf("seed: %v", err)
			}
			if _, err := daemonconn.Read(dir); err == nil {
				t.Error("Read = nil error, want a rejection")
			}
		})
	}
}

// The token travels to the daemon over this URL. Plain HTTP would put a
// bearer credential on the wire in clear text, and stem has no HTTP
// listener to begin with.
func TestPublishRejectsNonHTTPSURL(t *testing.T) {
	t.Parallel()

	for name, url := range map[string]string{
		"http":   "http://127.0.0.1:8444",
		"empty":  "",
		"scheme": "127.0.0.1:8444",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := sample()
			d.URL = url
			if err := daemonconn.Publish(t.TempDir(), d); err == nil {
				t.Errorf("Publish(%q) = nil error, want a rejection", url)
			}
		})
	}
}

func TestPublishRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	d := sample()
	d.Token = " "
	if err := daemonconn.Publish(t.TempDir(), d); err == nil {
		t.Error("Publish(empty token) = nil error, want a rejection")
	}
}

func TestPublishReplacesAtomically(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, token := range []string{"first", "second", "third"} {
		d := sample()
		d.Token = token
		if err := daemonconn.Publish(dir, d); err != nil {
			t.Fatalf("Publish(%q): %v", token, err)
		}
	}

	got, err := daemonconn.Read(dir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Token != "third" {
		t.Errorf("token = %q, want %q", got.Token, "third")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only the descriptor", len(entries))
	}
}

func TestWithdrawIsIdempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := daemonconn.Publish(dir, sample()); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	for range 2 {
		if err := daemonconn.Withdraw(dir); err != nil {
			t.Fatalf("Withdraw: %v", err)
		}
	}
	if _, err := os.Stat(daemonconn.Path(dir)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stat err = %v, want ErrNotExist", err)
	}
}

// The daemon runs as its own account, so an operator invoking the CLI as
// themselves gets a permission error. It has to be distinguishable from
// "no daemon here" — the remedy is completely different.
func TestReadReportsAnUnreadableDescriptor(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permissions")
	}

	dir := t.TempDir()
	if err := daemonconn.Publish(dir, sample()); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if err := os.Chmod(daemonconn.Path(dir), 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	_, err := daemonconn.Read(dir)
	if !errors.Is(err, daemonconn.ErrUnreadable) {
		t.Errorf("err = %v, want ErrUnreadable", err)
	}
	if errors.Is(err, daemonconn.ErrNotFound) {
		t.Error("an unreadable descriptor must not read as a missing one")
	}
}

// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// The install table of ADR-0012: how the running binary was installed decides
// how a release is applied and how the service restarts. The operator never
// chooses it, and an install that matches no row is refused.

// ErrUnsupportedInstall reports a binary that no install method owns, such as
// one run from a build tree.
var ErrUnsupportedInstall = errors.New("this stem was not installed by a release package")

const (
	serviceName   = "stem"
	launchdTarget = "system/com.stem"
	darwinBinary  = "/usr/local/stem/stem"
	windowsBinary = "stem.exe"
	binaryMode    = 0o755
)

// Runner executes an external command and returns its combined output.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// ExecRunner runs commands with os/exec.
func ExecRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

// Method applies a verified release artifact for one install type.
type Method interface {
	// Name identifies the install type in messages and events.
	Name() string
	// Asset is the release asset this install type takes for version
	// (without the leading v) on goarch.
	Asset(version, goarch string) (string, error)
	// Apply installs the verified artifact at path.
	Apply(ctx context.Context, path string) error
	// Reapply reinstalls the previous version's verified artifact at path.
	Reapply(ctx context.Context, path string) error
	// Stop stops the service; Start starts it again.
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
	// Restarts is the number of times the service manager has restarted the
	// service on its own. A change during the health check means the new
	// version crashed.
	Restarts(ctx context.Context) (int, error)
}

// Detect returns the install method that owns exe on goos.
func Detect(ctx context.Context, run Runner, goos, exe string) (Method, error) {
	switch goos {
	case "linux":
		if _, err := run(ctx, "dpkg", "-S", exe); err == nil {
			return debMethod{systemd{run}}, nil
		}
		if _, err := run(ctx, "rpm", "-qf", exe); err == nil {
			return rpmMethod{systemd{run}}, nil
		}
		return nil, fmt.Errorf("%w: neither dpkg nor rpm owns %s", ErrUnsupportedInstall, exe)
	case "darwin":
		if exe != darwinBinary {
			return nil, fmt.Errorf("%w: the macOS install runs %s, not %s", ErrUnsupportedInstall, darwinBinary, exe)
		}
		return darwinMethod{run: run, exe: exe}, nil
	case "windows":
		return windowsMethod{exe: exe}, nil
	default:
		return nil, fmt.Errorf("%w: no install method for %s", ErrUnsupportedInstall, goos)
	}
}

// systemd is the service half of the .deb and .rpm installs.
type systemd struct{ run Runner }

func (s systemd) Stop(ctx context.Context) error {
	_, err := s.run(ctx, "systemctl", "stop", serviceName)
	return err
}

// Start clears the unit's start-rate counter first. The unit allows three
// starts a minute so a refused start cannot respin forever, and an update
// that rolls back starts the service four times (the package's postinst and
// this restart, twice). These starts are deliberate; a crash loop is still
// caught by the restart count.
func (s systemd) Start(ctx context.Context) error {
	if _, err := s.run(ctx, "systemctl", "reset-failed", serviceName); err != nil {
		return err
	}
	_, err := s.run(ctx, "systemctl", "restart", serviceName)
	return err
}

func (s systemd) Restarts(ctx context.Context) (int, error) {
	out, err := s.run(ctx, "systemctl", "show", "--property=NRestarts", "--value", serviceName)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

type debMethod struct{ systemd }

func (debMethod) Name() string { return "deb" }

func (debMethod) Asset(version, goarch string) (string, error) {
	return "stem_" + version + "_" + goarch + ".deb", nil
}

func (m debMethod) Apply(ctx context.Context, path string) error {
	_, err := m.run(ctx, "dpkg", "-i", path)
	return err
}

// Reapply is the same dpkg -i: dpkg installs an older version over a newer
// one, with a warning.
func (m debMethod) Reapply(ctx context.Context, path string) error {
	return m.Apply(ctx, path)
}

type rpmMethod struct{ systemd }

func (rpmMethod) Name() string { return "rpm" }

func (rpmMethod) Asset(version, goarch string) (string, error) {
	arch, ok := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[goarch]
	if !ok {
		return "", fmt.Errorf("%w: no rpm is published for %s", ErrUnsupportedInstall, goarch)
	}
	return "stem-" + version + "-1." + arch + ".rpm", nil
}

func (m rpmMethod) Apply(ctx context.Context, path string) error {
	_, err := m.run(ctx, "rpm", "-U", path)
	return err
}

func (m rpmMethod) Reapply(ctx context.Context, path string) error {
	_, err := m.run(ctx, "rpm", "-U", "--oldpackage", path)
	return err
}

// darwinMethod is the tarball installed under /usr/local/stem by the macOS
// package. That tree is also the daemon's data directory, so only the binary
// is replaced; the state copy covers everything else.
type darwinMethod struct {
	run Runner
	exe string
}

func (darwinMethod) Name() string { return "macos" }

func (darwinMethod) Asset(version, goarch string) (string, error) {
	return "stem-" + version + "-darwin-" + goarch + ".tar.gz", nil
}

func (m darwinMethod) Apply(_ context.Context, path string) error {
	return replaceBinary(m.exe, func(w io.Writer) error { return extractTarGz(path, "stem", w) })
}

func (m darwinMethod) Reapply(ctx context.Context, path string) error {
	return m.Apply(ctx, path)
}

// Stop uses SIGTERM: the plist restarts the daemon only after a crash, so a
// clean exit stays down until kickstart.
func (m darwinMethod) Stop(ctx context.Context) error {
	_, err := m.run(ctx, "launchctl", "kill", "SIGTERM", launchdTarget)
	return err
}

func (m darwinMethod) Start(ctx context.Context) error {
	_, err := m.run(ctx, "launchctl", "kickstart", "-k", launchdTarget)
	return err
}

// Restarts reads launchd's "runs" counter. Kickstart counts as a run, so the
// health check compares it after the start, never across one.
func (m darwinMethod) Restarts(ctx context.Context) (int, error) {
	out, err := m.run(ctx, "launchctl", "print", launchdTarget)
	if err != nil {
		return 0, err
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "runs = "); ok {
			return strconv.Atoi(value)
		}
	}
	return 0, errors.New("launchctl print reported no run count for " + launchdTarget)
}

// windowsMethod is the release zip. Windows ships no service, so there is
// nothing to stop or restart; a running executable can be renamed but not
// overwritten, so the old one is moved aside first.
type windowsMethod struct{ exe string }

func (windowsMethod) Name() string { return "windows" }

func (windowsMethod) Asset(version, goarch string) (string, error) {
	return "stem-" + version + "-windows-" + goarch + ".zip", nil
}

func (m windowsMethod) Apply(_ context.Context, path string) error {
	return replaceBinary(m.exe, func(w io.Writer) error { return extractZip(path, windowsBinary, w) })
}

func (m windowsMethod) Reapply(ctx context.Context, path string) error {
	return m.Apply(ctx, path)
}

func (windowsMethod) Stop(context.Context) error            { return nil }
func (windowsMethod) Start(context.Context) error           { return nil }
func (windowsMethod) Restarts(context.Context) (int, error) { return 0, nil }

// replaceBinary writes the new executable beside exe and renames it into
// place, so exe is never half-written. The previous file is moved to
// exe.old first, which Windows needs for a running executable; a leftover
// from an earlier update is replaced.
func replaceBinary(exe string, write func(io.Writer) error) error {
	next := exe + ".new"
	f, err := os.OpenFile(next, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, binaryMode)
	if err != nil {
		return err
	}
	if err = write(f); err != nil {
		return errors.Join(err, f.Close(), os.Remove(next))
	}
	if err = f.Close(); err != nil {
		return errors.Join(err, os.Remove(next))
	}
	old := exe + ".old"
	if err = os.Remove(old); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err = os.Rename(exe, old); err != nil {
		return err
	}
	if err = os.Rename(next, exe); err != nil {
		return errors.Join(err, os.Rename(old, exe))
	}
	return nil
}

func extractTarGz(archive, member string, w io.Writer) error {
	f, err := os.Open(archive) // #nosec G304 -- a verified release artifact in the update's staging directory
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			return fmt.Errorf("%s holds no %s", filepath.Base(archive), member)
		}
		if nextErr != nil {
			return nextErr
		}
		if hdr.Typeflag == tar.TypeReg && hdr.Name == member {
			_, err = io.Copy(w, tr) // #nosec G110 -- the archive is verified against the release signature
			return err
		}
	}
}

func extractZip(archive, member string, w io.Writer) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != member {
			continue
		}
		rc, openErr := f.Open()
		if openErr != nil {
			return openErr
		}
		_, copyErr := io.Copy(w, rc) // #nosec G110 -- the archive is verified against the release signature
		return errors.Join(copyErr, rc.Close())
	}
	return fmt.Errorf("%s holds no %s", filepath.Base(archive), member)
}

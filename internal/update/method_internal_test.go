// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func writeTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if err = tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err = tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []interface{ Close() error }{tw, gz, f} {
		if err = c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path) // #nosec G304 -- test fixture
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, createErr := zw.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, err = w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err = zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}

// The archive installs replace only the binary, take it from the archive's
// own member, and keep the old one as .old.
func TestArchiveInstallsReplaceTheBinary(t *testing.T) {
	for _, tc := range []struct {
		name    string
		method  func(exe string) Method
		archive string
		write   func(*testing.T, string, map[string]string)
		member  string
	}{
		{"macos", func(exe string) Method { return darwinMethod{exe: exe} }, "stem.tar.gz", writeTarGz, "stem"},
		{"windows", func(exe string) Method { return windowsMethod{exe: exe} }, "stem.zip", writeZip, windowsBinary},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			exe := filepath.Join(dir, "stem")
			if err := os.WriteFile(exe, []byte("old"), binaryMode); err != nil {
				t.Fatal(err)
			}
			archive := filepath.Join(dir, tc.archive)
			tc.write(t, archive, map[string]string{tc.member: "new", "stem.yaml": "config"})

			if err := tc.method(exe).Apply(t.Context(), archive); err != nil {
				t.Fatalf("Apply: %v", err)
			}
			for path, want := range map[string]string{exe: "new", exe + ".old": "old"} {
				got, err := os.ReadFile(path) // #nosec G304 -- test fixture
				if err != nil || string(got) != want {
					t.Fatalf("%s = %q, %v; want %q", filepath.Base(path), got, err, want)
				}
			}
			info, err := os.Stat(exe)
			if err != nil || info.Mode().Perm()&0o100 == 0 {
				t.Fatalf("installed binary mode = %v, %v; want executable", info.Mode(), err)
			}
		})
	}
}

func TestArchiveWithoutTheBinaryLeavesTheInstallAlone(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "stem")
	if err := os.WriteFile(exe, []byte("old"), binaryMode); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "stem.tar.gz")
	writeTarGz(t, archive, map[string]string{"stem.yaml": "config"})

	if err := (darwinMethod{exe: exe}).Apply(t.Context(), archive); err == nil {
		t.Fatal("Apply succeeded without a binary in the archive")
	}
	got, err := os.ReadFile(exe) // #nosec G304 -- test fixture
	if err != nil || string(got) != "old" {
		t.Fatalf("binary = %q, %v; want it untouched", got, err)
	}
	if _, err = os.Stat(exe + ".new"); !os.IsNotExist(err) {
		t.Fatalf("partial download left behind: %v", err)
	}
}

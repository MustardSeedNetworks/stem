// SPDX-License-Identifier: BUSL-1.1

package support_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/support"
)

// Each sentinel is planted in a place the bundle reads or could be tempted to
// read; none may appear anywhere in the decompressed archive.
func sentinels() map[string]string {
	return map[string]string{
		"password hash":   "SENTINEL-PASSWORD-HASH-7f3a",
		"administrator":   "SENTINEL-ADMIN-NAME-91bc",
		"descriptor":      "SENTINEL-DAEMON-TOKEN-22de",
		"licence file":    "MSN1.SENTINELLICENSEPAYLOAD.SENTINELLICENSESIG",
		"first-run token": "SENTINEL-SETUP-TOKEN-5a5a",
		"recovery token":  "SENTINEL-RECOVERY-TOKEN-6b6b",
		"private key":     "SENTINEL-PRIVATE-KEY-BYTES-0c0c",
		"environment":     "SENTINEL-JWT-SECRET-1d1d",
		"update source":   "SENTINEL-MIRROR-PASSWORD-3e3e",
		"log value":       "SENTINEL-ONLY-IN-THE-LOG-4f4f",
	}
}

func bundleDataDir(t *testing.T) string {
	t.Helper()
	s := sentinels()
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"credentials.json": `{"username":"` + s["administrator"] + `","passwordHash":"` + s["password hash"] + `"}`,
		"reflector.json":   `{"interface":"eth0","profile":"netally"}`,
		"update.json":      `{"automatic":true,"source":"https://deploy:` + s["update source"] + `@mirror.example/stem"}`,
		".daemon.json": `{"url":"https://127.0.0.1:8444","token":"` + s["descriptor"] + `","caFile":"` + filepath.Join(
			dir,
			"certs",
			"server.crt",
		) + `"}`,
		".config/stem/.license": s["licence file"] + "\n",
		".setup-token":          s["first-run token"],
		".recovery-token":       s["recovery token"],
		"certs/server.crt":      "-----BEGIN CERTIFICATE-----\nPUBLIC\n-----END CERTIFICATE-----\n",
		"certs/server.key":      "-----BEGIN PRIVATE KEY-----\n" + s["private key"] + "\n-----END PRIVATE KEY-----\n",
		"stem.db":               "a file that is not on the allow-list",
	})
	return dir
}

func unpack(t *testing.T, archive []byte) map[string]string {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for tr := tar.NewReader(gz); ; {
		header, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			return files
		}
		if nextErr != nil {
			t.Fatal(nextErr)
		}
		body, readErr := io.ReadAll(tr)
		if readErr != nil {
			t.Fatal(readErr)
		}
		files[header.Name] = string(body)
	}
}

func TestBundleCarriesNoCredential(t *testing.T) {
	s := sentinels()
	dir := bundleDataDir(t)
	log := strings.Join([]string{
		`{"level":"info","message":"started","component":"api"}`,
		`{"level":"warn","message":"retry","password":"` + s["log value"] + `"}`,
		"token=" + s["descriptor"],
		"Authorization: Bearer " + s["descriptor"],
		"jwt secret is " + s["environment"],
		"admin " + s["administrator"] + " signed in",
		"hash " + s["password hash"],
		"licence " + s["licence file"],
		"setup " + s["first-run token"] + " recovery " + s["recovery token"],
		"mirror https://deploy:" + s["update source"] + "@mirror.example/stem",
	}, "\n")
	status := `{"state":"running","detail":"` + s["descriptor"] + `"}`

	var archive bytes.Buffer
	err := support.WriteBundle(support.BundleOptions{
		DataDir:     dir,
		Log:         []byte(log),
		Status:      []byte(status),
		Environment: []string{"STEM_JWT_SECRET=" + s["environment"], "STEM_LOG_LEVEL=debug", "HOME=/root"},
		Now:         restoreTime(),
	}, &archive)
	if err != nil {
		t.Fatalf("WriteBundle: %v", err)
	}

	files := unpack(t, archive.Bytes())
	var everything strings.Builder
	for name, body := range files {
		everything.WriteString(name + "\n" + body + "\n")
	}
	for what, sentinel := range s {
		if strings.Contains(everything.String(), sentinel) {
			t.Errorf("the bundle carries the %s (%q)", what, sentinel)
		}
	}

	for _, want := range []string{
		"manifest.json", "environment.txt", "status.json", "logs/stem.log",
		"state/credentials.json", "state/reflector.json", "state/update.json", "state/daemon.json",
	} {
		if _, ok := files[want]; !ok {
			t.Errorf("the bundle lacks %s; has %v", want, keys(files))
		}
	}
	for name := range files {
		if strings.Contains(name, "setup") || strings.Contains(name, "recovery") || strings.Contains(name, "license") ||
			strings.Contains(name, "certs") || strings.Contains(name, "stem.db") {
			t.Errorf("the bundle carries a file that is not on the allow-list: %s", name)
		}
	}

	if !strings.Contains(files["state/credentials.json"], `"passwordHash"`) {
		t.Error("the credential file lost its structure")
	}
	if !strings.Contains(files["state/reflector.json"], "netally") {
		t.Error("a settings file with no credential in it was altered")
	}
	if !strings.Contains(files["logs/stem.log"], `"message":"started"`) {
		t.Error("the log lost an ordinary line")
	}
	if got := files["environment.txt"]; !strings.Contains(got, "STEM_LOG_LEVEL=debug") ||
		!strings.Contains(got, "STEM_JWT_SECRET=(set)") || strings.Contains(got, "HOME") {
		t.Errorf("environment.txt = %q", got)
	}
}

func TestBundleManifestInventoriesWithoutReading(t *testing.T) {
	dir := bundleDataDir(t)

	var archive bytes.Buffer
	if err := support.WriteBundle(support.BundleOptions{DataDir: dir, Now: restoreTime()}, &archive); err != nil {
		t.Fatalf("WriteBundle: %v", err)
	}
	var summary struct {
		CollectedAt  string   `json:"collectedAt"`
		Certificates []string `json:"certificates"`
		StateFiles   []struct {
			Name string `json:"name"`
		} `json:"stateFiles"`
		Version map[string]string `json:"version"`
	}
	if err := json.Unmarshal([]byte(unpack(t, archive.Bytes())["manifest.json"]), &summary); err != nil {
		t.Fatal(err)
	}

	if summary.CollectedAt != "2026-11-21T09:00:00Z" {
		t.Errorf("collectedAt = %q", summary.CollectedAt)
	}
	if strings.Join(summary.Certificates, ",") != "server.crt,server.key" {
		t.Errorf("certificates = %v", summary.Certificates)
	}
	var inventoried []string
	for _, entry := range summary.StateFiles {
		inventoried = append(inventoried, entry.Name)
	}
	if !strings.Contains(strings.Join(inventoried, ","), ".setup-token") {
		t.Errorf("the inventory does not list the first-run token by name: %v", inventoried)
	}
	if summary.Version["version"] == "" {
		t.Error("the manifest carries no version")
	}
}

func TestBundleSurvivesAnUnreadableDataDirectory(t *testing.T) {
	var archive bytes.Buffer
	missing := filepath.Join(t.TempDir(), "missing")
	if err := support.WriteBundle(support.BundleOptions{DataDir: missing, Now: restoreTime()}, &archive); err != nil {
		t.Fatalf("WriteBundle: %v", err)
	}
	if got := unpack(t, archive.Bytes())["manifest.json"]; !strings.Contains(got, "unreadable") {
		t.Errorf("the manifest does not say the data directory was unreadable: %s", got)
	}
}

func TestBundleDoesNotCarryAFileThatIsNotJSON(t *testing.T) {
	s := sentinels()
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"credentials.json": "password=" + s["password hash"]})

	var archive bytes.Buffer
	if err := support.WriteBundle(support.BundleOptions{DataDir: dir, Now: restoreTime()}, &archive); err != nil {
		t.Fatalf("WriteBundle: %v", err)
	}
	files := unpack(t, archive.Bytes())
	if _, carried := files["state/credentials.json"]; carried {
		t.Error("a credential file that does not parse was carried")
	}
	if got := files["manifest.json"]; strings.Contains(got, s["password hash"]) ||
		!strings.Contains(got, "credentials.json") {
		t.Errorf("manifest = %s", got)
	}
	_ = os.Remove(dir)
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

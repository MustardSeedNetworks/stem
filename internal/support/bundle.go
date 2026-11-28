// SPDX-License-Identifier: BUSL-1.1

package support

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/MustardSeedNetworks/stem/internal/version"
)

// BundleOptions names what a support bundle collects. Everything is optional:
// a bundle taken where the daemon has never run still carries the manifest,
// which is the part support asks for first.
type BundleOptions struct {
	// DataDir is the daemon's data directory. Its settings files are carried
	// redacted; its credential, certificates and activation state are listed
	// by name and size and never read into the archive.
	DataDir string

	// Log is the log text to include. It is scrubbed here; the caller does not
	// need to.
	Log []byte

	// Status is the `stem status --json` document taken when the bundle was
	// collected, so support sees whether the daemon answered.
	Status []byte

	// Environment is the process environment, as KEY=VALUE pairs.
	Environment []string

	// Now is the collection time.
	Now time.Time
}

// settingsFiles are the data-directory files carried in the bundle, redacted,
// and whether every string in them is removed or only the ones whose key names
// a credential. The credential store is not among them: it is sealed, so there
// is nothing in it to redact, and it is listed by name and size only. Anything
// not listed here is never read:
// the bundle is built from an allow-list, so a file a later release adds to the
// data directory stays out until someone decides what is safe in it.
func settingsFiles() []settingsFile {
	return []settingsFile{
		{"reflector.json", false},
		{"update.json", false},
		{".daemon.json", false},
	}
}

type settingsFile struct {
	name        string
	everyString bool
}

// secretFiles are read only so their contents can be scrubbed out of the log by
// value. They are never written to the bundle.
func secretFiles() []string {
	return []string{".config/stem/.license", ".setup-token", ".recovery-token"}
}

// inventoryNames are listed in the manifest by name and size.
func inventoryNames() []string {
	return []string{
		".config/stem/.license", ".daemon.json", ".setup-token", ".recovery", ".recovery-token",
		"auth/credentials.sealed", "auth/credential.key", "reflector.json", "update.json",
	}
}

// safeEnvironment names the variables whose values are shown. Every other
// STEM_ variable is listed as set, and its value is scrubbed from the log:
// the JWT secret, the bootstrap credential and the licence key are all
// environment variables.
func safeEnvironment() []string {
	return []string{"STEM_DATA_DIR", "STEM_LOG_FORMAT", "STEM_LOG_LEVEL", "STEM_TLS_CERT", "STEM_TLS_KEY"}
}

type manifest struct {
	CollectedAt  string            `json:"collectedAt"`
	Version      map[string]string `json:"version"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	GoVersion    string            `json:"goVersion"`
	DataDir      string            `json:"dataDir"`
	StateFiles   []inventoryEntry  `json:"stateFiles"`
	Certificates []string          `json:"certificates,omitempty"`
	Unreadable   []string          `json:"unreadable,omitempty"`
}

type inventoryEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Mode string `json:"mode"`
}

// WriteBundle collects diagnostics into a gzipped tar on w.
//
// Nothing reaches the archive unredacted: settings files are redacted over
// their structure, and the log and status text are scrubbed of every
// credential value found in the settings, the secret files and the
// environment, plus anything still shaped like a credential. Private keys are
// never opened.
func WriteBundle(options BundleOptions, w io.Writer) error {
	gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("open gzip writer: %w", err)
	}
	tw := tar.NewWriter(gz)

	if collectErr := collect(tw, options); collectErr != nil {
		return collectErr
	}
	if closeErr := tw.Close(); closeErr != nil {
		return fmt.Errorf("close tar: %w", closeErr)
	}
	if closeErr := gz.Close(); closeErr != nil {
		return fmt.Errorf("close gzip: %w", closeErr)
	}
	return nil
}

func collect(tw *tar.Writer, options BundleOptions) error {
	summary := manifest{
		CollectedAt: options.Now.UTC().Format(time.RFC3339),
		Version:     version.Info(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoVersion:   runtime.Version(),
		DataDir:     options.DataDir,
	}
	secrets := environmentSecrets(options.Environment)

	root, err := os.OpenRoot(options.DataDir)
	if err != nil {
		summary.Unreadable = append(summary.Unreadable, "data directory: "+err.Error())
	} else {
		defer func() { _ = root.Close() }()
		secrets = append(secrets, secretFileValues(root)...)
		summary.StateFiles = inventory(root)
		summary.Certificates = certificateNames(root)

		for _, file := range settingsFiles() {
			redacted, removed, redactErr := redactSettings(root, file.name, file.everyString)
			switch {
			case errors.Is(redactErr, fs.ErrNotExist):
				continue
			case redactErr != nil:
				summary.Unreadable = append(summary.Unreadable, file.name+": "+redactErr.Error())
				continue
			}
			secrets = append(secrets, removed...)
			if writeErr := writeFile(
				tw,
				path.Join("state", strings.TrimPrefix(file.name, ".")),
				redacted,
			); writeErr != nil {
				return writeErr
			}
		}
	}

	encoded, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("render manifest: %w", err)
	}
	if writeErr := writeFile(tw, "manifest.json", append(encoded, '\n')); writeErr != nil {
		return writeErr
	}
	if writeErr := writeFile(tw, "environment.txt", []byte(environmentListing(options.Environment))); writeErr != nil {
		return writeErr
	}
	if len(options.Status) > 0 {
		if writeErr := writeFile(
			tw,
			"status.json",
			[]byte(ScrubText(string(options.Status), secrets)),
		); writeErr != nil {
			return writeErr
		}
	}
	if len(options.Log) > 0 {
		return writeFile(tw, "logs/stem.log", []byte(ScrubText(string(options.Log), secrets)))
	}
	return nil
}

// redactSettings reads one settings file and returns it redacted, with the
// credential values removed from it.
func redactSettings(root *os.Root, name string, everyString bool) ([]byte, []string, error) {
	data, err := root.ReadFile(name)
	if err != nil {
		return nil, nil, err
	}
	redacted, removed, redactErr := RedactJSON(data, everyString)
	if redactErr != nil {
		// A file that is not the JSON it should be may hold anything, so it is
		// not carried at all; the manifest says it was there.
		return nil, nil, fmt.Errorf("not carried, %d bytes that do not parse: %w", len(data), redactErr)
	}
	return redacted, removed, nil
}

// secretFileValues returns the contents of the files that are secrets in
// themselves, to be scrubbed from the log by value.
func secretFileValues(root *os.Root) []string {
	var values []string
	for _, name := range secretFiles() {
		if data, err := root.ReadFile(name); err == nil {
			values = append(values, strings.TrimSpace(string(data)))
		}
	}
	return values
}

// inventory lists the state files present, by name, size and mode.
func inventory(root *os.Root) []inventoryEntry {
	var entries []inventoryEntry
	for _, name := range inventoryNames() {
		if info, err := root.Lstat(name); err == nil {
			entries = append(entries, inventoryEntry{Name: name, Size: info.Size(), Mode: info.Mode().String()})
		}
	}
	return entries
}

// certificateNames lists what is in certs/ without opening any of it: the
// directory holds the private key.
func certificateNames(root *os.Root) []string {
	entries, err := fs.ReadDir(root.FS(), stateDir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// environmentSecrets returns the values of every STEM_ variable not known to be
// safe to show.
func environmentSecrets(environment []string) []string {
	var secrets []string
	for _, pair := range environment {
		name, value, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(name, "STEM_") && !isSafeVariable(name) {
			secrets = append(secrets, value)
		}
	}
	return secrets
}

func environmentListing(environment []string) string {
	var out strings.Builder
	for _, pair := range environment {
		name, value, _ := strings.Cut(pair, "=")
		switch {
		case !strings.HasPrefix(name, "STEM_"):
			continue
		case isSafeVariable(name):
			fmt.Fprintf(&out, "%s=%s\n", name, ScrubText(value, nil))
		default:
			fmt.Fprintf(&out, "%s=(set)\n", name)
		}
	}
	return out.String()
}

func isSafeVariable(name string) bool {
	return slices.Contains(safeEnvironment(), name)
}

func writeFile(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name:     name,
		Mode:     archiveFileMode,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
		Format:   tar.FormatPAX,
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write header %s: %w", name, err)
	}
	if _, err := tw.Write(content); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

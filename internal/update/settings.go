// SPDX-License-Identifier: BUSL-1.1

package update

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// SettingsName is the file in the data directory that holds the automatic
// update opt-in. The root-owned update timer reads it; it is given the data
// directory's owner so the service can change it from Settings.
const SettingsName = "update.json"

// DefaultSource is where releases come from unless an operator names another.
const DefaultSource = "github"

// maxSettingsBytes bounds the read: the timer runs as root and the file is
// writable by the service account.
const maxSettingsBytes = 4 << 10

// settingsMode lets the service's group read the opt-in.
const settingsMode = 0o640

// Settings is the persisted automatic update opt-in (ADR-0012). The zero
// value is the shipped default: off, from GitHub. While Automatic is false
// the update timer exits before any network call.
type Settings struct {
	Automatic bool   `json:"automatic"`
	Source    string `json:"source,omitempty"`
}

// SourceSpec returns the source to update from.
func (s Settings) SourceSpec() string {
	if s.Source == "" {
		return DefaultSource
	}
	return s.Source
}

// LoadSettings reads the opt-in from dataDir. A missing file is the default.
// The read stays inside dataDir: a symlink pointing out of it is refused.
func LoadSettings(dataDir string) (Settings, error) {
	var settings Settings
	root, err := os.OpenRoot(dataDir)
	if err != nil {
		return settings, err
	}
	defer root.Close()
	file, err := root.Open(SettingsName)
	if errors.Is(err, fs.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return settings, fmt.Errorf("read the update settings: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSettingsBytes+1))
	if err != nil {
		return settings, fmt.Errorf("read the update settings: %w", err)
	}
	if len(data) > maxSettingsBytes {
		return settings, fmt.Errorf("the update settings in %s are larger than %d bytes",
			filepath.Join(dataDir, SettingsName), maxSettingsBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&settings); err != nil {
		return Settings{}, fmt.Errorf("the update settings in %s are not valid: %w",
			filepath.Join(dataDir, SettingsName), err)
	}
	return settings, nil
}

// SaveSettings replaces the opt-in in dataDir atomically and gives the file
// the data directory's owner.
func SaveSettings(dataDir string, settings Settings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(dataDir)
	if err != nil {
		return err
	}
	defer root.Close()
	tmp := SettingsName + ".tmp"
	file, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, settingsMode)
	if err != nil {
		return fmt.Errorf("write the update settings: %w", err)
	}
	_, err = file.Write(append(data, '\n'))
	err = errors.Join(err, file.Close())
	if err == nil {
		err = shareWithDataDir(filepath.Join(dataDir, tmp), dataDir)
	}
	if err == nil {
		err = root.Rename(tmp, SettingsName)
	}
	if err != nil {
		return errors.Join(fmt.Errorf("write the update settings: %w", err), root.Remove(tmp))
	}
	return nil
}

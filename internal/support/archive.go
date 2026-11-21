// SPDX-License-Identifier: BUSL-1.1

// Package support implements the operator-facing recovery and diagnostics
// surfaces: a backup of the daemon's state, its restore, and a support bundle
// that carries diagnostics with every credential removed.
//
// A backup carries the files the daemon itself keeps in its data directory and
// nothing it regenerates: the administrator credential, the reflector and
// update settings, the TLS pair, the activation state and the published
// descriptor. The instance lock, the run lock and the first-run and recovery
// tokens stay out, both ways. They belong to a running process or grant access
// by themselves, and a backup that carried them would turn a stolen archive
// into a stolen admin session.
package support

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/MustardSeedNetworks/foundation/pkg/instance"

	"github.com/MustardSeedNetworks/stem/internal/update"
)

var (
	// ErrUnsafeEntry reports an archive entry that names a path outside the
	// data directory, one a stem backup never carries, or a type the archive
	// is not allowed to hold.
	ErrUnsafeEntry = errors.New("unsafe archive entry")

	// ErrNothingToBackUp reports a data directory with no stem state in it,
	// which almost always means the wrong directory was named.
	ErrNothingToBackUp = errors.New("no stem state found")

	// ErrEmptyArchive reports an archive that holds no stem state to restore.
	ErrEmptyArchive = errors.New("archive holds no stem state")

	// ErrDaemonRunning reports a restore refused because a daemon holds the
	// data directory.
	ErrDaemonRunning = errors.New("the daemon is running")
)

// stateFiles are the single files a backup carries, relative to the data
// directory. The activation state sits under the service account's home, which
// the packaged units point at the data directory.
func stateFiles() []string {
	return []string{
		".config/stem/.license",
		".daemon.json",
		"credentials.json",
		"reflector.json",
		"update.json",
	}
}

// stateDir is the one directory a backup carries recursively: the daemon's
// own TLS pair.
const stateDir = "certs"

const (
	// archiveFileMode and archiveDirMode are what every entry is written
	// with, and what it is restored as. The data directory is the service
	// account's alone, and the daemon refuses a descriptor or credential
	// anyone else could read, so a mode wider than this would break the
	// restored daemon rather than make the archive more portable.
	archiveFileMode = 0o600
	archiveDirMode  = 0o700

	// maxEntries and maxEntryBytes bound what a restore will expand, so a
	// crafted archive cannot fill the disk. A real backup holds a handful of
	// files, each well under a megabyte.
	maxEntries    = 256
	maxEntryBytes = 8 << 20

	stagingPrefix = ".restore-"
)

// managed reports whether name, a slash-separated path relative to the data
// directory, is part of what a backup carries.
func managed(name string) bool {
	if slices.Contains(stateFiles(), name) {
		return true
	}
	return name == stateDir || strings.HasPrefix(name, stateDir+"/")
}

// managedDir reports whether name may appear as a directory entry: the certs
// directory, and the parents of the activation file.
func managedDir(name string) bool {
	return name == ".config" || name == ".config/stem" || name == stateDir || strings.HasPrefix(name, stateDir+"/")
}

// openDataDir opens the data directory as an [os.Root] without following a
// link at its own name: the root is opened on the parent and the base name is
// checked with Lstat, so a data directory that has been swapped for a symlink
// is refused rather than followed to wherever it points.
func openDataDir(dataDir string) (*os.Root, fs.FileInfo, error) {
	absolute, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve %s: %w", dataDir, err)
	}
	parent, err := os.OpenRoot(filepath.Dir(absolute))
	if err != nil {
		return nil, nil, fmt.Errorf("open the parent of %s: %w", dataDir, err)
	}
	defer func() { _ = parent.Close() }()

	info, err := parent.Lstat(filepath.Base(absolute))
	switch {
	case err != nil:
		return nil, nil, fmt.Errorf("data directory %s: %w", dataDir, err)
	case info.Mode()&fs.ModeSymlink != 0:
		return nil, nil, fmt.Errorf("data directory %s is a symbolic link; name the directory itself", dataDir)
	case !info.IsDir():
		return nil, nil, fmt.Errorf("data directory %s is not a directory", dataDir)
	}
	root, err := parent.OpenRoot(filepath.Base(absolute))
	if err != nil {
		return nil, nil, fmt.Errorf("open data directory: %w", err)
	}
	return root, info, nil
}

// holdDataDir takes the instance lock the daemon holds for its whole life. A
// copy taken under a running daemon is not a backup, and a restore under one
// replaces files it holds open, so both refuse.
func holdDataDir(dataDir, action string) (*instance.Lock, error) {
	lock, err := instance.Acquire(dataDir)
	if err != nil {
		if held, isHeld := errors.AsType[*instance.HeldError](err); isHeld {
			return nil, fmt.Errorf("%w (%w): stop it before %s", ErrDaemonRunning, held, action)
		}
		return nil, err
	}
	return lock, nil
}

// Backup writes the daemon's state under dataDir into w as a gzipped tar. The
// daemon must not be running.
//
// The archive is deterministic: entries come in lexical order and carry no
// modification time or owner, so two backups of unchanged state are
// byte-identical.
func Backup(dataDir string, w io.Writer) error {
	root, _, err := openDataDir(dataDir)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	lock, err := holdDataDir(dataDir, "backing up")
	if err != nil {
		return err
	}
	defer func() { _ = lock.Release() }()

	names, err := presentState(root)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return fmt.Errorf("%w in %s", ErrNothingToBackUp, dataDir)
	}

	gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("open gzip writer: %w", err)
	}
	tw := tar.NewWriter(gz)
	for _, name := range names {
		if writeErr := writeEntry(tw, root, name); writeErr != nil {
			return writeErr
		}
	}
	if closeErr := tw.Close(); closeErr != nil {
		return fmt.Errorf("close tar: %w", closeErr)
	}
	if closeErr := gz.Close(); closeErr != nil {
		return fmt.Errorf("close gzip: %w", closeErr)
	}
	return nil
}

// presentState lists the entries of the managed set that exist, in the order
// they are archived: a directory before what it holds.
func presentState(root *os.Root) ([]string, error) {
	var names []string
	for _, name := range stateFiles() {
		info, err := root.Lstat(name)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return nil, fmt.Errorf("inspect %s: %w", name, err)
		case info.Mode().IsRegular():
			names = append(names, name)
		}
	}

	walkErr := fs.WalkDir(root.FS(), stateDir, func(name string, entry fs.DirEntry, err error) error {
		switch {
		case errors.Is(err, fs.ErrNotExist) && name == stateDir:
			return fs.SkipAll
		case err != nil:
			return err
		case entry.IsDir() || entry.Type().IsRegular():
			// A link or device in certs/ is skipped, never followed.
			names = append(names, name)
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("list %s: %w", stateDir, walkErr)
	}
	slices.Sort(names)
	return names, nil
}

func writeEntry(tw *tar.Writer, root *os.Root, name string) error {
	info, err := root.Lstat(name)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", name, err)
	}
	header := &tar.Header{Name: name, Mode: archiveFileMode, Typeflag: tar.TypeReg, Format: tar.FormatPAX}
	if info.IsDir() {
		header.Name = name + "/"
		header.Mode = archiveDirMode
		header.Typeflag = tar.TypeDir
	} else {
		header.Size = info.Size()
	}
	if headerErr := tw.WriteHeader(header); headerErr != nil {
		return fmt.Errorf("write header %s: %w", name, headerErr)
	}
	if info.IsDir() {
		return nil
	}

	file, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer func() { _ = file.Close() }()
	if _, copyErr := io.CopyN(tw, file, header.Size); copyErr != nil {
		return fmt.Errorf("copy %s: %w", name, copyErr)
	}
	return nil
}

// Restore replaces the daemon's state under dataDir with the archive read from
// r, and returns where the previous state was kept.
//
// The daemon must not be running: the instance lock the daemon holds for its
// whole life is taken here and kept until the restore ends, so a daemon cannot
// start halfway through either. The archive is expanded and vetted in full in a
// staging directory before any live file is touched, and every create, rename
// and removal goes through an [os.Root] on the data directory, so no link in it
// is followed. The previous state is copied beside the data directory first.
func Restore(r io.Reader, dataDir string, now time.Time) (string, error) {
	root, dirInfo, err := openDataDir(dataDir)
	if err != nil {
		return "", fmt.Errorf("%w: restore into the directory the service uses", err)
	}
	defer func() { _ = root.Close() }()

	lock, err := holdDataDir(dataDir, "restoring")
	if err != nil {
		return "", err
	}
	defer func() { _ = lock.Release() }()

	if vetErr := vetTargets(root, dirInfo); vetErr != nil {
		return "", vetErr
	}

	previous := filepath.Clean(dataDir) + ".pre-restore-" + now.UTC().Format("20060102T150405Z")
	if _, statErr := os.Lstat(previous); statErr == nil {
		return "", fmt.Errorf("%s already exists: a restore ran within the same second", previous)
	}
	if copyErr := update.CopyState(dataDir, previous); copyErr != nil {
		_ = os.RemoveAll(previous)
		return "", fmt.Errorf("keep a copy of the current state: %w", copyErr)
	}

	stagedName := fmt.Sprintf("%s%d-%d", stagingPrefix, os.Getpid(), now.UnixNano())
	if mkdirErr := root.Mkdir(stagedName, archiveDirMode); mkdirErr != nil {
		_ = os.RemoveAll(previous)
		return "", fmt.Errorf("stage the restore: %w", mkdirErr)
	}
	defer func() { _ = root.RemoveAll(stagedName) }()

	files, err := expand(r, root, stagedName)
	if err != nil {
		_ = os.RemoveAll(previous)
		return "", err
	}
	if swapErr := swap(root, dirInfo, stagedName, files); swapErr != nil {
		return previous, fmt.Errorf("%w (the state before the restore is kept at %s)", swapErr, previous)
	}
	return previous, nil
}

// vetTargets refuses a restore over state it should not touch: a managed path
// that is a link, or a file owned by someone other than the data directory's
// owner. Replacing either would act on whatever the link points at, or on a
// file another account put there.
func vetTargets(root *os.Root, dirInfo fs.FileInfo) error {
	dirUID, _, dirOwned := ownerOf(dirInfo)
	paths := append(stateFiles(), stateDir)
	for _, name := range paths {
		info, err := root.Lstat(name)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return fmt.Errorf("inspect %s: %w", name, err)
		case info.Mode()&fs.ModeSymlink != 0:
			return fmt.Errorf("%s is a symbolic link: refusing to restore over it", name)
		}
		if uid, _, owned := ownerOf(info); dirOwned && owned && uid != dirUID {
			return fmt.Errorf(
				"%s is owned by uid %d, not the data directory's owner (uid %d): refusing to restore over it",
				name,
				uid,
				dirUID,
			)
		}
	}
	return nil
}

// expand writes the archive into the staging directory and returns the regular
// files it wrote. An entry that names its way outside the directory fails
// instead of writing there.
func expand(r io.Reader, data *os.Root, stagedName string) ([]string, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("open gzip reader: %w", err)
	}
	defer func() { _ = gz.Close() }()

	root, err := data.OpenRoot(stagedName)
	if err != nil {
		return nil, fmt.Errorf("open staging directory: %w", err)
	}
	defer func() { _ = root.Close() }()

	var files []string
	tr := tar.NewReader(gz)
	for count := 0; ; count++ {
		header, nextErr := tr.Next()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return nil, fmt.Errorf("read archive: %w", nextErr)
		}
		if count >= maxEntries {
			return nil, fmt.Errorf("%w: more than %d entries", ErrUnsafeEntry, maxEntries)
		}
		name, isDir, entryErr := entryName(header)
		if entryErr != nil {
			return nil, entryErr
		}
		if isDir {
			if mkdirErr := root.MkdirAll(name, archiveDirMode); mkdirErr != nil {
				return nil, fmt.Errorf("create %s: %w", name, mkdirErr)
			}
			continue
		}
		if extractErr := extractFile(root, tr, name, header.Size); extractErr != nil {
			return nil, extractErr
		}
		files = append(files, name)
	}
	if len(files) == 0 {
		return nil, ErrEmptyArchive
	}
	return files, nil
}

// entryName vets one header and returns its slash-separated name. A backup
// this package wrote holds only regular files and directories of the managed
// set, so anything else is refused outright, whether it is hostile or merely
// not a stem backup.
func entryName(header *tar.Header) (string, bool, error) {
	isDir := header.Typeflag == tar.TypeDir
	if header.Typeflag != tar.TypeReg && !isDir {
		return "", false, fmt.Errorf("%w: %s is type %q", ErrUnsafeEntry, header.Name, header.Typeflag)
	}
	if path.IsAbs(header.Name) || filepath.IsAbs(header.Name) || strings.Contains(header.Name, `\`) {
		return "", false, fmt.Errorf("%w: %s is not a relative path", ErrUnsafeEntry, header.Name)
	}
	name := path.Clean(strings.TrimSuffix(header.Name, "/"))
	if name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return "", false, fmt.Errorf("%w: %s escapes the data directory", ErrUnsafeEntry, header.Name)
	}
	if (isDir && !managedDir(name)) || (!isDir && !managed(name)) {
		return "", false, fmt.Errorf("%w: %s is not part of a stem backup", ErrUnsafeEntry, header.Name)
	}
	if header.Size < 0 || header.Size > maxEntryBytes {
		return "", false, fmt.Errorf("%w: %s is %d bytes", ErrUnsafeEntry, header.Name, header.Size)
	}
	return name, isDir, nil
}

func extractFile(root *os.Root, tr io.Reader, name string, size int64) error {
	if parent := path.Dir(name); parent != "." {
		if err := root.MkdirAll(parent, archiveDirMode); err != nil {
			return fmt.Errorf("create %s: %w", parent, err)
		}
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, archiveFileMode)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	if _, copyErr := io.CopyN(file, tr, size); copyErr != nil {
		_ = file.Close()
		return fmt.Errorf("write %s: %w", name, copyErr)
	}
	if closeErr := file.Close(); closeErr != nil {
		return fmt.Errorf("write %s: %w", name, closeErr)
	}
	return nil
}

// swap moves the staged files into the data directory and removes the managed
// files the archive does not carry, so the restored state is the archived
// state and not a mixture of it and what was there. Each file arrives by
// rename, whole, and takes the data directory's owner.
func swap(root *os.Root, dirInfo fs.FileInfo, stagedName string, files []string) error {
	for _, name := range files {
		if err := install(root, dirInfo, path.Join(stagedName, name), name); err != nil {
			return err
		}
	}
	return removeStale(root, files)
}

// install renames one staged file to name, creating its parent directory first,
// and gives both the data directory's owner.
func install(root *os.Root, dirInfo fs.FileInfo, staged, name string) error {
	if parent := path.Dir(name); parent != "." {
		if err := root.MkdirAll(parent, archiveDirMode); err != nil {
			return fmt.Errorf("create %s: %w", parent, err)
		}
		if err := giveOwner(root, parent, dirInfo); err != nil {
			return fmt.Errorf("own %s: %w", parent, err)
		}
	}
	if err := root.Rename(staged, name); err != nil {
		return fmt.Errorf("install %s: %w", name, err)
	}
	if err := giveOwner(root, name, dirInfo); err != nil {
		return fmt.Errorf("own %s: %w", name, err)
	}
	return nil
}

// removeStale deletes the managed files the archive does not carry. A directory
// is left in place: it is emptied by removing what is in it.
func removeStale(root *os.Root, kept []string) error {
	live, err := presentState(root)
	if err != nil {
		return err
	}
	for _, name := range live {
		if slices.Contains(kept, name) {
			continue
		}
		if info, statErr := root.Lstat(name); statErr != nil || info.IsDir() {
			continue
		}
		if removeErr := root.Remove(name); removeErr != nil {
			return fmt.Errorf("remove %s: %w", name, removeErr)
		}
	}
	return nil
}

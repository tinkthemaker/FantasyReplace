package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// safeWriteFile writes through a temporary file in the destination directory.
// On platforms that cannot rename over an existing file, it keeps a rollback
// copy until the replacement is in place.
func safeWriteFile(path string, data []byte, defaultMode os.FileMode) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	mode := defaultMode
	info, statErr := os.Stat(path)
	if statErr == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(statErr) {
		return statErr
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err == nil {
		return nil
	}
	if os.IsNotExist(statErr) {
		return fmt.Errorf("replace %s: destination appeared while writing", path)
	}

	backupFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".backup-*")
	if err != nil {
		return fmt.Errorf("prepare rollback file: %w", err)
	}
	backup := backupFile.Name()
	if err := backupFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(backup); err != nil {
		return err
	}
	if err := os.Rename(path, backup); err != nil {
		return fmt.Errorf("prepare replacement: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Rename(backup, path)
		return fmt.Errorf("replace file: %w", err)
	}
	_ = os.Remove(backup)
	return nil
}

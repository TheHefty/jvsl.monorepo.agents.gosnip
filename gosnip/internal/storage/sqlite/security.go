package sqlite

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func prepareDatabasePath(path string) (bool, error) {
	dir := filepath.Dir(path)
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return false, fmt.Errorf("create database directory %s: %w", dir, err)
		}
		info, err = os.Lstat(dir)
	}
	if err != nil {
		return false, fmt.Errorf("inspect database directory %s: %w", dir, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, &UnsafePathError{Path: dir, Reason: "symbolic links are not allowed"}
	}
	if !info.IsDir() {
		return false, &UnsafePathError{Path: dir, Reason: "parent is not a directory"}
	}
	if err := requirePrivateMode(dir, info); err != nil {
		return false, err
	}

	info, err = os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		file, createErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
		if createErr != nil {
			return false, fmt.Errorf("create database %s: %w", path, createErr)
		}
		if closeErr := file.Close(); closeErr != nil {
			return false, fmt.Errorf("create database %s: close file: %w", path, closeErr)
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect database %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return false, &UnsafePathError{Path: path, Reason: "symbolic links are not allowed"}
	}
	if !info.Mode().IsRegular() {
		return false, &UnsafePathError{Path: path, Reason: "database is not a regular file"}
	}
	if err := requirePrivateMode(path, info); err != nil {
		return false, err
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := checkExistingSQLiteFile(path + suffix); err != nil {
			return false, err
		}
	}
	return false, nil
}

func checkExistingSQLiteFile(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect SQLite sidecar %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return &UnsafePathError{Path: path, Reason: "SQLite sidecar must be a regular file, not a symbolic link"}
	}
	return requirePrivateMode(path, info)
}

func secureSQLiteFiles(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	for _, candidate := range []string{path, path + "-wal", path + "-shm"} {
		info, err := os.Lstat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect SQLite file %s: %w", candidate, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return &UnsafePathError{Path: candidate, Reason: "SQLite file must be a regular file, not a symbolic link"}
		}
		if err := requirePrivateMode(candidate, info); err != nil {
			return err
		}
	}
	return nil
}

func requirePrivateMode(path string, info os.FileInfo) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		return &PermissionError{Path: path, Mode: uint32(mode)}
	}
	return nil
}

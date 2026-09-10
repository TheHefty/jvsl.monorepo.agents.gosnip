package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultPath resolves the platform storage path without creating it.
func DefaultPath() (string, error) {
	return defaultPath(runtime.GOOS, os.Getenv, os.UserHomeDir)
}

func defaultPath(goos string, getenv func(string) string, userHomeDir func() (string, error)) (string, error) {
	switch goos {
	case "linux":
		if dataHome := getenv("XDG_DATA_HOME"); dataHome != "" {
			return filepath.Join(dataHome, "gosnip", "gosnip.db"), nil
		}
		home, err := requiredHome(userHomeDir)
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "gosnip", "gosnip.db"), nil
	case "darwin":
		home, err := requiredHome(userHomeDir)
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "gosnip", "gosnip.db"), nil
	case "windows":
		localAppData := getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("resolve database path: LOCALAPPDATA is not set")
		}
		return filepath.Join(localAppData, "gosnip", "gosnip.db"), nil
	default:
		return "", fmt.Errorf("resolve database path: unsupported operating system %q", goos)
	}
}

func requiredHome(userHomeDir func() (string, error)) (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve database path: find home directory: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("resolve database path: home directory is empty")
	}
	return home, nil
}

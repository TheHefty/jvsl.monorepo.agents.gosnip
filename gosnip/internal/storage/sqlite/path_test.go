package sqlite

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		goos string
		env  map[string]string
		home string
		want string
	}{
		{"Linux XDG", "linux", map[string]string{"XDG_DATA_HOME": "/data"}, "/home/me", filepath.Join("/data", "gosnip", "gosnip.db")},
		{"Linux home", "linux", nil, "/home/me", filepath.Join("/home/me", ".local", "share", "gosnip", "gosnip.db")},
		{"macOS", "darwin", nil, "/Users/me", filepath.Join("/Users/me", "Library", "Application Support", "gosnip", "gosnip.db")},
		{"Windows", "windows", map[string]string{"LOCALAPPDATA": `C:\Users\me\AppData\Local`}, "", filepath.Join(`C:\Users\me\AppData\Local`, "gosnip", "gosnip.db")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := defaultPath(test.goos, func(key string) string { return test.env[key] }, func() (string, error) { return test.home, nil })
			if err != nil {
				t.Fatalf("defaultPath() error = %v", err)
			}
			if got != test.want {
				t.Errorf("defaultPath() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDefaultPathReportsMissingPlatformInputs(t *testing.T) {
	t.Parallel()

	if _, err := defaultPath("windows", func(string) string { return "" }, func() (string, error) { return "", nil }); err == nil {
		t.Error("defaultPath(windows) succeeded without LOCALAPPDATA")
	}
	homeErr := errors.New("no home")
	if _, err := defaultPath("darwin", func(string) string { return "" }, func() (string, error) { return "", homeErr }); !errors.Is(err, homeErr) {
		t.Errorf("defaultPath(darwin) error = %v, want wrapped home error", err)
	}
}

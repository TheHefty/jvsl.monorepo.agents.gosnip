package projecttest

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPrePushHook(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("bash.exe"); err != nil {
			t.Skip("bash.exe is not available")
		}
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test source")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	command := exec.Command("bash", filepath.Join(repoRoot, ".githooks", "pre-push.test.sh"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("pre-push hook test failed: %v\n%s", err, output)
	}
}

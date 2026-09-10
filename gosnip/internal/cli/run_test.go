package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunRoutesRootAndUsageResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantStatus int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "no arguments prints help",
			wantStatus: 0,
			wantStdout: "Usage: gosnip <command>\n\nCommands:\n  version  Print the version\n",
		},
		{
			name:       "unknown command is a usage failure",
			args:       []string{"unknown"},
			wantStatus: 2,
			wantStderr: "gosnip: unknown command \"unknown\"\nRun 'gosnip help' for usage.\n",
		},
		{
			name:       "help rejects extra arguments",
			args:       []string{"help", "extra"},
			wantStatus: 2,
			wantStderr: "gosnip: help accepts no arguments\nRun 'gosnip help' for usage.\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			status := Run(context.Background(), test.args, strings.NewReader("ignored"), &stdout, &stderr)

			if status != test.wantStatus {
				t.Errorf("Run() status = %d, want %d", status, test.wantStatus)
			}
			if got := stdout.String(); got != test.wantStdout {
				t.Errorf("Run() stdout = %q, want %q", got, test.wantStdout)
			}
			if got := stderr.String(); got != test.wantStderr {
				t.Errorf("Run() stderr = %q, want %q", got, test.wantStderr)
			}
		})
	}
}

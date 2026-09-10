package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/TheHefty/jvsl.monorepo.agents.gosnip/gosnip/internal/i18n"
)

// Run executes gosnip without taking ownership of process-global streams or exit handling.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_ = ctx
	_ = stdin

	if len(args) == 0 {
		_, _ = io.WriteString(stdout, i18n.English(i18n.Help))
		return 0
	}

	switch args[0] {
	case "help":
		if len(args) != 1 {
			return usageError(stderr, i18n.HelpAcceptsNoArgs)
		}
		_, _ = io.WriteString(stdout, i18n.English(i18n.Help))
		return 0
	case "version":
		if len(args) != 1 {
			return usageError(stderr, i18n.VersionAcceptsNoArgs)
		}
		_, _ = io.WriteString(stdout, i18n.English(i18n.DevelopmentVersion))
		return 0
	default:
		_, _ = fmt.Fprint(stderr, i18n.English(i18n.UnknownCommand, args[0]))
		_, _ = io.WriteString(stderr, i18n.English(i18n.UsageHint))
		return 2
	}
}

func usageError(stderr io.Writer, key i18n.Key) int {
	_, _ = io.WriteString(stderr, i18n.English(key))
	_, _ = io.WriteString(stderr, i18n.English(i18n.UsageHint))
	return 2
}

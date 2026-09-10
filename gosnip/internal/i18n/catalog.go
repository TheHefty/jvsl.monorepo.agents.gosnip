package i18n

import "fmt"

// Key identifies a human-facing message in an embedded catalogue.
type Key string

const (
	Help                 Key = "help"
	UnknownCommand       Key = "error.unknown_command"
	HelpAcceptsNoArgs    Key = "error.help_accepts_no_arguments"
	VersionAcceptsNoArgs Key = "error.version_accepts_no_arguments"
	UsageHint            Key = "error.usage_hint"
	DevelopmentVersion   Key = "version.development"
)

var keys = []Key{Help, UnknownCommand, HelpAcceptsNoArgs, VersionAcceptsNoArgs, UsageHint, DevelopmentVersion}

var english = map[Key]string{
	Help:                 "Usage: gosnip <command>\n\nCommands:\n  version  Print the version\n",
	UnknownCommand:       "gosnip: unknown command %q\n",
	HelpAcceptsNoArgs:    "gosnip: help accepts no arguments\n",
	VersionAcceptsNoArgs: "gosnip: version accepts no arguments\n",
	UsageHint:            "Run 'gosnip help' for usage.\n",
	DevelopmentVersion:   "dev\n",
}

// AllKeys returns a copy of every message key the application declares.
func AllKeys() []Key {
	return append([]Key(nil), keys...)
}

// English formats a message from the embedded English catalogue.
func English(key Key, args ...any) string {
	return fmt.Sprintf(english[key], args...)
}

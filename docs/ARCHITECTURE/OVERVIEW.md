# Architecture overview

What the `gosnip` system is, as implemented today.

## Current system

The product is an isolated Go module under `gosnip/` within the repository. Its executable entry
point in `cmd/gosnip` owns process-global arguments, streams, and exit handling, then delegates to
`internal/cli.Run`. That runner accepts its context, arguments, stdin, stdout, and stderr, so command
behavior can be tested without spawning or terminating a process.

Human-facing messages come from the embedded English catalogue in `internal/i18n`. The implemented
foundation handles root help, explicit help, development version output, and usage failures. It
does not open storage itself.

`internal/snippet` owns validation and canonicalization before persistence. `internal/storage/sqlite`
owns platform-default paths, private filesystem boundaries, schema migrations, online backups,
writer coordination, and atomic creation. Its current schema is v2; a populated v1 database is the
supported migration source. The command runner is not connected to the store yet, so no snippet
command is exposed by the executable.

The repository root owns cross-module infrastructure. `scripts/test-unit.sh` defines the fast test
suite used by native pull-request CI and the optional versioned pre-push hook. The hook also refuses
direct updates to `master`, while protected-branch settings and CI remain authoritative.

The accepted constraints are recorded in [`../SRS.md`](../SRS.md). Each update here describes what
exists at that commit, not what a later story intends to build.

## Development environment boundary

The `.code-server/` git submodule is tooling, not part of the `gosnip` product. It owns the
container image, code-server, agent sandbox, nested rootless Docker daemon, native launcher, and
selectable Go stack. Its authoritative design lives in
[`.code-server/docs/overview/`](../../.code-server/docs/overview/) and changes only through a tagged
submodule bump.

Project code, SQLite data behavior, command interfaces, release workflows, and their tests belong
to this repository. The generated `.code-server/Dockerfile` is outside that product architecture
and must not be edited manually.

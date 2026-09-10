# Architecture overview

What the `gosnip` system is, as implemented today.

## Current system

There is no product implementation yet. The repository contains an accepted charter and SRS, the
planning structure, release scaffolding inherited from the repository template, and a vendored
development environment. No Go module, executable, database schema, product workflow, or test suite
exists, so presenting a component architecture here would turn an unreviewed design into apparent
fact.

The accepted constraints are recorded in [`../SRS.md`](../SRS.md). Architecture will be added here
as accepted tasks create real boundaries and code. Each update must describe what exists at that
commit, not what a later story intends to build.

## Development environment boundary

The `.code-server/` git submodule is tooling, not part of the `gosnip` product. It owns the
container image, code-server, agent sandbox, nested rootless Docker daemon, native launcher, and
selectable Go stack. Its authoritative design lives in
[`.code-server/docs/overview/`](../../.code-server/docs/overview/) and changes only through a tagged
submodule bump.

Project code, SQLite data behavior, command interfaces, release workflows, and their tests belong
to this repository. The generated `.code-server/Dockerfile` is outside that product architecture
and must not be edited manually.

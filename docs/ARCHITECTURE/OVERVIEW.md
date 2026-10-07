# Architecture overview

What the `gosnip` system is, as implemented today.

## Current system

There is no product implementation yet. The repository contains an accepted charter and SRS, the
release scaffolding inherited from the repository template, and the manifest and generated
configuration for its development environment. No Go module, executable, database schema, product workflow, or test suite
exists, so presenting a component architecture here would turn an unreviewed design into apparent
fact.

The accepted constraints are recorded in [`../SRS.md`](../SRS.md). Architecture will be added here
as accepted tasks create real boundaries and code. Each update must describe what exists at that
commit, not what a later story intends to build.

## Development environment boundary

The Agent Container VS Code extension is tooling, not part of the `gosnip` product. It owns the
container image, agent sandbox, nested rootless Docker daemon, and selectable Go stack, and its
authoritative design lives in the extension's repository. This repository carries only its inputs
and outputs: `.agent-container.stack.json` and the generated `.devcontainer/devcontainer.json`.

Project code, SQLite data behavior, command interfaces, release workflows, and their tests belong
to this repository. The generated `.devcontainer/devcontainer.json` is outside that product
architecture and must not be edited manually.

# gosnip

`gosnip` is a personal, offline command-line application for storing reusable code snippets,
finding them quickly, and copying their exact contents to the native clipboard. It is a small Go
project built for hands-on learning without depending on a hosted gist service.

## Status

The project charter, software requirements, first story, and CLI-foundation task are accepted. The
Go executable and fast native test foundation are implemented; snippet persistence and operations
remain behind their task-design gates.

No `gosnip` product release has been published. The release baseline starts at 0.0.0 so the first
accepted product feature can produce the planned 0.1.0 release.

## Planned MVP

The accepted MVP will provide:

- creation, editing, deliberate deletion, listing, filtering, smart-case search, and inspection of
  snippets;
- exact transfer of snippet code to the native Linux, macOS, and Windows clipboard;
- local SQLite persistence with transactional migrations and recovery backups;
- human-readable terminal output and stable JSON contracts for automation;
- release archives for Linux, macOS, and Windows on amd64 and arm64.

The MVP deliberately excludes hosting, accounts, synchronization, collaboration, a graphical
interface, import/export, encryption, syntax highlighting, shell completion, telemetry, and secure
erasure.

## Development environment

Clone with the submodule initialized:

```bash
git clone --recurse-submodules https://github.com/TheHefty/jvsl.monorepo.agents.gosnip.git
cd jvsl.monorepo.agents.gosnip
```

Prepare the host once, build the selected Go development image, and open the environment:

```bash
.code-server/init
.code-server/setup
.code-server/dev
```

Enable the repository's local Git hooks once per clone:

```bash
git config core.hooksPath .githooks
```

The pre-push hook runs the fast unit suite and refuses direct pushes to `master`. It is a local
guard that can be bypassed with `--no-verify`; GitHub branch protection and CI remain authoritative.

The environment implementation and prerequisites are documented in
[`.code-server/README.md`](.code-server/README.md). The project-specific stack selection lives in
`.code-server.stack.json`; `.code-server/Dockerfile` is generated and must not be edited manually.

## Project documents

- [`docs/CHARTER.md`](docs/CHARTER.md) records why the project exists and its scope boundaries.
- [`docs/SRS.md`](docs/SRS.md) is the accepted source of truth for behavior, constraints, and the
  epic/story decomposition.
- [`docs/OVERVIEW.md`](docs/OVERVIEW.md) summarizes the intended product and current state.
- [`docs/ARCHITECTURE/OVERVIEW.md`](docs/ARCHITECTURE/OVERVIEW.md) records the architecture that
  exists today.
- [`docs/PLANNING/`](docs/PLANNING/) will contain accepted stories, scenarios, and tasks.

## License

[MIT](LICENSE) © 2026 João Lima.

# Overview

`gosnip` will give its owner a dependable local workflow for reusable code snippets. The owner
will be able to capture code with a unique name, language, and tags; find or inspect it later; edit
or deliberately delete it; and copy its exact bytes to the desktop clipboard.

## Product shape

The accepted product is a non-interactive Go CLI backed by a local SQLite database. Normal use is
offline and sends no telemetry. Linux, macOS, and Windows are supported on amd64 and arm64. Human
terminal output serves direct use, while stable JSON values, stderr diagnostics, and exit statuses
serve scripts.

Snippet bodies remain UTF-8 text exactly as entered. Metadata is normalized so names stay
unambiguous and language/tag filters remain predictable. The database owns schema versioning,
transactional changes, bounded lock waits, and a durable backup before migration.

The exact command behavior, validation, data map, performance target, and release requirements are
normative in [`SRS.md`](SRS.md). [`CHARTER.md`](CHARTER.md) governs purpose and scope when the two
could otherwise be confused.

## Current state

Initialization is complete: the charter and SRS have been accepted. No product code, test suite,
story, or task exists yet. The next change is the first story and its Gherkin acceptance scenarios;
implementation remains prohibited until that story and a task design pass their respective gates.

Release files inherited from the repository template still contain template history and version
1.8.0. They are not `gosnip` product releases. The accepted distribution story will replace that
baseline and produce the first product release as 0.1.0.

## Development environment

The development environment is vendored as the `.code-server/` submodule. It provides code-server,
the agent tooling, a sandbox, nested rootless Docker, and the selected Go stack. Its own design is
documented in [`.code-server/docs/overview/`](../.code-server/docs/overview/) and versions with the
template rather than this product.

The project manifest is `.code-server.stack.json`. Use `.code-server/setup` to regenerate the
development image and `.code-server/dev` to open it; never edit the generated
`.code-server/Dockerfile`.

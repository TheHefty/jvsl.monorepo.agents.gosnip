# Software Requirements Specification: gosnip

| | |
|---|---|
| **Status** | Accepted |
| **Date** | 2026-09-10 |
| **Author** | João Lima |

## Functional requirements

### Command-line interface

- **FR-001:** The system shall expose the commands `add`, `list`, `search`, `show`, `copy`, `edit`,
  `delete`, and `version` under the `gosnip` executable.
- **FR-002:** Commands that address one snippet shall accept a positional target. A target made
  only of decimal digits shall resolve as an ID; every other target shall resolve as a unique name.
- **FR-003:** The system shall reject snippet names that are empty after trimming, contain only
  decimal digits, exceed 100 Unicode characters, or case-fold to the name of an existing snippet.
  It shall preserve the accepted name's original spelling for display.
- **FR-004:** A global `--db PATH` option shall select the database for one invocation. If it is
  absent, `GOSNIP_DB` shall apply. If neither is set, the system shall use the platform default:
  `$XDG_DATA_HOME/gosnip/gosnip.db` on Linux when `XDG_DATA_HOME` is set,
  `~/.local/share/gosnip/gosnip.db` otherwise, `~/Library/Application Support/gosnip/gosnip.db` on
  macOS, and `%LOCALAPPDATA%\gosnip\gosnip.db` on Windows.
- **FR-005:** Human-readable output shall go to stdout, diagnostics shall go to stderr, and a
  successful command shall return exit status 0. Invalid command syntax, missing required
  arguments, and incompatible flags shall return 2. Validation, conflict, not-found, database,
  editor, clipboard, and other operational failures shall return 1.

### Creating snippets

- **FR-006:** `gosnip add NAME --language LANGUAGE [--tag TAG ...]` shall read the complete snippet
  body from stdin and create one snippet after all input passes validation.
- **FR-007:** `add` shall reject absent or empty stdin, invalid UTF-8, a body larger than 1 MiB, an
  invalid name, an invalid language, more than 20 tags, an invalid tag, or a duplicate name without
  writing a partial record.
- **FR-008:** On success, `add` shall print a human-readable summary containing the new ID, name,
  language, tags, and update time. With `--json`, it shall emit the complete snippet object.

### Listing, searching, and showing snippets

- **FR-009:** `gosnip list [--language LANGUAGE] [--tag TAG ...]` shall list all matching snippets.
  Language and every supplied tag shall be combined with logical AND.
- **FR-010:** `gosnip search QUERY [--language LANGUAGE] [--tag TAG ...]` shall require a non-empty
  query and search both name and code. Language, every tag, and the text match shall be combined
  with logical AND.
- **FR-011:** Search shall use Unicode smart-case substring matching: a query containing no
  uppercase Unicode letter shall compare without case; a query containing any uppercase Unicode
  letter shall compare with case.
- **FR-012:** Human `list` and `search` output shall be a summary table containing ID, name,
  language, tags, and update time without snippet bodies. `--json` shall emit an array of complete
  snippet objects.
- **FR-013:** `list` and `search` shall order results by `updated_at` descending and then ID
  descending. They shall return an empty result successfully when nothing matches and shall not
  paginate in the MVP.
- **FR-014:** `gosnip show TARGET` shall print only the stored code to stdout, adding or removing no
  bytes. With `--json`, it shall emit the complete snippet object.

### Copying snippets

- **FR-015:** `gosnip copy TARGET` shall place the stored code, byte for byte, in the native system
  clipboard on supported Linux, macOS, and Windows desktop sessions.
- **FR-016:** A successful human `copy` shall be silent. With `--json`, it shall emit an object with
  the snippet's `id`, `name`, and `copied: true`, without the code.
- **FR-017:** If no usable desktop clipboard is available, `copy` shall fail with an actionable
  diagnostic and shall not print the code as a fallback.

### Editing snippets

- **FR-018:** `gosnip edit TARGET [--name NAME] [--language LANGUAGE] [--tag TAG ...]
  [--clear-tags] [--editor COMMAND]` shall open a temporary file containing only the exact stored
  code and wait for the editor to exit.
- **FR-019:** Editor selection shall use `--editor`, then `VISUAL`, then `EDITOR`. If none identifies
  an executable editor, the command shall fail with instructions for configuring one rather than
  guessing a platform application.
- **FR-020:** Name and language flags shall replace their respective values. One or more `--tag`
  flags shall replace the complete tag set. No tag flag shall preserve it. `--clear-tags` shall
  remove all tags and shall be incompatible with `--tag`.
- **FR-021:** Editor failure, invalid edited content, invalid metadata, or a uniqueness conflict
  shall leave the entire snippet unchanged. An unchanged temporary file with unchanged metadata
  shall succeed without changing `updated_at`. A valid change shall commit code and metadata
  atomically.
- **FR-022:** On a valid change, `edit` shall print the same human summary shape as `add`; with
  `--json`, it shall emit the complete updated snippet object.

### Deleting snippets

- **FR-023:** `gosnip delete TARGET` shall refuse to delete unless `--yes` is supplied and shall
  explain the required authorization without opening an interactive prompt.
- **FR-024:** `delete --yes` shall atomically remove the selected row. Human output shall identify
  the deleted ID and name; `--json` shall emit the complete object as it existed before deletion.
- **FR-025:** Deletion shall make the record unavailable to subsequent commands but shall not
  promise secure erasure from SQLite pages, WAL files, filesystem storage, or migration backups.

### Structured output and versioning

- **FR-026:** The global `--json` option shall replace successful human output with exactly one
  JSON value. A snippet object shall contain `id`, `name`, `code`, `language`, `tags`, `created_at`,
  and `updated_at`; tags shall be an array and timestamps shall be UTC RFC 3339 strings.
- **FR-027:** When `--json` is active, failures shall emit
  `{"error":{"code":"CODE","message":"MESSAGE"}}` to stderr and no success value to stdout.
  Stable codes shall cover `usage`, `validation`, `conflict`, `not_found`, `database`, `editor`,
  `clipboard`, and `internal`.
- **FR-028:** `gosnip version` shall report the version embedded at build time. With `--json`, it
  shall emit `{"version":"VERSION"}`.

### Persistence and schema lifecycle

- **FR-029:** The first data-accessing command shall create missing parent directories and a new
  SQLite database. Commands such as help and version that do not need data shall work without
  opening or creating the database.
- **FR-030:** The logical snippet record shall contain an auto-incrementing integer ID, display
  name, Unicode case-folded unique name key, code, language, delimited tags, immutable creation
  time, and update time. A schema-version record shall track applied migrations.
- **FR-031:** Code shall be valid UTF-8 and shall retain its exact bytes, including whitespace and
  line endings. Language shall be required, trimmed, lowercased, at most 100 characters, and use
  Unicode letters or numbers plus `+`, `#`, `.`, `_`, and `-`.
- **FR-032:** Each tag shall be trimmed, lowercased, at most 100 Unicode characters, and consist of
  Unicode letters or numbers plus `_` and `-`. The system shall reject empty or comma-containing
  tags, remove duplicates, sort tags lexically, and store the canonical set with comma delimiters.
- **FR-033:** Creation shall set `created_at` and `updated_at` to the same UTC instant. A successful
  content or metadata edit shall change only `updated_at`. Reads and clipboard operations shall
  change neither timestamp. Human timestamps shall respect the user's locale and timezone.
- **FR-034:** Known schema upgrades shall run automatically in a transaction. Before changing an
  existing schema, the system shall create a uniquely timestamped backup beside the database and
  report its path. It shall never delete migration backups automatically.
- **FR-035:** A failed migration, corrupt database, or schema newer than the running binary
  supports shall leave the original data in place and fail with a diagnostic that identifies the
  database and recovery action.
- **FR-036:** The database shall use SQLite WAL mode and wait up to five seconds for a competing
  writer. It shall then fail clearly rather than wait forever, corrupt data, or apply a partial
  mutation.

### Distribution

- **FR-037:** Version 0.1.0 shall be the first `gosnip` product release; inherited template version
  and release history shall not be presented as product history.
- **FR-038:** Each product release shall provide binaries for Linux, macOS, and Windows on amd64 and
  arm64. Linux and macOS artifacts shall be `.tar.gz`; Windows artifacts shall be `.zip`.
- **FR-039:** Each GitHub Release shall include a SHA-256 checksum manifest covering every archive.
  Artifact creation shall run automatically after release-please creates the product tag.

## Non-functional requirements

- **NFR-001:** Normal command execution shall be fully offline and shall perform no telemetry,
  update check, or network request.
- **NFR-002:** The product test suite shall run natively on Linux, macOS, and Windows. Release CI
  shall build all six OS/architecture targets and smoke-test artifacts wherever a native runner is
  available.
- **NFR-003:** On each native CI platform, after one unmeasured warm-up, the median of five human
  `list` or `search` runs over 1,000 snippets whose bodies are no larger than 10 KiB shall complete
  in at most one second, including process startup and output formatting. Full-body JSON,
  clipboard, and editor time are excluded.
- **NFR-004:** All human-facing strings shall come from an embedded message catalogue. The MVP
  shall ship an English catalogue only; JSON field names and error codes shall never be translated.
- **NFR-005:** Human output shall use no color and shall communicate every state in text. Commands
  shall require no interactive prompt, making their behavior usable through keyboards, pipes, and
  screen-reader-compatible terminals.
- **NFR-006:** Mutations and migrations shall be transactional. Failures shall name the failed
  operation, its cause, and an actionable next step without printing snippet code or secrets.
- **NFR-007:** Newly created data directories and database or backup files shall be accessible only
  to the current user where the operating system exposes applicable permission controls.
- **NFR-008:** The application shall keep its in-memory work bounded by validated input sizes and
  the expected 1,000-record collection rather than loading unbounded external input.

## Data and legal

`gosnip` processes code snippets and their names, language labels, tags, IDs, and timestamps for
the sole purpose of local retrieval and reuse by the owner. The charter excludes credentials,
sensitive personal data, and data belonging to anyone else. The application has no account,
telemetry, network service, or third-party disclosure.

The owner is the sole controller of the data. Records remain in the selected SQLite database until
the owner deletes them or removes the database. Migration backups remain until the owner removes
them manually. A normal delete is not secure erasure. Because this is personal, non-economic use
of the owner's own nonsensitive data, the project adopts no additional privacy policy at this
stage. This section records the project's decision and is not legal advice.

## Epics and stories

### Epic: dependable-local-snippet-workflow

Make a reusable snippet collection available through a dependable local command-line workflow.

- **capture-snippets:** The owner can create a validated snippet in a durable local collection.
- **maintain-snippets:** The owner can edit or deliberately delete an existing snippet without
  partial or accidental changes.
- **discover-snippets:** The owner can list, filter, search, and inspect stored snippets
  predictably.
- **copy-snippets:** The owner can place an exact stored snippet in the native clipboard.
- **automate-cli-workflows:** The owner can compose every operation in scripts through stable
  streams, JSON values, database selection, and exit statuses.
- **distribute-gosnip:** The owner can install a verified release artifact on every supported OS
  and architecture.

## Alternatives considered

- **Do nothing:** rejected because hosted gist tools do not provide the intended local, offline
  retrieval workflow or the hands-on Go learning outcome.
- **Hosted service:** rejected because accounts, networking, synchronization, privacy obligations,
  and operations contradict the personal weekend scope.
- **JSON or loose text files:** rejected because concurrent consistency, schema evolution, and
  future portability are safer with a transactional data model.
- **A CGO SQLite driver:** rejected because a pure-Go driver keeps six-target builds and local
  installation simpler; `modernc.org/sqlite` is the selected driver.
- **Normalized tag tables:** rejected because a collection of tens to hundreds of snippets does
  not justify the additional schema and joins; canonical delimiter storage meets the required
  filters.
- **Interactive prompts:** rejected because deterministic arguments, stdin, stdout, stderr, and
  exit codes compose better in scripts. The editor remains the explicit exception for code edits.
- **Partial platform support:** rejected because the owner chose first-release binaries for Linux,
  macOS, and Windows on both amd64 and arm64.

## Outcome

Accepted by the owner on 2026-09-10 after the SRS grilling. The interview expanded the original
weekend concept with full lifecycle operations, Unicode smart-case search, explicit automation
contracts, transactional migration backups, six release targets, and native multiplatform
verification while keeping hosting, synchronization, and adjacent convenience features outside
the MVP.

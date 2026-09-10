---
status: Draft
story: dependable-local-snippet-workflow/capture-snippets
epic: dependable-local-snippet-workflow
pr:
depends-on:
  - establish-cli-foundation
---

# Task: persist-snippets

## Summary

Add the validated snippet domain and durable SQLite store inside the isolated `gosnip` module. The
store creates private platform-default storage, preserves exact code bytes, coordinates concurrent
writers, and upgrades a supported v1 database to schema v2 only after making a consistent backup.
This makes persistence independently trustworthy before the final task connects it to `gosnip add`.

## Problem

The implemented CLI foundation has no snippet model or storage. The accepted story requires the
first data command to create a database without preparation, while invalid input, unsafe files,
corruption, unsupported schemas, migration failures, and competing writers must not damage existing
data or create partial records.

Implementing those guarantees inside the command handler would couple validation, platform paths,
filesystem security, migration recovery, and SQL transactions. The owner would pay for that
coupling whenever another command needs the same invariants or a database failure must be diagnosed.

## Proposal

Create `internal/snippet` as the domain boundary. A `Draft` carries raw name, code, language, and
tags; `Normalize(Draft)` returns a `NewSnippet` whose invalid states cannot be constructed through
the normal API. It will:

- require non-empty valid UTF-8 code no larger than 1 MiB while preserving every accepted byte;
- trim Unicode whitespace outside names, preserve their remaining spelling, reject empty,
  decimal-only, or over-100-character names, and derive a Unicode case-folded key;
- require a trimmed, Unicode-lowercased language no longer than 100 characters using only the SRS
  character set; and
- trim and lowercase tags, reject invalid or excessive values, deduplicate them, sort them
  lexically, and serialize their canonical form with comma delimiters.

Use `golang.org/x/text` v0.42.0 for Unicode case folding and `modernc.org/sqlite` v1.58.0 through
`database/sql`. Both are compatible with the selected Go 1.26.5 module and carry permissive BSD
licenses; add the notices their binary-redistribution terms require.

Create `internal/storage/sqlite` with these internal interfaces:

```go
func Open(ctx context.Context, path string, options Options) (*Store, OpenReport, error)
func (s *Store) Create(ctx context.Context, input snippet.NewSnippet) (snippet.Snippet, error)
func (s *Store) Close() error
```

`Options` provides a clock and narrow backup/migration failure seams for deterministic tests.
`OpenReport` identifies whether storage was created or migrated, the old and current schema
versions, and the optional backup path. Typed errors distinguish validation, conflict, unsafe
permissions, corruption, future schema, lock timeout, and other storage failures without including
snippet code.

A platform-path helper will resolve the defaults fixed by FR-004. `Open` still accepts an explicit
path so tests are isolated and the future `--db` feature can reuse the store without changing it.
Missing `HOME`, `XDG_DATA_HOME`, or `LOCALAPPDATA` inputs fail with an actionable cause rather than
silently choosing a working directory.

On Linux and macOS, create the gosnip directory as `0700` and database or backup files as `0600`.
Before opening existing storage, reject symlinks and any group/other permission bits on the gosnip
directory, database, backup target, or existing SQLite sidecars. Do not require private mode bits on
ancestor directories such as `~/.local/share`. Windows uses the owner's `LOCALAPPDATA`; interpreting
Windows ACLs is outside this task because POSIX mode bits do not describe them reliably.

Open SQLite with foreign keys, WAL journal mode, a five-second busy timeout, and one pooled
connection per process. A context deadline earlier than five seconds wins. Before a migration,
acquire an immediate write lock, re-read the schema version, create a consistent online backup, and
apply every pending migration in the same critical section. Concurrent openers therefore cannot
back up one schema while another process changes it.

Embed ordered SQL migrations. Schema v1 creates `schema_migrations` and `snippets` with integer ID,
display name, folded name key, exact UTF-8 code, canonical language and tags, and UTC timestamps.
Schema v2 adds the unique index on folded name key. Persist timestamps as normalized UTC
RFC3339Nano `TEXT`; the domain exposes `time.Time`. A new database applies all migrations without a
backup. An existing v1 fixture exercises the real upgrade. Unversioned or future schemas are not
guessed.

Create the backup destination exclusively at `0600`, then use the driver's online backup API so a
WAL database is copied consistently. A nanosecond UTC timestamp plus exclusive-create collision
handling makes its sibling filename unique. Keep a completed backup after either migration success
or rollback; remove only an incomplete destination. Migration changes and snippet creation are
transactional.

The final `integrate-add-command` task will validate stdin before opening the store, render typed
errors through the message catalogue, report any migration backup, and register the story's Gherkin
file with Gocuke. This task does not change CLI syntax or claim that the feature file is executable.

## Three worst failure scenarios

| # | Scenario | How it manifests | Test that catches it |
|---|---|---|---|
| 1 | A failed migration changes the source or produces a useless backup. | An owner cannot reopen the old database after an injected migration failure, or the reported backup omits committed WAL data. | Build a real v1 WAL fixture with existing snippets, force the v2 migration to fail after backup, and assert rollback, unchanged source records and main database bytes, a readable complete backup, and no partial schema version. |
| 2 | Path handling follows a symlink or trusts unsafe permissions. | Gosnip reads or modifies a redirected database, or exposes snippets and backups to another local user. | On POSIX runners, exercise real directories, files, sidecars, and symlinks with unsafe modes and assert refusal before any byte or permission changes; separately assert new artifacts use `0700`/`0600`. |
| 3 | Concurrent writers bypass uniqueness or wait indefinitely. | Two equivalent names are committed, a mutation is partial, or a command hangs behind a lock. | Hold a real immediate transaction while another connection writes and assert timeout by five seconds; race case-fold-equivalent creates and assert exactly one complete row and one typed conflict. |

These failure tests are written and observed failing before their corresponding implementation.

## Blast radius

- [ ] The template submodule — needs a release and a pointer bump before any project sees it
- [ ] The image — needs `.code-server/setup`; nothing changes in a running environment until then
- [ ] The stack manifest (`.code-server.stack.json`)
- [ ] The agent's sandbox map, or where a capability is decided
- [x] A dependency fetched at build time — pin `modernc.org/sqlite` v1.58.0 and
  `golang.org/x/text` v0.42.0 in the product module and commit their module checksums
- [ ] The release/versioning discipline
- [x] Another story or task — `integrate-add-command` depends on the validated model, store,
  platform paths, migration report, and typed errors introduced here
- [ ] Nothing outside this repository

## Alternatives considered

- **Filesystem copying for backups:** rejected because a WAL database spans coordinated state that
  cannot be captured safely by copying only the main file.
- **Backup before acquiring the migration lock:** rejected because another process could change the
  schema between the copy and migration.
- **Start at schema v1:** rejected because the accepted story explicitly requires exercising an
  older populated database before the first release.
- **Unix integer timestamps:** rejected because canonical RFC3339Nano UTC text is inspectable,
  lexically ordered, and already matches the later JSON contract.
- **Store raw drafts and validate only in the CLI:** rejected because later callers could persist
  states the product considers invalid.
- **Inspect Windows ACLs now:** rejected because correct owner and inherited-ACE evaluation requires
  a separate native security design; `LOCALAPPDATA` remains the Windows boundary for this MVP task.
- **Follow symlinks:** rejected because resolved destinations make permission and backup guarantees
  ambiguous.
- **Do nothing:** rejected because the final `add` task would otherwise combine every durability and
  command concern in one untestable slice.

## Verification

This design and the story index will run diff and Markdown-size checks in their own pull request.
Implementation verification will use temporary real SQLite databases on native Linux, macOS, and
Windows, plus the existing unit command, `go vet`, the race detector, and six-target compilation.
POSIX permission assertions will be skipped only on Windows with an explicit reason.

## Open questions

None. The owner selected schema v2 with a migratable v1 fixture, UTC RFC3339Nano text timestamps,
`modernc.org/sqlite` v1.58.0, strict POSIX permissions, lock-before-backup migration coordination,
complete domain validation, trimmed display names, retained completed backups after rollback, and
symlink refusal.

## Outcome

Pending owner review of this task gate.

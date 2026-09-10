---
status: Draft
story: dependable-local-snippet-workflow/capture-snippets
epic: dependable-local-snippet-workflow
pr:
depends-on: []
---

# Task: establish-cli-foundation

## Summary

Establish the Go executable boundary, deterministic command runner, embedded English message
catalogue, native pull-request unit-test matrix, and a versioned pre-push hook. This gives later
storage and `add` work one testable interface and one shared definition of the fast test suite
without creating or opening a database.

## Problem

The repository has an accepted `capture-snippets` story but no product module, executable, test
suite, or product CI. Implementing persistence first would force command parsing, process streams,
exit statuses, and platform behavior to be tested indirectly through global process state. It would
also leave ordinary pull requests mergeable without an enforced product check and give local pushes
a different test path from CI.

The owner pays for that ambiguity while learning Go: failures would arrive as coupled CLI and
database defects instead of small, attributable test failures.

## Proposal

Create a Go module for `gosnip` with a thin entry point under `cmd/gosnip` and application dispatch
under `internal/cli`. The stable internal boundary is:

```go
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int
```

`main` passes process-owned values to `Run` and exits with its result. `Run` uses standard-library
argument handling and manual subcommand dispatch; it does not use Cobra or another CLI framework.
With no arguments it writes help to stdout and returns 0. An unknown command or invalid syntax
writes an actionable diagnostic to stderr and returns 2. Help and version-facing foundation paths
must not inspect or create local storage.

Place human-facing text behind typed message keys in `internal/i18n`, with one embedded English
catalogue. Tests must prove that every declared key has a non-empty English value. JSON field names
and error codes remain outside translation, and JSON itself remains deferred to the
`automate-cli-workflows` story.

Define `scripts/test-unit.sh` as the single fast-test entry point. It runs the Go unit suite in
short mode and propagates its exit status. A pull-request workflow calls it on native Linux, macOS,
and Windows runners. Actions are pinned to immutable commit SHAs, while the Go version follows the
repository's selected Go stack version.

Add `.githooks/pre-push`, which consumes the pushed refs, refuses any update or deletion whose
remote ref is `refs/heads/master`, then executes `scripts/test-unit.sh`. Document the one-time,
per-clone activation command `git config core.hooksPath .githooks` and the fact that `--no-verify`
can bypass it. CI and protected-branch checks remain the authoritative gate.

The later `persist-snippets` task will depend on this boundary. A final `integrate-add-command` task
will add Gocuke v1.1.1 as a test-only dependency, register the story's exact
`capture-snippets.feature`, add a distinct executable E2E suite, and make both run on the
release-please PR across Linux, macOS, and Windows. This task only creates the fast unit-test lane;
it must not claim that the Gherkin scenarios are executable yet.

This task moves the story toward every scenario by creating its command boundary, but it completes
none of the storage-backed acceptance scenarios on its own.

## Three worst failure scenarios

| # | Scenario | How it manifests | Test that catches it |
|---|---|---|---|
| 1 | The runner violates its stream or exit-status contract. | Help, usage failures, or unknown commands write to the wrong stream, return the wrong status, or terminate the test process through global `os.Exit`. | Table-driven tests call `cli.Run` with in-memory streams and assert exact status, stdout, and stderr for no arguments, unknown commands, and invalid syntax. |
| 2 | The pre-push safety gate permits a direct update to `master` or hides a failing unit suite. | A forbidden push reaches the remote, or a developer believes tests passed because the hook returned zero after the shared test command failed. | A shell test drives the real hook in a temporary Git repository with controlled ref input and a controlled unit-test script, asserting rejection of every `master` ref operation and exact propagation of a non-zero test result. |
| 3 | Fast tests work only on the development Linux container or local and CI commands drift apart. | A platform-specific path, shell assumption, or different CI command breaks macOS or Windows after a locally green push. | Every pull request runs the same `scripts/test-unit.sh` used by the hook on native Linux, macOS, and Windows; a script test verifies short-mode invocation and failure propagation. |

The failure tests above are written and observed failing before their corresponding implementation.

## Blast radius

- [ ] The template submodule — needs a release and a pointer bump before any project sees it
- [ ] The image — needs `.code-server/setup`; nothing changes in a running environment until then
- [ ] The stack manifest (`.code-server.stack.json`)
- [ ] The agent's sandbox map, or where a capability is decided
- [ ] A dependency fetched at build time — with its pin and digest
- [x] The release/versioning discipline — the release-please mechanism stays unchanged, but product
  checks begin to provide merge gates that its release PR will later reuse
- [x] Another story or task — `persist-snippets` and `integrate-add-command` depend on the runner and
  fast-test lane established here
- [ ] Nothing outside this repository

The repository owner must configure the workflow checks as required checks for protected
`master`; a committed workflow cannot change repository branch-protection settings by itself.

## Alternatives considered

- **Cobra:** rejected because the current command grammar does not justify a framework dependency;
  standard-library parsing keeps the first slice small and explicit.
- **Testify:** rejected because standard `testing` tables cover the contracts without another
  dependency or assertion dialect.
- **Execute only through subprocesses:** rejected because injected streams and arguments make unit
  failures faster and more precise; executable behavior will still receive separate E2E coverage.
- **Run Gocuke in the fast suite now:** rejected because no accepted behavior is implemented yet and
  registering the feature before the integration task would produce a permanently red ordinary-PR
  gate rather than a test-first implementation step.
- **Install the hook automatically:** rejected because Git intentionally makes hook activation a
  per-clone trust decision; setup documentation will state the explicit command.
- **Do nothing:** rejected because later work would invent its own process boundary and have no
  product CI gate.

## Verification

The task document and story index will be checked with the repository Markdown-size validator and
reviewed as their own pull request. No Go or workflow implementation exists at this gate, so product
tests cannot run yet. The implementation PR will record the initial red runs and the final native CI
matrix results.

## Open questions

None. The owner selected standard-library CLI and tests, native Linux/macOS/Windows CI, an injectable
runner, typed English messages, fast unit tests on every pull request and pre-push, and deferred
Gocuke plus distinct E2E tests on the release-please PR.

## Outcome

Pending owner review of this task gate.

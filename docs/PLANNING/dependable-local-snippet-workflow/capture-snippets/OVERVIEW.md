# Story: Capture snippets

| | |
|---|---|
| **Status** | Accepted |
| **Epic** | dependable-local-snippet-workflow |
| **Date** | 2026-09-10 |

## Summary

The owner can create a validated snippet in a durable local collection on a clean machine without
preparing a database first.

## Why

Every later workflow needs a trustworthy way to put the first exact piece of code into the
collection. This story establishes the human `add` path and the persistence guarantees it depends
on: unambiguous names, bounded and normalized metadata, exact UTF-8 code, a versioned local schema,
atomic creation, and failures that preserve existing data.

It delivers the `add` portion of FR-001, FR-003, FR-005 through FR-008, FR-029 through FR-036, and
the applicable constraints in NFR-001 and NFR-004 through NFR-008.

## Acceptance criteria

The agreed behavior lives in [`capture-snippets.feature`](capture-snippets.feature). It is
documentation, not an executable test, until an accepted task explicitly registers this exact file
with a Gherkin runner and observes it fail for missing behavior.

## Tasks

| Order | Task | Status | Depends on |
|---|---|---|---|
| 1 | [`establish-cli-foundation`](tasks/establish-cli-foundation.md) | Accepted | — |
| 2 | [`persist-snippets`](tasks/persist-snippets.md) | Draft | `establish-cli-foundation` |

The complete `add` integration task will be designed after persistence is accepted. It will register
this story's exact feature file with Gocuke and add the release-PR acceptance and end-to-end gates.

## Out of scope

- Structured output and the public `--db`/`GOSNIP_DB` overrides belong to
  `automate-cli-workflows`.
- Listing, filtering, searching, and showing stored snippets belong to `discover-snippets`.
- Editing and deletion belong to `maintain-snippets`.
- Clipboard integration belongs to `copy-snippets`.
- Release artifacts and installation belong to `distribute-gosnip`.
- This story creates no graphical, hosted, synchronized, multi-user, import/export, encryption,
  syntax-highlighting, shell-completion, telemetry, or secure-erasure behavior.

## Outcome

Accepted by the owner on 2026-09-10 after the story grilling. The agreed scope includes complete
schema lifecycle foundations and failure preservation, while JSON and public database overrides
remain reserved for the automation story.

# CLAUDE.md

Guidance for agents working in this repository.

## Standing answers

- **Work here is Navigator Mode.** The user writes the code; the agent investigates, guides,
  reviews, and verifies. The user chose this mode to learn Go by implementing the project
  personally.
- **The documentation language is English.** Every project file written from here on inherits it,
  including commit messages.
- **This is the `gosnip` project.** It consumes the development-environment template vendored at
  `.code-server/`; it is not the template repository itself.
- **Project initialization is still in progress.** The charter is the first gate. The SRS must be
  accepted before any story is written, a story and its scenarios before its tasks, and a task's
  design before its code.

## If the imports below did not load

The normative documents live inside the `.code-server/` submodule, which is empty until
`git submodule update --init`. Imports that do not resolve fail silently. If the modes, rules, or
initialization procedure are unavailable, stop and report it rather than proceeding without them.

@.code-server/docs/agent/en/MODES.md
@docs/RULES.md
@.code-server/docs/agent/en/INITIALIZATION.md

## What this repository is

`gosnip` is a new, personal Go command-line application for keeping reusable code snippets in a
local SQLite database, finding them quickly, and copying them to the clipboard. The accepted scope
and standing decisions live in `docs/CHARTER.md`. Detailed behavior and architecture remain
undecided until the SRS and subsequent planning gates are accepted.

The environment template is a git submodule at `.code-server/`. Its implementation and design
rationale belong to the template repository; project code, requirements, architecture, and local
rules belong in this repository.

## Current development state

There is no product code or project test suite yet. Do not create either before the charter and SRS
gates have been completed and the first story and task have been accepted.

The selected development stack is recorded in `.code-server.stack.json`. The generated
`.code-server/Dockerfile` must never be hand-edited.

## Environment commands

Prepare the host once:

```bash
.code-server/init
```

Build or rebuild the development image:

```bash
.code-server/setup
```

Open the environment:

```bash
.code-server/dev
```

## Planning workflow

The mandatory chain is:

1. Accepted charter in `docs/CHARTER.md`.
2. Accepted SRS in `docs/SRS.md`.
3. Accepted story overview and Gherkin scenarios under `docs/PLANNING/`.
4. Accepted task design, including its three worst failure scenarios.
5. Product code written test-first from those scenarios.

Each document is its own pull request and is agreed before the next link is written. The full
procedure is `.code-server/docs/agent/en/WORKFLOW.md`.

## Branches and releases

The default branch is protected. Every change goes through a pull request; do not push directly to
`master`, merge a pull request, or cut a release without the user's explicit direction.

Use conventional commits. When a feature PR is merged with a merge commit, its PR title must be
non-conventional so release-please does not count it twice. Never rename a release-please release
PR, because its title carries the version used to create the tag.

When updating `.code-server/`, pin it to a released tag, read the template changelog, and rerun
`.code-server/setup` after the bump.

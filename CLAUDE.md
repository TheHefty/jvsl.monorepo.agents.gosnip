# CLAUDE.md

Guidance for agents working in this repository.

## Standing answers

- **Work here is Navigator Mode.** The user writes the code; the agent investigates, guides,
  reviews, and verifies. The user chose this mode to learn Go by implementing the project
  personally.
- **The documentation language is English.** Every project file written from here on inherits it,
  including commit messages.
- **This is the `gosnip` project.** Its development environment is provided by the Agent Container
  VS Code extension (`jvsl.env.agents.container`); this is not that extension's repository.
- **Project initialization is complete.** The charter and SRS are accepted. A story and its
  scenarios must be accepted before its tasks, and a task's design before its code.

## If the normative documents did not load

The inherited modes and rules ship inside the agent container's image and are loaded from
`/config/.claude/rules/` (`MODES.md` and `RULES.md`). Outside that container they are absent, and
nothing reports the absence. If the modes or rules are unavailable, stop and report it rather than
proceeding without them.

@docs/RULES.md

## What this repository is

`gosnip` is a new, personal Go command-line application for keeping reusable code snippets in a
local SQLite database, finding them quickly, and copying them to the clipboard. The accepted scope
and standing decisions live in `docs/CHARTER.md`. Detailed behavior and architecture remain
undecided until the SRS and subsequent planning gates are accepted.

The development environment is not vendored here. The Agent Container extension composes and builds
the image and generates `.devcontainer/devcontainer.json`; its implementation and design rationale
belong to the extension's repository. Project code, requirements, architecture, and local rules
belong in this repository.

## Current development state

There is no product code or project test suite yet. Do not create either before the relevant story
and task have been accepted.

The selected development stack is recorded in `.agent-container.stack.json`. The generated
`.devcontainer/devcontainer.json` is rewritten by the extension and must never be hand-edited.

## Environment

Open the repository in VS Code on the host with the Agent Container extension installed, and use
its `Agent Container:` commands to build the image and open the container. Nothing needs to be run
from this repository to prepare the host.

## Work tracker

Epics, stories, tasks, and debts live in the `bd` tracker, not in Markdown; the charter and SRS
remain files. The tracker travels by this repository's remote as `refs/dolt/data`: run
`bd dolt push` alongside every approved `git push`, and never on its own.

## Planning workflow

The mandatory chain is:

1. Accepted charter in `docs/CHARTER.md`.
2. Accepted SRS in `docs/SRS.md`.
3. Accepted story, with its overview and Gherkin scenarios, in the `bd` tracker.
4. Accepted task design in the tracker, including its three worst failure scenarios.
5. Product code written test-first from those scenarios.

Each link is agreed before the next is written: the charter and SRS through their own pull requests,
tracker items as proposals that the user accepts.

## Branches and releases

The default branch is protected. Every change goes through a pull request; do not push directly to
`master`, merge a pull request, or cut a release without the user's explicit direction.

Use conventional commits. When a feature PR is merged with a merge commit, its PR title must be
non-conventional so release-please does not count it twice. Never rename a release-please release
PR, because its title carries the version used to create the tag.

Updating the environment means updating the Agent Container extension and rebuilding the image
through it; read the extension's changelog first, because a new version can change the inherited
rules as well as the image.

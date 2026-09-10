# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

## Standing answers

- **The mode is Pair Programming Mode.** The agent drives, the user navigates. It is described in
  the import below, and it governs every session until someone says otherwise.
- **The documentation language is English.** Every file written from here on inherits it, including
  the commit messages.
- **This repository is the template itself, so the rules it ships do not govern it.** They govern
  projects built on it. A change here is a change to what projects will inherit.
- **There is no initialization to run here, and the interview is deliberately not imported.**
  Initialization is a project's first act, and this is not a project — it is the reference monorepo
  that ships the interview to others. While the process documents were copies kept at this root,
  importing `INITIALIZATION.md` at least held the text projects would inherit in front of the agent.
  That reason left with the documents: they ship from the submodule now, so the checklist was
  costing every session 13.4 KiB to describe a moment that never happens here. It is read at
  `.code-server/docs/agent/en/INITIALIZATION.md` when the subject is the interview itself.

## If the imports below did not load

The normative documents — the modes and the rules — ship from the template and are pulled in by the
`@path` lines below. They live inside the `.code-server/` submodule, which is **empty until
`git submodule update --init`**, and an import that resolves to nothing **resolves to nothing
silently: no error, no warning**. That is observed behaviour, not a guess — see "Open edges" in
`docs/ARCHITECTURE/OVERVIEW.md` for the fixture and the output. The one cue left is that the `@`
line below stays visible with no content behind it.

So: if you cannot see the pairing modes or the ground rules in your context, **stop and say so**
rather than proceeding. An agent working without them is not working under a lighter process, it is
working with no mode, no rules and no gates, and nothing failed to tell anybody. Four gates in
particular exist and are not optional — the charter agreed with the user before the SRS, the SRS
before any story, a story's scenarios before its tasks, and a task's design before its code. The
chain is `.code-server/docs/agent/en/WORKFLOW.md`.

@.code-server/docs/agent/en/MODES.md
@docs/RULES.md

## What this repository is

A reference monorepo that has already adopted the code-server dev-container template. The template
itself — `core/`, `stacks/`, `start/`, `setup`, the process documents in `docs/agent/`, and the full
design-rationale doc — lives in its own repo,
[`jvsl.env.agents.code-server`](https://github.com/TheHefty/jvsl.env.agents.code-server), vendored
here as a git submodule at `.code-server/` (the officially documented way to consume it — not a
copy-paste drop-in). Splitting it out means any monorepo consuming the template pulls in updates
with a plain `git submodule update` (bump the pinned tag) instead of needing a rebase against
upstream template history.

This repo is not itself an application — there is no product code here, only the root-level
scaffolding a consuming monorepo keeps outside the submodule: this `CLAUDE.md`, `README.md`,
`docs/OVERVIEW.md`, `docs/RULES.md` (one import line plus whatever the project adds), the empty
`docs/PLANNING/` and `docs/DEBTS/` folders, and `.code-server.stack.json` (the per-project stack
selection — see "Manifest" in `.code-server/docs/overview/setup.md` for why it cannot live inside the
submodule).

Full design rationale for the template — every decision made, the exact structure of
`core/`/`stacks/`, the manifest format, how the process documents are delivered, and every build
error hit and fixed along the way — is recorded in `.code-server/docs/overview/`, inside the
submodule, so it versions together with the template rather than with this consuming repo. Treat
that file as the authoritative, up-to-date spec for anything under `.code-server/`.
`docs/OVERVIEW.md` at this repo's root is the short user-facing "how to use this template" version.

## Commands

Build the dev image (interactive; also how you add/remove stacks later):
```bash
.code-server/setup
```
Requires `jq`, `whiptail` and `docker` on the host. It runs on the host, before any image exists,
so it cannot depend on anything from inside the image it builds.

Build the native launcher (Rust/Tauri; once, or after editing `.code-server/start/src/main.rs`):
```bash
cd .code-server/start && cargo build --release
```
Requires Rust and the Tauri Linux prerequisites — exact packages per distro are in
`.code-server/docs/overview/start.md`.

Launch the environment:
```bash
.code-server/start/target/release/start
```

Point an already-created project at the template's process documents instead of at the copies it
was created with (dry run without `--apply`):
```bash
.code-server/migrate-agent-docs.sh --lang en
```

**There is no test suite and no lint config in this repo.** The template has CI inside the
submodule (`.code-server/.github/workflows/ci.yml`): `bash -n` over the shell scripts,
`cargo check --release --locked` on `start/`, the process-document checks, and a `docker build` per
stack — plus `ci-green`, the single check the template's branch protection requires. Changes under
`.code-server/` are therefore verifiable there; nothing in this consuming repo runs any of it.

## Architecture

The detail is in `.code-server/docs/overview/` and is not repeated here. What a reader needs
before opening it:

- **`core/`** — the mandatory base layer: code-server, Node.js, the Claude Code CLI, `ai-jail`, the
  GitHub CLI, Rust and the Tauri Linux libs. Two things are load-bearing and easy to get wrong.
  Docker inside the container is a **nested rootless daemon**, not the host's socket: mounting the
  host socket made everything in the container root-equivalent on the host and silently voided the
  sandbox. And `CLAUDE_CONFIG_DIR=/config/.claude` keeps the whole CLI state in the bind-mounted
  directory rather than only its credentials.
- **`stacks/<name>/`** — one directory per selectable tech stack, each with a `Dockerfile.frag`
  using a `{{VERSION}}` placeholder, a `versions.json`, and optionally a `requires.json` and a
  `cont-init/`. `java` was the original and is the pattern to copy; `android` is the only one with
  a `requires.json`, and `php` the only one pinning a third-party signing key.
- **`setup`** — reads `.code-server.stack.json` at this repo's root, offers a `whiptail`
  checklist, rewrites the manifest, composes `.code-server/Dockerfile` and builds. The generated
  Dockerfile is gitignored and never hand-edited; removing a stack is excluding it from the
  manifest.
- **`start/`** — a Tauri v2 Rust app with no JS frontend. It ensures the container is running,
  reads back the host port Docker published, and opens a native window rather than a browser tab —
  deliberately, so the browser cannot intercept editor shortcuts. It publishes to
  `127.0.0.1:0:8443` rather than using `--network host`, so several projects can run at once, and
  passes `/dev/fuse`, `/dev/net/tun` and `/dev/kvm` through only when the host actually has them.
- **`docs/agent/`** — the process documents this repo imports: one folder per language, delivered
  by bump, with `check-parity.sh` guarding the languages against drifting apart.
- **Versioning** — the template is released with release-please from its own conventional commits.
  Bump `.code-server/` to a **tag**, never a bare commit: a commit reachable only from a branch
  becomes unreachable once that branch is squash-merged, and then every fresh clone fails its
  `git submodule update`.

## Releases

Released the same way the template is: release-please (`.github/workflows/release-please.yml`)
keeps a release PR open on `master` and cuts the tag when it is merged, with `version.txt` and
`CHANGELOG.md` as the only versioned artifacts (`release-type: simple` — there is nothing here to
publish).

`master` is protected, so **there is no direct push to it** — every change, including a submodule
bump or a one-line doc fix, goes through a pull request. No approvals are required (single
maintainer), but the rule applies to administrators too, and force-pushes and branch deletion are
blocked. There are no required status checks, because this repo has no CI of its own: a PR here is
mergeable as soon as it is open. Head branches are deleted automatically on merge.

**Give a feature PR a non-conventional title when it is merged with a merge commit. Never the
release PR.** The two halves are opposite and both were learned the hard way at 1.6.0.

`gh pr merge --merge` writes the pull request's title into the body of the merge commit, so a
feature PR titled `feat: ...` is read by release-please a second time and the entry lands in the
changelog twice — once for the real commit, once for the merge that carried it. Observed through
1.4.0 and 1.5.0, and again at 1.6.0.

The release PR is the exact opposite: release-please parses the version out of its title, against
the pattern `chore${scope}: release ${version}`. Rename it and the next run finds the merged PR,
cannot read a version from it, and aborts with `There are untagged, merged release PRs outstanding`
— the release is merged, the changelog is on `main`, and no tag exists. Recovering means restoring
the title and re-running the workflow.

This file used to say the duplicate was "confirmed fixed in 1.5.2". It was not: no feature PR
between 1.5.2 and 1.6.0 happened to carry a conventional title, and an absence of symptoms was
written down as a repair.

When a duplicate does land, it has to be removed in two places. Drop the line from `CHANGELOG.md`
on the release branch before merging it (`cc672fb` in the template is the precedent) — and then fix
the GitHub release body too, because release-please generates the release notes from the commits
rather than from the file, so the file being clean does not make the release page clean.

`bootstrap-sha` pins the changelog's starting point at `54628be`, the commit that vendored the
template as a submodule. Everything before it describes `core/`/`stacks/` work that has since moved
into the template's own repo.

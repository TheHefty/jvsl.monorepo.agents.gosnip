# Overview

Dev environment template based on [code-server](https://github.com/coder/code-server), with the
Claude Code CLI, `ai-jail`, `ai-memory`, the GitHub CLI, and a nested rootless Docker daemon
already set up. The template itself lives in its own repo,
[`jvsl.env.agents.code-server`](https://github.com/TheHefty/jvsl.env.agents.code-server),
vendored here as a git submodule at `.code-server/` (doesn't clutter the root of the monorepo that
adopts it) and driven by three executables: `init` (prepares the host, once per machine), `setup`
(chooses the tech stacks that go into the image) and `dev` (builds the launcher if it is stale and
opens the environment in a native window).

The full design — decisions made, the structure of `core/`/`stacks/`, the manifest format, bugs
already hit and fixed — is in
[`.code-server/docs/overview/`](../.code-server/docs/overview/), inside the submodule (so it
versions with the template, not with this consuming repo). This document is just the "how to use
it" summary, and covers what's specific to being a *consumer* of the template.
[`RULES.md`](RULES.md) alongside it carries the ground rules a project built on this template is
expected to follow, and why each one exists; [`PLANNING/`](PLANNING/) holds the charter, the SRS and
the epics, stories and tasks work is broken into — the decisions that were expensive enough to be
worth writing down, and the Gherkin acceptance criteria beside each story;
[`DEBTS/`](DEBTS/) holds the fixes and shortcuts made outside that chain;
[`ARCHITECTURE/OVERVIEW.md`](ARCHITECTURE/OVERVIEW.md) describes the system a project builds, as it
stands today.

## Updating the template

The template is released with [release-please](https://github.com/googleapis/release-please), so it
carries tags and a `CHANGELOG.md`. Prefer pinning to a tag over a bare commit:
```bash
cd .code-server && git fetch --tags && git checkout v1.0.1 && cd ..
git add .code-server && git commit -m "chore: bump .code-server template to v1.0.1"
```
This is the whole point of the submodule split: pulling in template changes (new stacks, fixes)
never requires rebasing this repo's own history against the template's. Rerun `.code-server/setup`
after a bump — the image is what actually carries the change, and it is never updated in place.

## Starting the environment

1. **Prepare the host** (once per machine):
   ```bash
   .code-server/init
   ```
   It checks the display before anything else — on WSL that means WSLg, where "installed but not
   running" is `wsl --shutdown` from Windows rather than a package — then checks every build and
   runtime dependency and offers to install what is missing, naming each one per package manager.
   `cargo` is the one exception it will not install: a distribution's Rust is routinely older than
   the Tauri crates need, so it prints the `rustup` line and leaves the choice to you.

   Skipping `init` works, but each missing piece fails in a way that doesn't name itself: a missing
   `libwebkit2gtk-4.1-dev` surfaces forty seconds into a build as `cannot find -lwebkit2gtk-4.1`,
   a missing `whiptail` surfaces as `setup` exiting with a blank screen, and a WSL without the X
   client libraries opens a window that is simply blank, with nothing logged.

2. **Build the image** (choose the monorepo's stacks, generate the Dockerfile, and build it):
   ```bash
   .code-server/setup
   ```
   Rerun `setup` any time you want to add or remove a stack. A stack can depend on another one
   (today only `android`, which needs `java`); `setup` refuses a selection that leaves the
   dependency unchecked rather than adding it silently, since its version is yours to choose.

3. **Open the environment**:
   ```bash
   .code-server/dev
   ```
   `dev` builds the launcher whenever it is older than its own sources and then runs it, so there
   is no separate `cargo build` step to remember and no stale binary to notice. On WSL it also
   sets `WEBKIT_DISABLE_COMPOSITING_MODE=1` — WebKit's own switch for a disagreement with WSLg's
   compositor that otherwise opens the window and leaves it blank.

   The first run creates the container (`docker run`, with the workspace mounted); after that it
   just ensures the container is running (`docker start`) and opens the window. No env var needs
   to be passed to run it from within the repo's own structure — image/container name and
   `START_WORKSPACE_DIR` have an automatically derived default.

What the host provides decides part of what you get inside, and `start` adapts instead of failing:
`docker` in the container is a nested rootless daemon that needs `/dev/fuse` and `/dev/net/tun` —
without `/dev/fuse` the daemon stays down and `docker` simply isn't available — and the `android`
stack's headless emulator needs `/dev/kvm`, which exists on Linux hosts with VT-x/AMD-V but has no
equivalent under Docker Desktop's macOS/Windows VM. The container is also capped at 8g of memory
(plus 2g of swap) and pinned to half the host's cores.

## Long-term memory (`ai-memory`)

[`ai-memory`](https://github.com/akitaonrails/ai-memory) gives the agent memory that outlives a
session: quit mid-task, come back tomorrow (or in a different agent CLI), and the next session opens
with a handoff instead of a blank slate. The memory itself is plain markdown in a git repo —
`grep`-able and readable without any of this — kept on the container's persistent volume at
`/config/ai-memory`, not in your repository.

**It is off until you ask for it, per project.** The switch is a marker file at the repo root:

```bash
printf '[project]\nproject_strategy = "repo-root"\n' > .ai-memory.toml
```

Then restart the container — the marker is read at boot. `ai-memory status` shows the paths and
counts once it is up.

Without that file nothing runs and nothing is recorded: the server never starts listening, the
installed hooks are in allowlist mode and emit no event at all, and the agent's sandbox is not
widened to reach the store. Forgetting the marker costs you recall, never confidentiality.

Two things worth knowing before you turn it on. **Memory is per project**, because the server runs
inside this project's container — a second project gets its own, and they never see each other.
And **no LLM provider is configured**, so nothing captured leaves the machine: you get full-text,
entity and graph-neighbour search plus rule-based summaries. Adding a provider buys consolidated
pages and contradiction linting, and costs sending captured prompts and tool excerpts to it.

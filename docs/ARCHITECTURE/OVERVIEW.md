# Architecture overview

What this repository is, right now.

**There is no product architecture to describe.** This is the reference monorepo for the
code-server dev-container template: root-level scaffolding, a submodule, and no application code.
What the environment itself is made of — the container, the sandbox, the nested rootless Docker
daemon, the launcher — belongs to the template and is documented in
[`.code-server/docs/overview/`](../../.code-server/docs/overview/), which versions with it
rather than with this repository.

A project built on this template fills this file in as its system grows. How to write it, and the
sections to fill, ship with the template:
[`.code-server/docs/agent/en/ARCHITECTURE.md`](../../.code-server/docs/agent/en/ARCHITECTURE.md).

## Verified behaviour of the `@path` imports

The whole delivery mechanism rests on these, so they were tested rather than assumed. Claude Code
2.1.241, fixture directories with unique canary strings, `claude -p` reading them back.

- **A top-level `@` reaches into a subdirectory, including the submodule.** `@sub/deep/LEVEL1.md`
  from a `CLAUDE.md` returned `L1=XKCDTOPLEVEL7788`. The live case is this repository's own session:
  `MODES.md` is imported from inside `.code-server/` and arrives.
- **A second-level import resolves relative to the file that contains it, not to the repository
  root.** `docs/MID.md` importing `@../sub/deep/LEVEL2.md` returned `L2=ZQPNESTED4412`. This is what
  makes `docs/RULES.md` write `@../.code-server/docs/agent/en/RULES.md`; without the `..` it would
  point at `docs/.code-server/`, which does not exist.
- **An import that cannot be resolved fails silently.** With the target directory present but empty
  — exactly what an uninitialised submodule looks like — the run reported the canary from the file
  that did resolve, `no` to having received the missing ones, and `no` to any error or warning. The
  harness says nothing at all.
- **The one cue is the `@` line itself.** It stays visible in the loaded file with no content behind
  it, and the agent in the fixture could quote it. So the rescue text in `CLAUDE.md` is not shouting
  into a void: an agent that reads the import lines and notices the missing content can act on it.
  It has to know to look, which is why the mode, the gates and the stop-and-say-so instruction stay
  written in `CLAUDE.md` itself.

## Open edges

- **The process documents now depend on a Claude Code feature.** `@path` imports are not a Markdown
  convention; while these documents sat at this root, any CLI that reads files reached them, and now
  a CLI without imports gets the literal `@` line and nothing else. The rescue text in `CLAUDE.md`
  only helps a CLI that at least reads `CLAUDE.md` — Codex reads `AGENTS.md`, Gemini CLI reads
  `GEMINI.md`, and neither is this file.

  What mitigating costs, cheapest first: an `AGENTS.md` that is a one-line pointer to
  `.code-server/docs/agent/<lang>/` — near-free, and works for any agent that reads its own entry
  file and can open a path. Or a generated flattened copy with the imports inlined — which buys
  exactness and costs a build step, a file that goes stale silently, and a check that it is current;
  that is the copy-drift problem this whole change exists to end, reintroduced one layer down. The
  first is worth doing when a second CLI is actually in use; the second is not worth doing at all
  until the first has failed.

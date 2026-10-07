# AGENTS.md

Read and follow `CLAUDE.md`. It is the canonical instruction file for this repository.

One thing that file cannot tell you itself: the inherited modes and rules reach Claude Code from
`/config/.claude/rules/` inside the agent container, and the project's own rules through an `@path`
import, which is a Claude Code feature and not a Markdown convention. If your CLI does not load
them, open them by hand — `/config/.claude/rules/MODES.md`, `/config/.claude/rules/RULES.md` and
`docs/RULES.md` — because a document that fails to load fails *silently*: no error, no warning. Working without those
two documents is not working under a lighter process; it is working with no mode, no rules and no
gates.

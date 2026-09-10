# AGENTS.md

Read and follow `CLAUDE.md`. It is the canonical instruction file for this repository.

One thing that file cannot tell you itself: it pulls the normative documents in with `@path`
import lines, which are a Claude Code feature and not a Markdown convention. If your CLI does not
resolve them, open the two paths by hand — `.code-server/docs/agent/en/MODES.md` and
`docs/RULES.md` — because an import that resolves to nothing resolves to nothing *silently*: no
error, no warning, only the `@` line left visible with no content behind it. Working without those
two documents is not working under a lighter process; it is working with no mode, no rules and no
gates.

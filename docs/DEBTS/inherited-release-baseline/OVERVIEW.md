# Debt: Inherited release baseline

| | |
|---|---|
| **Status** | Paid back |
| **Date** | 2026-09-10 |
| **Kind** | shortcut |

## Problem

After project initialization was marked complete, `CHANGELOG.md` still presented releases 1.0.0
through 1.8.0 from `jvsl.monorepo.agents.template` as this repository's history. `version.txt` and
`.release-please-manifest.json` also remained at 1.8.0, while `release-please-config.json` stopped at
a bootstrap commit that does not exist in the `gosnip` history. A future product feature would
therefore continue the template version line instead of producing the accepted first `gosnip`
release, 0.1.0.

## Root cause

The initialization cleanup deliberately deferred inherited release files to the later distribution
story. That treated false product identity and a broken release baseline as future implementation
rather than initialization data that had to be reset before the project was true.

## Fix

The changelog was reset to an unreleased `gosnip` state, both version sources were reset to 0.0.0,
and the bootstrap boundary was changed to this repository's actual initial commit. Project
documentation now describes that clean baseline. The existing release-please workflow verifies the
state before attempting release automation.

## Regression scenario

`scripts/release-state.test.sh` was written first and observed failing because the changelog named
the template repository and the configured bootstrap commit was absent. It verifies that release
history never names the template, the manifest and version file agree, the bootstrap commit exists,
and the latest changelog entry matches the tracked version after releases begin. The
`release-please` workflow runs it with complete git history.

## Payback

Paid back in the same correction that records this debt. The later `distribute-gosnip` story still
owns product artifact generation; it does not need to revisit template history or the initial
version baseline.

## Outcome

Accepted as permanently resolved on 2026-09-10. The repository now starts at 0.0.0 with no product
release entries, allowing the first accepted feature release to become 0.1.0.

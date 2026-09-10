# Project Charter: gosnip

| | |
|---|---|
| **Status** | Accepted |
| **Date** | 2026-09-10 |
| **Author** | João Lima |
| **Kind** | new |

## Purpose

`gosnip` exists so its owner can keep reusable code snippets on their own machine, find them
quickly, and copy them into the work at hand without depending on a hosted service. It is also a
small, weekend-sized project through which the owner can learn idiomatic Go by implementing a
complete command-line application.

## In scope / out of scope

The first version stores snippets with a name, code, language, and simple tags in a local SQLite
database. It lets the owner create, list, search, copy, edit, and delete those snippets through a
command-line interface.

The first version deliberately has no graphical interface, accounts, hosted service, cloud
storage, synchronization between machines, collaboration, or multi-user behavior. It does not
store credentials, identity documents, sensitive personal data, or data belonging to anyone other
than the owner.

## Stakeholders

The sole owner is the user, learner, product authority, and final judge of what "done" means. No
other stakeholder needs to be consulted unless the ownership or audience changes.

## Standing decisions

- **Mode of work:** Navigator Mode, so the owner writes the Go code while the agent guides design,
  highlights concrete risks and idioms, and performs verification.
- **Documentation language:** English for documentation, planning artifacts, and commit messages,
  preventing the project record from becoming split across languages.
- **Long-term memory:** enabled locally to preserve decisions and handoffs across sessions. Prompts
  and tool excerpts may be recorded on disk for this project; no LLM provider is configured, so
  that record does not leave the machine unless the owner later adds one.
- **Licence:** MIT, allowing reuse with attribution and a warranty disclaimer. The selected SQLite
  driver, `modernc.org/sqlite`, is distributed under the compatible BSD 3-Clause licence.
- **Use and privacy:** personal, non-economic use of the owner's own nonsensitive snippet data. No
  additional privacy policy is adopted at this stage because the system has no other users and
  does not process anyone else's personal data.

## What would change this charter

This charter must be reconsidered before the project serves another person, charges for access or
use, processes data belonging to someone else, stores sensitive data or credentials, introduces a
hosted or synchronized service, changes ownership, or takes on a purpose beyond personal snippet
management and Go learning.

## Outcome

Accepted by the owner on 2026-09-10 after the initialization grilling. The proposed minimal
operation set was expanded during review so the first version includes editing and deletion and
therefore provides the complete local snippet lifecycle.

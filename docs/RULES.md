# Rules

The inherited rules ship inside the agent container's image and are loaded from
`/config/.claude/rules/RULES.md`; they are updated by updating the Agent Container extension. They
are not edited here: a rule that needs changing is changed in the extension's repository, where
changing it reaches every project rather than only this one.

---

Everything below this line belongs to this repository.

`gosnip` has no additional project-specific rules yet. The inherited rules above govern the
project; standing decisions belong in `docs/CHARTER.md` and `CLAUDE.md` rather than being duplicated
here.

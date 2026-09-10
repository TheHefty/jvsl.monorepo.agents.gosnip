# Debts

Fixes and shortcuts made outside the charter/SRS/story/task chain — a production hotfix that could
not wait for a story, or a corner cut knowingly with the intent of paying it back. One folder per
item, each with an `OVERVIEW.md` in the shape problem → root cause → fix → regression scenario. The
form ships with the template: see
[`.code-server/docs/agent/en/DEBT-TEMPLATE.md`](../../.code-server/docs/agent/en/DEBT-TEMPLATE.md).

Recorded items:

- [`inherited-release-baseline/`](inherited-release-baseline/) — the template changelog, version,
  and invalid bootstrap commit left behind when initialization was first marked complete; paid back
  by resetting the product release state and adding its regression check.

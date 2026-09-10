# Security Policy

## Supported versions

`gosnip` has no product release yet. Once 0.1.0 is released, only the most recent product release
will receive fixes. The 0.0.0 repository baseline is not a supported release.

## Reporting a vulnerability

Report privately through GitHub:
**[Security → Report a vulnerability](https://github.com/TheHefty/jvsl.monorepo.agents.gosnip/security/advisories/new)**.

Please do not open a public issue for something you believe is exploitable. This is a personal
project with one maintainer, so acknowledgement is best-effort with no guaranteed response time or
bounty.

## Scope

Product security reports belong here when they concern the future `gosnip` executable, local
SQLite data, migrations or backups, editor and clipboard integration, command/output contracts, or
release artifacts and workflows. There is no product code yet, but these boundaries are fixed by
the accepted SRS.

`gosnip` is designed for the owner's nonsensitive snippets. It does not promise secure erasure and
must not be used to store credentials, tokens, identity documents, or sensitive personal data.
Normal operation is offline and performs no telemetry or update checks.

The development container, agent sandbox, nested Docker daemon, and native launcher belong to the
template vendored at `.code-server/`. Report vulnerabilities in those components through the
[template security policy](https://github.com/TheHefty/jvsl.env.agents.code-server/blob/main/SECURITY.md)
instead.

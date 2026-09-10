# Security Policy

## Supported versions

Only the most recent release is supported. See [`CHANGELOG.md`](CHANGELOG.md).

## Reporting a vulnerability

Report privately through GitHub: **[Security → Report a
vulnerability](https://github.com/TheHefty/jvsl.monorepo.agents.template/security/advisories/new)**.

Please do not open a public issue for something you believe is exploitable.

This is a personal project with a single maintainer — expect a best-effort acknowledgement, with no
guaranteed response time and no bounty.

## Where the security-relevant code actually lives

This repository is scaffolding. It contains no application code and no build logic — only this
policy, the documentation, the release configuration, and `.code-server.stack.json`, which selects
which tech stacks the dev container is built with.

Everything that builds or runs anything — the container image, the nested rootless Docker daemon,
`ai-jail`, and the native launcher and its `docker run` flags — comes from the dev-container
template vendored at `.code-server/`, which is a git submodule of
[`jvsl.env.agents.code-server`](https://github.com/TheHefty/jvsl.env.agents.code-server).

**If your finding concerns the container, the sandbox, or the launcher, report it there instead:**
[its security policy](https://github.com/TheHefty/jvsl.env.agents.code-server/blob/main/SECURITY.md)
covers the threat model, what is in scope, and which permissive behaviours are deliberate. Reporting
it against this repository only adds a hop.

What is worth reporting *here* is something specific to this repo as a consumer of that template —
for example a stack selection in `.code-server.stack.json` that pulls in something it should not, or
a workflow in `.github/workflows/` that mishandles a token.

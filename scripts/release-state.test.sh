#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

failed=0

if rg -q 'jvsl\.monorepo\.agents\.template|github\.com/TheHefty/jvsl\.monorepo\.agents\.template' CHANGELOG.md; then
    echo "release-state: CHANGELOG.md still contains template release history" >&2
    failed=1
fi

manifest_version="$(jq -er '.["."]' .release-please-manifest.json)"
file_version="$(tr -d '[:space:]' < version.txt)"
if [[ "$manifest_version" != "$file_version" ]]; then
    echo "release-state: manifest version '$manifest_version' differs from version.txt '$file_version'" >&2
    failed=1
fi

bootstrap_sha="$(jq -er '."bootstrap-sha"' release-please-config.json)"
if ! git cat-file -e "${bootstrap_sha}^{commit}" 2>/dev/null; then
    echo "release-state: bootstrap-sha '$bootstrap_sha' is not a commit in this repository" >&2
    failed=1
fi

if [[ "$file_version" == "0.0.0" ]]; then
    if rg -q '^## ' CHANGELOG.md; then
        echo "release-state: unreleased baseline contains a release section" >&2
        failed=1
    fi
else
    first_release="$(sed -nE 's/^## \[?([0-9]+\.[0-9]+\.[0-9]+)\]?.*/\1/p' CHANGELOG.md | head -n 1)"
    if [[ "$first_release" != "$file_version" ]]; then
        echo "release-state: latest changelog version '$first_release' differs from version.txt '$file_version'" >&2
        failed=1
    fi
fi

if ((failed)); then
    exit 1
fi

echo "release-state: changelog identity, version files, and bootstrap commit are consistent"

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOK="$ROOT/.githooks/pre-push"

fail() {
    echo "pre-push test: $*" >&2
    exit 1
}

run_hook() {
    local repo="$1"
    local refs="$2"
    local output_file="$3"

    (cd "$repo" && printf '%s\n' "$refs" | bash "$HOOK") >"$output_file" 2>&1
}

test_repo="$(mktemp -d)"
trap 'rm -rf "$test_repo"' EXIT

git -C "$test_repo" init -q
mkdir -p "$test_repo/scripts"
printf '#!/usr/bin/env bash\nexit "${UNIT_STATUS:-0}"\n' >"$test_repo/scripts/test-unit.sh"
chmod +x "$test_repo/scripts/test-unit.sh"

output="$test_repo/output"
master_ref="refs/heads/topic $(printf '1%.0s' {1..40}) refs/heads/master $(printf '2%.0s' {1..40})"
if UNIT_STATUS=0 run_hook "$test_repo" "$master_ref" "$output"; then
    fail "a direct push to master was allowed"
fi
grep -q "refusing to push straight to master" "$output" ||
    fail "master rejection did not explain the failure"

topic_ref="refs/heads/topic $(printf '1%.0s' {1..40}) refs/heads/topic $(printf '2%.0s' {1..40})"
if UNIT_STATUS=23 run_hook "$test_repo" "$topic_ref" "$output"; then
    fail "a failing unit suite was ignored"
else
    status=$?
fi
[ "$status" -eq 23 ] || fail "unit status was $status, want 23"

UNIT_STATUS=0 run_hook "$test_repo" "$topic_ref" "$output" ||
    fail "a valid topic push with passing tests was rejected"

echo "pre-push test: ok"

#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "unit: Go"
(cd "$root/gosnip" && go test -short ./...)

echo "unit: pre-push hook"
bash "$root/.githooks/pre-push.test.sh"

echo "unit: ok"

#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT/scripts/deploy-eks.sh"
fail() { echo "FAIL: $*" >&2; exit 1; }
[[ -f "$SCRIPT" ]] || fail "deployment helper missing"
bash -n "$SCRIPT" || fail "Bash syntax invalid"
if bash "$SCRIPT" --image ghcr.io/example/kubecis:latest >/dev/null 2>&1; then
  fail "mutable image was accepted"
fi
if bash "$SCRIPT" --image 'ghcr.io/example/kubecis@sha256:123' >/dev/null 2>&1; then
  fail "short digest was accepted"
fi
if bash "$SCRIPT" --image "ghcr.io/example/kubecis@sha256:$(printf 'a%.0s' {1..64})" --namespace default >/dev/null 2>&1; then
  fail "unsupported namespace was accepted"
fi
echo 'PASS: deployment helper validates syntax and rejects unsafe arguments'

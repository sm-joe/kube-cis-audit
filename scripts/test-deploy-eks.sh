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
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
case "$*" in
  *"version -o json"*) printf '{"serverVersion":{"gitVersion":"v1.35.0-eks-test"}}\n' ;;
  *"config current-context"*) echo test-eks ;;
  *"kustomize "*) printf 'image: ghcr.io/sm-joe/kubecis:eks-provider-rc3-identity\n%.0s' {1..3} ;;
  *"apply -f -"*) cat > "$CAPTURE" ;;
  *"rollout status"*) exit 0 ;;
  *) exit 1 ;;
esac
EOF
chmod +x "$tmp/bin/kubectl"
export CAPTURE="$tmp/applied"
digest="ghcr.io/example/kubecis@sha256:$(printf 'a%.0s' {1..64})"
PATH="$tmp/bin:$PATH" bash "$SCRIPT" --context test-eks --image "$digest"
[[ "$(grep -Fc "image: $digest" "$CAPTURE")" == 3 ]] || fail "not all three containers pinned"
if grep -q 'eks-provider-rc3-identity' "$CAPTURE"; then fail "legacy tag retained"; fi
echo 'PASS: deployment helper rejects unsafe arguments and pins all three containers (mock Kubernetes API)'

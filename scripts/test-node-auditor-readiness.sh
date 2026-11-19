#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/collect-eks.sh"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/scripts" "$tmp/bin" "$tmp/benchmarks/cis-eks-2.1.0"
cp "$root/scripts/collect-eks.sh" "$tmp/scripts/"
printf 'id: cis-eks-2.1.0\n' > "$tmp/benchmarks/cis-eks-2.1.0/benchmark.yaml"
cat > "$tmp/kubecis" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == version ]]; then echo fixture-version; exit 0; fi
if [[ "$1" == scan ]]; then
  while (($#)); do
    if [[ "$1" == --output-file ]]; then
      shift; printf '{"auditMode":"native","cloudProvider":"eks","summary":{"total":0},"results":[]}' > "$1"
      exit 0
    fi
    shift
  done
fi
exit 1
EOF
chmod +x "$tmp/kubecis"
cat > "$tmp/bin/aws" <<'EOF'
#!/usr/bin/env bash
cat <<'JSON'
{"cluster":{"name":"test-eks","endpoint":"https://eks.example.invalid","version":"1.35"}}
JSON
EOF
cat > "$tmp/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
args=" $* "
if [[ "$args" == *" config current-context"* ]]; then echo test-context; exit 0; fi
if [[ "$args" == *" version -o json"* ]]; then echo '{"serverVersion":{"gitVersion":"v1.35.1-eks-example"}}'; exit 0; fi
if [[ "$args" == *" config view --minify -o json"* ]]; then echo '{"clusters":[{"cluster":{"server":"https://eks.example.invalid"}}]}'; exit 0; fi
if [[ "$args" == *" get nodes -o json"* ]]; then
  echo '{"items":[{"metadata":{"name":"ip-10-10-0-165.ec2.internal","labels":{"kubernetes.io/os":"linux","eks.amazonaws.com/nodegroup":"workers"}},"spec":{"taints":[{"key":"dedicated","effect":"NoSchedule"}]}}]}'
  exit 0
fi
if [[ "$args" == *" get daemonset kubecis-node-auditor"* ]]; then
  [[ "$FIXTURE_CASE" != missing-ds ]] || exit 1
  echo '{"status":{"desiredNumberScheduled":1,"currentNumberScheduled":1,"numberReady":0}}'
  exit 0
fi
if [[ "$args" == *" get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json"* ]]; then
  if [[ "$FIXTURE_CASE" == pending ]]; then
    echo '{"items":[{"metadata":{"name":"kubecis-node-auditor-123"},"spec":{"nodeName":"ip-10-10-0-165.ec2.internal"},"status":{"phase":"Pending","initContainerStatuses":[{"name":"node-metadata-init","ready":false,"state":{"waiting":{"reason":"ImagePullBackOff"}},"restartCount":0}]}}]}'
  else
    echo '{"items":[]}'
  fi
  exit 0
fi
exit 1
EOF
chmod +x "$tmp/bin/aws" "$tmp/bin/kubectl"
args=(--context test-context --cluster-name test-eks --region us-east-1 --binary "$tmp/kubecis")
if FIXTURE_CASE=missing-ds PATH="$tmp/bin:$PATH" bash "$tmp/scripts/collect-eks.sh" "${args[@]}" > "$tmp/log" 2>&1; then
  echo 'FAIL: missing DaemonSet was accepted' >&2; exit 1
fi
grep -q 'DaemonSet is not installed' "$tmp/log" || { cat "$tmp/log"; exit 1; }
if FIXTURE_CASE=pending PATH="$tmp/bin:$PATH" bash "$tmp/scripts/collect-eks.sh" "${args[@]}" > "$tmp/log" 2>&1; then
  echo 'FAIL: pending node auditor was accepted' >&2; exit 1
fi
grep -q 'ImagePullBackOff' "$tmp/log" || { cat "$tmp/log"; exit 1; }
grep -q 'dedicated:NoSchedule' "$tmp/log" || { cat "$tmp/log"; exit 1; }
echo 'PASS: absent DaemonSet and Pending Pod readiness diagnostics'

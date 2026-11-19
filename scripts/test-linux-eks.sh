#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
for name in collect-eks-provider.sh collect-eks-ecr.sh collect-eks.sh validate-eks-preview.sh deploy-eks.sh; do
  bash -n "$root/scripts/$name"
done
if bash "$root/scripts/validate-eks-preview.sh" --release-level full --report /dev/null --cluster-name sample --region us-east-1 >/dev/null 2>&1; then
  echo 'FAIL: full CIS compliance gate unexpectedly accepted' >&2
  exit 1
fi
if bash "$root/scripts/collect-eks.sh" --context dummy --cluster-name sample --region us-east-1 --namespace default >/dev/null 2>&1; then
  echo 'FAIL: unsupported namespace unexpectedly accepted' >&2
  exit 1
fi
# Exercise the contract gate using an explicit invalid document, independent
# of any live AWS identity. Partial input cannot become a passing assessment.
tmp="$(mktemp)"
trap 'rm -f -- "$tmp"' EXIT
echo '{"auditMode":"consolidated","cloudProvider":"eks","compliance":{"compliancePercent":100}}' > "$tmp"
if bash "$root/scripts/validate-eks-preview.sh" --report "$tmp" --cluster-name sample --region us-east-1 >/dev/null 2>&1; then
  echo 'FAIL: invented 100% CIS compliance incorrectly accepted' >&2
  exit 1
fi
echo 'PASS: POSIX script syntax and fail-closed input contracts'

# A minimal structurally valid preview dossier must pass. Altered identities,
# omitted node evidence or an invented compliance percentage must be rejected.
now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
jq -n --arg now "$now" '
  def nodefinding($id):
    {id:$id,target:"node",scope:"worker-a",status:"PASS",
     evidence:[{name:"etc/kubernetes/kubelet/config.json"}]};
  def awsfinding($id):
    {id:$id,target:"policies",method:"aws_eks_describe_cluster",status:"PASS",
     evidence:[{name:"EKS"}]};
  {auditMode:"consolidated",cloudProvider:"eks",benchmark:"cis-eks-2.1.0",
   benchmarkCompatibilityStatus:"provisional",clusterVersion:"v1.35.9-eks-test",
   expectedNodes:["worker-a"],expectedControlPlaneNodes:[],
   sources:[{auditMode:"native",generatedAt:$now,findings:0},
            {auditMode:"node-local",nodeName:"worker-a",generatedAt:$now,findings:3},
            {auditMode:"aws-provider",generatedAt:$now,findings:3}],
   compliance:{compliancePercent:null,scoreStatus:"withheld: provisional",
               scoredTotal:6,scoredPass:6,scoredFail:0},
   providerEvidence:{schemaVersion:1,assessmentStatus:"INVENTORY_ONLY",
     cloudProvider:"eks",clusterName:"sample",region:"us-east-1",
     clusterArn:"arn:aws:eks:us-east-1:123456789012:cluster/sample",
     kubernetesVersion:"1.35",generatedAt:$now,
     controlPlaneLogging:{api:true,audit:true,authenticator:true,
                          controllerManager:true,scheduler:true}},
   results:[nodefinding("3.1.3"),nodefinding("3.2.1"),nodefinding("3.2.2"),
            awsfinding("2.1.1"),awsfinding("5.4.1"),awsfinding("5.4.2")],
   summary:{total:6}
  }' > "$tmp"
bash "$root/scripts/validate-eks-preview.sh" --report "$tmp" --cluster-name sample --region us-east-1 >/dev/null
if bash "$root/scripts/validate-eks-preview.sh" --report "$tmp" --cluster-name wrong --region us-east-1 >/dev/null 2>&1; then
  echo 'FAIL: wrong AWS cluster identity accepted' >&2; exit 1
fi
jq '.results |= map(select(.id != "3.2.2")) | .summary.total = 5' "$tmp" > "$tmp.new"
mv "$tmp.new" "$tmp"
if bash "$root/scripts/validate-eks-preview.sh" --report "$tmp" --cluster-name sample --region us-east-1 >/dev/null 2>&1; then
  echo 'FAIL: missing essential kubelet control accepted' >&2; exit 1
fi
echo 'PASS: positive fixture and missing-evidence/identity regression checks'


# Regression: a missing compiled CLI and a --binary cmd/kubecis directory
# should both trigger a source build instead of a directory execution failure.
# Test with a disposable fixture repository and a MOCK Go builder. This does
# not replace the real Go build/test CI job.
fixture="$(mktemp -d)"
trap 'rm -rf -- "$tmp" "$fixture"' EXIT
mkdir -p "$fixture/scripts" "$fixture/cmd/kubecis" "$fixture/benchmarks/cis-eks-2.1.0" "$fixture/tools"
cp "$root/scripts/collect-eks.sh" "$fixture/scripts/collect-eks.sh"
cp "$root/go.mod" "$fixture/go.mod"
touch "$fixture/benchmarks/cis-eks-2.1.0/benchmark.yaml"
cat > "$fixture/tools/go" <<'EOF'
#!/usr/bin/env bash
[[ "$1" == build ]] || exit 5
while (($#)); do
  if [[ "$1" == -o ]]; then
    shift
    printf '#!/bin/sh\necho fixture-version\n' > "$1"
    chmod 755 "$1"
    exit 0
  fi
  shift
done
exit 5
EOF
chmod +x "$fixture/tools/go"
for binary_arg in default source; do
  rm -f "$fixture/kubecis"
  args=(--context dummy --cluster-name sample --region us-east-1)
  if [[ "$binary_arg" == source ]]; then
    args+=(--binary "$fixture/cmd/kubecis")
  fi
  # The fixture must reach the missing kubectl prerequisite, not attempt to
  # execute the source directory or fail finding the local binary.
  if PATH="$fixture/tools:$PATH" bash "$fixture/scripts/collect-eks.sh" "${args[@]}" > "$fixture/log" 2>&1; then
    echo 'FAIL: fixture reached a successful live scan unexpectedly' >&2; exit 1
  fi
  [[ -x "$fixture/kubecis" ]] || { cat "$fixture/log" >&2; echo 'FAIL: source build was not created' >&2; exit 1; }
  if grep -q 'Is a directory\|CLI version unavailable\|Missing executable' "$fixture/log"; then
    cat "$fixture/log" >&2
    echo 'FAIL: source directory/binary resolution regressed' >&2
    exit 1
  fi
done
echo 'PASS: CLI missing/source-directory arguments build a runnable executable (mocked Go)'

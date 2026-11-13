#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
for name in collect-eks-provider.sh collect-eks.sh validate-eks-preview.sh deploy-eks.sh; do
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


#!/usr/bin/env bash
# Fixture-only AWS provider parser test: real AWS IAM/EKS acceptance is separate.
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/aws" <<'EOF'
#!/bin/sh
test "$1" = eks && test "$2" = describe-cluster || exit 12
cat "$MOCK_EKS_JSON"
EOF
chmod +x "$tmp/bin/aws"
export PATH="$tmp/bin:$PATH"
export MOCK_EKS_JSON="$tmp/cluster.json"
jq -n '
 {cluster:{name:"test-eks",status:"ACTIVE",
   arn:"arn:aws:eks:us-east-1:123456789012:cluster/test-eks",
   endpoint:"https://example.gr7.us-east-1.eks.amazonaws.com",
   version:"1.35",
   resourcesVpcConfig:{endpointPrivateAccess:true,endpointPublicAccess:false,
                       publicAccessCidrs:["10.0.0.0/8"]},
   accessConfig:{authenticationMode:"API"},
   logging:{clusterLogging:[
     {types:["api"],enabled:true},{types:["audit"],enabled:true},
     {types:["authenticator"],enabled:false},
     {types:["controllerManager"],enabled:true},
     {types:["scheduler"],enabled:true}
   ]}}}
' > "$MOCK_EKS_JSON"
collect() {
 bash "$root/scripts/collect-eks-provider.sh" \
   --cluster-name test-eks --region us-east-1 \
   --expected-server https://example.gr7.us-east-1.eks.amazonaws.com \
   --output-file "$tmp/provider.json"
}
collect >/dev/null
jq -e '
  .schemaVersion == 1 and .assessmentStatus == "INVENTORY_ONLY" and
  .clusterName == "test-eks" and .region == "us-east-1" and
  .controlPlaneLogging.authenticator == false and
  .accessInventoryStatus == "UNAVAILABLE" and
  .ecrInventoryStatus == "UNAVAILABLE" and
  .vpcCniStatus == "UNAVAILABLE" and
  .endpointPublicAccess == false
' "$tmp/provider.json" >/dev/null
jq '.cluster.endpoint = "https://attacker.invalid"' "$MOCK_EKS_JSON" > "$tmp/bad.json"
cp "$tmp/bad.json" "$MOCK_EKS_JSON"
if collect >/dev/null 2>&1; then
  echo 'FAIL: AWS EKS endpoint mismatch was accepted' >&2
  exit 1
fi
echo 'PASS: provider JSON maps AWS evidence and rejects wrong cluster identity'

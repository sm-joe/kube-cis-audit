#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/collect-eks-access-cni.sh"
tmp="$(mktemp -d)"; trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/aws" <<'MOCK'
#!/usr/bin/env bash
case "$*" in
 *"list-access-entries"*)
  if [[ "${MOCK_ACCESS_DENY:-}" == true ]]; then echo 'AccessDenied' >&2;exit 1;fi
  echo '{"accessEntries":["arn:aws:iam::123456789012:role/eks-admin"]}' ;;
 *"describe-access-entry"*)
  echo '{"accessEntry":{"principalArn":"arn:aws:iam::123456789012:role/eks-admin","type":"STANDARD","kubernetesGroups":["admins"]}}' ;;
 *"list-associated-access-policies"*)
  echo '{"associatedAccessPolicies":[{"policyArn":"arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy","accessScope":{"type":"namespace","namespaces":["apps"]}}]}' ;;
 *"describe-addon"*)
  if [[ "${MOCK_CNI_UNKNOWN:-}" == true ]];then
   echo '{"addon":{"status":"ACTIVE","addonVersion":"v1.20.0","configurationValues":"enableNetworkPolicy: true"}}'
  else
   echo '{"addon":{"status":"ACTIVE","addonVersion":"v1.20.0","configurationValues":"{\"enableNetworkPolicy\":\"true\"}"}}'
  fi ;;
 *) exit 1 ;;
esac
MOCK
chmod +x "$tmp/bin/aws"
PATH="$tmp/bin:$PATH" bash "$root/scripts/collect-eks-access-cni.sh" \
 --cluster-name demo-eks --region us-east-1 --authentication-mode API \
 --output-file "$tmp/evidence.json"
jq -e '
 .accessInventoryStatus=="COLLECTED" and (.accessEntries|length)==1 and
 .accessEntries[0].principalArn=="arn:aws:iam::123456789012:role/eks-admin" and
 .accessEntries[0].kubernetesGroups==["admins"] and
 .vpcCniStatus=="ACTIVE" and .vpcCniNetworkPolicy==true
' "$tmp/evidence.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_ACCESS_DENY=true bash "$root/scripts/collect-eks-access-cni.sh" \
 --cluster-name demo-eks --region us-east-1 --authentication-mode API \
 --output-file "$tmp/evidence.json"
jq -e '.accessInventoryStatus=="UNAVAILABLE" and (.accessEntries|length)==0' "$tmp/evidence.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CNI_UNKNOWN=true bash "$root/scripts/collect-eks-access-cni.sh" \
 --cluster-name demo-eks --region us-east-1 --authentication-mode API \
 --output-file "$tmp/evidence.json"
jq -e '.vpcCniStatus=="ACTIVE" and .vpcCniNetworkPolicy==null' "$tmp/evidence.json" >/dev/null
PATH="$tmp/bin:$PATH" bash "$root/scripts/collect-eks-access-cni.sh" \
 --cluster-name demo-eks --region us-east-1 --authentication-mode CONFIG_MAP \
 --output-file "$tmp/evidence.json"
jq -e '.accessInventoryStatus=="UNAVAILABLE"' "$tmp/evidence.json" >/dev/null
echo 'PASS: IAM access-entry policy scopes and VPC CNI settings fail closed when unknown/denied'

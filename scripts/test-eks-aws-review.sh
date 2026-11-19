#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/collect-eks-aws-review.sh"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/provider.json" <<'JSON'
{"schemaVersion":1,"clusterArn":"arn:aws:eks:us-east-1:123456789012:cluster/demo-eks","ecrImages":[{"registry":"123456789012.dkr.ecr.us-east-1.amazonaws.com","repository":"kubecis","digest":"sha256:aaaa"}]}
JSON
cat > "$tmp/bin/aws" <<'MOCK'
#!/usr/bin/env bash
if [[ "${MOCK_FAIL:-}" == all ]]; then echo 'AccessDeniedException' >&2;exit 1;fi
case "$1 $2" in
 "ecr get-repository-policy") echo 'RepositoryPolicyNotFoundException' >&2;exit 1;;
 "ecr get-registry-policy") echo 'RegistryPolicyNotFoundException' >&2;exit 1;;
 "eks list-nodegroups") echo '["workers"]';;
 "eks describe-nodegroup")
   if [[ "$*" == *nodegroup.subnets* ]];then echo '["subnet-123"]'
   else echo '{"nodeRole":"arn:aws:iam::123456789012:role/eks-node","subnets":["subnet-123"]}';fi;;
 "iam list-attached-role-policies") echo '["arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPullOnly"]';;
 "iam list-role-policies") echo '[]';;
 "eks describe-cluster") echo '[{"resources":["secrets"],"provider":{"keyArn":"arn:aws:kms:us-east-1:123456789012:key/abcd"}}]';;
 "kms describe-key") echo '{"KeyManager":"CUSTOMER","KeyState":"Enabled"}';;
 "ec2 describe-subnets") echo '{"SubnetId":"subnet-123","VpcId":"vpc-123","MapPublicIpOnLaunch":false}';;
 *) echo "unexpected AWS call $*" >&2;exit 1;;
esac
MOCK
chmod +x "$tmp/bin/aws"
PATH="$tmp/bin:$PATH" bash "$root/scripts/collect-eks-aws-review.sh" \
 --cluster-name demo-eks --region us-east-1 --evidence-file "$tmp/provider.json"
jq -e '
(.awsManualEvidence|keys|sort)==["5.1.2","5.1.3","5.3.1","5.4.3"] and
.awsManualEvidence["5.1.2"].items[0].repositories[0].repository=="kubecis" and
.awsManualEvidence["5.1.3"].items[0].nodeRole=="arn:aws:iam::123456789012:role/eks-node" and
.awsManualEvidence["5.3.1"].items[0].keyManager=="CUSTOMER" and
.awsManualEvidence["5.4.3"].items[0].MapPublicIpOnLaunch==false and
.awsManualEvidence["5.4.3"].status=="PARTIAL" and
all(.awsManualEvidence[];has("status") and has("note") and has("items"))
' "$tmp/provider.json" >/dev/null
MOCK_FAIL=all PATH="$tmp/bin:$PATH" bash "$root/scripts/collect-eks-aws-review.sh" \
 --cluster-name demo-eks --region us-east-1 --evidence-file "$tmp/provider.json"
jq -e '.awsManualEvidence["5.1.3"].status=="PARTIAL" and
.awsManualEvidence["5.3.1"].status=="PARTIAL" and
(.awsManualEvidence["5.4.3"].items|length)==0' "$tmp/provider.json" >/dev/null
echo 'PASS: AWS manual 4-control collector, failed IAM/AWS API states and no automatic compliance decisions'

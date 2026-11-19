#!/usr/bin/env bash
set -Eeuxo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/collect-eks-node-network.sh"
tmp="$(mktemp -d)";trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/provider.json" <<'JSON'
{"awsManualEvidence":{"5.4.3":{"status":"PARTIAL","note":"old","items":[{"SubnetId":"subnet-123","VpcId":"vpc-123","nodeGroup":"workers","MapPublicIpOnLaunch":false}]}}}
JSON
cat > "$tmp/bin/kubectl" <<'MOCK'
#!/usr/bin/env bash
if [[ "${TEST_UNAVAILABLE:-}" == kubectl ]];then exit 1;fi
echo '{"items":[{"metadata":{"name":"worker"},"spec":{"providerID":"aws:///us-east-1a/i-123"}}]}'
MOCK
cat > "$tmp/bin/aws" <<'MOCK'
#!/usr/bin/env bash
if [[ "${TEST_UNAVAILABLE:-}" == aws ]];then echo 'AccessDeniedException' >&2;exit 1;fi
case "$1 $2" in
"ec2 describe-route-tables")
 if [[ "$*" == *"Name=association.subnet-id"* ]];then
  if [[ "${TEST_ROUTE:-}" == main ]];then echo '{"RouteTables":[]}'
  else echo '{"RouteTables":[{"RouteTableId":"rtb-123","Associations":[{"SubnetId":"subnet-123"}],"Routes":[{"DestinationCidrBlock":"0.0.0.0/0","GatewayId":"igw-123","State":"active"}]}]}';fi
 else
  echo '{"RouteTables":[{"RouteTableId":"rtb-main","Associations":[{"Main":true}],"Routes":[{"DestinationIpv6CidrBlock":"::/0","GatewayId":"igw-456","State":"active"}]}]}'
 fi;;
"ec2 describe-instances")
 echo '[{"InstanceId":"i-123","SubnetId":"subnet-123","PublicIpAddress":null,"State":"running"}]';;
*) echo "Unexpected AWS call $*" >&2;exit 1;;
esac
MOCK
chmod +x "$tmp/bin/"*
run(){ PATH="$tmp/bin:$PATH" bash "$root/scripts/collect-eks-node-network.sh" --evidence-file "$tmp/provider.json" --region us-east-1 --context demo; }
run
jq -e '.awsManualEvidence["5.4.3"] | .status=="COLLECTED" and .nodeInventoryStatus=="COLLECTED" and .items[0].routeSource=="EXPLICIT" and .items[0].routes[0].gateway=="igw-123" and .nodes[0].InstanceId=="i-123"' "$tmp/provider.json" >/dev/null
TEST_ROUTE=main run
jq -e '.awsManualEvidence["5.4.3"] | .status=="COLLECTED" and .items[0].routeSource=="MAIN" and .items[0].routes[0].ipv6=="::/0" and .items[0].routes[0].gateway=="igw-456"' "$tmp/provider.json" >/dev/null
TEST_UNAVAILABLE=aws run
jq -e '.awsManualEvidence["5.4.3"] | .status=="PARTIAL" and .items[0].routeSource=="UNAVAILABLE"' "$tmp/provider.json" >/dev/null
TEST_UNAVAILABLE=kubectl run
jq -e '.awsManualEvidence["5.4.3"] | .status=="PARTIAL" and .nodeInventoryStatus=="UNAVAILABLE"' "$tmp/provider.json" >/dev/null
echo 'PASS: CIS EKS 5.4.3 explicit IGW, VPC main IPv6 IGW, EC2 node, denied AWS API and incomplete kubectl evidence'

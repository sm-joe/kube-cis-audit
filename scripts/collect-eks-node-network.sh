#!/usr/bin/env bash
# CIS EKS 2.1.0 5.4.3: read-only node subnet / actual EC2 / effective routing.
# Private-node PASS requires full cluster node inventory; failures can be
# confirmed from any verified node-group subnet or observed public EC2 node.
set -Eeuo pipefail
die(){ echo "EKS node network: $*" >&2;exit 1; }
evidence='' region='' context=''
while (($#)); do
 case "$1" in
 --evidence-file|--region|--context)
  (($#>=2)) || die "Missing value for $1"
  case "$1" in
   --evidence-file) evidence=$2;; --region) region=$2;; --context) context=$2;;
  esac;shift 2;;
 *) die "Unknown argument $1";;
 esac
done
[[ -s "$evidence" && -n "$region" ]] || die 'Require provider evidence file and region'
for app in aws jq; do command -v "$app" >/dev/null || die "$app is required"; done
umask 077
tmp="$(mktemp -d)"; trap 'rm -rf -- "$tmp"' EXIT
records="$(jq -c '.awsManualEvidence["5.4.3"].items // []' "$evidence")"
count="$(jq 'length' <<<"$records")"
(( count <= 40 )) || die 'Too many nodegroup subnets'
errors=0
enriched='[]'
for ((i=0;i<count;i++)); do
 item="$(jq -c --argjson i "$i" '.[$i]' <<<"$records")"
 subnet="$(jq -r '.SubnetId // empty' <<<"$item")"
 vpc="$(jq -r '.VpcId // empty' <<<"$item")"
 routes='[]'; route_table=''; route_source='UNAVAILABLE'
 if [[ "$subnet" =~ ^subnet-[a-f0-9]+$ && "$vpc" =~ ^vpc-[a-f0-9]+$ ]]; then
  if tables="$(aws ec2 describe-route-tables --region "$region" --filters "Name=association.subnet-id,Values=$subnet" --output json 2>/dev/null)" &&
    jq -e '.RouteTables | type == "array"' <<<"$tables" >/dev/null 2>&1;then
   candidate="$(jq -c --arg subnet "$subnet" '[.RouteTables[] | select(any(.Associations[]?; .SubnetId == $subnet))]' <<<"$tables")"
   if [[ "$(jq 'length' <<<"$candidate")" == 1 ]]; then
    selected="$(jq -c '.[0]' <<<"$candidate")";route_source='EXPLICIT'
   elif [[ "$(jq 'length' <<<"$candidate")" == 0 ]];then
    # AWS uses VPC main route table when a subnet has no explicit association.
    if main="$(aws ec2 describe-route-tables --region "$region" --filters "Name=vpc-id,Values=$vpc" "Name=association.main,Values=true" --output json 2>/dev/null)" &&
      jq -e '.RouteTables | type == "array" and length == 1' <<<"$main" >/dev/null 2>&1;then
     selected="$(jq -c '.RouteTables[0]' <<<"$main")";route_source='MAIN'
    fi
   fi
  fi
 fi
 if [[ "$route_source" != UNAVAILABLE ]];then
  route_table="$(jq -r '.RouteTableId // empty' <<<"$selected")"
  if [[ ! "$route_table" =~ ^rtb-[a-f0-9]+$ ]];then
   route_source='UNAVAILABLE'
  else
   routes="$(jq -c '[.Routes[]? | {
    ipv4:(.DestinationCidrBlock // ""),ipv6:(.DestinationIpv6CidrBlock // ""),
    gateway:(.GatewayId // ""),state:(.State // "")
   }]' <<<"$selected")"
  fi
 fi
 [[ "$route_source" != UNAVAILABLE ]] || ((errors+=1))
 enriched="$(jq -c --argjson item "$item" --argjson routes "$routes" \
  --arg table "$route_table" --arg source "$route_source" \
  '.+[$item+{routeTableId:$table,routeSource:$source,routes:$routes}]' <<<"$enriched")"
done
# The Kubernetes Node inventory is the authoritative set of currently
# registered workers, including self-managed instances not in managed groups.
nodes='[]';node_status='UNAVAILABLE'
if [[ -n "$context" ]] && command -v kubectl >/dev/null;then
 if raw="$(kubectl --context "$context" get nodes -o json 2>/dev/null)" &&
    jq -e '.items | type=="array" and length>0 and length<=200' <<<"$raw" >/dev/null 2>&1;then
  node_status='COLLECTED'
  while IFS=$'\t' read -r node provider;do
   [[ -n "$node" ]] || continue
   if [[ "$provider" =~ /i-[a-f0-9]+$ ]];then
    instance_id="${BASH_REMATCH[0]}"
    instance_id="${instance_id#/}"
    if instance="$(aws ec2 describe-instances --instance-ids "$instance_id" --region "$region" \
      --query 'Reservations[].Instances[].{InstanceId:InstanceId,SubnetId:SubnetId,PublicIpAddress:PublicIpAddress,State:State.Name}' --output json 2>/dev/null)" &&
      jq -e --arg id "$instance_id" 'type=="array" and length==1 and .[0].InstanceId==$id and (.[0].SubnetId|type)=="string"' <<<"$instance" >/dev/null 2>&1;then
     entry="$(jq -c --arg node "$node" '.[0]+{node:$node}' <<<"$instance")"
     nodes="$(jq -c --argjson x "$entry" '.+[$x]' <<<"$nodes")"
    else node_status='PARTIAL';((errors+=1));fi
   else
    # Hybrid, Fargate, and unknown provider IDs cannot justify an EC2 PASS.
    node_status='PARTIAL';((errors+=1))
   fi
  done < <(jq -r '.items[] | [.metadata.name, (.spec.providerID // "")] | @tsv' <<<"$raw")
 fi
fi
if [[ "$node_status" == COLLECTED ]];then
 # Every observed EC2 node must be associated with an inspected subnet.
 if ! jq -e --argjson items "$enriched" 'all(.[]; .SubnetId as $id | any($items[]; .SubnetId==$id))' <<<"$nodes" >/dev/null 2>&1;then
  node_status='PARTIAL';((errors+=1))
 fi
fi
status='COLLECTED'
if ((errors>0||count==0)) || [[ "$node_status" != COLLECTED ]];then status='PARTIAL';fi
updated="$(jq -c --argjson items "$enriched" --argjson nodes "$nodes" \
 --arg status "$status" --arg nodeStatus "$node_status" \
 '.awsManualEvidence["5.4.3"] |= (.items=$items | .status=$status | .nodeInventoryStatus=$nodeStatus | .nodes=$nodes | .note="Inspect verified effective subnet routes and actual worker EC2 public IPs. Private PASS requires complete node inventory, no IGW default route, and public IP assignment disabled; exceptions remain separate.")' "$evidence")"
printf '%s\n' "$updated" > "$tmp/merged.json"
mv -- "$tmp/merged.json" "$evidence"

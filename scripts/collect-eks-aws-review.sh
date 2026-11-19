#!/usr/bin/env bash
# Read-only reviewer inputs for four CIS EKS controls. NEVER emits PASS.
set -Eeuo pipefail
die(){ echo "AWS review: $*" >&2; exit 1; }
cluster='' region='' output=''
while (($#)); do
  case "$1" in
    --cluster-name|--region|--evidence-file)
      (($#>=2)) || die "Missing $1"
      case "$1" in --cluster-name) cluster=$2;; --region) region=$2;; --evidence-file) output=$2;; esac
      shift 2;;
    *) die "Unknown option $1";;
  esac
done
[[ -n "$cluster" && -n "$region" && -s "$output" ]] || die 'Cluster, region and provider evidence required'
for cmd in jq aws sha256sum; do command -v "$cmd" >/dev/null || die "$cmd required"; done
account="$(jq -r '.clusterArn | split(":")[4] // empty' "$output")"
[[ "$account" =~ ^[0-9]{12}$ ]] || die 'AWS account could not be bound to cluster ARN'
umask 077
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
review='{}'
put(){
 review="$(jq -c --arg id "$1" --arg status "$2" --arg note "$3" --argjson items "$4" \
   '.[$id]={status:$status,note:$note,items:$items}' <<<"$review")"
}
# 5.1.2: repo/registry policy presence, with policy hashes only.
repos='[]';errors=0;count=0
while IFS= read -r repository; do
 [[ -n "$repository" ]] || continue
 ((count+=1))
 if ((count>40));then ((errors+=1));break;fi
 policy_status='UNAVAILABLE';hash=''
 if p="$(aws ecr get-repository-policy --registry-id "$account" --repository-name "$repository" --region "$region" --query policyText --output text 2>"$tmp/err")";then
   policy_status='PRESENT';hash="$(printf '%s' "$p" | sha256sum | cut -d' ' -f1)"
 elif grep -q RepositoryPolicyNotFoundException "$tmp/err";then
   policy_status='NONE'
 else ((errors+=1));fi
 repos="$(jq -c --arg repo "$repository" --arg status "$policy_status" --arg sha "$hash" \
  '.+[{repository:$repo,policyStatus:$status,policySha256:$sha}]' <<<"$repos")"
done < <(jq -r --arg prefix "$account.dkr.ecr.$region.amazonaws.com" '.ecrImages[]? | select(.registry==$prefix) | .repository' "$output" | sort -u)
registry_status='UNAVAILABLE';registry_hash=''
if p="$(aws ecr get-registry-policy --region "$region" --query policyText --output text 2>"$tmp/err")";then
 registry_status='PRESENT';registry_hash="$(printf '%s' "$p" | sha256sum | cut -d' ' -f1)"
elif grep -q RegistryPolicyNotFoundException "$tmp/err";then registry_status='NONE'
else ((errors+=1));fi
items="$(jq -n --argjson repos "$repos" --arg s "$registry_status" --arg h "$registry_hash" \
 '[{type:"ECR",repositories:$repos,registryPolicyStatus:$s,registryPolicySha256:$h}]')"
status='COLLECTED';((errors==0))||status='PARTIAL'
put '5.1.2' "$status" 'Review actual IAM, registry and repository policy scopes; a missing resource policy is not failure.' "$items"
# 5.1.3: managed node roles and attached/inline policy identities.
groups='[]';roles='[]';role_errors=0
if g="$(aws eks list-nodegroups --cluster-name "$cluster" --region "$region" --query nodegroups --output json 2>/dev/null)" &&
 jq -e 'type=="array"' <<<"$g" >/dev/null;then groups="$g";else ((role_errors+=1));fi
idx=0
while IFS= read -r group;do
 [[ -n "$group" ]] || continue
 ((idx+=1))
 if ((idx>40));then ((role_errors+=1));break;fi
 if d="$(aws eks describe-nodegroup --cluster-name "$cluster" --nodegroup-name "$group" --region "$region" --query 'nodegroup.{nodeRole:nodeRole,subnets:subnets}' --output json 2>/dev/null)" &&
 jq -e '(.nodeRole|type)=="string"' <<<"$d" >/dev/null;then
  arn="$(jq -r '.nodeRole' <<<"$d")"; name="${arn##*/}"
  attached='[]';inline='[]';pstatus='COLLECTED'
  if p="$(aws iam list-attached-role-policies --role-name "$name" --query 'AttachedPolicies[].PolicyArn' --output json 2>/dev/null)" &&
    jq -e 'type=="array"' <<<"$p" >/dev/null;then attached="$p";else pstatus='UNAVAILABLE';((role_errors+=1));fi
  if p="$(aws iam list-role-policies --role-name "$name" --query PolicyNames --output json 2>/dev/null)" &&
    jq -e 'type=="array"' <<<"$p" >/dev/null;then inline="$p";else pstatus='UNAVAILABLE';((role_errors+=1));fi
  roles="$(jq -c --arg ng "$group" --arg arn "$arn" --arg s "$pstatus" --argjson a "$attached" --argjson i "$inline" \
    '.+[{nodeGroup:$ng,nodeRole:$arn,policyInventoryStatus:$s,managedPolicies:$a,inlinePolicyNames:$i}]' <<<"$roles")"
 else ((role_errors+=1));fi
done < <(jq -r '.[]' <<<"$groups")
status='COLLECTED';((role_errors==0))||status='PARTIAL'
put '5.1.3' "$status" 'Reviewer must inspect IAM policy documents, self-managed nodes and Fargate pull roles; policy names alone are insufficient.' "$roles"
# 5.3.1: EKS encryption association and KMS key metadata.
keys='[]';key_errors=0
if d="$(aws eks describe-cluster --name "$cluster" --region "$region" --query cluster.encryptionConfig --output json 2>/dev/null)" &&
 jq -e 'type=="array" or .==null' <<<"$d" >/dev/null;then
 while IFS= read -r key;do
  [[ -n "$key" ]] || continue
  state='UNAVAILABLE';manager='UNKNOWN'
  if m="$(aws kms describe-key --key-id "$key" --region "$region" --query 'KeyMetadata.{KeyState:KeyState,KeyManager:KeyManager}' --output json 2>/dev/null)" &&
   jq -e '(.KeyManager|type)=="string"' <<<"$m" >/dev/null;then
   state="$(jq -r '.KeyState' <<<"$m")";manager="$(jq -r '.KeyManager' <<<"$m")"
  else ((key_errors+=1));fi
  keys="$(jq -c --arg k "$key" --arg s "$state" --arg m "$manager" \
    '.+[{keyArn:$k,keyState:$s,keyManager:$m}]' <<<"$keys")"
 done < <(jq -r '.[]? | select(.resources | index("secrets")) | .provider.keyArn // empty' <<<"$d")
else ((key_errors+=1));fi
status='COLLECTED';((key_errors==0))||status='PARTIAL'
put '5.3.1' "$status" 'Reviewer must confirm customer-managed KMS key ownership, enabled grants, key policy and rotation.' "$keys"
# 5.4.3: nodegroup subnet inventory; public-IP defaults are not sufficient.
subnets='[]';net_errors=0
while IFS= read -r group;do
 [[ -n "$group" ]] || continue
 if d="$(aws eks describe-nodegroup --cluster-name "$cluster" --nodegroup-name "$group" --region "$region" --query nodegroup.subnets --output json 2>/dev/null)" &&
 jq -e 'type=="array"' <<<"$d" >/dev/null;then
  while IFS= read -r subnet;do
   [[ -n "$subnet" ]] || continue
   if x="$(aws ec2 describe-subnets --subnet-ids "$subnet" --region "$region" --query 'Subnets[0].{SubnetId:SubnetId,MapPublicIpOnLaunch:MapPublicIpOnLaunch,VpcId:VpcId}' --output json 2>/dev/null)" &&
     jq -e '.SubnetId!=null and (.MapPublicIpOnLaunch|type)=="boolean"' <<<"$x" >/dev/null;then
    subnets="$(jq -c --arg group "$group" --argjson x "$x" '.+[$x+{nodeGroup:$group}]' <<<"$subnets")"
   else ((net_errors+=1));fi
  done < <(jq -r '.[]' <<<"$d")
 else ((net_errors+=1));fi
done < <(jq -r '.[]' <<<"$groups")
put '5.4.3' 'PARTIAL' 'Reviewer must inspect effective route tables, public IP addresses, and unmanaged/Fargate node placements; subnet defaults do not prove private nodes.' "$subnets"
merge="$(mktemp)"
trap 'rm -rf -- "$tmp" "$merge"' EXIT
jq -n --argjson v "$review" '{awsManualEvidence:$v}' > "$tmp/new.json"
jq -s '.[0]*.[1]' "$output" "$tmp/new.json" > "$merge" || die 'Cannot merge manual AWS inventory'
mv -- "$merge" "$output"

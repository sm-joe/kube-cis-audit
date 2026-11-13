#!/usr/bin/env bash
# Amazon EKS provider evidence. Read-only and deliberately conservative:
# optional AWS inventories remain UNAVAILABLE rather than becoming PASS.
set -Eeuo pipefail
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
usage() {
  echo 'Usage: bash scripts/collect-eks-provider.sh --cluster-name NAME --region REGION --expected-server HTTPS_URL --output-file PATH' 
}
cluster='' region='' endpoint='' outfile=''
while (($#)); do
  case "$1" in
    --cluster-name|--region|--expected-server|--output-file)
      (($# >= 2)) || die "Missing value for $1"
      case "$1" in
        --cluster-name) cluster=$2 ;;
        --region) region=$2 ;;
        --expected-server) endpoint=$2 ;;
        --output-file) outfile=$2 ;;
      esac
      shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown argument: $1" ;;
  esac
done
[[ -n "$cluster" && -n "$region" && -n "$endpoint" && -n "$outfile" ]] || { usage >&2; exit 2; }
command -v aws >/dev/null || die 'AWS CLI v2 required'
command -v jq >/dev/null || die 'jq required'
[[ "$endpoint" == https://* ]] || die 'Expected HTTPS EKS API server'
raw="$(aws eks describe-cluster --name "$cluster" --region "$region" --output json)" || die 'AWS DescribeCluster failed'
# AWS ARN, cluster version, five log types and HTTPS API identity must be
# present. Compare URL normalized only for a trailing slash; never trust a
# context alias as an AWS cluster identity.
jq -e --arg name "$cluster" --arg region "$region" --arg server "${endpoint%/}" '
  .cluster as $c |
  ($c.name == $name and $c.status == "ACTIVE"
   and ($c.endpoint | rtrimstr("/")) == $server
   and ($c.endpoint | startswith("https://"))
   and ($c.arn | test("^arn:aws[a-z-]*:eks:" + $region + ":[0-9]{12}:cluster/" + $name + "$"))
   and ($c.version | test("^[0-9]+[.][0-9]+$"))
   and ($c.resourcesVpcConfig.endpointPrivateAccess | type) == "boolean"
   and ($c.resourcesVpcConfig.endpointPublicAccess | type) == "boolean"
   and ([ "api", "audit", "authenticator", "controllerManager", "scheduler" ] |
       all(. as $t | ([$c.logging.clusterLogging[]? | select(.types | index($t)) | .enabled] | length) == 1))
  ' <<<"$raw" >/dev/null || die 'AWS EKS identity or mandatory provider fields invalid'
# jq never persists credentials. Other API evidence requires independent
# permission-scoped collection; mark it explicitly unavailable for this
# first POSIX collector rather than claiming coverage.
mkdir -p "$(dirname -- "$outfile")"
jq -n --argjson aws "$raw" --arg region "$region" --arg now "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '
  $aws.cluster as $c |
  {schemaVersion:1,cloudProvider:"eks",clusterName:$c.name,clusterArn:$c.arn,
   region:$region,kubernetesVersion:$c.version,generatedAt:$now,
   endpointPrivateAccess:$c.resourcesVpcConfig.endpointPrivateAccess,
   endpointPublicAccess:$c.resourcesVpcConfig.endpointPublicAccess,
   publicAccessCidrs:($c.resourcesVpcConfig.publicAccessCidrs // []),
   controlPlaneLogging:(reduce ["api","audit","authenticator","controllerManager","scheduler"][] as $t
     ({}; .[$t] = ([ $c.logging.clusterLogging[] | select(.types | index($t)) | .enabled ][0]))),
   assessmentStatus:"INVENTORY_ONLY",
   assessmentNote:"Read-only AWS inventory; optional IAM/CNI/ECR evidence not collected by POSIX preview; no official compliance score",
   authenticationMode:($c.accessConfig.authenticationMode // ""),
   accessInventoryStatus:"UNAVAILABLE",accessEntries:[],
   vpcCniStatus:"UNAVAILABLE",ecrInventoryStatus:"UNAVAILABLE",ecrImages:[],nonEcrImageCount:0
  }' > "$outfile" || die 'Cannot write provider evidence'
jq -e '.schemaVersion == 1 and .assessmentStatus == "INVENTORY_ONLY"' "$outfile" >/dev/null || die 'Invalid generated evidence'
echo "Collected provisional EKS provider inventory for $cluster ($region)"

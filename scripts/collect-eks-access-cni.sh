#!/usr/bin/env bash
# Bounded read-only EKS access entries and managed VPC CNI configuration.
# Never print API errors containing account internals or treat denied calls as PASS.
set -Eeuo pipefail
die(){ echo "ERROR: $*" >&2; exit 1; }
cluster='' region='' mode='' outfile=''
while (($#)); do
 case "$1" in
 --cluster-name|--region|--authentication-mode|--output-file)
  (($#>=2)) || die "Missing $1"
  case "$1" in
   --cluster-name) cluster=$2;; --region) region=$2;;
   --authentication-mode) mode=$2;; --output-file) outfile=$2;;
  esac
  shift 2;;
 *) die "Unknown option $1";;
 esac
done
[[ -n "$cluster" && -n "$region" && -n "$mode" && -n "$outfile" ]] || die 'Cluster, region, mode and output required'
for tool in aws jq; do command -v "$tool" >/dev/null || die "$tool required"; done
umask 077
access_status=UNAVAILABLE
entries='[]'
if [[ "$mode" == API || "$mode" == API_AND_CONFIG_MAP ]]; then
 if raw="$(aws eks list-access-entries --cluster-name "$cluster" --region "$region" --output json 2>/dev/null)" &&
    jq -e '(.accessEntries | type) == "array" and (.accessEntries | length) <= 60' <<<"$raw" >/dev/null; then
  access_status=COLLECTED
  while IFS= read -r principal; do
   if [[ ! "$principal" == arn:aws* ]]; then access_status=UNAVAILABLE;break;fi
   if ! item="$(aws eks describe-access-entry --cluster-name "$cluster" --principal-arn "$principal" --region "$region" --output json 2>/dev/null)" ||
      ! policies="$(aws eks list-associated-access-policies --cluster-name "$cluster" --principal-arn "$principal" --region "$region" --output json 2>/dev/null)" ||
      ! jq -e --arg principal "$principal" '.accessEntry.principalArn == $principal and (.accessEntry.kubernetesGroups // [] | type) == "array"' <<<"$item" >/dev/null ||
      ! jq -e '(.associatedAccessPolicies | type) == "array" and (.associatedAccessPolicies | length) <= 60' <<<"$policies" >/dev/null; then
    access_status=UNAVAILABLE
    break
   fi
   type="$(jq -r '.accessEntry.type // "UNKNOWN"' <<<"$item")"
   groups="$(jq -c '.accessEntry.kubernetesGroups // []' <<<"$item")"
   scopes="$(jq -c '[.associatedAccessPolicies[] | (.policyArn // "UNKNOWN") + "@" + (.accessScope.type // "UNKNOWN") + ":" + ((.accessScope.namespaces // []) | join(","))]' <<<"$policies")"
   entries="$(jq -c --arg principal "$principal" --arg type "$type" --argjson groups "$groups" --argjson scopes "$scopes" '. + [{principalArn:$principal,type:$type,kubernetesGroups:$groups,policyScopes:$scopes}]' <<<"$entries")"
  done < <(jq -r '.accessEntries[]' <<<"$raw")
 fi
fi
if [[ "$access_status" != COLLECTED ]]; then entries='[]';fi
cni_status=UNAVAILABLE
cni_version=''
cni_policy='null'
if addon="$(aws eks describe-addon --cluster-name "$cluster" --addon-name vpc-cni --region "$region" --output json 2>/dev/null)" &&
   jq -e '(.addon | type) == "object" and (.addon.status | type) == "string"' <<<"$addon" >/dev/null; then
 cni_status="$(jq -r '.addon.status' <<<"$addon")"
 cni_version="$(jq -r '.addon.addonVersion // ""' <<<"$addon")"
 config="$(jq -r '.addon.configurationValues // ""' <<<"$addon")"
 cni_policy="$(jq -n --arg cv "$config" '
   (try ($cv | fromjson) catch null) as $parsed |
   ($parsed.enableNetworkPolicy // null) as $value |
   if $value == true or $value == "true" then true
   elif $value == false or $value == "false" then false
   else null end
 ')"
fi
tmp="$(mktemp)"
trap 'rm -f -- "$tmp"' EXIT
jq -n --arg access "$access_status" --argjson entries "$entries" \
 --arg cniStatus "$cni_status" --arg cniVersion "$cni_version" \
 --argjson cniPolicy "$cni_policy" \
 '{accessInventoryStatus:$access,accessEntries:$entries,vpcCniStatus:$cniStatus,vpcCniVersion:$cniVersion,vpcCniNetworkPolicy:$cniPolicy}' > "$tmp"
mkdir -p "$(dirname -- "$outfile")"
mv -- "$tmp" "$outfile"

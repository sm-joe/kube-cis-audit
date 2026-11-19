#!/usr/bin/env bash
# Bounded, read-only ECR inventory of deployed Pod container images.
set -Eeuo pipefail
die(){ echo "ERROR: $*" >&2; exit 1; }
context='' region='' outfile=''
while (($#)); do
 case "$1" in
  --context|--region|--output-file)
   (($#>=2)) || die "Missing $1"
   case "$1" in --context) context=$2;; --region) region=$2;; --output-file) outfile=$2;; esac
   shift 2;;
  -h|--help) echo 'Usage: collect-eks-ecr.sh --context CTX --region REGION --output-file FILE';exit 0;;
  *) die "Unknown argument $1";;
 esac
done
[[ -n "$context" && -n "$region" && -n "$outfile" ]] || die 'Context, region, output required'
for tool in jq kubectl aws; do command -v "$tool" >/dev/null || die "$tool required"; done
umask 077
work="$(mktemp -d)"
trap 'rm -rf -- "$work"' EXIT
status=UNAVAILABLE; non_ecr=0; images='[]'; count=0
declare -A seen=() non_ecr_seen=()
if kubectl --context "$context" get pods --all-namespaces -o json > "$work/pods.json" &&
   jq -e '(.items | type) == "array"' "$work/pods.json" >/dev/null; then
 jq -r '
   .items[] | select(.status.phase != "Succeeded" and .status.phase != "Failed") as $p |
   (([$p.spec.containers[]? | {name,image,group:"containerStatuses"}] +
     [$p.spec.initContainers[]? | {name,image,group:"initContainerStatuses"}] +
     [$p.spec.ephemeralContainers[]? | {name,image,group:"ephemeralContainerStatuses"}])[]) as $c |
   ([($p.status[$c.group] // [])[] | select(.name == $c.name) | .imageID][0] // "") as $imageID |
   [$c.image, $imageID] | @tsv
 ' "$work/pods.json" > "$work/rows"
 status=COLLECTED
 while IFS=$'\t' read -r image image_id; do
  [[ -n "$image" ]] || { status=UNAVAILABLE; break; }
  if [[ ! "$image" =~ ^([0-9]{12})[.]dkr[.]ecr[.]([a-z0-9-]+)[.]amazonaws[.]com([.]cn)?/(.+)$ ]]; then
   non_ecr_seen["$image"]=1;continue
  fi
  account="${BASH_REMATCH[1]}"; image_region="${BASH_REMATCH[2]}"
  registry="${image%%/*}"; path="${BASH_REMATCH[4]}"
  repository="${path%%@*}"; repository="${repository%%:*}"
  [[ "$repository" =~ ^[a-z0-9][a-z0-9._/-]*$ ]] || { status=UNAVAILABLE; break; }
  digest=''
  if [[ "$path" =~ @sha256:([a-fA-F0-9]{64})$ ]]; then
   digest="sha256:${BASH_REMATCH[1],,}"
  elif [[ "$image_id" =~ sha256:([a-fA-F0-9]{64}) ]]; then
   digest="sha256:${BASH_REMATCH[1],,}"
  fi
  key="$account/$image_region/$repository/$digest"
  [[ -z "${seen[$key]+x}" ]] || continue
  seen["$key"]=1
  ((count+=1))
  if ((count>80)); then status=UNAVAILABLE;break;fi
  scan_type=''; scan_status=''; error=''
  if [[ -z "$digest" ]];then
   error='runtime image digest unavailable'
  else
   if config="$(aws ecr get-registry-scanning-configuration --registry-id "$account" --region "$image_region" --output json 2>/dev/null)" &&
      scan_type="$(jq -er '.scanningConfiguration.scanType | select(. == "BASIC" or . == "ENHANCED")' <<<"$config")"; then
    if detail="$(aws ecr describe-image-scan-findings --registry-id "$account" --repository-name "$repository" --image-id "imageDigest=$digest" --region "$image_region" --query '{imageId:imageId,scanStatus:imageScanStatus.status}' --output json 2>/dev/null)" &&
       jq -e --arg digest "$digest" '.imageId.imageDigest == $digest and (.scanStatus | type) == "string"' <<<"$detail" >/dev/null;then
     scan_status="$(jq -r '.scanStatus' <<<"$detail")"
    else
     error='scan status unavailable or image digest mismatch'
    fi
   else
    error='scan configuration unavailable'
   fi
  fi
  images="$(jq -c --arg registry "$registry" --arg repository "$repository" --arg digest "$digest" --arg scanType "$scan_type" --arg scanStatus "$scan_status" --arg error "$error" '. + [{registry:$registry,repository:$repository,digest:$digest,scanType:$scanType,scanStatus:$scanStatus,error:$error}]' <<<"$images")"
 done < "$work/rows"
 non_ecr="${#non_ecr_seen[@]}"
 [[ -s "$work/rows" ]] || status=UNAVAILABLE
else
 echo 'WARN: unable to obtain complete Kubernetes Pod image inventory' >&2
fi
if [[ "$status" != COLLECTED ]];then images='[]';non_ecr=0;fi
mkdir -p "$(dirname -- "$outfile")"
jq -n --arg status "$status" --argjson images "$images" --argjson nonECR "$non_ecr" '{ecrInventoryStatus:$status,ecrImages:$images,nonEcrImageCount:$nonECR}' > "$work/out.json"
mv -- "$work/out.json" "$outfile"
echo "ECR evidence: $status (ECR $count, non-ECR $non_ecr)" >&2

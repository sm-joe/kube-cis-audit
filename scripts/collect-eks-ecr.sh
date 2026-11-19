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
for tool in jq kubectl aws base64; do command -v "$tool" >/dev/null || die "$tool required"; done
# External evidence is opt-in and requires all three independent trust inputs.
attest_dir="${KUBECIS_ATTESTATION_DIR:-}"
attest_pub="${KUBECIS_ATTESTATION_PUBLIC_KEY:-}"
attest_pin="${KUBECIS_ATTESTATION_KEY_SHA256:-}"
if [[ -n "$attest_dir$attest_pub$attest_pin" && ( -z "$attest_dir" || -z "$attest_pub" || -z "$attest_pin" ) ]]; then
 die 'External attestations require directory, pinned public key and its SHA256 together'
fi
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
umask 077
work="$(mktemp -d)"
trap 'rm -rf -- "$work"' EXIT
status=UNAVAILABLE; non_ecr=0; images='[]'; count=0
# Caller identity is used only for cross-account diagnostics. Missing caller
# identity must not cause Pod inventory to be classified as absent.
caller_account="$(aws sts get-caller-identity --query Account --output text 2>/dev/null || true)"
[[ "$caller_account" =~ ^[0-9]{12}$ ]] || caller_account=''
classify_aws_error() {
 local output="$1"
 case "$output" in
  *AccessDenied*|*UnauthorizedOperation*|*"not authorized"*) echo 'AWS_READ_ACCESS_DENIED' ;;
  *ScanNotFoundException*) echo 'SCAN_NOT_FOUND' ;;
  *UnsupportedImageTypeException*) echo 'UNSUPPORTED_IMAGE_TYPE' ;;
  *ImageNotFoundException*) echo 'IMAGE_NOT_FOUND' ;;
  *RepositoryNotFoundException*) echo 'REPOSITORY_NOT_FOUND' ;;
  *Throttling*|*TooManyRequests*) echo 'AWS_THROTTLED' ;;
  *) echo 'AWS_API_UNAVAILABLE' ;;
 esac
}
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
  scan_type=''; scan_status=''; error=''; resolved_scan_digests='[]'
  if [[ -z "$digest" ]];then
   error='runtime image digest unavailable'
  else
   config_error=''
   # GetRegistryScanningConfiguration has NO --registry-id input option.
   # It describes only the caller's current regional registry. Never reuse
   # that response for images published in a different AWS account.
   if [[ -z "$caller_account" ]]; then
    config_error='CALLER_ACCOUNT_UNVERIFIED'
    error="REGISTRY_SCAN_CONFIG_$config_error"
   elif [[ "$account" != "$caller_account" ]]; then
    config_error='CROSS_ACCOUNT_REGISTRY_MODE_UNAVAILABLE'
    error="REGISTRY_SCAN_CONFIG_$config_error"
   elif config="$(aws ecr get-registry-scanning-configuration --region "$image_region" --output json 2>"$work/aws-error")"; then
    # Verify the returned registry owner as well as the scanning mode.
    if jq -e --arg account "$account" '.registryId == $account' <<<"$config" >/dev/null 2>&1; then
     scan_type="$(jq -er '.scanningConfiguration.scanType | select(. == "BASIC" or . == "ENHANCED")' <<<"$config" 2>/dev/null || true)"
     if [[ -z "$scan_type" ]]; then
      error='REGISTRY_SCAN_TYPE_UNKNOWN'
     fi
    else
     config_error='REGISTRY_ID_MISMATCH'
     error="REGISTRY_SCAN_CONFIG_$config_error"
    fi
   else
    config_error="$(classify_aws_error "$(cat "$work/aws-error")")"
    error="REGISTRY_SCAN_CONFIG_$config_error"
   fi
   # The scan API can be independently authorized even when the caller lacks
   # GetRegistryScanningConfiguration. Keep both observations separate.
   if detail="$(aws ecr describe-image-scan-findings --registry-id "$account" --repository-name "$repository" --image-id "imageDigest=$digest" --region "$image_region" --query '{imageId:imageId,scanStatus:imageScanStatus.status}' --output json 2>"$work/aws-error")"; then
    if jq -e --arg digest "$digest" '.imageId.imageDigest == $digest and (.scanStatus | type) == "string" and (.scanStatus | length) > 0' <<<"$detail" >/dev/null; then
     scan_status="$(jq -r '.scanStatus' <<<"$detail")"
     if [[ -z "$scan_type" ]]; then
      if [[ -n "$config_error" ]]; then
       error="REGISTRY_SCAN_CONFIG_$config_error"
      else
       error='REGISTRY_SCAN_TYPE_UNKNOWN'
      fi
     else
      error=''
     fi
    else
     error='SCAN_RESPONSE_DIGEST_OR_STATUS_UNVERIFIED'
    fi
   else
    scan_error="$(classify_aws_error "$(cat "$work/aws-error")")"
    error="SCAN_RESULT_$scan_error"
    if [[ -n "$caller_account" && "$account" != "$caller_account" ]]; then
     error="CROSS_ACCOUNT_ECR_$scan_error"
    fi
   fi
   # A deployed OCI image index (or Docker manifest list) is an immutable
   # pointer to platform-specific image manifests. ECR Basic does not scan
   # index digests. Resolve the exact index through ECR, then require every
   # runnable Linux child digest to have an independently verified scan.
   # Ignore OCI attestation descriptors with unknown/unknown platform.
   # Never infer a PASS from a child that cannot be read or matched.
   if [[ ( "$error" == 'SCAN_RESULT_SCAN_NOT_FOUND' || "$error" == 'SCAN_RESULT_UNSUPPORTED_IMAGE_TYPE' ) && "$account" == "$caller_account" ]]; then
    if manifest="$(aws ecr batch-get-image --registry-id "$account" --repository-name "$repository" --image-ids "imageDigest=$digest" --region "$image_region" --query 'images[0].imageManifest' --output text 2>"$work/aws-error")"; then
     if jq -e '(.mediaType == "application/vnd.oci.image.index.v1+json" or .mediaType == "application/vnd.docker.distribution.manifest.list.v2+json") and (.manifests | type == "array")' <<<"$manifest" >/dev/null 2>&1; then
      child_digests="$(jq -cr '[.manifests[] | select(.platform.os == "linux" and (.platform.architecture == "amd64" or .platform.architecture == "arm64")) | select(.mediaType == "application/vnd.oci.image.manifest.v1+json" or .mediaType == "application/vnd.docker.distribution.manifest.v2+json") | .digest] | unique' <<<"$manifest")"
      child_count="$(jq 'length' <<<"$child_digests")"
      if ((child_count == 0 || child_count > 8)); then
       error='OCI_INDEX_NO_BOUNDED_LINUX_CHILDREN'
      elif ! jq -e 'all(.[]; test("^sha256:[a-fA-F0-9]{64}$"))' <<<"$child_digests" >/dev/null; then
       error='OCI_INDEX_INVALID_CHILD_DIGEST'
      else
       error=''
       all_complete=true
       all_active=false
       while IFS= read -r child_digest; do
        if ! child_result="$(aws ecr describe-image-scan-findings --registry-id "$account" --repository-name "$repository" --image-id "imageDigest=$child_digest" --region "$image_region" --query '{imageId:imageId,scanStatus:imageScanStatus.status}' --output json 2>"$work/aws-error")"; then
         all_complete=false
         error="OCI_CHILD_$(classify_aws_error "$(cat "$work/aws-error")")"
         break
        fi
        if ! jq -e --arg digest "$child_digest" '.imageId.imageDigest == $digest and (.scanStatus == "COMPLETE" or .scanStatus == "ACTIVE")' <<<"$child_result" >/dev/null 2>&1; then
         all_complete=false
         error='OCI_CHILD_SCAN_NOT_VERIFIED'
         break
        fi
        [[ "$(jq -r '.scanStatus' <<<"$child_result")" == 'ACTIVE' ]] && all_active=true
        resolved_scan_digests="$(jq -c --arg digest "$child_digest" '. + [$digest]' <<<"$resolved_scan_digests")"
       done < <(jq -r '.[]' <<<"$child_digests")
       if [[ "$all_complete" == true && "$(jq 'length' <<<"$resolved_scan_digests")" -eq "$child_count" ]]; then
        if [[ "$all_active" == true ]]; then scan_status='ACTIVE'; else scan_status='COMPLETE'; fi
        if [[ -z "$scan_type" ]]; then
         if [[ -n "$config_error" ]]; then
          error="REGISTRY_SCAN_CONFIG_$config_error"
         else
          error='REGISTRY_SCAN_TYPE_UNKNOWN'
         fi
        fi
       else
        scan_status=''
       fi
      fi
     fi
    else
     error="OCI_INDEX_LOOKUP_$(classify_aws_error "$(cat "$work/aws-error")")"
    fi
   fi
  fi
  attestation_scanner='' attestation_key='' attestation_date=''
  attestation_claim_b64='' attestation_signature_b64=''
  if [[ -n "$attest_dir" && -n "$digest" && -n "$error" && -f "$attest_dir/${digest#sha256:}.json" ]]; then
   if bash "$script_dir/verify-eks-scan-attestation.sh" \
     --dir "$attest_dir" --public-key "$attest_pub" --key-sha256 "$attest_pin" \
     --registry "$registry" --repository "$repository" --digest "$digest" \
     --output-file "$work/verified-attestation.json" >/dev/null 2>"$work/attestation-error"; then
     scan_type='THIRD_PARTY'
     scan_status='COMPLETE'
     error=''
     resolved_scan_digests="$(jq -c '.scanDigests' "$work/verified-attestation.json")"
     attestation_scanner="$(jq -r '.scanner' "$work/verified-attestation.json")"
     attestation_key="$(jq -r '.signerKeySha256' "$work/verified-attestation.json")"
     attestation_date="$(jq -r '.scannedAt' "$work/verified-attestation.json")"
     # Include only the bounded signed claim and detached signature so the
     # Go evaluator can re-check the trust boundary. Signed claim schema
     # rejects unknown fields; raw vulnerability details are never embedded.
     attestation_claim_b64="$(base64 -w0 -- "$attest_dir/${digest#sha256:}.json")"
     attestation_signature_b64="$(base64 -w0 -- "$attest_dir/${digest#sha256:}.sig")"
   else
     error='EXTERNAL_ATTESTATION_INVALID'
   fi
  fi
  images="$(jq -c --arg registry "$registry" --arg repository "$repository" --arg digest "$digest" --arg scanType "$scan_type" --arg scanStatus "$scan_status" --arg error "$error" --argjson scanDigests "$resolved_scan_digests" --arg scanner "$attestation_scanner" --arg signer "$attestation_key" --arg scannedAt "$attestation_date" --arg signedClaim "$attestation_claim_b64" --arg signedSignature "$attestation_signature_b64" '. + [{registry:$registry,repository:$repository,digest:$digest,scanType:$scanType,scanStatus:$scanStatus,error:$error,scanDigests:$scanDigests,attestationScanner:$scanner,attestationSignerKeySha256:$signer,attestationScannedAt:$scannedAt,attestationClaimBase64:$signedClaim,attestationSignatureBase64:$signedSignature}]' <<<"$images")"
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

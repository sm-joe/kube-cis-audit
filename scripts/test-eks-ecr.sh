#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/collect-eks-ecr.sh"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
export MOCK_PODS="$tmp/pods.json" MOCK_DIGEST="sha256:$(printf '%064d' 0)"
export MOCK_CHILD_DIGEST="sha256:$(printf '%064d' 1)"
cat > "$tmp/bin/kubectl" <<'SH'
#!/usr/bin/env bash
[[ "${MOCK_CASE:-}" != deny ]] || exit 1
cat "$MOCK_PODS"
SH
cat > "$tmp/bin/aws" <<'SH'
#!/usr/bin/env bash
if [[ "$*" == *"sts get-caller-identity"* ]]; then
 echo "${MOCK_ACCOUNT:-999999999999}"
elif [[ "$*" == *get-registry-scanning-configuration* ]]; then
 if [[ "${MOCK_CASE:-}" == awsdeny ]]; then echo 'AccessDeniedException' >&2; exit 1; fi
 echo '{"scanningConfiguration":{"scanType":"ENHANCED"}}'
elif [[ "$*" == *batch-get-image* ]]; then
 if [[ "${MOCK_CASE:-}" != index && "${MOCK_CASE:-}" != indexpartial && "${MOCK_CASE:-}" != indexinvalid ]]; then exit 1; fi
 if [[ "${MOCK_CASE:-}" == indexinvalid ]]; then
  echo '{"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[{"platform":{"os":"unknown","architecture":"unknown"},"digest":"sha256:bad"}]}'
 else
  printf '{"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[{"mediaType":"application/vnd.oci.image.manifest.v1+json","platform":{"os":"linux","architecture":"amd64"},"digest":"%s"},{"mediaType":"application/vnd.oci.image.manifest.v1+json","platform":{"os":"unknown","architecture":"unknown"},"digest":"sha256:attestation"}]}\n' "$MOCK_CHILD_DIGEST"
 fi
elif [[ "$*" == *describe-image-scan-findings* ]]; then
 if [[ "${MOCK_CASE:-}" == awsdeny ]]; then echo 'AccessDeniedException' >&2; exit 1; fi
 if [[ "${MOCK_CASE:-}" == scannotfound ]]; then echo 'ScanNotFoundException' >&2; exit 1; fi
 if [[ "${MOCK_CASE:-}" == mismatch ]]; then echo '{"imageId":{"imageDigest":"sha256:bad"},"scanStatus":"ACTIVE"}'; exit 0; fi
 if [[ "${MOCK_CASE:-}" == index || "${MOCK_CASE:-}" == indexpartial || "${MOCK_CASE:-}" == indexinvalid ]]; then
  if [[ "$*" == *"imageDigest=$MOCK_DIGEST"* ]]; then echo 'ScanNotFoundException' >&2; exit 1; fi
  if [[ "$*" == *"imageDigest=$MOCK_CHILD_DIGEST"* ]]; then
   if [[ "${MOCK_CASE:-}" == indexpartial ]]; then echo 'ScanNotFoundException' >&2; exit 1; fi
   printf '{"imageId":{"imageDigest":"%s"},"scanStatus":"COMPLETE"}\n' "$MOCK_CHILD_DIGEST"; exit 0
  fi
  exit 1
 fi
 printf '{"imageId":{"imageDigest":"%s"},"scanStatus":"ACTIVE"}\n' "$MOCK_DIGEST"
else exit 1;fi
SH
chmod +x "$tmp/bin/"*
cat > "$MOCK_PODS" <<EOF
{"items":[{"spec":{"containers":[{"name":"ecr","image":"123456789012.dkr.ecr.us-east-1.amazonaws.com/nested/app:tag"},{"name":"external","image":"docker.io/example:tag"}]},"status":{"phase":"Running","containerStatuses":[{"name":"ecr","imageID":"123456789012.dkr.ecr.us-east-1.amazonaws.com/nested/app@$MOCK_DIGEST"}]}}]}
EOF
PATH="$tmp/bin:$PATH" bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e --arg d "$MOCK_DIGEST" '.ecrInventoryStatus=="COLLECTED" and .nonEcrImageCount==1 and (.ecrImages|length)==1 and .ecrImages[0].digest==$d and .ecrImages[0].scanStatus=="ACTIVE"' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=awsdeny bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrInventoryStatus=="COLLECTED" and (.ecrImages[0].error | startswith("CROSS_ACCOUNT_ECR_AWS_READ_ACCESS_DENIED")) and .ecrImages[0].scanStatus==""' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=scannotfound bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrImages[0].error=="CROSS_ACCOUNT_ECR_SCAN_NOT_FOUND"' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=mismatch bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrImages[0].error=="SCAN_RESPONSE_DIGEST_OR_STATUS_UNVERIFIED" and .ecrImages[0].scanStatus==""' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=index MOCK_ACCOUNT=123456789012 bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e --arg parent "$MOCK_DIGEST" --arg child "$MOCK_CHILD_DIGEST" '.ecrImages[0].digest==$parent and .ecrImages[0].scanStatus=="COMPLETE" and .ecrImages[0].error=="" and .ecrImages[0].scanDigests==[$child]' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=indexpartial MOCK_ACCOUNT=123456789012 bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrImages[0].scanStatus=="" and .ecrImages[0].error=="OCI_CHILD_SCAN_NOT_FOUND"' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=indexinvalid MOCK_ACCOUNT=123456789012 bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrImages[0].scanStatus=="" and .ecrImages[0].error=="OCI_INDEX_NO_BOUNDED_LINUX_CHILDREN"' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=deny bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrInventoryStatus=="UNAVAILABLE" and (.ecrImages|length)==0' "$tmp/out.json" >/dev/null
echo 'PASS: ECR digest/status, OCI index child scans, partial child, cross-account denial and incomplete Pod inventory'

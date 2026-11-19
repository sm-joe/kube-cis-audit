#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/collect-eks-ecr.sh"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin"
export MOCK_PODS="$tmp/pods.json" MOCK_DIGEST="sha256:$(printf '%064d' 0)"
cat > "$tmp/bin/kubectl" <<'SH'
#!/usr/bin/env bash
[[ "${MOCK_CASE:-}" != deny ]] || exit 1
cat "$MOCK_PODS"
SH
cat > "$tmp/bin/aws" <<'SH'
#!/usr/bin/env bash
if [[ "$*" == *get-registry-scanning-configuration* ]]; then
 [[ "${MOCK_CASE:-}" != awsdeny ]] || exit 1
 echo '{"scanningConfiguration":{"scanType":"ENHANCED"}}'
elif [[ "$*" == *describe-image-scan-findings* ]]; then
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
jq -e '.ecrInventoryStatus=="COLLECTED" and .ecrImages[0].error != "" and .ecrImages[0].scanStatus==""' "$tmp/out.json" >/dev/null
PATH="$tmp/bin:$PATH" MOCK_CASE=deny bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/out.json"
jq -e '.ecrInventoryStatus=="UNAVAILABLE" and (.ecrImages|length)==0' "$tmp/out.json" >/dev/null
echo 'PASS: ECR digest/status, denied AWS API and denied Pod inventory'

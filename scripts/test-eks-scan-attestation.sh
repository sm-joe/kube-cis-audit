#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/verify-eks-scan-attestation.sh"
tmp="$(mktemp -d)"; trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/claims"
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$tmp/private.pem" >/dev/null 2>&1
openssl pkey -in "$tmp/private.pem" -pubout -out "$tmp/public.pem" >/dev/null 2>&1
pin="$(sha256sum "$tmp/public.pem" | cut -d' ' -f1)"
digest="sha256:$(printf '%064d' 1)"
child="sha256:$(printf '%064d' 2)"
file="$tmp/claims/${digest#sha256:}.json"
sig="$tmp/claims/${digest#sha256:}.sig"
now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo '{"Results":[],"Metadata":{"scanner":"approved-scanner"}}' > "$tmp/claims/${digest#sha256:}.scan.json"
report_sha="$(sha256sum "$tmp/claims/${digest#sha256:}.scan.json" | cut -d' ' -f1)"
jq -n --arg digest "$digest" --arg scanned "$now" --arg child "$child" --arg reportSha "$report_sha" '{
 schemaVersion:1,subject:{registry:"602401143452.dkr.ecr.us-east-1.amazonaws.com",repository:"eks/coredns",digest:$digest},
 scanner:"approved-scanner",scanId:"run-001",status:"COMPLETE",findingsAvailable:true,scannerReportSha256:$reportSha,scannedAt:$scanned,scannedDigests:[$child]
}' > "$file"
openssl dgst -sha256 -sign "$tmp/private.pem" -out "$sig" "$file"
verify() {
 bash "$root/scripts/verify-eks-scan-attestation.sh" \
 --dir "$tmp/claims" --public-key "$tmp/public.pem" --key-sha256 "$pin" \
 --registry "602401143452.dkr.ecr.us-east-1.amazonaws.com" \
 --repository "eks/coredns" --digest "$digest" --output-file "$tmp/evidence.json"
}
verify
jq -e --arg child "$child" '.verified == true and .scanDigests == [$child]' "$tmp/evidence.json" >/dev/null
# Verify actual collector integration against an AWS-owned registry with
# deliberately denied cross-account ECR scan-read API calls.
mkdir -p "$tmp/bin"
cat > "$tmp/bin/kubectl" <<'MOCK'
#!/usr/bin/env bash
cat "$MOCK_PODS"
MOCK
cat > "$tmp/bin/aws" <<'MOCK'
#!/usr/bin/env bash
if [[ "$*" == *"sts get-caller-identity"* ]]; then
 echo '638631156706'
else
 echo 'AccessDeniedException' >&2
 exit 1
fi
MOCK
chmod +x "$tmp/bin/"*
export MOCK_PODS="$tmp/pods.json"
cat > "$MOCK_PODS" <<EOF
{"items":[{"spec":{"containers":[{"name":"coredns","image":"602401143452.dkr.ecr.us-east-1.amazonaws.com/eks/coredns@$digest"}]},"status":{"phase":"Running","containerStatuses":[{"name":"coredns","imageID":"602401143452.dkr.ecr.us-east-1.amazonaws.com/eks/coredns@$digest"}]}}]}
EOF
PATH="$tmp/bin:$PATH" KUBECIS_ATTESTATION_DIR="$tmp/claims" \
 KUBECIS_ATTESTATION_PUBLIC_KEY="$tmp/public.pem" KUBECIS_ATTESTATION_KEY_SHA256="$pin" \
 bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/provider.json"
jq -e --arg key "$pin" '.ecrInventoryStatus == "COLLECTED" and .ecrImages[0].scanType == "THIRD_PARTY" and .ecrImages[0].scanStatus == "COMPLETE" and .ecrImages[0].error == "" and .ecrImages[0].attestationSignerKeySha256 == $key' "$tmp/provider.json" >/dev/null
if bash "$root/scripts/verify-eks-scan-attestation.sh" \
 --dir "$tmp/claims" --public-key "$tmp/public.pem" --key-sha256 "$pin" \
 --registry "602401143452.dkr.ecr.us-east-1.amazonaws.com" \
 --repository "eks/kube-proxy" --digest "$digest" --output-file "$tmp/evidence.json" >/dev/null 2>&1; then
 echo 'FAIL: accepted attestation for wrong repository' >&2;exit 1
fi
jq '.scanId = "tampered"' "$file" > "$tmp/tampered"
mv "$tmp/tampered" "$file"
if verify >/dev/null 2>&1; then echo 'FAIL: accepted tampered scan evidence' >&2;exit 1;fi
PATH="$tmp/bin:$PATH" KUBECIS_ATTESTATION_DIR="$tmp/claims" \
 KUBECIS_ATTESTATION_PUBLIC_KEY="$tmp/public.pem" KUBECIS_ATTESTATION_KEY_SHA256="$pin" \
 bash "$root/scripts/collect-eks-ecr.sh" --context demo --region us-east-1 --output-file "$tmp/provider.json"
jq -e '.ecrImages[0].scanStatus == "" and .ecrImages[0].error == "EXTERNAL_ATTESTATION_INVALID"' "$tmp/provider.json" >/dev/null
jq '.scanId = "run-001" | .scannedAt = "2020-01-01T00:00:00Z"' "$file" > "$tmp/stale"
mv "$tmp/stale" "$file"
openssl dgst -sha256 -sign "$tmp/private.pem" -out "$sig" "$file"
if verify >/dev/null 2>&1; then echo 'FAIL: accepted signed stale evidence' >&2;exit 1;fi
# Even with a valid claim signature, a changed raw scan report must fail.
jq '.scannedAt = "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"' "$file" > "$tmp/fresh"
mv "$tmp/fresh" "$file"
openssl dgst -sha256 -sign "$tmp/private.pem" -out "$sig" "$file"
echo '{"Results":[{"modified":true}]}' > "$tmp/claims/${digest#sha256:}.scan.json"
if verify >/dev/null 2>&1; then echo 'FAIL: accepted modified scan report with stale digest' >&2;exit 1;fi
echo 'PASS: signed digest-bound scan, mismatched subject, tampering, signed stale claim'

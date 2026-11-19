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
jq -n --arg digest "$digest" --arg scanned "$now" --arg child "$child" '{
 schemaVersion:1,subject:{registry:"602401143452.dkr.ecr.us-east-1.amazonaws.com",repository:"eks/coredns",digest:$digest},
 scanner:"approved-scanner",scanId:"run-001",status:"COMPLETE",findingsAvailable:true,scannedAt:$scanned,scannedDigests:[$child]
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
if bash "$root/scripts/verify-eks-scan-attestation.sh" \
 --dir "$tmp/claims" --public-key "$tmp/public.pem" --key-sha256 "$pin" \
 --registry "602401143452.dkr.ecr.us-east-1.amazonaws.com" \
 --repository "eks/kube-proxy" --digest "$digest" --output-file "$tmp/evidence.json" >/dev/null 2>&1; then
 echo 'FAIL: accepted attestation for wrong repository' >&2;exit 1
fi
jq '.scanId = "tampered"' "$file" > "$tmp/tampered"
mv "$tmp/tampered" "$file"
if verify >/dev/null 2>&1; then echo 'FAIL: accepted tampered scan evidence' >&2;exit 1;fi
jq '.scanId = "run-001" | .scannedAt = "2020-01-01T00:00:00Z"' "$file" > "$tmp/stale"
mv "$tmp/stale" "$file"
openssl dgst -sha256 -sign "$tmp/private.pem" -out "$sig" "$file"
if verify >/dev/null 2>&1; then echo 'FAIL: accepted signed stale evidence' >&2;exit 1;fi
echo 'PASS: signed digest-bound scan, mismatched subject, tampering, signed stale claim'

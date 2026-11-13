#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
for name in collect-eks-provider.sh collect-eks.sh validate-eks-preview.sh deploy-eks.sh; do
  bash -n "$root/scripts/$name"
done
if bash "$root/scripts/validate-eks-preview.sh" --release-level full --report /dev/null --cluster-name sample --region us-east-1 >/dev/null 2>&1; then
  echo 'FAIL: full CIS compliance gate unexpectedly accepted' >&2
  exit 1
fi
if bash "$root/scripts/collect-eks.sh" --context dummy --cluster-name sample --region us-east-1 --namespace default >/dev/null 2>&1; then
  echo 'FAIL: unsupported namespace unexpectedly accepted' >&2
  exit 1
fi
# Exercise the contract gate using an explicit invalid document, independent
# of any live AWS identity. Partial input cannot become a passing assessment.
tmp="$(mktemp)"
trap 'rm -f -- "$tmp"' EXIT
echo '{"auditMode":"consolidated","cloudProvider":"eks","compliance":{"compliancePercent":100}}' > "$tmp"
if bash "$root/scripts/validate-eks-preview.sh" --report "$tmp" --cluster-name sample --region us-east-1 >/dev/null 2>&1; then
  echo 'FAIL: invented 100% CIS compliance incorrectly accepted' >&2
  exit 1
fi
echo 'PASS: POSIX script syntax and fail-closed input contracts'

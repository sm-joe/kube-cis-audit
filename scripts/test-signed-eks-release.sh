#!/usr/bin/env bash
# Contract tests use mock Cosign; tagged publication performs real verification.
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
script="$root/scripts/accept-signed-eks-release.sh"
bash -n "$script"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/assets" "$tmp/pkg/scripts"
case "$(uname -m)" in x86_64) arch=amd64;; aarch64|arm64) arch=arm64;; *) exit 1;; esac
tag=v3.9.9-rc1
archive="kubecis_linux_${arch}.tar.gz"
cat > "$tmp/bin/cosign" <<'EOF'
#!/bin/sh
test "${REJECT_SIGNATURE:-0}" = 0
EOF
printf '#!/bin/sh\necho v3.9.9-rc1\n' > "$tmp/pkg/kubecis"
printf '#!/bin/sh\nexit 0\n' > "$tmp/pkg/scripts/test-linux-eks.sh"
chmod +x "$tmp/bin/cosign" "$tmp/pkg/kubecis"
export PATH="$tmp/bin:$PATH"
tar -czf "$tmp/assets/$archive" -C "$tmp/pkg" .
(cd "$tmp/assets" && sha256sum "$archive" > SHA256SUMS)
jq -n --arg tag "$tag" --arg commit "$(printf 'a%.0s' {1..40})" --arg image "ghcr.io/sm-joe/kubecis@sha256:$(printf 'b%.0s' {1..64})" '{schemaVersion:1,version:$tag,commit:$commit,image:$image}' > "$tmp/assets/release-manifest.json"
for file in "$archive" SHA256SUMS release-manifest.json; do echo '{}' > "$tmp/assets/$file.sigstore.json"; done
run(){ bash "$script" --release "$tag" --assets "$tmp/assets" --verify-only; }
run
if REJECT_SIGNATURE=1 run >/dev/null 2>&1; then echo 'FAIL: rejected signature accepted'; exit 1; fi
cp "$tmp/assets/release-manifest.json" "$tmp/original"
jq '.version="v0.0.1"' "$tmp/original" > "$tmp/assets/release-manifest.json"
if run >/dev/null 2>&1; then echo 'FAIL: wrong release accepted'; exit 1; fi
cp "$tmp/original" "$tmp/assets/release-manifest.json"
echo tamper >> "$tmp/assets/$archive"
if run >/dev/null 2>&1; then echo 'FAIL: tampered archive accepted'; exit 1; fi
echo 'PASS: mock-signature release acceptance, wrong version and altered archive rejection'

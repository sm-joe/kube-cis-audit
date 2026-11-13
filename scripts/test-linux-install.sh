#!/usr/bin/env bash
# Test installer mechanics with explicitly MOCKED cosign. This is not a
# cryptographic verification test; real keyless signing runs at tagged release.
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
bash -n "$root/scripts/install-linux.sh"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
mkdir -p "$tmp/bin" "$tmp/pkg/benchmarks" "$tmp/install"
cat > "$tmp/bin/cosign" <<'EOF'
#!/bin/sh
# Test-only signer mock: real verification occurs in tagged-release job.
test "$1" = verify-blob
EOF
chmod +x "$tmp/bin/cosign"
export PATH="$tmp/bin:$PATH"
case "$(uname -m)" in
  x86_64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; *) echo 'Unsupported test architecture'; exit 1 ;;
esac
archive="$tmp/kubecis_linux_${arch}.tar.gz"
package() {
  version=$1
  printf '#!/bin/sh\nif [ "$1" = version ]; then echo "%s"; else exit 1; fi\n' "$version" > "$tmp/pkg/kubecis"
  chmod 755 "$tmp/pkg/kubecis"
  tar -czf "$archive" -C "$tmp/pkg" .
  (cd "$tmp" && sha256sum "kubecis_linux_${arch}.tar.gz" > SHA256SUMS)
  echo '{}' > "$archive.sigstore.json"
}
install() {
  bash "$root/scripts/install-linux.sh" --archive "$archive" \
    --checksums "$tmp/SHA256SUMS" --bundle "$archive.sigstore.json" \
    --release "$1" --prefix "$tmp/install" --bin-dir "$tmp/bin"
}
package v3.9.1-rc1
install v3.9.1-rc1
(cd / && "$tmp/bin/kubecis" version | grep -Fx 'v3.9.1-rc1')
package v3.9.2-rc2
install v3.9.2-rc2
(cd / && "$tmp/bin/kubecis" version | grep -Fx 'v3.9.2-rc2')
bash "$root/scripts/install-linux.sh" --rollback --prefix "$tmp/install" --bin-dir "$tmp/bin"
(cd / && "$tmp/bin/kubecis" version | grep -Fx 'v3.9.1-rc1')
echo 'tamper' >> "$archive"
if install v3.9.3-rc3 >/dev/null 2>&1; then
  echo 'FAIL: archive tampering accepted' >&2
  exit 1
fi
echo 'PASS: mocked-signature Linux package install, upgrade, rollback and checksum rejection'

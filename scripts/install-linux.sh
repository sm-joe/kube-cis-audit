#!/usr/bin/env bash
# Install or upgrade a signed KubeCIS Linux release archive without root.
# Rollback switches to the last installed release. The release archive, its
# Sigstore bundle and a checksum manifest must come from the same GitHub release.
set -Eeuo pipefail
die() { echo "ERROR: $*" >&2; exit 1; }
usage() {
  cat <<'EOF'
Usage:
  bash scripts/install-linux.sh --archive kubecis_linux_amd64.tar.gz \
    --checksums SHA256SUMS --bundle kubecis_linux_amd64.tar.gz.sigstore.json \
    --release vX.Y.Z-rcN [--prefix ~/.local/share/kubecis] [--bin-dir ~/.local/bin]
  bash scripts/install-linux.sh --rollback [--prefix DIR] [--bin-dir DIR]

Requires GNU coreutils, tar and Cosign 3.x for signed installation.
Never use a checksum alone as a substitute for signed provenance.
EOF
}
archive='' sums='' bundle='' release='' rollback=false
prefix="$HOME/.local/share/kubecis" bin_dir="$HOME/.local/bin"
while (($#)); do
  case "$1" in
    --archive|--checksums|--bundle|--release|--prefix|--bin-dir)
      (($# >= 2)) || die "Missing value for $1"
      case "$1" in
        --archive) archive=$2 ;; --checksums) sums=$2 ;;
        --bundle) bundle=$2 ;; --release) release=$2 ;;
        --prefix) prefix=$2 ;; --bin-dir) bin_dir=$2 ;;
      esac
      shift 2 ;;
    --rollback) rollback=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown argument: $1" ;;
  esac
done
# Only support simple absolute filesystem paths: reject whitespace, shell
# metacharacters and path traversal before generating a persistent wrapper.
[[ "$prefix" =~ ^/[A-Za-z0-9_./-]+$ && "$bin_dir" =~ ^/[A-Za-z0-9_./-]+$ ]] || die 'Prefix and bin directory must be simple absolute paths'
[[ "$prefix" != *'..'* && "$bin_dir" != *'..'* ]] || die 'Path traversal refused'
mkdir -p "$prefix/releases" "$bin_dir"
old_target=''
if [[ -L "$prefix/current" ]]; then
  old_target="$(readlink "$prefix/current")"
fi
swap_current() {
  local target=$1 link="$prefix/.current-$$"
  ln -s "$target" "$link"
  mv -Tf -- "$link" "$prefix/current"
}
if $rollback; then
  [[ -z "$archive$sums$bundle$release" ]] || die 'Do not supply an archive when rolling back'
  [[ -L "$prefix/previous" ]] || die 'No previous release to roll back to'
  target="$(readlink "$prefix/previous")"
  [[ -x "$target/kubecis" && -d "$target/benchmarks" ]] || die 'Previous version is not a valid release'
  swap_current "$target"
  if [[ -n "$old_target" ]]; then
    ln -s "$old_target" "$prefix/.previous-$$"
    mv -Tf -- "$prefix/.previous-$$" "$prefix/previous"
  fi
  echo "Rolled back to $("$target/kubecis" version)"
  exit 0
fi
[[ -f "$archive" && -f "$sums" && -f "$bundle" ]] || die 'Release archive, SHA256SUMS and Sigstore bundle required'
[[ "$release" =~ ^v[0-9]+[.][0-9]+[.][0-9]+(-[0-9A-Za-z.-]+)?$ ]] || die 'Invalid release tag'
command -v cosign >/dev/null || die 'Cosign 3.x required for verified installation'
for tool in sha256sum tar uname; do command -v "$tool" >/dev/null || die "Missing tool $tool"; done
case "$(uname -m)" in
  x86_64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;;
  *) die 'Only Linux amd64 and arm64 are supported for this installer' ;;
esac
[[ "$(uname -s)" == Linux ]] || die 'Linux required'
filename="$(basename -- "$archive")"
[[ "$filename" == "kubecis_linux_${arch}.tar.gz" ]] || die "Archive platform mismatch: $filename"
expected="$(awk -v file="$filename" '$2 == file {print $1}' "$sums")"
[[ "$expected" =~ ^[0-9a-fA-F]{64}$ ]] || die 'Expected exactly one SHA256SUMS digest for this archive'
actual="$(sha256sum "$archive")"
actual="${actual%% *}"
[[ "${actual,,}" == "${expected,,}" ]] || die 'Archive SHA256 mismatch'
identity="https://github.com/sm-joe/kube-cis-audit/.github/workflows/release.yml@refs/tags/$release"
cosign verify-blob "$archive" --bundle "$bundle" \
  --certificate-identity "$identity" \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' >/dev/null || die 'Release signature invalid'
final="$prefix/releases/$release"
stage="$(mktemp -d "$prefix/.incoming-XXXXXXXX")"
trap 'rm -rf -- "$stage"' EXIT
tar -xzf "$archive" -C "$stage" --no-same-owner --no-same-permissions
[[ -x "$stage/kubecis" && -d "$stage/benchmarks" ]] || die 'Release archive lacks executable or benchmarks'
[[ "$("$stage/kubecis" version)" == "$release" ]] || die 'Release binary version does not match signed archive tag'
if [[ -e "$final" ]]; then
  [[ -x "$final/kubecis" && "$("$final/kubecis" version)" == "$release" ]] || die 'Existing release directory is inconsistent'
else
  mv -- "$stage" "$final"
  stage=''
fi
# The wrapper ensures benchmarks resolve correctly even when invoked from an
# unrelated working directory. Install it only after the signed payload is valid.
wrapper="$bin_dir/.kubecis-wrapper-$$"
printf '#!/usr/bin/env bash\nexport KUBECIS_CONFIG_DIR="%s/current/benchmarks"\nexec "%s/current/kubecis" "$@"\n' "$prefix" "$prefix" > "$wrapper"
chmod 755 "$wrapper"
mv -f -- "$wrapper" "$bin_dir/kubecis"
if [[ -n "$old_target" && "$old_target" != "$final" ]]; then
  ln -s "$old_target" "$prefix/.previous-$$"
  mv -Tf -- "$prefix/.previous-$$" "$prefix/previous"
fi
swap_current "$final"
echo "Installed verified $release. Binary: $bin_dir/kubecis"
echo "Prior version: ${old_target:-none}. Run with --rollback if the previous release is required."

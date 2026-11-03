#!/usr/bin/env bash
set -Eeuo pipefail

# Reproducible upstream import. Update the tag and commit together only after
# reviewing the upstream release and its configuration changes.
readonly UPSTREAM_REPO="https://github.com/aquasecurity/kube-bench.git"
readonly UPSTREAM_TAG="v0.16.0"
readonly EXPECTED_COMMIT="5c6c22d"
readonly DEST="benchmarks/upstream/kube-bench"

command -v git >/dev/null 2>&1 || { echo "ERROR: git is required" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo "ERROR: Go is required" >&2; exit 1; }

repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$repo_root"
tmp_dir="$(mktemp -d)"
cleanup() { rm -rf "$tmp_dir"; }
trap cleanup EXIT

echo "[1/5] Fetching kube-bench ${UPSTREAM_TAG}..."
git clone --quiet --depth 1 --branch "$UPSTREAM_TAG" "$UPSTREAM_REPO" "$tmp_dir/repo"
actual_commit="$(git -C "$tmp_dir/repo" rev-parse HEAD)"
if [[ "$actual_commit" != "$EXPECTED_COMMIT"* ]]; then
  echo "ERROR: expected commit ${EXPECTED_COMMIT} for ${UPSTREAM_TAG}, got ${actual_commit}" >&2
  exit 1
fi

[[ -d "$tmp_dir/repo/cfg" ]] || { echo "ERROR: upstream cfg/ directory is missing" >&2; exit 1; }
[[ -f "$tmp_dir/repo/LICENSE" ]] || { echo "ERROR: upstream LICENSE is missing" >&2; exit 1; }

# Stage first, then replace only the dedicated upstream directory. Existing
# KubeCIS profiles under benchmarks/cis-* are deliberately untouched.
stage="$tmp_dir/stage"
mkdir -p "$stage"
cp -a "$tmp_dir/repo/cfg" "$stage/cfg"
cp "$tmp_dir/repo/LICENSE" "$stage/LICENSE"
cat > "$stage/UPSTREAM_VERSION" <<VERSION
repository=${UPSTREAM_REPO}
tag=${UPSTREAM_TAG}
commit=${actual_commit}
imported_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION

mkdir -p "$(dirname "$DEST")"
rm -rf "$DEST"
mv "$stage" "$DEST"

echo "[2/5] Imported upstream configuration and license."
echo "[3/5] Generating structural inventory..."
go run ./cmd/benchmark-inventory \
  --root "$DEST/cfg" \
  --out docs/UPSTREAM-BENCHMARK-INVENTORY.md \
  --json docs/UPSTREAM-BENCHMARK-INVENTORY.json

echo "[4/5] Running inventory tests..."
go test ./cmd/benchmark-inventory -count=1

echo "[5/5] Upstream import complete."
echo "Pinned release: ${UPSTREAM_TAG} (${actual_commit})"
echo "Imported files: ${DEST}/cfg"
echo "Inventory: docs/UPSTREAM-BENCHMARK-INVENTORY.md"

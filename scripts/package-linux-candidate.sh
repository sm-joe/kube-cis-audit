#!/usr/bin/env bash
# Shared candidate layout for archive creation and credential-free acceptance.
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
[[ $# == 1 && -n "$1" ]] || { echo "Usage: $0 STAGE_DIRECTORY" >&2; exit 2; }
stage="$1"
mkdir -p "$stage/internal/audit/catalogs"
cp -R "$root/benchmarks" "$root/scripts" "$root/deploy" "$root/docs" "$stage/"
cp "$root/README.md" "$root/LICENSE" "$root/ATTRIBUTION.md" "$stage/"
cp "$root/internal/audit/catalogs/eks-2.1.0.json" "$stage/internal/audit/catalogs/"
test -s "$stage/internal/audit/catalogs/eks-2.1.0.json"

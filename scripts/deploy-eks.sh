#!/usr/bin/env bash
# Fail closed on unpinned release images. Deploy only the node collector; this
# command creates/updates Kubernetes resources, it does not run a scan.
set -Eeuo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/deploy-eks.sh --image ghcr.io/OWNER/IMAGE@sha256:DIGEST [--context CONTEXT] [--namespace kube-system]

Requires kubectl, an existing authorized EKS kubeconfig and a verified image
digest from the same source commit as the KubeCIS CLI. The deployment changes
Kubernetes resources (ServiceAccount, DaemonSet and RBAC).
EOF
}
image='' context='' namespace='kube-system'
while (($#)); do
  case "$1" in
    --image|--context|--namespace)
      (($# >= 2)) || { echo "Missing value for $1" >&2; exit 2; }
      case "$1" in
        --image) image="$2" ;;
        --context) context="$2" ;;
        --namespace) namespace="$2" ;;
      esac
      shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done
if [[ ! "$image" =~ ^[a-z0-9][a-z0-9./_-]*@sha256:[a-f0-9]{64}$ ]]; then
  echo "Refusing non-immutable image: supply a registry/repository@sha256:<64 hex> reference" >&2
  exit 2
fi
if [[ "$namespace" != 'kube-system' ]]; then
  echo "This release manifest is fixed to kube-system; other namespaces are not supported" >&2
  exit 2
fi
command -v kubectl >/dev/null || { echo "kubectl required" >&2; exit 2; }
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "$script_dir/.." && pwd)"
kubectl_args=()
if [[ -n "$context" ]]; then kubectl_args+=(--context "$context"); fi
server="$(kubectl "${kubectl_args[@]}" version -o json)"
# Version is checked from the server response, not from the client.
if ! printf '%s' "$server" | grep -Eq -- '-eks-'; then
  echo "Selected context does not identify an EKS server; refusing deployment" >&2
  exit 2
fi
echo "Selected kubeconfig context: $(kubectl "${kubectl_args[@]}" config current-context)"
echo "Image (immutable): $image"
# Replace only the known candidate image reference in the rendered output.
# Refuse unexpected manifest changes rather than accidentally shipping a
# mutable image or deploying a partially updated DaemonSet.
manifest="$(kubectl kustomize "$repo_dir/deploy/eks")"
placeholder='ghcr.io/sm-joe/kubecis:eks-provider-rc3-identity'
old_count="$(printf '%s\n' "$manifest" | grep -Fc "image: $placeholder" || true)"
if [[ "$old_count" -ne 3 ]]; then
  echo "Expected exactly three known candidate image references; found $old_count" >&2
  exit 1
fi
patched="${manifest//image: $placeholder/image: $image}"
count="$(printf '%s\n' "$patched" | grep -Fc "image: $image" || true)"
if [[ "$count" -ne 3 ]]; then
  echo "Expected three digest-pinned image references; found $count; refusing apply" >&2
  exit 1
fi
printf '%s\n' "$patched" | kubectl "${kubectl_args[@]}" apply -f -
kubectl "${kubectl_args[@]}" -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=300s
echo 'EKS node auditor deployed; run evidence collection separately.'

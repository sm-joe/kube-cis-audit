#!/usr/bin/env bash
# Collect EKS API, all managed EC2 worker metadata and AWS provider snapshots.
# Requires: Linux, Bash, jq, kubectl, AWS CLI v2, matched kubecis CLI+image.
# The CLI enforces freshness, source uniqueness and official score withholding.
set -Eeuo pipefail
die() { echo "ERROR: $*" >&2; exit 1; }
usage() {
  echo 'Usage: bash scripts/collect-eks.sh --context KUBE_CONTEXT --cluster-name NAME --region REGION [--binary ./kubecis] [--output-dir reports/eks] [--namespace kube-system]'
}
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd -- "$script_dir/.." && pwd)"
context='' cluster='' region='' binary="$root/kubecis" output="$root/reports/eks" namespace='kube-system'
while (($#)); do
  case "$1" in
    --context|--cluster-name|--region|--binary|--output-dir|--namespace)
      (($# >= 2)) || die "Missing value for $1"
      case "$1" in
        --context) context=$2 ;; --cluster-name) cluster=$2 ;;
        --region) region=$2 ;; --binary) binary=$2 ;;
        --output-dir) output=$2 ;; --namespace) namespace=$2 ;;
      esac
      shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown argument: $1" ;;
  esac
done
[[ -n "$context" && -n "$cluster" && -n "$region" ]] || { usage >&2; exit 2; }
[[ "$namespace" == kube-system ]] || die 'Only kube-system deployment is supported'
for tool in jq kubectl aws; do command -v "$tool" >/dev/null || die "$tool missing"; done
[[ -x "$binary" ]] || die "Missing executable $binary"
[[ -s "$root/benchmarks/cis-eks-2.1.0/benchmark.yaml" ]] || die 'Missing benchmark profile'
[[ "$(kubectl config current-context)" == "$context" ]] || die "Selected context must equal current kubeconfig context ($context)"
version_json="$(kubectl --context "$context" version -o json)"
jq -e '.serverVersion.gitVersion | contains("-eks-")' <<<"$version_json" >/dev/null || die 'Expected EKS API server'
kube_version="$(jq -r '.serverVersion.gitVersion' <<<"$version_json")"
kube_minor="$(sed -E 's/^v([0-9]+[.][0-9]+).*/\1/' <<<"$kube_version")"
api_server="$(kubectl --context "$context" config view --minify -o json | jq -r '.clusters[0].cluster.server // empty')"
[[ "$api_server" == https://* ]] || die 'Kubeconfig EKS API endpoint missing'
aws_raw="$(aws eks describe-cluster --region "$region" --name "$cluster" --output json)" || die 'EKS DescribeCluster failed'
jq -e --arg url "${api_server%/}" --arg ver "$kube_minor" --arg name "$cluster" '
 .cluster.name == $name and (.cluster.endpoint | rtrimstr("/")) == $url and .cluster.version == $ver
' <<<"$aws_raw" >/dev/null || die 'AWS EKS endpoint/version mismatch'
nodes_json="$(kubectl --context "$context" get nodes -o json)"
mapfile -t nodes < <(jq -r '.items[] | select(.metadata.labels["kubernetes.io/os"] == "linux") | .metadata.name' <<<"$nodes_json" | sort -u)
(("${#nodes[@]}" > 0)) || die 'No Linux workers'
# Reject Fargate/Auto Mode/self-managed and Windows workers. Inventory scope is
# exact: every reported node must have the supported managed EC2 nodegroup.
jq -e --argjson count "${#nodes[@]}" '
 (.items | length) == $count and
 all(.items[]; .metadata.labels["kubernetes.io/os"] == "linux" and
     (.metadata.labels["eks.amazonaws.com/nodegroup"] // "") != "")
' <<<"$nodes_json" >/dev/null || die 'Mixed or unsupported compute modes; refusing incomplete assessment'
pods_json="$(kubectl --context "$context" -n "$namespace" get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json)"
local_version="$("$binary" version)" || die 'CLI version unavailable'
[[ -n "$local_version" ]] || die 'CLI version missing'
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
inputs=()
"$binary" scan --context "$context" --config-dir "$root/benchmarks" --benchmark cis-eks-2.1.0 \
  --target policies --engine native --output json --output-file "$tmp/cluster.json" --exit-code-on-fail=false
jq -e '.auditMode == "native" and .cloudProvider == "eks" and (.summary.total == (.results | length))' "$tmp/cluster.json" >/dev/null || die 'Incomplete native API scan'
inputs+=("$tmp/cluster.json")
for node in "${nodes[@]}"; do
  mapfile -t matches < <(jq -r --arg n "$node" '.items[] | select(.spec.nodeName == $n and .status.phase == "Running") | .metadata.name' <<<"$pods_json")
  (("${#matches[@]}" == 1)) || die "Need exactly one running node-auditor Pod for $node; got ${#matches[@]}"
  pod="${matches[0]}"
  node_version="$(kubectl --context "$context" -n "$namespace" exec "$pod" -c node-auditor -- /usr/local/bin/kubecis version)" || die "Cannot inspect node auditor version for $node"
  [[ "$node_version" == "$local_version" ]] || die "Version mismatch on $node: CLI=$local_version auditor=$node_version"
  safe="${node//[^a-zA-Z0-9_.-]/_}"
  report="$tmp/node-$safe.json"
  kubectl --context "$context" -n "$namespace" exec "$pod" -c node-auditor -- \
    /usr/local/bin/kubecis node-audit --provider eks --benchmark cis-eks-2.1.0 \
    --target node --metadata-file /metadata/snapshot.json --node-name "$node" \
    --output json --exit-code-on-fail=false > "$report" || die "Node audit failed on $node"
  jq -e --arg node "$node" '.auditMode == "node-local" and .nodeName == $node and (.summary.total == (.results | length))' "$report" >/dev/null || die "Invalid node audit for $node"
  # Optional nodes/proxy permission. No raw configz is retained in the output.
  if kubectl --context "$context" get --raw "/api/v1/nodes/$node/proxy/configz" > "$tmp/configz.json" 2>/dev/null; then
    "$binary" enrich-node --input "$report" --configz-file "$tmp/configz.json" --node-name "$node" --output-file "$report" || die "Invalid configz evidence on $node"
  else
    echo "WARN: node/proxy configz unavailable for $node; affected controls remain unknown" >&2
  fi
  rm -f -- "$tmp/configz.json"
  inputs+=("$report")
done
bash "$script_dir/collect-eks-provider.sh" --cluster-name "$cluster" --region "$region" \
  --expected-server "$api_server" --output-file "$tmp/provider.json"
args=(consolidate --cluster "$tmp/cluster.json" --provider-evidence "$tmp/provider.json"
  --expected-nodes "$(IFS=,; echo "${nodes[*]}")" --max-age 30m --output-file "$tmp/consolidated.json")
for report in "${inputs[@]:1}"; do args+=(--local "$report"); done
"$binary" "${args[@]}" || die 'Consolidation failed (missing/stale/inconsistent evidence)'
bash "$script_dir/validate-eks-preview.sh" --report "$tmp/consolidated.json" --cluster-name "$cluster" --region "$region"
"$binary" render --input "$tmp/consolidated.json" --output-file "$tmp/results.txt"
"$binary" review-init --input "$tmp/consolidated.json" --output-file "$tmp/review-template.json"
"$binary" finalize --input "$tmp/consolidated.json" --output-dir "$tmp"
# Publish only after the complete preview contract passes.
mkdir -p "$output"
cp -- "$tmp/cluster.json" "$tmp/provider.json" "$tmp/consolidated.json" "$tmp/results.txt" \
 "$tmp/review-template.json" "$tmp/final-assessment.json" "$tmp/final-assessment.txt" "$output/"
for report in "${inputs[@]:1}"; do cp -- "$report" "$output/"; done
echo "PASS (PREVIEW ONLY): output $output; compliance score WITHHELD"

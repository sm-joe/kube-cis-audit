#!/usr/bin/env bash
# Structural acceptance of an explicitly provisional EKS v2.1.0 report.
# Not a certification gate: a complete official CIS benchmark requires
# independent mapping and conformance evidence.
set -Eeuo pipefail
die() { echo "EKS preview rejected: $*" >&2; exit 1; }
usage() {
  echo 'Usage: bash scripts/validate-eks-preview.sh --report PATH --cluster-name NAME --region REGION [--release-level preview|full]'
}
report='' name='' region='' level='preview'
while (($#)); do
  case "$1" in
    --report|--cluster-name|--region|--release-level)
      (($# >= 2)) || die "Missing value $1"
      case "$1" in
        --report) report=$2 ;; --cluster-name) name=$2 ;;
        --region) region=$2 ;; --release-level) level=$2 ;;
      esac
      shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) die "Unknown argument $1" ;;
  esac
done
[[ "$level" == preview ]] || die 'FULL CIS EKS release is blocked pending independently validated official benchmark coverage'
[[ -n "$report" && -n "$name" && -n "$region" ]] || { usage >&2; exit 2; }
[[ -s "$report" ]] || die "Missing report $report"
command -v jq >/dev/null || die 'jq required'
jq -e --arg name "$name" --arg region "$region" '
  .auditMode == "consolidated" and .cloudProvider == "eks"
  and .benchmark == "cis-eks-2.1.0"
  and .benchmarkCompatibilityStatus == "provisional"
  and (.compliance.compliancePercent == null)
  and (.compliance.scoreStatus | startswith("withheld:"))
  and (.compliance.scoredTotal >= 1)
  and ((.compliance.scoredPass + .compliance.scoredFail) <= .compliance.scoredTotal)
  and (.summary.total == (.results | length))
  and (.expectedNodes | length) > 0
  and ((.expectedControlPlaneNodes // []) | length) == 0
  and (.sources | length) == ((.expectedNodes | length) + 2)
  and ([.sources[] | select(.auditMode == "native" and (.nodeName // "") == "")] | length) == 1
  and ([.sources[] | select(.auditMode == "aws-provider")] | length) == 1
  and ([.sources[] | select(.auditMode == "master-local")] | length) == 0
  and (.providerEvidence.schemaVersion == 1)
  and (.providerEvidence.assessmentStatus == "INVENTORY_ONLY")
  and (.providerEvidence.clusterName == $name)
  and (.providerEvidence.region == $region)
  and (.providerEvidence.cloudProvider == "eks")
  and (.providerEvidence.clusterArn | test("^arn:aws[a-z-]*:eks:" + $region + ":[0-9]{12}:cluster/" + $name + "$"))
  and (.clusterVersion | startswith("v" + .providerEvidence.kubernetesVersion + "."))
  and (["api","audit","authenticator","controllerManager","scheduler"] |
       all(. as $t | .providerEvidence.controlPlaneLogging | has($t)))
' "$report" >/dev/null || die 'Report fails identity, accounting or withheld-score contract'
# Temporal checks are evaluated using date(1), not timestamps supplied by the
# report. Accept at most two minutes future clock skew.
now="$(date -u +%s)"
while IFS= read -r timestamp; do
  [[ -n "$timestamp" ]] || die 'Missing source timestamp'
  epoch="$(date -u -d "$timestamp" +%s 2>/dev/null)" || die "Invalid timestamp $timestamp"
  (( epoch <= now + 120 && epoch >= now - 1800 )) || die "Stale/future source $timestamp"
done < <(jq -r '.sources[].generatedAt, .providerEvidence.generatedAt' "$report")
jq -e '
  . as $r |
  all(.expectedNodes[]; . as $n |
    ([ $r.sources[] | select(.auditMode == "node-local" and .nodeName == $n)] | length) == 1
    and all(["3.1.3","3.2.1","3.2.2"][]; . as $id |
      ([ $r.results[] | select(.id == $id and .scope == $n and
         (.status == "PASS" or .status == "FAIL") and
         any(.evidence[]?; .name == "etc/kubernetes/kubelet/config.json")) ] | length) == 1))
' "$report" >/dev/null || die 'Missing supported AL2023 kubelet evidence for all workers'
jq -e '
  . as $r |
  ([ $r.results[] | select(.method == "aws_eks_describe_cluster" or
    .method == "aws_eks_access_entries" or .method == "aws_ecr_image_scanning") ] | length) as $count |
  ($r.sources | map(select(.auditMode == "aws-provider"))[0].findings == $count)
  and all(["2.1.1","5.4.1","5.4.2"][]; . as $id |
    ([ $r.results[] | select(.id == $id and .method == "aws_eks_describe_cluster" and
      (.status == "PASS" or .status == "FAIL" or .status == "MANUAL") and (.evidence | length > 0))] | length) == 1)
' "$report" >/dev/null || die 'AWS provider findings incomplete'
echo 'PASS (PREVIEW ONLY): EKS sources, node evidence and score-withholding validated'

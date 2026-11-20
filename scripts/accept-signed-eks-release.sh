#!/usr/bin/env bash
# Verify release artifacts before touching an existing EKS deployment.
# No deployment or AWS mutation is performed by this script.
set -Eeuo pipefail
umask 077
die(){ echo "Release acceptance: $*" >&2; exit 1; }
release='' assets='' context='' cluster='' region='' output='' verify_only=false
while (($#)); do
  case "$1" in
    --release|--assets|--context|--cluster-name|--region|--output-dir)
      (($# >= 2)) || die "Missing value for $1"
      case "$1" in
        --release) release=$2;; --assets) assets=$2;; --context) context=$2;;
        --cluster-name) cluster=$2;; --region) region=$2;; --output-dir) output=$2;;
      esac
      shift 2;;
    --verify-only) verify_only=true; shift;;
    *) die "Unknown argument $1";;
  esac
done
[[ "$release" =~ ^v[0-9]+[.][0-9]+[.][0-9]+(-[0-9A-Za-z.-]+)?$ && -d "$assets" ]] || die 'Provide --release TAG --assets DIRECTORY'
[[ "$(uname -s)" == Linux ]] || die 'Linux required'
case "$(uname -m)" in x86_64) arch=amd64;; aarch64|arm64) arch=arm64;; *) die 'Unsupported architecture';; esac
for tool in cosign jq sha256sum tar; do command -v "$tool" >/dev/null || die "$tool required"; done
archive="kubecis_linux_${arch}.tar.gz"
identity="https://github.com/sm-joe/kube-cis-audit/.github/workflows/release.yml@refs/tags/$release"
for file in "$archive" SHA256SUMS release-manifest.json; do
  [[ -s "$assets/$file" && -s "$assets/$file.sigstore.json" ]] || die "Missing signed asset $file"
  cosign verify-blob "$assets/$file" --bundle "$assets/$file.sigstore.json" \
    --certificate-identity "$identity" --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' >/dev/null || die "Signature rejected: $file"
done
expected="$(awk -v name="$archive" '$2==name {print $1}' "$assets/SHA256SUMS")"
actual="$(sha256sum "$assets/$archive")"; actual="${actual%% *}"
[[ "$expected" =~ ^[a-f0-9]{64}$ && "$expected" == "$actual" ]] || die 'Archive checksum mismatch'
jq -e --arg tag "$release" '.schemaVersion==1 and .version==$tag and (.commit|test("^[a-f0-9]{40}$")) and (.image|test("^ghcr[.]io/sm-joe/kubecis@sha256:[a-f0-9]{64}$"))' "$assets/release-manifest.json" >/dev/null || die 'Manifest identity mismatch'
image="$(jq -r .image "$assets/release-manifest.json")"
cosign verify "$image" --certificate-identity "$identity" --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' >/dev/null || die 'Image signature rejected'
stage="$(mktemp -d)"
trap 'rm -rf -- "$stage"' EXIT
tar -xzf "$assets/$archive" -C "$stage" --no-same-owner --no-same-permissions
[[ "$("$stage/kubecis" version)" == "$release" ]] || die 'Binary version mismatch'
bash "$stage/scripts/test-linux-eks.sh" --packaged >/dev/null
echo "Verified signed release $release: $image"
if $verify_only; then exit 0; fi
[[ -n "$context" && -n "$cluster" && -n "$region" && -n "$output" ]] || die 'Live mode requires --context --cluster-name --region --output-dir'
[[ ! -e "$output" ]] || die 'Use a new output directory to avoid mixing evidence'
command -v kubectl >/dev/null || die 'kubectl required'
ds="$(kubectl --context "$context" -n kube-system get daemonset kubecis-node-auditor -o json)"
jq -e --arg image "$image" '
  .spec.template.spec as $spec |
  ([ $spec.containers[], $spec.initContainers[]? ] | length) == 3 and
  all($spec.containers[], $spec.initContainers[]?; .image==$image) and
  .status.observedGeneration==.metadata.generation and
  .status.desiredNumberScheduled>0 and
  .status.numberReady==.status.desiredNumberScheduled and
  .status.updatedNumberScheduled==.status.desiredNumberScheduled
' <<<"$ds" >/dev/null || die 'DaemonSet is not ready on the signed digest; deploy the verified release first'
pods="$(kubectl --context "$context" -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json)"
jq -e --arg image "$image" '
  (.items|length)>0 and all(.items[];
    .status.phase=="Running" and
    all(.spec.containers[], .spec.initContainers[]?; .image==$image) and
    all(.status.containerStatuses[]; .ready==true))
' <<<"$pods" >/dev/null || die 'Node auditor Pods do not match the signed image'
bash "$stage/scripts/collect-eks.sh" --context "$context" --cluster-name "$cluster" --region "$region" --binary "$stage/kubecis" --output-dir "$output"
jq -e --arg version "$release" '.buildVersion==$version and .compliance.compliancePercent==null' "$output/consolidated.json" >/dev/null || die 'Report version or score mismatch'
report_hash="$(sha256sum "$output/consolidated.json")"; report_hash="${report_hash%% *}"
jq -e --arg hash "$report_hash" '.sourceReportSha256==$hash and (.scoreStatus|startswith("WITHHELD"))' "$output/final-assessment.json" >/dev/null || die 'Final dossier is not bound to the report'
jq -n --arg tag "$release" --arg commit "$(jq -r .commit "$assets/release-manifest.json")" --arg image "$image" --arg archiveSHA256 "$actual" --arg reportSHA256 "$report_hash" --arg timestamp "$(date -u +%FT%TZ)" \
  '{schemaVersion:1,release:$tag,commit:$commit,image:$image,archiveSha256:$archiveSHA256,reportSha256:$reportSHA256,generatedAt:$timestamp,status:"OPERATOR_RUNTIME_ACCEPTANCE_PREVIEW",certificationClaim:false}' > "$output/release-acceptance.json"
echo 'PASS: signed artifact and live collection acceptance (not independent CIS conformance)'

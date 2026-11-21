# KubeCIS

**Native Kubernetes security auditing | Linux | Amazon EKS**

[![CI](https://github.com/sm-joe/kube-cis-audit/actions/workflows/ci.yml/badge.svg)](https://github.com/sm-joe/kube-cis-audit/actions/workflows/ci.yml)
[![Security](https://github.com/sm-joe/kube-cis-audit/actions/workflows/security.yml/badge.svg)](https://github.com/sm-joe/kube-cis-audit/actions/workflows/security.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

KubeCIS is an open-source Go security scanner that assesses Kubernetes clusters, Linux worker nodes and Amazon EKS configurations against versioned CIS benchmark profiles. It gathers evidence, identifies configuration weaknesses and produces actionable reports without automatically modifying your workloads.

## Features

| Area | Capabilities |
| --- | --- |
| Cluster | RBAC, service accounts, workload security, admission and network policies |
| Nodes | File permissions, ownership, kubelet authentication and configuration |
| EKS | Control-plane logging, endpoint exposure, IAM access entries, CNI, ECR and subnet routing |
| Reports | JSON and readable text, remediation guidance, baseline comparisons |
| Evidence | Multi-node consolidation, freshness and identity validation, manual review |
| Distribution | Signed Linux AMD64/ARM64 binaries and container images |

## Supported environments

| Profile | Scope |
| --- | --- |
| cis-eks-2.1.0 | 47 control IDs; managed EC2 Linux workers, including AL2023 |
| cis-2.1.0 | Generic Kubernetes native starter profile |
| cis-2.0.1 | Legacy native Kubernetes profile |

See [support matrix](docs/SUPPORT-MATRIX.md) for supported versions and limitations. Checks requiring operator judgment or unavailable evidence are reported separately; coverage is not a compliance percentage.

## Installation

Get the Linux archive from [GitHub Releases](https://github.com/sm-joe/kube-cis-audit/releases). Requires Bash, GNU coreutils, tar, jq, kubectl and Cosign for verification. For EKS, install AWS CLI v2 and configure an authorized identity.

```bash
VERSION=v1.0.0
case "$(uname -m)" in
  x86_64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported architecture"; exit 1 ;;
esac
ARCHIVE="kubecis_linux_$ARCH.tar.gz"
BASE="https://github.com/sm-joe/kube-cis-audit/releases/download/$VERSION"
mkdir -p kubecis-assets
cd kubecis-assets
for file in "$ARCHIVE" "$ARCHIVE.sigstore.json" SHA256SUMS SHA256SUMS.sigstore.json release-manifest.json release-manifest.json.sigstore.json; do
  curl -fL "$BASE/$file" -o "$file" || exit 1
done
IDENTITY="https://github.com/sm-joe/kube-cis-audit/.github/workflows/release.yml@refs/tags/$VERSION"
for file in "$ARCHIVE" SHA256SUMS release-manifest.json; do
  cosign verify-blob "$file" --bundle "$file.sigstore.json" \
    --certificate-identity "$IDENTITY" \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com || exit 1
done
sha256sum -c --ignore-missing SHA256SUMS || exit 1
mkdir -p kubecis
tar -xzf "$ARCHIVE" -C kubecis
cd kubecis
./kubecis version
```

## Scan a Kubernetes cluster

```bash
kubectl config current-context
mkdir -p reports
./kubecis scan --benchmark auto --target policies --output json \
  --output-file reports/cluster.json --exit-code-on-fail=false
./kubecis render --input reports/cluster.json --output-file reports/results.txt
cat reports/results.txt
```

Remove the exit-code override to use findings as a CI gate. An API scan cannot measure host-file permissions without node evidence.

## Scan Amazon EKS

Configure the kubeconfig for your target cluster and use the same signed release for the CLI and node auditor. Review the supplied manifests and RBAC before deployment.

```bash
CLUSTER=my-eks-cluster
REGION=us-east-1
aws eks update-kubeconfig --name "$CLUSTER" --region "$REGION"
CONTEXT="$(kubectl config current-context)"
IMAGE="$(jq -er '.image' ../release-manifest.json)"
cosign verify "$IMAGE" --certificate-identity "$IDENTITY" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com || exit 1
bash scripts/deploy-eks.sh --context "$CONTEXT" --image "$IMAGE"
bash scripts/collect-eks.sh --context "$CONTEXT" --cluster-name "$CLUSTER" \
  --region "$REGION" --binary ./kubecis --output-dir reports/eks
```

See [EKS deployment](deploy/eks/README.md), [collection](docs/LINUX-EKS-COLLECTION.md) and [security](SECURITY.md). The collector rejects unsupported compute modes, mismatched identities and incomplete required node evidence.

## Understand the results

| Status | Meaning |
| --- | --- |
| PASS | Observed evidence meets the check criterion |
| FAIL | Observed evidence violates the criterion |
| WARN | Investigate a potential issue |
| MANUAL | Operator assessment needed |
| NOT_ASSESSED | Insufficient or unsupported evidence |

| Report | Purpose |
| --- | --- |
| cluster.json / node-*.json / provider.json | Source evidence |
| consolidated.json | Combined assessment |
| results.txt | Readable findings and remediation |
| review-template.json | Manual review candidates |
| final-assessment.json / .txt | Final dossier |

```bash
./kubecis finalize --input reports/eks/consolidated.json --output-dir reports/eks
```

Manual review is optional and never rewrites original automated findings. See command help for review and comparison options.

## CLI reference

| Command | Function |
| --- | --- |
| scan | API assessment |
| node-audit | Node assessment |
| node-metadata | Bounded metadata collection |
| enrich-node | Effective kubelet evidence |
| consolidate | Merge validated sources |
| render | Generate text output |
| compare | Compare assessments |
| review-init / review | Prepare and record manual decisions |
| finalize | Generate final dossier |

## Build from source

```bash
git clone https://github.com/sm-joe/kube-cis-audit.git
cd kube-cis-audit
go mod download
go vet ./...
go test ./... -count=1
CGO_ENABLED=0 go build -trimpath -o kubecis ./cmd/kubecis
./kubecis version
```

Use the Go toolchain pinned in go.mod. Buildx and QEMU are needed only for local multi-architecture image builds; published assets require no local container build.

## Security and support

Use scoped Kubernetes and AWS access, protect reports, and verify release signatures. See [SECURITY.md](SECURITY.md) for the security model and vulnerability disclosure process. Report ordinary bugs through [GitHub Issues](https://github.com/sm-joe/kube-cis-audit/issues).

[Apache 2.0](LICENSE) · [Changelog](CHANGELOG.md) · [Release scope](docs/RELEASE-SCOPE.md) · [Attribution](ATTRIBUTION.md)

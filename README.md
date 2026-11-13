# KubeCIS Native

**Linux-native Kubernetes security assessment with Amazon EKS evidence and auditable review workflows.**

[![CI](https://github.com/sm-joe/kube-cis-audit/actions/workflows/ci.yml/badge.svg)](https://github.com/sm-joe/kube-cis-audit/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

KubeCIS Native is an open-source, Go-based Kubernetes and Linux-node security auditor. It performs read-only Kubernetes API checks, evaluates explicitly collected Linux kubelet and host metadata, supports provider-aware Amazon EKS findings, and produces structured reports with human review attestations.

**Designed for Linux/Unix environments and Kubernetes containers.** Its native scanning engine does not depend on the kube-bench executable, kube-bench configuration corpus or runtime image. The project is not a vulnerability scanner, admission controller or automatic remediation service.

> **Benchmark accuracy:** The `cis-eks-2.1.0` assessment profile is **provisional**, not a certified or fully verified reproduction of the official benchmark. PASS and FAIL apply to specific implemented checks; MANUAL and NOT_ASSESSED are not silently counted as PASS. The official compliance score remains **WITHHELD**, including after manual reviews.

## Capabilities

| Component | Function |
| --- | --- |
| Kubernetes API auditor | Read-only workloads, namespace, RBAC, service account, Pod Security Admission and network-policy assessment |
| Linux node auditor | Host-file permission checks and allowlisted kubelet configuration evidence, including EKS AL2023 nodeadm sources |
| EKS provider evidence | Separate read-only AWS inventory for control-plane logging, API endpoint exposure, IAM access entries, ECR image scanning and VPC CNI configuration |
| Consolidation | Combine compatible Kubernetes, worker-node and provider reports; reject stale, incomplete or mismatched collection inputs |
| Reporting | JSON, kube-bench-style text, stable finding identities and baseline comparisons |
| Human reviews | SHA-256-bound review candidates and APPROVED / REJECTED / EXCEPTION decisions |
| Final dossier | One JSON and text assessment preserving original automated results and separating reviewer judgments |

### Supported deployment scope

KubeCIS runs against Kubernetes clusters through their APIs. Linux node checks require authorized access to the node's relevant host metadata; API-only scans cannot determine host-file permissions.

The EKS node-auditor manifests target **EKS managed EC2 Linux nodes**. EKS control-plane nodes are AWS-managed and not mounted or scanned as customer hosts. Fargate, EKS Auto Mode and self-managed worker nodes are not included in the strict managed-node collection contract.

## Prerequisites

- Linux distribution with a POSIX-compatible shell and standard tooling.
- Go **1.27.4** for source builds (see [go.mod](go.mod)).
- `kubectl` and a Kubernetes kubeconfig with permitted read access.
- Docker or Podman to build/run the Linux image, or access to a published GHCR image.
- For Amazon EKS: AWS CLI v2 with relevant IAM permissions, an EKS kubeconfig, and appropriate Kubernetes authorization to install the auditor's resources.

## Quick start: Linux CLI

Clone and build:

```bash
git clone https://github.com/sm-joe/kube-cis-audit.git
cd kube-cis-audit

go version
go mod download
go vet ./...
go test ./... -count=1
CGO_ENABLED=0 go build -trimpath -o ./kubecis ./cmd/kubecis
./kubecis version
```

Use the native Kubernetes API scanner:

```bash
kubectl config current-context
kubectl get nodes

mkdir -p reports
./kubecis scan \
  --engine native \
  --benchmark auto \
  --target policies \
  --output json \
  --output-file reports/cluster.json \
  --exit-code-on-fail=false

./kubecis render \
  --input reports/cluster.json \
  --output-file reports/results.txt

cat reports/results.txt
```

The `--exit-code-on-fail=false` option permits security findings to be reported without treating a detected FAIL as a CLI execution error. It does not hide failed checks.

Inspect available CLI flags with `./kubecis <command> --help`. `--engine upstream` and `--engine hybrid` are not supported.

## Containerized operation

Build a Linux image locally:

```bash
docker build --build-arg VERSION=local-dev -t kubecis-native:local .
docker run --rm kubecis-native:local version
```

For an API scan, supply kubeconfig or an in-cluster service account with carefully scoped access. An external kubeconfig may reference credential helpers or certificates not present in the container; make those available securely, rather than copying secrets into an image. In-cluster deployment examples are under [deploy/](deploy/).

### Amazon EKS deployment

For Linux/Unix operators, [the digest-pinned EKS deployment helper](docs/LINUX-EKS-DEPLOY.md) provides a fail-closed installation entry point. It does **not** replace the currently unverified Linux provider evidence collection workflow.

```bash
bash scripts/test-deploy-eks.sh
bash scripts/deploy-eks.sh --context YOUR_EKS_CONTEXT \\
  --image ghcr.io/sm-joe/kubecis@sha256:YOUR_VERIFIED_64_CHARACTER_DIGEST
```

The digest is an example placeholder, not a deployable image. Obtain a real published digest before running the helper.


Use a deliberately selected cluster context and confirm AWS/Kubernetes identity:

```bash
aws sts get-caller-identity
aws eks update-kubeconfig --name my-eks-cluster --region us-east-1
kubectl config current-context
kubectl get nodes -L eks.amazonaws.com/nodegroup
```

The EKS DaemonSet is version-coupled to the local scanner. **Do not deploy its placeholder image as-is.** Build and publish a Linux container from the same Git commit as the CLI, or use an existing published matching build, and set the images for all three DaemonSet containers:

```bash
# Replace with a published image built from the matching source commit.
IMAGE="ghcr.io/sm-joe/kubecis:cis210-eks135-<40-character-commit-sha>"

kubectl apply -k deploy/eks
kubectl -n kube-system set image daemonset/kubecis-node-auditor \
  "node-metadata-init=$IMAGE" \
  "node-auditor=$IMAGE" \
  "node-metadata-helper=$IMAGE"
kubectl -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=300s
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
```

These commands **create or update KubeCIS deployment resources**, including RBAC and a host-metadata helper. Review [EKS deployment manifests](deploy/eks/) and their access requirements before deployment. Assessments do not intentionally change the cluster's security configuration.

**Current limitation:** The repository's complete EKS collection/orchestration and release-acceptance workflow is not yet available as a verified POSIX-shell runner. Full provider-plus-node consolidation on Linux has **not been validated end to end**. The Go CLI and Linux node-auditor image are Linux-native, but the commands above should **not** be presented as a complete EKS provider assessment or as an equivalent to the proven integrated collection workflow. Linux/Unix support is the product's target; the remaining orchestration portability work is separate from the scanning engine.

## Reporting and review

KubeCIS can generate a reviewable dossier from a consolidated report produced by its native collection/consolidation process:

```bash
./kubecis review-init \
  --input reports/eks/consolidated.json \
  --output-file reports/eks/review-template.json

# An authorized reviewer supplies evidence-bound decisions in JSON.
./kubecis review \
  --input reports/eks/consolidated.json \
  --decisions reports/eks/review-decisions.json \
  --output-file reports/eks/review-validated.json

./kubecis finalize \
  --input reports/eks/consolidated.json \
  --decisions reports/eks/review-decisions.json \
  --output-dir reports/eks
```

To produce a dossier with unresolved manual items, omit `--decisions` from `finalize`:

```bash
./kubecis finalize \
  --input reports/eks/consolidated.json \
  --output-dir reports/eks
```

Typical report files:

| File | Description |
| --- | --- |
| `cluster.json` | Kubernetes API findings |
| `node-<node>.json` | Linux node-local findings |
| `provider.json` | Read-only AWS provider inventory |
| `consolidated.json` | Original, source-of-truth assessment |
| `results.txt` | Grouped kube-bench-style summary and remediations |
| `review-template.json` | Manual controls with supporting evidence, observations and guidance |
| `review-decisions.json` | Decisions made by an authorized reviewer |
| `review-validated.json` | Validated review results tied to the source SHA-256 |
| `final-assessment.json` | Complete assessment, automated statuses and reviewer decisions |
| `final-assessment.txt` | Human-readable final dossier |

**Review integrity:** Every decision requires a matching manual control, reviewer identity, evidence reference, justification and valid timestamp. A decision for a non-manual control or from another scan is rejected. Human approval never changes the underlying PASS/FAIL/NOT_ASSESSED result.

**Audit-grade limitation:** The current review validator verifies that external evidence references are present, but does not independently retrieve or authenticate those external records. Text such as `approved` is not independently verifiable proof. An assessment labeled `REVIEWED` is not an assertion of CIS compliance.

## Findings and coverage interpretation

| Status | Meaning |
| --- | --- |
| `PASS` | The implemented check observed a compliant configuration for its defined condition |
| `FAIL` | The implemented check observed a condition that does not meet its defined criterion |
| `WARN` | A security concern that needs investigation |
| `MANUAL` | Additional human assessment or a documented exception is required |
| `NOT_ASSESSED` | Evidence, evaluator coverage or reliable applicability is insufficient |
| `REVIEWED` | All manual items received recorded reviewer decisions, regardless of their conclusions |
| `WITHHELD` | Official compliance scoring is not available for this provisional benchmark |

`coveragePercent` is a decisive-outcome fraction among configured scored checks, **not** the percentage of official CIS controls implemented, not a compliance percentage, and not a benchmark certification.

## Native command reference

| Command | Role |
| --- | --- |
| `kubecis version` | Print executable build identity |
| `kubecis scan` | Kubernetes API scanner |
| `kubecis node-audit` | Linux host and kubelet checks |
| `kubecis node-metadata` | Bounded host metadata collection |
| `kubecis enrich-node` | Add validated effective kubelet `/configz` evidence |
| `kubecis consolidate` | Validate and merge fresh API, node and provider evidence |
| `kubecis render` | Generate grouped readable report |
| `kubecis compare` | Compare two reports using stable finding identities |
| `kubecis review-init` | Prepare manual assessment candidates |
| `kubecis review` | Validate reviewer attestations |
| `kubecis finalize` | Produce final JSON and text dossier |

## Security and permissions

Use least-privilege access for Kubernetes API reads. AWS EKS evidence collection may require `eks:DescribeCluster`, `eks:ListAccessEntries`, `eks:DescribeAccessEntry`, `eks:ListAssociatedAccessPolicies`, `eks:DescribeAddon`, `ecr:GetRegistryScanningConfiguration` and `ecr:DescribeImageScanFindings`. Some metadata enrichment uses `nodes/proxy`, a sensitive Kubernetes permission; grant it only where reviewed and necessary. The Kubernetes deployment requires additional privileges to create its own ServiceAccounts, RBAC and DaemonSet.

Store kubeconfigs, identity metadata, scan reports and reviewer records in access-restricted locations. Do not commit credentials, private cluster inventories or attestations into public source control. Verify published image tags, digests and provenance before deployment. Missing evidence must not be interpreted as a secure default.

## CI, project structure and support

Go CI, security scans and benchmark contract tests run under [GitHub Actions](.github/workflows/). Release packaging and smoke tests exist, but **the currently published automated artifact workflow is not a general Linux/Unix release installer**. Users building on Linux should use the native Go build or the Linux container image, with matching source/image versions.

```text
cmd/kubecis/          Go CLI, rendering and review engine
internal/audit/       Native checks, consolidation, node evidence and AWS attachment
benchmarks/           Native profile definitions
deploy/eks/           EKS Kubernetes deployment
deploy/kind-native/   Local Kubernetes/Kind examples
scripts/              Collection and validation helpers
docs/                 Security, validation and release-scope notes
.github/workflows/    CI, security and publishing automation
```

See [native validation](docs/NATIVE-VALIDATION.md), [release scope](docs/RELEASE-SCOPE.md), and [security guidance](docs/SECURITY.md). Report problems through [GitHub Issues](https://github.com/sm-joe/kube-cis-audit/issues), with sanitized logs and build metadata.

## License

[Apache License 2.0](LICENSE). Provided **as is**, without warranty. KubeCIS findings and review attestations must be assessed by authorized operators before making production security changes.

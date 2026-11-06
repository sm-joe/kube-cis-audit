# KubeCIS Native

**Native Kubernetes security assessment, AWS EKS evidence collection, and reviewable security reports.**

[![CI](https://github.com/sm-joe/kube-cis-audit/actions/workflows/ci.yml/badge.svg)](https://github.com/sm-joe/kube-cis-audit/actions/workflows/ci.yml)
[![Windows EKS build](https://github.com/sm-joe/kube-cis-audit/actions/workflows/eks-135-windows-binary.yml/badge.svg)](https://github.com/sm-joe/kube-cis-audit/actions/workflows/eks-135-windows-binary.yml)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

KubeCIS Native is a **Go-based, self-hosted security assessment tool** for Kubernetes clusters and Linux nodes. It collects Kubernetes API evidence, examines allowlisted node security settings, enriches Amazon EKS findings with AWS read-only API inventory, and produces consolidated JSON, kube-bench-style text, and a human-reviewable assessment dossier.

**No kube-bench runtime dependency.** Assessments run through the project's native Go implementation. KubeCIS is not a vulnerability scanner, admission controller, automatic hardening engine, or independently certified CIS benchmark implementation.

> **Release scope:** The Amazon EKS `cis-eks-2.1.0` profile is **provisional**. The program reports automated PASS/FAIL separately from MANUAL and NOT_ASSESSED findings. A successful scan, 100% review completion, or high decisive-coverage percentage **does not establish official CIS compliance**. The official compliance percentage stays **WITHHELD** until mapping, applicability, evidence, and release scope are fully verified. Do not report preview percentages as a CIS score.

## Features

| Capability | What it provides |
| --- | --- |
| Kubernetes API checks | Read-only evaluation of configured workload, service account, RBAC, namespace, Pod Security Admission and network-security observations |
| Linux node assessment | Native host-file permission and kubelet security-setting checks using allowlisted metadata |
| Amazon EKS integration | Read-only inventory of control-plane logging, endpoint configuration, authentication/access entries, VPC CNI and ECR scanning evidence |
| EKS managed worker nodes | Per-node results for supported **EKS managed EC2 Linux** nodes; Kubernetes control-plane host access is not assumed |
| Evidence provenance | Source and build identity, timestamps, per-node scope, method, evidence and expected values; consolidation rejects stale or mismatched inputs |
| Reporting | Machine-readable JSON, kube-bench-style `results.txt`, stable control/finding identities and cross-report comparison |
| Human review | Evidence-backed review candidates and interactive APPROVED / REJECTED / EXCEPTION decisions bound to the immutable scan SHA-256 |
| Final assessment | `final-assessment.json` and `final-assessment.txt` retain automated statuses while showing reviewer attestations and pending controls separately |
| Release gates | Go tests, vet, PowerShell collection tests, Windows executable smoke test, version-matched Linux image and strict provider/source completeness checks |

## Quick start — Amazon EKS on Windows

### Requirements

- **Windows PowerShell 5.1+** (PowerShell 7 is also suitable); commands below use PowerShell.
- **AWS CLI v2** with valid credentials, `aws eks update-kubeconfig` access, and read-only access to the relevant EKS/ECR inventory.
- **kubectl** configured for the target Amazon EKS cluster.
- A supported Amazon EKS cluster with **managed EC2 Linux worker nodes** and sufficient permissions to install KubeCIS's Kubernetes ServiceAccount/RBAC and node-auditor DaemonSet in `kube-system`.
- Network access to the specific published `ghcr.io/sm-joe/kubecis` image.
- If building from source instead of using the packaged executable: **Go 1.27.3** (see [go.mod](go.mod)).

**Supported EKS collection is not equivalent to all EKS compute types.** Fargate, EKS Auto Mode and self-managed worker fleets are outside the strict managed-EC2 consolidation path. Managed EKS control-plane host paths cannot be audited by customer node agents.

### 1. Obtain a version-matched package

Download the `kubecis-eks-135-windows-amd64` artifact from a successful [Windows EKS build workflow](https://github.com/sm-joe/kube-cis-audit/actions/workflows/eks-135-windows-binary.yml). Expand the downloaded workflow artifact and its packaged ZIP if GitHub presents nested archives. The **package directory** should contain:

```text
kubecis-eks-135/
  kubecis.exe
  IMAGE.txt
  SHA256SUMS
  benchmarks/
  deploy/eks/
  scripts/
    Run-KubeCIS-EKS.ps1
    Collect-ConsolidatedEKSReport.ps1
    Collect-EKSProviderEvidence.ps1
    Validate-EKSRelease.ps1
    Review-KubeCIS-EKS.ps1
```

Use the executable and `IMAGE.txt` from **the same build**. Do **not** reuse an old `kubecis.exe` or change the container tag independently; consolidation rejects mixed-version evidence.

### 2. Authenticate and verify cluster identity

```powershell
aws sts get-caller-identity
aws eks describe-cluster --region us-east-1 --name my-eks-cluster --query "cluster.status"
kubectl config current-context
kubectl get nodes -L eks.amazonaws.com/nodegroup
```

Confirm the AWS account, cluster name, region and managed EC2 worker nodes before running any deployment command. KubeCIS independently verifies EKS provider and Kubernetes API endpoint identity.

### 3. Run a fresh assessment

From the extracted **`kubecis-eks-135`** directory:

```powershell
.\scripts\Run-KubeCIS-EKS.ps1 `
  -ClusterName "my-eks-cluster" `
  -Region "us-east-1"
```

The runner verifies the local binary and `IMAGE.txt` build identity, updates the target kubeconfig, **applies KubeCIS resources** under `deploy/eks`, updates all node-auditor images to the matching version, waits for rollout, checks the binary version inside the auditor Pods, collects fresh Kubernetes/node/AWS evidence, enforces the preview release gate, and renders the assessment.

**Deployment changes Kubernetes resources**; the subsequent audit collects configuration without intentionally remediating the cluster. Review Kubernetes RBAC, host mounts, image provenance and change-control policies before installation. The KubeCIS node metadata helper has read-only host access but may require elevated Linux permissions to inspect restricted kubelet metadata.

To choose another output directory:

```powershell
.\scripts\Run-KubeCIS-EKS.ps1 `
  -ClusterName "my-eks-cluster" -Region "us-east-1" `
  -OutputDir "D:\security-reports\eks"
```

### 4. Review the generated reports

```powershell
Get-Content .\reports\eks\results.txt
Get-Content .\reports\eks\final-assessment.txt
```

Default files:

| File | Purpose |
| --- | --- |
| `cluster.json` | Fresh Kubernetes API findings |
| `node-<node>.json` | Individual managed Linux worker findings |
| `provider.json` | Bounded AWS EKS/ECR provider evidence inventory |
| `consolidated.json` | Immutable source report for the review workflow |
| `results.txt` | Grouped kube-bench-style text with remediations |
| `review-template.json` | Manual review candidates, attached evidence, control titles and guidance |
| `final-assessment.json` | Original source plus review state, decisions and separate automated/reviewer summary |
| `final-assessment.txt` | Readable finalized dossier, even if some reviews are pending |
| `review-decisions.json` | Saved reviewer decisions (created by the interactive review step) |
| `review-validated.json` | Validated review attestation output (created by the interactive review step) |

`final-assessment` is a **report artifact**, not a declaration that all controls passed. Its status may be `INCOMPLETE` or `REVIEWED`; automated FAIL and NOT_ASSESSED remain visible in either case.

## Human review and evidence-backed attestations

After a scan, an authorized reviewer can work through pending controls **without scanning the cluster again**:

```powershell
.\scripts\Review-KubeCIS-EKS.ps1 `
  -Reviewer "SecurityReviewer" `
  -ShowEvidence
```

Choose **A** (APPROVED), **R** (REJECTED), **E** (EXCEPTION), **S** (SKIP), or **Q** (QUIT). Approved/rejected/exception decisions require the reviewer, a **real independently resolvable** evidence reference (for example a ticket, document or controlled review URL), a specific justification and UTC time. Decisions are saved incrementally. Re-running the review command resumes unreviewed items from the same scan. Skipped items remain pending.

For another report directory or explicit decisions file:

```powershell
.\scripts\Review-KubeCIS-EKS.ps1 `
  -Reviewer "SecurityReviewer" `
  -ReportDir "D:\security-reports\eks" `
  -DecisionsFile "D:\security-reports\eks\review-decisions.json" `
  -ShowEvidence
```

To validate and generate the dossier without interactive prompts:

```powershell
.\kubecis.exe review `
  --input .\reports\eks\consolidated.json `
  --decisions .\reports\eks\review-decisions.json `
  --output-file .\reports\eks\review-validated.json

.\kubecis.exe finalize `
  --input .\reports\eks\consolidated.json `
  --decisions .\reports\eks\review-decisions.json `
  --output-dir .\reports\eks
```

`finalize` can run **without** `--decisions`; all MANUAL controls then remain PENDING. Reviewer decisions never convert automated FAIL/NOT_ASSESSED to PASS, and never release a provisional compliance percentage. Decisions are bound to the original report's **SHA-256**; decisions from a different scan are rejected. If you rescan, collect new evidence and perform a fresh review; an old decision file is not automatically valid for a new source.

**Current attestation limitation:** The review validator checks that `evidenceRef` and `justification` are nonempty and that the reviewer/timestamp/control mapping is valid. It **does not independently fetch, authenticate or prove the contents of an external ticket or document**. Generic references such as `approved` are not audit-quality evidence even if accepted syntactically. A completed `REVIEWED` workflow is not an independent compliance certification.

## Understanding assessment status

| Result | Meaning |
| --- | --- |
| `PASS` | Specific configured native check passed with the available asserted evidence |
| `FAIL` | Check identified a configuration or policy condition that does not satisfy its configured criterion |
| `WARN` | Caution requiring investigation; do not interpret as PASS |
| `MANUAL` | Human verification or an approved exception is needed; recorded decisions are separate from the source result |
| `NOT_ASSESSED` | Necessary evidence, reliable evaluator or applicability information was unavailable |
| `REVIEWED` | Every MANUAL finding received a syntactically validated reviewer decision; it does not mean compliant |
| `WITHHELD` | An official CIS compliance score is not published |

The EKS `coveragePercent` is a **decisive PASS/FAIL fraction of currently scored starter-profile checks**, not the share of official CIS controls implemented and not a compliance percentage. The `cis-eks-2.1.0` control mapping and benchmark compatibility remain provisional. Do not use preview output as certification, or equate every marked `APPROVED` attestation with a PASS.

## Supported command-line interfaces

| Command | Description |
| --- | --- |
| `kubecis version` | Print the embedded build version |
| `kubecis scan` | Run native Kubernetes API checks; use `--engine native` |
| `kubecis node-audit` | Run Linux host/node checks, normally inside the matching node-auditor image |
| `kubecis node-metadata` | Produce bounded, allowlisted host metadata for node-audit |
| `kubecis enrich-node` | Enrich a matching node-local report with effective kubelet `/configz` facts |
| `kubecis consolidate` | Combine fresh reports and validate source identity and expected node inventory |
| `kubecis render` | Render JSON assessment in kube-bench-style text |
| `kubecis review-init` | Generate SHA-256-bound manual review candidates |
| `kubecis review` | Validate reviewer decisions against the original report |
| `kubecis finalize` | Produce final JSON and readable text dossier with separate automated and reviewer results |
| `kubecis compare` | Compare two native reports by stable finding identity |

For actual flags, run `.\kubecis.exe <command> --help`. The upstream `kube-bench` and legacy `--engine upstream/hybrid` modes are not supported.

## Build from source and run tests

```powershell
git clone https://github.com/sm-joe/kube-cis-audit.git
cd kube-cis-audit
go version   # 1.27.3 per go.mod
go mod download
go fmt ./...
go vet ./...
go test ./... -count=1
go build -trimpath -o .\kubecis.exe ./cmd/kubecis
.\kubecis.exe version
```

A locally built binary by itself is **not** an EKS deployment release. To scan node data through EKS, publish the Linux image from **the same commit** with matching embedded `audit.BuildVersion`; use the CI-produced ZIP/image pair where possible. See [EKS deployment details](deploy/eks/README.md), [native validation](docs/NATIVE-VALIDATION.md), and [release scope](docs/RELEASE-SCOPE.md).

## Required permissions and operational security

The collector uses `aws eks describe-cluster`, access-entry APIs, `aws eks describe-addon`, ECR scanning status APIs, and Kubernetes read-only list/get requests. A least-privilege AWS policy for complete inventory typically needs permission to call:

```text
eks:DescribeCluster
eks:ListAccessEntries
eks:DescribeAccessEntry
eks:ListAssociatedAccessPolicies
eks:DescribeAddon
ecr:GetRegistryScanningConfiguration
ecr:DescribeImageScanFindings
sts:GetCallerIdentity                 # identity preflight
```

AWS permissions must be scoped to the target account, cluster and repositories where the actions support resource-level authorization. The **installation** identity additionally needs Kubernetes authorization to apply the explicitly reviewed `deploy/eks` manifests. Optional kubelet `/configz` enrichment uses authenticated Kubernetes node proxy access (`nodes/proxy`), which is sensitive; grant only where approved. Do not grant cluster-admin merely to suppress an authorization error. Missing permissions or unavailable evidence should remain `MANUAL`/`NOT_ASSESSED`.

The agent does not intentionally modify cluster security settings or remediate detected findings. It does create/update its own Kubernetes deployment resources. Avoid committing kubeconfig files, temporary credentials, provider inventory, scan JSON or reviewer records to public source control. These may contain infrastructure metadata, principal names and operationally sensitive facts. Store reports and attestations in restricted storage with an appropriate retention policy. Treat `SHA256SUMS` as package integrity checking, **not** a digital publisher signature.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Missing `kubecis.exe`, `IMAGE.txt` or benchmark | Run inside the extracted `kubecis-eks-135` directory, not an older repo checkout |
| Local and DaemonSet build mismatch | Download a single workflow build; never combine binaries and images from different commits |
| `ImagePullBackOff` | Verify published GHCR tag, registry/network access, and node image pull permissions |
| No running node auditor | Verify DaemonSet rollout, supported managed Linux nodes, namespace, Pod Security exceptions, host mount access |
| AWS provider evidence failure | Check `aws sts get-caller-identity`, `eks:DescribeCluster`, cluster name, region, API endpoint identity, provider collector diagnostics |
| Missing `/configz` evidence | Check authenticated node proxy permission, kubelet endpoint, and explicit config availability; never assume defaults |
| `NOT_ASSESSED` for a kubelet control | Required effective setting/applicability could not be established; this is not a product crash |
| Review SHA-256 mismatch | The source report changed or decisions belong to another scan; do not copy approvals across runs |
| `REVIEWED` but score WITHHELD | Expected for a provisional benchmark, even with all manual attestations completed |
| Security finding remains FAIL | Investigate the actual EKS/Kubernetes setting. The tool will not silently remediate or mask it |

## Continuous integration and release criteria

The [CI workflow](.github/workflows/ci.yml) checks Go formatting, vet, tests and build. The [EKS Windows build workflow](.github/workflows/eks-135-windows-binary.yml) runs Go and PowerShell tests, packages the Windows CLI/manifests/scripts, publishes a **matching multi-architecture Linux image**, and smoke-tests the **packaged executable on Windows** with JSON-to-text and final-assessment generation. [Security workflows](.github/workflows/security.yml) provide additional gates.

A successful workflow validates those build and packaging contracts. It is **not** a live end-to-end test of every customer EKS configuration, complete proof of CIS mapping, external evidence verification, or a certificate of compliance.

## Project structure

```text
cmd/kubecis/           CLI, rendering, review and final assessment
internal/audit/        Native evaluators, evidence, node collectors, consolidation
benchmarks/            Native benchmark profiles and compatibility declarations
deploy/eks/            EKS Kubernetes manifests and deployment notes
deploy/kind-native/    Optional local Kind-based integration examples
scripts/               Windows collection, validation, scan and review runners
docs/                  Security, design, validation and release-scope documentation
.github/workflows/     Tests, security gates, binary/image packaging
```

## License and support

Distributed under the [Apache License, Version 2.0](LICENSE). See [security documentation](docs/SECURITY.md) for handling sensitive findings and [GitHub Issues](https://github.com/sm-joe/kube-cis-audit/issues) to report defects or request controls. Please include the CLI build version, Kubernetes version, provider, reproduction steps and **sanitized** logs; never attach credentials, kubeconfig tokens or unrestricted security inventories.

KubeCIS Native is provided **as is**. Review findings and change-control implications before applying remediations to production workloads.

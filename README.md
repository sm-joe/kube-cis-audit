# KubeCIS

**Read-only Kubernetes security auditing from your terminal.** KubeCIS evaluates Kubernetes API configuration and supported node-local settings, then produces human-readable findings and machine-readable JSON reports.

KubeCIS is an auditing tool, not a remediation controller. It does not change cluster resources, retrieve Secret values, or require cluster-admin for the supplied cluster-audit ServiceAccount.

> **Benchmark scope:** the included `cis-2.1.0` profile is a limited starter profile with a provisional Kubernetes 1.37 mapping. It is not the complete CIS Kubernetes Benchmark, is not CIS-certified, and must not be represented as a complete compliance assessment. Obtain and validate the official benchmark content and compatibility for your distribution. KubeCIS-specific extensions are identified separately from official CIS mappings.

## Install on Windows

Open PowerShell and run:

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/sm-joe/kube-cis-audit/main/install.ps1" -OutFile ".\install-kubecis.ps1"
.\install-kubecis.ps1
```

The installer downloads the latest Windows x64 release, installs `kubecis.exe` and its benchmark profiles under `$HOME\.kubecis\bin`, and adds that directory to your user PATH.

To install a specific version:

```powershell
.\install-kubecis.ps1 -Version 3.0.0
```

Open a new PowerShell window (or use the updated PATH in the current one), then verify:

```powershell
kubecis --version
kubecis --help
```

## Run your first audit

KubeCIS uses the current kubeconfig context by default. It does not create, delete, or modify your Kubernetes cluster.

```powershell
kubectl config current-context
kubecis scan --benchmark auto --target all --output json --output-file .\kubecis-report.json
```

The command prints a non-zero exit code when one or more checks are `FAIL`, while still writing the complete report. This is useful for CI/CD quality gates.

For an observational scan that should not fail the process based on findings:

```powershell
kubecis scan --benchmark auto --target all --output json --output-file .\kubecis-report.json --exit-code-on-fail=false
```

Select another context or kubeconfig:

```powershell
kubecis scan --context kind-kind --benchmark auto --target all
kubecis scan --kubeconfig "$HOME\.kube\config" --context production --output json --output-file .\reports\production.json
```

Useful commands:

```powershell
kubecis scan --help
kubecis version
kubecis scan --target policies --output text
kubecis scan --benchmark cis-2.1.0 --target all --output json
```

## Run with Podman

Build locally:

```powershell
podman build --build-arg VERSION=3.0.0-dev -t kubecis:3.0.0-dev .
```

Run with your kubeconfig mounted read-only:

```powershell
podman run --rm `
  -v "$HOME\.kube:/home/nonroot/.kube:ro" `
  -e KUBECONFIG=/home/nonroot/.kube/config `
  kubecis:3.0.0-dev scan --benchmark auto --target all --output json --exit-code-on-fail=false
```

The container runs as UID/GID 65532, uses a Distroless non-root runtime, and includes the benchmark profiles. The host filesystem is not mounted by default; host-file checks require the separate node-audit deployment and its explicit read-only mounts. A standalone container can reach only API endpoints routable from its own network namespace, and kubeconfigs that require external `exec` credential plugins (such as `aws`, `az`, or `gcloud`) are not supported by this minimal image. For local Kind on Podman, use the in-cluster Job below: a kubeconfig endpoint such as `127.0.0.1` points to the container itself, not the Windows host. See [Container and E2E details](docs/DOCKER-AND-E2E.md).

## Run inside Kubernetes

The supplied Job uses the dedicated `kubecis` ServiceAccount and read-only ClusterRole. Review the manifests and image reference before applying:

```powershell
# If GHCR is private, create the pull secret first:
.\deploy\create-ghcr-pull-secret.ps1 -GitHubUsername "YOUR_GITHUB_USERNAME"

kubectl apply -f .\deploy\rbac.yaml
kubectl apply -f .\deploy\job.yaml
kubectl -n kube-system logs job/kubecis-audit
```

If the GHCR package is public, remove the `imagePullSecrets` stanza from the manifests and skip the pull-secret step. The Job is observational (`--exit-code-on-fail=false`): findings are emitted in the report without marking the Job itself failed. The RBAC role does not grant Secret access, Node inventory access, or workload mutation permissions.

For per-node host metadata checks, review and deploy `deploy/node-auditor-daemonset.yaml`. The main auditor remains unprivileged and has no ServiceAccount token. The separate metadata helper is narrowly scoped to the allowlisted kubelet metadata path and the minimal capability required by the design. Do not replace this boundary with a privileged auditor container.

## CLI reference

| Command | Purpose |
|---|---|
| `kubecis scan` | Cluster/API audit using kubeconfig or in-cluster credentials |
| `kubecis node-audit` | Node-local read-only host checks |
| `kubecis node-metadata` | Allowlisted metadata snapshot helper used by the DaemonSet |
| `kubecis version` / `kubecis --version` | Print the build version |

Important `scan` options:

| Option | Default | Meaning |
|---|---|---|
| `--context` | current context | Select a kubeconfig context |
| `--kubeconfig` | standard loading rules | Explicit kubeconfig path |
| `--benchmark` | `auto` | Benchmark profile ID or version-based selection |
| `--target` | `all` | `all`, `master`, `node`, or `policies` |
| `--namespace` | all namespaces | Limit supported namespaced checks |
| `--output` | `text` | `text` or `json` |
| `--output-file` | stdout | Write report to a file |
| `--exit-code-on-fail` | `true` | Return non-zero when findings include FAIL |

`--benchmark auto` selects only a profile with an explicit Kubernetes minor-version mapping. If no mapping exists, KubeCIS fails closed rather than silently applying an unrelated profile.

## Result semantics

- `PASS`: sufficient evidence demonstrates the configured requirement is met.
- `FAIL`: sufficient evidence demonstrates a violation.
- `WARN`: risk or incomplete configuration needs review; a definitive failure is not justified.
- `MANUAL`: human, provider or runtime verification is required.
- `NOT_ASSESSED`: evidence is unavailable, incomplete or unsupported.
- `INFO`: informational result, not a security verdict.

Compliance scoring is withheld when the assessment is materially incomplete. KubeCIS does not convert `WARN`, `MANUAL`, or `NOT_ASSESSED` into `PASS`.

## Security model and limitations

- Cluster audit uses read-only Kubernetes API operations.
- The supplied auditor ClusterRole grants only the required `get`/`list` operations on the inventoried resource types.
- Secret values are not retrieved.
- Node-local auditing is separate from API auditing and does not use Kubernetes API credentials.
- The node auditor runs non-root, drops all capabilities, uses a read-only root filesystem and RuntimeDefault seccomp, and has ServiceAccount token automount disabled.
- The metadata helper is a separate, narrowly scoped container. Its root identity and `CAP_DAC_READ_SEARCH` exception are intentional; do not broaden its mounts or capabilities.
- NetworkPolicy object presence does not prove effective isolation or dataplane enforcement.
- Managed control-plane host settings may be inaccessible and should remain `MANUAL` or `NOT_ASSESSED` where evidence is unavailable.
- The included profile is not a complete CIS Benchmark. Validate mappings and distribution-specific applicability before compliance use.

## Build from source

Requires Go 1.26 or later:

Run the consolidated Windows build, formatting, vet and test script:

```powershell
.\scripts\Build-KubeCIS.ps1
```

It creates `dist\kubecis.exe` only after the full test suite and `go vet` pass. To run the audit:

```powershell
.\dist\kubecis.exe scan --benchmark auto --target all --output json --output-file .\kubecis-report.json
```

For container builds, use Podman or Docker. The GitHub release workflow publishes Windows, Linux and macOS binaries for amd64/arm64 where applicable, SHA-256 checksums, and a multi-architecture GHCR image on version tags.

## Project

- Repository: https://github.com/sm-joe/kube-cis-audit
- Container: `ghcr.io/sm-joe/kubecis`
- CIS Kubernetes Benchmark: https://www.cisecurity.org/benchmark/kubernetes
- Reference implementation: https://github.com/aquasecurity/kube-bench

KubeCIS does not bundle or reproduce the full CIS Benchmark.

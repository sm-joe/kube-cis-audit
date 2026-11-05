# KubeCIS Native — independent Kubernetes security auditor

KubeCIS Native implements its own Kubernetes API and Linux host assessments in Go. The `kube-bench` executable, runtime invocation, upstream configuration corpus and runtime image dependency have been removed. This is **not** an independent reproduction of the full CIS Kubernetes Benchmark and **not** a certified CIS compliance tool. Existing starter benchmark profiles retain historical IDs for report compatibility and are explicitly provisional; compliance percentages remain gated by compatibility and assessment coverage.

## Amazon EKS (experimental managed-node integration)

EKS API scans auto-detect the `-eks-` Kubernetes server version and label
AWS-managed control-plane host checks `NOT_ASSESSED`. Worker-host checks
require a Linux node metadata collector; a Windows-hosted API scan cannot
determine kubelet file permissions. Scored results remain provisional.

The separate [EKS deployment and collection guide](deploy/eks/README.md)
covers EKS managed EC2 node groups and the fail-closed PowerShell collector.
Build the `3.1.0-native-eks-rc2` candidate CLI and publish the matching
`ghcr.io/sm-joe/kubecis:eks-provider-rc2` image before deploying. The
example does **not** deploy a control-plane host auditor or claim complete
EKS CIS compliance.

## Requirements

- Go **1.26+** (repository `go.mod` specifies `go 1.26.0`); Go 1.27 on Windows 11 is appropriate.
- Windows PowerShell for local API-only scanning; Linux container/Kind for node filesystem checks.
- Podman, Kind and kubectl for the optional in-cluster workflow.

## PowerShell build and tests

```powershell
$ErrorActionPreference = 'Stop'
go version
go mod download
go vet ./...
go test ./... -count=1
go build -trimpath -o ./kubecis.exe ./cmd/kubecis
./kubecis.exe version
./kubecis.exe scan --engine native --benchmark auto --target policies --output json --exit-code-on-fail=false
```

Run from the repository root so `benchmarks/` can be discovered. API-only scan does not need Linux. `--engine upstream` and `--engine hybrid` now fail with an actionable migration error. `--upstream-benchmark` is retained as a compatibility flag but rejects any non-default value.

## Native Linux node audit

Read-only host access is needed for node-specific checks. Do **not** run host audits against Windows paths. An example on Linux with a read-only mount:

```powershell
podman build -t localhost/kubecis-native:integrated-rc3.2 .
podman run --rm --user 0 --read-only --security-opt=no-new-privileges `
  --mount type=bind,src=/var/lib/kubelet,dst=/host/var/lib/kubelet,ro `
  localhost/kubecis-native:integrated-rc3.2 node-audit --node-name example-node --benchmark cis-2.1.0 `
  --host-root /host --target node --output json --exit-code-on-fail=false
```

The host path example runs on a **Linux** Podman host, not a Windows PowerShell Podman machine by assumption. In Kind, use `deploy/kind-native/` for local Kind deployment or the `deploy/` GHCR manifests after reviewing image names, host mounts, and RBAC. The metadata helper exports file metadata and narrowly allowlisted kubelet authentication/authorization fields. Other content checks remain `NOT_ASSESSED` when required evidence is unavailable. See [validation instructions](docs/NATIVE-VALIDATION.md).

## Findings and reports

- `scan`: Kubernetes API-backed policy, workload, RBAC, network and visible control-plane checks; host-file checks only with appropriate host mounts.
- `node-audit`: Linux-local host file permissions and direct kubelet YAML inspection.
- `node-metadata`: allowlisted host metadata and narrowly selected kubelet security settings, not the full configuration file.
- `compare`: stable logical finding IDs; absent findings and `NOT_ASSESSED` never imply remediation.
- JSON output preserves existing report field names (including historical `upstreamBenchmark`, which is now unused).

Native kubelet controls `KUBECIS-NATIVE-NODE-1` and `KUBECIS-NATIVE-NODE-2` examine explicit values of anonymous authentication and authorization mode. Missing settings or inaccessible files are `NOT_ASSESSED` rather than assumed to use a secure default. These controls are original KubeCIS checks and are not counted as official CIS scored controls.

## Security and limitations

See [security notes](docs/SECURITY.md), [release scope](docs/RELEASE-SCOPE.md), and [cleanup inventory](SOURCE-CLEANUP.md). Profiles are provisional and do not establish official CIS compliance. Real image-build and Kind E2E validation is required before promotion to a stable public release. No root project license is assigned by this archive; the copyright owner must select one.

### Consolidated report (integrated RC3)

Once the Podman/Kind auditors are deployed, use the supported PowerShell collection workflow:

```powershell
New-Item -ItemType Directory -Force .\dist | Out-Null
go build -trimpath -o .\dist\kubecis.exe ./cmd/kubecis
.\scripts\Collect-ConsolidatedKindReport.ps1 -ClusterName control-plane -Binary .\dist\kubecis.exe
```

The output `reports/consolidated.json` contains cluster, node and control-plane findings, node inventory, source provenance, and overall versus scored-control coverage. It **does not** claim complete CIS compliance. See `docs/CONSOLIDATED-RELEASE.md` and `RELEASE-READINESS-NATIVE-RC3.md` for safety constraints and status.

## Release source hygiene

This distribution intentionally excludes historical kube-bench source/configurations, local scan reports, binaries, Terraform state, Git internals, and milestone/patch artifacts. See `SOURCE-CLEANUP.md` for the inclusion policy. To test a clean source extraction, run `go test ./... -count=1` and `go vet ./...` with Go 1.26+.

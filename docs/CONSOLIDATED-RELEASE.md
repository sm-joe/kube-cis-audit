# Consolidation and production release gate

KubeCIS Native `consolidate` accepts a **fresh cluster policy report** and **one local node report per expected Linux node**, plus **one master-local report per expected control-plane node**. It performs no Kubernetes API calls and needs no cluster credentials. Node inventory is supplied explicitly to prevent silently losing nodes.

## Windows 11 / PowerShell / Podman-backed Kind

From the source repository root (Go 1.26+ installed):

```powershell
$ErrorActionPreference = 'Stop'
go mod download
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
New-Item -ItemType Directory -Force .\dist | Out-Null
go build -trimpath -o .\dist\kubecis.exe ./cmd/kubecis
$env:KIND_EXPERIMENTAL_PROVIDER = 'podman'
kubectl config use-context kind-control-plane
.\scripts\Deploy-KindNative.ps1 -ClusterName control-plane
.\scripts\Collect-ConsolidatedKindReport.ps1 -ClusterName control-plane -Binary .\dist\kubecis.exe
$report = Get-Content .\reports\consolidated.json -Raw | ConvertFrom-Json
$report | Select-Object auditMode,benchmark,benchmarkCompatibilityStatus,summary,coverageMetrics,compliance
```

The collection script uses a fresh one-shot `node-audit` inside the running unprivileged node and master auditor containers, using their existing metadata snapshots. It pulls the cluster scan from the dedicated Job. It never scrapes rolling 24-hour JSON logs. It rejects missing auditor pods and fails if any source fails to produce parseable JSON. The local CLI then validates all sources, including timestamps (default max age 30 minutes), identical build version/benchmark/compatibility, counts, logical finding identities, and per-node completeness.

**IMPORTANT:** Re-run the cluster Job before collecting if it is more than 30 minutes old. `Deploy-KindNative.ps1` creates a new Job. Never resolve a stale report by increasing max age for production evidence.

## Benchmark / CIS truthfulness

The repository's `cis-2.1.0` is explicitly labelled a **provisional starter profile**. It has a small subset of controls, and its Kubernetes 1.37 compatibility mapping is not validated against a complete licensed upstream benchmark. `cis-2.0.1` is also not a full compliance implementation. **Passing a scan is not proof of complete CIS Kubernetes Benchmark conformance.** Native checks and non-scored extensions are useful vulnerability/configuration assessments but do not substitute for licensed control-by-control benchmark implementation and validation. Every consolidated report withholds a CIS compliance percentage, even if every assessed check passes.

## Release readiness matrix

| Gate | Current evidence | Status |
|---|---|---|
| Source formatting / manifest and ZIP integrity | Can be checked in release workspace | Check results in release notes |
| Full Go unit + integration suite | User reports tests are working, no captured full test output supplied | User-confirmed, not independently audited |
| Podman + Kind runtime | User supplied functioning three-scanner output | Verified from user logs for integrated RC2 |
| Node helper config extraction | PASS evidence for anonymous-auth and authorization.mode | Verified from user logs |
| Multi-node consolidation | New CLI, tests; needs Kind E2E with all nodes | Pending environment E2E |
| Profile coverage and version mapping | Starter subset, provisional Kubernetes 1.37 | Not CIS compliance ready |
| Root project license and dependency notices | Root license owner decision missing | Publication blocker |
| Signed versioned release, SBOM, vulnerability scan | Not supplied | Publication hardening pending |

Do not label this source archive a production-ready CIS compliance product before all gates are independently met. It is a native Kubernetes security auditor release candidate with a deliberately limited profile.

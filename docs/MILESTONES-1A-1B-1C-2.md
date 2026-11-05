# KubeCIS consolidated delivery — Milestones 1A, 1B, 1C and 2

## Scope

This delivery consolidates cluster-version provenance, benchmark compatibility enforcement, report-integrity regression coverage, and stable finding identity with report-to-report comparison.

### 1A — Kubernetes version detection and propagation

- Native API scans record the Kubernetes server version and `clusterVersionSource=kubernetes-api`.
- Upstream kube-bench output remains the preferred version source when it provides a real version.
- Hybrid reports fall back to the native Kubernetes API version when upstream returns an empty/`none`/unknown version.
- Reports identify the source of the version. Node-local reports explicitly state that no API version was queried.

### 1B — benchmark compatibility enforcement

- Reports carry `benchmarkCompatibilityStatus`.
- Only the explicit value `supported` permits an official compliance percentage.
- `provisional`, `unsupported`, `unknown`, or missing compatibility metadata withholds the official percentage and emits an explicit reason.
- Hybrid reports identify the upstream profile and the KubeCIS extension profile separately, including the extension profile compatibility status.
- The compatibility gate does not relabel raw PASS/FAIL findings; it qualifies the benchmark-level compliance claim.

Important: kube-bench profile directory names are not treated as official CIS publication versions. A profile is not marked supported merely because kube-bench executed it.

### 1C — report integrity regression coverage

- Summary counts continue to be recomputed from result records in the upstream/hybrid finalizer.
- Compliance scoring continues to exclude unscored KubeCIS extensions.
- Regression tests cover stable finding identity, comparison classifications, malformed reports, and compatibility score withholding.

### 2 — stable finding identity and scan comparison

Each result receives a deterministic `findingId` derived from the check ID, target, and sorted resource identity evidence (`resource`, `namespace`, `name`, `field`). Observed values and status are excluded, so drift does not change the identity.

Compare two JSON reports:

```powershell
.\dist\kubecis.exe compare `
  --baseline .\reports\baseline.json `
  --current .\reports\current.json `
  --output-file .\reports\comparison.json
```

The comparison classifies findings as `NEW`, `RESOLVED`, `UNCHANGED`, or `STATUS_CHANGED`, or `EVIDENCE_CHANGED` when status is unchanged but the evidence/remediation/details have drifted. Summary/compliance roll-up records are excluded from finding comparisons. Older reports without `findingId` are supported by deriving identity from their result content.

## Files changed

- `internal/audit/types.go` — finding identity and report provenance fields.
- `internal/audit/runner.go` — Kubernetes version source, finding IDs and compatibility gate.
- `internal/audit/kubebench.go` — upstream version fallback, hybrid profile provenance, compatibility gate and finalization identity.
- `internal/audit/comparison.go` — deterministic identity, report validation and comparison model.
- `internal/audit/comparison_test.go` — identity, comparison, decoding and compatibility tests.
- `cmd/kubecis/main.go` — `kubecis compare` command.
- `docs/MILESTONES-1A-1B-1C-2.md` — consolidated implementation and usage notes.

## Validation status

Formatting was applied with `gofmt`. Full `go test`, `go vet`, and build validation were not executable in the authoring environment because the repository pins Go 1.26.0 and the Go toolchain download endpoint was unreachable. Do not treat this package as release-validated until the PowerShell validation commands below pass in the user's Go 1.26+ environment.

```powershell
gofmt -w .\cmd .\internal
go vet ./...
go test ./... -count=1
go build -trimpath -o .\dist\kubecis.exe .\cmd\kubecis
```

## Known limitation

The upstream profile resolver/compatibility mapping is not yet a complete authoritative matrix for every Kubernetes distribution and version. Consequently, the integrated hybrid path reports compatibility as `unknown` unless a supported mapping is explicitly supplied. This intentionally withholds official compliance percentages rather than presenting an unverified score as CIS compliance.

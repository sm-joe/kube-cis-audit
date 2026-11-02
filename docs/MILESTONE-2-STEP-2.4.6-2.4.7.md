# Milestone 2.4.6–2.4.7 — HostPath Assessment and Workload Security Consolidation

## 2.4.6 — HostPath volume assessment

KubeCIS inventories `hostPath` volumes on non-terminal Pods and correlates each volume with regular, init, and ephemeral container mounts and block-device mappings.

### Classification and status

- `FAIL`: a sensitive host path is mounted writable, or a mount uses `Bidirectional` mount propagation.
- `WARN`: a sensitive host path is mounted read-only; a non-sensitive host path is mounted writable; a HostPath volume is declared but not mounted; a path is invalid; or mount propagation is `HostToContainer`.
- `INFO`: a non-sensitive host path is mounted read-only.
- `PASS`: eligible Pods were inspected and no HostPath volumes were found.
- `NOT_ASSESSED`: Pod inventory failed or no non-terminal Pods were available.

Sensitive lexical path classes include host root, system configuration, process/kernel interfaces, device files, runtime state, Kubernetes node state, container runtime state, host logs, root/user home directories, boot files and system binaries.

Path matching is lexical and uses path boundaries (`/etc` matches `/etc/kubernetes`, not `/etc2`). KubeCIS does not resolve symlinks, inspect host filesystem ownership/permissions, or verify effective runtime mounts. HostPath type is recorded but not treated as proof that the host object exists or has the declared type.

## 2.4.7 — Workload security consolidation

A final `KUBECIS-2.4.7` result consolidates the existing KubeCIS workload checks 2.4.1 through 2.4.6. It reuses results already generated in the same report and does not make additional Kubernetes API calls.

Status precedence:

1. `FAIL` if any underlying workload check fails.
2. Otherwise `WARN` if any underlying check warns.
3. Otherwise `NOT_ASSESSED` if a required result is missing, manual, or not assessed.
4. Otherwise `INFO` if at least one underlying result is informational.
5. `PASS` only when all six underlying checks pass.

The summary contains one evidence entry per underlying check. Detailed findings remain in their individual results.

## Validation

Run from the repository root in PowerShell:

```powershell
gofmt -w internal/audit/runner.go internal/audit/types.go internal/audit/hostpath_assessment.go internal/audit/hostpath_assessment_test.go internal/audit/workload_security_summary.go internal/audit/workload_security_summary_test.go
go test ./internal/audit -count=1
go test ./... -count=1
go build ./cmd/kubecis
```

Cluster report:

```powershell
go run ./cmd/kubecis scan --benchmark cis-2.1.0 --output json
```

Expected report includes `KUBECIS-2.4.6` (`api_pod_hostpath`) and `KUBECIS-2.4.7` (`report_workload_security_summary`), with build version `2.4.7-workload-security-consolidation`.

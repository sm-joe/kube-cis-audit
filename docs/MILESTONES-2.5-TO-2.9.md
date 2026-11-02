# KubeCIS Milestones 2.5–2.9 — Implementation and validation

This document describes the consolidated implementation added on top of the 2.4.7 workload-security baseline.

## Implemented scope

| Milestone | Check | Implementation |
|---|---|---|
| 2.5 RBAC security | `KUBECIS-2.5.1` / `api_rbac_risk` | Roles, ClusterRoles, RoleBindings, ClusterRoleBindings, wildcard and sensitive permissions, escalation verbs, service-account automount metadata |
| 2.6 Network security | `KUBECIS-2.6.1` / `api_network_security` | Per-Pod ingress/egress NetworkPolicy selection, broad empty-peer rules, Services, Ingress hosts and hostPort mappings |
| 2.7 Control plane | `KUBECIS-2.7.1` / `api_control_plane_security` | Visible control-plane Pod argument checks for API server, etcd, scheduler and controller-manager |
| 2.8 Audit logging | `KUBECIS-2.8.1` / `api_audit_logging` | Visible API-server audit policy, log path and webhook backend flags |
| 2.9 Compliance reporting | `KUBECIS-2.9.1` / `report_compliance_summary` | Scored-control counts, coverage and a compliance percentage only when all scored checks resolve to PASS/FAIL |

All five entries are KubeCIS extensions and are marked `scored: false`; they are not represented as official CIS controls. The 2.9 summary is a reporting extension, not a control.

## Important assessment boundaries

- RBAC permission findings describe configured grants. They do not prove that a subject has exercised a permission or establish every effective authorization decision.
- Service-account automount evidence is based on Pod and ServiceAccount specs. Token Secret contents are never retrieved.
- NetworkPolicy selection is a configuration assessment. It does not verify CNI enforcement, packet flow, cloud firewall behavior or public reachability.
- LoadBalancer, NodePort, Ingress and hostPort records are exposure indicators, not proof of internet accessibility.
- Control-plane checks require visible control-plane Pods. Managed control planes commonly hide these Pods and must be verified through provider-side evidence; absent visibility produces MANUAL/NOT_ASSESSED rather than PASS.
- Control-plane argument inspection does not read mounted configuration files or determine all effective defaults.
- Audit backend flags do not prove that referenced files are readable, policy content is adequate, logs are delivered/retained, or alerts work.
- Compliance coverage is `(PASS + FAIL) / scored controls`. A compliance percentage is withheld if any scored check is WARN, MANUAL, NOT_ASSESSED, missing or otherwise unresolved. KubeCIS extensions are excluded.

## Files added

- `internal/audit/rbac_assessment.go`
- `internal/audit/network_assessment.go`
- `internal/audit/controlplane_assessment.go`
- `internal/audit/compliance_summary.go`
- `internal/audit/milestones_25_29_test.go`

## Files updated

- `internal/audit/runner.go` — evaluator dispatch and post-run compliance summary integration.
- `internal/audit/types.go` — build version and structured `compliance` report section.
- `benchmarks/cis-2.1.0/benchmark.yaml` — five new KubeCIS extension entries.
- `README.md` — feature and evaluator documentation.

Build version: `2.9.0-consolidated-security-assessment`.

## Windows PowerShell validation

Run from the repository root with the required Go 1.26.0 toolchain:

```powershell
gofmt -w ./cmd ./internal
go vet ./...
go test ./internal/audit -count=1
go test ./... -count=1
go build -o .\dist\kubecis.exe .\cmd\kubecis
```

Validate the benchmark and run the cluster audit:

```powershell
.\dist\kubecis.exe scan --benchmark cis-2.1.0 --target all --output json | Set-Content -Encoding utf8 .\kubecis-report.json
```

Expected new result IDs: `KUBECIS-2.5.1`, `KUBECIS-2.6.1`, `KUBECIS-2.7.1`, `KUBECIS-2.8.1`, and `KUBECIS-2.9.1`.

Expected build version: `2.9.0-consolidated-security-assessment`.

The report should also contain a top-level `compliance` object with scored-control counts, coverage percentage, score status and (only when complete) `compliancePercent`.

## Validation status

Formatting and YAML structure/unique IDs were checked in the authoring environment. Go tests and build were not executed: the available environment has Go 1.23.2 and no cached Kubernetes dependencies, while this repository requires Go 1.26.0 and Kubernetes client-go v0.37.2. Run the PowerShell commands above before treating this package as build-verified.

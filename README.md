# KubeCIS B2 — Safe Evaluator Adapter

Standalone Go module. It translates a normalized upstream `TestDescriptor` into a typed request only when an exact mapping is allowlisted. It never executes upstream `audit` strings, shell commands, scripts, or arbitrary paths.

## Safety contract

- Exact match required on control ID, target, source kind, and source value.
- Comparator operator must be explicitly allowlisted.
- Evaluator must be registered by trusted Go code.
- Mapping must set `allow_evaluation: true`.
- Evidence must be supplied by a trusted collector and match the declared evidence kind/key.
- Missing or mismatched evidence returns `NOT_ASSESSED`, never `PASS`.
- Manual upstream controls remain `MANUAL`.
- No filesystem, process, network, or Kubernetes client access exists in this module.

The example mapping is intentionally disabled (`allow_evaluation: false`). No production mappings are asserted by this package; each must be reviewed against the exact CIS requirement and KubeCIS evidence contract before approval.

## Build and test

```powershell
Set-Location .\kubecis-b2-standalone
.\Build-KubeCIS-B2.ps1
```

# KubeCIS Native v3.1.0-rc1 — release candidate

## Changed

- Retired Aqua kube-bench invocation, binary, Go parser package, profile corpus, installer scripts and container build stage.
- Unified `scan` under the already implemented native Go assessment engine; removed hybrid/upstream selection. Legacy CLI flags continue to parse and reject unsupported modes with clear errors.
- Preserved public JSON result fields, baseline comparison, API checks, node metadata helper and Linux node-audit interface.
- Added original native kubelet YAML checks for explicit anonymous authentication and authorization mode; fail-closed evaluation on missing/unavailable configuration.
- Removed outdated upstream implementation documentation and build recipes.

## Validation and blockers

Source formatting and archive checks can be validated in this session; full Go build and integration tests require Go >=1.26 and available Kubernetes dependencies. Podman/Kind E2E and native host evidence tests require an accessible Linux runtime, not provided by this execution environment. Existing profiles are limited/provisional, not complete CIS profiles. Root copyright/license selection remains the owner's responsibility. Do not publish as a stable certified CIS auditor.

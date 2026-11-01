# Milestone 2 — Audit Accuracy and Node-Level Assessment

## Baseline

Milestone 2 starts from `kube-cis-audit` Milestone 1 Build Fix 4. The Kubernetes 1.37 / CIS 2.1.0 profile remains provisional and limited; this milestone does not claim complete benchmark coverage.

## Step 2.1 — Evaluator semantics and evidence

Implemented in this increment:

- Distinguish an explicitly incompatible security setting (`fail`) from a missing setting (`missing`) and an unsupported/unknown setting (`unknown`).
- For the non-root check, missing Pod-spec evidence remains a warning/unknown rather than an assertion that the image runs as root.
- For the seccomp Pod-definition check, an unset profile remains a failed configuration requirement, but is counted and described separately from an explicitly incompatible profile such as `Unconfined`.
- Preserve Pod-level seccomp inheritance and container-level override behavior.
- Add regression tests for missing seccomp evidence, accurate summary wording, and a container-level `Unconfined` override.

## Step 2.2 — Node-level assessment design

Node filesystem checks must execute in a Linux process with an explicitly mounted host filesystem. A normal API-only Job cannot inspect node files. The implementation will use a dedicated node-audit DaemonSet (or equivalent per-node execution) with:

- Explicit node identity in each result.
- Read-only host filesystem mounts, limited to the configured paths required by the selected checks wherever the Kubernetes distribution permits.
- No host PID, host network, privileged mode, or added Linux capabilities by default.
- A non-root execution model where host directory traversal and metadata access permit it; any check requiring broader host permissions must be explicitly documented and opt-in.
- Per-node output that identifies missing paths, permission-denied paths, unsupported layouts, and actual permission findings separately.
- No claim that one Job's hostPath mount represents every node.

Before enabling the DaemonSet, verify the target distribution's actual kubelet configuration path and control-plane manifest layout. Managed control planes may not expose control-plane host files at all; those checks must remain `NOT_ASSESSED` or `MANUAL` unless provider evidence is available.

## Acceptance criteria

- `gofmt` is clean.
- `go test ./...` passes with the Go version required by `go.mod` (Go 1.26.0 or later).
- Missing non-root settings do not become confirmed root execution failures.
- Missing seccomp profile is reported as a missing explicit Pod-spec requirement, not mislabeled as an explicit incompatible profile.
- Pod-level inheritance and container overrides are covered by tests.
- Node-level checks produce per-node evidence and never imply cluster-wide coverage from a single node.


## Step 2.3.1 — Admission mechanism discovery

The first admission-policy increment inventories Pod Security Admission enforce labels, ValidatingAdmissionPolicy objects/bindings, and validating webhooks. It matches actual admission rules against core/v1 Pod CREATE/UPDATE requests instead of relying on vendor names. VAP Deny actions and matching validating webhooks are marked MANUAL candidates; Warn/Audit-only VAP bindings are WARN. API inventory errors are reported as NOT_ASSESSED. This does not prove effective enforcement: CEL logic, selectors, exclusions, parameter resolution, webhook reachability/behavior, and API-server plugin configuration still require review. See `docs/MILESTONE-2-STEP-2.3.md`.

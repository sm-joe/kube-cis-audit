# Milestone 1 — CIS evaluator correctness and evidence quality

## Scope delivered

- Added explicit result states: `MANUAL` and `NOT_ASSESSED` alongside `PASS`, `FAIL`, `WARN`, and `INFO`.
- Added structured evidence fields to each result and included them in JSON output and terminal logs.
- Host file access failures now produce `NOT_ASSESSED`; they are not treated as failed security controls.
- Pod security evaluation now covers regular, init, and ephemeral containers across all namespaces. Container-level security context fields override Pod-level defaults.
- Non-root evaluation distinguishes an explicit root UID / unconstrained identity from a proven non-root UID or `runAsNonRoot=true`. Missing image-user evidence is `WARN`, not an assertion that the process is root.
- Seccomp evaluation distinguishes RuntimeDefault, valid Localhost, Unconfined, malformed Localhost, and missing profiles.
- Policy-control discovery checks PSA enforce labels, VAP bindings with Deny actions, and recognizable Kyverno/Gatekeeper/Kubewarden validating webhooks. Candidate discovery is not treated as proof of enforcement; it requires manual review.
- NetworkPolicy presence is evaluated across every namespace, including system namespaces, with explicit evidence for each namespace. Presence is not proof of dataplane isolation.
- Cluster-admin bindings include both ClusterRoleBindings and namespaced RoleBindings, and include ServiceAccount subjects.
- Updated the Kubernetes client libraries to v0.37.2 for the Kubernetes 1.37 cluster target and Go 1.26.0 minimum.
- Added focused unit tests for evaluator edge cases.

## Unit test coverage

`internal/audit/runner_test.go` covers:

1. Pod/container inheritance and overrides for `runAsNonRoot` / `runAsUser`.
2. Root UID conflicts and explicit non-root UID handling.
3. RuntimeDefault, Localhost, Unconfined, and absent seccomp profiles.
4. Regular, init, and ephemeral container enumeration.
5. No PSA labels does not produce a false FAIL.
6. Fully labelled PSA namespaces require control-plane/manual confirmation.
7. A bound VAP with Deny is a manual-review candidate, not an automatic PASS.
8. NetworkPolicy scope includes system namespaces.
9. ServiceAccount and namespaced RoleBinding cluster-admin subjects are included.
10. Host permission checks use a permission-bit mask.
11. Missing host files are NOT_ASSESSED.

## Known limitations

- The current policy-control evaluator discovers candidates but does not evaluate the semantic content and full match scope of Kyverno, Gatekeeper, Kubewarden, or ValidatingAdmissionPolicy rules. Such findings remain `MANUAL` or `NOT_ASSESSED`.
- Pod checks inspect live non-terminal Pods, not all controller templates, future admission requests, image config metadata, or runtime process identity.
- Seccomp checks inspect the Pod API specification; they do not prove the runtime's effective seccomp state.
- NetworkPolicy checks inspect object presence, not selectors, default-deny behavior, CNI support, or packet-level enforcement.
- Host checks require a correctly mounted host filesystem. A single Job sees only its own node; full worker-node coverage needs a DaemonSet or equivalent per-node collector.
- The benchmark directory remains a small provisional starter profile, not the complete CIS Kubernetes Benchmark. The 1.37 mapping is not an official CIS compatibility assertion.

## Validation performed in the build environment

- `gofmt` validation: passed.
- YAML parsing for the CIS profile, Job, and RBAC manifests: passed.
- `go mod tidy` / `go test ./...`: not completed because this environment could not resolve `proxy.golang.org`. The source is therefore not represented as compiled or test-passed. Run the documented Go commands in an environment with Go 1.26+ and module network access.

# Production CIS Kubernetes compliance release gates

## Current disposition

**Not a production CIS compliance release.** The shipped `cis-2.1.0` profile is a provisional starter profile with 20 entries: 8 CIS-style scored controls, 11 KubeCIS extensions, and 1 report-only summary. It is not the complete CIS Kubernetes Benchmark v2.1.0.

The official CIS Kubernetes Benchmark v2.1.0 covers Kubernetes 1.36 and 1.37. A Kubernetes minor version alone does not determine benchmark applicability; the benchmark publication and platform-specific CIS benchmark must be selected explicitly. See the official CIS benchmark page and release announcement before adding or claiming mappings.

## Work completed in this candidate

- Release workflow now validates the semantic-version tag format.
- Release publication runs formatting, `go vet`, the full Go test suite, and a CLI build before publishing binaries or the container image.
- Container publication depends on the validation job.
- Prerelease tags such as `v3.0.0-rc4` publish a versioned image and GitHub prerelease, but do not move the stable `latest` image tag.
- Stable `latest` is only published for tags matching `vMAJOR.MINOR.PATCH`.
- Kustomize image override is aligned to `v3.0.0-rc4` for this candidate.
- README release status explicitly says the 20-entry profile is incomplete.

## Required before a production compliance claim

1. **Benchmark source and rights**
   - Pin the exact official CIS Kubernetes Benchmark v2.1.0 source, revision/publication date, and allowed redistribution/use terms.
   - Obtain and pin the corresponding official EKS, AKS, GKE, OKE, or OpenShift benchmark when claiming those platforms.
   - Do not copy or redistribute benchmark text outside the applicable CIS terms.

2. **Complete recommendation inventory**
   - Build a traceable inventory of every recommendation, recommendation title, profile applicability, scored/unscored designation, audit procedure, remediation, and evidence requirement.
   - Map every applicable recommendation to a native KubeCIS collector/evaluator or an explicitly documented manual assessment.
   - Include tests proving every benchmark recommendation is represented exactly once in the selected platform profile.

3. **Native execution and evidence**
   - Do not execute arbitrary shell commands embedded in kube-bench YAML.
   - Implement typed, allowlisted collectors for files, process arguments, API objects, and provider APIs.
   - Keep `PASS`, `FAIL`, `WARN`, `MANUAL`, `NOT_ASSESSED`, and unsupported states distinct. Missing evidence must never become PASS.
   - Record evidence provenance, node identity, collection time, source, and limitations.

4. **Assessment coverage**
   - Separate cluster/API, worker-node, self-managed control-plane, and managed-control-plane evidence.
   - Define an aggregation mechanism that combines per-node and cluster reports without dropping node-level failures or double-counting controls.
   - Validate all workers and all applicable control-plane nodes, not a single representative node.
   - Keep provider-managed controls `MANUAL`/`NOT_ASSESSED` until provider evidence is collected and verified.

5. **Correctness and security**
   - Independently validate each evaluator against the benchmark audit/remediation text.
   - Test API errors, RBAC denial, unsupported distributions, stale/mismatched snapshots, symlinks, permission edge cases, and partial node coverage.
   - Threat-model the metadata helper's `CAP_DAC_READ_SEARCH` use and hostPath boundary; verify the helper only performs the intended metadata operations.
   - Test that the scanner does not mutate cluster resources or host files.

6. **Platform and release validation**
   - Run Go formatting, vet, unit, integration, and E2E tests on the pinned Go toolchain.
   - Build and install from a clean checkout on Windows, Linux amd64/arm64, and macOS amd64/arm64.
   - Test supported Kubernetes versions and distribution-specific profiles.
   - Generate and verify SBOM, checksums, provenance/attestation, license notices, vulnerability reports, and immutable image digests.
   - Perform a clean install from the published release assets and published GHCR image.

## RC4 report findings that must remain visible until independently remediated

- `4.1.9`: node-local kubelet config metadata was not available to the cluster scan. The node metadata collector can assess the file separately; a production report needs reliable per-node collection and aggregation.
- `5.1.1`: cluster-admin bindings require owner/necessity review.
- `5.2.1`: active policy enforcement is not established by the current API inventory; external/provider-managed admission controls require evidence.
- `5.2.7`: explicit root containers are findings; missing `runAsNonRoot`/image-user evidence is unknown, not proof of root.
- `5.3.2`: namespaces without NetworkPolicies remain findings under the current check definition; do not blindly apply default-deny policies to system namespaces.
- `5.6.2`: missing explicit acceptable seccomp configuration remains a finding under the current check definition.
- `KUBECIS-2.8.1`: audit logging configuration and delivery/retention need provider/control-plane evidence.

## Validation limitations for this archive

The current execution environment has Go 1.23.2, while `go.mod` requires Go 1.26.0. The Go toolchain download is unavailable in this environment. Therefore `go test`, `go vet`, and Go builds have **not** been run here. YAML parsing and benchmark-profile structural checks passed, but they are not substitutes for the Go and cluster validation gates.

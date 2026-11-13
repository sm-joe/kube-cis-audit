# KubeCIS 3.0.0 release scope

## Product scope

KubeCIS is a read-only Kubernetes security assessment tool. This release package contains the native KubeCIS evaluator set and the explicitly versioned benchmark profiles shipped under `benchmarks/`.

The `cis-2.1.0` profile in this repository is **provisional** and contains 20 native entries: 8 numbered CIS-style controls, 11 KubeCIS extension checks, and 1 report-only summary. It is not the complete official CIS Kubernetes Benchmark v2.1.0 and must not be represented as complete CIS coverage, CIS certification, or a compliance attestation.

## Result semantics

- `PASS`: the evaluator obtained sufficient evidence and the check passed.
- `FAIL`: the evaluator obtained sufficient evidence and the check failed.
- `WARN`: evidence indicates a risk requiring review; it is not silently converted to PASS.
- `MANUAL`: a human/provider-side verification is required.
- `NOT_ASSESSED`: evidence was unavailable, unsupported, unsafe to inspect, or insufficient.
- `INFO`: informational output that is not a scored pass/fail result.

Missing evidence is never a PASS. KubeCIS extension checks are excluded from official CIS scoring. A percentage is withheld when a scored control remains indeterminate.

## Assessment boundaries

- Cluster/API scan: Kubernetes API evidence only.
- Worker-node assessment: allowlisted kubelet metadata through the dedicated node collector.
- Self-managed control-plane assessment: allowlisted API-server manifest metadata through the dedicated control-plane collector.
- Managed control planes: provider-managed settings remain manual or not assessed unless provider evidence is supplied and verified.

Node-local and cluster/API reports are separate outputs; this release does not silently merge them into a single cluster-wide compliance attestation.

## Release classification

This is a **technical-preview / release-candidate source package**, not a stable full-CIS compliance release. Stable compliance claims require a complete authorized benchmark inventory, native evidence mappings for every applicable recommendation, distribution-specific applicability, multi-node aggregation, and passing clean-room build, unit, integration, and E2E validation.

## Linux EKS preview workflow (November 2026)

The CLI + Bash collection pipeline produces validated, per-worker and AWS
DescribeCluster source reports, but has not yet completed clean-room Linux
EKS runtime acceptance. IAM access entries, VPC CNI and ECR enrichment are
currently uncollected in the Linux provider adapter and must not be treated as
secure or conformant. The full official benchmark remains unverified, so the
stable full-CIS release gate is intentionally blocked.

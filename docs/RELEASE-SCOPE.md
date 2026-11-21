# KubeCIS release scope

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

Stable product versions describe a tested, supported distribution of KubeCIS, not certification of CIS benchmark conformance. Benchmark compatibility and result accuracy are tracked separately. The generic Kubernetes starter profile remains limited; the CIS EKS 2.1.0 profile routes all 47 control IDs and remains provisional. Official compliance scores remain withheld.

Product releases require engineering checks, secure packaging and signed distribution. Independent benchmark verification is not required to publish the product. Stronger benchmark claims require separate supporting evidence.

## Linux EKS assessment

The native Linux workflow collects Kubernetes API, per-worker kubelet metadata and AWS evidence (EKS, access entries, CNI, ECR and subnet routing), then consolidates and validates the exact control inventory. A live two-worker assessment has completed. Supported scope is managed EC2 Linux workers with the documented kubelet layout; unsupported compute modes are rejected by the integrated collector. See [support matrix](SUPPORT-MATRIX.md) and [deployment guide](../deploy/eks/README.md).

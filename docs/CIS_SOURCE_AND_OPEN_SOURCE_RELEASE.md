# Benchmark provenance and open-source distribution

KubeCIS Native is an independent implementation of security assessment logic.
It does **not** depend on the kube-bench runtime or execute kube-bench scripts.

## Profile boundaries

- `cis-eks-2.1.0` covers Amazon EKS 1.34–1.37 (EC2 worker nodes). AWS manages the standard EKS control plane, so generic kubeadm control-plane checks must not be substituted.
- `cis-2.1.0` is presently a provisional **generic Kubernetes** profile. Generic Kubernetes v2.1.0 PDF Target Technology Details states Kubernetes 1.34–1.35, notwithstanding inconsistent introductory wording.
- EKS 2.1.0 is **not** a valid benchmark for EKS 1.31 or Fargate node management.
- Planned GKE, AKS, k3s and other distribution support requires separate verified applicability and controls; generic checks alone do not certify each vendor distribution.

## CIS benchmark licensing boundary

CIS PDFs were provided privately by the benchmark owner/operator for implementation verification.
The PDF terms explicitly prohibit third-party redistribution of a full CIS benchmark
and direct vendors to contact `legalnotices@cisecurity.org` for guidance
about incorporating portions of benchmark recommendations into third-party tooling.

**Never commit the supplied PDF documents to this repository, package them in
KubeCIS release archives/images, or copy substantial licensed text into the
open-source control catalog.** The machine-readable inventory uses factual
control identifiers, classification and implementation evidence/status; product
descriptions should be independently authored and licensing reviewed.

An open-source KubeCIS license does not imply CIS Benchmark certification,
CIS endorsement, permission to redistribute licensed PDFs, or that an
incomplete profile is compliant.

## Release security invariant

The all-controls acceptance gate must fail closed until EVERY official
automated recommendation has decisive supported evidence, EVERY manual
recommendation has an authenticated reviewer disposition with audit trace,
each supported Kubernetes version and distribution profile is validated,
and source identity/freshness/worker inventory remain consistent.

A `NOT_ASSESSED`, `MANUAL`, missing, unsupported-version or duplicate
finding must never be converted into PASS to manufacture 100% coverage.
An intentionally insecure test setting produces a FAIL finding; it is
not a reason to change the user's cluster as a prerequisite for scanning.

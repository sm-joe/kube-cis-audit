# KubeCIS Native support matrix and release acceptance

Last reviewed: 2026-11-21. This table is an **engineering scope
statement**, not an official CIS benchmark compatibility certification.

| Environment | API assessment | Linux node-local assessment | AWS provider evidence | Integrated acceptance |
| --- | --- | --- | --- | --- |
| Self-managed Kubernetes on Linux | Native available | Allowlisted metadata / node auditor | Not applicable | Requires distribution-specific E2E |
| Kind Linux nodes | Native available | Dedicated Kind auditor | Not applicable | Development integration path exists |
| Amazon EKS managed EC2 Linux, Kubernetes 1.35 | Native preview | AL2023 nodeadm supported | DescribeCluster, ECR, IAM access entries, CNI and subnet routes | Operator-reported Linux test scan passed; signed published artifact acceptance pending |
| Amazon EKS managed EC2 Linux, Kubernetes 1.34 / 1.36 / 1.37 | Provisional compatibility entries | Depends on node OS, manifest and kubelet sources | Linux DescribeCluster preview | **Not independently certified** |
| Amazon EKS Fargate / EKS Auto Mode / self-managed EKS workers | API-only where authorized | Not supported by strict managed EC2 collector | Provider inventory possible | Rejected for complete EKS consolidation |
| EKS managed control plane | Provider-owned metadata only | No customer host access | DescribeCluster snapshot | Manual / not assessed as applicable |
| Windows workers | API-only where authorized | Linux node auditor unsupported | Depends on provider | Not supported by Linux integrated collector |

## Product support levels

- **Implemented**: Code path exists and is covered by defined tests.
- **CI validated**: Source-level automated checks pass for a specific commit.
- **Integration validated**: Authorized operator has executed live cluster tests
  against the identical published artifact versions, including negative cases.
- **Benchmark verified**: Official control revision, audit procedure, applicability
  and actual evaluator correctness have independent, traceable acceptance.
- **Stable product**: Engineering and distribution gates have passed. Independent benchmark verification is a separate classification, not a product publication prerequisite.

A passing benchmark-profile parser or Kubernetes API scan does **not** imply
verified official CIS compliance. Provisional EKS and generic benchmark
profiles must retain a withheld official compliance score.

## Engineering validation and ongoing benchmark improvements

1. Independently obtain and verify the licensed official benchmark recommendations,
   version support and all control applicability.
2. Create a traceable audit-control-to-evaluator matrix and exercise positive,
   negative and missing-evidence fixtures for every applicable control.
3. Complete version-pinned Linux integrated collector acceptance on an authorized EKS managed EC2 cluster,
   including multiple nodes, one missing node, stale report, wrong version,
   wrong AWS identity and RBAC denial.
4. Verify Linux artifact install, upgrade, rollback, image provenance, signatures,
   SBOM, source/build version matching and clean uninstallation.
5. Freeze supported versions and remove obsolete candidate documentation after
   references, reproducibility and licensing have been reviewed.

The stronger benchmark verification gate remains separate from stable product publication; the provisional profile and score withholding remain unchanged.

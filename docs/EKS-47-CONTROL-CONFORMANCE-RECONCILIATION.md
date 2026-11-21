# CIS EKS 2.1.0 — Actual Runtime Mapping and Verification Boundary

## What the catalog means

This table is mapped against the official CIS Amazon EKS Benchmark v2.1.0 PDF supplied privately by the operator. The licensed audit text is not redistributed. **Candidate** means a functioning route exists for assessment or human evidence review; it does **not** establish exact procedure conformance or PASS.

All **47** controls are now routed: **29 automated** audit paths and **18 manual** reviewer workflows. The conformance ledger still records **0 independently verified** controls and retains `releaseReady=false`, `compatibilityStatus=provisional` and the withheld official compliance score.

| Official control | Assessment type | Runtime evaluator or evidence route | Independent verification |
|---|---|---|---|
| 2.1.1 | automated | `aws_eks_describe_cluster` | Pending |
| 3.1.1 | automated | `host_file_mode` | Pending |
| 3.1.2 | automated | `host_file_owner` | Pending |
| 3.1.3 | automated | `host_file_mode` | Pending |
| 3.1.4 | automated | `host_file_owner` | Pending |
| 3.2.1 | automated | `host_kubelet_config_field` | Pending |
| 3.2.2 | automated | `host_kubelet_config_field` | Pending |
| 3.2.3 | automated | `host_kubelet_config_field` | Pending |
| 3.2.4 | automated | `host_kubelet_config_field` | Pending |
| 3.2.5 | automated | `host_kubelet_config_field` | Pending |
| 3.2.6 | automated | `host_kubelet_config_field` | Pending |
| 3.2.7 | automated | `host_kubelet_config_field` | Pending |
| 3.2.8 | automated | `host_kubelet_config_field` | Pending |
| 4.1.1 | manual | `kubernetes_inventory_human_attestation` | Pending |
| 4.1.2 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.3 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.4 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.5 | automated | `eks_default_service_account` | Pending |
| 4.1.6 | automated | `eks_token_mounts` | Pending |
| 4.1.7 | automated | `aws_eks_access_entries` | Pending |
| 4.1.8 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.9 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.10 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.11 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.1.12 | manual | `kubernetes_bound_rbac_inventory` | Pending |
| 4.2.1 | automated | `eks_psa_admission` | Pending |
| 4.2.2 | automated | `eks_psa_admission` | Pending |
| 4.2.3 | automated | `eks_psa_admission` | Pending |
| 4.2.4 | automated | `eks_psa_admission` | Pending |
| 4.2.5 | automated | `eks_psa_admission` | Pending |
| 4.3.1 | automated | `eks_vpc_cni_network_policy` | Pending |
| 4.3.2 | manual | `kubernetes_inventory_human_attestation` | Pending |
| 4.4.1 | automated | `eks_secret_env` | Pending |
| 4.4.2 | manual | `kubernetes_inventory_human_attestation` | Pending |
| 4.5.1 | manual | `kubernetes_inventory_human_attestation` | Pending |
| 4.5.2 | automated | `eks_default_namespace` | Pending |
| 5.1.1 | automated | `aws_ecr_image_scanning` | Pending |
| 5.1.2 | manual | `aws_manual_review_inventory` | Pending |
| 5.1.3 | manual | `aws_manual_review_inventory` | Pending |
| 5.1.4 | manual | `kubernetes_inventory_human_attestation` | Pending |
| 5.2.1 | automated | `eks_dedicated_service_account` | Pending |
| 5.3.1 | manual | `aws_manual_review_inventory` | Pending |
| 5.4.1 | automated | `aws_eks_describe_cluster` | Pending |
| 5.4.2 | automated | `aws_eks_describe_cluster` | Pending |
| 5.4.3 | manual | `aws_manual_review_inventory` | Pending |
| 5.4.4 | manual | `kubernetes_inventory_human_attestation` | Pending |
| 5.5.1 | automated | `aws_eks_access_entries` | Pending |

## Specific evidence boundaries

The 12 worker-node controls assess host-file ownership/permissions or bounded kubelet configuration with missing-evidence detection; client certificate rotation is conditional on active kubelet authentication method. Kubernetes 4.x checks use live Kubernetes API or MANUAL RBAC/network/Secrets review. The AWS 2.x/5.x automated checks use verified EKS/ECR provider inventories, including signed third-party image scanner evidence where applicable.

The four newly attached AWS manual inventories are **review inputs, not approvals**:

- `5.1.2` lists deployed first-party ECR repositories and records hashes/presence of repository and registry policies, not their effective IAM authorization.
- `5.1.3` inventories EKS managed node-group IAM roles and attached/inline policy identities; complete policy documents and Fargate/self-managed roles still need human review.
- `5.3.1` inventories EKS encryption KMS key associations and key manager/state; grants, rotation and organization approval need review.
- `5.4.3` inventories managed node-group subnets, effective routes and registered worker EC2 public addresses; observed public placement produces FAIL. Coverage outside the strict managed-worker contract needs review.

A denied AWS API request produces PARTIAL or UNAVAILABLE reviewer evidence, **never PASS**. Reviewer decisions must be bound to a specific report digest using the existing `kubecis review` / `kubecis finalize` flow.

## Verification commands

```bash
bash scripts/test-eks-aws-review.sh
bash scripts/test-eks-provider.sh
python3 scripts/test-eks-evaluator-conformance.py
python3 scripts/verify-eks-evaluator-conformance.py
# Expected: 47 candidate, 0 not_implemented, 0 verified, releaseReady=false
python3 scripts/verify-eks-evaluator-conformance.py --require-full
# Expected: nonzero until each audit procedure has independent evidence
```

The optional stronger benchmark-verification gate requires control-specific positive and negative fixtures, documented official audit-procedure review, real runtime acceptance evidence, and reviewer attribution for **every** control. No assistant can fabricate the required live AWS/account evidence, make real human decisions, or infer a CIS certification merely from CI being green.

## Release-readiness source alignment

The operator-held source was reopened and its SHA-256 matched the pinned digest. The engineering review found and addressed narrower RBAC inventory gaps and missing init/ephemeral registry inventory. This is not independent approval of every evaluator. The ledger remains unchanged until real control-specific fixtures, runtime evidence and reviewer approval are supplied. Release validation builds and checks archives without publishing on pull requests; stable product publication does not require independent benchmark verification. A stronger benchmark compatibility claim still requires supporting evidence and the separate verification gate. The ledger's releaseReady field describes benchmark verification readiness, not product distribution readiness.

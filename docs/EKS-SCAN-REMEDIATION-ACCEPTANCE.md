# EKS findings: remediation and re-scan acceptance

This guide is based on a real two-worker EKS 1.37 assessment dated
2026-11-19 (59 findings: 22 PASS, 7 FAIL, 27 MANUAL, 3 NOT_ASSESSED).
It contains **no account IDs, private endpoint names, kubeconfigs or Secrets**.
These findings are from an immutable historical report, not a live read.

## 1. Inspect before changing cluster infrastructure

```bash
export CLUSTER_NAME="YOUR_EKS_CLUSTER"
export AWS_REGION="YOUR_AWS_REGION"
aws sts get-caller-identity
aws eks describe-cluster --name "$CLUSTER_NAME" --region "$AWS_REGION" \
  --query 'cluster.{Status:status,Version:version,Access:accessConfig.authenticationMode,Endpoint:resourcesVpcConfig}' --output json
kubectl -n kube-system get daemonset eks-pod-identity-agent -o yaml
kubectl -n kube-system get deployment ebs-csi-controller -o yaml
```

Do not copy unredacted Pod environment values, Secret contents or credentials
into support issues or public repositories.

## 2. Public endpoint access: 5.4.1 and 5.4.2

The report observed `endpointPrivateAccess=true`,
`endpointPublicAccess=true` and `publicAccessCidrs=["0.0.0.0/0"]`.
These are **two controls triggered by one underlying endpoint configuration**;
the findings must remain FAIL until the AWS API confirms remediation.

**Option A (preferred where private administrative connectivity is verified):**
after confirming a working route/VPN/bastion into the VPC, arrange a change
window, preserve AWS administration access, then update:

```bash
aws eks update-cluster-config --name "$CLUSTER_NAME" --region "$AWS_REGION" \
  --resources-vpc-config endpointPrivateAccess=true,endpointPublicAccess=false
```

**Option B (public administration required):** replace the permissive internet
CIDR with a validated fixed admin IP/CIDR. Avoid arbitrary guessed CIDRs.

```bash
# Replace with your actual, currently reachable, approved administrative CIDR.
export ADMIN_CIDR="YOUR_APPROVED_ADMIN_CIDR"
# Inspect the proposed CIDR and access path before running:
aws eks update-cluster-config --name "$CLUSTER_NAME" --region "$AWS_REGION" \
  --resources-vpc-config "endpointPrivateAccess=true,endpointPublicAccess=true,publicAccessCidrs=$ADMIN_CIDR"
```

A restricted public endpoint may remove the 5.4.1 exposure, but **5.4.2 is
private-only** and continues to FAIL until public access is disabled.
Wait for the EKS update to finish and re-query the exact fields before rescanning.
Never disable public access while all administrators are outside the VPC.

## 3. EKS access entries: 4.1.7 and 5.5.1

The report shows `authenticationMode=CONFIG_MAP`; two separate control IDs
reference that same configuration. This is genuine AWS evidence, not a
duplicate Kubernetes event.

**Do not switch immediately to API-only.** Inventory current `aws-auth`
role/user mappings, node bootstrap identities, existing RBAC and administrator
access, then prepare equivalent EKS access entries. AWS authentication-mode
changes are one-way. Once approved, migrate incrementally:

```bash
aws eks update-cluster-config --name "$CLUSTER_NAME" --region "$AWS_REGION" \
  --access-config authenticationMode=API_AND_CONFIG_MAP
```

After the update completes, configure and test the correct access entries,
policy associations and IAM principals. Keep the ConfigMap mode during the
transition; an API-only switch is a separate controlled decision. The
Linux collector's access-entry inventory is currently **UNAVAILABLE**, so
its score cannot attest least privilege even after the mode is changed.

## 4. Default ServiceAccount: 4.1.5 and 5.2.1

The historical report captured two EKS Pod Identity Agent Pods and their
owning DaemonSet with the default ServiceAccount. These are real Kubernetes
references, but their owner/exception status cannot be verified from the
reference alone. Do **not** patch an AWS-managed add-on DaemonSet merely to
satisfy the report. Check add-on ownership, effective token mounting, RBAC and
supported vendor configuration first.

The scanner now marks **platform-only** default-SA evidence `MANUAL` with
the original resource and field evidence retained. Non-`kube-system`
application default-ServiceAccount use remains `FAIL`. Two control IDs
still exist because they are distinct benchmark recommendations, but Pod and
controller references are explicitly noted as duplicate underlying sources.

## 5. EBS CSI Secret environment references: 4.4.1

The historical report captured four `secretKeyRef` references for
`AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` in two EBS CSI controller Pods.
These are references to Kubernetes Secrets; the scanner did **not** read
credentials or establish credential compromise.

Inspect deployment/add-on ownership and the supported IAM model. For your own
workloads, prefer short-lived IAM credentials through EKS Pod Identity or
IRSA, and remove static access keys after verifying the controller functions.
For EKS-managed add-ons, use supported AWS add-on configuration instead of
directly editing managed Pods. Preserve manual review findings until the
approved configuration has been independently verified.

## 6. Unresolved evidence, not automatic failures

- **5.1.1 ECR/third-party scanning:** Linux provider evidence reports
  `ecrInventoryStatus=UNAVAILABLE`. Implement/authorize full deployed-image
  registry/digest scanning inventory or obtain validated third-party attestations
  before determining PASS/FAIL. No invented PASS.
- **3.2.7 on both workers:** the effective kubelet `/configz` omitted
  `rotateCertificates`. Review active kubelet authentication mode and
  configuration; do not infer defaults. Keep NOT_ASSESSED without new evidence.
- **Other MANUAL controls:** require explicit evidence-backed human review.
  A reviewer approval is not a replacement for automated source observations.

## 7. Acceptance: only a fresh authorized EKS scan counts

```bash
git pull origin main
go test ./... -count=1
bash scripts/test-linux-eks.sh

bash scripts/collect-eks.sh \
  --context YOUR_EKS_CONTEXT \
  --cluster-name "$CLUSTER_NAME" \
  --region "$AWS_REGION" \
  --output-dir reports/eks

jq -r '.results[] | select(.status=="FAIL" or .status=="NOT_ASSESSED") |
 [.id,.status,.title,.details] | @tsv' reports/eks/consolidated.json
jq '.summary,.compliance,.providerEvidence |
  if type == "object" then . else . end' reports/eks/consolidated.json
```

Success means every supported worker has current evidence, all source-count
and freshness gates pass, AWS endpoint/authentication controls reflect actual
post-change values, and the previously overconfident platform-only checks
retain evidence under MANUAL review. **Official CIS EKS 2.1.0 compliance
remains WITHHELD** pending independent benchmark verification.

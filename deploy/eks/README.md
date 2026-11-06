# EKS native audit (experimental, deliberately provisional)

This EKS integration is scoped to EC2 **managed node groups**. No
control-plane host mount or master-local DaemonSet is deployed. Fargate,
EKS Auto Mode, and self-managed EC2 nodes are not covered by this example.

## Build and publish the matching binary and image

Build the CLI from this branch (`go build -o kubecis.exe ./cmd/kubecis` on Windows).
Build and publish the image from **the same commit**, updating the placeholder
`ghcr.io/sm-joe/kubecis:eks-provider-rc3` in
`node-auditor-daemonset.yaml` to its immutable digest or release tag.
Do not deploy the placeholder until an image exists. The collector deliberately
requires the local CLI and in-cluster node auditor to have matching build
versions, validated by `consolidate`.

## Deploy

Select the EKS kubeconfig context deliberately. Verify the target using
`kubectl config current-context` and `kubectl get nodes -L eks.amazonaws.com/nodegroup`.
Use `kubectl apply -k deploy/eks` from the repository root. This applies
the read-only Kubernetes API ServiceAccount/RBAC and the Linux-only node
auditor. The node auditor's metadata helper is root with a narrowly
scoped hostPath to `/var/lib/kubelet`; review this privileged host
access exception and the Pod Security policy requirements before deployment.

Wait for the DaemonSet with:
`kubectl -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=180s`.

## Collect

Run `./scripts/Collect-ConsolidatedEKSReport.ps1 -Context <your-eks-context>\n-ClusterName <actual-eks-name> -Region <aws-region> -Binary ./kubecis.exe`\nin PowerShell after verifying AWS credentials
and permissions. The script produces a fresh cluster policy report
from the Kubernetes API, collects one node-local report for each EKS
managed Linux node, and validates all inputs with `consolidate`.
Missing pods, old reports, version mismatches, and inconsistent node
inventory fail closed. There is **no** expected control-plane-node list.

After a successful collection, run `./scripts/Validate-EKSRelease.ps1` to assert
provider identity, node-source completeness, freshness, and withheld scoring.
This acceptance gate is intentionally strict and does not replace live AWS
configuration validation or a certified benchmark assessment.\n\nThe provisional `cis-2.1.0` profile is a starter set, not full CIS or
EKS benchmark certification. Its `coveragePercent` measures the seven
configured scored controls before aggregation, not upstream CIS coverage.
The compliance percentage stays withheld.

## Limitations

EKS exposes control-plane configuration through provider APIs, not by
customer access to master host files. AWS settings such as control-plane
logging, encryption, networking, and access entries need separately
designed read-only AWS API collectors with explicit evidence models and
AWS IAM permissions. This branch does not claim those evaluations.
NetworkPolicy presence is not proof of enforcement; RBAC inventory can
surface system controllers and needs customer-ownership triage.

## AL2023 nodeadm evidence (RC3)

The EKS node metadata helper mounts `/etc/kubernetes/kubelet` read-only
as `/host-nodeadm` and only evaluates `config.json`. Legacy Kind/kubeadm
uses `/var/lib/kubelet/config.yaml`; the EKS overlay does not use this
fallback. The audit output records the actual source path, and the EKS
acceptance gate now fails unless all three worker-kubelet checks are explicitly
PASS or FAIL with AL2023 source evidence. A security FAIL is an assessed outcome
and is not the same as an unavailable check. This still does not validate full
upstream CIS EKS compliance.

## RC3 versioned applicability and evidence model

RC3 rejects explicit benchmark selections for Kubernetes minors not declared in
`benchmarks/cis-2.1.0/benchmark.yaml`. This intentionally **provisional**
starter profile lists only minors 1.31 and 1.37, without asserting that its
historical generic CIS control IDs represent a verified EKS-specific mapping.
The cluster report records the actual server version and provider.

On EKS worker nodes only, the generic CIS Kubernetes kubelet config.yaml mode
control `4.1.9` is replaced by **unscored**
`KUBECIS-EKS-NODE-CONFIG-MODE`, with AL2023 nodeadm `config.json` evidence
and a 0644 maximum. It is an EKS security extension, not a claim of verified
CIS EKS 3.1.3 applicability; ownership/effective kubelet flags need review.
Existing Kind/kubeadm node assessments preserve their original generic check.

The RC3 collector additionally uses read-only `aws eks describe-cluster` to
record the EKS API endpoint exposure settings and all five control-plane
logging categories in `reports/eks/provider.json`, then attaches this
**unscored inventory** to `consolidated.json`. It contains no AWS credentials,
IAM tokens, raw API endpoints, or encryption secrets. Missing/expired/wrong
cluster-version evidence is rejected. Requires valid AWS CLI credentials and
`eks:DescribeCluster` permission. This is *not* proof of effective logging,
retention, identity enforcement, or compliance.

An evidence-level `ownershipHint` distinguishes names associated with typical
EKS system components and the KubeCIS auditor from other resources. Name
matching is only for triage, not a trusted allowlist: FAIL/WARN and CIS-scored
status are **never suppressed** by these hints. An EKS-specific CIS ruleset
must be independently verified before making official compliance claims.

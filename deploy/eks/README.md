# EKS native audit (experimental, deliberately provisional)

This EKS integration is scoped to EC2 **managed node groups**. No
control-plane host mount or master-local DaemonSet is deployed. Fargate,
EKS Auto Mode, and self-managed EC2 nodes are not covered by this example.

## Build and publish the matching binary and image

Build the CLI from this branch (`go build -o kubecis.exe ./cmd/kubecis` on Windows).
Build and publish the image from **the same commit**, updating the placeholder
`ghcr.io/sm-joe/kubecis:eks-provider-rc` in
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

Run `./scripts/Collect-ConsolidatedEKSReport.ps1 -Context <your-eks-context>
-Binary ./kubecis.exe` in PowerShell after verifying AWS credentials
and permissions. The script produces a fresh cluster policy report
from the Kubernetes API, collects one node-local report for each EKS
managed Linux node, and validates all inputs with `consolidate`.
Missing pods, old reports, version mismatches, and inconsistent node
inventory fail closed. There is **no** expected control-plane-node list.

The provisional `cis-2.1.0` profile is a starter set, not full CIS or
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

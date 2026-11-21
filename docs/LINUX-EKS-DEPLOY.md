# Native Linux EKS deployment: fail-closed release helper

The EKS node auditor is installed through `scripts/deploy-eks.sh`. The helper
requires a Linux/Unix Bash environment, `kubectl`, a selected EKS kubeconfig
context, and an immutable container image digest.

```bash
# Download a signed release or matching main candidate and verify its digest.
kubectl config current-context
kubectl get nodes -L eks.amazonaws.com/nodegroup
bash scripts/deploy-eks.sh \
  --context YOUR_EKS_CONTEXT \
  --image ghcr.io/sm-joe/kubecis@sha256:YOUR_VERIFIED_64_CHARACTER_DIGEST
```

The script refuses mutable image tags and non-EKS server identities, pins all
three DaemonSet images **before** applying any objects, and waits for rollout.
It does not configure AWS authentication, collect EKS provider evidence,
consolidate results, or confer CIS compliance.

The native Linux collector is available through scripts/collect-eks.sh and has completed a live managed-worker EKS assessment. Use the matching signed CLI and published image digest. See [EKS installation and collection](../deploy/eks/README.md) for supported scope and limitations.

For reproducible shell checks:

```bash
bash scripts/test-deploy-eks.sh
```

These shell checks are input validation tests and do not replace a live
Kubernetes rollout or scan acceptance.

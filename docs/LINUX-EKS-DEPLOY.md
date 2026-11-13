# Native Linux EKS deployment: fail-closed release helper

The EKS node auditor is installed through `scripts/deploy-eks.sh`. The helper
requires a Linux/Unix Bash environment, `kubectl`, a selected EKS kubeconfig
context, and an immutable container image digest.

```bash
# Build/publish a Linux image from the same source commit as the CLI and verify
# its provenance, architecture and digest before deployment.
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

The native CLI is already cross-platform, but the repository's complete
EKS provider + worker collection is currently implemented in PowerShell.
Until a POSIX collection workflow passes real AWS/EKS end-to-end tests, do not
represent the above installation as a complete Linux EKS assessment.

For reproducible shell checks:

```bash
bash scripts/test-deploy-eks.sh
```

These shell checks are input validation tests and do not replace a live
Kubernetes rollout or scan acceptance.

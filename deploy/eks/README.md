# Amazon EKS managed-node integration

This deployment targets Amazon EKS **managed EC2 Linux worker nodes**. The node-auditor DaemonSet collects bounded, read-only kubelet and host metadata, rather than mounting or impersonating the AWS-managed control plane. Fargate, EKS Auto Mode and self-managed workers are outside the verified managed-node source contract.

> **Benchmark status:** `cis-eks-2.1.0` remains a **provisional** native profile. Scan and review completion never grant an official CIS compliance score; incomplete or inapplicable checks remain MANUAL or NOT_ASSESSED.

## Linux prerequisites

- Linux host with a shell, `kubectl`, AWS CLI v2, and a Kubernetes kubeconfig.
- A Kubernetes identity authorized to apply the project's RBAC and DaemonSet to `kube-system`.
- Published `ghcr.io/sm-joe/kubecis` Linux image built from the **same source commit** as the local KubeCIS CLI. Never reuse an old image with a newer CLI.
- For source builds: Go 1.27.4.

## Build CLI from source

From the repository root:

```bash
go mod download
go vet ./...
go test ./... -count=1
CGO_ENABLED=0 go build -trimpath -o ./kubecis ./cmd/kubecis
./kubecis version
```

A local development build is **not automatically matched** to a published EKS node image. Use a published tag/digest and ensure the local and in-cluster `kubecis version` values match before consolidating.

## Verify target cluster

```bash
aws sts get-caller-identity
aws eks update-kubeconfig --name my-eks-cluster --region us-east-1
kubectl config current-context
kubectl get nodes -L eks.amazonaws.com/nodegroup
```

## Deploy and collect using Linux

Use an immutable verified GHCR image digest. Do not deploy the DaemonSet's
candidate image tag from the raw manifest. The helper pins all three containers
before submitting any objects.

```bash
bash scripts/deploy-eks.sh --context YOUR_EKS_CONTEXT \
  --image ghcr.io/sm-joe/kubecis@sha256:YOUR_VERIFIED_64_CHARACTER_DIGEST

bash scripts/collect-eks.sh --context YOUR_EKS_CONTEXT \
  --cluster-name YOUR_CLUSTER --region YOUR_REGION \
  --binary ./kubecis --output-dir reports/eks
```

The image value shown is a non-deployable placeholder. Use the actual digest
published from the matching CLI source. The collection is read-only but the
deploy step creates/updates its own Kubernetes resources and requires review.
The preview gate requires evidence for every supported Linux managed EC2 worker.
Mixed compute modes, mismatched AWS identity, inconsistent build versions and
missing essential AL2023 kubelet evidence fail closed. Unavailable optional AWS
inventories and `nodes/proxy` results never become PASS.

See [Linux EKS collection](../../docs/LINUX-EKS-COLLECTION.md). The Go scanner
and Bash workflow are Linux-native, but **live Linux EKS end-to-end acceptance
remains outstanding**. A complete official CIS EKS score is withheld.

## Evidence and security boundaries

- Never assume customer access to EKS control-plane host files.
- Check each managed node separately, and refuse missing, stale, duplicate or mismatched evidence.
- `nodes/proxy` for authenticated `/configz` data is a sensitive permission; grant only when justified.
- Missing AWS inventory or kubelet fields do not imply PASS; they require MANUAL or NOT_ASSESSED disposition.
- RBAC names associated with platform controllers are triage hints, **not** unconditional allowlists.
- Some CI releases test platform-specific packaging; they do not establish complete CIS implementation or independent compliance certification.

See the repository [README](../../README.md), [release scope](../../docs/RELEASE-SCOPE.md) and [security guidance](../../docs/SECURITY.md).

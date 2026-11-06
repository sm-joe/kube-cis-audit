# Amazon EKS managed-node integration

This deployment targets Amazon EKS **managed EC2 Linux worker nodes**. The node-auditor DaemonSet collects bounded, read-only kubelet and host metadata, rather than mounting or impersonating the AWS-managed control plane. Fargate, EKS Auto Mode and self-managed workers are outside the verified managed-node source contract.

> **Benchmark status:** `cis-eks-2.1.0` remains a **provisional** native profile. Scan and review completion never grant an official CIS compliance score; incomplete or inapplicable checks remain MANUAL or NOT_ASSESSED.

## Linux prerequisites

- Linux host with a shell, `kubectl`, AWS CLI v2, and a Kubernetes kubeconfig.
- A Kubernetes identity authorized to apply the project's RBAC and DaemonSet to `kube-system`.
- Published `ghcr.io/sm-joe/kubecis` Linux image built from the **same source commit** as the local KubeCIS CLI. Never reuse an old image with a newer CLI.
- For source builds: Go 1.27.3.

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

## Deploy matching Linux node auditor

Review [node-auditor-daemonset.yaml](node-auditor-daemonset.yaml) and [RBAC](rbac.yaml), especially the restricted host-metadata helper, before applying the resources.

```bash
# Replace this placeholder with a real image published for the matching commit.
IMAGE="ghcr.io/sm-joe/kubecis:cis210-eks135-<40-character-commit-sha>"

kubectl apply -k deploy/eks
kubectl -n kube-system set image daemonset/kubecis-node-auditor \
  "node-metadata-init=$IMAGE" \
  "node-auditor=$IMAGE" \
  "node-metadata-helper=$IMAGE"

kubectl -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=300s
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
```

These commands install/update KubeCIS resources but do **not** remediate security findings. Validate the image's registry, digest, and the node-auditor version. The node metadata helper reads restricted paths from `/var/lib/kubelet` and, for EKS AL2023, allowlisted kubelet settings under `/etc/kubernetes/kubelet/config.json`. Raw credentials and unrestricted kubelet configuration should not be exported.

## Execute a Kubernetes API scan on Linux

```bash
mkdir -p reports/eks
./kubecis scan \
  --engine native \
  --benchmark cis-eks-2.1.0 \
  --target policies \
  --output json \
  --output-file reports/eks/cluster.json \
  --exit-code-on-fail=false

./kubecis render --input reports/eks/cluster.json --output-file reports/eks/results.txt
```

The API scan is useful independently but **does not include all managed-node and AWS provider findings**. Cluster identity, per-node sources, AWS provider inventory, freshness and compatibility must be validated before issuing a consolidated EKS dossier. The current repository does **not yet provide a verified POSIX-shell replacement** for the full integrated EKS collector; the Go scanner and Linux node image themselves are native. Do not label API-only output as a complete EKS assessment.

For a validated consolidated JSON produced by the supported collector, the Go review/finalization CLI runs on Linux:

```bash
./kubecis review-init --input reports/eks/consolidated.json --output-file reports/eks/review-template.json
./kubecis finalize --input reports/eks/consolidated.json --output-dir reports/eks
```

When verified human decisions are available, add `--decisions reports/eks/review-decisions.json` to `finalize`. Decisions remain distinct from automated findings; the score stays WITHHELD for provisional benchmark coverage.

## Evidence and security boundaries

- Never assume customer access to EKS control-plane host files.
- Check each managed node separately, and refuse missing, stale, duplicate or mismatched evidence.
- `nodes/proxy` for authenticated `/configz` data is a sensitive permission; grant only when justified.
- Missing AWS inventory or kubelet fields do not imply PASS; they require MANUAL or NOT_ASSESSED disposition.
- RBAC names associated with platform controllers are triage hints, **not** unconditional allowlists.
- Some CI releases test platform-specific packaging; they do not establish complete CIS implementation or independent compliance certification.

See the repository [README](../../README.md), [release scope](../../docs/RELEASE-SCOPE.md) and [security guidance](../../docs/SECURITY.md).

# Amazon EKS managed-node integration

This deployment targets Amazon EKS **managed EC2 Linux worker nodes**. The node-auditor DaemonSet collects bounded, read-only kubelet and host metadata, rather than mounting or impersonating the AWS-managed control plane. Fargate, EKS Auto Mode and self-managed workers are outside the verified managed-node source contract.

> **Benchmark status:** `cis-eks-2.1.0` remains a **provisional** native profile. Scan and review completion never grant an official CIS compliance score; incomplete or inapplicable checks remain MANUAL or NOT_ASSESSED.

## Linux prerequisites

- Linux host with a shell, `kubectl`, AWS CLI v2, and a Kubernetes kubeconfig.
- A Kubernetes identity authorized to apply the project's RBAC and DaemonSet to `kube-system`.
- Published `ghcr.io/sm-joe/kubecis` Linux image built from the **same source commit** as the local KubeCIS CLI. Never reuse an old image with a newer CLI.
- For source builds: Go 1.27.4.

## Install a CI-published candidate (no local Docker build)

Open [EKS Candidate Image workflow runs](https://github.com/sm-joe/kube-cis-audit/actions/workflows/eks-candidate-image.yml).
Choose a successful main-branch run and download its signed candidate artifact.
Extract the downloaded artifact ZIP, then run these commands in the directory
containing the archives. Cosign is required.

```bash
case "$(uname -m)" in
  x86_64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *) echo "Unsupported platform"; exit 1 ;;
esac
ARCHIVE="kubecis_candidate_linux_$ARCH.tar.gz"
IDENTITY=https://github.com/sm-joe/kube-cis-audit/.github/workflows/eks-candidate-image.yml@refs/heads/main
cosign verify-blob --bundle "$ARCHIVE.sigstore.json" \
  --certificate-identity "$IDENTITY" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com "$ARCHIVE"
sha256sum -c --ignore-missing SHA256SUMS
mkdir -p kubecis-candidate
tar -xzf "$ARCHIVE" -C kubecis-candidate
cd kubecis-candidate
IMAGE="$(cat IMAGE_DIGEST)"
cosign verify "$IMAGE" --certificate-identity "$IDENTITY" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
./kubecis version
bash scripts/deploy-eks.sh --context "$(kubectl config current-context)" --image "$IMAGE"
```

The archive and image use the same commit-specific build version and the
archive includes the exact OCI digest. GHCR must permit the worker nodes to
pull the package (public visibility or approved registry access). A successful
candidate publication is not a stable CIS release or independent conformance.
Use the packaged CLI and scripts together; do not mix older binaries and images.
If a deployment fails, inspect pod Events and verify image digest and registry
access before attempting a rebuild. To roll back, redeploy the previous verified
digest with its matching candidate archive.

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

## Build and publish a matching image using Docker Buildx

This Dockerfile uses BuildKit's `BUILDPLATFORM` and `TARGETARCH`
arguments. Docker's deprecated legacy builder cannot build it and fails
with `failed to parse platform: ""`. Install the
[Docker Buildx plugin](https://docs.docker.com/build/buildx/install/)
(e.g. `sudo apt-get install docker-buildx-plugin` when using Docker's
official Ubuntu/Debian package repository). Verify with `docker buildx version`.

For an EKS cluster where **all** Linux workers are AMD64, from the repository
root after configuring `IMAGE_REPO` and `IMAGE_TAG`:

```bash
docker buildx version
KUBECIS_VERSION="$(./kubecis version)"
docker buildx build --platform linux/amd64 \
  --build-arg VERSION="$KUBECIS_VERSION" \
  --load -t "${IMAGE_REPO}:${IMAGE_TAG}" .
docker run --rm "${IMAGE_REPO}:${IMAGE_TAG}" version
docker push "${IMAGE_REPO}:${IMAGE_TAG}"
```

The local and image versions must be identical. For mixed AMD64/ARM64
workers, use a `docker-container` Buildx builder and
`docker buildx build --platform linux/amd64,linux/arm64 --push ...`
instead of `--load`. Fetch the digest from your registry and pass the
immutable `repository@sha256:<digest>` to the deployment helper.
An image built locally is not an independently signed release artifact.

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

## Troubleshoot: no Running auditor Pod on a worker

If collection reports `Expected exactly one Running node-auditor Pod`,
the CLI is working, but node-local evidence cannot be collected yet.
Check the selected EKS context and the required DaemonSet first:

```bash
kubectl -n kube-system get daemonset kubecis-node-auditor
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
kubectl -n kube-system describe daemonset kubecis-node-auditor
kubectl get node ip-10-10-0-165.ec2.internal --show-labels
```

If the DaemonSet is absent, install it with `scripts/deploy-eks.sh` and
an actual, verified immutable image digest built from the matching CLI source.
Do **not** apply the raw manifest with its legacy candidate image tag.

If a Pod is `Pending`, inspect `kubectl -n kube-system describe pod POD_NAME`
for insufficient resources, node affinity, taints, or missing host directories.
If it is `ImagePullBackOff`, verify the digest, GHCR image availability,
architecture and image pull access. If `Init:CrashLoopBackOff`, review
`kubectl -n kube-system logs POD_NAME -c node-metadata-init --previous` and
the allowlisted hostPath mounts (`/var/lib/kubelet` and
`/etc/kubernetes/kubelet`); do not add broad host privileges as a workaround.

Once the DaemonSet has a Running Pod on **every** supported managed EC2 worker,
re-run `scripts/collect-eks.sh`. Missing nodes must not be silently excluded
from a consolidated assessment.

## Evidence and security boundaries

- Never assume customer access to EKS control-plane host files.
- Check each managed node separately, and refuse missing, stale, duplicate or mismatched evidence.
- `nodes/proxy` for authenticated `/configz` data is a sensitive permission; grant only when justified.
- Missing AWS inventory or kubelet fields do not imply PASS; they require MANUAL or NOT_ASSESSED disposition.
- RBAC names associated with platform controllers are triage hints, **not** unconditional allowlists.
- Some CI releases test platform-specific packaging; they do not establish complete CIS implementation or independent compliance certification.

See the repository [README](../../README.md), [release scope](../../docs/RELEASE-SCOPE.md) and [security guidance](../../SECURITY.md).

# Container image, Kubernetes deployment and E2E validation

KubeCIS has two separate execution modes:

1. **Cluster/API audit** — a short-lived Job uses a dedicated read-only ServiceAccount and Kubernetes API permissions. It does not mount the node filesystem.
2. **Node-local audit** — a DaemonSet runs an unprivileged auditor without a Kubernetes API token. A separate, narrowly scoped metadata helper can `Lstat` only the allowlisted kubelet configuration file and publishes metadata (not file contents) to a shared memory volume.

The container image is built with Go 1.26 and runs from a Distroless non-root image as UID/GID `65532:65532`. It contains the CLI and benchmark profiles, but no shell, package manager, `kubectl`, cloud CLI, or credential helper.

## Build and inspect with Podman

Run from PowerShell at the repository root:

```powershell
podman build --pull --build-arg VERSION=3.0.0-dev -t kubecis:3.0.0-dev .
podman run --rm kubecis:3.0.0-dev version
podman image inspect --format '{{.Config.User}}' kubecis:3.0.0-dev
podman run --rm kubecis:3.0.0-dev scan --help
```

Expected version output is `3.0.0-dev`; the image user should be `65532:65532`.

For a release build, pass the intended release version explicitly:

```powershell
podman build --pull --build-arg VERSION=3.0.0 -t ghcr.io/sm-joe/kubecis:3.0.0 .
podman run --rm ghcr.io/sm-joe/kubecis:3.0.0 version
```

Push only after the image has been validated and the GHCR package/repository permissions are configured:

```powershell
podman push ghcr.io/sm-joe/kubecis:3.0.0
```

## Running the CLI container against a cluster

A standalone container can run `kubecis scan` only when both conditions are true:

- The Kubernetes API endpoint in the supplied kubeconfig is reachable from the container's network namespace.
- The kubeconfig uses credentials available to the image. KubeCIS supports ordinary kubeconfig credentials (for example, embedded CA/client certificate data or a token). Kubeconfigs that depend on an `exec` credential plugin, such as `aws`, `az`, or `gcloud`, will not work in this minimal image because those executables are intentionally not installed.

Example for a cluster with a routable API endpoint and self-contained kubeconfig credentials:

```powershell
podman run --rm `
  -v "${env:USERPROFILE}\.kube:/home/nonroot/.kube:ro" `
  -e KUBECONFIG=/home/nonroot/.kube/config `
  kubecis:3.0.0-dev scan --benchmark auto --target all --output json --exit-code-on-fail=false
```

**Important for local Kind on Podman:** a kubeconfig endpoint such as `https://127.0.0.1:49213` refers to the container's own loopback address when KubeCIS runs inside a separate Podman container. It does not refer to the Windows host or the Kind control-plane container. Replacing it with `host.docker.internal` is not a general fix: the Kind API port must also be bound and reachable on the host interface that Podman's VM can access. Do not change the Kind cluster's API binding just to make this standalone-container path work.

For the existing local Kind cluster, use the in-cluster Job workflow below. The Job runs inside Kubernetes and uses its mounted ServiceAccount token, so it avoids the host-to-container API networking problem and does not require copying kubeconfig credentials into the image.

## Cluster/API audit deployment

Review the image reference in `deploy/job.yaml`. Apply the dedicated read-only RBAC and Job:

```powershell
kubectl apply -f .\deploy\rbac.yaml
kubectl apply -f .\deploy\job.yaml
kubectl -n kube-system wait --for=condition=complete job/kubecis-audit --timeout=10m
kubectl -n kube-system logs job/kubecis-audit
```

The Job runs the selected profile (`--target all`) and is observational (`--exit-code-on-fail=false`): findings are emitted in the report without marking the Job failed. The ClusterRole is limited to the `get`/`list` operations used by the evaluators. It does not grant mutation, Secret reads, `exec`, or impersonation permissions.

If GHCR is private, create the pull secret before applying the Job. If the package is public, remove the `imagePullSecrets` stanza from the manifest and skip the pull-secret step.

## Node-local DaemonSet deployment

Review the image reference in `deploy/node-auditor-daemonset.yaml`, then apply:

```powershell
kubectl apply -f .\deploy\node-auditor-daemonset.yaml
kubectl -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=5m
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor -c node-auditor --prefix=true --tail=100
kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor -c node-metadata-helper --prefix=true --tail=100
```

The E2E DaemonSet runs a one-shot `node-metadata-init` init container first. Kubernetes does not start the auditor or periodic helper until that init container exits successfully, so the initial metadata snapshot exists before the auditor starts; no fixed sleep is needed. The main node auditor runs as UID 65532, drops all capabilities, has a read-only root filesystem, receives no ServiceAccount token, and can read only the metadata snapshot. The helper is the explicit security exception: UID 0 with only `CAP_DAC_READ_SEARCH`, no privilege escalation, RuntimeDefault seccomp, a read-only `/var/lib/kubelet` mount, and a fixed benchmark-path allowlist. It calls `Lstat` and publishes file type/mode metadata; it never reads host file contents. Do not replace this with a privileged container or a writable host mount.

The snapshot is node-name bound and expires after two minutes. The helper refreshes it periodically; the main auditor consumes the snapshot. A missing, stale, malformed, or wrong-node snapshot produces `NOT_ASSESSED`.

## E2E validation using the existing Kind cluster

The PowerShell E2E runner is Podman-based and **reuses the existing Kind cluster**. It does not create, delete, or switch clusters. Its default context is `kind-kind`.

Prerequisites on Windows:

- Podman available and its machine/runtime running.
- `kind` and `kubectl` available in `PATH`.
- The existing Kind cluster named `kind` is running and reachable through context `kind-kind`.
- PowerShell 7 or Windows PowerShell 5.1.

Run from the repository root:

```powershell
.\scripts\e2e\Invoke-KubeCisE2E.ps1
```

If your existing context has a different name, specify it explicitly:

```powershell
.\scripts\e2e\Invoke-KubeCisE2E.ps1 -Context "YOUR-KIND-CONTEXT"
```

The runner builds and inspects the Podman image, checks the non-root image user and version, exports the image and loads it into the existing Kind cluster, applies the E2E fixtures, runs the cluster/API Job, then runs the node-local DaemonSet. It verifies that `node-metadata-init` terminated successfully and that control `4.1.9` contains an observed octal mode with a `PASS` or `FAIL` result (never `NOT_ASSESSED`) before validating the metadata-helper security boundary. The runner uses bounded readiness polling rather than a fixed 40-second delay. It uses explicit `kubectl --context` arguments and leaves the current kubeconfig context unchanged. Logs and reports are written to `artifacts\e2e`.

The E2E fixtures intentionally create security findings. The test verifies that selected expected checks are `FAIL`; the E2E Job uses `--exit-code-on-fail=false` so the Kubernetes Job completes after emitting its report. The runner replaces only its own named E2E Job when rerun. It does not delete the cluster or unrelated resources.

## Validation limits

The E2E runner validates the existing Kind cluster, not every managed Kubernetes distribution. Managed control-plane configuration and external audit-log delivery remain `MANUAL`/`NOT_ASSESSED` where the Kubernetes API does not expose evidence. The included benchmark profile is a limited, provisional CIS 2.1.0 mapping for Kubernetes minor 37, not a claim of full CIS coverage or certification.

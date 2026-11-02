# Docker image, Kubernetes deployment and E2E validation

This deployment package keeps KubeCIS in two separate execution modes:

1. **Cluster/API audit** — a short-lived Job uses a dedicated read-only ServiceAccount and Kubernetes API permissions. It does not mount the node filesystem.
2. **Node-local audit** — a DaemonSet runs an unprivileged auditor without a Kubernetes API token. A separate, narrowly scoped metadata helper can `Lstat` only the allowlisted kubelet configuration file and publishes metadata (not file contents) to a shared memory volume.

## Build the image

The multi-stage Dockerfile builds a static Linux binary with Go 1.26 and copies only the binary and benchmark profiles into the Distroless runtime image. The default runtime user is UID/GID 65532. The image has no shell or package manager.

From PowerShell at the repository root:

```powershell
docker build --pull -t ghcr.io/sm-joe/kubecis:2.9.0 .
docker run --rm ghcr.io/sm-joe/kubecis:2.9.0 version
docker run --rm ghcr.io/sm-joe/kubecis:2.9.0 --help
```

Expected version:

```text
2.9.0-consolidated-security-assessment
```

Publish to GHCR after authenticating with a package-write token:

```powershell
docker push ghcr.io/sm-joe/kubecis:2.9.0
```

## Cluster/API audit deployment

Create the GHCR pull secret in `kube-system` if the package is private, then apply RBAC and the Job:

```powershell
kubectl apply -f .\deploy\rbac.yaml
kubectl apply -f .\deploy\job.yaml
kubectl -n kube-system wait --for=condition=complete job/kubecis-audit --timeout=10m
kubectl -n kube-system logs job/kubecis-audit
```

The Job runs the full selected profile (`--target all`) and writes JSON to stdout. A non-zero process exit is expected when the audit discovers FAIL findings; retrieve and review the JSON report even when the Job phase is Failed. `NOT_ASSESSED` is not converted to PASS.

The ClusterRole is limited to `get`/`list` on the resource types used by the evaluators. It does not grant mutation, Secret reads, `exec`, or impersonation permissions.

## Node-local DaemonSet deployment

Create the GHCR pull secret if required, then apply:

```powershell
kubectl apply -f .\deploy\node-auditor-daemonset.yaml
kubectl -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=5m
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor -c node-auditor --prefix=true --tail=100
kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor -c node-metadata-helper --prefix=true --tail=100
```

The main node auditor runs as UID 65532, drops all capabilities, has a read-only root filesystem, receives no service-account token, and can read only the metadata snapshot. The helper is the explicit security exception: UID 0 with only `CAP_DAC_READ_SEARCH`, no privilege escalation, RuntimeDefault seccomp, a read-only `/var/lib/kubelet` mount, and a fixed benchmark-path allowlist. It calls `Lstat` and publishes file type/mode metadata; it never reads the host file contents. Do not replace this with a privileged container or a writable host mount.

The snapshot is node-name bound and expires after two minutes. The helper refreshes every 30 seconds; the main auditor also runs every 30 seconds. A missing, stale, malformed, or wrong-node snapshot produces `NOT_ASSESSED`.

## End-to-end tests on a disposable Kind cluster

Prerequisites on Windows:

- Docker Desktop running with Linux containers enabled.
- `kubectl` available in `PATH`.
- `kind` available in `PATH`.
- PowerShell 7 or Windows PowerShell 5.1.

Run from the repository root:

```powershell
.\scripts\e2e\Invoke-KubeCisE2E.ps1
```

The script creates a disposable Kind cluster, builds and loads `kubecis:e2e`, applies intentionally insecure fixtures, runs the cluster/API Job, validates the JSON report and the 2.4.7/2.5–2.9 check IDs, then runs the node-local DaemonSet and validates the metadata-helper boundary. It saves logs and JSON under `artifacts\e2e` and deletes the Kind cluster on completion.

To keep the cluster for manual inspection:

```powershell
.\scripts\e2e\Invoke-KubeCisE2E.ps1 -KeepCluster
```

To inspect a retained cluster:

```powershell
kubectl --context kind-kubecis-e2e get pods -A -o wide
kubectl --context kind-kubecis-e2e -n kube-system get jobs,daemonsets,pods
```

The E2E fixtures intentionally create security findings. Therefore, the cluster audit Job may finish in Failed phase because the CLI returns a non-zero exit when `Summary.Fail` is greater than zero. The test treats that as an expected finding-gate outcome only after it has retrieved the report and verified the expected check IDs and build version.

## Validation limits

The E2E script validates a disposable Kind cluster, not every managed Kubernetes distribution. Managed control-plane configuration and external audit-log delivery remain `MANUAL`/`NOT_ASSESSED` where the Kubernetes API does not expose evidence. The benchmark profile is a limited, provisional CIS 2.1.0 mapping for Kubernetes minor 37, not a claim of full CIS certification.

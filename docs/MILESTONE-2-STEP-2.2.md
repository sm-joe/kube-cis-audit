# Milestone 2 — Step 2.2: Node-level assessment

## What this adds

- A separate `kubecis node-audit` command that performs node-local checks without Kubernetes API credentials.
- The command loads an explicitly selected benchmark profile and evaluates only checks with `target: node` and kind `host_file_mode` or `host_file_exists`.
- Node identity is supplied through `--node-name` or `NODE_NAME` and is included in the JSON report.
- Host paths must be relative to the mounted host root. Path traversal is rejected, and final-component symlinks are not followed.
- Host filesystem access is read-only. Missing/inaccessible paths are `NOT_ASSESSED`, not `FAIL`.
- A Linux DaemonSet runs an unprivileged auditor and a separate metadata helper. The auditor repeats every 24 hours; the helper refreshes an allowlisted metadata snapshot every 15 seconds.
- The auditor runs as UID/GID 10001 with all capabilities dropped. The helper runs as UID 0 with only `CAP_DAC_READ_SEARCH`, no privilege escalation, RuntimeDefault seccomp, and a read-only root filesystem.
- The helper mounts only `/var/lib/kubelet` read-only, and its code allowlists only `var/lib/kubelet/config.yaml`. It uses `lstat` and publishes file type and permission bits; it never opens or reads the configuration file contents.
- The helper writes an atomic JSON metadata snapshot to a shared `emptyDir`; the auditor mounts that volume read-only. The snapshot is node-bound and rejected if older than two minutes.

## Why a separate command

The normal `scan` command uses Kubernetes API credentials and cannot infer the node's host filesystem from the workstation or from an ordinary API Pod. `node-audit` is intentionally separate: it only inspects the explicitly mounted host filesystem and does not contact the Kubernetes API.

## Run locally on a Linux node

Build the binary for Linux and mount the kubelet directory read-only at `/host-kubelet` (the helper's metadata mode is intended for the DaemonSet deployment):

```text
kubecis node-audit --benchmark cis-2.1.0 --host-root /host --node-name worker-01 --output json
```

The benchmark must be explicit because node-local mode does not query the Kubernetes API to discover the cluster version.

## Deploy in Kubernetes

1. Build and publish the updated KubeCIS image. Publish the new `milestone-2-step-2.2-metadata-helper` tag and use it in both image references in the manifest.
2. Create the GHCR pull secret in `kube-system` if the package is private.
3. Review the `/var/lib/kubelet` hostPath and the helper's `DAC_READ_SEARCH` capability with your cluster security team. Admission policies may need an explicit exception for this capability; do not enable privileged mode.
4. Apply `deploy/node-auditor-daemonset.yaml`.
5. Inspect per-node output:

```powershell
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor --all-containers=true --prefix=true
```

Each Pod reports its own `nodeName`. The DaemonSet runs on Linux nodes, including tainted nodes due to the broad toleration. If your cluster policy disallows hostPath `/var/lib/kubelet` or the helper capability, do not deploy this manifest; use an approved node-management/host assessment mechanism instead.

## Security and limitations

- The helper has `CAP_DAC_READ_SEARCH`, which can bypass discretionary read/search checks. Its host mount is limited to `/var/lib/kubelet`, and the implementation allowlists only `config.yaml`; nevertheless, this is sensitive host access and requires security review. The capability can technically read files within the mounted directory if the helper is compromised, even though the supplied code only calls `lstat`.
- The main auditor remains UID/GID 10001 with all capabilities dropped and receives only the metadata JSON through a read-only mount. Neither container is privileged, and neither receives a Kubernetes API token.
- The helper reports metadata collection errors as `NOT_ASSESSED`; it does not convert access failures into control failures.
- This first iteration assesses only checks explicitly targeted at `node`. Control-plane checks such as API-server static pod manifest permissions are not run by this DaemonSet. Managed control-plane files may not be exposed to customer workloads at all.
- Windows nodes are excluded. This iteration checks POSIX file modes only.
- A node-level PASS proves only that the inspected file's observed permission bits satisfy the profile. It does not prove the entire node is CIS compliant.
- The included Kubernetes 1.37 mapping remains provisional and the profile remains a limited starter profile, not the complete CIS benchmark.
- The DaemonSet repeats every 24 hours. Logs are held according to your cluster's container log retention policy; export them to a durable system if audit retention is required.

## Validation

```powershell
gofmt -w .\cmd\kubecis\main.go .\internal\audit\runner.go .\internal\audit\metadata.go .\internal\audit\types.go .\internal\audit\node_test.go
go test ./...
go build -o kubecis.exe ./cmd/kubecis
```

The repository's `go.mod` requires Go 1.26.0. Run validation with that toolchain. Do not lower the module's Go version to bypass the requirement.

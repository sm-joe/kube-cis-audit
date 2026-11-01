# Milestone 2 — Step 2.2: Node-level assessment

## What this adds

- A separate `kubecis node-audit` command that performs node-local checks without Kubernetes API credentials.
- The command loads an explicitly selected benchmark profile and evaluates only checks with `target: node` and kind `host_file_mode` or `host_file_exists`.
- Node identity is supplied through `--node-name` or `NODE_NAME` and is included in the JSON report.
- Host paths must be relative to the mounted host root. Path traversal is rejected, and final-component symlinks are not followed.
- Host filesystem access is read-only. Missing/inaccessible paths are `NOT_ASSESSED`, not `FAIL`.
- A Linux DaemonSet manifest runs one auditor per node and repeats every 24 hours. It has no Kubernetes API token, no host PID/network namespace, no privileged mode, no Linux capabilities, a read-only root filesystem, a read-only host-root mount, and RuntimeDefault seccomp.

## Why a separate command

The normal `scan` command uses Kubernetes API credentials and cannot infer the node's host filesystem from the workstation or from an ordinary API Pod. `node-audit` is intentionally separate: it only inspects the explicitly mounted host filesystem and does not contact the Kubernetes API.

## Run locally on a Linux node

Build the binary for Linux and mount the node filesystem read-only at `/host`:

```text
kubecis node-audit --benchmark cis-2.1.0 --host-root /host --node-name worker-01 --output json
```

The benchmark must be explicit because node-local mode does not query the Kubernetes API to discover the cluster version.

## Deploy in Kubernetes

1. Build and publish the KubeCIS image. Replace `ghcr.io/your-org/kubecis:replace-with-your-tag` in the manifest with the published image.
2. Review the hostPath exposure and your cluster's node security policy.
3. Apply `deploy/node-auditor-daemonset.yaml`.
4. Inspect per-node output:

```powershell
kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor --all-containers=true --prefix=true
```

Each Pod reports its own `nodeName`. The DaemonSet runs on Linux nodes, including tainted nodes due to the broad toleration. If your cluster policy disallows hostPath `/`, do not deploy this manifest; use an approved node-management/host assessment mechanism instead.

## Security and limitations

- The host root is mounted read-only, but the container can still see host files. This is a sensitive capability; use only the supplied constrained manifest after security review.
- The process runs as UID/GID 10001 with all capabilities dropped. If a host directory denies traversal, the result is `NOT_ASSESSED`; do not solve this by making the container privileged. A narrowly scoped, approved node-side mechanism is preferable if additional access is required.
- This first iteration assesses only checks explicitly targeted at `node`. Control-plane checks such as API-server static pod manifest permissions are not run by this DaemonSet. Managed control-plane files may not be exposed to customer workloads at all.
- Windows nodes are excluded. This iteration checks POSIX file modes only.
- A node-level PASS proves only that the inspected file's observed permission bits satisfy the profile. It does not prove the entire node is CIS compliant.
- The included Kubernetes 1.37 mapping remains provisional and the profile remains a limited starter profile, not the complete CIS benchmark.
- The DaemonSet repeats every 24 hours. Logs are held according to your cluster's container log retention policy; export them to a durable system if audit retention is required.

## Validation

```powershell
gofmt -w .\cmd\kubecis\main.go .\internal\audit\runner.go .\internal\audit\types.go .\internal\audit\print.go
go test ./...
go build -o kubecis.exe ./cmd/kubecis
```

The repository's `go.mod` requires Go 1.26.0. Run validation with that toolchain. Do not lower the module's Go version to bypass the requirement.

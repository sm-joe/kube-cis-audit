# Control-plane manifest metadata collector

## Purpose

Assess CIS control `1.1.1` using metadata for the API-server static Pod manifest. The collector calls `Lstat` only; it never opens or reads the manifest contents.

## Isolation and least privilege

- Separate DaemonSet, scheduled only to Linux nodes labelled `node-role.kubernetes.io/control-plane`.
- Dedicated read-only hostPath mount: `/etc/kubernetes/manifests` only.
- HostPath must already exist (`type: Directory`); it is never created on worker nodes.
- Helper runs as UID 0 with only `CAP_DAC_READ_SEARCH`, drops all other capabilities, and cannot escalate privileges.
- Auditor runs as UID 65532, read-only root filesystem, drops all capabilities, has no ServiceAccount token, and receives only a metadata JSON snapshot through a read-only mount.
- Strict allowlist permits only `etc/kubernetes/manifests/kube-apiserver.yaml` for target `master`.
- Snapshot includes node identity and timestamp; the auditor rejects mismatched or stale snapshots.
- No Kubernetes API permissions, Secret access, host root mount, file-content read, or remediation.

The E2E DaemonSet includes a one-shot init collector so the first snapshot exists before the auditor starts. A periodic helper refreshes the snapshot thereafter.

## Deploy and validate (PowerShell)

Build and load the image using the repository's existing Podman/Kind E2E build process. Do not recreate or switch the existing Kind cluster. Then run:

```powershell
.\scripts\e2e\Invoke-ControlPlaneMetadataE2E.ps1 `
  -Context "kind-kind" `
  -Image "kubecis:e2e-3-0-0-rc3"
```

The script applies only `e2e/controlplane-auditor-daemonset.yaml`, waits for its rollout, and requires a `master-local` report containing control `1.1.1` with `PASS` or `FAIL` and observed permission-mode evidence. If the host path is unavailable, the collector must not fabricate a result; the script fails validation.

## Report semantics

This produces a separate control-plane-local report. It does not silently rewrite or merge the cluster/API report. A future report aggregation change must define node identity, benchmark compatibility, evidence freshness, duplicate handling and summary reconciliation before merging these results.

## Limitations

The collector is for distributions exposing the kubeadm-style static-pod manifest at `/etc/kubernetes/manifests/kube-apiserver.yaml`. Managed control planes or distributions without that host path remain `NOT_ASSESSED`; absence of a Pod or host path is not a PASS. CIS mapping and applicability still require validation against the exact benchmark publication and platform.

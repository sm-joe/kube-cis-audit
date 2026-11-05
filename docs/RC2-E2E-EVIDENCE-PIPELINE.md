# RC2 E2E evidence-pipeline correction

## Defect fixed

The E2E node DaemonSet previously started `node-auditor` and
`node-metadata-helper` as concurrent containers sharing an emptyDir. The auditor
could start before the helper's first snapshot write, causing a transient
`NOT_ASSESSED` result even though a later periodic helper refresh succeeded.

## Correction

- Added a one-shot `node-metadata-init` init container.
- It uses the same metadata-only helper, exact benchmark allowlist, read-only
  `/var/lib/kubelet` hostPath mount, UID 0 plus `CAP_DAC_READ_SEARCH`, no
  privilege escalation, and the shared memory-backed metadata volume.
- It runs once (`--interval 0s`) and must exit successfully before Kubernetes
  starts either long-running container.
- Kept the periodic helper as a sidecar to refresh snapshots.
- Kept the node auditor non-root, capability-free, read-only-rootfs, tokenless,
  and limited to the snapshot volume.
- Replaced the fixed 40-second sleep with bounded readiness polling and
  assertions on init-container exit status and control `4.1.9` evidence.
- The E2E test now fails if the node-local `4.1.9` result is `NOT_ASSESSED` or
  lacks an observed octal mode. A legitimate `PASS` or `FAIL` is accepted.

## Deliberate boundary

The cluster/API Job still does not mount the host filesystem. Therefore the
cluster-level API report may retain `NOT_ASSESSED` for host-file controls such
as the API-server manifest permissions and kubelet configuration permissions.
The separate node-local report is the source of node-file metadata evidence.
No broad `/` hostPath mount, privileged container, or Secret access was added.

The API-server manifest permission control (`1.1.1`) remains unassessed until a
separately reviewed, narrowly scoped control-plane metadata collector is
implemented. This RC2 change does not claim to solve that control.

## Validation status

This package was prepared from the uploaded RC1 source. YAML and source-level
structure were inspected. The execution environment does not have PowerShell
or Go 1.26+, so the Windows PowerShell E2E script and full Go build must be
validated on the user's Windows/Go 1.27 environment before RC2 is accepted.

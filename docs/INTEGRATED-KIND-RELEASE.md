# KubeCIS Native integrated Linux node evidence release candidate

## Changes

This release changes the helper/auditor snapshot contract to schema version 2.
The root-owned helper with `CAP_DAC_READ_SEARCH` and a read-only mount of
`/var/lib/kubelet` now extracts only two allowlisted kubelet settings:
`authentication.anonymous.enabled` (boolean) and `authorization.mode`
(known mode string). No arbitrary YAML configuration is written to the snapshot.
The main auditor still runs without host access or a Kubernetes API token.
Control-plane metadata stays limited to manifest file metadata.

Snapshots are bound to node name, benchmark, target, schema version and a short
freshness interval. Missing evidence, invalid YAML, inaccessible files and
incompatible snapshots cause `NOT_ASSESSED`, not false PASS/FAIL. Existing
finding identity semantics are maintained. The old 3.0.0 helper is **not**
compatible with the updated auditor; always deploy all three containers together.

The cluster policy job remains an API-only assessment. Host-file checks are
performed by dedicated node and control-plane DaemonSets, rather than by granting
host access to the cluster scanner. These are distinct JSON reports; this release
does **not** claim to provide automatic multi-report consolidation.

## Windows 11 + Podman + Kind

From PowerShell in the extracted repository root:

```powershell
$env:KIND_EXPERIMENTAL_PROVIDER = 'podman'
kind get clusters
.\scripts\Deploy-KindNative.ps1 -ClusterName control-plane
```

The script builds with Podman, exports an archive, imports to Kind, applies RBAC
and the dedicated deployment manifests, waits for the DaemonSet rollout and
prints cluster, node and control-plane logs. It will not weaken container
capabilities or host mount permissions and does not overwrite the old rc4 Job.
The generated manifests use the image `localhost/kubecis-native:integrated` with
`imagePullPolicy: Never` and therefore require importing the image to each
applicable Kind node. This procedure is appropriate to the single-node Kind
cluster in the supplied results; multi-node deployment needs image loading on
every node (Kind's archive load normally targets all nodes).

The old 3.0.0 DaemonSet should be upgraded, not left with a 3.1.0 auditor against
an older helper. If rollout fails, inspect the Pod event/security policy logs;
do not add `privileged: true`.

## Tests

```powershell
go version
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
```

`TestSplitContainerKubeletEvidenceEndToEnd` verifies two setting checks and
file permissions through the actual helper-to-auditor JSON interface, changes
in status with stable finding IDs, redaction, and rejection of old/stale and
mismatched snapshots. `TestKubeletCollectorNeverFollowsSymlinks` verifies
that symlink targets are rejected.

## Scope and limitations

* CIS mapping for Kubernetes 1.37 / cis-2.1.0 remains **provisional**; the
  compliance percentage is withheld.
* Node-local Kubernetes version is not queried: benchmark must be explicitly
  selected, not inferred by the node helper.
* Actual Linux runtime execution, Podman build, and Kind E2E have not been
  executed in this authoring environment. Run the commands above and preserve
  results before a stable release.
* A root project license requires approval from the repository owner.
* File metadata checks and the two kubelet setting checks do not provide
  complete CIS benchmark coverage.

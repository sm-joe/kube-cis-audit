# Upstream kube-bench execution engine integration

## Pinned upstream implementation

KubeCIS builds Aqua Security's upstream `github.com/aquasecurity/kube-bench`
Go application from source at `v0.16.0`, commit
`5c6c22d51b926020e7414e1f1e051851d1a2feff`. The build does not substitute a
KubeCIS reimplementation for the upstream command interpreter.

The container build performs these operations:

1. `go install github.com/aquasecurity/kube-bench@v0.16.0` compiles the pinned
   upstream source.
2. The exact module's `cfg/`, `LICENSE`, `NOTICE`, and source tree are retained
   in the image.
3. KubeCIS executes `/usr/local/bin/kube-bench` with the pinned configuration
   directory and JSON output enabled.
4. KubeCIS maps each upstream control result into its report contract. The
   upstream control ID, target, score flag, status, actual/expected values,
   reason, and remediation are retained.
5. Audit-command execution errors are normalized to `NOT_ASSESSED` rather than
   being presented as evidence that a configuration setting is insecure.
6. `hybrid` mode takes numbered CIS results from kube-bench and adds only
   KubeCIS extension controls, avoiding duplicate scoring of overlapping
   numbered controls.

## Build on Windows PowerShell

Run from the repository root with Go 1.27 installed and available on `PATH`:

```powershell
$PSVersionTable.PSVersion
 go version

.\scripts\install-upstream-engine.ps1

go test ./... -count=1
go vet ./...
go build -trimpath -o .\bin\kubecis.exe .\cmd\kubecis
```

The install script places `kube-bench.exe` beside `kubecis.exe` under `bin/`,
imports the matching upstream configuration corpus, and copies the pinned
upstream source into `third_party/kube-bench/source/`.

Run the upstream engine through KubeCIS:

```powershell
.\bin\kubecis.exe scan `
  --engine upstream `
  --benchmark auto `
  --upstream-benchmark auto `
  --output json `
  --exit-code-on-fail=false
```

Run the combined upstream CIS plus KubeCIS extension assessment:

```powershell
.\bin\kubecis.exe scan `
  --engine hybrid `
  --benchmark auto `
  --upstream-benchmark auto `
  --output json `
  --exit-code-on-fail=false
```

`--benchmark` selects the existing KubeCIS extension profile. The separate
`--upstream-benchmark` accepts kube-bench profile identifiers such as
`cis-1.12`, `eks-1.8.0`, `gke-1.9.0`, or `k3s-cis-1.24`. `auto` delegates
Kubernetes-version and platform resolution to upstream kube-bench.

## Host evidence requirements

The upstream engine is designed to inspect the node on which it executes. Its
host checks require access to the host process namespace and read-only host
configuration paths. Running it inside an ordinary unprivileged Job without
those mounts does not assess the Kubernetes node. The upstream project's
reference Job uses `hostPID` and read-only hostPath mounts; managed control
planes remain inaccessible where the provider does not expose their nodes.

The KubeCIS image runs as UID 65532 by default. A node-local upstream scanner
must explicitly opt into the minimum host access required by the upstream
checks: host PID namespace, UID 0 for reading protected host configuration,
read-only hostPath mounts, no privileged mode, no privilege escalation, and
all Linux capabilities dropped. Do not silently broaden mounts or privileges.

Kube-bench's policy checks use `kubectl`. When KubeCIS is running in a Pod, it
creates a short-lived kubeconfig that references the mounted ServiceAccount
token file and CA file; the token is not copied into the kubeconfig or report.
The ServiceAccount must have only the read permissions needed by the selected
upstream policy checks. When `readOnlyRootFilesystem: true` is enabled, mount a
writable `emptyDir` at `/tmp` because KubeCIS creates a short-lived, mode-0600
kubeconfig there. For an explicit `--context`, KubeCIS writes a separate
mode-0600 temporary kubeconfig with that context selected; the original file is
not modified, and the temporary file is deleted after the scan.

## Important profile-selection behaviour

The upstream `v0.16.0` global configuration maps Kubernetes 1.32–1.34 to
`cis-1.12`. For newer Kubernetes versions, kube-bench falls back to the nearest
lower version mapping. This is upstream's compatibility behaviour; it does
not mean a newer CIS benchmark release has been selected. Keep the upstream
profile identifier and detected Kubernetes version visible in the report.

## Validation status

The integration includes unit tests for status normalization, config validation,
report aggregation, and fake-engine JSON normalization. A release must still
pass `go test ./...`, `go vet ./...`, the container build, and node-level E2E
checks on the supported distributions before being labelled stable. A
successful build alone does not prove that every control is assessable on every
managed Kubernetes platform.

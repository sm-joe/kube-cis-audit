# Dockerization and E2E validation status

## Completed in this package

- Multi-stage Go 1.26 build and Distroless runtime image, default UID/GID 65532.
- Updated cluster audit Job and read-only least-privilege ClusterRole for all API resources used by the included evaluators.
- Node-local DaemonSet split into an unprivileged auditor and a metadata-only helper.
- Wired `node-metadata` CLI command and metadata snapshot consumption into `node-audit`.
- Metadata snapshots are restricted to the allowlisted kubelet `config.yaml`, contain metadata only, are bound to a node name, and expire after two minutes.
- Disposable Kind fixtures and PowerShell E2E runner for cluster/API and node-local modes.

## Validation performed in this authoring environment

| Check | Result |
|---|---|
| YAML parsing (8 files) | PASS |
| Benchmark IDs (20 unique checks) | PASS |
| Deployment RBAC verbs limited to `get`/`list` | PASS |
| No Secret-read permission in auditor ClusterRole | PASS |
| Node auditor/helper security-context split | PASS |
| Go formatting (`gofmt`) | PASS |

## Not run here

- `go test ./...`: blocked because the environment has Go 1.23.2 and the module requires Go 1.26.0; automatic toolchain download failed because network access is unavailable.
- Docker image build/run: Docker is not installed in this environment.
- Kind/Kubernetes E2E: `kind` and `kubectl` are not installed, and no disposable cluster is available.

Run the exact Windows PowerShell commands in `docs/DOCKER-AND-E2E.md`. The E2E runner writes cluster and node logs/reports to `artifacts\e2e` and reports a failure if expected checks or security boundaries are missing.

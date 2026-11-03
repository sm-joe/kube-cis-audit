# Changelog

## v3.0.0 — CLI-first product release

- Added kubeconfig context selection with `kubecis scan --context`.
- Added `--output-file` for text and JSON reports.
- Added `--exit-code-on-fail` control while preserving the default CI-gate behavior.
- Added root `--version` support and version injection for release builds.
- Made benchmark discovery work from the executable's installation directory.
- Added a Windows PowerShell installer with SHA-256 verification.
- Added a consolidated PowerShell build, vet and test script.
- Added GitHub Actions CI and multi-platform release workflows.
- Updated the cluster Job for observational execution.
- Updated the node DaemonSet with an initial metadata snapshot init container and the validated 15-second metadata refresh / 24-hour node audit cadence.
- Preserved the non-root node auditor and the separate narrowly privileged metadata helper.

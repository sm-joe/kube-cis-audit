# Changelog

## Unreleased — control-plane local report accounting

- Populate compliance counters for node-local reports using only the supported host-file checks in the selected local target.
- Add control-plane metadata regression assertions for both compliant (`0600`) and non-compliant (`0644`) API-server manifest modes.
- Mark the legacy combined `cis-2.0.1` pod-security check as `MANUAL` because no equivalent combined native evaluator exists.
- Refresh README and release-readiness documentation to reflect the separate control-plane collector and current provisional benchmark limitations.


## 3.0.0-rc2 (candidate)

- Fix E2E node metadata startup race with a synchronous init-container snapshot.
- Replace fixed 40-second E2E delay with bounded readiness and evidence assertions.
- Preserve the metadata-only, read-only helper boundary and keep cluster Job host mounts absent.
- API-server manifest metadata remains an explicit release limitation pending a separately scoped collector.


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

## Unreleased — control-plane manifest metadata collector

- Added a control-plane-only metadata collection path for CIS control `1.1.1`.
- Added strict target/path allowlisting and node-identity/freshness-validated metadata snapshots.
- Added dedicated control-plane E2E and deployment DaemonSets. The collector mounts only `/etc/kubernetes/manifests` read-only and records file metadata without reading manifest contents.
- Kept the cluster/API auditor unprivileged and unchanged; control-plane evidence is emitted as a separate `master-local` report.

## 3.1.0-native-integrated-rc3 — consolidation release candidate

- Added validated `consolidate` CLI for cluster policy, per-node and per-control-plane JSON reports.
- Added PowerShell Kind report collection workflow and separate overall/scored coverage metrics.
- Added missing/stale/duplicate-source regression tests.
- Kept CIS score withheld for provisional incomplete benchmark mapping.
- Production release gates documented in `RELEASE-READINESS-NATIVE-RC3.md`.

# RC4 candidate patch notes

This patch hardens release behavior and corrects release metadata. It does not claim to complete the CIS benchmark.

- Added a release validation job that checks tag syntax, Go formatting, `go vet`, all Go tests, and a CLI build.
- Made binary and container publication depend on release validation.
- Marked hyphenated tags as GitHub prereleases and prevented them from updating the GHCR `latest` tag.
- Restricted the stable `latest` tag to `vMAJOR.MINOR.PATCH` tags.
- Updated the deployment Kustomize image override to `v3.0.0-rc4`.
- Clarified in README and release documentation that the current 20-entry profile is a limited starter profile, not full CIS Kubernetes Benchmark coverage.
- Added production compliance release gates and retained RC4 findings that must not be suppressed.

## Validation executed in the patch environment

- Parsed all 303 YAML/YML files (including multi-document manifests): PASS.
- Checked the benchmark profile has 20 unique entries and all `KUBECIS-*` extensions are unscored: PASS.
- Checked release workflow binary and container jobs depend on validation: PASS.
- Go formatting/vet/tests/build: NOT RUN; available Go is 1.23.2 and the module requires 1.26.0. Automatic toolchain download failed because network DNS is unavailable.
- Kustomize render: NOT RUN; `kubectl` is unavailable in the patch environment.

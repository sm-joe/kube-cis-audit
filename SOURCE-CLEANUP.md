# Source cleanup inventory — KubeCIS Native RC3.2

This archive is a **cleaned product source distribution**, prepared from the user's uploaded repository. The original ZIP was not modified.

## Retained

- Native Go CLI, library packages and all remaining relevant Go unit and integration tests
- KubeCIS `cis-2.0.1` and `cis-2.1.0` provisional profiles (not complete CIS benchmarks)
- Native Kind/Podman and GHCR deployment manifests, least-privilege RBAC and E2E test fixtures
- Working PowerShell build, deploy, consolidation, installation and packaging scripts
- Go module/sum, Dockerfile, CI and release workflows
- User, security, architecture/consolidation and release-scope documentation

## Excluded

- `.git/` private/local history, objects and hooks
- `bin/kube-bench.exe`, `test-results/kubecis.exe`, generated image tar and other binaries
- `artifacts/`, `reports/`, `test-results/`, generated Kubernetes manifests and scan output
- `eks-test/` Terraform state and developer infrastructure examples containing local values
- Entire vendored upstream kube-bench benchmark corpus, unused `upstream/` parser package,
  legacy `internal/audit/kubebench*.go`, and old upstream-only CLI utilities
- Legacy hybrid deployment YAML, milestone patch documents and obsolete build scripts

## Preserved functionality and limitations

Existing native code, its public report structures and comparison/consolidation logic were not rewritten.
The deprecated legacy CLI flags remain for actionable error reporting. Some public field names
retain the historical `upstream` term for backward JSON compatibility. No root `LICENSE`
was present; the maintainer must authorize and add a root open-source license before publication.

## Validation

The archive is ZIP-checked and structurally audited. The local sandbox has Go 1.23.2;
this module requires Go 1.26+, so no fresh full Go compilation/tests or Podman/Kind
execution can be claimed for this cleaned archive. Run those checks on the extracted source.

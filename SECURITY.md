# Security Policy

## Supported releases

Security fixes target the latest stable release of KubeCIS. Use the Linux AMD64 or ARM64 distribution from GitHub Releases and keep the CLI and node auditor on the same version.

## Report a vulnerability

Use [GitHub private vulnerability reporting](https://github.com/sm-joe/kube-cis-audit/security/advisories/new). If private reporting is unavailable, contact the repository maintainer privately before publishing exploit details. Do not include credentials, tokens, raw Secrets, kubeconfigs or sensitive live reports in public issues.

Include the affected version and digest, deployment type, reproduction steps, expected and observed behavior, impact, and sanitized evidence. The maintainers will triage reports and coordinate disclosure and remediation; no guaranteed response SLA is implied.

## Security architecture

- API assessments are read-only. Installation creates dedicated Kubernetes resources and requires a separately authorized operator.
- Node helpers use bounded read-only mounts and allowlisted metadata. The main scanner runs as a non-root user; do not broaden privileges to compensate for missing evidence.
- AWS evidence uses the operator's scoped IAM identity. Never grant broad permissions merely to obtain a passing result.
- Treat nodes/proxy as sensitive access. Missing or denied evidence is not a passing assessment.
- Reports may contain sensitive infrastructure identifiers and settings. Restrict permissions, redact before sharing and avoid committing reports.

## Supply chain

Verify Sigstore signatures, SHA-256 checksums and the signed release manifest before installing. Deploy containers by immutable digest. CI runs tests and supply-chain security checks; release workflows produce signed archives and multi-architecture images with SBOM and provenance.

## Scope

KubeCIS reports configuration findings and remediation guidance. It does not automatically enforce or remediate cluster settings. Operators are responsible for authorizing scans and reviewing findings before making changes.

See [release scope](docs/RELEASE-SCOPE.md) and [EKS deployment](deploy/eks/README.md).

# Linux EKS collection and acceptance

## Support classification

This is an **engineering preview**. The scanner runs natively on Linux, and
the Bash adapter collects API, EKS managed EC2 Linux node metadata and AWS
DescribeCluster evidence. Full EKS v2.1.0 compliance is **not** established;
the official score remains withheld even when preview acceptance succeeds.
A real Linux-to-EKS acceptance run is still required.

## Dependencies and permissions

Linux with Bash, GNU `date`, `jq`, `kubectl`, AWS CLI v2, and a
`kubecis` CLI built from the exact version as the published DaemonSet image.
The selected Kubernetes context must resolve to the named EKS cluster's API
endpoint and Kubernetes minor version. The AWS identity needs
`eks:DescribeCluster`. Kubernetes access needs read-only workload, cluster
and node inventory rights; cluster administration rights are necessary **only**
for initial DaemonSet deployment. The optional `nodes/proxy` permission
exposes kubelet configuration data and should be tightly scoped.

## Operating procedure

Review `deploy/eks/rbac.yaml` and the DaemonSet's hostPath mounts, then run
the digest-pinned installer. This operation changes Kubernetes resources.
Replace placeholder values before execution.

```bash
bash scripts/deploy-eks.sh --context EKS_CONTEXT \
  --image ghcr.io/sm-joe/kubecis@sha256:YOUR_VERIFIED_64_CHARACTER_DIGEST

bash scripts/collect-eks.sh --context EKS_CONTEXT \
  --cluster-name EKS_CLUSTER_NAME --region us-east-1 \
  --binary ./kubecis --output-dir ./reports/eks
```

Each supported worker must have exactly one running auditor Pod. The collector
verifies matching local/remote CLI versions, uses node-local read-only metadata
snapshots, optionally requests live kubelet `/configz` via Kubernetes and
deletes that sensitive raw response. It collects AWS provider configuration,
consolidates reports with a 30-minute freshness limit, and validates the
provisional release contract before publishing output.

The Linux provider adapter intentionally omits optional access-entry, VPC CNI
and ECR enrichment: their status is **UNAVAILABLE**, not PASS.
Do not equate its reduced provider coverage to the richer, separately
implemented PowerShell adapter. Do not publish a full CIS score.

## Verified Linux archive installation and rollback

Download the platform release archive, `SHA256SUMS`, and the archive's
matching `.sigstore.json` bundle from the same GitHub release tag. Verify the
release archive before deploying its matching image digest. Use non-root user
directories for the CLI and benchmarks:

```bash
# Example for an x86-64 Linux host. Replace the tag and use real downloaded assets.
bash scripts/install-linux.sh \
  --archive kubecis_linux_amd64.tar.gz \
  --checksums SHA256SUMS \
  --bundle kubecis_linux_amd64.tar.gz.sigstore.json \
  --release vX.Y.Z-rcN

"$HOME/.local/bin/kubecis" version

# After installing a different signed version, restore the prior version:
bash scripts/install-linux.sh --rollback
```

This installer checks architecture, checksum, Sigstore certificate identity,
OIDC issuer, benchmark bundle and binary version; then switches an installation
symlink. It retains the previous version for rollback. It uses a wrapper that
points `KUBECIS_CONFIG_DIR` at the active release, so scans do not depend on
the current working directory.

Run `bash scripts/test-linux-install.sh` for mocked-signature install,
upgrade and rollback functional testing. **It does not verify Sigstore
cryptography**; the tagged release workflow must successfully sign and verify
the actual published artifacts. Neither CLI rollback nor an archive smoke test
establishes safe Kubernetes DaemonSet upgrades; those require a live EKS
rollout, canary and rollback acceptance.

## Evidence files

- `cluster.json`: Kubernetes API findings
- `node-*.json`: per-managed-worker findings
- `provider.json`: bounded AWS DescribeCluster inventory
- `consolidated.json`: merged report with withheld score
- `results.txt`: readable findings
- `review-template.json`: human review tasks
- `final-assessment.json`, `final-assessment.txt`: original automated findings
  plus review status; these do **not** constitute CIS certification

## Go/no-go gates

- Source security, Go build/vet/test, shell contract checks must pass CI.
- All Linux managed EC2 worker nodes must be included. Mixed compute modes
  are deliberately rejected.
- A live, authorized EKS test must prove image rollout, API/source collection,
  positive and negative fixture outcomes, output freshness, cleanup and
  reproducibility from published digests.
- The independent official CIS EKS recommendation inventory, applicability,
  every evaluator mapping, and proof of correctness are still required.
- `--release-level full` on the preview validator is intentionally rejected.

Validate syntax and fail-closed contract checks on Linux with:

```bash
bash scripts/test-linux-eks.sh
go test ./... -count=1
go vet ./...
```

Do not store kubeconfigs, raw kubelet configz, AWS credentials or unsanitized
private reports in a public repository. Handle assessment dossiers as security
sensitive files.

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
The collector builds `./kubecis` automatically when it is missing and Go is
installed. `./cmd/kubecis` is a source directory; it is not an executable.
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
  --output-dir ./reports/eks
```

Each supported worker must have exactly one running auditor Pod. The collector
verifies matching local/remote CLI versions, uses node-local read-only metadata
snapshots, optionally requests live kubelet `/configz` via Kubernetes and
deletes that sensitive raw response. It collects AWS provider configuration,
consolidates reports with a 30-minute freshness limit, and validates the
provisional release contract before publishing output.

Linux now collects bounded, read-only ECR deployed-image evidence for active
Pod containers, init containers and ephemeral containers. It requests
`ecr:GetRegistryScanningConfiguration` and
`ecr:DescribeImageScanFindings` for resolved deployed digests. It does not
copy raw vulnerability descriptions or image contents into reports. A complete
Pod inventory is mandatory; unavailable AWS image scan data is preserved as
MANUAL/NOT_ASSESSED, and non-ECR images require third-party review. Denied
Pod inventory is explicitly UNAVAILABLE.

EKS access-entry and managed VPC CNI AWS inventories remain unavailable in
the Linux adapter. This is less provider coverage than the PowerShell collector.
The official score must remain withheld.

Kubelet metadata schema v5 adds a bounded parse of the active allowlisted
`/var/lib/kubelet/kubeconfig`, recording only whether the actual client
authenticates with a recognized AWS IAM exec-token provider, a client
certificate/key pair, or an unknown mechanism. It never exports credentials.
The official 3.2.7 check is not applicable to IAM-token-only kubelet clients,
but the existing report model records that as MANUAL pending reviewer
disposition; no automatic PASS is issued. Client-cert nodes require explicit
rotation evidence. A missing /configz field must not erase trusted host data,
and unknown remains NOT_ASSESSED.

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

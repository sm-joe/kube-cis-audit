# EKS release evidence and acceptance

The 47-control EKS v2.1.0 inventory is implemented through automated or review routes. Independent procedure conformance and official scoring remain blocked. A successful scan can contain FAIL, MANUAL or NOT_ASSESSED findings.

## Source conformance

Keep the licensed benchmark PDF outside the repository. Run the source inventory verifier, evaluator tests and ledger verifier:

```bash
python3 scripts/verify-eks-210-source.py --pdf /private/path/to/benchmark.pdf
go test ./... -count=1
python3 scripts/test-eks-evaluator-conformance.py
python3 scripts/verify-eks-evaluator-conformance.py
python3 scripts/verify-eks-evaluator-conformance.py --print-evaluator-digest
```

The independent reviewer records actual per-control evidence and signs the exact ledger including evaluatorSourceSha256. Changes to evaluator, test, benchmark, deployment and collector sources invalidate the stable approval. The full gate also requires the pinned reviewer key and its signature. Never populate verification fields from synthetic or invented runtime evidence.

## Signed release acceptance (Linux)

A tagged release publishes a signed release-manifest.json linking the tag, source commit and immutable OCI image digest, plus signed archives and SHA256SUMS. Download these from the same release. With Cosign, jq, AWS CLI, kubectl and GNU tools installed, verify without touching the cluster:

```bash
bash scripts/accept-signed-eks-release.sh --release "$TAG" --assets "$ASSETS" --verify-only
```

Only after this succeeds, deploy that specific verified image using the matching archive's deploy script (this updates the node-auditor Kubernetes resources). Then run:

```bash
bash scripts/accept-signed-eks-release.sh --release "$TAG" --assets "$ASSETS" \
  --context "$EKS_CONTEXT" --cluster-name "$EKS_CLUSTER" --region "$AWS_REGION" \
  --output-dir reports/signed-release
```

Use a new output directory. The helper checks deployment readiness and exact image selection, performs a version-matched scan using the signed archive, and records report and archive digests in release-acceptance.json. This is an operator runtime record, not a signed independent CIS conformance decision. Do not commit sensitive reports.

Stable tags remain blocked until independent approval and the verified benchmark compatibility mapping are established. A signed preview can be used for artifact acceptance without claiming full CIS coverage.

# Trusted third-party EKS image scan attestations

CIS EKS 2.1.0 control 5.1.1 permits ECR or third-party scanning of
**deployed image digests**, with findings available for review. AWS EKS
system add-on images can reside in externally owned ECR accounts where
image pull access does not grant scan-read access.

## Security model

KubeCIS accepts an external claim **only** when a detached SHA-256/RSA
signature verifies against an explicitly pinned, operator-approved RSA
public key. The signed JSON must match exact deployed registry, repository
and digest, include a real scanner/run ID, be less than 30 days old,
declare findings available, list the scanned child/platform digests, and
bind the **actual raw scanner report** through its SHA-256.

This establishes an auditable trust chain, **not** a guarantee of zero
vulnerabilities. The scanner operator must independently confirm
that all required runtime platforms and the right artifact are scanned,
and approve the signing identity through change control. Do not put
private keys in the cluster, this repository, or collector environments.
The raw findings stay outside provider.json.

## Collect a real scan and prepare the signed claim

Choose a deployed registry/repository/digest from provider.json. The
following uses illustrative values; replace digest with a real one.

\`\`\`bash
export REGISTRY="602401143452.dkr.ecr.us-east-1.amazonaws.com"
export REPOSITORY="eks/coredns"
export DEPLOYED_DIGEST="sha256:REPLACE_WITH_64_HEX_CHARS"
export CLAIM_DIR="/secure/kubecis/claims"
export HEX="${DEPLOYED_DIGEST#sha256:}"
mkdir -p "$CLAIM_DIR"

# Trivy is an example. Use an independently approved scanner/run.
trivy image --platform linux/amd64 --format json \
  --output "$CLAIM_DIR/$HEX.scan.json" \
  "$REGISTRY/$REPOSITORY@$DEPLOYED_DIGEST"
sha256sum "$CLAIM_DIR/$HEX.scan.json"
\`\`\`

An authorized signing workflow creates \`$CLAIM_DIR/$HEX.json\` with
these fields; substitute the real values of all placeholders:

\`\`\`json
{
  "schemaVersion": 1,
  "subject": {
    "registry": "602401143452.dkr.ecr.us-east-1.amazonaws.com",
    "repository": "eks/coredns",
    "digest": "sha256:EXACT_DEPLOYED_DIGEST"
  },
  "scanner": "approved-scanner",
  "scanId": "real-scan-run-identifier",
  "status": "COMPLETE",
  "findingsAvailable": true,
  "scannerReportSha256": "SHA256_OF_ACTUAL_SCAN_REPORT_BYTES",
  "scannedAt": "2026-11-19T12:00:00Z",
  "scannedDigests": ["sha256:EXACT_SCANNED_PLATFORM_MANIFEST_DIGEST"]
}
\`\`\`

Do not sign an unsuccessful or incomplete scan. For an OCI image
index, the approved operator must attest platform-specific manifests
actually scanned, not merely the index digest.

Sign the exact JSON bytes with the **approved scanner private key**:

\`\`\`bash
openssl dgst -sha256 \
  -sign /secure/approved-scanner/private.pem \
  -out "$CLAIM_DIR/$HEX.sig" "$CLAIM_DIR/$HEX.json"

sha256sum /secure/approved-scanner/public.pem
\`\`\`

The public-key digest must be pinned from an independent trusted
approval process, not taken uncritically from a submitted claim.
Enable the feature explicitly:

\`\`\`bash
export KUBECIS_ATTESTATION_DIR="/secure/kubecis/claims"
export KUBECIS_ATTESTATION_PUBLIC_KEY="/secure/approved-scanner/public.pem"
export KUBECIS_ATTESTATION_KEY_SHA256="64_HEX_CHARACTERS_OF_APPROVED_PUBLIC_KEY"

bash scripts/collect-eks.sh \
  --context "$EKS_CONTEXT" --cluster-name "$CLUSTER_NAME" \
  --region "$AWS_REGION" --output-dir reports/eks
\`\`\`

Valid evidence yields \`scanType=THIRD_PARTY\`, \`scanStatus=COMPLETE\`,
\`error=""\`, and signer/scanner/scan-date metadata. Missing evidence
preserves the AWS error; invalid evidence yields
\`EXTERNAL_ATTESTATION_INVALID\`. No credentials or scan report data
are copied into provider.json.

## Independent benchmark conformance gate

\`\`\`bash
python3 scripts/verify-eks-evaluator-conformance.py
python3 scripts/verify-eks-evaluator-conformance.py --require-full
\`\`\`

The first command checks all official source IDs and the truthful
implementation status; the second requires each of the 47 mappings to
have independently reviewed official audit procedures, positive/negative
fixtures, runtime evidence and a named reviewer. It **fails** while
any remain unverified or the independent signed reviewer approval is absent.

For an independently approved full-conformance release, the reviewer must
provide signed approval of the exact ledger bytes plus SHA-256-bound positive,
negative and fresh runtime evidence files for every control:

```bash
python3 scripts/verify-eks-evaluator-conformance.py --require-full \
  --approval-public-key /approved/reviewer-public.pem \
  --approval-key-sha256 YOUR_INDEPENDENTLY_PINNED_64_HEX_SHA256 \
  --approval-signature /approved/eks-2.1.0-evaluator-review.sig
```

The reviewer signing key must never be created or stored by the build
pipeline. A self-signed local file is **not** independent verification. A
successful structural audit without `--require-full` is not release approval.

Current audit baseline: 47 official control identities with concrete node/API/AWS
evaluator or evidence-bound human-review routes, **47 candidates, zero independently
accepted evaluator mappings**. A candidate is not a verified implementation of
every audit procedure, nor a completed human security review.
This is a gap register rather than a full verified CIS EKS implementation.
Original licensed CIS source remains private. Full compliance score:
WITHHELD. No CIS certification claim is made.

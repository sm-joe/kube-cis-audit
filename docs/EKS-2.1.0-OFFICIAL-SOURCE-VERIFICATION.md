# Official CIS Amazon EKS v2.1.0 Source Verification

## Independent source

The operator-provided, licensed **CIS Amazon Elastic Kubernetes Service (EKS)
Benchmark v2.1.0** was read privately on 2026-11-19. The exact
document SHA-256 is:

```text
c3c9eb01253643a0176b21aafd19cc7c08af0a19872f869712d2d82be8a967e9
```

The benchmark itself is **not in this repository or the distribution** because
the source's redistribution terms must be respected. Source inventory checking
uses only factual recommendation IDs and Automated/Manual classification.

| Structural source property | Independently verified |
| --- | --- |
| Distinct EKS 2.1.0 recommendations in official Appendix Summary Table | 47 |
| Recommendations marked Automated | 29 |
| Recommendations marked Manual | 18 |
| Controls in local EKS profile | 41 |
| Controls derived separately from AWS provider evidence | 6 |
| Duplicate / missing IDs in source inventory | None |

The operator can rerun the source-independent verifier without committing the
PDF anywhere:

```bash
# Install poppler / pdftotext if required. Keep the original PDF private.
python3 scripts/verify-eks-210-source.py \
  --pdf "/private/path/to/CIS_Amazon_Elastic_Kubernetes_Service_(EKS)_Benchmark_v2.1.0.pdf"
```

The program requires the exact source digest and checks the repository's
catalog and benchmark ID sets, plus assessment classes. The Go unit test
`TestPinnedEKS210SourceControlIdentitiesAndClasses` independently pins the
IDs/types without storing licensed benchmark text.

## What this verifies — and what it does not

**Verified:** The official Appendix recommendation inventory and its
Automated/Manual classes agree with the repository's 47-ID catalog, including
the distinct AWS and node/API source split.

**Not yet established:** Every check's precise applicability and audit
procedure has been independently accepted against supported EKS node OS
versions, IAM/authentication modes, effective kubelet options, and negative
runtime fixtures. ECR/image scanner results and policy/manual decisions depend
on the AWS/Kubernetes account permissions and deployed image inventory.
Manual controls still need authenticated review and source-bound evidence.

The source clearly distinguishes the client certificate rotation control
`3.2.7` on client-certificate nodes from IAM exec-token-only EKS nodeadm
clients. The engine records IAM-client evidence without exporting secrets,
and uses MANUAL applicability review in the current outcome model rather than
falsely issuing PASS. The ECR control `5.1.1` requires evidence of actual
deployed-image scan results, not just a registry scanning flag. Both rules
are implemented in the native evaluator/adapter and covered by tests, but
they still require live acceptance on matching versioned artifacts.

## Full-CIS release gate remains fail-closed

The current benchmark profile must retain
`compatibilityStatus: provisional` and the official compliance score
`WITHHELD` until independent evaluator conformance, supported version
matrix, manual reviews, signed artifact publication and end-to-end acceptance
are documented and verified. An ID catalog match or a successful source
inventory test **cannot** be used to claim CIS tool certification.

For a live EKS preview, run `bash scripts/collect-eks.sh` and separately
review all MANUAL and NOT_ASSESSED findings. The operator's test cluster
settings should not be changed to manufacture a clean assessment.

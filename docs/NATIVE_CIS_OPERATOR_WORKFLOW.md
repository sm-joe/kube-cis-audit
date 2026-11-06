# KubeCIS Native operator workflow (development branch)

This CLI is **not a completed CIS compliance product**. It does not execute or require kube-bench. The EKS 2.1.0 assessment remains provisional until all official automated controls are natively evaluated, all designated manual recommendations have an auditable review outcome, and live EKS + generic Kubernetes acceptance is complete.

## Read-only evidence collection

For EKS, keep the existing validated three-source workflow: Kubernetes API scanner, managed EC2 node metadata auditor, and read-only AWS EKS DescribeCluster evidence collector. Use `scripts/Collect-ConsolidatedEKSReport.ps1` and `scripts/Validate-EKSRelease.ps1` as shipped with a **matching** binary/image revision. The collector produces a consolidated JSON report. Do not assume a kubeconfig of a connected Kubernetes cluster is the same as proof that the host files were accessed.

Do not replace the managed EKS cluster configuration just to make tests pass.

## Human-readable report

After producing a fresh `consolidated.json` from the matching release candidate:

```powershell
.\kubecis.exe render --input .\consolidated.json --output-file .\kubecis-results.txt
```

Alternatively `kubecis consolidate --format text` renders human-readable results directly when all required cluster/local/provider inputs are supplied. `--format json` remains the default and is required for machine processing.

The text renderer groups controls by CIS control prefix, preserves node scopes and evidence-unknown states, prints per-section remediation text where configured, and provides pass/fail/warn/manual/not-assessed counts. These are **observations**, not an official compliance score.

## Manual review and exception attestation

Generate a review template whose `reportSha256` binds to the exact consolidated JSON bytes:

```powershell
.\kubecis.exe review-init --input .\consolidated.json --output-file .\review-template.json
```

The template's `pendingReviews` is a list of MANUAL findings. Do not modify its `reportSha256`. Add reviewed findings to the `decisions` array. The following is an illustrative decision only, not a recommendation to approve any control:

```json
{
  "id": "4.1.1",
  "decision": "APPROVED",
  "reviewer": "security-reviewer",
  "evidenceRef": "internal-change-or-audit-record",
  "justification": "Documented scope and evidence reviewed",
  "reviewedAt": "2026-11-06T14:00:00Z"
}
```

Use `REJECTED` or `EXCEPTION` where appropriate. For node-specific findings include the exact `scope` value from the report. Leave unreviewed findings out of `decisions`; they remain pending.

```powershell
.\kubecis.exe review --input .\consolidated.json --decisions .\review-template.json --output-file .\review-validated.json
```

Validation fails if the source JSON changes, an unreviewed/not-implemented/native FAIL control is being overridden, a decision is duplicated, or reviewer/evidence/justification/timestamp is missing. A completed review **does not** overwrite the native scan status or convert NOT_ASSESSED into PASS.

## Acceptance gates before release

- Every official EKS 2.1.0 and generic Kubernetes 2.1.0 control has a verified native evaluator or designated manual-review path, with benchmark-ID parity tests.
- Linux nodeadm/kubeadm metadata collectors are exercised against real eligible nodes and fail closed for inaccessible data.
- AWS provider evidence is fresh, scoped to the correct cluster and IAM permissions are least-privilege.
- Reports are deterministic and include correct evidence, detailed original remediations, exception records, and accurate score withholding.
- `go fmt`, `go vet`, full tests, security workflows, Windows CLI, Linux multi-architecture image, and cross-platform acceptance pass.
- Kind and live EKS 1.35 integration runs validate all 47 controls; generic Kubernetes version-targeted integration validates its full official catalog.

Until these conditions hold, do not mark benchmarkCompatibilityStatus verified, release a compliance score, or publish a final GitHub release.

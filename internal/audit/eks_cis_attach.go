package audit

import (
    "fmt"
    "math"
    "sort"
    "time"
)

// AttachEKSCISProviderEvidence attaches native CIS Amazon EKS 2.1.0 findings to
// a previously validated API+worker consolidated report. Each finding remains
// traceable to its AWS source; this does not override provisional benchmark
// status or permit a compliance score without full verified coverage.
func AttachEKSCISProviderEvidence(out *ConsolidatedReport, evidence EKSProviderEvidence) error {
    if out == nil || out.CloudProvider != "eks" || out.AuditMode != "consolidated" {
        return fmt.Errorf("validated consolidated EKS report required")
    }
    if out.ProviderEvidence != nil {
        return fmt.Errorf("AWS EKS evidence already attached")
    }
    findings, err := EvaluateEKSCISProviderControls(evidence)
    if err != nil { return err }
    existing := make(map[string]bool, len(out.Results))
    for _, r := range out.Results { existing[StableFindingID(r)] = true }
    for i := range findings {
        findings[i].EvidenceSource = "aws-eks-describe-cluster"
        findings[i].FindingID = StableFindingID(findings[i])
        if existing[findings[i].FindingID] {
            return fmt.Errorf("duplicate CIS EKS finding %s", findings[i].ID)
        }
    }
    out.ProviderEvidence = &evidence
    for _, r := range findings {
        out.Results = append(out.Results, r)
        incrementSummary(&out.Summary, r.Status)
        out.CoverageMetrics.Total++
        if r.Status == "PASS" || r.Status == "FAIL" {
            out.CoverageMetrics.Assessed++
        }
        if r.Scored {
            out.Compliance.ScoredTotal++
            switch r.Status {
            case "PASS": out.Compliance.ScoredPass++
            case "FAIL": out.Compliance.ScoredFail++
            case "WARN": out.Compliance.ScoredWarn++
            case "MANUAL": out.Compliance.ScoredManual++
            case "NOT_ASSESSED": out.Compliance.ScoredNotAssessed++
            default: out.Compliance.ScoredOther++
            }
        }
    }
    out.Summary.Total = len(out.Results)
    out.CoverageMetrics.ScoredTotal = out.Compliance.ScoredTotal
    out.CoverageMetrics.ScoredAssessed = out.Compliance.ScoredPass + out.Compliance.ScoredFail
    if out.CoverageMetrics.Total > 0 {
        out.CoverageMetrics.OverallPercent = math.Round(float64(out.CoverageMetrics.Assessed)*10000/float64(out.CoverageMetrics.Total))/100
    }
    if out.CoverageMetrics.ScoredTotal > 0 {
        out.CoverageMetrics.ScoredPercent = math.Round(float64(out.CoverageMetrics.ScoredAssessed)*10000/float64(out.CoverageMetrics.ScoredTotal))/100
    }
    out.Compliance.CoveragePercent = out.CoverageMetrics.ScoredPercent
    // In particular, these three actual CIS findings do not imply that the
    // other official benchmark controls have been tested.
    out.Compliance.CompliancePercent = nil
    out.Compliance.ScoreStatus = "withheld: incomplete official CIS EKS 2.1.0 control coverage"
    out.Sources = append(out.Sources, SourceRecord{AuditMode: "aws-provider", GeneratedAt: evidence.GeneratedAt, Findings: len(findings)})
    sort.Slice(out.Results, func(i, j int) bool { return StableFindingID(out.Results[i]) < StableFindingID(out.Results[j]) })
    // The caller has already validated freshness and identity through
    // DecodeEKSProviderEvidence. A zero timestamp must never reach output.
    if evidence.GeneratedAt.IsZero() || evidence.GeneratedAt.After(time.Now().Add(2*time.Minute)) {
        return fmt.Errorf("missing or future AWS evidence timestamp")
    }
    return nil
}

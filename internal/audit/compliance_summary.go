package audit

import (
	"fmt"
	"math"
	"strings"
)

// ComplianceSummary deliberately withholds a compliance percentage when any
// scored check is WARN, MANUAL, INFO, NOT_ASSESSED or missing. Such checks are
// not silently converted into passes or failures.
type ComplianceSummary struct {
	ScoredTotal       int      `json:"scoredTotal"`
	ScoredPass        int      `json:"scoredPass"`
	ScoredFail        int      `json:"scoredFail"`
	ScoredWarn        int      `json:"scoredWarn"`
	ScoredManual      int      `json:"scoredManual"`
	ScoredNotAssessed int      `json:"scoredNotAssessed"`
	ScoredOther       int      `json:"scoredOther"`
	Extensions        int      `json:"extensions"`
	CoveragePercent   float64  `json:"coveragePercent"`
	CompliancePercent *float64 `json:"compliancePercent,omitempty"`
	ScoreStatus       string   `json:"scoreStatus"`
}

func buildComplianceSummary(checks []Check, results []Result) (ComplianceSummary, Result) {
	byID := make(map[string]Result, len(results))
	for _, result := range results {
		byID[result.ID] = result
	}
	var summary ComplianceSummary
	r := Result{ID: "KUBECIS-2.9.1", Title: "Summarize benchmark coverage and compliance scoring", Target: "report", Status: "INFO", Method: "report_compliance_summary", Scored: false}
	for _, check := range checks {
		if !check.Scored {
			if strings.HasPrefix(check.ID, "KUBECIS-") && !strings.HasPrefix(check.Kind, "report_") {
				summary.Extensions++
			}
			continue
		}
		summary.ScoredTotal++
		result, ok := byID[check.ID]
		status := "NOT_ASSESSED"
		if ok {
			status = strings.ToUpper(result.Status)
		}
		switch status {
		case "PASS":
			summary.ScoredPass++
		case "FAIL":
			summary.ScoredFail++
		case "WARN":
			summary.ScoredWarn++
		case "MANUAL":
			summary.ScoredManual++
		case "NOT_ASSESSED":
			summary.ScoredNotAssessed++
		default:
			summary.ScoredOther++
		}
		observed := status
		if !ok {
			observed = "missing result; treated as NOT_ASSESSED"
		}
		if len(r.Evidence) < 250 {
			r.Evidence = append(r.Evidence, Evidence{Resource: "BenchmarkCheck", Name: check.ID, Field: "status", Observed: observed, Expected: "explicit assessment result for every scored benchmark control"})
		}
	}
	if summary.ScoredTotal > 0 {
		summary.CoveragePercent = math.Round(float64(summary.ScoredPass+summary.ScoredFail)/float64(summary.ScoredTotal)*10000) / 100
	}
	unresolved := summary.ScoredWarn + summary.ScoredManual + summary.ScoredNotAssessed + summary.ScoredOther
	if summary.ScoredTotal == 0 {
		summary.ScoreStatus = "unavailable: no scored controls in selected scope"
	} else if unresolved > 0 {
		summary.ScoreStatus = "withheld: scored controls include WARN, MANUAL, NOT_ASSESSED, or unsupported statuses"
	} else {
		percentage := math.Round(float64(summary.ScoredPass)/float64(summary.ScoredTotal)*10000) / 100
		summary.CompliancePercent = &percentage
		summary.ScoreStatus = "complete: all scored controls resolved to PASS or FAIL"
	}
	r.Details = fmt.Sprintf("Scored controls=%d; PASS=%d; FAIL=%d; WARN=%d; MANUAL=%d; NOT_ASSESSED=%d; other=%d; KubeCIS extensions=%d. Assessment coverage=%0.2f%% (PASS+FAIL divided by scored controls). Compliance percentage: %s. WARN, MANUAL, NOT_ASSESSED and missing results are never counted as PASS; KubeCIS extensions are excluded from official scoring.", summary.ScoredTotal, summary.ScoredPass, summary.ScoredFail, summary.ScoredWarn, summary.ScoredManual, summary.ScoredNotAssessed, summary.ScoredOther, summary.Extensions, summary.CoveragePercent, summary.ScoreStatus)
	return summary, r
}

package audit

import "math"

// Recalculate after AWS provider evidence and the possible CNI reconciliation.
// This prevents coverage and section totals from disagreeing with findings.
func recalculateEKSConsolidatedMetrics(out *ConsolidatedReport) {
	var total Summary
	var compliance ComplianceSummary
	compliance.Extensions = out.Compliance.Extensions
	for _, result := range out.Results {
		incrementSummary(&total, result.Status)
		if result.Scored {
			compliance.ScoredTotal++
			switch result.Status {
			case "PASS":
				compliance.ScoredPass++
			case "FAIL":
				compliance.ScoredFail++
			case "WARN":
				compliance.ScoredWarn++
			case "MANUAL":
				compliance.ScoredManual++
			case "NOT_ASSESSED":
				compliance.ScoredNotAssessed++
			default:
				compliance.ScoredOther++
			}
		}
	}
	total.Total = len(out.Results)
	out.Summary = total
	out.CoverageMetrics.Total = total.Total
	out.CoverageMetrics.Assessed = total.Pass + total.Fail
	out.CoverageMetrics.ScoredTotal = compliance.ScoredTotal
	out.CoverageMetrics.ScoredAssessed = compliance.ScoredPass + compliance.ScoredFail
	if total.Total > 0 {
		out.CoverageMetrics.OverallPercent = math.Round(float64(out.CoverageMetrics.Assessed)*10000/float64(total.Total)) / 100
	}
	if compliance.ScoredTotal > 0 {
		out.CoverageMetrics.ScoredPercent = math.Round(float64(out.CoverageMetrics.ScoredAssessed)*10000/float64(compliance.ScoredTotal)) / 100
	}
	compliance.CoveragePercent = out.CoverageMetrics.ScoredPercent
	compliance.ScoreStatus = "withheld: incomplete official CIS EKS 2.1.0 control coverage"
	compliance.CompliancePercent = nil
	out.Compliance = compliance
}

package audit

import (
	"fmt"
	"strings"
)

var workloadSecurityCheckIDs = []string{
	"KUBECIS-2.4.1",
	"KUBECIS-2.4.2",
	"KUBECIS-2.4.3",
	"KUBECIS-2.4.4",
	"KUBECIS-2.4.5",
	"KUBECIS-2.4.6",
}

// buildWorkloadSecuritySummary consolidates the individual workload checks
// already present in the report. It does not issue additional Kubernetes API
// requests or reinterpret their field-level evidence.
func buildWorkloadSecuritySummary(check Check, results []Result) Result {
	r := Result{
		ID: check.ID, Title: check.Title, Target: check.Target,
		Status: "NOT_ASSESSED", Method: check.Kind,
		Remediation: check.Remediation, Scored: check.Scored,
	}
	byID := make(map[string]Result, len(results))
	for _, result := range results {
		byID[result.ID] = result
	}

	counts := map[string]int{
		"PASS": 0, "FAIL": 0, "WARN": 0, "INFO": 0,
		"MANUAL": 0, "NOT_ASSESSED": 0,
	}
	missing := 0
	for _, id := range workloadSecurityCheckIDs {
		result, ok := byID[id]
		if !ok {
			missing++
			r.Evidence = append(r.Evidence, Evidence{
				Resource: "KubeCIS workload check", Name: id,
				Field: "status", Observed: "missing from this report",
				Expected: "individual workload assessment result",
			})
			continue
		}
		status := strings.ToUpper(result.Status)
		if _, known := counts[status]; !known {
			status = "NOT_ASSESSED"
		}
		counts[status]++
		r.Evidence = append(r.Evidence, Evidence{
			Resource: "KubeCIS workload check", Name: id,
			Field: "status", Observed: status,
			Expected: "review individual check evidence and remediation",
		})
	}

	switch {
	case counts["FAIL"] > 0:
		r.Status = "FAIL"
	case counts["WARN"] > 0:
		r.Status = "WARN"
	case missing > 0 || counts["NOT_ASSESSED"] > 0 || counts["MANUAL"] > 0:
		r.Status = "NOT_ASSESSED"
	case counts["INFO"] > 0:
		r.Status = "INFO"
	case counts["PASS"] == len(workloadSecurityCheckIDs):
		r.Status = "PASS"
	default:
		r.Status = "NOT_ASSESSED"
	}

	r.Details = fmt.Sprintf(
		"Consolidated workload checks 2.4.1-2.4.6 without re-scanning the cluster: PASS=%d, FAIL=%d, WARN=%d, INFO=%d, MANUAL=%d, NOT_ASSESSED=%d, missing=%d. Overall precedence is FAIL, WARN, NOT_ASSESSED, INFO, then PASS. Review each underlying check for detailed evidence; this summary does not establish runtime state.",
		counts["PASS"], counts["FAIL"], counts["WARN"], counts["INFO"],
		counts["MANUAL"], counts["NOT_ASSESSED"], missing,
	)
	return r
}

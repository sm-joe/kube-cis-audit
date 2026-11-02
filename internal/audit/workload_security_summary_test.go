package audit

import (
	"strings"
	"testing"
)

func completeWorkloadResults(statuses ...string) []Result {
	results := make([]Result, 0, len(workloadSecurityCheckIDs))
	for i, id := range workloadSecurityCheckIDs {
		status := "PASS"
		if i < len(statuses) {
			status = statuses[i]
		}
		results = append(results, Result{ID: id, Status: status})
	}
	return results
}

func TestBuildWorkloadSecuritySummaryPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		statuses []string
		want     string
	}{
		{"all pass", nil, "PASS"},
		{"failure dominates", []string{"PASS", "FAIL", "WARN", "PASS", "PASS", "PASS"}, "FAIL"},
		{"warning dominates unknown", []string{"PASS", "WARN", "NOT_ASSESSED", "PASS", "PASS", "PASS"}, "WARN"},
		{"unknown prevents pass", []string{"PASS", "PASS", "NOT_ASSESSED", "PASS", "PASS", "PASS"}, "NOT_ASSESSED"},
		{"manual prevents pass", []string{"PASS", "PASS", "MANUAL", "PASS", "PASS", "PASS"}, "NOT_ASSESSED"},
		{"info is preserved", []string{"PASS", "PASS", "INFO", "PASS", "PASS", "PASS"}, "INFO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := Check{ID: "KUBECIS-2.4.7", Title: "Workload summary", Target: "policies", Kind: "report_workload_security_summary"}
			got := buildWorkloadSecuritySummary(check, completeWorkloadResults(tt.statuses...))
			if got.Status != tt.want {
				t.Fatalf("status=%s want=%s details=%s", got.Status, tt.want, got.Details)
			}
			if len(got.Evidence) != len(workloadSecurityCheckIDs) {
				t.Fatalf("evidence=%d want=%d", len(got.Evidence), len(workloadSecurityCheckIDs))
			}
		})
	}
}

func TestBuildWorkloadSecuritySummaryMissingCheckPreventsPass(t *testing.T) {
	check := Check{ID: "KUBECIS-2.4.7", Kind: "report_workload_security_summary"}
	results := completeWorkloadResults()
	results = results[:len(results)-1]
	got := buildWorkloadSecuritySummary(check, results)
	if got.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s want=NOT_ASSESSED", got.Status)
	}
	if !strings.Contains(got.Details, "missing=1") {
		t.Fatalf("missing check not described: %s", got.Details)
	}
}

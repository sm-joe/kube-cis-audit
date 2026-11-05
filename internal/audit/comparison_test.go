package audit

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStableFindingIDIgnoresObservedValueAndStatus(t *testing.T) {
	base := Result{ID: "5.2.7", Target: "policies", Status: "FAIL", Evidence: []Evidence{{Resource: "Pod", Namespace: "payments", Name: "api", Field: "spec.securityContext.runAsNonRoot", Observed: "false"}}}
	changed := base
	changed.Status = "PASS"
	changed.Evidence = []Evidence{{Resource: "Pod", Namespace: "payments", Name: "api", Field: "spec.securityContext.runAsNonRoot", Observed: "true"}}
	if StableFindingID(base) != StableFindingID(changed) {
		t.Fatal("finding identity changed when only status/observed value changed")
	}
	changed.Evidence[0].Name = "worker"
	if StableFindingID(base) != StableFindingID(changed) {
		t.Fatal("mutable evidence resource name must not change identity")
	}
	changed.Scope = "worker-2"
	if StableFindingID(base) == StableFindingID(changed) {
		t.Fatal("explicit stable scope must distinguish independent node findings")
	}
}

func TestCompareReportsClassifiesFindingChanges(t *testing.T) {
	old := Report{GeneratedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Results: []Result{
		{ID: "A", Target: "node", Status: "FAIL"},
		{ID: "B", Target: "node", Status: "FAIL"},
		{ID: "C", Target: "node", Status: "PASS"},
	}}
	old.Results[2].Evidence = []Evidence{{Resource: "Node", Name: "worker-1", Field: "kubelet", Observed: "old"}}
	current := Report{GeneratedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Results: []Result{
		{ID: "A", Target: "node", Status: "PASS"},
		{ID: "C", Target: "node", Status: "PASS", Evidence: []Evidence{{Resource: "Node", Name: "worker-1", Field: "kubelet", Observed: "new"}}},
		{ID: "D", Target: "node", Status: "FAIL"},
	}}
	got := CompareReports(old, current)
	if got.Summary.StatusChanged != 2 || got.Summary.Resolved != 0 || got.Summary.New != 1 || got.Summary.Unchanged != 0 || got.Summary.EvidenceChanged != 1 || got.Summary.Total != 4 {
		t.Fatalf("unexpected comparison summary: %+v", got.Summary)
	}
}

func TestDecodeReportRejectsMalformedResult(t *testing.T) {
	if _, err := DecodeReport([]byte(`{"results":[{"id":"","status":"FAIL"}]}`)); err == nil {
		t.Fatal("expected malformed report to be rejected")
	}
}

func TestCompatibilityGateWithholdsOfficialScore(t *testing.T) {
	pct := 100.0
	r := Report{CompatibilityStatus: "unknown", Compliance: ComplianceSummary{CompliancePercent: &pct}, Results: []Result{{Method: "report_compliance_summary", Details: "summary"}}}
	r.enforceNativeCompatibility()
	if r.Compliance.CompliancePercent != nil {
		t.Fatal("score must be withheld when compatibility is unknown")
	}
	if r.Compliance.ScoreStatus == "" {
		t.Fatal("withheld reason must be explicit")
	}
}

func TestComparisonJSONContract(t *testing.T) {
	got := CompareReports(Report{}, Report{})
	if _, err := json.Marshal(got); err != nil {
		t.Fatalf("comparison must serialize: %v", err)
	}
}

func TestNotAssessedCannotImplyRemediation(t *testing.T) {
	old := Report{Results: []Result{{ID: "A", Target: "policies", Status: "FAIL", FindingID: "legacy-id", Evidence: []Evidence{{Resource: "Pod", Namespace: "prod", Name: "api", Field: "hostNetwork", Observed: "true"}}}}}
	current := Report{Results: []Result{{ID: "A", Target: "policies", Status: "NOT_ASSESSED", Evidence: nil}}}
	got := CompareReports(old, current)
	if got.Summary.Resolved != 0 || got.Summary.StatusChanged != 1 || got.Changes[0].Previous != "FAIL" || got.Changes[0].Current != "NOT_ASSESSED" {
		t.Fatalf("not assessed incorrectly classified: %+v", got)
	}
	if StableFindingID(old.Results[0]) != StableFindingID(current.Results[0]) {
		t.Fatal("evidence disappearance changed stable ID")
	}
}
func TestMissingFindingNotResolved(t *testing.T) {
	got := CompareReports(Report{Results: []Result{{ID: "A", Target: "node", Status: "FAIL"}}}, Report{})
	if got.Summary.Resolved != 0 || got.Summary.StatusChanged != 1 || got.Changes[0].Current != "NOT_ASSESSED" {
		t.Fatalf("missing current evidence implied remediation: %+v", got)
	}
}
func TestDecodeRejectsDuplicateLogicalFinding(t *testing.T) {
	_, err := DecodeReport([]byte(`{"results":[{"id":"A","target":"node","status":"FAIL"},{"id":"A","target":"node","status":"PASS"}]}`))
	if err == nil {
		t.Fatal("expected duplicate rejection")
	}
}

package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEKSAWSManualEvidenceNeverAutoApprovesFindings(t *testing.T) {
	out := &ConsolidatedReport{Report: Report{
		Results: []Result{
			{ID: "5.1.2", Target: "policies", Status: "MANUAL", Details: "Reviewer pending."},
			{ID: "5.1.3", Target: "policies", Status: "MANUAL"},
			{ID: "5.3.1", Target: "policies", Status: "MANUAL"},
			{ID: "5.4.3", Target: "policies", Status: "MANUAL"},
			{ID: "5.4.2", Target: "policies", Status: "FAIL"},
		},
	}}
	evidence := EKSProviderEvidence{
		ClusterARN: "arn:aws:eks:us-east-1:123456789012:cluster/demo-eks",
		AWSManualEvidence: map[string]EKSAWSManualEvidence{
			"5.1.2": {Status: "COLLECTED", Note: "Requires IAM reviewer", Items: []json.RawMessage{json.RawMessage(`{"repository":"kubecis"}`)}},
			"5.1.3": {Status: "PARTIAL", Note: "Policy details unavailable"},
			"5.3.1": {Status: "COLLECTED", Note: "Key grants require review"},
			"5.4.3": {Status: "PARTIAL", Note: "Verify routes"},
		},
	}
	ApplyEKSAWSManualEvidence(out, evidence)
	for _, r := range out.Results[:4] {
		if r.Status != "MANUAL" || len(r.Evidence) != 1 || !strings.Contains(r.Details, "AWS manual evidence:") {
			t.Fatalf("manual control %s was misclassified or lost provenance: %+v", r.ID, r)
		}
	}
	if r := out.Results[4]; r.Status != "FAIL" || len(r.Evidence) != 0 {
		t.Fatalf("provider review must not rewrite unrelated automated FAIL: %+v", r)
	}
}

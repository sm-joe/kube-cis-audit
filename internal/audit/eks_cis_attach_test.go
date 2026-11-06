package audit

import (
	"testing"
	"time"
)

func TestAttachEKSCISProviderEvidenceTracksScoreAndProvenance(t *testing.T) {
	now := time.Now().UTC()
	report := ConsolidatedReport{
		Report: Report{
			CloudProvider: "eks", AuditMode: "consolidated",
			CompatibilityStatus: "provisional",
			Compliance:          ComplianceSummary{ScoredTotal: 1, ScoredFail: 1, ScoreStatus: "withheld"},
			Results:             []Result{{ID: "4.1.5", Target: "policies", Status: "FAIL", Scored: true}},
			Summary:             Summary{Fail: 1, Total: 1},
		},
		CoverageMetrics: CoverageMetrics{Total: 1, ScoredTotal: 1, Assessed: 1, ScoredAssessed: 1},
	}
	evidence := EKSProviderEvidence{
		SchemaVersion: 1, CloudProvider: "eks", AssessmentStatus: "INVENTORY_ONLY",
		ClusterARN: "arn:aws:eks:us-east-1:123456789012:cluster/test", GeneratedAt: now,
		EndpointPrivate: true, EndpointPublic: true, PublicAccessCIDRs: []string{"0.0.0.0/0"},
		ControlPlaneLogging: map[string]bool{
			"api": false, "audit": false, "authenticator": false, "controllerManager": false, "scheduler": false,
		},
	}
	if err := AttachEKSCISProviderEvidence(&report, evidence); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Total != 4 || report.Summary.Fail != 4 || report.Compliance.ScoredTotal != 4 ||
		report.Compliance.ScoredFail != 4 || len(report.Sources) != 1 ||
		report.Sources[0].AuditMode != "aws-provider" || report.Sources[0].Findings != 3 {
		t.Fatalf("wrong CIS EKS AWS aggregation: %+v", report)
	}
	if report.Compliance.CompliancePercent != nil || report.CompatibilityStatus != "provisional" {
		t.Fatal("partial CIS EKS coverage incorrectly upgraded to full compliance")
	}
	if err := AttachEKSCISProviderEvidence(&report, evidence); err == nil {
		t.Fatal("accepted duplicate AWS EKS source")
	}
}

func TestAttachEKSCISProviderEvidenceRefusesInvalidReport(t *testing.T) {
	r := ConsolidatedReport{Report: Report{CloudProvider: "unknown", AuditMode: "consolidated"}}
	if err := AttachEKSCISProviderEvidence(&r, EKSProviderEvidence{}); err == nil {
		t.Fatal("accepted non-EKS input")
	}
	if len(r.Results) != 0 || r.ProviderEvidence != nil {
		t.Fatal("invalid report mutated")
	}
}

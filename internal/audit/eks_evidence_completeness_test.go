package audit

import (
	"strings"
	"testing"
)

func TestEKSECRScanEvidenceRequiresVerifiedDigestAndCompleteInventory(t *testing.T) {
	good := EKSECRImageScanEvidence{
		Registry:   "123456789012.dkr.ecr.us-east-1.amazonaws.com",
		Repository: "nested/app",
		Digest:     "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ScanType:   "ENHANCED",
		ScanStatus: "ACTIVE",
	}
	e := EKSProviderEvidence{ECRInventoryStatus: "COLLECTED", ECRImages: []EKSECRImageScanEvidence{good}}
	result := EvaluateEKSAdditionalAWSControls(e)
	if len(result) != 1 || result[0].ID != "5.1.1" || result[0].Status != "PASS" {
		t.Fatalf("confirmed ECR digest should be independently assessable: %+v", result)
	}
	e.NonECRImageCount = 1
	result = EvaluateEKSAdditionalAWSControls(e)
	if result[0].Status != "MANUAL" {
		t.Fatalf("unverified third-party images must not pass: %+v", result)
	}
	e.NonECRImageCount = 0
	e.ECRImages[0].Error = "access denied"
	result = EvaluateEKSAdditionalAWSControls(e)
	if result[0].Status != "MANUAL" {
		t.Fatalf("missing ECR permissions must not pass: %+v", result)
	}
	e.ECRImages[0].Error = "CROSS_ACCOUNT_ECR_AWS_READ_ACCESS_DENIED"
	result = EvaluateEKSAdditionalAWSControls(e)
	if result[0].Status != "MANUAL" || !strings.Contains(result[0].Details, "CROSS_ACCOUNT_ECR_AWS_READ_ACCESS_DENIED=1") {
		t.Fatalf("cross-account ECR diagnostic must remain reviewable, not PASS: %+v", result)
	}
	e.ECRInventoryStatus = "UNAVAILABLE"
	result = EvaluateEKSAdditionalAWSControls(e)
	if result[0].Status != "NOT_ASSESSED" {
		t.Fatalf("incomplete Pod inventory must remain unknown: %+v", result)
	}
}

func TestEKSConfigzMissingFieldPreservesTrustedHostEvidence(t *testing.T) {
	node := "worker-a"
	report := &Report{
		AuditMode: "node-local", CloudProvider: "eks", Benchmark: "cis-eks-2.1.0", NodeName: node,
		Results: []Result{
			{ID: "3.2.5", Scope: node, Target: "node", Method: "host_kubelet_config_field", Status: "PASS", Scored: true, EvidenceSource: "node-local-metadata", Evidence: []Evidence{{Field: "makeIPTablesUtilChains", Observed: "true"}}},
			{ID: "3.2.6", Scope: node, Target: "node", Method: "host_kubelet_config_field", Status: "MANUAL", Scored: true, EvidenceSource: "node-local-metadata", Evidence: []Evidence{{Field: "eventRecordQPS", Observed: "50"}}},
			{ID: "3.2.7", Scope: node, Target: "node", Method: "host_kubelet_config_field", Status: "PASS", Scored: true, EvidenceSource: "node-local-metadata", Evidence: []Evidence{{Field: "rotateCertificates", Observed: "true", Resource: "host-file", Name: "etc/kubernetes/kubelet/config.json"}}},
		},
	}
	if err := EnrichNodeWithConfigz(report, node, []byte(`{"kubeletconfig":{"makeIPTablesUtilChains":true,"eventRecordQPS":50}}`)); err != nil {
		t.Fatal(err)
	}
	r := report.Results[2]
	if r.Status != "PASS" || r.EvidenceSource != "node-local-metadata" || len(r.Evidence) != 1 || !strings.Contains(r.Details, "preserving explicit") {
		t.Fatalf("missing /configz field erased explicit host setting: %+v", r)
	}
	if report.Compliance.ScoredPass != 2 || report.Compliance.ScoredNotAssessed != 0 || report.Compliance.CompliancePercent != nil {
		t.Fatalf("counters/official score incorrect after preserved host evidence: %+v", report.Compliance)
	}
	unknown := *report
	unknown.Results = append([]Result(nil), report.Results...)
	unknown.Results[2] = Result{ID: "3.2.7", Scope: node, Target: "node", Method: "host_kubelet_config_field", Status: "NOT_ASSESSED", Scored: true}
	if err := EnrichNodeWithConfigz(&unknown, node, []byte(`{"kubeletconfig":{"makeIPTablesUtilChains":true,"eventRecordQPS":50}}`)); err != nil {
		t.Fatal(err)
	}
	if unknown.Results[2].Status != "NOT_ASSESSED" {
		t.Fatalf("missing both host/configz evidence must stay unknown: %+v", unknown.Results[2])
	}
}

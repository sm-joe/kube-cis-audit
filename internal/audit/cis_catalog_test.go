package audit

import (
	"strings"
	"testing"
)

func TestOfficialCIS210CompleteInventory(t *testing.T) {
	expected := map[string]struct{ controls, automated, manual int }{
		"eks":        {47, 29, 18},
		"kubernetes": {134, 65, 69},
	}
	for flavor, want := range expected {
		inv, err := LoadCIS210ControlInventory(flavor)
		if err != nil {
			t.Fatalf("%s inventory invalid: %v", flavor, err)
		}
		automatic, manual := 0, 0
		for _, c := range inv.Controls {
			if c.Assessment == "automated" {
				automatic++
			} else {
				manual++
			}
		}
		if len(inv.Controls) != want.controls || automatic != want.automated || manual != want.manual {
			t.Fatalf("%s inventory mismatch: got %d/%d/%d, want %+v", flavor, len(inv.Controls), automatic, manual, want)
		}
		if err := ValidateCIS210FullCoverage(inv, nil, nil); err == nil ||
			!strings.Contains(err.Error(), "full release blocked") {
			t.Fatalf("%s partial benchmark falsely certified: %v", flavor, err)
		}
	}
	if _, err := LoadCIS210ControlInventory("unknown"); err == nil {
		t.Fatal("accepted unverified flavor")
	}
}

func TestFullCoverageGateValidatesEvidenceAndHumanReview(t *testing.T) {
	inv := CISControlInventory{Benchmark: "cis-eks-test", Controls: []CISControlIdentity{
		{ID: "2.1.1", Assessment: "automated", Implemented: true, PDFPage: 19},
		{ID: "4.1.1", Assessment: "manual", Implemented: true, PDFPage: 80},
	}}
	auto := Result{ID: "2.1.1", Status: "PASS", Scored: true, Evidence: []Evidence{{Resource: "AWS::EKS::Cluster", Observed: "enabled"}}}
	manual := Result{ID: "4.1.1", Status: "MANUAL", Scored: false, Details: "Manual review required"}
	if err := ValidateCIS210FullCoverage(inv, []Result{auto, manual}, nil); err == nil ||
		!strings.Contains(err.Error(), "manual_review_pending") {
		t.Fatalf("manual item silently passed: %v", err)
	}
	if err := ValidateCIS210FullCoverage(inv, []Result{auto, manual}, map[string]bool{"4.1.1": true}); err != nil {
		t.Fatalf("documented manual review rejected: %v", err)
	}
	auto.Evidence = nil
	if err := ValidateCIS210FullCoverage(inv, []Result{auto, manual}, map[string]bool{"4.1.1": true}); err == nil {
		t.Fatal("unsupported PASS accepted without evidence")
	}
}

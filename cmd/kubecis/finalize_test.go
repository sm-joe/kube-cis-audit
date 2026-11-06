package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
)

func TestFinalAssessmentImmutableResultsAndHumanSeparation(t *testing.T) {
	raw := []byte("exact original report bytes")
	sum := sha256.Sum256(raw)
	now := time.Date(2026, 11, 6, 13, 0, 0, 0, time.UTC)
	source := audit.Report{
		Benchmark: "cis-eks-2.1.0", CompatibilityStatus: "provisional",
		Results: []audit.Result{
			{ID: "2.1.1", Status: "FAIL", Title: "Control plane logging", Scored: true},
			{ID: "3.2.7", Status: "NOT_ASSESSED", Title: "Certificate rotation", Scored: true},
			{ID: "4.1.1", Status: "MANUAL", Title: "Admin privileges", Evidence: []audit.Evidence{{Resource: "ClusterRoleBinding", Name: "cluster-admin", Observed: "binding"}}, Scored: false},
			{ID: "4.1.2", Status: "MANUAL", Title: "Secrets read", Scored: false},
			{ID: "3.1.3", Status: "PASS", Title: "Permissions", Scored: true},
		},
	}
	input := reviewInput{ReportSHA256: hex.EncodeToString(sum[:]), Decisions: []reviewDecision{
		{ID: "4.1.1", Decision: "APPROVED", Reviewer: "security", EvidenceRef: "ticket-24", Justification: "Reviewed binding scope", ReviewedAt: now},
	}}
	final, err := assembleFinalAssessment(raw, source, input, now)
	if err != nil {
		t.Fatal(err)
	}
	s := final.Summary
	if s.AutomatedPass != 1 || s.AutomatedFail != 1 || s.NotAssessed != 1 || s.ManualTotal != 2 || s.AwaitingReview != 1 || s.ReviewedApproved != 1 || s.Total != 5 {
		t.Fatalf("incorrect final summary %+v", s)
	}
	if final.Source.Results[0].Status != "FAIL" || final.Source.Results[1].Status != "NOT_ASSESSED" {
		t.Fatal("review mutated the immutable original scan")
	}
	seen := false
	for _, c := range final.Controls {
		if c.ID == "4.1.1" {
			seen = true
			if c.AutomatedStatus != "MANUAL" || c.ReviewStatus != "APPROVED" || c.Decision == nil || len(c.Evidence) != 1 {
				t.Fatalf("lost manual provenance %+v", c)
			}
		}
		if c.ID == "2.1.1" && c.ReviewStatus != "NOT_REQUIRED" {
			t.Fatalf("automated FAIL changed review status: %+v", c)
		}
	}
	if !seen {
		t.Fatal("manual control missing")
	}
	txt := renderFinalAssessment(final)
	for _, fragment := range []string{"[FAIL] 2.1.1", "[NOT_ASSESSED] 3.2.7", "Human review: APPROVED", "Pending: 1", "Compliance score: WITHHELD"} {
		if !strings.Contains(txt, fragment) {
			t.Errorf("missing %q in final text", fragment)
		}
	}
}

func TestFinalAssessmentRejectsTamperingAndInventedApprovals(t *testing.T) {
	now := time.Now().UTC()
	raw := []byte("source report")
	sum := sha256.Sum256(raw)
	report := audit.Report{Results: []audit.Result{{ID: "3.2.7", Status: "NOT_ASSESSED"}, {ID: "4.1.1", Status: "MANUAL"}}}
	input := reviewInput{ReportSHA256: hex.EncodeToString(sum[:]), Decisions: []reviewDecision{{
		ID: "3.2.7", Decision: "APPROVED", Reviewer: "test", EvidenceRef: "ticket", Justification: "not allowed", ReviewedAt: now,
	}}}
	if _, err := assembleFinalAssessment(raw, report, input, now); err == nil {
		t.Fatal("approved NOT_ASSESSED implementation gap")
	}
	input.Decisions = nil
	input.ReportSHA256 = strings.Repeat("a", 64)
	if _, err := assembleFinalAssessment(raw, report, input, now); err == nil {
		t.Fatal("accepted changed report hash")
	}
	input.ReportSHA256 = hex.EncodeToString(sum[:])
	f, err := assembleFinalAssessment(raw, report, input, now)
	if err != nil {
		t.Fatal(err)
	}
	if f.Summary.AwaitingReview != 1 || f.Review.ReviewStatus != "INCOMPLETE" {
		t.Fatal("silent approval of unresolved review")
	}
}

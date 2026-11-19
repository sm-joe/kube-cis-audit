package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
)

func TestManualReviewAttestationRequiresBoundEvidence(t *testing.T) {
	raw := []byte("immutable source bytes")
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	report := audit.Report{
		Benchmark: "cis-eks-2.1.0",
		Results: []audit.Result{
			{ID: "4.1.1", Status: "MANUAL"},
			{ID: "3.1.1", Status: "NOT_ASSESSED"},
		},
	}
	entry := reviewDecision{
		ID: "4.1.1", Decision: "APPROVED", Reviewer: "security-reviewer",
		EvidenceRef: "ticket-42", Justification: "Least privilege reviewed",
		ReviewedAt: now,
	}
	approved, err := validateReviews(raw, report, reviewInput{
		ReportSHA256: hash, Decisions: []reviewDecision{entry},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if approved.ReviewStatus != "REVIEWED" || len(approved.Pending) != 0 {
		t.Fatalf("unexpected review result: %+v", approved)
	}
	if !strings.Contains(approved.ScoreStatus, "WITHHELD") {
		t.Fatal("manual review must not release a CIS score")
	}
	_, err = validateReviews(raw, report, reviewInput{
		ReportSHA256: strings.Repeat("0", 64), Decisions: []reviewDecision{entry},
	}, now)
	if err == nil {
		t.Fatal("tampered source report accepted")
	}
	entry.ID = "3.1.1"
	_, err = validateReviews(raw, report, reviewInput{
		ReportSHA256: hash, Decisions: []reviewDecision{entry},
	}, now)
	if err == nil {
		t.Fatal("manual review must not override NOT_ASSESSED implementation gap")
	}
}

func TestManualReviewRejectsUnattributedDecisions(t *testing.T) {
	raw := []byte("report")
	sum := sha256.Sum256(raw)
	now := time.Now().UTC()
	report := audit.Report{Results: []audit.Result{{ID: "4.1.1", Status: "MANUAL"}}}
	_, err := validateReviews(raw, report, reviewInput{
		ReportSHA256: hex.EncodeToString(sum[:]),
		Decisions: []reviewDecision{{
			ID: "4.1.1", Decision: "APPROVED", Reviewer: "",
			EvidenceRef: "test-ticket", Justification: "review", ReviewedAt: now,
		}},
	}, now)
	if err == nil {
		t.Fatal("review without identified reviewer accepted")
	}
}

func TestCIS543PublicNodeExceptionDoesNotOverrideFailure(t *testing.T) {
	raw := []byte("immutable public node scan")
	hash := sha256.Sum256(raw)
	now := time.Now().UTC()
	report := audit.Report{Benchmark: "cis-eks-2.1.0", Results: []audit.Result{
		{ID: "5.4.3", Target: "policies", Status: "FAIL", Details: "IGW route verified"},
	}}
	exception := reviewDecision{
		ID: "5.4.3", Decision: "EXCEPTION",
		Reviewer: "independent-security-owner",
		EvidenceRef: "change-control/CIS-543",
		Justification: "Approved test-cluster public-node deviation",
		ReviewedAt: now,
	}
	input := reviewInput{ReportSHA256: hex.EncodeToString(hash[:]), Decisions: []reviewDecision{exception}}
	accepted, err := validateReviews(raw, report, input, now)
	if err != nil || len(accepted.Decisions) != 1 {
		t.Fatalf("legitimate evidence-bound exception rejected: %+v %v", accepted, err)
	}
	exception.Decision = "APPROVED"
	_, err = validateReviews(raw, report, reviewInput{
		ReportSHA256: input.ReportSHA256, Decisions: []reviewDecision{exception},
	}, now)
	if err == nil {
		t.Fatal("an automated FAIL must not accept a PASS-equivalent approval")
	}
	exception.Decision = "EXCEPTION"
	exception.ID = "5.4.2"
	_, err = validateReviews(raw, report, reviewInput{
		ReportSHA256: input.ReportSHA256, Decisions: []reviewDecision{exception},
	}, now)
	if err == nil {
		t.Fatal("exception cannot apply to a different finding")
	}
}

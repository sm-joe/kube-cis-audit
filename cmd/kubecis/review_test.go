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

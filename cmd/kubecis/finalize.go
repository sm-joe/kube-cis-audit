package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
	"github.com/spf13/cobra"
)

// FinalAssessment is deliberately separate from the raw, immutable scan. A
// review is an attestation, NOT a mutation or reinterpretation of CIS results.
type finalAssessment struct {
	SchemaVersion      int                `json:"schemaVersion"`
	GeneratedAt        time.Time          `json:"generatedAt"`
	SourceReportSHA256 string             `json:"sourceReportSha256"`
	Source             audit.Report       `json:"source"`
	Review             reviewOutput       `json:"review"`
	Summary            finalReviewSummary `json:"summary"`
	Controls           []finalControl     `json:"controls"`
	ScoreStatus        string             `json:"scoreStatus"`
}

type finalReviewSummary struct {
	Total             int `json:"total"`
	AutomatedPass     int `json:"automatedPass"`
	AutomatedFail     int `json:"automatedFail"`
	AutomatedWarn     int `json:"automatedWarn"`
	NotAssessed       int `json:"notAssessed"`
	ManualTotal       int `json:"manualTotal"`
	ReviewedApproved  int `json:"reviewedApproved"`
	ReviewedRejected  int `json:"reviewedRejected"`
	ReviewedException int `json:"reviewedException"`
	AwaitingReview    int `json:"awaitingReview"`
}

type finalControl struct {
	ID              string           `json:"id"`
	Scope           string           `json:"scope,omitempty"`
	Title           string           `json:"title"`
	AutomatedStatus string           `json:"automatedStatus"`
	Scored          bool             `json:"scored"`
	Method          string           `json:"method,omitempty"`
	EvidenceSource  string           `json:"evidenceSource,omitempty"`
	Evidence        []audit.Evidence `json:"evidence,omitempty"`
	Details         string           `json:"details,omitempty"`
	Remediation     string           `json:"remediation,omitempty"`
	ReviewStatus    string           `json:"reviewStatus"`
	Decision        *reviewDecision  `json:"decision,omitempty"`
}

func assembleFinalAssessment(raw []byte, report audit.Report, input reviewInput, now time.Time) (finalAssessment, error) {
	review, err := validateReviews(raw, report, input, now)
	if err != nil {
		return finalAssessment{}, err
	}
	sum := sha256.Sum256(raw)
	out := finalAssessment{
		SchemaVersion: 1, GeneratedAt: now,
		SourceReportSHA256: hex.EncodeToString(sum[:]),
		Source:             report, Review: review,
		Controls:    make([]finalControl, 0, len(report.Results)),
		ScoreStatus: "WITHHELD: CIS benchmark is provisional; manual decisions do not override scan findings or certify compliance",
	}
	decisions := make(map[string]reviewDecision, len(input.Decisions))
	for _, d := range input.Decisions {
		decisions[d.ID+"|"+d.Scope] = d
	}
	for _, r := range report.Results {
		c := finalControl{
			ID: r.ID, Scope: r.Scope, Title: r.Title, AutomatedStatus: r.Status,
			Scored: r.Scored, Method: r.Method, EvidenceSource: r.EvidenceSource,
			Evidence: r.Evidence, Details: r.Details, Remediation: r.Remediation,
			ReviewStatus: "NOT_REQUIRED",
		}
		out.Summary.Total++
		switch r.Status {
		case "PASS":
			out.Summary.AutomatedPass++
		case "FAIL":
			out.Summary.AutomatedFail++
		case "WARN":
			out.Summary.AutomatedWarn++
		case "NOT_ASSESSED":
			out.Summary.NotAssessed++
		case "MANUAL":
			out.Summary.ManualTotal++
			c.ReviewStatus = "PENDING"
			if d, ok := decisions[r.ID+"|"+r.Scope]; ok {
				c.Decision = &d
				c.ReviewStatus = d.Decision
				switch d.Decision {
				case "APPROVED":
					out.Summary.ReviewedApproved++
				case "REJECTED":
					out.Summary.ReviewedRejected++
				case "EXCEPTION":
					out.Summary.ReviewedException++
				}
			} else {
				out.Summary.AwaitingReview++
			}
		}
		out.Controls = append(out.Controls, c)
	}
	sort.Slice(out.Controls, func(i, j int) bool {
		if out.Controls[i].ID == out.Controls[j].ID {
			return out.Controls[i].Scope < out.Controls[j].Scope
		}
		return cisIDLess(out.Controls[i].ID, out.Controls[j].ID)
	})
	if out.Summary.AwaitingReview > 0 {
		out.Review.ReviewStatus = "INCOMPLETE"
	}
	return out, nil
}

func cisIDLess(a, b string) bool {
	ai, bi := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(ai) && i < len(bi); i++ {
		var x, y int
		_, xe := fmt.Sscanf(ai[i], "%d", &x)
		_, ye := fmt.Sscanf(bi[i], "%d", &y)
		if xe == nil && ye == nil && x != y {
			return x < y
		}
		if ai[i] != bi[i] {
			return ai[i] < bi[i]
		}
	}
	return len(ai) < len(bi)
}

func renderFinalAssessment(out finalAssessment) string {
	var b strings.Builder
	fmt.Fprintf(&b, "KubeCIS Native — final assessment dossier\nBuild: %s\nBenchmark: %s (%s)\nKubernetes: %s\nProvider: %s\nSource SHA256: %s\nReview: %s\nCompliance: %s\n\n", out.Source.BuildVersion, out.Source.Benchmark, out.Source.CompatibilityStatus, out.Source.ClusterVersion, out.Source.CloudProvider, out.SourceReportSHA256, out.Review.ReviewStatus, out.ScoreStatus)
	for _, c := range out.Controls {
		scope := ""
		if c.Scope != "" {
			scope = " [node: " + c.Scope + "]"
		}
		fmt.Fprintf(&b, "[%s] %s %s%s\n", c.AutomatedStatus, c.ID, c.Title, scope)
		if c.AutomatedStatus == "MANUAL" {
			fmt.Fprintf(&b, "  Human review: %s\n", c.ReviewStatus)
			if c.Decision != nil {
				fmt.Fprintf(&b, "  Attestation: %s by %s at %s\n  Evidence reference: %s\n  Justification: %s\n", c.Decision.Decision, c.Decision.Reviewer, c.Decision.ReviewedAt.Format(time.RFC3339), c.Decision.EvidenceRef, c.Decision.Justification)
			}
		}
		if c.Details != "" && c.AutomatedStatus != "PASS" {
			fmt.Fprintf(&b, "  Evidence: %s\n", strings.ReplaceAll(c.Details, "\n", " "))
		}
	}
	s := out.Summary
	fmt.Fprintf(&b, "\n== Final assessment summary ==\nAutomated PASS: %d\nAutomated FAIL: %d\nAutomated WARN: %d\nNOT_ASSESSED: %d\nManual controls: %d\nApproved: %d\nRejected: %d\nExceptions: %d\nPending: %d\nControl findings: %d\nCompliance score: WITHHELD\n", s.AutomatedPass, s.AutomatedFail, s.AutomatedWarn, s.NotAssessed, s.ManualTotal, s.ReviewedApproved, s.ReviewedRejected, s.ReviewedException, s.AwaitingReview, s.Total)
	return b.String()
}

func newFinalizeCommand() *cobra.Command {
	var input, decisions, outputDir string
	cmd := &cobra.Command{
		Use: "finalize", Short: "Create one immutable-evidence assessment dossier with separate human review decisions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(outputDir) == "" {
				return fmt.Errorf("output-dir is required")
			}
			raw, err := os.ReadFile(input)
			if err != nil {
				return fmt.Errorf("read consolidated report: %w", err)
			}
			report, err := audit.DecodeReport(raw)
			if err != nil {
				return fmt.Errorf("decode report: %w", err)
			}
			hash := sha256.Sum256(raw)
			in := reviewInput{ReportSHA256: hex.EncodeToString(hash[:]), Decisions: []reviewDecision{}}
			if decisions != "" {
				d, err := os.ReadFile(decisions)
				if err != nil {
					return fmt.Errorf("read decisions: %w", err)
				}
				if err := json.Unmarshal(d, &in); err != nil {
					return fmt.Errorf("invalid decisions JSON: %w", err)
				}
			}
			final, err := assembleFinalAssessment(raw, report, in, time.Now().UTC())
			if err != nil {
				return err
			}
			if err := os.MkdirAll(outputDir, 0700); err != nil {
				return err
			}
			dst := filepath.Join(outputDir, "final-assessment.json")
			data, err := json.MarshalIndent(final, "", "  ")
			if err != nil {
				return err
			}
			if err := writeFinalAtomic(dst, append(data, '\n')); err != nil {
				return err
			}
			txt := filepath.Join(outputDir, "final-assessment.txt")
			if err := writeFinalAtomic(txt, []byte(renderFinalAssessment(final))); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Final assessment: %s\nReadable assessment: %s\nManual pending: %d, automated failures: %d, score WITHHELD\n", dst, txt, final.Summary.AwaitingReview, final.Summary.AutomatedFail)
			return nil
		},
	}
	cmd.Flags().StringVar(&input, "input", "", "Immutable consolidated scan JSON")
	cmd.Flags().StringVar(&decisions, "decisions", "", "Optional review decisions JSON from review-init; absent = all manual controls pending")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory for final-assessment.json and final-assessment.txt")
	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("output-dir")
	return cmd
}

func writeFinalAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".kubecis-final-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

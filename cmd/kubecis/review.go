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

// Review records are kept separate from native scan results: human judgments
// must never overwrite source evidence, or complete an unimplemented control.
type reviewDecision struct {
	ID            string    `json:"id"`
	Scope         string    `json:"scope,omitempty"`
	Decision      string    `json:"decision"`
	Reviewer      string    `json:"reviewer"`
	EvidenceRef   string    `json:"evidenceRef"`
	Justification string    `json:"justification"`
	ReviewedAt    time.Time `json:"reviewedAt"`
}

type reviewInput struct {
	ReportSHA256 string           `json:"reportSha256"`
	Decisions    []reviewDecision `json:"decisions"`
}

type reviewCandidate struct {
	ID             string           `json:"id"`
	Scope          string           `json:"scope,omitempty"`
	Title          string           `json:"title"`
	Details        string           `json:"details"`
	Scored         bool             `json:"scored"`
	Method         string           `json:"method,omitempty"`
	EvidenceSource string           `json:"evidenceSource,omitempty"`
	Evidence       []audit.Evidence `json:"evidence,omitempty"`
	Remediation    string           `json:"remediation,omitempty"`
	ReviewGuidance string           `json:"reviewGuidance"`
}

type reviewTemplate struct {
	ReportSHA256   string            `json:"reportSha256"`
	Decisions      []reviewDecision  `json:"decisions"`
	PendingReviews []reviewCandidate `json:"pendingReviews"`
}

type reviewOutput struct {
	Benchmark    string           `json:"benchmark"`
	ReportSHA256 string           `json:"reportSha256"`
	GeneratedAt  time.Time        `json:"generatedAt"`
	Decisions    []reviewDecision `json:"decisions"`
	Pending      []string         `json:"pending"`
	ReviewStatus string           `json:"reviewStatus"`
	ScoreStatus  string           `json:"scoreStatus"`
}

func newReviewCommand() *cobra.Command {
	var reportPath, decisionsPath, output string
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Validate human review attestations against an immutable KubeCIS JSON report",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(reportPath)
			if err != nil {
				return fmt.Errorf("read scan report: %w", err)
			}
			report, err := audit.DecodeReport(raw)
			if err != nil {
				return fmt.Errorf("decode scan report: %w", err)
			}
			data, err := os.ReadFile(decisionsPath)
			if err != nil {
				return fmt.Errorf("read review input: %w", err)
			}
			var input reviewInput
			if err := json.Unmarshal(data, &input); err != nil {
				return fmt.Errorf("decode review input: %w", err)
			}
			attestation, err := validateReviews(raw, report, input, time.Now().UTC())
			if err != nil {
				return err
			}
			rendered, err := json.MarshalIndent(attestation, "", "  ")
			if err != nil {
				return err
			}
			if output == "" {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), string(rendered))
				return err
			}
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				return err
			}
			return os.WriteFile(output, append(rendered, '\n'), 0600)
		},
	}
	cmd.Flags().StringVar(&reportPath, "input", "", "Original KubeCIS consolidated JSON report")
	cmd.Flags().StringVar(&decisionsPath, "decisions", "", "JSON human review attestations bound to report SHA-256")
	cmd.Flags().StringVar(&output, "output-file", "", "Write validated review JSON (default stdout)")
	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("decisions")
	return cmd
}

func newReviewInitCommand() *cobra.Command {
	var input, output string
	cmd := &cobra.Command{
		Use:   "review-init",
		Short: "Generate an evidence-bound human review template for a native scan",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(input)
			if err != nil {
				return fmt.Errorf("read scan report: %w", err)
			}
			report, err := audit.DecodeReport(raw)
			if err != nil {
				return fmt.Errorf("decode scan report: %w", err)
			}
			digest := sha256.Sum256(raw)
			template := reviewTemplate{
				ReportSHA256:   hex.EncodeToString(digest[:]),
				Decisions:      []reviewDecision{},
				PendingReviews: []reviewCandidate{},
			}
			for _, r := range report.Results {
				if r.Status == "FAIL" && r.ID == "5.4.3" {
					template.PendingReviews = append(template.PendingReviews, reviewCandidate{
						ID: r.ID, Scope: r.Scope, Title: r.Title, Details: r.Details,
						Scored: r.Scored, Method: r.Method, EvidenceSource: r.EvidenceSource,
						Evidence: r.Evidence, Remediation: r.Remediation,
						ReviewGuidance: "Optional documented EXCEPTION only. The CIS FAIL remains unchanged; record reviewer, dated justification and resolvable evidenceRef.",
					})
				}
				if r.Status == "MANUAL" {
					guidance := "Inspect evidence; record APPROVED, REJECTED or EXCEPTION only after human verification. Include reviewer, dated justification and independently resolvable evidenceRef. No auto-approval."
					switch r.ID {
					case "3.2.6":
						guidance = "Assess the effective kubelet eventRecordQPS for this specific node against actual workload event throughput and throttling evidence. The default value 50 is not an automatic PASS; zero disables rate limiting and is not an automatic PASS. Record the reviewer, justification and evidence reference."
					case "3.2.7":
						guidance = "Verify the active kubelet client authentication mechanism for this node. IAM exec-token-only evidence can support a documented non-applicability judgment, not a PASS. Check whether any bootstrap or client-certificate path remains active; omitted rotateCertificates from /configz does not prove it is disabled. Record applicability reasoning and an independently resolvable evidence reference."
					case "4.1.7":
						guidance = "Review every IAM principal, EKS access-policy scope, Kubernetes group and aws-auth mapping against an independently maintained least-privilege access approval. Explicitly justify any cluster-wide administrative grant. APPROVED means the documented authorization review was completed; it does not rewrite native findings or release a CIS score."
					case "5.5.1":
						guidance = "Verify that API or API_AND_CONFIG_MAP is active, that intended IAM principals use EKS access entries, and that any remaining aws-auth mappings are required for migration or node bootstrap. API_AND_CONFIG_MAP does not prove migration completeness. Record a separately evidenced reviewer decision; do not infer approval from entry counts."
					}
					template.PendingReviews = append(template.PendingReviews, reviewCandidate{
						ID: r.ID, Scope: r.Scope, Title: r.Title, Details: r.Details,
						Scored: r.Scored, Method: r.Method, EvidenceSource: r.EvidenceSource,
						Evidence: r.Evidence, Remediation: r.Remediation,
						ReviewGuidance: guidance,
					})
				}
			}
			sort.Slice(template.PendingReviews, func(i, j int) bool {
				if template.PendingReviews[i].ID == template.PendingReviews[j].ID {
					return template.PendingReviews[i].Scope < template.PendingReviews[j].Scope
				}
				return template.PendingReviews[i].ID < template.PendingReviews[j].ID
			})
			encoded, err := json.MarshalIndent(template, "", "  ")
			if err != nil {
				return err
			}
			if output == "" {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
				return err
			}
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				return err
			}
			return os.WriteFile(output, append(encoded, '\n'), 0600)
		},
	}
	cmd.Flags().StringVar(&input, "input", "", "Original KubeCIS consolidated JSON report")
	cmd.Flags().StringVar(&output, "output-file", "", "Review template JSON path (default stdout)")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

func validateReviews(raw []byte, report audit.Report, input reviewInput, now time.Time) (reviewOutput, error) {
	sum := sha256.Sum256(raw)
	digest := hex.EncodeToString(sum[:])
	if input.ReportSHA256 != digest {
		return reviewOutput{}, fmt.Errorf("review reportSha256 mismatch: attestations must bind to exact source report bytes")
	}
	pending := map[string]bool{}
	exceptionEligible := map[string]bool{}
	for _, result := range report.Results {
		key := result.ID + "|" + result.Scope
		if result.Status == "MANUAL" {
			pending[key] = true
		}
		if result.ID == "5.4.3" && result.Status == "FAIL" {
			// An exception documents a deviation but cannot reclassify it.
			exceptionEligible[key] = true
		}
	}
	seen := map[string]bool{}
	for _, decision := range input.Decisions {
		key := decision.ID + "|" + decision.Scope
		if !pending[key] && !(exceptionEligible[key] && decision.Decision == "EXCEPTION") {
			return reviewOutput{}, fmt.Errorf("review %s: only MANUAL controls or an EXCEPTION for a verified 5.4.3 FAIL may be attested", key)
		}
		if seen[key] {
			return reviewOutput{}, fmt.Errorf("duplicate review for control %s", key)
		}
		seen[key] = true
		if decision.Decision != "APPROVED" && decision.Decision != "REJECTED" && decision.Decision != "EXCEPTION" {
			return reviewOutput{}, fmt.Errorf("review %s: decision must be APPROVED, REJECTED, or EXCEPTION", key)
		}
		if strings.TrimSpace(decision.Reviewer) == "" || strings.TrimSpace(decision.EvidenceRef) == "" || strings.TrimSpace(decision.Justification) == "" {
			return reviewOutput{}, fmt.Errorf("review %s: reviewer, evidenceRef, and justification are required", key)
		}
		if decision.ReviewedAt.IsZero() || decision.ReviewedAt.After(now.Add(time.Minute)) {
			return reviewOutput{}, fmt.Errorf("review %s: invalid review timestamp", key)
		}
		delete(pending, key)
	}
	output := reviewOutput{
		Benchmark:    report.Benchmark,
		ReportSHA256: digest,
		GeneratedAt:  now,
		Decisions:    input.Decisions,
		ReviewStatus: "INCOMPLETE",
		ScoreStatus:  "WITHHELD: native scan results are unchanged; manual decisions do not prove complete CIS assessment",
	}
	for key := range pending {
		output.Pending = append(output.Pending, key)
	}
	sort.Strings(output.Pending)
	if len(pending) == 0 {
		output.ReviewStatus = "REVIEWED"
	}
	return output, nil
}

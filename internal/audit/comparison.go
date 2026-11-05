package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// StableFindingID identifies a check finding independently of observed values.
// Observed values and status are deliberately excluded so configuration drift
// changes the finding state rather than creating a different identity.
func StableFindingID(r Result) string {
	// Resource observations belong to evidence, not identity: they can vanish
	// when a permission failure or unavailable resource yields NOT_ASSESSED.
	// Scope is explicit when a logical control is node-specific.
	parts := []string{
		strings.ToLower(strings.TrimSpace(r.ID)),
		strings.ToLower(strings.TrimSpace(r.Target)),
		strings.ToLower(strings.TrimSpace(r.Scope)),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return "kcf_" + hex.EncodeToString(sum[:16])
}

type FindingChange struct {
	FindingID string `json:"findingId"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Target    string `json:"target"`
	Change    string `json:"change"` // NEW, RESOLVED (reserved), UNCHANGED, STATUS_CHANGED, EVIDENCE_CHANGED
	Previous  string `json:"previousStatus,omitempty"`
	Current   string `json:"currentStatus,omitempty"`
}

type ScanComparison struct {
	SchemaVersion string            `json:"schemaVersion"`
	BaselineAt    string            `json:"baselineGeneratedAt"`
	CurrentAt     string            `json:"currentGeneratedAt"`
	BaselineBuild string            `json:"baselineBuildVersion,omitempty"`
	CurrentBuild  string            `json:"currentBuildVersion,omitempty"`
	Summary       ComparisonSummary `json:"summary"`
	Changes       []FindingChange   `json:"changes"`
}

type ComparisonSummary struct {
	New             int `json:"new"`
	Resolved        int `json:"resolved"`
	Unchanged       int `json:"unchanged"`
	StatusChanged   int `json:"statusChanged"`
	EvidenceChanged int `json:"evidenceChanged"`
	Total           int `json:"total"`
}

// CompareReports compares two completed reports using stable finding identity.
func CompareReports(baseline, current Report) ScanComparison {
	// Recompute identities rather than trusting older findingId fields, which
	// could incorporate mutable evidence. Legacy reports remain comparable.
	oldByID := make(map[string]Result, len(baseline.Results))
	newByID := make(map[string]Result, len(current.Results))
	for _, r := range baseline.Results {
		if isReportSummary(r) {
			continue
		}
		if r.Scope == "" && baseline.NodeName != "" {
			r.Scope = baseline.NodeName
		}
		oldByID[StableFindingID(r)] = r
	}
	for _, r := range current.Results {
		if isReportSummary(r) {
			continue
		}
		if r.Scope == "" && current.NodeName != "" {
			r.Scope = current.NodeName
		}
		newByID[StableFindingID(r)] = r
	}
	ids := make([]string, 0, len(oldByID)+len(newByID))
	seen := make(map[string]bool)
	for id := range oldByID {
		ids = append(ids, id)
		seen[id] = true
	}
	for id := range newByID {
		if !seen[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	out := ScanComparison{SchemaVersion: "1.0", BaselineAt: baseline.GeneratedAt.UTC().Format(time.RFC3339), CurrentAt: current.GeneratedAt.UTC().Format(time.RFC3339), BaselineBuild: baseline.BuildVersion, CurrentBuild: current.BuildVersion, Changes: make([]FindingChange, 0, len(ids))}
	for _, id := range ids {
		old, hadOld := oldByID[id]
		cur, hasCurrent := newByID[id]
		item := FindingChange{FindingID: id}
		switch {
		case !hadOld:
			item.ID, item.Title, item.Target, item.Change, item.Current = cur.ID, cur.Title, cur.Target, "NEW", cur.Status
			out.Summary.New++
		case !hasCurrent:
			// Absence from a scan is not proof of remediation: filters, denied
			// permissions, outages or an incomplete scan can hide a check.
			item.ID, item.Title, item.Target, item.Change, item.Previous, item.Current = old.ID, old.Title, old.Target, "STATUS_CHANGED", old.Status, "NOT_ASSESSED"
			out.Summary.StatusChanged++
		case !strings.EqualFold(old.Status, cur.Status):
			item.ID, item.Title, item.Target, item.Change, item.Previous, item.Current = cur.ID, cur.Title, cur.Target, "STATUS_CHANGED", old.Status, cur.Status
			out.Summary.StatusChanged++
		case old.Details != cur.Details || !reflect.DeepEqual(old.Evidence, cur.Evidence):
			// Remediation text is guidance, not scan evidence.
			item.ID, item.Title, item.Target, item.Change, item.Previous, item.Current = cur.ID, cur.Title, cur.Target, "EVIDENCE_CHANGED", old.Status, cur.Status
			out.Summary.EvidenceChanged++
		default:
			item.ID, item.Title, item.Target, item.Change, item.Previous, item.Current = cur.ID, cur.Title, cur.Target, "UNCHANGED", old.Status, cur.Status
			out.Summary.Unchanged++
		}
		out.Changes = append(out.Changes, item)
	}
	out.Summary.Total = len(out.Changes)
	return out
}

func isReportSummary(r Result) bool {
	return r.Method == "report_compliance_summary" || r.Method == "report_workload_security_summary"
}

// DecodeReport validates the minimum report contract before comparison.
func DecodeReport(data []byte) (Report, error) {
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return Report{}, fmt.Errorf("decode report JSON: %w", err)
	}
	if len(report.Results) == 0 {
		return Report{}, fmt.Errorf("report contains no results")
	}
	seen := make(map[string]bool)
	for i, r := range report.Results {
		if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.Status) == "" {
			return Report{}, fmt.Errorf("result %d is missing id or status", i)
		}
		if isReportSummary(r) {
			continue
		}
		if r.Scope == "" && report.NodeName != "" {
			r.Scope = report.NodeName
		}
		key := StableFindingID(r)
		if seen[key] {
			return Report{}, fmt.Errorf("duplicate logical finding identity for result %d (%s/%s); cannot compare without losing results", i, r.ID, r.Target)
		}
		seen[key] = true
	}
	return report, nil
}

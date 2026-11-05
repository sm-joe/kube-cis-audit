package audit

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// SourceRecord identifies an assessment included in a consolidated report.
type SourceRecord struct {
	AuditMode   string    `json:"auditMode"`
	NodeName    string    `json:"nodeName,omitempty"`
	GeneratedAt time.Time `json:"generatedAt"`
	Findings    int       `json:"findings"`
}

type CoverageMetrics struct {
	Assessed       int     `json:"assessed"`
	Total          int     `json:"total"`
	OverallPercent float64 `json:"overallPercent"`
	ScoredAssessed int     `json:"scoredAssessed"`
	ScoredTotal    int     `json:"scoredTotal"`
	ScoredPercent  float64 `json:"scoredPercent"`
}

// ConsolidatedReport retains the original Report JSON fields for existing consumers.
// Completeness is based on an explicit expected-node inventory, never inferred
// from whichever reports happened to be supplied.
type ConsolidatedReport struct {
	Report
	SchemaVersion             string          `json:"schemaVersion"`
	Sources                   []SourceRecord  `json:"sources"`
	ExpectedNodes             []string        `json:"expectedNodes"`
	ExpectedControlPlaneNodes []string        `json:"expectedControlPlaneNodes"`
	CoverageMetrics           CoverageMetrics `json:"coverageMetrics"`
}

// ConsolidateReports fails closed on missing, stale, inconsistent or duplicate
// source reports. It does not reinterpret NOT_ASSESSED as remediation and it
// never upgrades a provisional benchmark to a compliant one.
func ConsolidateReports(cluster Report, locals []Report, expectedNodes, expectedMasters []string, maxAge time.Duration, now time.Time) (ConsolidatedReport, error) {
	if maxAge <= 0 {
		return ConsolidatedReport{}, fmt.Errorf("maxAge must be positive")
	}
	if cluster.AuditMode != "native" || cluster.NodeName != "" {
		return ConsolidatedReport{}, fmt.Errorf("cluster input must be a cluster-scoped native scan")
	}
	nodes, err := normalizeExpected(expectedNodes)
	if err != nil {
		return ConsolidatedReport{}, fmt.Errorf("expected nodes: %w", err)
	}
	masters, err := normalizeOptionalExpected(expectedMasters)
	if err != nil {
		return ConsolidatedReport{}, fmt.Errorf("expected control-plane nodes: %w", err)
	}
	nodeSet := map[string]bool{}
	for _, n := range nodes {
		nodeSet[n] = true
	}
	for _, n := range masters {
		if !nodeSet[n] {
			return ConsolidatedReport{}, fmt.Errorf("control-plane node %q is not in expected nodes", n)
		}
	}
	all := append([]Report{cluster}, locals...)
	seenSources := map[string]bool{}
	seenFindings := map[string]bool{}
	out := ConsolidatedReport{Report: cluster, SchemaVersion: "1.0", ExpectedNodes: nodes, ExpectedControlPlaneNodes: masters}
	out.Report.AuditMode = "consolidated"
	out.Report.NodeName = ""
	out.Report.Results = make([]Result, 0)
	out.Report.Summary = Summary{}
	out.Report.Compliance = ComplianceSummary{}
	out.Sources = make([]SourceRecord, 0, len(all))
	var scored ComplianceSummary
	for _, src := range all {
		if src.Benchmark != cluster.Benchmark || src.CompatibilityStatus != cluster.CompatibilityStatus {
			return ConsolidatedReport{}, fmt.Errorf("source %s/%s benchmark or compatibility mismatch", src.AuditMode, src.NodeName)
		}
		if src.GeneratedAt.IsZero() || src.GeneratedAt.After(now.Add(2*time.Minute)) || now.Sub(src.GeneratedAt) > maxAge {
			return ConsolidatedReport{}, fmt.Errorf("stale or future source %s/%s timestamp %s", src.AuditMode, src.NodeName, src.GeneratedAt)
		}
		if strings.TrimSpace(src.BuildVersion) == "" || src.BuildVersion != cluster.BuildVersion {
			return ConsolidatedReport{}, fmt.Errorf("source %s/%s build version mismatch", src.AuditMode, src.NodeName)
		}
		key := src.AuditMode + "/" + src.NodeName
		if seenSources[key] {
			return ConsolidatedReport{}, fmt.Errorf("duplicate assessment source %s", key)
		}
		seenSources[key] = true
		if src.Summary.Total != len(src.Results) {
			return ConsolidatedReport{}, fmt.Errorf("source %s summary total mismatch", key)
		}
		var counted Summary
		sourceScored := 0
		for _, r := range src.Results {
			if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.Target) == "" || strings.TrimSpace(r.Status) == "" {
				return ConsolidatedReport{}, fmt.Errorf("source %s contains incomplete finding", key)
			}
			switch r.Status {
			case "PASS", "FAIL", "WARN", "INFO", "MANUAL", "NOT_ASSESSED":
			default:
				return ConsolidatedReport{}, fmt.Errorf("source %s unsupported status %q", key, r.Status)
			}
			incrementSummary(&counted, r.Status)
			if src.AuditMode == "native" && r.Target != "policies" && r.Target != "report" {
				return ConsolidatedReport{}, fmt.Errorf("cluster report has non-policy control %s; run scan --target policies", r.ID)
			}
			if src.AuditMode != "native" {
				if r.Scope != "" && r.Scope != src.NodeName {
					return ConsolidatedReport{}, fmt.Errorf("source %s finding %s scope mismatch", key, r.ID)
				}
				r.Scope = src.NodeName
				if r.Target != strings.TrimSuffix(src.AuditMode, "-local") {
					return ConsolidatedReport{}, fmt.Errorf("source %s finding %s target mismatch", key, r.ID)
				}
			} else if r.Scope != "" {
				return ConsolidatedReport{}, fmt.Errorf("cluster-scoped finding %s unexpectedly has node scope", r.ID)
			}
			id := StableFindingID(r)
			if seenFindings[id] {
				return ConsolidatedReport{}, fmt.Errorf("duplicate logical finding %s across sources", id)
			}
			seenFindings[id] = true
			r.FindingID = id
			out.Results = append(out.Results, r)
			incrementSummary(&out.Summary, r.Status)
			if r.Status == "PASS" || r.Status == "FAIL" {
				out.CoverageMetrics.Assessed++
			}
			out.CoverageMetrics.Total++
			if r.Scored {
				sourceScored++
				scored.ScoredTotal++
				switch r.Status {
				case "PASS":
					scored.ScoredPass++
				case "FAIL":
					scored.ScoredFail++
				case "WARN":
					scored.ScoredWarn++
				case "MANUAL":
					scored.ScoredManual++
				case "NOT_ASSESSED":
					scored.ScoredNotAssessed++
				default:
					scored.ScoredOther++
				}
			} else if strings.HasPrefix(r.ID, "KUBECIS-") {
				scored.Extensions++
			}
		}
		counted.Total = len(src.Results)
		if counted != src.Summary {
			return ConsolidatedReport{}, fmt.Errorf("source %s status counts inconsistent with findings", key)
		}
		if sourceScored != src.Compliance.ScoredTotal {
			return ConsolidatedReport{}, fmt.Errorf("source %s scoredTotal mismatch", key)
		}
		out.Sources = append(out.Sources, SourceRecord{AuditMode: src.AuditMode, NodeName: src.NodeName, GeneratedAt: src.GeneratedAt, Findings: len(src.Results)})
		if src.AuditMode != "native" {
			if src.NodeName == "" || !nodeSet[src.NodeName] {
				return ConsolidatedReport{}, fmt.Errorf("unexpected or empty node name %q", src.NodeName)
			}
			switch src.AuditMode {
			case "node-local":
			case "master-local":
				ok := false
				for _, n := range masters {
					if n == src.NodeName {
						ok = true
						break
					}
				}
				if !ok {
					return ConsolidatedReport{}, fmt.Errorf("control-plane report from unexpected node %q", src.NodeName)
				}
			default:
				return ConsolidatedReport{}, fmt.Errorf("unsupported source auditMode %q", src.AuditMode)
			}
		}
	}
	for _, n := range nodes {
		if !seenSources["node-local/"+n] {
			return ConsolidatedReport{}, fmt.Errorf("missing node-local report for %s", n)
		}
	}
	for _, n := range masters {
		if !seenSources["master-local/"+n] {
			return ConsolidatedReport{}, fmt.Errorf("missing master-local report for %s", n)
		}
	}
	out.Summary.Total = len(out.Results)
	out.CoverageMetrics.ScoredTotal = scored.ScoredTotal
	out.CoverageMetrics.ScoredAssessed = scored.ScoredPass + scored.ScoredFail
	if out.CoverageMetrics.Total > 0 {
		out.CoverageMetrics.OverallPercent = math.Round(float64(out.CoverageMetrics.Assessed)*10000/float64(out.CoverageMetrics.Total)) / 100
	}
	if scored.ScoredTotal > 0 {
		out.CoverageMetrics.ScoredPercent = math.Round(float64(out.CoverageMetrics.ScoredAssessed)*10000/float64(scored.ScoredTotal)) / 100
	}
	scored.CoveragePercent = out.CoverageMetrics.ScoredPercent
	// Do not expose a CIS compliance percentage for this limited starter profile.
	scored.CompliancePercent = nil
	scored.ScoreStatus = "withheld: consolidated KubeCIS native assessment is not a verified complete CIS benchmark; benchmark compatibility is " + cluster.CompatibilityStatus
	out.Compliance = scored
	sort.Slice(out.Results, func(i, j int) bool { return StableFindingID(out.Results[i]) < StableFindingID(out.Results[j]) })
	return out, nil
}

func normalizeExpected(in []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		n := strings.TrimSpace(raw)
		if n == "" {
			continue
		}
		if seen[n] {
			return nil, fmt.Errorf("duplicate node %q", n)
		}
		seen[n] = true
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("explicit expected nodes required")
	}
	sort.Strings(out)
	return out, nil
}

func normalizeOptionalExpected(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, nil
	}
	return normalizeExpected(in)
}

package audit

import (
	"strings"
	"testing"
	"time"
)

func fixtureSource(mode, node string, when time.Time, status string) Report {
	target := "policies"
	if mode == "node-local" {
		target = "node"
	}
	if mode == "master-local" {
		target = "master"
	}
	r := Result{ID: "C-" + target, Target: target, Title: "fixture", Status: status, Scope: node, Scored: true}
	r.FindingID = StableFindingID(r)
	s := Summary{Total: 1}
	incrementSummary(&s, status)
	c := ComplianceSummary{ScoredTotal: 1}
	if status == "PASS" {
		c.ScoredPass = 1
	} else if status == "FAIL" {
		c.ScoredFail = 1
	} else {
		c.ScoredNotAssessed = 1
	}
	return Report{BuildVersion: "test", AuditMode: mode, NodeName: node, GeneratedAt: when, Tool: "KubeCIS Native", Benchmark: "cis-2.1.0", CompatibilityStatus: "provisional", Results: []Result{r}, Summary: s, Compliance: c}
}
func TestConsolidateThreeSources(t *testing.T) {
	now := time.Date(2026, 11, 4, 14, 50, 0, 0, time.UTC)
	cluster := fixtureSource("native", "", now, "FAIL")
	node := fixtureSource("node-local", "kind-control-plane", now, "PASS")
	master := fixtureSource("master-local", "kind-control-plane", now, "NOT_ASSESSED")
	got, err := ConsolidateReports(cluster, []Report{node, master}, []string{"kind-control-plane"}, []string{"kind-control-plane"}, 30*time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary.Total != 3 || got.Summary.Fail != 1 || got.Summary.Pass != 1 || got.Summary.NotAssessed != 1 {
		t.Fatalf("bad merged summary: %+v", got.Summary)
	}
	if got.CoverageMetrics.OverallPercent != 66.67 || got.CoverageMetrics.ScoredPercent != 66.67 {
		t.Fatalf("bad coverage %+v", got.CoverageMetrics)
	}
	if got.Compliance.CompliancePercent != nil || !strings.Contains(got.Compliance.ScoreStatus, "withheld") {
		t.Fatal("provisional score not withheld")
	}
	if got.Results[0].FindingID == "" {
		t.Fatal("missing finding ID")
	}
}
func TestConsolidateRejectsMissingStaleOrWrongScope(t *testing.T) {
	now := time.Now().UTC()
	c := fixtureSource("native", "", now, "FAIL")
	n := fixtureSource("node-local", "n1", now, "PASS")
	tests := []struct {
		name   string
		locals []Report
		want   string
	}{
		{"missing", nil, "missing node-local"},
		{"duplicate", []Report{n, n}, "duplicate assessment source"},
		{"unrelated node", []Report{fixtureSource("node-local", "n2", now, "PASS")}, "unexpected"},
		{"stale", []Report{fixtureSource("node-local", "n1", now.Add(-time.Hour), "PASS")}, "stale"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ConsolidateReports(c, tc.locals, []string{"n1"}, nil, 30*time.Minute, now)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v; wanted %s", err, tc.want)
			}
		})
	}
}
func TestConsolidateRejectsWrongBenchmarkAndMissingSourceCounts(t *testing.T) {
	now := time.Now().UTC()
	c := fixtureSource("native", "", now, "FAIL")
	n := fixtureSource("node-local", "n1", now, "PASS")
	n.Benchmark = "other"
	if _, err := ConsolidateReports(c, []Report{n}, []string{"n1"}, nil, 30*time.Minute, now); err == nil {
		t.Fatal("accepted wrong benchmark")
	}
	n.Benchmark = c.Benchmark
	n.Summary.Total = 2
	if _, err := ConsolidateReports(c, []Report{n}, []string{"n1"}, nil, 30*time.Minute, now); err == nil {
		t.Fatal("accepted incorrect source totals")
	}
}

func TestConsolidateEKSManagedNodesWithoutControlPlaneHost(t *testing.T) {
	now := time.Now().UTC()
	cluster := fixtureSource("native", "", now, "FAIL")
	cluster.CloudProvider = "eks"
	node := fixtureSource("node-local", "ip-10-0-1-1", now, "PASS")
	combined, err := ConsolidateReports(cluster, []Report{node}, []string{"ip-10-0-1-1"}, nil, 30*time.Minute, now)
	if err != nil {
		t.Fatalf("valid EKS aggregation rejected: %v", err)
	}
	if combined.CloudProvider != "eks" || len(combined.Sources) != 2 {
		t.Fatalf("lost EKS identity or node audit source: %+v", combined)
	}
	master := fixtureSource("master-local", "ip-10-0-1-1", now, "NOT_ASSESSED")
	_, err = ConsolidateReports(cluster, []Report{node, master}, []string{"ip-10-0-1-1"}, []string{"ip-10-0-1-1"}, 30*time.Minute, now)
	if err == nil || !strings.Contains(err.Error(), "AWS-managed") {
		t.Fatalf("expected managed control-plane rejection, got: %v", err)
	}
}

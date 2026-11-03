package audit

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrintToGroupsResultsAndPreservesSummary(t *testing.T) {
	var output bytes.Buffer
	percentage := 75.0
	report := Report{
		BuildVersion:   "3.0.0-dev",
		GeneratedAt:    time.Date(2026, 11, 3, 8, 0, 0, 0, time.UTC),
		ClusterVersion: "v1.37.0",
		Benchmark:      "cis-2.1.0",
		Results: []Result{
			{ID: "1.1.1", Title: "API server configuration", Target: "master", Status: "FAIL", Details: "A setting is not compliant."},
			{ID: "5.3.2", Title: "NetworkPolicies are defined", Target: "policies", Status: "WARN", Details: "Some namespaces are uncovered."},
			{ID: "KUBECIS-2.4.1", Title: "Workload security context", Target: "policies", Status: "INFO"},
			{ID: "KUBECIS-2.9.1", Title: "Compliance summary", Target: "report", Status: "INFO"},
		},
		Summary: Summary{Pass: 0, Fail: 1, Warn: 1, Info: 2, Total: 4},
		Compliance: ComplianceSummary{
			ScoredTotal: 4, CoveragePercent: 25, CompliancePercent: &percentage,
			ScoreStatus: "complete",
		},
	}

	PrintTo(&output, report)
	got := output.String()
	for _, expected := range []string{
		"KUBECIS", "3.0.0-dev", "v1.37.0", "cis-2.1.0",
		"MASTER NODE / CONTROL PLANE", "POLICY & WORKLOAD SECURITY",
		"KUBECIS EXTENSIONS", "REPORTING", "[FAIL]", "[WARN]",
		"ASSESSMENT SUMMARY", "TOTAL", "COMPLIANCE", "75.00%",
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("text report missing %q:\n%s", expected, got)
		}
	}
	if strings.Contains(got, "\033[") {
		t.Fatal("buffered report should not contain ANSI escape sequences")
	}
}

func TestPrintToLimitsEvidenceInTerminalOutput(t *testing.T) {
	var output bytes.Buffer
	evidence := []Evidence{
		{Resource: "Pod", Namespace: "default", Name: "one", Field: "securityContext", Observed: "bad", Expected: "good"},
		{Resource: "Pod", Namespace: "default", Name: "two", Field: "securityContext", Observed: "bad", Expected: "good"},
		{Resource: "Pod", Namespace: "default", Name: "three", Field: "securityContext", Observed: "bad", Expected: "good"},
		{Resource: "Pod", Namespace: "default", Name: "four", Field: "securityContext", Observed: "bad", Expected: "good"},
	}
	PrintTo(&output, Report{
		Results: []Result{{ID: "5.2.7", Target: "policies", Status: "FAIL", Title: "Root containers", Evidence: evidence}},
	})
	got := output.String()
	if !strings.Contains(got, "3 more evidence item(s); see JSON report") {
		t.Fatalf("expected evidence truncation note, got:\n%s", got)
	}
	if strings.Contains(got, "default/four") {
		t.Fatalf("terminal output should not print evidence beyond the configured limit:\n%s", got)
	}
}

func TestPrintToPreservesLongDetailsAndNormalizesWindowsPaths(t *testing.T) {
	var output bytes.Buffer
	detail := `Host evidence unavailable at \host\etc\kubernetes\manifests\kube-apiserver.yaml: access denied. ` +
		strings.Repeat("Additional diagnostic context. ", 12)
	PrintTo(&output, Report{
		ClusterVersion: "v1.37.0",
		Results: []Result{{
			ID: "1.1.1", Target: "master", Status: "NOT_ASSESSED",
			Title: "API server manifest permissions", Details: detail,
			Evidence: []Evidence{{
				Resource: "host-file", Name: `\host\etc\kubernetes\manifests\kube-apiserver.yaml`,
				Field: "mode", Observed: "unavailable", Expected: "readable host file",
			}},
		}},
	})
	got := output.String()
	if strings.Contains(got, `\host\`) {
		t.Fatalf("human-readable output should normalize Windows path separators:\n%s", got)
	}
	if !strings.Contains(got, "access denied.") || !strings.Contains(got, "Additional diagnostic context.") {
		t.Fatalf("long diagnostic details should not be truncated:\n%s", got)
	}
	if !strings.Contains(got, "Expected: readable host file") ||
		!strings.Contains(got, "Observed: unavailable") {
		t.Fatalf("evidence labels should remain clearly separated:\n%s", got)
	}
}

func TestPrintToWithContextDisplaysExplicitContext(t *testing.T) {
	var output bytes.Buffer
	PrintToWithContext(&output, Report{ClusterVersion: "v1.37.0"}, "kind-kind")
	got := output.String()
	if !strings.Contains(got, "Context:      kind-kind") {
		t.Fatalf("expected selected kubeconfig context in text output:\n%s", got)
	}
	if !strings.Contains(got, "Kubernetes:   v1.37.0") {
		t.Fatalf("expected Kubernetes version to remain visible:\n%s", got)
	}
}

func TestPrintSummaryStatusCountsAreAligned(t *testing.T) {
	var output bytes.Buffer
	PrintTo(&output, Report{
		Summary: Summary{Pass: 0, Fail: 9, Warn: 5, Manual: 2, NotAssessed: 3, Info: 1, Total: 20},
	})
	const countColumn = 19 // two-space indent + 16-character label field + separator
	rows := 0
	for _, line := range strings.Split(output.String(), "\n") {
		if !strings.Contains(line, "[PASS]") &&
			!strings.Contains(line, "[FAIL]") &&
			!strings.Contains(line, "[WARN]") &&
			!strings.Contains(line, "[MANUAL]") &&
			!strings.Contains(line, "[NOT_ASSESSED]") &&
			!strings.Contains(line, "[INFO]") &&
			!strings.Contains(line, "[TOTAL]") {
			continue
		}
		if len(line) <= countColumn || strings.TrimSpace(line[countColumn:]) == "" {
			t.Fatalf("summary count missing or misaligned:\n%s", output.String())
		}
		rows++
	}
	if rows != 7 {
		t.Fatalf("expected 7 summary rows, got %d:\n%s", rows, output.String())
	}
}

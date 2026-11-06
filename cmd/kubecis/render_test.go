package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
)

func TestRenderBenchStyleDoesNotMaskUnknowns(t *testing.T) {
	r := audit.Report{
		BuildVersion:        "test",
		Benchmark:           "cis-eks-2.1.0",
		CompatibilityStatus: "provisional",
		ClusterVersion:      "v1.35.9",
		CloudProvider:       "eks",
		Results: []audit.Result{
			{ID: "4.1.5", Title: "Service accounts", Status: "MANUAL", Details: "Needs review", Remediation: "Review workloads"},
			{ID: "3.2.2", Scope: "node-a", Title: "Authorization", Status: "PASS"},
			{ID: "3.2.1", Scope: "node-a", Title: "Anonymous authentication", Status: "FAIL", Remediation: "Disable anonymous auth"},
			{ID: "3.1.1", Scope: "node-a", Title: "Missing evidence", Status: "NOT_ASSESSED", Details: "Collector not implemented"},
		},
	}
	var buf bytes.Buffer
	if err := renderBenchStyle(&buf, r); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"[FAIL] 3.2.1", "[PASS] 3.2.2", "[MANUAL] 4.1.5",
		"[NOT_ASSESSED] 3.1.1", "1 checks PASS", "1 checks FAIL",
		"1 checks MANUAL", "1 checks NOT_ASSESSED",
		"Compliance score: WITHHELD", "== Remediations 3.2 ==",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
	if strings.Index(out, "[NOT_ASSESSED] 3.1.1") > strings.Index(out, "[FAIL] 3.2.1") {
		t.Error("CIS identifiers are not sorted")
	}
}

func TestRenderBenchStyleKeepsMultipleNodes(t *testing.T) {
	r := audit.Report{
		CompatibilityStatus: "provisional",
		Results: []audit.Result{
			{ID: "3.1.3", Scope: "node-a", Status: "PASS"},
			{ID: "3.1.3", Scope: "node-b", Status: "FAIL"},
		},
	}
	var buf bytes.Buffer
	if err := renderBenchStyle(&buf, r); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "[node: node-a]") || !strings.Contains(out, "[node: node-b]") {
		t.Error("node-specific evidence was dropped")
	}
}

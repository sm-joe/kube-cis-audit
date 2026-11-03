package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
)

func TestWriteReportJSONToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "report.json")
	report := audit.Report{Tool: "KubeCIS", Benchmark: "cis-2.1.0", Summary: audit.Summary{Pass: 1, Total: 1}}

	if err := writeReport("json", path, report); err != nil {
		t.Fatalf("writeReport returned error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var got audit.Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	if got.Tool != "KubeCIS" || got.Summary.Total != 1 {
		t.Fatalf("unexpected report content: %+v", got)
	}
}

func TestWriteReportTextToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.txt")
	report := audit.Report{Tool: "KubeCIS", Benchmark: "cis-2.1.0", Summary: audit.Summary{Pass: 1, Total: 1}}

	if err := writeReport("text", path, report); err != nil {
		t.Fatalf("writeReport returned error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	if !strings.Contains(string(data), "KUBECIS") || !strings.Contains(string(data), "TOTAL") {
		t.Fatalf("text report is missing expected content: %s", data)
	}
}

func TestWriteReportRejectsUnknownFormatBeforeCreatingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.out")
	err := writeReport("yaml", path, audit.Report{})
	if err == nil {
		t.Fatal("expected unsupported format error")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("output file should not be created for unsupported format; stat error: %v", statErr)
	}
}

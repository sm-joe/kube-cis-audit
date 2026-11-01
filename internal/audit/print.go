package audit

import (
	"fmt"
	"strings"
)

func Print(r Report) {
	fmt.Printf("\nKubeCIS | Kubernetes CIS configuration audit\n")
	if r.AuditMode == "node-local" {
		fmt.Printf("Mode: node-local   Node: %s   Benchmark: %s\n", r.NodeName, r.Benchmark)
	} else {
		fmt.Printf("Cluster: %s   Benchmark: %s\n", r.ClusterVersion, r.Benchmark)
	}
	fmt.Printf("Generated: %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	if r.CompatibilityStatus != "" {
		fmt.Printf("Benchmark compatibility: %s\n", strings.ToUpper(r.CompatibilityStatus))
	}
	if r.CompatibilityNote != "" {
		fmt.Printf("NOTE: %s\n", r.CompatibilityNote)
	}
	fmt.Println(strings.Repeat("─", 100))
	for _, x := range r.Results {
		fmt.Printf("%-13s %-8s %-66s %s\n", symbol(x.Status), x.ID, truncate(x.Title, 66), x.Details)
		if x.Status == "FAIL" && x.Remediation != "" {
			fmt.Printf("              Remediation: %s\n", x.Remediation)
		}
		for _, e := range x.Evidence {
			name := e.Name
			if e.Namespace != "" {
				name = e.Namespace + "/" + name
			}
			fmt.Printf("              Evidence: %s %s %s observed=%q expected=%q\n", e.Resource, name, e.Field, e.Observed, e.Expected)
		}
	}
	fmt.Println(strings.Repeat("─", 100))
	fmt.Printf("SUMMARY  PASS: %d  FAIL: %d  WARN: %d  MANUAL: %d  NOT_ASSESSED: %d  INFO: %d  TOTAL: %d\n\n", r.Summary.Pass, r.Summary.Fail, r.Summary.Warn, r.Summary.Manual, r.Summary.NotAssessed, r.Summary.Info, r.Summary.Total)
}

func symbol(s string) string {
	switch s {
	case "PASS":
		return "[PASS]"
	case "FAIL":
		return "[FAIL]"
	case "WARN":
		return "[WARN]"
	case "MANUAL":
		return "[MANUAL]"
	case "NOT_ASSESSED":
		return "[NOT ASSESSED]"
	default:
		return "[INFO]"
	}
}
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

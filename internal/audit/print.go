package audit

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	reportWidth     = 88
	maxTextEvidence = 1
)

// Print writes a human-readable report to standard output.
func Print(r Report) {
	PrintTo(os.Stdout, r)
}

// PrintTo writes a human-readable report to the supplied destination.
// ANSI colors are enabled only for an interactive terminal; redirected output
// and test buffers remain plain text. JSON output is handled separately.
func PrintTo(w io.Writer, r Report) {
	PrintToWithContext(w, r, "")
}

// PrintToWithContext writes the text report and includes the explicitly
// selected kubeconfig context when the CLI supplied one. It does not alter
// the report model or JSON representation.
func PrintToWithContext(w io.Writer, r Report, kubeContext string) {
	color := supportsColor(w)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s  %s\n", paint(color, ansiBold+ansiCyan, "KUBECIS"), paint(color, ansiBold, "Kubernetes Security Assessment"))
	if r.BuildVersion != "" {
		fmt.Fprintf(w, "Version:      %s\n", r.BuildVersion)
	}
	if r.AuditMode == "node-local" {
		fmt.Fprintf(w, "Mode:         Node-local\nNode:         %s\n", valueOrDash(r.NodeName))
	} else {
		fmt.Fprintf(w, "Context:      %s\n", valueOrDefault(kubeContext, "current kubeconfig context"))
		fmt.Fprintf(w, "Kubernetes:   %s\n", valueOrDash(r.ClusterVersion))
	}
	fmt.Fprintf(w, "Benchmark:    %s\n", valueOrDash(r.Benchmark))
	fmt.Fprintf(w, "Generated:    %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	if r.CompatibilityStatus != "" {
		fmt.Fprintf(w, "Compatibility: %s\n", strings.ToUpper(r.CompatibilityStatus))
	}
	if r.CompatibilityNote != "" {
		fmt.Fprintf(w, "Note:         %s\n", r.CompatibilityNote)
	}
	printRule(w)

	section := ""
	for _, result := range r.Results {
		nextSection := resultSection(result)
		if nextSection != section {
			if section != "" {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s\n", paint(color, ansiBold+ansiCyan, nextSection))
			section = nextSection
		}
		printResult(w, color, result)
	}
	printRule(w)
	printSummary(w, color, r)
	fmt.Fprintln(w)
}

func printResult(w io.Writer, color bool, result Result) {
	status := strings.ToUpper(result.Status)
	id := result.ID
	if id == "" {
		id = "UNKNOWN"
	}
	fmt.Fprintf(w, "  %-15s %-12s %s\n",
		paintStatus(color, status),
		paint(color, ansiBold, id),
		result.Title,
	)
	if detail := strings.TrimSpace(result.Details); detail != "" {
		fmt.Fprintf(w, "                   %s\n", wrapText(normalizeDisplayPaths(detail), 76, "                   "))
	}

	// Evidence can be extensive (for example, cluster-wide RBAC inventory).
	// Keep the terminal report readable while preserving every evidence item
	// in the JSON report.
	if len(result.Evidence) > 0 && status != "PASS" && status != "INFO" {
		limit := len(result.Evidence)
		if limit > maxTextEvidence {
			limit = maxTextEvidence
		}
		for _, evidence := range result.Evidence[:limit] {
			fmt.Fprintf(w, "                   • %s\n", wrapText(evidenceSummary(evidence), 76, "                     "))
		}
		if remaining := len(result.Evidence) - limit; remaining > 0 {
			fmt.Fprintf(w, "                   … %d more evidence item(s); see JSON report\n", remaining)
		}
	}
	if result.Remediation != "" && status != "PASS" && status != "INFO" {
		fmt.Fprintf(w, "                   %s %s\n",
			paint(color, ansiBold+ansiYellow, "Remediation:"),
			wrapText(normalizeDisplayPaths(result.Remediation), 65, "                                "))
	}
}

func printSummary(w io.Writer, color bool, r Report) {
	fmt.Fprintf(w, "%s\n", paint(color, ansiBold, "ASSESSMENT SUMMARY"))
	fmt.Fprintf(w, "  %-16s %5s\n", "STATUS", "COUNT")
	printStatusCount(w, color, "PASS", r.Summary.Pass)
	printStatusCount(w, color, "FAIL", r.Summary.Fail)
	printStatusCount(w, color, "WARN", r.Summary.Warn)
	printStatusCount(w, color, "MANUAL", r.Summary.Manual)
	printStatusCount(w, color, "NOT_ASSESSED", r.Summary.NotAssessed)
	printStatusCount(w, color, "INFO", r.Summary.Info)
	printStatusCount(w, color, "TOTAL", r.Summary.Total)

	if r.Compliance.ScoreStatus != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", paint(color, ansiBold, "COMPLIANCE"))
		fmt.Fprintf(w, "  Scored controls: %d\n", r.Compliance.ScoredTotal)
		fmt.Fprintf(w, "  Assessment coverage: %.2f%%\n", r.Compliance.CoveragePercent)
		if r.Compliance.CompliancePercent == nil {
			fmt.Fprintf(w, "  Compliance score: %s\n",
				paint(color, ansiBold+ansiYellow, "WITHHELD"))
			fmt.Fprintf(w, "  Reason: %s\n", r.Compliance.ScoreStatus)
		} else {
			fmt.Fprintf(w, "  Compliance score: %.2f%%\n", *r.Compliance.CompliancePercent)
			fmt.Fprintf(w, "  Status: %s\n", r.Compliance.ScoreStatus)
		}
	}
}

// printStatusCount pads the uncolored label before applying ANSI styling.
// This keeps the count column aligned in both colored terminals and plain text.
func printStatusCount(w io.Writer, color bool, status string, count int) {
	label := "[" + status + "]"
	padding := 16 - len(label)
	if padding < 1 {
		padding = 1
	}
	fmt.Fprintf(w, "  %s%s %5d\n",
		paintStatus(color, status), strings.Repeat(" ", padding), count)
}

func resultSection(result Result) string {
	if strings.HasPrefix(result.ID, "KUBECIS-") && strings.ToLower(result.Target) != "report" {
		return "KUBECIS EXTENSIONS"
	}
	switch strings.ToLower(result.Target) {
	case "master":
		return "MASTER NODE / CONTROL PLANE"
	case "node":
		return "NODE SECURITY"
	case "policies":
		return "POLICY & WORKLOAD SECURITY"
	case "report":
		return "REPORTING"
	default:
		if strings.HasPrefix(result.ID, "KUBECIS-") {
			return "KUBECIS EXTENSIONS"
		}
		return "SECURITY ASSESSMENT"
	}
}

func evidenceSummary(e Evidence) string {
	var parts []string
	resource := strings.TrimSpace(e.Resource)
	name := strings.TrimSpace(e.Name)
	if e.Namespace != "" {
		if name != "" {
			name = e.Namespace + "/" + name
		} else {
			name = e.Namespace
		}
	}
	if resource != "" || name != "" {
		parts = append(parts, strings.TrimSpace(resource+" "+name))
	}
	if field := strings.TrimSpace(e.Field); field != "" {
		parts = append(parts, "Field: "+field)
	}
	if observed := strings.TrimSpace(e.Observed); observed != "" {
		parts = append(parts, "Observed: "+observed)
	}
	if expected := strings.TrimSpace(e.Expected); expected != "" {
		parts = append(parts, "Expected: "+expected)
	}
	return normalizeDisplayPaths(strings.Join(parts, " | "))
}

// normalizeDisplayPaths changes Windows separators only in the human-readable
// report. The underlying evidence and JSON values remain untouched.
func normalizeDisplayPaths(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func wrapText(value string, width int, indent string) string {
	value = strings.Join(strings.Fields(value), " ")
	if width <= 0 || utf8.RuneCountInString(value) <= width {
		return value
	}
	words := strings.Fields(value)
	var lines []string
	line := ""
	for _, word := range words {
		if line == "" {
			line = word
			continue
		}
		if utf8.RuneCountInString(line)+1+utf8.RuneCountInString(word) > width {
			lines = append(lines, line)
			line = indent + word
		} else {
			line += " " + word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func printRule(w io.Writer) {
	fmt.Fprintln(w, strings.Repeat("─", reportWidth))
}

const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiCyan    = "\033[36m"
	ansiMagenta = "\033[35m"
)

func supportsColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func paint(enabled bool, style, value string) string {
	if !enabled {
		return value
	}
	return style + value + ansiReset
}

func paintStatus(enabled bool, status string) string {
	style := ansiCyan
	switch status {
	case "PASS":
		style = ansiGreen
	case "FAIL":
		style = ansiRed
	case "WARN":
		style = ansiYellow
	case "MANUAL":
		style = ansiMagenta
	case "NOT_ASSESSED":
		style = ansiYellow
	case "INFO":
		style = ansiCyan
	}
	return paint(enabled, ansiBold+style, "["+status+"]")
}

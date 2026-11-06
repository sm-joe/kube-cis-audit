package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
	"github.com/spf13/cobra"
)

func main() {
	var kubeconfig, kubeContext, configDir, benchmark, target, output, outputFile, namespace, engine, upstreamBenchmark string
	var hostRoot string
	var exitCodeOnFail bool

	root := &cobra.Command{
		Use:           "kubecis",
		Short:         "Read-only native Go Kubernetes security auditor",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       audit.BuildVersion,
	}

	scan := &cobra.Command{
		Use:   "scan",
		Short: "Audit a Kubernetes cluster using the Kubernetes API",
		Long:  "Run a read-only Kubernetes security assessment. Credentials are taken from --kubeconfig, in-cluster ServiceAccount credentials, KUBECONFIG, or the default kubeconfig.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var report audit.Report
			var err error
			// kube-bench has been removed; retain the legacy flags so older
			// automation receives an actionable error rather than an unknown flag.
			if strings.ToLower(engine) != "native" {
				return fmt.Errorf("engine %q is no longer available: KubeCIS Native supports --engine native only; remove --upstream-benchmark and use native node-audit for host checks", engine)
			}
			if upstreamBenchmark != "auto" {
				return fmt.Errorf("--upstream-benchmark has been removed; use --benchmark for KubeCIS Native")
			}
			report, err = audit.Run(audit.Options{
				Kubeconfig: kubeconfig, KubeContext: kubeContext, ConfigDir: configDir,
				Benchmark: benchmark, Target: target, HostRoot: hostRoot, Namespace: namespace,
			})
			if err != nil {
				return err
			}
			if err := writeReportWithContext(output, outputFile, report, kubeContext); err != nil {
				return err
			}
			if exitCodeOnFail && report.Summary.Fail > 0 {
				return fmt.Errorf("audit completed with %d failed check(s)", report.Summary.Fail)
			}
			return nil
		},
	}
	scan.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig (defaults to KUBECONFIG or ~/.kube/config)")
	scan.Flags().StringVar(&kubeContext, "context", "", "Kubeconfig context to use (defaults to the current context)")
	scan.Flags().StringVar(&configDir, "config-dir", defaultConfigDir(), "Directory containing benchmark YAML files")
	scan.Flags().StringVar(&benchmark, "benchmark", "auto", "Benchmark ID (e.g. cis-2.1.0) or auto-detect from Kubernetes version")
	scan.Flags().StringVar(&target, "target", "all", "Assessment target: all, master, node, or policies")
	scan.Flags().StringVar(&output, "output", "text", "Report format: text or json")
	scan.Flags().StringVar(&outputFile, "output-file", "", "Write the report to this file instead of standard output")
	scan.Flags().StringVar(&hostRoot, "host-root", "/host", "Mounted host filesystem root for host-file checks")
	scan.Flags().StringVar(&namespace, "namespace", "", "Limit namespaced API checks to one namespace (empty means all)")
	scan.Flags().StringVar(&engine, "engine", "native", "Audit engine (native only; legacy values are rejected)")
	scan.Flags().StringVar(&upstreamBenchmark, "upstream-benchmark", "auto", "Deprecated: upstream profile selection is unsupported; only auto accepted")
	scan.Flags().BoolVar(&exitCodeOnFail, "exit-code-on-fail", true, "Exit non-zero when one or more checks are FAIL (set false for observational scans)")

	var nodeBenchmark, nodeConfigDir, nodeHostRoot, nodeName, nodeOutput, nodeOutputFile, nodeMetadataFile, nodeTarget, nodeProvider string
	var nodeInterval time.Duration
	var nodeExitCodeOnFail bool
	nodeAudit := &cobra.Command{
		Use:   "node-audit",
		Short: "Perform read-only host configuration checks on this Kubernetes node",
		RunE: func(cmd *cobra.Command, args []string) error {
			if nodeName == "" {
				nodeName = os.Getenv("NODE_NAME")
			}
			if nodeName == "" {
				return fmt.Errorf("node name is required; pass --node-name or set NODE_NAME")
			}
			run := func() error {
				report, err := audit.RunNode(audit.Options{ConfigDir: nodeConfigDir, Benchmark: nodeBenchmark, Target: nodeTarget, HostRoot: nodeHostRoot, MetadataFile: nodeMetadataFile, Provider: nodeProvider}, nodeName)
				if err != nil {
					return err
				}
				if err := writeReport(nodeOutput, nodeOutputFile, report); err != nil {
					return err
				}
				if nodeInterval <= 0 && nodeExitCodeOnFail && report.Summary.Fail > 0 {
					return fmt.Errorf("node audit completed with %d failed check(s)", report.Summary.Fail)
				}
				return nil
			}
			if nodeInterval <= 0 {
				return run()
			}
			for {
				if err := run(); err != nil {
					fmt.Fprintln(os.Stderr, "node audit error:", err)
				}
				time.Sleep(nodeInterval)
			}
		},
	}
	nodeAudit.Flags().StringVar(&nodeBenchmark, "benchmark", "", "Explicit benchmark ID (required in host mode)")
	nodeAudit.Flags().StringVar(&nodeTarget, "target", "node", "Host target: node or master (control-plane)")
	nodeAudit.Flags().StringVar(&nodeConfigDir, "config-dir", defaultConfigDir(), "Directory containing benchmark YAML files")
	nodeAudit.Flags().StringVar(&nodeHostRoot, "host-root", "/host", "Read-only mounted node filesystem root (direct mode)")
	nodeAudit.Flags().StringVar(&nodeMetadataFile, "metadata-file", "", "Fresh metadata snapshot path produced by node-metadata")
	nodeAudit.Flags().StringVar(&nodeProvider, "provider", "", "Explicit node provider (eks for AL2023 nodeadm EKS extension; empty for generic Kubernetes)")
	nodeAudit.Flags().StringVar(&nodeName, "node-name", "", "Kubernetes node name (defaults to NODE_NAME)")
	nodeAudit.Flags().StringVar(&nodeOutput, "output", "json", "Report format: json or text")
	nodeAudit.Flags().StringVar(&nodeOutputFile, "output-file", "", "Write the report to this file instead of standard output")
	nodeAudit.Flags().DurationVar(&nodeInterval, "interval", 0, "Repeat assessment interval (e.g. 24h); zero runs once")
	nodeAudit.Flags().BoolVar(&nodeExitCodeOnFail, "exit-code-on-fail", true, "Exit non-zero when a one-shot node audit has FAIL results")

	var metadataBenchmark, metadataConfigDir, metadataHostRoot, metadataKubeletConfigRoot, metadataNodeName, metadataOutputFile, metadataTarget string
	var metadataInterval time.Duration
	nodeMetadata := &cobra.Command{
		Use:   "node-metadata",
		Short: "Collect allowlisted node file metadata for the unprivileged node auditor",
		RunE: func(cmd *cobra.Command, args []string) error {
			if metadataNodeName == "" {
				metadataNodeName = os.Getenv("NODE_NAME")
			}
			if metadataNodeName == "" {
				return fmt.Errorf("node name is required; pass --node-name or set NODE_NAME")
			}
			run := func() error {
				return audit.RunHostMetadataSnapshotWithKubeletRoot(metadataConfigDir, metadataBenchmark, metadataTarget, metadataHostRoot, metadataKubeletConfigRoot, metadataNodeName, metadataOutputFile)
			}
			if metadataInterval <= 0 {
				return run()
			}
			for {
				if err := run(); err != nil {
					fmt.Fprintln(os.Stderr, "node metadata collection error:", err)
				} else {
					fmt.Fprintln(os.Stderr, "node metadata snapshot refreshed")
				}
				time.Sleep(metadataInterval)
			}
		},
	}
	nodeMetadata.Flags().StringVar(&metadataBenchmark, "benchmark", "", "Explicit benchmark ID (required)")
	nodeMetadata.Flags().StringVar(&metadataTarget, "target", "node", "Host target: node or master (control-plane)")
	nodeMetadata.Flags().StringVar(&metadataConfigDir, "config-dir", defaultConfigDir(), "Directory containing benchmark YAML files")
	nodeMetadata.Flags().StringVar(&metadataHostRoot, "host-root", "/host-kubelet", "Read-only mount of the node kubelet directory")
	nodeMetadata.Flags().StringVar(&metadataKubeletConfigRoot, "kubelet-config-root", "", "Optional read-only AL2023 /etc/kubernetes/kubelet mount containing config.json")
	nodeMetadata.Flags().StringVar(&metadataNodeName, "node-name", "", "Kubernetes node name (defaults to NODE_NAME)")
	nodeMetadata.Flags().StringVar(&metadataOutputFile, "output-file", "/metadata/snapshot.json", "Writable shared-volume path for metadata-only JSON snapshot")
	nodeMetadata.Flags().DurationVar(&metadataInterval, "interval", 30*time.Second, "Snapshot refresh interval; zero runs once")

	version := &cobra.Command{Use: "version", Short: "Print KubeCIS build version", Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(audit.BuildVersion)
	}}
	var compareBaseline, compareCurrent, compareOutput string
	compare := &cobra.Command{
		Use:   "compare",
		Short: "Compare two KubeCIS JSON reports using stable finding identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			baselineBytes, err := os.ReadFile(compareBaseline)
			if err != nil {
				return fmt.Errorf("read baseline report: %w", err)
			}
			currentBytes, err := os.ReadFile(compareCurrent)
			if err != nil {
				return fmt.Errorf("read current report: %w", err)
			}
			baseline, err := audit.DecodeReport(baselineBytes)
			if err != nil {
				return fmt.Errorf("baseline report: %w", err)
			}
			current, err := audit.DecodeReport(currentBytes)
			if err != nil {
				return fmt.Errorf("current report: %w", err)
			}
			comparison := audit.CompareReports(baseline, current)
			var out io.Writer = os.Stdout
			if compareOutput != "" {
				if err := os.MkdirAll(filepath.Dir(compareOutput), 0o755); err != nil && filepath.Dir(compareOutput) != "." {
					return err
				}
				f, err := os.Create(compareOutput)
				if err != nil {
					return err
				}
				defer f.Close()
				out = f
			}
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(comparison)
		},
	}
	compare.Flags().StringVar(&compareBaseline, "baseline", "", "Baseline KubeCIS JSON report (required)")
	compare.Flags().StringVar(&compareCurrent, "current", "", "Current KubeCIS JSON report (required)")
	compare.Flags().StringVar(&compareOutput, "output-file", "", "Write comparison JSON to this file (defaults to stdout)")
	_ = compare.MarkFlagRequired("baseline")
	_ = compare.MarkFlagRequired("current")
	var mergeCluster, mergeOutput, mergeNodes, mergeMasters, mergeProviderEvidence, mergeFormat string
	var mergeLocals []string
	var mergeMaxAge time.Duration
	consolidate := &cobra.Command{
		Use: "consolidate", Short: "Validate and merge cluster, node and control-plane reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			read := func(path string) (audit.Report, error) {
				data, err := os.ReadFile(path)
				if err != nil {
					return audit.Report{}, fmt.Errorf("read %s: %w", path, err)
				}
				return audit.DecodeReport(data)
			}
			clusterReport, err := read(mergeCluster)
			if err != nil {
				return err
			}
			locals := make([]audit.Report, 0, len(mergeLocals))
			for _, path := range mergeLocals {
				r, err := read(path)
				if err != nil {
					return err
				}
				locals = append(locals, r)
			}
			merged, err := audit.ConsolidateReports(clusterReport, locals, strings.Split(mergeNodes, ","), splitOptionalCSV(mergeMasters), mergeMaxAge, time.Now().UTC())
			if err != nil {
				return err
			}
			if mergeProviderEvidence != "" {
				providerData, err := os.ReadFile(mergeProviderEvidence)
				if err != nil {
					return fmt.Errorf("read EKS provider evidence: %w", err)
				}
				evidence, err := audit.DecodeEKSProviderEvidence(providerData, clusterReport, mergeMaxAge, time.Now().UTC())
				if err != nil {
					return err
				}
				if err := audit.AttachEKSCISProviderEvidence(&merged, evidence); err != nil {
					return fmt.Errorf("assess CIS EKS provider evidence: %w", err)
				}
			}
			var dst io.Writer = os.Stdout
			if mergeOutput != "" {
				if err := os.MkdirAll(filepath.Dir(mergeOutput), 0o755); err != nil {
					return err
				}
				f, err := os.Create(mergeOutput)
				if err != nil {
					return err
				}
				defer f.Close()
				dst = f
			}
			switch strings.ToLower(mergeFormat) {
			case "json":
				encoder := json.NewEncoder(dst)
				encoder.SetIndent("", "  ")
				return encoder.Encode(merged)
			case "text":
				return renderBenchStyle(dst, merged.Report)
			default:
				return fmt.Errorf("unsupported consolidated output format %q (use json or text)", mergeFormat)
			}
		},
	}
	consolidate.Flags().StringVar(&mergeCluster, "cluster", "", "Cluster-policy JSON report")
	consolidate.Flags().StringVar(&mergeProviderEvidence, "provider-evidence", "", "Optional validated read-only AWS EKS DescribeCluster inventory JSON")
	consolidate.Flags().StringSliceVar(&mergeLocals, "local", nil, "Node/control-plane JSON reports (repeat flag)")
	consolidate.Flags().StringVar(&mergeNodes, "expected-nodes", "", "Comma-separated names of ALL expected Linux nodes")
	consolidate.Flags().StringVar(&mergeMasters, "expected-control-planes", "", "Comma-separated names of expected control-plane nodes")
	consolidate.Flags().DurationVar(&mergeMaxAge, "max-age", 30*time.Minute, "Maximum age of each source report")
	consolidate.Flags().StringVar(&mergeOutput, "output-file", "", "Consolidated output file (default stdout)")
	consolidate.Flags().StringVar(&mergeFormat, "format", "json", "Consolidated output format: json or text")
	_ = consolidate.MarkFlagRequired("cluster")
	_ = consolidate.MarkFlagRequired("expected-nodes")
	root.AddCommand(consolidate)
	root.AddCommand(newRenderCommand())
	root.AddCommand(newReviewCommand())
	root.AddCommand(newReviewInitCommand())
	root.AddCommand(scan, nodeAudit, nodeMetadata, compare, version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func writeReport(format, path string, report audit.Report) error {
	return writeReportWithContext(format, path, report, "")
}

func writeReportWithContext(format, path string, report audit.Report, kubeContext string) error {
	if normalized := strings.ToLower(format); normalized != "json" && normalized != "text" && normalized != "" {
		return fmt.Errorf("unsupported output %q (use text or json)", format)
	}
	var out io.Writer = os.Stdout
	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
			return fmt.Errorf("create report directory: %w", err)
		}
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create report file %q: %w", path, err)
		}
		defer file.Close()
		out = file
	}

	switch strings.ToLower(format) {
	case "json":
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write JSON report: %w", err)
		}
	case "text", "":
		audit.PrintToWithContext(out, report, kubeContext)
	default:
		return fmt.Errorf("unsupported output %q (use text or json)", format)
	}
	return nil
}

func defaultConfigDir() string {
	if value := os.Getenv("KUBECIS_CONFIG_DIR"); value != "" {
		return value
	}
	// Prefer a working-directory profile for source checkouts and containers.
	local := filepath.Join("benchmarks")
	if info, err := os.Stat(local); err == nil && info.IsDir() {
		return local
	}
	// Release archives install the binary and profiles side by side. Resolve
	// profiles relative to the executable so callers can run from any directory.
	if executable, err := os.Executable(); err == nil {
		adjacent := filepath.Join(filepath.Dir(executable), "benchmarks")
		if info, statErr := os.Stat(adjacent); statErr == nil && info.IsDir() {
			return adjacent
		}
	}
	return local
}

func splitOptionalCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.Split(s, ",")
}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
	"github.com/spf13/cobra"
)

func main() {
	var kubeconfig, configDir, benchmark, target, output, namespace string
	var hostRoot string
	root := &cobra.Command{Use: "kubecis", Short: "Kubernetes CIS security configuration auditor", SilenceUsage: true}
	scan := &cobra.Command{Use: "scan", Short: "Audit a Kubernetes cluster", RunE: func(cmd *cobra.Command, args []string) error {
		opts := audit.Options{Kubeconfig: kubeconfig, ConfigDir: configDir, Benchmark: benchmark, Target: target, HostRoot: hostRoot, Namespace: namespace}
		report, err := audit.Run(opts)
		if err != nil {
			return err
		}
		switch strings.ToLower(output) {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(report); err != nil {
				return err
			}
		case "text", "":
			audit.Print(report)
		default:
			return fmt.Errorf("unsupported output %q (use text or json)", output)
		}
		if report.Summary.Fail > 0 {
			return fmt.Errorf("audit completed with %d failed check(s)", report.Summary.Fail)
		}
		return nil
	}}
	scan.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file (otherwise auto-detect in-cluster credentials or standard KUBECONFIG/~/.kube/config)")
	scan.Flags().StringVar(&configDir, "config-dir", defaultConfigDir(), "Directory containing benchmark YAML files")
	scan.Flags().StringVar(&benchmark, "benchmark", "auto", "Benchmark identifier, e.g. cis-2.0.1, or auto")
	scan.Flags().StringVar(&target, "target", "all", "Targets: all, master, node, policies")
	scan.Flags().StringVar(&output, "output", "text", "Output format: text or json")
	scan.Flags().StringVar(&hostRoot, "host-root", "/host", "Mounted host filesystem root for node/control-plane file checks")
	scan.Flags().StringVar(&namespace, "namespace", "", "Limit namespaced API checks to one namespace (empty means all)")
	var nodeBenchmark, nodeConfigDir, nodeHostRoot, nodeName, nodeOutput, nodeMetadataFile string
	var nodeInterval time.Duration
	nodeAudit := &cobra.Command{Use: "node-audit", Short: "Perform read-only host configuration checks on this Kubernetes node", RunE: func(cmd *cobra.Command, args []string) error {
		if nodeName == "" {
			nodeName = os.Getenv("NODE_NAME")
		}
		if nodeName == "" {
			return fmt.Errorf("node name is required; pass --node-name or set NODE_NAME")
		}
		if nodeMetadataFile != "" {
			if err := waitForFile(nodeMetadataFile, 60*time.Second); err != nil {
				return fmt.Errorf("wait for metadata helper: %w", err)
			}
		}
		run := func() error {
			report, err := audit.RunNode(audit.Options{ConfigDir: nodeConfigDir, Benchmark: nodeBenchmark, Target: "node", HostRoot: nodeHostRoot, MetadataFile: nodeMetadataFile}, nodeName)
			if err != nil {
				return err
			}
			switch strings.ToLower(nodeOutput) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(report); err != nil {
					return err
				}
			case "text", "":
				audit.Print(report)
			default:
				return fmt.Errorf("unsupported output %q (use text or json)", nodeOutput)
			}
			if nodeInterval <= 0 && report.Summary.Fail > 0 {
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
	}}
	nodeAudit.Flags().StringVar(&nodeBenchmark, "benchmark", "", "Explicit benchmark identifier, e.g. cis-2.1.0 (required; node mode does not query API version)")
	nodeAudit.Flags().StringVar(&nodeConfigDir, "config-dir", defaultConfigDir(), "Directory containing benchmark YAML files")
	nodeAudit.Flags().StringVar(&nodeHostRoot, "host-root", "/host", "Read-only mounted node filesystem root (standalone mode)")
	nodeAudit.Flags().StringVar(&nodeMetadataFile, "metadata-file", "", "Metadata snapshot produced by the node-metadata helper")
	nodeAudit.Flags().StringVar(&nodeName, "node-name", "", "Kubernetes node name (defaults to NODE_NAME)")
	nodeAudit.Flags().StringVar(&nodeOutput, "output", "json", "Output format: json or text")
	nodeAudit.Flags().DurationVar(&nodeInterval, "interval", 0, "Repeat assessment interval (e.g. 24h); zero runs once")

	var metadataBenchmark, metadataConfigDir, metadataHostRoot, metadataNodeName, metadataOutputFile string
	var metadataInterval time.Duration
	nodeMetadata := &cobra.Command{Use: "node-metadata", Short: "Collect allowlisted node file metadata for node-audit", RunE: func(cmd *cobra.Command, args []string) error {
		if metadataNodeName == "" {
			metadataNodeName = os.Getenv("NODE_NAME")
		}
		if metadataNodeName == "" {
			return fmt.Errorf("node name is required; pass --node-name or set NODE_NAME")
		}
		collect := func() error {
			return audit.RunNodeMetadataSnapshot(metadataConfigDir, metadataBenchmark, metadataHostRoot, metadataNodeName, metadataOutputFile)
		}
		if metadataInterval <= 0 {
			return collect()
		}
		for {
			if err := collect(); err != nil {
				fmt.Fprintln(os.Stderr, "node metadata error:", err)
			}
			time.Sleep(metadataInterval)
		}
	}}
	nodeMetadata.Flags().StringVar(&metadataBenchmark, "benchmark", "", "Explicit benchmark identifier")
	nodeMetadata.Flags().StringVar(&metadataConfigDir, "config-dir", defaultConfigDir(), "Directory containing benchmark YAML files")
	nodeMetadata.Flags().StringVar(&metadataHostRoot, "host-root", "/host-kubelet", "Read-only mount of the kubelet directory")
	nodeMetadata.Flags().StringVar(&metadataNodeName, "node-name", "", "Kubernetes node name (defaults to NODE_NAME)")
	nodeMetadata.Flags().StringVar(&metadataOutputFile, "output-file", "/shared/node-metadata.json", "Shared metadata snapshot output path")
	nodeMetadata.Flags().DurationVar(&metadataInterval, "interval", 15*time.Second, "Metadata refresh interval")

	root.AddCommand(scan)
	root.AddCommand(nodeAudit)
	root.AddCommand(nodeMetadata)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
func defaultConfigDir() string {
	if v := os.Getenv("KUBECIS_CONFIG_DIR"); v != "" {
		return v
	}
	return filepath.Join("benchmarks")
}

func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("metadata file %s was not created within %s", path, timeout)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

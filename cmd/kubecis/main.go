package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	root.AddCommand(scan)
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

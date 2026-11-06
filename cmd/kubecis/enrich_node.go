package main

import (
	"fmt"
	"os"

	"github.com/sm-joe/kube-cis-audit/internal/audit"
	"github.com/spf13/cobra"
)

func newEnrichNodeCommand() *cobra.Command {
	var input, configzPath, nodeName, output string
	cmd := &cobra.Command{
		Use:   "enrich-node",
		Short: "Merge authenticated live kubelet configz facts into a matching native node report",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportData, err := os.ReadFile(input)
			if err != nil {
				return fmt.Errorf("read node report: %w", err)
			}
			report, err := audit.DecodeReport(reportData)
			if err != nil {
				return fmt.Errorf("decode node report: %w", err)
			}
			configz, err := os.ReadFile(configzPath)
			if err != nil {
				return fmt.Errorf("read authenticated configz snapshot: %w", err)
			}
			if err := audit.EnrichNodeWithConfigz(&report, nodeName, configz); err != nil {
				return err
			}
			if output == "" {
				return fmt.Errorf("output-file is required to avoid accidentally emitting raw evidence")
			}
			return writeReport("json", output, report)
		},
	}
	cmd.Flags().StringVar(&input, "input", "", "Matching node-local JSON report")
	cmd.Flags().StringVar(&configzPath, "configz-file", "", "Fresh live Kubernetes node/proxy/configz JSON")
	cmd.Flags().StringVar(&nodeName, "node-name", "", "Kubernetes node name, must match report")
	cmd.Flags().StringVar(&output, "output-file", "", "Enhanced node JSON report")
	_ = cmd.MarkFlagRequired("input")
	_ = cmd.MarkFlagRequired("configz-file")
	_ = cmd.MarkFlagRequired("node-name")
	_ = cmd.MarkFlagRequired("output-file")
	return cmd
}

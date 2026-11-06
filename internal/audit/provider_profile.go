package audit

import (
	"fmt"
	"strings"
)

const eksKubeletModeExtensionID = "KUBECIS-EKS-NODE-CONFIG-MODE"

// applyProviderProfile creates a copy of the provisional profile with
// provider-specific assessments. The generic Kubernetes 4.1.9 requirement
// cannot be reused as an official EKS control: the AL2023 nodeadm kubelet
// config lives at a different path and the EKS permission guidance differs.
// This is an explicitly non-CIS-scored KubeCIS extension until an independently
// verified, versioned EKS CIS mapping is available.
func applyProviderProfile(input Benchmark, provider string) Benchmark {
	if provider != "eks" {
		return input
	}
	out := input
	out.Checks = append([]Check(nil), input.Checks...)
	for i, c := range out.Checks {
		if c.Target != "node" || c.Kind != "host_file_mode" ||
			c.Path != "var/lib/kubelet/config.yaml" {
			continue
		}
		c.ID = eksKubeletModeExtensionID
		c.Title = "Inspect EKS AL2023 nodeadm kubelet configuration permissions"
		c.Description = "KubeCIS EKS extension (not an official CIS score): inspect allowlisted /etc/kubernetes/kubelet/config.json file mode; no broader than 0644. This does not verify file ownership or confirm effective kubelet flags."
		c.MaxMode = "0644"
		c.Scored = false
		c.Remediation = "Review EKS nodeadm kubelet configuration file ownership and permissions against the applicable CIS EKS benchmark and AMI guidance; do not change AWS-managed node files solely to satisfy a generic Kubernetes control."
		out.Checks[i] = c
	}
	return out
}

// validateBenchmarkMinor rejects explicit benchmark selections that silently
// contradict the Kubernetes server minor. A provisional profile is still
// provisional even after matching a listed minor; no CIS certification follows.
func validateBenchmarkMinor(b Benchmark, minor string) error {
	if strings.TrimSpace(minor) == "" {
		return fmt.Errorf("Kubernetes minor version cannot be empty")
	}
	for _, m := range b.KubernetesMinor {
		if m == minor {
			return nil
		}
	}
	return fmt.Errorf("benchmark %q does not declare Kubernetes minor %s (listed: %s); select a matching profile or add an explicitly provisional mapping", b.ID, minor, strings.Join(b.KubernetesMinor, ","))
}

// benchmarkCompatibilityNote names the detected version and provider without
// suggesting that a partial profile is an independently validated CIS mapping.
func benchmarkCompatibilityNote(b Benchmark, gitVersion, provider string) string {
	note := strings.TrimSpace(b.CompatibilityNote)
	if note != "" {
		note += " "
	}
	return fmt.Sprintf("%sDetected server=%s, provider=%s. Matching this starter profile does not validate official CIS EKS control IDs or applicability.", note, gitVersion, provider)
}

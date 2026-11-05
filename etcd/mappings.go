package etcd

import "fmt"

// DefaultMappings are deliberately conservative examples. Validate exact applicability and
// control identifiers against the pinned upstream benchmark before enabling in a release profile.
func DefaultMappings() []Mapping {
	return []Mapping{
		{
			ControlID: "2.1", Title: "Ensure that the --client-cert-auth argument is set to true",
			Source: SourceFlag, Key: "--client-cert-auth", Expected: "true",
			Remediation: "Configure etcd client certificate authentication and verify the trusted CA configuration.",
		},
		{
			ControlID: "2.2", Title: "Ensure that the --auto-tls argument is not set to true",
			Source: SourceFlag, Key: "--auto-tls", Expected: "false",
			Remediation: "Disable etcd auto-TLS and configure certificates issued by the cluster's trusted CA.",
		},
		{
			ControlID: "2.3", Title: "Ensure that the --peer-client-cert-auth argument is set to true",
			Source: SourceFlag, Key: "--peer-client-cert-auth", Expected: "true",
			Remediation: "Enable peer client certificate authentication and configure the peer trusted CA.",
		},
		{
			ControlID: "2.4", Title: "Ensure that the --peer-auto-tls argument is not set to true",
			Source: SourceFlag, Key: "--peer-auto-tls", Expected: "false",
			Remediation: "Disable peer auto-TLS and use explicitly managed peer certificates.",
		},
		{
			ControlID: "2.5", Title: "Ensure that the --data-dir argument is set",
			Source: SourceFlag, Key: "--data-dir", Expected: "__PRESENT__",
			Remediation: "Set an explicit etcd data directory on appropriately protected storage.",
		},
		{
			ControlID: "2.6", Title: "Ensure that the --client-cert-auth argument is enabled in parsed configuration",
			Source: SourceConfig, Key: "client-transport-security.client-cert-auth", Expected: "true",
			Remediation: "Set client-transport-security.client-cert-auth to true in the etcd configuration.",
		},
	}
}

func ValidateMappings(m []Mapping) error {
	seen := map[string]bool{}
	for i, x := range m {
		if x.ControlID == "" || x.Title == "" || x.Key == "" || x.Expected == "" {
			return fmt.Errorf("mapping %d: control ID, title, key and expected are required", i)
		}
		if x.Source != SourceFlag && x.Source != SourceConfig {
			return fmt.Errorf("mapping %s: unsupported source %q", x.ControlID, x.Source)
		}
		if x.Sensitive {
			return fmt.Errorf("mapping %s: sensitive evidence mappings are prohibited", x.ControlID)
		}
		if seen[x.ControlID] {
			return fmt.Errorf("duplicate control mapping %s", x.ControlID)
		}
		seen[x.ControlID] = true
	}
	return nil
}

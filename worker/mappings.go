package worker

import "fmt"

// DefaultMappings are conservative starter examples. Confirm exact IDs and applicability
// against the pinned CIS benchmark and distribution profile before enabling.
func DefaultMappings() []Mapping {
	return []Mapping{
		{
			ControlID: "4.2.1", Title: "Ensure that the --anonymous-auth argument is set to false",
			Component: ComponentKubelet, Source: SourceFlag, Key: "--anonymous-auth", Expected: "false",
			Remediation: "Set kubelet --anonymous-auth=false, or the equivalent supported kubelet configuration.",
		},
		{
			ControlID: "4.2.2", Title: "Ensure that the --authorization-mode argument is not set to AlwaysAllow",
			Component: ComponentKubelet, Source: SourceFlag, Key: "--authorization-mode", Expected: "Webhook",
			Remediation: "Configure kubelet authorization mode to Webhook and verify API-server authorization integration.",
		},
		{
			ControlID: "4.2.3", Title: "Ensure that the --client-ca-file argument is set",
			Component: ComponentKubelet, Source: SourceFlag, Key: "--client-ca-file", Expected: "__PRESENT__",
			Remediation: "Configure the kubelet client CA file using the distribution-supported configuration mechanism.",
		},
		{
			ControlID: "4.2.4", Title: "Ensure that the --read-only-port argument is set to 0",
			Component: ComponentKubelet, Source: SourceFlag, Key: "--read-only-port", Expected: "0",
			Remediation: "Set kubelet --read-only-port=0.",
		},
		{
			ControlID: "4.2.6", Title: "Ensure that the --protect-kernel-defaults argument is set to true",
			Component: ComponentKubelet, Source: SourceFlag, Key: "--protect-kernel-defaults", Expected: "true",
			Remediation: "Set kubelet --protect-kernel-defaults=true after validating required node kernel settings.",
		},
		{
			ControlID: "4.2.7", Title: "Ensure that the --make-iptables-util-chains argument is set to true",
			Component: ComponentKubelet, Source: SourceFlag, Key: "--make-iptables-util-chains", Expected: "true",
			Remediation: "Set kubelet --make-iptables-util-chains=true where required by the cluster networking implementation.",
		},
		{
			ControlID: "4.2.1-config", Title: "Kubelet configuration anonymous authentication is disabled",
			Component: ComponentKubelet, Source: SourceConfig, Key: "authentication.anonymous.enabled", Expected: "false",
			Remediation: "Set authentication.anonymous.enabled to false in the effective kubelet configuration.",
		},
		{
			ControlID: "4.2.2-config", Title: "Kubelet configuration authorization mode uses Webhook",
			Component: ComponentKubelet, Source: SourceConfig, Key: "authorization.mode", Expected: "Webhook",
			Remediation: "Set authorization.mode to Webhook in the effective kubelet configuration.",
		},
		{
			ControlID: "4.2.4-config", Title: "Kubelet configuration read-only port is disabled",
			Component: ComponentKubelet, Source: SourceConfig, Key: "readOnlyPort", Expected: "0",
			Remediation: "Set readOnlyPort to 0 in the effective kubelet configuration.",
		},
		{
			ControlID: "4.2.5-proxy", Title: "Kube-proxy metrics bind address is not exposed on all interfaces",
			Component: ComponentKubeProxy, Source: SourceFlag, Key: "--metrics-bind-address", Expected: "127.0.0.1:10249",
			Remediation: "Bind kube-proxy metrics to loopback or an explicitly approved restricted address.",
		},
	}
}

func ValidateMappings(mappings []Mapping) error {
	seen := make(map[string]struct{}, len(mappings))
	for i, m := range mappings {
		if m.ControlID == "" || m.Title == "" || m.Component == "" || m.Key == "" || m.Expected == "" {
			return fmt.Errorf("mapping %d: control ID, title, component, key and expected are required", i)
		}
		if m.Component != ComponentKubelet && m.Component != ComponentKubeProxy {
			return fmt.Errorf("mapping %s: unsupported component %q", m.ControlID, m.Component)
		}
		if m.Source != SourceFlag && m.Source != SourceConfig {
			return fmt.Errorf("mapping %s: unsupported source %q", m.ControlID, m.Source)
		}
		identity := string(m.Component) + "|" + m.ControlID
		if _, ok := seen[identity]; ok {
			return fmt.Errorf("duplicate mapping %s", identity)
		}
		seen[identity] = struct{}{}
	}
	return nil
}

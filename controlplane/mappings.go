package controlplane

import "fmt"

// DefaultMappings contains conservative examples of common control-plane hardening
// requirements. These are evaluator mappings, not a complete CIS control catalogue.
// Confirm exact control IDs, benchmark versions, and applicability before enabling.
func DefaultMappings() []Mapping {
	return []Mapping{
		{
			ControlID: "1.2.1", Title: "Ensure that the --anonymous-auth argument is set to false",
			Component: ComponentAPIServer, Source: SourceFlag, Key: "--anonymous-auth",
			ExpectedValue: "false", Remediation: "Set --anonymous-auth=false in the API server configuration.",
		},
		{
			ControlID: "1.2.2", Title: "Ensure that the --token-auth-file argument is not set",
			Component: ComponentAPIServer, Source: SourceFlag, Key: "--token-auth-file",
			ExpectedValue: "__ABSENT__", Remediation: "Remove --token-auth-file unless an explicitly approved exception applies.",
		},
		{
			ControlID: "1.2.6", Title: "Ensure that the --kubelet-certificate-authority argument is set",
			Component: ComponentAPIServer, Source: SourceFlag, Key: "--kubelet-certificate-authority",
			ExpectedValue: "__PRESENT__", Remediation: "Configure --kubelet-certificate-authority with the trusted kubelet CA bundle.",
		},
		{
			ControlID: "1.3.2", Title: "Ensure that the --profiling argument is set to false",
			Component: ComponentControllerManager, Source: SourceFlag, Key: "--profiling",
			ExpectedValue: "false", Remediation: "Set --profiling=false in the controller-manager configuration.",
		},
		{
			ControlID: "1.3.6", Title: "Ensure that the -- RotateKubeletServerCertificate argument is enabled",
			Component: ComponentControllerManager, Source: SourceFlag, Key: "--feature-gates",
			ExpectedValue: "RotateKubeletServerCertificate=true", AllowMultiple: true,
			Remediation: "Enable RotateKubeletServerCertificate where supported by the Kubernetes version.",
		},
		{
			ControlID: "1.4.1", Title: "Ensure that the --profiling argument is set to false",
			Component: ComponentScheduler, Source: SourceFlag, Key: "--profiling",
			ExpectedValue: "false", Remediation: "Set --profiling=false in the scheduler configuration.",
		},
	}
}

func ValidateMappings(mappings []Mapping) error {
	seen := make(map[string]struct{}, len(mappings))
	for i, m := range mappings {
		if m.ControlID == "" || m.Title == "" || m.Component == "" || m.Key == "" || m.ExpectedValue == "" {
			return fmt.Errorf("mapping %d: control ID, title, component, key and expected value are required", i)
		}
		if m.Source != SourceFlag && m.Source != SourceConfig {
			return fmt.Errorf("mapping %s: unsupported source %q", m.ControlID, m.Source)
		}
		switch m.Component {
		case ComponentAPIServer, ComponentControllerManager, ComponentScheduler:
		default:
			return fmt.Errorf("mapping %s: unsupported component %q", m.ControlID, m.Component)
		}
		identity := string(m.Component) + "|" + m.ControlID
		if _, ok := seen[identity]; ok {
			return fmt.Errorf("duplicate mapping for %s", identity)
		}
		seen[identity] = struct{}{}
	}
	return nil
}

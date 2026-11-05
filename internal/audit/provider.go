package audit

import "strings"

// DetectCloudProvider is conservative. A Kubernetes version without a
// recognizable provider signature is "unknown", not a guessed cloud platform.
func DetectCloudProvider(gitVersion string) string {
	v := strings.ToLower(strings.TrimSpace(gitVersion))
	switch {
	case strings.Contains(v, "-eks-"):
		return "eks"
	case strings.Contains(v, "-gke."):
		return "gke"
	case strings.Contains(v, "-aks"):
		return "aks"
	default:
		return "unknown"
	}
}

// annotateProviderAssessment preserves the underlying security result. It
// provides a separate responsibility/applicability interpretation so that
// missing managed-host evidence is never represented as a compliant PASS.
func annotateProviderAssessment(provider string, check Check, result Result) Result {
	if provider != "eks" {
		return result
	}
	switch {
	case check.Target == "master" && (strings.HasPrefix(check.Kind, "host_") || check.Kind == "manual"):
		result.Applicability = "AWS_MANAGED_CONTROL_PLANE"
		result.EvidenceSource = "aws-provider-evidence"
		if strings.HasPrefix(check.Kind, "host_") {
			result.Status = "NOT_ASSESSED"
			result.Details = "Amazon EKS manages the control-plane hosts; their filesystem is not exposed to customers. This host check cannot be evaluated through the Kubernetes API. Review AWS-managed control-plane evidence and the applicable EKS benchmark; no PASS or FAIL is inferred."
			result.Evidence = []Evidence{{Resource: "EKSControlPlane", Name: check.ID, Field: "hostAccess", Observed: "AWS-managed; host not customer-accessible", Expected: "provider evidence or an EKS-applicable control"}}
		}
	case check.Target == "node" && strings.HasPrefix(check.Kind, "host_"):
		result.Applicability = "NODE_AUDIT_REQUIRED"
		result.EvidenceSource = "linux-node-metadata"
		result.Status = "NOT_ASSESSED"
		result.Details = "This cluster-API scan cannot inspect EC2 worker-node host files. Deploy the read-only Linux node metadata auditor on supported EKS EC2 nodes and consolidate its report; no PASS or FAIL is inferred."
		result.Evidence = []Evidence{{Resource: "EKSWorkerNode", Name: check.ID, Field: "hostAccess", Observed: "not collected by cluster API scanner", Expected: "read-only per-node Linux host metadata"}}
	case check.Target == "master":
		result.Applicability = "PROVIDER_EVIDENCE_REQUIRED"
		result.EvidenceSource = "aws-provider-evidence"
	case check.Kind == "api_control_plane_security" || check.Kind == "api_audit_logging":
		result.Applicability = "PROVIDER_EVIDENCE_REQUIRED"
		result.EvidenceSource = "kubernetes-api-and-aws-provider-evidence"
	case check.Target == "policies":
		result.Applicability = "KUBERNETES_API"
		result.EvidenceSource = "kubernetes-api"
	default:
		result.Applicability = "REVIEW_REQUIRED"
	}
	return result
}

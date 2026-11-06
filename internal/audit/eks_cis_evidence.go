package audit

import (
    "fmt"
    "net/netip"
    "strings"
)

// EvaluateEKSCISProviderControls interprets only the exact AWS DescribeCluster
// evidence needed by three automated controls in CIS Amazon EKS v2.1.0.
// It does not evaluate any other recommendation and does not certify benchmark
// coverage. Source: user-supplied official CIS Amazon EKS v2.1.0 PDF.
func EvaluateEKSCISProviderControls(e EKSProviderEvidence) ([]Result, error) {
    if e.SchemaVersion != 1 || e.CloudProvider != "eks" || e.AssessmentStatus != "INVENTORY_ONLY" {
        return nil, fmt.Errorf("invalid AWS EKS inventory for CIS provider controls")
    }
    for _, name := range []string{"api", "audit", "authenticator", "controllerManager", "scheduler"} {
        if _, ok := e.ControlPlaneLogging[name]; !ok {
            return nil, fmt.Errorf("missing EKS control plane log type %q", name)
        }
    }
    results := make([]Result, 0, 3)
    logging := Result{
        ID: "2.1.1", Title: "EKS control plane logging", Target: "policies",
        Status: "PASS", Method: "aws_eks_describe_cluster", Scored: true,
        Details: "All five EKS control-plane logging types are enabled.",
        Remediation: "Enable EKS control plane logging for api, audit, authenticator, controllerManager and scheduler.",
    }
    for _, name := range []string{"api", "audit", "authenticator", "controllerManager", "scheduler"} {
        observed := "disabled"
        if e.ControlPlaneLogging[name] {
            observed = "enabled"
        } else {
            logging.Status = "FAIL"
            logging.Details = "One or more required control-plane logging types are disabled."
        }
        logging.Evidence = append(logging.Evidence, Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "logging."+name, Observed: observed, Expected: "enabled"})
    }
    results = append(results, logging)

    endpoint := Result{
        ID: "5.4.1", Title: "Restrict EKS control plane endpoint with private access and public CIDR allowlists",
        Target: "policies", Status: "FAIL", Scored: true, Method: "aws_eks_describe_cluster",
        Remediation: "Enable private API endpoint access and limit any public endpoint to explicitly approved source CIDRs.",
    }
    if e.EndpointPrivate {
        if !e.EndpointPublic {
            endpoint.Status = "PASS"
            endpoint.Details = "Private endpoint enabled and public endpoint disabled."
        } else {
            endpoint.Status = "MANUAL"
            endpoint.Details = "Private endpoint enabled; public CIDRs need verification against the organization's approved source ranges."
            if len(e.PublicAccessCIDRs) == 0 {
                endpoint.Status = "FAIL"
                endpoint.Details = "Public endpoint enabled without an explicit CIDR allowlist."
            }
            for _, raw := range e.PublicAccessCIDRs {
                parsed, err := netip.ParsePrefix(strings.TrimSpace(raw))
                if err != nil {
                    endpoint.Status = "NOT_ASSESSED"
                    endpoint.Details = "Invalid public CIDR evidence; cannot verify endpoint restriction."
                    break
                }
                if parsed.Bits() == 0 {
                    endpoint.Status = "FAIL"
                    endpoint.Details = "Public endpoint allows traffic from the entire IPv4 or IPv6 internet."
                    break
                }
            }
        }
    } else {
        endpoint.Details = "Private EKS API endpoint access is disabled."
    }
    privateObserved, publicObserved := "false", "false"
    if e.EndpointPrivate { privateObserved = "true" }
    if e.EndpointPublic { publicObserved = "true" }
    endpoint.Evidence = append(endpoint.Evidence,
        Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "endpointPrivateAccess", Observed: privateObserved, Expected: "true"},
        Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "endpointPublicAccess", Observed: publicObserved, Expected: "false or restricted to approved CIDRs"},
    )
    for _, cidr := range e.PublicAccessCIDRs {
        endpoint.Evidence = append(endpoint.Evidence, Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "publicAccessCidrs", Observed: cidr, Expected: "approved restricted CIDR"})
    }
    results = append(results, endpoint)

    privateOnly := Result{
        ID: "5.4.2", Title: "Private-only EKS control plane access",
        Target: "policies", Status: "FAIL", Scored: true, Method: "aws_eks_describe_cluster",
        Remediation: "Enable private endpoint access and disable the public API endpoint after confirming private administration connectivity.",
    }
    if e.EndpointPrivate && !e.EndpointPublic {
        privateOnly.Status = "PASS"
        privateOnly.Details = "Private endpoint is enabled and public endpoint is disabled."
    } else {
        privateOnly.Details = "Private-only control requires endpointPrivateAccess=true and endpointPublicAccess=false."
    }
    privateOnly.Evidence = append(privateOnly.Evidence,
        Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "endpointPrivateAccess", Observed: privateObserved, Expected: "true"},
        Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "endpointPublicAccess", Observed: publicObserved, Expected: "false"},
    )
    results = append(results, privateOnly)
    return results, nil
}

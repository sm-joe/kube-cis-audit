package audit

import (
	"fmt"
	"strings"
)

// EvaluateEKSAdditionalAWSControls evaluates the read-only AWS evidence
// separately from Kubernetes API findings. Missing IAM permissions or
// third-party scan attestation never becomes a successful assessment.
func EvaluateEKSAdditionalAWSControls(e EKSProviderEvidence) []Result {
	var results []Result
	if e.AccessInventoryStatus != "" {
		for _, id := range []string{"4.1.7", "5.5.1"} {
			title := "Scoped EKS access entries and IAM principal authorization"
			if id == "5.5.1" {
				title = "Use EKS access entries for IAM principal access"
			}
			r := Result{ID: id, Title: title, Target: "policies", Status: "NOT_ASSESSED", Method: "aws_eks_access_entries", Scored: true, EvidenceSource: "aws-eks-access-entries", Remediation: "Migrate approved IAM principal mappings to EKS access entries; review EKS access policies, Kubernetes groups, scopes and deprecated aws-auth mappings."}
			if e.AuthenticationMode != "" {
				r.Evidence = append(r.Evidence, Evidence{Resource: "AWS::EKS::Cluster", Name: e.ClusterARN, Field: "accessConfig.authenticationMode", Observed: e.AuthenticationMode, Expected: "API or API_AND_CONFIG_MAP"})
			}
			for _, entry := range e.AccessEntries {
				r.Evidence = append(r.Evidence, Evidence{Resource: "AWS::EKS::AccessEntry", Name: entry.PrincipalARN, Field: "authorization", Observed: fmt.Sprintf("type=%s groups=%d scopes=%v error=%s", entry.Type, len(entry.Groups), entry.PolicyScopes, entry.Error), Expected: "approved IAM principal and least-privilege namespace/cluster scope"})
			}
			switch {
			case e.AuthenticationMode == "CONFIG_MAP":
				r.Status = "FAIL"
				r.Details = "EKS authentication mode relies exclusively on legacy aws-auth ConfigMap. No EKS access-entry authentication path is active."
			case e.AuthenticationMode != "API" && e.AuthenticationMode != "API_AND_CONFIG_MAP":
				r.Details = "EKS authentication mode not verified as API or API_AND_CONFIG_MAP."
			case e.AccessInventoryStatus != "COLLECTED":
				r.Details = "Access entries or associated policy scopes could not be fully inventoried; verify eks:ListAccessEntries, eks:DescribeAccessEntry and eks:ListAssociatedAccessPolicies."
			case len(e.AccessEntries) == 0:
				r.Status = "MANUAL"
				r.Details = "Access entries API is enabled but no principal entries were returned; verify intended cluster administrator and bootstrap identities."
			default:
				r.Status = "MANUAL"
				r.Details = fmt.Sprintf("EKS authentication mode %s with %d access entries and inventoried policy scopes; review approved principals, RBAC and any legacy aws-auth mappings.", e.AuthenticationMode, len(e.AccessEntries))
			}
			results = append(results, r)
		}
	}
	if e.ECRInventoryStatus != "" {
		r := Result{ID: "5.1.1", Title: "ECR or approved third-party image vulnerability scanning", Target: "policies", Status: "NOT_ASSESSED", Method: "aws_ecr_image_scanning", Scored: true, EvidenceSource: "aws-ecr-image-scan", Remediation: "Enable ECR Basic scan-on-push or Enhanced continuous scanning for all deployed ECR images, confirm image digest scan status and remediate findings; approve third-party coverage separately."}
		successful, problems := 0, 0
		for _, item := range e.ECRImages {
			observed := item.ScanType + "/" + item.ScanStatus
			if item.Error != "" {
				observed = "unavailable: " + item.Error
				problems++
			} else if item.Digest == "" || (item.ScanStatus != "COMPLETE" && item.ScanStatus != "ACTIVE") {
				problems++
			} else {
				successful++
			}
			r.Evidence = append(r.Evidence, Evidence{Resource: "AWS::ECR::Image", Name: item.Registry + "/" + item.Repository, Field: "scanStatus", Observed: observed, Expected: "verified vulnerability scan for deployed image digest"})
		}
		switch {
		case e.ECRInventoryStatus != "COLLECTED":
			r.Details = "Could not complete the deployed image inventory and ECR scanning checks; permissions or image digest evidence unavailable."
		case len(e.ECRImages) == 0 || e.NonECRImageCount > 0:
			r.Status = "MANUAL"
			r.Details = fmt.Sprintf("%d ECR images and %d non-ECR image references; verify approved third-party scanning of non-ECR images.", len(e.ECRImages), e.NonECRImageCount)
		case problems > 0:
			r.Status = "MANUAL"
			r.Details = fmt.Sprintf("%d of %d ECR images lack verified scan results; check scan eligibility, registry rules and alternate approved scanner evidence.", problems, len(e.ECRImages))
		default:
			r.Status = "PASS"
			r.Details = fmt.Sprintf("%d of %d deployed ECR image digests have confirmed ECR scan status; review vulnerability findings separately.", successful, len(e.ECRImages))
		}
		results = append(results, r)
	}
	return results
}

// ApplyEKSCNIProviderEvidence finalizes the VPC CNI policy only when managed
// add-on configuration and Kubernetes node-agent readiness agree.
func ApplyEKSCNIProviderEvidence(out *ConsolidatedReport, e EKSProviderEvidence) {
	if e.VpcCNIStatus == "" {
		return
	}
	for i := range out.Results {
		r := &out.Results[i]
		if r.ID != "4.3.1" || r.Target != "policies" {
			continue
		}
		r.Evidence = append(r.Evidence, Evidence{Resource: "AWS::EKS::Addon", Name: "vpc-cni", Field: "enableNetworkPolicy", Observed: fmt.Sprintf("addonStatus=%s version=%s configured=%v", e.VpcCNIStatus, e.VpcCNIVersion, e.VpcCNINetworkPolicy), Expected: "ACTIVE supported add-on with enableNetworkPolicy=true and node policy agents ready"})
		switch {
		case e.VpcCNIStatus != "ACTIVE" || e.VpcCNINetworkPolicy == nil:
			r.Status = "MANUAL"
			r.Details = "Managed VPC CNI network policy configuration is unavailable or add-on is not ACTIVE; review self-managed CNI and alternative enforcement."
		case !*e.VpcCNINetworkPolicy:
			r.Status = "FAIL"
			r.Details = "Managed EKS vpc-cni add-on explicitly has enableNetworkPolicy=false; no VPC CNI policy enforcement is configured."
		case strings.Contains(r.Details, "ready network policy agent"):
			r.Status = "PASS"
			r.Details = "EKS managed VPC CNI enables network policy and the aws-node network policy agent is ready. Workload policy coverage requires separate manual control 4.3.2."
		default:
			r.Status = "MANUAL"
			r.Details = "EKS managed VPC CNI enables network policy, but live aws-node agent readiness is not fully verified."
		}
		return
	}
}

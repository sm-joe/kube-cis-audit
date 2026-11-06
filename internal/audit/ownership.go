package audit

import "strings"

const (
	ownershipPlatformNamed = "PLATFORM_NAMED_REVIEW"
	ownershipAuditorSelf   = "KUBECIS_SELF_REVIEW"
	ownershipUnverified    = "UNVERIFIED"
)

// annotateEKSOwnershipHints adds triage context to evidence without changing
// the observed security outcome, controls, or scores. Names are hints, NOT
// proof that AWS manages a Pod or that a risky permission is acceptable.
func annotateEKSOwnershipHints(check Check, result Result) Result {
	if check.Target != "policies" || len(result.Evidence) == 0 {
		return result
	}
	foundPlatform, foundUnknown := false, false
	for i := range result.Evidence {
		e := &result.Evidence[i]
		name := e.Name
		if strings.Contains(name, "/") {
			name = strings.SplitN(name, "/", 2)[0]
		}
		switch {
		case e.Namespace == "kube-system" && (strings.HasPrefix(name, "aws-node-") ||
			strings.HasPrefix(name, "kube-proxy-") ||
			strings.HasPrefix(name, "coredns-")):
			e.OwnershipHint = ownershipPlatformNamed
			foundPlatform = true
		case e.Namespace == "kube-system" && strings.HasPrefix(name, "kubecis-node-auditor-"):
			e.OwnershipHint = ownershipAuditorSelf
			foundPlatform = true
		case e.Resource == "ClusterRole" || e.Resource == "Role" ||
			e.Resource == "ClusterRoleBinding" || e.Resource == "RoleBinding":
			if strings.HasPrefix(e.Name, "system:") || strings.HasPrefix(e.Name, "eks:") {
				e.OwnershipHint = ownershipPlatformNamed
				foundPlatform = true
			} else {
				e.OwnershipHint = ownershipUnverified
				foundUnknown = true
			}
		case e.Namespace != "" && e.Name != "":
			e.OwnershipHint = ownershipUnverified
			foundUnknown = true
		}
	}
	switch {
	case foundPlatform && foundUnknown:
		result.OwnershipReview = "MIXED_PLATFORM_NAMED_AND_UNVERIFIED"
	case foundPlatform:
		result.OwnershipReview = "PLATFORM_NAMED_EVIDENCE_REVIEW"
	case foundUnknown:
		result.OwnershipReview = "UNVERIFIED_OWNERSHIP"
	}
	return result
}

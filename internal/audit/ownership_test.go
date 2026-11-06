package audit

import "testing"

func TestEKSPlatformHintsDoNotDowngradeFindings(t *testing.T) {
	check := Check{ID: "KUBECIS-2.4.6", Target: "policies", Kind: "api_pod_hostpath"}
	in := Result{ID: check.ID, Status: "FAIL", Scored: false, Details: "sensitive writable host mount", Evidence: []Evidence{
		{Resource: "container", Namespace: "kube-system", Name: "aws-node-123/aws-node", Field: "volumeMount.readOnly", Observed: "false"},
		{Resource: "container", Namespace: "customer-workloads", Name: "frontend/api", Field: "volumeMount.readOnly", Observed: "false"},
	}}
	got := annotateProviderAssessment("eks", check, in)
	if got.Status != "FAIL" || got.Details != in.Details || got.Scored != in.Scored {
		t.Fatalf("risk status was suppressed: %+v", got)
	}
	if got.OwnershipReview != "MIXED_PLATFORM_NAMED_AND_UNVERIFIED" ||
		got.Evidence[0].OwnershipHint != ownershipPlatformNamed ||
		got.Evidence[1].OwnershipHint != ownershipUnverified {
		t.Fatalf("missing evidence ownership triage: %+v", got)
	}
}

func TestEKSSystemNamesAreNotTreatedAsPass(t *testing.T) {
	check := Check{ID: "KUBECIS-2.4.5", Target: "policies", Kind: "api_pod_host_namespaces"}
	in := Result{Status: "FAIL", Evidence: []Evidence{
		{Resource: "Pod", Namespace: "kube-system", Name: "kube-proxy-aaa", Observed: "hostNetwork=true"},
		{Resource: "Pod", Namespace: "kube-system", Name: "kubecis-node-auditor-123", Observed: "root metadata helper"},
	}}
	got := annotateProviderAssessment("eks", check, in)
	if got.Status != "FAIL" || got.OwnershipReview != "PLATFORM_NAMED_EVIDENCE_REVIEW" {
		t.Fatalf("system role misinterpreted: %+v", got)
	}
	if got.Evidence[0].OwnershipHint != ownershipPlatformNamed || got.Evidence[1].OwnershipHint != ownershipAuditorSelf {
		t.Fatalf("wrong system evidence classification: %+v", got.Evidence)
	}
}

func TestUnknownProviderDoesNotClaimSystemOwnership(t *testing.T) {
	c := Check{Target: "policies", Kind: "api_pod_hostpath"}
	r := Result{Status: "FAIL", Evidence: []Evidence{{Resource: "Pod", Namespace: "kube-system", Name: "aws-node-a"}}}
	got := annotateProviderAssessment("unknown", c, r)
	if got.OwnershipReview != "" || got.Evidence[0].OwnershipHint != "" {
		t.Fatalf("provider hint leaked into non-EKS report: %+v", got)
	}
}

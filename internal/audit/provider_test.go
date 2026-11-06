package audit

import (
	"strings"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestDetectCloudProvider(t *testing.T) {
	tests := []struct{ version, want string }{
		{"v1.37.1-eks-577a51e", "eks"},
		{"v1.32.0-eks-1234567", "eks"},
		{"v1.32.0", "unknown"},
		{"v1.30.0-gke.123", "gke"},
		{"v1.31.0-aks", "aks"},
	}
	for _, tt := range tests {
		if got := DetectCloudProvider(tt.version); got != tt.want {
			t.Errorf("DetectCloudProvider(%q) = %q; want %q", tt.version, got, tt.want)
		}
	}
}

func TestEKSHostApplicabilityNeverPasses(t *testing.T) {
	tests := []struct{ target, want string }{
		{"master", "AWS_MANAGED_CONTROL_PLANE"},
		{"node", "NODE_AUDIT_REQUIRED"},
	}
	for _, tt := range tests {
		c := Check{ID: "test", Kind: "host_file_mode", Target: tt.target, Scored: true}
		r := annotateProviderAssessment("eks", c, Result{Status: "PASS", Scored: true})
		if r.Status != "NOT_ASSESSED" || r.Applicability != tt.want || r.EvidenceSource == "" {
			t.Errorf("%s host result = %#v", tt.target, r)
		}
		if !strings.Contains(r.Details, "no PASS or FAIL is inferred") {
			t.Errorf("missing explicit evidence limitation: %q", r.Details)
		}
	}
}

func TestKindHostResultsAreNotReclassified(t *testing.T) {
	r := Result{Status: "PASS", Details: "original"}
	got := annotateProviderAssessment("unknown", Check{Target: "node", Kind: "host_file_mode"}, r)
	if got.Status != r.Status || got.Details != r.Details || got.Applicability != "" {
		t.Fatalf("Kind result changed: %#v", got)
	}
}

func TestEKSAPIResultRetainsStatus(t *testing.T) {
	r := annotateProviderAssessment("eks", Check{Target: "policies", Kind: "api_rbac_risk"}, Result{Status: "FAIL", Details: "risky permission"})
	if r.Status != "FAIL" || r.Applicability != "KUBERNETES_API" {
		t.Fatalf("EKS API finding was incorrectly suppressed: %#v", r)
	}
}

func TestPlatformNamedRBACBindingIsOnlyAnOwnershipHint(t *testing.T) {
	tests := []struct {
		name    string
		binding string
		role    string
		subject string
		want    bool
	}{
		{"EKS binding", "eks:addon-cluster-admin", "cluster-admin", "eks:addon-manager", true},
		{"controller role", "workload", "system:controller:deployment-controller", "deployment", true},
		{"customer", "app-prod", "cluster-admin", "app-admin", false},
	}
	for _, tt := range tests {
		got := platformNamedRBACBinding(tt.binding, rbacv1.RoleRef{Kind: "ClusterRole", Name: tt.role}, []rbacv1.Subject{{Kind: "User", Name: tt.subject}})
		if got != tt.want {
			t.Fatalf("%s: got %t, want %t", tt.name, got, tt.want)
		}
	}
}

func TestEKSProviderEvidenceForAPILimitations(t *testing.T) {
	for _, kind := range []string{"api_control_plane_security", "api_audit_logging"} {
		r := annotateProviderAssessment("eks", Check{Target: "policies", Kind: kind}, Result{Status: "MANUAL"})
		if r.Status != "MANUAL" || r.Applicability != "PROVIDER_EVIDENCE_REQUIRED" {
			t.Fatalf("%s: %#v", kind, r)
		}
	}
}

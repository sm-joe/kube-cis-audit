package audit

import (
	"context"
	"fmt"
	"strings"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestEKSBoundRBACGrantsAndNamespaceScope(t *testing.T) {
	rules := []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"secrets"}, Verbs: []string{"get", "list"}},
		{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}},
		{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"roles"}, Verbs: []string{"bind", "escalate"}},
		{APIGroups: []string{""}, Resources: []string{"nodes/proxy"}, Verbs: []string{"get"}},
		{APIGroups: []string{"admissionregistration.k8s.io"}, Resources: []string{"mutatingwebhookconfigurations"}, Verbs: []string{"patch"}},
		{APIGroups: []string{""}, Resources: []string{"serviceaccounts/token"}, Verbs: []string{"create"}},
	}
	role := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "aggregated-permissions"}, Rules: rules}
	binding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "dev-use", Namespace: "dev"},
		RoleRef:  rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "aggregated-permissions"},
		Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "app", Namespace: "dev"}}}
	client := fake.NewSimpleClientset(role, binding)
	for _, id := range []string{"4.1.2", "4.1.3", "4.1.8", "4.1.10", "4.1.11", "4.1.12"} {
		t.Run(id, func(t *testing.T) {
			result := EvaluateEKSManualEvidence(context.Background(), client, Check{ID: id, Target: "policies", Title: id})
			if result.Status != "MANUAL" || len(result.Evidence) == 0 {
				t.Fatalf("missing bound evidence: %+v", result)
			}
			found := false
			for _, e := range result.Evidence {
				if e.Resource == "RoleBinding" && e.Namespace == "dev" &&
					strings.Contains(e.Observed, "scope=namespace:dev") && strings.Contains(e.Observed, "ServiceAccount:dev:app") {
					found = true
				}
			}
			if !found {
				t.Fatalf("bound namespaced subject/scope absent: %+v", result.Evidence)
			}
		})
	}
}

func TestEKSBoundRBACUnboundRoleIsNotAnEffectiveGrant(t *testing.T) {
	role := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused"},
		Rules: []rbacv1.PolicyRule{{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}}}}
	client := fake.NewSimpleClientset(role)
	for _, id := range []string{"4.1.2", "4.1.3", "4.1.8", "4.1.10", "4.1.11", "4.1.12"} {
		got := EvaluateEKSManualEvidence(context.Background(), client, Check{ID: id})
		if got.Status != "MANUAL" || len(got.Evidence) != 0 || !strings.Contains(got.Details, "0 sensitive grant(s)") {
			t.Fatalf("%s unbound role incorrectly reported as grant: %+v", id, got)
		}
	}
}

func TestEKSBoundRBACSensitiveResourceAndVerbMatching(t *testing.T) {
	cases := []struct {
		id   string
		rule rbacv1.PolicyRule
		want bool
	}{
		{"4.1.12", rbacv1.PolicyRule{APIGroups: []string{""}, Resources: []string{"serviceaccounts/token"}, Verbs: []string{"get"}}, false},
		{"4.1.12", rbacv1.PolicyRule{APIGroups: []string{""}, Resources: []string{"serviceaccounts/*"}, Verbs: []string{"create"}}, true},
		{"4.1.10", rbacv1.PolicyRule{APIGroups: []string{""}, Resources: []string{"nodes"}, Verbs: []string{"get"}}, false},
		{"4.1.10", rbacv1.PolicyRule{APIGroups: []string{""}, Resources: []string{"nodes/*"}, Verbs: []string{"get"}}, true},
		{"4.1.11", rbacv1.PolicyRule{APIGroups: []string{"apps"}, Resources: []string{"validatingwebhookconfigurations"}, Verbs: []string{"delete"}}, false},
		{"4.1.11", rbacv1.PolicyRule{APIGroups: []string{"admissionregistration.k8s.io"}, Resources: []string{"validatingwebhookconfigurations"}, Verbs: []string{"delete"}}, true},
		{"4.1.2", rbacv1.PolicyRule{APIGroups: []string{"apps"}, Resources: []string{"secrets"}, Verbs: []string{"get"}}, false},
		{"4.1.2", rbacv1.PolicyRule{APIGroups: []string{""}, Resources: []string{"secrets"}, Verbs: []string{"watch"}}, true},
		{"4.1.8", rbacv1.PolicyRule{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"roles"}, Verbs: []string{"bind"}}, true},
		{"4.1.8", rbacv1.PolicyRule{APIGroups: []string{""}, Resources: []string{"users"}, Verbs: []string{"impersonate"}}, true},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			if got := eksRBACSensitiveRule(tc.id, tc.rule); got != tc.want {
				t.Errorf("%s: got %t want %t", tc.id, got, tc.want)
			}
		})
	}
}

func TestEKSBoundRBACIncompleteEvidenceNeverPass(t *testing.T) {
	binding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "broken"},
		RoleRef:  rbacv1.RoleRef{Kind: "ClusterRole", Name: "missing"},
		Subjects: []rbacv1.Subject{{Kind: "User", Name: "operator"}}}
	client := fake.NewSimpleClientset(binding)
	got := EvaluateEKSManualEvidence(context.Background(), client, Check{ID: "4.1.2"})
	if got.Status != "NOT_ASSESSED" || !strings.Contains(got.Details, "1 unresolved") {
		t.Fatalf("missing role must not pass: %+v", got)
	}
	denied := fake.NewSimpleClientset()
	denied.PrependReactor("list", "clusterroles", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("forbidden")
	})
	got = EvaluateEKSManualEvidence(context.Background(), denied, Check{ID: "4.1.2"})
	if got.Status != "NOT_ASSESSED" {
		t.Fatalf("denied inventory must not pass: %+v", got)
	}
}

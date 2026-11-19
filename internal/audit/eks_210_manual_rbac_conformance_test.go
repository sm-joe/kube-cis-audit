package audit

import (
	"context"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// '*' in RBAC resources/verbs/apiGroups includes the sensitive resources.
// All CIS manual controls retain MANUAL status, but the reviewer must see
// the bound subject and role rather than an unbound role inventory.
func TestEKS210ManualRBACWildcardInventoryFindsEverySensitivePermission(t *testing.T) {
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "unrestricted"},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}},
	}
	client := fake.NewSimpleClientset(role, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "explicit-test-grant"},
		RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", Name: "unrestricted"},
		Subjects: []rbacv1.Subject{{Kind: "User", Name: "test-subject"}},
	})
	for _, id := range []string{
		"4.1.2", "4.1.3", "4.1.4", "4.1.8", "4.1.9",
		"4.1.10", "4.1.11", "4.1.12",
	} {
		c := Check{ID: id, Target: "policies", Kind: "manual", Manual: true, Scored: false}
		r := EvaluateEKSManualEvidence(context.Background(), client, c)
		if r.Status != "MANUAL" || len(r.Evidence) == 0 {
			t.Fatalf("%s missed wildcard access requiring human RBAC review: %+v", id, r)
		}
	}
}

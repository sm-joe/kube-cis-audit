package audit

import (
	"context"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCISManualRBACWildcardMatchesRestrictedResourceChecks(t *testing.T) {
	client := fake.NewSimpleClientset(
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{Name: "broad", Namespace: "workload"},
			Rules: []rbacv1.PolicyRule{{
				APIGroups: []string{""}, Resources: []string{"*"}, Verbs: []string{"*"},
			}},
		},
	)
	for _, id := range []string{"4.1.2", "4.1.4", "4.1.9", "4.1.10", "4.1.11", "4.1.12"} {
		t.Run(id, func(t *testing.T) {
			got := EvaluateEKSManualEvidence(context.Background(), client, Check{ID: id, Title: "RBAC evidence", Target: "policies", Manual: true})
			if got.Status != "MANUAL" || len(got.Evidence) != 1 || got.Evidence[0].Name != "broad" {
				t.Fatalf("wildcard grant must be reviewable under %s: %+v", id, got)
			}
		})
	}
}

package audit

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEKSWorkloadPolicyCoverageDirections(t *testing.T) {
	pod := func(name, namespace, app string) *corev1.Pod {
		return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": app}}}
	}
	policy := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "ingress-only", Namespace: "apps"},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "selected"}},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
		},
	}
	client := fake.NewSimpleClientset(pod("selected", "apps", "selected"), pod("other", "apps", "other"), pod("different-namespace", "other", "selected"), policy)
	got := EvaluateEKSManualEvidence(context.Background(), client, Check{ID: "4.3.2"})
	if got.Status != "MANUAL" || len(got.Evidence) != 3 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	for _, e := range got.Evidence {
		want := "ingress=false egress=false"
		if e.Name == "selected" {
			want = "ingress=true egress=false"
		}
		if !strings.Contains(e.Observed, want) {
			t.Fatalf("%s: expected %s, got %s", e.Name, want, e.Observed)
		}
	}
}

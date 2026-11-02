package audit

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestClassifyRBACRule(t *testing.T) {
	tests := []struct {
		name string
		rule rbacv1.PolicyRule
		want int
	}{
		{"wildcard", rbacv1.PolicyRule{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}, 1},
		{"impersonate", rbacv1.PolicyRule{Verbs: []string{"impersonate"}, APIGroups: []string{""}, Resources: []string{"users"}}, 1},
		{"secret read", rbacv1.PolicyRule{Verbs: []string{"get", "list"}, APIGroups: []string{""}, Resources: []string{"secrets"}}, 1},
		{"safe read pods", rbacv1.PolicyRule{Verbs: []string{"get"}, APIGroups: []string{""}, Resources: []string{"pods"}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(classifyRBACRule(tt.rule)); got != tt.want {
				t.Fatalf("risk count=%d want=%d", got, tt.want)
			}
		})
	}
}

func TestEvaluateRBACRisk(t *testing.T) {
	role := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "reader", Namespace: "apps"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get", "list"}, APIGroups: []string{""}, Resources: []string{"secrets"}}}}
	binding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "reader-binding", Namespace: "apps"}, RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "reader"}, Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Name: "app", Namespace: "apps"}}}
	client := fake.NewSimpleClientset(role, binding, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "apps"}})
	got := evaluateRBACRisk(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{Status: "NOT_ASSESSED"})
	if got.Status != "WARN" {
		t.Fatalf("status=%s details=%s", got.Status, got.Details)
	}
	if len(got.Evidence) == 0 {
		t.Fatal("expected RBAC evidence")
	}
}

func TestEvaluateRBACRiskEscalationIsFail(t *testing.T) {
	role := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "escalator"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"escalate"}, APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"clusterroles"}}}}
	binding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "escalator-binding"}, RoleRef: rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "ClusterRole", Name: "escalator"}, Subjects: []rbacv1.Subject{{Kind: "User", Name: "operator"}}}
	client := fake.NewSimpleClientset(role, binding)
	got := evaluateRBACRisk(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "FAIL" {
		t.Fatalf("status=%s details=%s", got.Status, got.Details)
	}
}

func TestEvaluateNetworkSecurityCoverageAndExposure(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "apps", Labels: map[string]string{"app": "web"}}}
	policy := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "default-deny", Namespace: "apps"}, Spec: networkingv1.NetworkPolicySpec{PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}}, PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress}}}
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "apps"}, Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer}}
	client := fake.NewSimpleClientset(pod, policy, service)
	got := evaluateNetworkSecurity(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if got.Status != "WARN" {
		t.Fatalf("status=%s details=%s", got.Status, got.Details)
	}
}

func TestEvaluateNetworkSecurityMissingPolicyWarns(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "unprotected", Namespace: "apps"}})
	got := evaluateNetworkSecurity(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if got.Status != "WARN" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestEvaluateNetworkSecurityPodListError(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) { return true, nil, errors.New("forbidden") })
	got := evaluateNetworkSecurity(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestEvaluateControlPlaneSecurityVisibleRisk(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "kube-apiserver-control-1", Namespace: "kube-system"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "kube-apiserver", Command: []string{"kube-apiserver"}, Args: []string{"--anonymous-auth=true", "--authorization-mode=AlwaysAllow,RBAC"}}}}}
	client := fake.NewSimpleClientset(pod)
	got := evaluateControlPlaneSecurity(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "FAIL" {
		t.Fatalf("status=%s details=%s", got.Status, got.Details)
	}
}

func TestEvaluateControlPlaneSecurityManagedPlaneManual(t *testing.T) {
	client := fake.NewSimpleClientset()
	got := evaluateControlPlaneSecurity(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "MANUAL" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestEvaluateAuditLoggingVisibleConfigurationManual(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "kube-apiserver-control-1", Namespace: "kube-system"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "kube-apiserver", Args: []string{"--audit-policy-file=/etc/kubernetes/audit-policy.yaml", "--audit-log-path=/var/log/kubernetes/audit.log"}}}}}
	client := fake.NewSimpleClientset(pod)
	got := evaluateAuditLogging(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "MANUAL" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestEvaluateAuditLoggingNoBackendWarns(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "kube-apiserver-control-1", Namespace: "kube-system"}}
	client := fake.NewSimpleClientset(pod)
	got := evaluateAuditLogging(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "WARN" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestBuildComplianceSummaryWithholdsIncompleteScore(t *testing.T) {
	checks := []Check{{ID: "1.1", Scored: true}, {ID: "1.2", Scored: true}, {ID: "KUBECIS-2.5.1", Scored: false}}
	results := []Result{{ID: "1.1", Status: "PASS"}, {ID: "1.2", Status: "NOT_ASSESSED"}}
	summary, result := buildComplianceSummary(checks, results)
	if summary.ScoredTotal != 2 || summary.ScoredPass != 1 || summary.ScoredNotAssessed != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.CompliancePercent != nil {
		t.Fatal("compliance percentage must be withheld for incomplete scored controls")
	}
	if summary.Extensions != 1 {
		t.Fatalf("extensions=%d want=1", summary.Extensions)
	}
	if result.Status != "INFO" {
		t.Fatalf("summary result status=%s", result.Status)
	}
}

func TestBuildComplianceSummaryCalculatesCompleteScore(t *testing.T) {
	checks := []Check{{ID: "1.1", Scored: true}, {ID: "1.2", Scored: true}}
	results := []Result{{ID: "1.1", Status: "PASS"}, {ID: "1.2", Status: "FAIL"}}
	summary, _ := buildComplianceSummary(checks, results)
	if summary.CompliancePercent == nil || *summary.CompliancePercent != 50 {
		t.Fatalf("unexpected compliance summary: %+v", summary)
	}
}

func TestBenchmarkHasUniqueRegisteredMilestoneKinds(t *testing.T) {
	benchmark, err := readBench("../../benchmarks/cis-2.1.0/benchmark.yaml")
	if err != nil {
		t.Fatalf("read benchmark: %v", err)
	}
	seen := map[string]bool{}
	want := map[string]bool{"KUBECIS-2.5.1": true, "KUBECIS-2.6.1": true, "KUBECIS-2.7.1": true, "KUBECIS-2.8.1": true, "KUBECIS-2.9.1": true}
	for _, check := range benchmark.Checks {
		if seen[check.ID] {
			t.Fatalf("duplicate benchmark ID %q", check.ID)
		}
		seen[check.ID] = true
		if want[check.ID] && check.Scored {
			t.Fatalf("KubeCIS extension %s must not be scored as official CIS", check.ID)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("missing milestone check %s", id)
		}
	}
}

func TestEvaluateRBACRiskListFailureIsNotAssessed(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "clusterroles", func(k8stesting.Action) (bool, runtime.Object, error) { return true, nil, errors.New("forbidden") })
	got := evaluateRBACRisk(context.Background(), client, Check{}, Options{}, Result{})
	if got.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestFlagValueSupportsEqualsAndSeparateArgument(t *testing.T) {
	if got, ok := flagValue([]string{"--anonymous-auth=false"}, "--anonymous-auth"); !ok || got != "false" {
		t.Fatalf("equals form: got %q, %t", got, ok)
	}
	if got, ok := flagValue([]string{"--anonymous-auth", "false"}, "--anonymous-auth"); !ok || got != "false" {
		t.Fatalf("separate form: got %q, %t", got, ok)
	}
}

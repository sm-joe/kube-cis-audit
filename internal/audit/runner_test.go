package audit

import (
	"context"
	"os"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func boolPtr(v bool) *bool       { return &v }
func int64Ptr(v int64) *int64    { return &v }
func stringPtr(v string) *string { return &v }

func TestEvaluateContainerNonRoot(t *testing.T) {
	tests := []struct {
		name      string
		pod       *corev1.PodSecurityContext
		container *corev1.SecurityContext
		want      string
	}{
		{"explicit true", nil, &corev1.SecurityContext{RunAsNonRoot: boolPtr(true)}, "pass"},
		{"explicit false", nil, &corev1.SecurityContext{RunAsNonRoot: boolPtr(false)}, "fail"},
		{"explicit false but fixed nonroot UID", nil, &corev1.SecurityContext{RunAsNonRoot: boolPtr(false), RunAsUser: int64Ptr(10001)}, "pass"},
		{"uid zero", nil, &corev1.SecurityContext{RunAsUser: int64Ptr(0)}, "fail"},
		{"fixed nonroot uid", nil, &corev1.SecurityContext{RunAsUser: int64Ptr(10001)}, "pass"},
		{"missing is unknown", nil, nil, "unknown"},
		{"pod default inherited", &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)}, nil, "pass"},
		{"container override wins", &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)}, &corev1.SecurityContext{RunAsNonRoot: boolPtr(false)}, "fail"},
		{"nonroot plus uid zero conflict", nil, &corev1.SecurityContext{RunAsNonRoot: boolPtr(true), RunAsUser: int64Ptr(0)}, "fail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := evaluateContainerSecurity("api_pod_non_root", tt.pod, tt.container)
			if got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEvaluateContainerSeccomp(t *testing.T) {
	tests := []struct {
		name    string
		profile *corev1.SeccompProfile
		want    string
	}{
		{"runtime default", &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}, "pass"},
		{"localhost", &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost, LocalhostProfile: stringPtr("profiles/custom.json")}, "pass"},
		{"localhost without profile path", &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost}, "fail"},
		{"localhost nil profile pointer", &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost, LocalhostProfile: nil}, "fail"},
		{"localhost whitespace profile", &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost, LocalhostProfile: stringPtr("  ")}, "fail"},
		{"unconfined", &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeUnconfined}, "fail"},
		{"unset", nil, "missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := evaluateContainerSecurity("api_pod_seccomp", &corev1.PodSecurityContext{SeccompProfile: tt.profile}, nil)
			if got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEvaluateContainerSeccompContainerOverrideWins(t *testing.T) {
	podSC := &corev1.PodSecurityContext{SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}}
	containerSC := &corev1.SecurityContext{SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeUnconfined}}
	_, got := evaluateContainerSecurity("api_pod_seccomp", podSC, containerSC)
	if got != "fail" {
		t.Fatalf("container-level Unconfined override got %s, want fail", got)
	}
}

func TestSeccompMissingProfileIsFailureWithAccurateEvidence(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "app-pod", Namespace: "app"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	client := fake.NewSimpleClientset(pod)
	r := evaluatePodSecurity(context.Background(), client, Check{Kind: "api_pod_seccomp"}, Options{}, Result{})
	if r.Status != "FAIL" {
		t.Fatalf("got %s, want FAIL: %s", r.Status, r.Details)
	}
	if !strings.Contains(r.Details, "lack an explicit acceptable seccomp profile") || strings.Contains(r.Details, "explicitly incompatible security setting") {
		t.Fatalf("misleading seccomp summary: %s", r.Details)
	}
	if len(r.Evidence) != 1 || !strings.Contains(r.Evidence[0].Observed, "missing") {
		t.Fatalf("missing-profile evidence not recorded accurately: %#v", r.Evidence)
	}
}

func TestPodContainersIncludesAllContainerTypes(t *testing.T) {
	p := corev1.Pod{Spec: corev1.PodSpec{
		InitContainers:      []corev1.Container{{Name: "init"}},
		Containers:          []corev1.Container{{Name: "app"}},
		EphemeralContainers: []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug"}}},
	}}
	got := podContainers(p)
	if len(got) != 3 {
		t.Fatalf("got %d containers, want 3", len(got))
	}
	if got[0].Kind != "initContainer" || got[1].Kind != "container" || got[2].Kind != "ephemeralContainer" {
		t.Fatalf("unexpected container kinds: %#v", got)
	}
}

func TestPolicyControlDoesNotFalseFailWhenNoPSALabels(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app"}})
	r := evaluatePolicyControl(context.Background(), client, Check{}, Options{}, Result{})
	if r.Status != "NOT_ASSESSED" {
		t.Fatalf("got %s, want NOT_ASSESSED: %s", r.Status, r.Details)
	}
}

func TestPolicyControlRequiresManualConfirmationWhenAllNamespacesEnforcePSA(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app", Labels: map[string]string{"pod-security.kubernetes.io/enforce": "restricted"}}})
	r := evaluatePolicyControl(context.Background(), client, Check{}, Options{}, Result{})
	if r.Status != "MANUAL" {
		t.Fatalf("got %s, want MANUAL: %s", r.Status, r.Details)
	}
}

func TestPolicyControlAlternativeCandidateRequiresManualReview(t *testing.T) {
	policy := &admissionv1.ValidatingAdmissionPolicy{ObjectMeta: metav1.ObjectMeta{Name: "deny-root"}, Spec: admissionv1.ValidatingAdmissionPolicySpec{Validations: []admissionv1.Validation{{Expression: "object.spec.containers.all(c, has(c.securityContext) && c.securityContext.runAsNonRoot == true)"}}}}
	binding := &admissionv1.ValidatingAdmissionPolicyBinding{ObjectMeta: metav1.ObjectMeta{Name: "deny-root-binding"}, Spec: admissionv1.ValidatingAdmissionPolicyBindingSpec{PolicyName: "deny-root", ValidationActions: []admissionv1.ValidationAction{admissionv1.Deny}}}
	client := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app"}}, policy, binding)
	r := evaluatePolicyControl(context.Background(), client, Check{}, Options{}, Result{})
	if r.Status != "MANUAL" {
		t.Fatalf("got %s, want MANUAL: %s", r.Status, r.Details)
	}
}

func TestNetworkPolicyScopeIncludesSystemNamespaces(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-public"}},
	)
	r := evaluateNetworkPolicies(context.Background(), client, Check{}, Options{}, Result{})
	if r.Status != "FAIL" {
		t.Fatalf("got %s, want FAIL", r.Status)
	}
	if len(r.Evidence) != 3 {
		t.Fatalf("got %d namespace evidence entries, want 3", len(r.Evidence))
	}
}

func TestClusterAdminIncludesServiceAccounts(t *testing.T) {
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "sa-admin"},
		RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "cluster-admin"},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: "automation", Namespace: "ops"}},
	}
	client := fake.NewSimpleClientset(binding)
	r := evaluateClusterAdmin(context.Background(), client, Check{}, Result{})
	if r.Status != "WARN" {
		t.Fatalf("got %s, want WARN", r.Status)
	}
	if len(r.Evidence) != 1 || r.Evidence[0].Observed != "ServiceAccount:automation (namespace=ops)" {
		t.Fatalf("service account subject missing from evidence: %#v", r.Evidence)
	}
}

func TestNetworkPolicyPassesWhenEveryNamespaceHasPolicy(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "default-deny", Namespace: "app"}},
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "system-guard", Namespace: "kube-system"}},
	)
	r := evaluateNetworkPolicies(context.Background(), client, Check{}, Options{}, Result{})
	if r.Status != "PASS" {
		t.Fatalf("got %s, want PASS: %s", r.Status, r.Details)
	}
}

func TestPodSecurityMissingNonRootConstraintIsWarnNotRootAssertion(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "app-pod", Namespace: "app"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	client := fake.NewSimpleClientset(pod)
	r := evaluatePodSecurity(context.Background(), client, Check{Kind: "api_pod_non_root"}, Options{}, Result{})
	if r.Status != "WARN" {
		t.Fatalf("got %s, want WARN: %s", r.Status, r.Details)
	}
	if len(r.Evidence) != 1 || !strings.Contains(r.Evidence[0].Observed, "image user not inspected") {
		t.Fatalf("missing evidence about unknown image user: %#v", r.Evidence)
	}
}

func TestPodSecurityChecksAllNamespacesIncludingSystem(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "system-pod", Namespace: "kube-system"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", SecurityContext: &corev1.SecurityContext{RunAsNonRoot: boolPtr(true)}}}}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	client := fake.NewSimpleClientset(pod)
	r := evaluatePodSecurity(context.Background(), client, Check{Kind: "api_pod_non_root"}, Options{}, Result{})
	if r.Status != "PASS" {
		t.Fatalf("got %s, want PASS: %s", r.Status, r.Details)
	}
}

func TestClusterAdminIncludesNamespacedRoleBindings(t *testing.T) {
	binding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "namespace-admin", Namespace: "apps"}, RoleRef: rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "cluster-admin"}, Subjects: []rbacv1.Subject{{Kind: "Group", Name: "ops-admins"}}}
	client := fake.NewSimpleClientset(binding)
	r := evaluateClusterAdmin(context.Background(), client, Check{}, Result{})
	if r.Status != "WARN" || len(r.Evidence) != 1 || r.Evidence[0].Resource != "RoleBinding" {
		t.Fatalf("namespaced cluster-admin binding not reported: %#v", r)
	}
}

func TestHostFileModeUsesPermissionMask(t *testing.T) {
	c := Check{ID: "1.1.1", Kind: "host_file_mode", Path: "manifest.yaml", MaxMode: "0600"}
	r := evaluateHostFileMode(os.FileMode(0644), c, Result{ID: "1.1.1", Status: "NOT_ASSESSED"})
	if r.Status != "FAIL" {
		t.Fatalf("got %s, want FAIL: %s", r.Status, r.Details)
	}
	if len(r.Evidence) != 1 || r.Evidence[0].Observed != "0644" {
		t.Fatalf("unexpected permission evidence: %#v", r.Evidence)
	}
	r = evaluateHostFileMode(os.FileMode(0600), c, Result{ID: "1.1.1", Status: "NOT_ASSESSED"})
	if r.Status != "PASS" {
		t.Fatalf("got %s, want PASS: %s", r.Status, r.Details)
	}
}

func TestMissingHostFileIsNotAssessed(t *testing.T) {
	client := fake.NewSimpleClientset()
	r := evaluate(context.Background(), client, Check{ID: "4.1.9", Kind: "host_file_mode", Path: "missing/config.yaml", MaxMode: "0600"}, Options{HostRoot: t.TempDir()})
	if r.Status != "NOT_ASSESSED" {
		t.Fatalf("got %s, want NOT_ASSESSED", r.Status)
	}
}

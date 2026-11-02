package audit

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluatePodPrivilegeSecurity(t *testing.T) {
	tests := []struct {
		name       string
		sc         *corev1.SecurityContext
		wantStatus string
		wantText   string
	}{
		{
			name:       "secure explicit settings",
			sc:         &corev1.SecurityContext{Privileged: boolPtr(false), AllowPrivilegeEscalation: boolPtr(false)},
			wantStatus: "PASS",
		},
		{
			name:       "privileged true fails",
			sc:         &corev1.SecurityContext{Privileged: boolPtr(true), AllowPrivilegeEscalation: boolPtr(false)},
			wantStatus: "FAIL", wantText: "privileged=true",
		},
		{
			name:       "allow escalation true fails",
			sc:         &corev1.SecurityContext{Privileged: boolPtr(false), AllowPrivilegeEscalation: boolPtr(true)},
			wantStatus: "FAIL", wantText: "allowPrivilegeEscalation=true",
		},
		{
			name:       "unset escalation uses true default",
			sc:         &corev1.SecurityContext{Privileged: boolPtr(false)},
			wantStatus: "FAIL", wantText: "default true",
		},
		{
			name: "sys admin forces escalation despite false",
			sc: &corev1.SecurityContext{
				Privileged: boolPtr(false), AllowPrivilegeEscalation: boolPtr(false),
				Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"SYS_ADMIN"}},
			},
			wantStatus: "FAIL", wantText: "CAP_SYS_ADMIN",
		},
		{
			name:       "privilege defaults false but escalation defaults true",
			wantStatus: "FAIL", wantText: "privileged unset (Kubernetes default false)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "apps"},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app", SecurityContext: tt.sc}}},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			}
			client := fake.NewSimpleClientset(pod)
			result := evaluatePodPrivilegeSecurity(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
			if result.Status != tt.wantStatus {
				t.Fatalf("status=%s, want %s; details=%s", result.Status, tt.wantStatus, result.Details)
			}
			if tt.wantText != "" {
				found := false
				for _, e := range result.Evidence {
					if strings.Contains(e.Observed, tt.wantText) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected evidence containing %q; got %#v", tt.wantText, result.Evidence)
				}
			}
		})
	}
}

func TestEvaluatePodPrivilegeSecurityExcludesTerminalPods(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "completed", Namespace: "apps"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	client := fake.NewSimpleClientset(pod)
	result := evaluatePodPrivilegeSecurity(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s, want NOT_ASSESSED", result.Status)
	}
}

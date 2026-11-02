package audit

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluatePodCapabilities(t *testing.T) {
	tests := []struct {
		name       string
		caps       *corev1.Capabilities
		wantStatus string
		wantText   string
	}{
		{
			name:       "drop all and add none passes",
			caps:       &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
			wantStatus: "PASS",
			wantText:   "ALL dropped",
		},
		{
			name:       "missing capability configuration warns",
			caps:       nil,
			wantStatus: "WARN",
			wantText:   "does not include ALL",
		},
		{
			name:       "partial drop does not satisfy drop all",
			caps:       &corev1.Capabilities{Drop: []corev1.Capability{"NET_RAW"}},
			wantStatus: "WARN",
			wantText:   "ALL not dropped",
		},
		{
			name:       "added capability requires review even when all dropped",
			caps:       &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}, Add: []corev1.Capability{"NET_BIND_SERVICE"}},
			wantStatus: "WARN",
			wantText:   "NET_BIND_SERVICE",
		},
		{
			name:       "cap prefix and case normalized for all",
			caps:       &corev1.Capabilities{Drop: []corev1.Capability{"cap_all"}},
			wantStatus: "PASS",
			wantText:   "ALL dropped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "apps"},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name: "app", SecurityContext: &corev1.SecurityContext{Capabilities: tt.caps},
				}}},
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
			}
			client := fake.NewSimpleClientset(pod)
			result := evaluatePodCapabilities(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
			if result.Status != tt.wantStatus {
				t.Fatalf("status=%s, want %s; details=%s", result.Status, tt.wantStatus, result.Details)
			}
			if tt.wantText != "" {
				found := false
				for _, evidence := range result.Evidence {
					if strings.Contains(evidence.Observed, tt.wantText) {
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

func TestEvaluatePodCapabilitiesIncludesAllContainerTypes(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "workload", Namespace: "apps"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Name: "init", SecurityContext: &corev1.SecurityContext{Capabilities: &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}}}}},
			Containers:     []corev1.Container{{Name: "app", SecurityContext: &corev1.SecurityContext{Capabilities: &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}}}}},
			EphemeralContainers: []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{
				Name: "debug", SecurityContext: &corev1.SecurityContext{Capabilities: &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}}},
			}}},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	client := fake.NewSimpleClientset(pod)
	result := evaluatePodCapabilities(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if result.Status != "PASS" {
		t.Fatalf("status=%s, want PASS; details=%s", result.Status, result.Details)
	}
	if len(result.Evidence) != 6 {
		t.Fatalf("evidence entries=%d, want 6 (two per container); evidence=%#v", len(result.Evidence), result.Evidence)
	}
}

func TestEvaluatePodCapabilitiesExcludesTerminalPods(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "completed", Namespace: "apps"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	client := fake.NewSimpleClientset(pod)
	result := evaluatePodCapabilities(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s, want NOT_ASSESSED", result.Status)
	}
}

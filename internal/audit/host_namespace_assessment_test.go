package audit

import (
	"context"
	"errors"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestEvaluatePodHostNamespaces(t *testing.T) {
	tests := []struct {
		name        string
		hostPID     bool
		hostIPC     bool
		hostNetwork bool
		wantStatus  string
		wantField   string
	}{
		{
			name:       "all host namespaces disabled",
			wantStatus: "PASS",
		},
		{
			name:       "host PID enabled",
			hostPID:    true,
			wantStatus: "FAIL",
			wantField:  "spec.hostPID",
		},
		{
			name:       "host IPC enabled",
			hostIPC:    true,
			wantStatus: "FAIL",
			wantField:  "spec.hostIPC",
		},
		{
			name:        "host network enabled",
			hostNetwork: true,
			wantStatus:  "FAIL",
			wantField:   "spec.hostNetwork",
		},
		{
			name:        "multiple host namespaces enabled",
			hostPID:     true,
			hostIPC:     true,
			hostNetwork: true,
			wantStatus:  "FAIL",
			wantField:   "spec.hostPID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "apps"},
				Spec: corev1.PodSpec{
					HostPID:     tt.hostPID,
					HostIPC:     tt.hostIPC,
					HostNetwork: tt.hostNetwork,
				},
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
			}
			client := fake.NewSimpleClientset(pod)
			result := evaluatePodHostNamespaces(
				context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{},
			)

			if result.Status != tt.wantStatus {
				t.Fatalf("status=%s, want %s; details=%s", result.Status, tt.wantStatus, result.Details)
			}
			if len(result.Evidence) != 3 {
				t.Fatalf("evidence entries=%d, want 3; evidence=%#v", len(result.Evidence), result.Evidence)
			}

			if tt.wantField != "" {
				found := false
				for _, evidence := range result.Evidence {
					if evidence.Field == tt.wantField && strings.Contains(evidence.Observed, "true") {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected enabled evidence for %q; got %#v", tt.wantField, result.Evidence)
				}
			}
		})
	}
}

func TestEvaluatePodHostNamespacesTreatsUnsetFieldsAsDisabled(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "defaults", Namespace: "apps"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	client := fake.NewSimpleClientset(pod)

	result := evaluatePodHostNamespaces(
		context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{},
	)

	if result.Status != "PASS" {
		t.Fatalf("status=%s, want PASS; details=%s", result.Status, result.Details)
	}
	for _, evidence := range result.Evidence {
		if !strings.Contains(evidence.Observed, "Kubernetes default false") {
			t.Errorf("evidence for %s does not describe the default: %q", evidence.Field, evidence.Observed)
		}
	}
}

func TestEvaluatePodHostNamespacesExcludesTerminalPods(t *testing.T) {
	pods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "succeeded", Namespace: "apps"},
			Spec:       corev1.PodSpec{HostPID: true},
			Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "failed", Namespace: "apps"},
			Spec:       corev1.PodSpec{HostIPC: true},
			Status:     corev1.PodStatus{Phase: corev1.PodFailed},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "running", Namespace: "apps"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
	}
	client := fake.NewSimpleClientset(pods[0], pods[1], pods[2])

	result := evaluatePodHostNamespaces(
		context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{},
	)

	if result.Status != "PASS" {
		t.Fatalf("status=%s, want PASS; details=%s", result.Status, result.Details)
	}
	if len(result.Evidence) != 3 {
		t.Fatalf("evidence entries=%d, want 3 for the one non-terminal Pod; evidence=%#v", len(result.Evidence), result.Evidence)
	}
	for _, evidence := range result.Evidence {
		if evidence.Name != "running" {
			t.Errorf("terminal Pod included in evidence: %#v", evidence)
		}
	}
}

func TestEvaluatePodHostNamespacesReturnsNotAssessedForEmptyInventory(t *testing.T) {
	client := fake.NewSimpleClientset()
	result := evaluatePodHostNamespaces(
		context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{},
	)

	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s, want NOT_ASSESSED; details=%s", result.Status, result.Details)
	}
	if len(result.Evidence) != 0 {
		t.Fatalf("evidence entries=%d, want 0", len(result.Evidence))
	}
}

func TestEvaluatePodHostNamespacesReturnsNotAssessedOnListError(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("simulated Pod list failure")
	})

	result := evaluatePodHostNamespaces(
		context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{},
	)

	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s, want NOT_ASSESSED; details=%s", result.Status, result.Details)
	}
	if !strings.Contains(result.Details, "simulated Pod list failure") {
		t.Fatalf("details do not include the API error: %q", result.Details)
	}
}

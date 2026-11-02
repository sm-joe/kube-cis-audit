package audit

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluatePodReadOnlyRootFilesystem(t *testing.T) {
	tests := []struct {
		name       string
		sc         *corev1.SecurityContext
		mounts     []corev1.VolumeMount
		wantStatus string
		wantText   string
	}{
		{
			name:       "read-only root filesystem passes",
			sc:         &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPtr(true)},
			wantStatus: "PASS",
			wantText:   "readOnlyRootFilesystem=true",
		},
		{
			name:       "explicit writable root filesystem fails",
			sc:         &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPtr(false)},
			wantStatus: "FAIL",
			wantText:   "readOnlyRootFilesystem=false",
		},
		{
			name:       "unset root filesystem setting fails as Kubernetes default is writable",
			sc:         nil,
			wantStatus: "FAIL",
			wantText:   "Kubernetes default false",
		},
		{
			name:       "read-only root with writable data mount warns separately",
			sc:         &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPtr(true)},
			mounts:     []corev1.VolumeMount{{Name: "data", MountPath: "/var/lib/app"}},
			wantStatus: "WARN",
			wantText:   "readOnly=false (writable mount)",
		},
		{
			name:       "read-only root and read-only mount pass",
			sc:         &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPtr(true)},
			mounts:     []corev1.VolumeMount{{Name: "config", MountPath: "/etc/app", ReadOnly: true}},
			wantStatus: "PASS",
			wantText:   "readOnly=true",
		},
		{
			name:       "writable root fails even when all volume mounts are read-only",
			sc:         &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPtr(false)},
			mounts:     []corev1.VolumeMount{{Name: "config", MountPath: "/etc/app", ReadOnly: true}},
			wantStatus: "FAIL",
			wantText:   "readOnlyRootFilesystem=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "apps"},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name: "app", SecurityContext: tt.sc, VolumeMounts: tt.mounts,
				}}},
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
			}
			client := fake.NewSimpleClientset(pod)
			result := evaluatePodReadOnlyRootFilesystem(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
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

func TestEvaluatePodReadOnlyRootFilesystemIncludesAllContainerTypes(t *testing.T) {
	readOnly := &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPtr(true)}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "workload", Namespace: "apps"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{Name: "init", SecurityContext: readOnly}},
			Containers:     []corev1.Container{{Name: "app", SecurityContext: readOnly}},
			EphemeralContainers: []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{
				Name: "debug", SecurityContext: readOnly,
			}}},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	client := fake.NewSimpleClientset(pod)
	result := evaluatePodReadOnlyRootFilesystem(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if result.Status != "PASS" {
		t.Fatalf("status=%s, want PASS; details=%s", result.Status, result.Details)
	}
	if len(result.Evidence) != 3 {
		t.Fatalf("evidence entries=%d, want 3; evidence=%#v", len(result.Evidence), result.Evidence)
	}
}

func TestEvaluatePodReadOnlyRootFilesystemExcludesTerminalPods(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "completed", Namespace: "apps"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	client := fake.NewSimpleClientset(pod)
	result := evaluatePodReadOnlyRootFilesystem(context.Background(), client, Check{}, Options{Namespace: "apps"}, Result{})
	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s, want NOT_ASSESSED", result.Status)
	}
}

package audit

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluatePodGroupSecurityContext(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "app", Name: "inherits"},
			Spec: corev1.PodSpec{
				SecurityContext: &corev1.PodSecurityContext{RunAsGroup: int64Ptr(10002), FSGroup: int64Ptr(10003)},
				Containers:      []corev1.Container{{Name: "api"}, {Name: "worker", SecurityContext: &corev1.SecurityContext{RunAsGroup: int64Ptr(0)}}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "app", Name: "unset"},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "sidecar"}}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "app", Name: "completed"},
			Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "done"}}},
		},
	}
	client := fake.NewSimpleClientset(&pods[0], &pods[1], &pods[2])
	check := Check{ID: "KUBECIS-2.4.1", Title: "Group security-context inventory", Target: "policies", Kind: "api_pod_group_security_context"}
	result := evaluatePodGroupSecurityContext(context.Background(), client, check, Options{}, Result{ID: check.ID, Title: check.Title, Target: check.Target, Method: check.Kind})

	if result.Status != "WARN" {
		t.Fatalf("status = %q, want WARN; details=%s", result.Status, result.Details)
	}
	if !strings.Contains(result.Details, "2 non-terminal Pods and 3 containers") {
		t.Fatalf("unexpected pod/container counts: %s", result.Details)
	}
	if len(result.Evidence) != 5 {
		t.Fatalf("evidence count = %d, want 5", len(result.Evidence))
	}
	foundOverride := false
	foundFSGroup := false
	for _, evidence := range result.Evidence {
		if evidence.Name == "inherits/worker" && evidence.Observed == "runAsGroup=0 (root group)" {
			foundOverride = true
		}
		if evidence.Name == "inherits" && evidence.Field == "securityContext.fsGroup" && evidence.Observed == "fsGroup=10003" {
			foundFSGroup = true
		}
	}
	if !foundOverride {
		t.Fatal("container-level runAsGroup override was not reported")
	}
	if !foundFSGroup {
		t.Fatal("Pod-level fsGroup was not reported")
	}
}

func TestEvaluatePodGroupSecurityContextAllUnsetIsNotAssessed(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "app", Name: "unset"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "api"}}},
	}
	client := fake.NewSimpleClientset(pod)
	result := evaluatePodGroupSecurityContext(context.Background(), client, Check{}, Options{}, Result{})
	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status = %q, want NOT_ASSESSED", result.Status)
	}
}

package audit

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEKSRegistryReviewIncludesAllContainerClasses(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "registry-test", Namespace: "demo"},
		Spec: corev1.PodSpec{
			Containers:     []corev1.Container{{Name: "app", Image: "nginx:latest"}},
			InitContainers: []corev1.Container{{Name: "init", Image: "init.example.org/helper:v1"}},
			EphemeralContainers: []corev1.EphemeralContainer{
				{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug", Image: "debug.example.org/tool:v1"}},
			},
		},
	}
	client := fake.NewSimpleClientset(pod)
	got := EvaluateEKSManualEvidence(context.Background(), client, Check{ID: "5.1.4"})
	if got.Status != "MANUAL" {
		t.Fatalf("registry inventory cannot prove admission enforcement: %+v", got)
	}
	seen := map[string]bool{}
	for _, e := range got.Evidence {
		seen[e.Name] = true
	}
	for _, registry := range []string{"docker.io (implicit)", "init.example.org", "debug.example.org"} {
		if !seen[registry] {
			t.Fatalf("missing registry %q: %+v", registry, got.Evidence)
		}
	}
}

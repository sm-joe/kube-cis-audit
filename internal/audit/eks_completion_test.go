package audit

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAdditionalEKSPolicySecretEnvAndDefaultNamespace(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "default"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "app", Image: "example",
				Env: []corev1.EnvVar{{Name: "TOKEN", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "credential"}, Key: "token"}}}},
			}},
		},
	}
	client := fake.NewSimpleClientset(pod)
	for _, tc := range []struct{ id, kind string }{
		{"4.4.1", "eks_secret_env"}, {"4.5.2", "eks_default_namespace"},
	} {
		r := EvaluateEKSAdditionalPolicy(context.Background(), client, Check{ID: tc.id, Kind: tc.kind, Target: "policies", Scored: true})
		if r.Status != "FAIL" || len(r.Evidence) == 0 {
			t.Fatalf("%s should detect actual workload configuration: %+v", tc.id, r)
		}
	}
}

func TestEffectiveKubeletConfigzNeverAssumesDefault(t *testing.T) {
	report := Report{
		AuditMode: "node-local", CloudProvider: "eks", Benchmark: "cis-eks-2.1.0", NodeName: "worker-a",
		Results: []Result{
			{ID: "3.2.5", Scope: "worker-a", Target: "node", Method: "host_kubelet_config_field", Status: "NOT_ASSESSED", Scored: true},
			{ID: "3.2.6", Scope: "worker-a", Target: "node", Method: "host_kubelet_config_field", Status: "NOT_ASSESSED", Scored: true},
			{ID: "3.2.7", Scope: "worker-a", Target: "node", Method: "host_kubelet_config_field", Status: "NOT_ASSESSED", Scored: true},
		},
	}
	if err := EnrichNodeWithConfigz(&report, "other-worker", []byte(`{"kubeletconfig":{}}`)); err == nil {
		t.Fatal("accepted configz evidence for a different node")
	}
	if err := EnrichNodeWithConfigz(&report, "worker-a", []byte(`{"kubeletconfig":{"makeIPTablesUtilChains":true,"eventRecordQPS":50,"rotateCertificates":false}}`)); err != nil {
		t.Fatal(err)
	}
	want := []string{"PASS", "MANUAL", "MANUAL"}
	for i, r := range report.Results {
		if r.Status != want[i] || r.EvidenceSource != "kubernetes-node-proxy-configz" {
			t.Fatalf("control %s incorrect: %+v", r.ID, r)
		}
	}
	if report.Summary.Pass != 1 || report.Summary.Manual != 2 || report.Compliance.ScoredNotAssessed != 0 || report.Compliance.CompliancePercent != nil {
		t.Fatalf("inconsistent enriched counters: %+v %+v", report.Summary, report.Compliance)
	}
}

func TestEKSAccessAndECRMissingEvidenceFailClosed(t *testing.T) {
	e := EKSProviderEvidence{AuthenticationMode: "API_AND_CONFIG_MAP", AccessInventoryStatus: "UNAVAILABLE", ECRInventoryStatus: "UNAVAILABLE"}
	results := EvaluateEKSAdditionalAWSControls(e)
	if len(results) != 3 {
		t.Fatalf("expected 3 provider-native controls, got %d", len(results))
	}
	for _, r := range results {
		if r.Status == "PASS" {
			t.Fatalf("missing AWS evidence passed control %s", r.ID)
		}
	}
	e.AuthenticationMode = "CONFIG_MAP"
	results = EvaluateEKSAdditionalAWSControls(e)
	if results[0].Status != "FAIL" || results[1].Status != "FAIL" {
		t.Fatalf("CONFIG_MAP-only authentication incorrectly accepted: %+v", results)
	}
}

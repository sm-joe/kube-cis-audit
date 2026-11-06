package audit

import (
	"testing"
)

func TestExtendedKubeletMetadataAndDecisions(t *testing.T) {
	cfg := map[string]interface{}{
		"authentication": map[string]interface{}{
			"x509": map[string]interface{}{"clientCAFile": "/etc/kubernetes/pki/ca.crt"},
		},
		"readOnlyPort":           0,
		"makeIPTablesUtilChains": true,
		"eventRecordQPS":         0,
		"rotateCertificates":     true,
		"serverTLSBootstrap":     true,
		"featureGates":           map[string]interface{}{"RotateKubeletServerCertificate": true},
	}
	values := map[string]KubeletSetting{}
	collectExtendedKubeletSettings(cfg, values)
	for field, value := range values {
		value.SourcePath = "etc/kubernetes/kubelet/config.json"
		values[field] = value
	}
	snapshot := NodeMetadataSnapshot{KubeletSettings: values}
	cases := []struct {
		id    string
		field string
	}{
		{"3.2.3", "authentication.x509.clientCAFile"},
		{"3.2.4", "readOnlyPort"},
		{"3.2.5", "makeIPTablesUtilChains"},
		{"3.2.6", "eventRecordQPS"},
		{"3.2.7", "rotateCertificates"},
		{"3.2.8", "serverTLSBootstrap"},
	}
	for _, tc := range cases {
		check := Check{ID: tc.id, Field: tc.field, Path: "var/lib/kubelet/config.yaml", Kind: "host_kubelet_config_field", Target: "node", Scored: true}
		got := evaluateExtendedKubeletSetting("worker-a", check, snapshot)
		expected := "PASS"
		if tc.id == "3.2.6" {
			expected = "MANUAL" // 0 means unlimited event traffic, not automatic compliance.
		}
		if got.Status != expected {
			t.Errorf("%s got %s, want %s: %s", tc.id, got.Status, expected, got.Details)
		}
	}
	values["readOnlyPort"] = KubeletSetting{SourcePath: "etc/kubernetes/kubelet/config.json", Value: "10255"}
	check := Check{ID: "3.2.4", Field: "readOnlyPort", Path: "var/lib/kubelet/config.yaml"}
	if got := evaluateExtendedKubeletSetting("worker-a", check, snapshot); got.Status != "FAIL" {
		t.Fatalf("nonzero read-only port should fail; got %+v", got)
	}
	delete(values, "readOnlyPort")
	if got := evaluateExtendedKubeletSetting("worker-a", check, snapshot); got.Status != "NOT_ASSESSED" {
		t.Fatalf("missing evidence must fail closed; got %+v", got)
	}
}

func TestExtendedKubeletCertificateRotationRequiresFeatureGate(t *testing.T) {
	fields := map[string]KubeletSetting{
		"serverTLSBootstrap": {SourcePath: "etc/kubernetes/kubelet/config.json", Value: "true"},
	}
	check := Check{ID: "3.2.8", Field: "serverTLSBootstrap", Path: "var/lib/kubelet/config.yaml"}
	report := NodeMetadataSnapshot{KubeletSettings: fields}
	if got := evaluateExtendedKubeletSetting("worker-a", check, report); got.Status != "NOT_ASSESSED" {
		t.Fatalf("missing server rotation evidence must fail closed: %+v", got)
	}
}

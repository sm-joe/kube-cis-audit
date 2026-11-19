package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKubeletClientAuthMetadataUsesOnlyActiveContext(t *testing.T) {
	dir := t.TempDir()
	raw := "apiVersion: v1\nkind: Config\ncurrent-context: aws-context\ncontexts:\n" +
		"- name: aws-context\n  context:\n    user: aws-user\n" +
		"users:\n" +
		"- name: unused-cert\n  user:\n    client-certificate-data: PRIVATE-CERT-DATA-MUST-NOT-LEAK\n    client-key-data: PRIVATE-KEY-DATA-MUST-NOT-LEAK\n    token: SUPER-SECRET-UNUSED-TOKEN\n" +
		"- name: aws-user\n  user:\n    exec:\n      command: /usr/local/bin/aws\n      args: [eks, get-token, --cluster-name, private-cluster]\n"
	if err := os.WriteFile(filepath.Join(dir, "kubeconfig"), []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	got := collectKubeletClientAuthentication(dir)
	if got.Mode != KubeletIAMExecToken || got.Error != "" {
		t.Fatalf("did not recognize active AWS exec token: %+v", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"SUPER-SECRET", "PRIVATE-CERT", "PRIVATE-KEY", "private-cluster"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("authentication metadata leaked kubeconfig contents: %s", secret)
		}
	}
}

func TestKubeletClientAuthCertificateAndUnsafeInputs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "kubeconfig")
	write := func(raw string) {
		t.Helper()
		if err := os.WriteFile(file, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("current-context: cert-context\ncontexts:\n- name: cert-context\n  context:\n    user: cert-user\nusers:\n- name: cert-user\n  user:\n    client-certificate: /etc/kubernetes/pki/kubelet.crt\n    client-key: /etc/kubernetes/pki/kubelet.key\n")
	if got := collectKubeletClientAuthentication(dir); got.Mode != KubeletClientCert {
		t.Fatalf("client certificate authentication not detected: %+v", got)
	}
	write("current-context: cert-context\ncontexts:\n- name: cert-context\n  context:\n    user: cert-user\nusers:\n- name: cert-user\n  user:\n    client-certificate: /etc/kubernetes/pki/kubelet.crt\n")
	if got := collectKubeletClientAuthentication(dir); got.Mode != KubeletAuthUnknown {
		t.Fatalf("incomplete client cert credentials were trusted: %+v", got)
	}
	write("current-context: aws-context\ncontexts:\n- name: aws-context\n  context:\n    user: aws-user\nusers:\n- name: aws-user\n  user:\n    token: unsafe-mixed-auth\n    exec:\n      command: aws\n      args: [eks, get-token]\n")
	if got := collectKubeletClientAuthentication(dir); got.Mode != KubeletAuthUnknown {
		t.Fatalf("mixed active token plus exec credentials were trusted: %+v", got)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "private")
	if err := os.WriteFile(outside, []byte("token: PRIVATE-SECRET"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, file); err != nil {
		t.Fatal(err)
	}
	if got := collectKubeletClientAuthentication(dir); got.Mode != KubeletAuthUnknown || !strings.Contains(got.Error, "symlink") {
		t.Fatalf("symlink should never be followed: %+v", got)
	}
}

func TestEKSRotationRequiresVerifiedApplicability(t *testing.T) {
	c := Check{ID: "3.2.7", Field: "rotateCertificates", Path: "var/lib/kubelet/config.yaml", Kind: "host_kubelet_config_field", Target: "node", Scored: true}
	settings := map[string]KubeletSetting{"rotateCertificates": {SourcePath: "etc/kubernetes/kubelet/config.json", Value: "false"}}
	snap := NodeMetadataSnapshot{KubeletSettings: settings, KubeletClientAuthentication: &KubeletClientAuthentication{Mode: KubeletIAMExecToken}}
	if got := evaluateExtendedKubeletSetting("worker-a", c, snap); got.Status != "MANUAL" || len(got.Evidence) != 1 {
		t.Fatalf("AWS IAM exec-token client must be reviewed as not applicable: %+v", got)
	}
	snap.KubeletClientAuthentication.Mode = KubeletClientCert
	if got := evaluateExtendedKubeletSetting("worker-a", c, snap); got.Status != "FAIL" {
		t.Fatalf("client certificate with rotation disabled must FAIL: %+v", got)
	}
	delete(settings, "rotateCertificates")
	if got := evaluateExtendedKubeletSetting("worker-a", c, snap); got.Status != "NOT_ASSESSED" {
		t.Fatalf("missing rotation setting for cert client must remain unknown: %+v", got)
	}
}

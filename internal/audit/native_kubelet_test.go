package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNativeKubeletConfiguration(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only host configuration evaluation")
	}
	root := t.TempDir()
	path := filepath.Join(root, "var", "lib", "kubelet", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	check := Check{ID: "KUBECIS-NATIVE-NODE-1", Target: "node", Kind: "host_kubelet_config_field", Path: "var/lib/kubelet/config.yaml", Field: "authentication.anonymous.enabled"}
	write := func(content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("authentication:\n  anonymous:\n    enabled: false\nauthorization:\n  mode: Webhook\n")
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "PASS" {
		t.Fatalf("expected PASS: %+v", got)
	}
	check.Field = "authorization.mode"
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "PASS" {
		t.Fatalf("expected Webhook PASS: %+v", got)
	}
	check.Field = "authentication.anonymous.enabled"
	write("authentication:\n  anonymous:\n    enabled: true\n")
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "FAIL" {
		t.Fatalf("expected FAIL: %+v", got)
	}
	write("authentication: {}\n")
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "NOT_ASSESSED" {
		t.Fatalf("missing value must be NOT_ASSESSED: %+v", got)
	}
	write("authentication: [broken")
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "NOT_ASSESSED" {
		t.Fatalf("malformed YAML must be NOT_ASSESSED: %+v", got)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "outside"), path); err != nil {
		t.Fatal(err)
	}
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "NOT_ASSESSED" {
		t.Fatalf("symlink must be NOT_ASSESSED: %+v", got)
	}
	check.Path = "../../outside"
	if got := evaluateKubeletConfigField(root, "worker-a", check); got.Status != "NOT_ASSESSED" {
		t.Fatalf("path escape must be NOT_ASSESSED: %+v", got)
	}
}

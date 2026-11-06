package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEKSAL2023NodeadmKubeletMetadata(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux node auditor")
	}
	base := t.TempDir()
	profiles := filepath.Join(base, "benchmarks", "eks-test")
	if err := os.MkdirAll(profiles, 0700); err != nil {
		t.Fatal(err)
	}
	benchmark := `id: eks-test
name: AL2023 nodeadm audit
compatibilityStatus: provisional
checks:
  - id: 4.1.9
    title: kubelet config permissions
    target: node
    kind: host_file_mode
    path: var/lib/kubelet/config.yaml
    maxMode: "0600"
    scored: true
  - id: KUBECIS-NATIVE-NODE-1
    title: kubelet anonymous auth
    target: node
    kind: host_kubelet_config_field
    path: var/lib/kubelet/config.yaml
    field: authentication.anonymous.enabled
    scored: false
  - id: KUBECIS-NATIVE-NODE-2
    title: kubelet authorization
    target: node
    kind: host_kubelet_config_field
    path: var/lib/kubelet/config.yaml
    field: authorization.mode
    scored: false
`
	if err := os.WriteFile(filepath.Join(profiles, "benchmark.yaml"), []byte(benchmark), 0600); err != nil {
		t.Fatal(err)
	}
	legacyRoot := filepath.Join(base, "legacy")
	nodeadmRoot := filepath.Join(base, "nodeadm")
	for _, path := range []string{legacyRoot, nodeadmRoot} {
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	configPath := filepath.Join(nodeadmRoot, "config.json")
	config := `{"apiVersion":"kubelet.config.k8s.io/v1beta1","kind":"KubeletConfiguration","authentication":{"anonymous":{"enabled":false}},"authorization":{"mode":"Webhook"},"staticPodPath":"SECRET-DO-NOT-LEAK"}`
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	snapshotPath := filepath.Join(base, "metadata", "snapshot.json")
	collect := func() {
		t.Helper()
		if err := RunHostMetadataSnapshotWithKubeletRoot(filepath.Join(base, "benchmarks"), "eks-test", "node", legacyRoot, nodeadmRoot, "eks-worker-1", snapshotPath); err != nil {
			t.Fatal(err)
		}
	}
	collect()
	raw, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "SECRET-DO-NOT-LEAK") || strings.Contains(string(raw), "staticPodPath") {
		t.Fatal("node metadata snapshot leaked raw kubelet configuration")
	}
	report, err := RunNode(Options{ConfigDir: filepath.Join(base, "benchmarks"), Benchmark: "eks-test", Target: "node", MetadataFile: snapshotPath}, "eks-worker-1")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"4.1.9": "PASS", "KUBECIS-NATIVE-NODE-1": "PASS", "KUBECIS-NATIVE-NODE-2": "PASS"}
	for _, result := range report.Results {
		if result.Status != want[result.ID] {
			t.Fatalf("%s got %s: %s", result.ID, result.Status, result.Details)
		}
		found := false
		for _, e := range result.Evidence {
			if e.Name == "etc/kubernetes/kubelet/config.json" {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s has no AL2023 config source evidence: %#v", result.ID, result.Evidence)
		}
	}
	if err := os.Chmod(configPath, 0644); err != nil {
		t.Fatal(err)
	}
	collect()
	report, err = RunNode(Options{ConfigDir: filepath.Join(base, "benchmarks"), Benchmark: "eks-test", Target: "node", MetadataFile: snapshotPath}, "eks-worker-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range report.Results {
		if result.ID == "4.1.9" && result.Status != "FAIL" {
			t.Fatalf("0644 must fail max mode 0600: %+v", result)
		}
	}
	if err := os.Remove(configPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(base, "secret.json"), configPath); err != nil {
		t.Fatal(err)
	}
	collect()
	report, err = RunNode(Options{ConfigDir: filepath.Join(base, "benchmarks"), Benchmark: "eks-test", Target: "node", MetadataFile: snapshotPath}, "eks-worker-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range report.Results {
		if result.Status != "NOT_ASSESSED" {
			t.Fatalf("symlink must never be assessed: %+v", result)
		}
	}
}

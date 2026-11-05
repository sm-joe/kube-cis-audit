package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSplitContainerKubeletEvidenceEndToEnd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux host collector")
	}
	root := t.TempDir()
	profiles := filepath.Join(root, "benchmarks", "test")
	if err := os.MkdirAll(profiles, 0700); err != nil {
		t.Fatal(err)
	}
	benchmark := `id: test
compatibilityStatus: provisional
name: integration
checks:
  - id: 4.1.9
    title: kubelet mode
    target: node
    kind: host_file_mode
    path: var/lib/kubelet/config.yaml
    maxMode: "0600"
    scored: true
  - id: KUBECIS-NATIVE-NODE-1
    title: anonymous
    target: node
    kind: host_kubelet_config_field
    path: var/lib/kubelet/config.yaml
    field: authentication.anonymous.enabled
    scored: false
  - id: KUBECIS-NATIVE-NODE-2
    title: authorization
    target: node
    kind: host_kubelet_config_field
    path: var/lib/kubelet/config.yaml
    field: authorization.mode
    scored: false
`
	if err := os.WriteFile(filepath.Join(profiles, "benchmark.yaml"), []byte(benchmark), 0600); err != nil {
		t.Fatal(err)
	}
	host := filepath.Join(root, "host")
	if err := os.MkdirAll(host, 0700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(host, "config.yaml")
	content := "authentication:\n  anonymous:\n    enabled: false\nauthorization:\n  mode: AlwaysAllow\nclientCAFile: SECRET-MUST-NOT-LEAK\n"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "shared", "snapshot.json")
	profileRoot := filepath.Join(root, "benchmarks")
	if err := RunHostMetadataSnapshot(profileRoot, "test", "node", host, "worker-a", output); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "SECRET-MUST-NOT-LEAK") || strings.Contains(string(data), "clientCAFile") {
		t.Fatal("snapshot leaked non-allowlisted kubelet configuration")
	}
	report, err := RunNode(Options{ConfigDir: profileRoot, Benchmark: "test", Target: "node", MetadataFile: output}, "worker-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Results) != 3 {
		t.Fatalf("expected three results, got %d", len(report.Results))
	}
	want := map[string]string{"4.1.9": "FAIL", "KUBECIS-NATIVE-NODE-1": "PASS", "KUBECIS-NATIVE-NODE-2": "FAIL"}
	for _, r := range report.Results {
		if r.Status != want[r.ID] || r.Scope != "worker-a" || r.FindingID == "" {
			t.Fatalf("incorrect node assessment: %+v", r)
		}
	}
	if report.Compliance.CompliancePercent != nil {
		t.Fatal("provisional benchmark must withhold compliance score")
	}
	if err := os.WriteFile(configPath, []byte("authentication:\n  anonymous:\n    enabled: true\nauthorization:\n  mode: Webhook\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// WriteFile preserves the existing file mode; explicitly model the
	// permission remediation before collecting the refreshed snapshot.
	if err := os.Chmod(configPath, 0600); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(configPath); err != nil {
		t.Fatal(err)
	} else if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("fixture permissions were not remediated: got %04o, want 0600", got)
	}
	if err := RunHostMetadataSnapshot(profileRoot, "test", "node", host, "worker-a", output); err != nil {
		t.Fatal(err)
	}
	changed, err := RunNode(Options{ConfigDir: profileRoot, Benchmark: "test", Target: "node", MetadataFile: output}, "worker-a")
	if err != nil {
		t.Fatal(err)
	}
	want2 := map[string]string{"4.1.9": "PASS", "KUBECIS-NATIVE-NODE-1": "FAIL", "KUBECIS-NATIVE-NODE-2": "PASS"}
	for _, r := range changed.Results {
		if r.Status != want2[r.ID] {
			t.Fatalf("changed status not observed: %+v", r)
		}
		for _, old := range report.Results {
			if old.ID == r.ID && old.FindingID != r.FindingID {
				t.Fatalf("unstable finding identity %s", r.ID)
			}
		}
	}
	snap, err := readNodeMetadataSnapshot(output, "worker-a", "test", "node")
	if err != nil {
		t.Fatal(err)
	}
	snap.SchemaVersion = 1
	invalid, _ := json.Marshal(snap)
	if err := os.WriteFile(output, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readNodeMetadataSnapshot(output, "worker-a", "test", "node"); err == nil {
		t.Fatal("old schema should be rejected")
	}
	staleReport, err := RunNode(Options{ConfigDir: profileRoot, Benchmark: "test", Target: "node", MetadataFile: output}, "worker-a")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range staleReport.Results {
		if r.Status != "NOT_ASSESSED" {
			t.Fatalf("invalid snapshot falsely assessed: %+v", r)
		}
	}
	snap.SchemaVersion = nodeSnapshotSchemaVersion
	snap.GeneratedAt = time.Now().Add(-10 * time.Minute)
	invalid, _ = json.Marshal(snap)
	if err := os.WriteFile(output, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readNodeMetadataSnapshot(output, "worker-a", "test", "node"); err == nil {
		t.Fatal("stale snapshot accepted")
	}
	snap.GeneratedAt = time.Now()
	invalid, _ = json.Marshal(snap)
	if err := os.WriteFile(output, invalid, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readNodeMetadataSnapshot(output, "worker-b", "test", "node"); err == nil {
		t.Fatal("wrong node accepted")
	}
	if _, err := readNodeMetadataSnapshot(output, "worker-a", "different", "node"); err == nil {
		t.Fatal("wrong benchmark accepted")
	}
	if _, err := readNodeMetadataSnapshot(output, "worker-a", "test", "master"); err == nil {
		t.Fatal("wrong target accepted")
	}
}

func TestKubeletCollectorNeverFollowsSymlinks(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux collector")
	}
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.yaml")
	if err := os.WriteFile(outside, []byte("authentication:\n  anonymous:\n    enabled: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "config.yaml")); err != nil {
		t.Fatal(err)
	}
	for _, s := range collectKubeletSettings(dir) {
		if s.Error == "" || s.Value != "" {
			t.Fatalf("symlink followed: %+v", s)
		}
	}
}

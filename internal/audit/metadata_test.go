package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunNodeUsesMetadataSnapshot(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("node metadata helper is Linux-only")
	}
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	profileDir := filepath.Join(configDir, "cis-test")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	benchmark := `id: cis-test
name: node metadata test
checks:
  - id: 4.1.9
    title: kubelet config permissions
    target: node
    kind: host_file_mode
    path: var/lib/kubelet/config.yaml
    maxMode: "0600"
    scored: true
`
	if err := os.WriteFile(filepath.Join(profileDir, "benchmark.yaml"), []byte(benchmark), 0o600); err != nil {
		t.Fatal(err)
	}

	hostRoot := filepath.Join(root, "host-kubelet")
	if err := os.MkdirAll(hostRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(hostRoot, "config.yaml")
	if err := os.WriteFile(configPath, []byte("metadata-only fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(root, "metadata", "snapshot.json")
	if err := RunNodeMetadataSnapshot(configDir, "cis-test", hostRoot, "worker-a", metadataPath); err != nil {
		t.Fatalf("collect metadata snapshot: %v", err)
	}

	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", MetadataFile: metadataPath}, "worker-a")
	if err != nil {
		t.Fatalf("run node audit from metadata: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected one node check, got %d", len(report.Results))
	}
	if report.Results[0].Status != "FAIL" {
		t.Fatalf("expected 0644 to fail the fixture's default maxMode, got %s (%s)", report.Results[0].Status, report.Results[0].Details)
	}
}

func TestRunNodeRejectsMetadataForDifferentNode(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("node metadata helper is Linux-only")
	}
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	profileDir := filepath.Join(configDir, "cis-test")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	benchmark := `id: cis-test
name: node metadata test
checks:
  - id: 4.1.9
    title: kubelet config permissions
    target: node
    kind: host_file_mode
    path: var/lib/kubelet/config.yaml
    maxMode: "0600"
    scored: true
`
	if err := os.WriteFile(filepath.Join(profileDir, "benchmark.yaml"), []byte(benchmark), 0o600); err != nil {
		t.Fatal(err)
	}
	hostRoot := filepath.Join(root, "host-kubelet")
	if err := os.MkdirAll(hostRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hostRoot, "config.yaml"), []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(root, "snapshot.json")
	if err := RunNodeMetadataSnapshot(configDir, "cis-test", hostRoot, "worker-a", metadataPath); err != nil {
		t.Fatal(err)
	}
	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", MetadataFile: metadataPath}, "worker-b")
	if err != nil {
		t.Fatal(err)
	}
	if report.Results[0].Status != "NOT_ASSESSED" {
		t.Fatalf("node identity mismatch must not use another node's evidence: %#v", report.Results[0])
	}
}

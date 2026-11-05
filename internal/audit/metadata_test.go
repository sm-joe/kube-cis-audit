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
compatibilityStatus: supported
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
compatibilityStatus: supported
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

func TestControlPlaneMetadataAssessesAPIServerManifestMode(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("host metadata helper is Linux-only")
	}
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	profileDir := filepath.Join(configDir, "cis-test")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	benchmark := `id: cis-test
compatibilityStatus: supported
name: control-plane metadata test
checks:
  - id: 1.1.1
    title: API server manifest permissions
    target: master
    kind: host_file_mode
    path: etc/kubernetes/manifests/kube-apiserver.yaml
    maxMode: "0600"
    scored: true
`
	if err := os.WriteFile(filepath.Join(profileDir, "benchmark.yaml"), []byte(benchmark), 0o600); err != nil {
		t.Fatal(err)
	}
	hostRoot := filepath.Join(root, "host-manifests")
	if err := os.MkdirAll(hostRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(hostRoot, "kube-apiserver.yaml")
	if err := os.WriteFile(manifest, []byte("must never be read by collector"), 0o644); err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(root, "metadata", "snapshot.json")
	if err := RunHostMetadataSnapshot(configDir, "cis-test", "master", hostRoot, "control-plane-a", metadataPath); err != nil {
		t.Fatal(err)
	}
	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", Target: "master", MetadataFile: metadataPath}, "control-plane-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Results) != 1 || report.Results[0].ID != "1.1.1" {
		t.Fatalf("unexpected results: %#v", report.Results)
	}
	if report.Results[0].Status != "FAIL" {
		t.Fatalf("expected 0644 to fail, got %#v", report.Results[0])
	}
	if report.Compliance.ScoredTotal != 1 || report.Compliance.ScoredFail != 1 || report.Compliance.ScoredPass != 0 {
		t.Fatalf("unexpected master-local compliance aggregation for FAIL: %#v", report.Compliance)
	}
	if report.Compliance.CoveragePercent != 100 || report.Compliance.CompliancePercent == nil || *report.Compliance.CompliancePercent != 0 {
		t.Fatalf("unexpected master-local coverage/compliance for FAIL: %#v", report.Compliance)
	}
	if report.Compliance.ScoreStatus != "complete: all scored controls resolved to PASS or FAIL" {
		t.Fatalf("unexpected master-local score status: %q", report.Compliance.ScoreStatus)
	}

	// Recollect after changing the fixture to the compliant mode and verify
	// that the PASS is also reflected in the local compliance summary.
	if err := os.Chmod(manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RunHostMetadataSnapshot(configDir, "cis-test", "master", hostRoot, "control-plane-a", metadataPath); err != nil {
		t.Fatalf("refresh compliant metadata snapshot: %v", err)
	}
	passReport, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", Target: "master", MetadataFile: metadataPath}, "control-plane-a")
	if err != nil {
		t.Fatalf("run compliant control-plane audit: %v", err)
	}
	if len(passReport.Results) != 1 || passReport.Results[0].Status != "PASS" {
		t.Fatalf("expected 0600 to pass, got %#v", passReport.Results)
	}
	if passReport.Compliance.ScoredTotal != 1 || passReport.Compliance.ScoredPass != 1 || passReport.Compliance.ScoredFail != 0 {
		t.Fatalf("unexpected master-local compliance aggregation for PASS: %#v", passReport.Compliance)
	}
	if passReport.Compliance.CoveragePercent != 100 || passReport.Compliance.CompliancePercent == nil || *passReport.Compliance.CompliancePercent != 100 {
		t.Fatalf("unexpected master-local coverage/compliance for PASS: %#v", passReport.Compliance)
	}
	if report.AuditMode != "master-local" {
		t.Fatalf("unexpected audit mode %q", report.AuditMode)
	}
	if _, ok := approvedHostMetadataPath("master", "etc/kubernetes/manifests/../../shadow"); ok {
		t.Fatal("path traversal unexpectedly approved")
	}
}

func TestControlPlaneMetadataRejectsNodeTargetSnapshot(t *testing.T) {
	if _, ok := approvedHostMetadataPath("node", "etc/kubernetes/manifests/kube-apiserver.yaml"); ok {
		t.Fatal("control-plane path approved for node target")
	}
	if _, ok := approvedHostMetadataPath("master", "var/lib/kubelet/config.yaml"); ok {
		t.Fatal("kubelet path approved for control-plane target")
	}
}

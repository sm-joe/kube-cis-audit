package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeNodeBenchmark(t *testing.T, root string) {
	t.Helper()
	profileDir := filepath.Join(root, "cis-test")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data := `id: cis-test
name: node test profile
compatibilityStatus: test
checks:
  - id: 4.1.1
    title: node file exists
    target: node
    kind: host_file_mode
    path: var/lib/kubelet/config.yaml
    maxMode: "0600"
    scored: true
  - id: 1.1.1
    title: control plane file exists
    target: master
    kind: host_file_exists
    path: etc/kubernetes/manifests/kube-apiserver.yaml
    scored: true
  - id: 5.1.1
    title: API check must not run node-local
    target: policies
    kind: api_cluster_admin
    scored: true
`
	if err := os.WriteFile(filepath.Join(profileDir, "benchmark.yaml"), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRunNodeOnlyRunsNodeLocalChecksAndReportsNodeIdentity(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	writeNodeBenchmark(t, configDir)
	hostRoot := filepath.Join(root, "host")
	configPath := filepath.Join(hostRoot, "var", "lib", "kubelet", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("apiVersion: kubelet.config.k8s.io/v1beta1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", HostRoot: hostRoot}, "worker-a")
	if err != nil {
		t.Fatal(err)
	}
	if report.AuditMode != "node-local" || report.NodeName != "worker-a" {
		t.Fatalf("node identity missing from report: %#v", report)
	}
	if len(report.Results) != 1 || report.Results[0].ID != "4.1.1" {
		t.Fatalf("node-local run included unexpected checks: %#v", report.Results)
	}
	if report.Results[0].Status != "PASS" {
		t.Fatalf("got %s, want PASS: %s", report.Results[0].Status, report.Results[0].Details)
	}
}

func TestRunNodeMissingHostFileIsNotAssessed(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	writeNodeBenchmark(t, configDir)
	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", HostRoot: filepath.Join(root, "empty-host")}, "worker-b")
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.NotAssessed != 1 || report.Results[0].Status != "NOT_ASSESSED" {
		t.Fatalf("missing host evidence should be NOT_ASSESSED: %#v", report)
	}
	if !strings.Contains(report.Results[0].Details, "worker-b") {
		t.Fatalf("node name missing from inaccessible evidence: %s", report.Results[0].Details)
	}
}

func TestRunNodeRequiresExplicitBenchmarkAndNodeName(t *testing.T) {
	root := t.TempDir()
	if _, err := RunNode(Options{ConfigDir: root, Benchmark: "auto", HostRoot: root}, "worker-a"); err == nil {
		t.Fatal("expected explicit benchmark requirement")
	}
	if _, err := RunNode(Options{ConfigDir: root, Benchmark: "cis-test", HostRoot: root}, ""); err == nil {
		t.Fatal("expected node name requirement")
	}
}

func TestNodeHostCheckRejectsPathTraversal(t *testing.T) {
	r := evaluateNodeHostCheck(t.TempDir(), "worker-a", Check{ID: "4.1.1", Target: "node", Kind: "host_file_mode", Path: "../../etc/shadow", MaxMode: "0600"})
	if r.Status != "NOT_ASSESSED" || !strings.Contains(r.Details, "unsafe host path") {
		t.Fatalf("path traversal was not rejected: %#v", r)
	}
}

func TestNodeHostCheckDoesNotFollowSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	r := evaluateNodeHostCheck(root, "worker-a", Check{ID: "4.1.1", Target: "node", Kind: "host_file_mode", Path: "link", MaxMode: "0600"})
	if r.Status != "NOT_ASSESSED" || !strings.Contains(r.Details, "symbolic link") {
		t.Fatalf("symlink target should not be followed: %#v", r)
	}
}

func TestRunNodeUsesMetadataSnapshotWithoutHostRootAccess(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("metadata helper requires Linux POSIX file modes")
	}
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	writeNodeBenchmark(t, configDir)
	hostKubelet := filepath.Join(root, "host-kubelet")
	if err := os.MkdirAll(hostKubelet, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hostKubelet, "config.yaml"), []byte("metadata only test"), 0o644); err != nil {
		t.Fatal(err)
	}
	metadataFile := filepath.Join(root, "shared", "node-metadata.json")
	if err := RunNodeMetadataSnapshot(configDir, "cis-test", hostKubelet, "worker-metadata", metadataFile); err != nil {
		t.Fatal(err)
	}
	snapshotBytes, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(snapshotBytes), "metadata only test") {
		t.Fatal("metadata snapshot must never contain host file contents")
	}

	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", MetadataFile: metadataFile}, "worker-metadata")
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Fail != 1 || report.Results[0].Status != "FAIL" {
		t.Fatalf("0644 should fail the test profile's 0600 requirement: %#v", report.Results)
	}
	if len(report.Results[0].Evidence) != 1 || report.Results[0].Evidence[0].Observed != "0644" {
		t.Fatalf("expected permission metadata evidence, got %#v", report.Results[0].Evidence)
	}
}

func TestRunNodeMetadataSnapshotReportsMissingFileAsNotAssessed(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("metadata helper requires Linux")
	}
	root := t.TempDir()
	configDir := filepath.Join(root, "benchmarks")
	writeNodeBenchmark(t, configDir)
	metadataFile := filepath.Join(root, "shared", "node-metadata.json")
	if err := RunNodeMetadataSnapshot(configDir, "cis-test", filepath.Join(root, "empty-kubelet"), "worker-missing", metadataFile); err != nil {
		t.Fatal(err)
	}

	report, err := RunNode(Options{ConfigDir: configDir, Benchmark: "cis-test", MetadataFile: metadataFile}, "worker-missing")
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.NotAssessed != 1 || report.Results[0].Status != "NOT_ASSESSED" {
		t.Fatalf("missing host metadata should remain NOT_ASSESSED: %#v", report.Results)
	}
}

func TestNodeMetadataRejectsUnapprovedPath(t *testing.T) {
	if _, ok := approvedNodeMetadataPath("etc/shadow"); ok {
		t.Fatal("metadata helper must reject paths outside its explicit allowlist")
	}
	if got, ok := approvedNodeMetadataPath("var/lib/kubelet/config.yaml"); !ok || got != "config.yaml" {
		t.Fatalf("unexpected approved path mapping: %q, %v", got, ok)
	}
}

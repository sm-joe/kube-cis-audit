package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	if err := os.WriteFile(
		filepath.Join(profileDir, "benchmark.yaml"),
		[]byte(benchmark),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	hostRoot := filepath.Join(root, "host-kubelet")
	if err := os.MkdirAll(hostRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(hostRoot, "config.yaml")
	if err := os.WriteFile(
		configPath,
		[]byte("metadata-only fixture"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	metadataPath := filepath.Join(root, "metadata", "snapshot.json")

	if err := RunNodeMetadataSnapshot(
		configDir,
		"cis-test",
		hostRoot,
		"worker-a",
		metadataPath,
	); err != nil {
		t.Fatalf("collect metadata snapshot: %v", err)
	}

	report, err := RunNode(
		Options{
			ConfigDir:    configDir,
			Benchmark:    "cis-test",
			MetadataFile: metadataPath,
		},
		"worker-a",
	)
	if err != nil {
		t.Fatalf("run node audit from metadata: %v", err)
	}

	if len(report.Results) != 1 {
		t.Fatalf("expected one node check, got %d", len(report.Results))
	}

	if report.Results[0].Status != "FAIL" {
		t.Fatalf(
			"expected 0644 to fail the fixture's default maxMode, got %s (%s)",
			report.Results[0].Status,
			report.Results[0].Details,
		)
	}

	if report.AuditMode != "node-local" {
		t.Fatalf("expected node-local audit mode, got %q", report.AuditMode)
	}

	if report.Compliance.ScoredTotal != 1 ||
		report.Compliance.ScoredFail != 1 ||
		report.Compliance.ScoredPass != 0 {
		t.Fatalf(
			"unexpected node-local compliance aggregation: %#v",
			report.Compliance,
		)
	}

	assertLocalComplianceWithheld(t, report, 100)
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

	if err := os.WriteFile(
		filepath.Join(profileDir, "benchmark.yaml"),
		[]byte(benchmark),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	hostRoot := filepath.Join(root, "host-kubelet")
	if err := os.MkdirAll(hostRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(hostRoot, "config.yaml"),
		[]byte("fixture"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	metadataPath := filepath.Join(root, "snapshot.json")

	if err := RunNodeMetadataSnapshot(
		configDir,
		"cis-test",
		hostRoot,
		"worker-a",
		metadataPath,
	); err != nil {
		t.Fatal(err)
	}

	report, err := RunNode(
		Options{
			ConfigDir:    configDir,
			Benchmark:    "cis-test",
			MetadataFile: metadataPath,
		},
		"worker-b",
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Results) != 1 {
		t.Fatalf(
			"expected one node result, got %d",
			len(report.Results),
		)
	}

	if report.Results[0].Status != "NOT_ASSESSED" {
		t.Fatalf(
			"node identity mismatch must not use another node's evidence: %#v",
			report.Results[0],
		)
	}

	if report.Compliance.ScoredNotAssessed != 1 {
		t.Fatalf(
			"mismatched node evidence must remain NOT_ASSESSED: %#v",
			report.Compliance,
		)
	}

	assertLocalComplianceWithheld(t, report, 0)
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

	if err := os.WriteFile(
		filepath.Join(profileDir, "benchmark.yaml"),
		[]byte(benchmark),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	hostRoot := filepath.Join(root, "host-manifests")
	if err := os.MkdirAll(hostRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := filepath.Join(hostRoot, "kube-apiserver.yaml")

	if err := os.WriteFile(
		manifest,
		[]byte("must never be read by collector"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	metadataPath := filepath.Join(root, "metadata", "snapshot.json")

	// Scenario 1: manifest permissions are 0644.
	// Expected assessment result: FAIL.
	if err := RunHostMetadataSnapshot(
		configDir,
		"cis-test",
		"master",
		hostRoot,
		"control-plane-a",
		metadataPath,
	); err != nil {
		t.Fatal(err)
	}

	report, err := RunNode(
		Options{
			ConfigDir:    configDir,
			Benchmark:    "cis-test",
			Target:       "master",
			MetadataFile: metadataPath,
		},
		"control-plane-a",
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Results) != 1 || report.Results[0].ID != "1.1.1" {
		t.Fatalf("unexpected results: %#v", report.Results)
	}

	if report.Results[0].Status != "FAIL" {
		t.Fatalf(
			"expected 0644 to fail, got %#v",
			report.Results[0],
		)
	}

	if report.AuditMode != "master-local" {
		t.Fatalf("unexpected audit mode %q", report.AuditMode)
	}

	if report.Compliance.ScoredTotal != 1 ||
		report.Compliance.ScoredFail != 1 ||
		report.Compliance.ScoredPass != 0 {
		t.Fatalf(
			"unexpected master-local compliance aggregation for FAIL: %#v",
			report.Compliance,
		)
	}

	// All selected checks were assessed, so coverage is 100%.
	// However, a single control-plane check does not establish
	// complete cluster-wide CIS compliance.
	assertLocalComplianceWithheld(t, report, 100)

	// Scenario 2: change manifest permissions to 0600.
	// Expected assessment result: PASS.
	if err := os.Chmod(manifest, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RunHostMetadataSnapshot(
		configDir,
		"cis-test",
		"master",
		hostRoot,
		"control-plane-a",
		metadataPath,
	); err != nil {
		t.Fatalf("refresh compliant metadata snapshot: %v", err)
	}

	passReport, err := RunNode(
		Options{
			ConfigDir:    configDir,
			Benchmark:    "cis-test",
			Target:       "master",
			MetadataFile: metadataPath,
		},
		"control-plane-a",
	)
	if err != nil {
		t.Fatalf("run compliant control-plane audit: %v", err)
	}

	if len(passReport.Results) != 1 ||
		passReport.Results[0].Status != "PASS" {
		t.Fatalf(
			"expected 0600 to pass, got %#v",
			passReport.Results,
		)
	}

	if passReport.AuditMode != "master-local" {
		t.Fatalf(
			"expected master-local audit mode, got %q",
			passReport.AuditMode,
		)
	}

	if passReport.Compliance.ScoredTotal != 1 ||
		passReport.Compliance.ScoredPass != 1 ||
		passReport.Compliance.ScoredFail != 0 {
		t.Fatalf(
			"unexpected master-local compliance aggregation for PASS: %#v",
			passReport.Compliance,
		)
	}

	// A PASS is reflected in assessment coverage,
	// but it must not create a cluster-wide compliance score.
	assertLocalComplianceWithheld(t, passReport, 100)

	// Ensure unsafe traversal paths remain rejected.
	if _, ok := approvedHostMetadataPath(
		"master",
		"etc/kubernetes/manifests/../../shadow",
	); ok {
		t.Fatal("path traversal unexpectedly approved")
	}
}

func TestControlPlaneMetadataRejectsNodeTargetSnapshot(t *testing.T) {
	if _, ok := approvedHostMetadataPath(
		"node",
		"etc/kubernetes/manifests/kube-apiserver.yaml",
	); ok {
		t.Fatal("control-plane path approved for node target")
	}

	if _, ok := approvedHostMetadataPath(
		"master",
		"var/lib/kubelet/config.yaml",
	); ok {
		t.Fatal("kubelet path approved for control-plane target")
	}
}

// assertLocalComplianceWithheld verifies that local assessment
// coverage is calculated correctly while cluster-wide CIS
// compliance scoring remains withheld.
//
// Node-local and master-local reports assess only their selected
// host controls, not the complete Kubernetes CIS benchmark.
func assertLocalComplianceWithheld(
	t *testing.T,
	report Report,
	expectedCoverage float64,
) {
	t.Helper()

	if report.Compliance.CoveragePercent != expectedCoverage {
		t.Fatalf(
			"unexpected local coverage: expected %.2f, got %.2f",
			expectedCoverage,
			report.Compliance.CoveragePercent,
		)
	}

	if report.Compliance.CompliancePercent != nil {
		t.Fatalf(
			"%s report must not publish cluster-wide CIS compliance: %#v",
			report.AuditMode,
			report.Compliance,
		)
	}

	if !strings.HasPrefix(
		report.Compliance.ScoreStatus,
		"withheld:",
	) {
		t.Fatalf(
			"expected withheld local compliance score, got %q",
			report.Compliance.ScoreStatus,
		)
	}
}

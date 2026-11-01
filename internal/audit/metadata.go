package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// NodeMetadataSnapshot contains metadata only; it never contains host file contents.
type NodeMetadataSnapshot struct {
	NodeName    string                      `json:"nodeName"`
	GeneratedAt time.Time                   `json:"generatedAt"`
	Files       map[string]NodeFileMetadata `json:"files"`
}

type NodeFileMetadata struct {
	Mode    uint32 `json:"mode,omitempty"`
	Regular bool   `json:"regular"`
	Symlink bool   `json:"symlink"`
	Error   string `json:"error,omitempty"`
}

const nodeMetadataMaxAge = 2 * time.Minute

// RunNodeMetadataSnapshot inspects only explicitly approved node paths and
// atomically publishes their metadata for the unprivileged auditor container.
func RunNodeMetadataSnapshot(configDir, benchmark, hostRoot, nodeName, outputFile string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("node metadata helper requires Linux; current OS is %s", runtime.GOOS)
	}
	if strings.TrimSpace(nodeName) == "" {
		return fmt.Errorf("node name is required")
	}
	if strings.TrimSpace(hostRoot) == "" || strings.TrimSpace(outputFile) == "" {
		return fmt.Errorf("host root and output file are required")
	}
	bench, err := readSelectedBenchmark(configDir, benchmark)
	if err != nil {
		return err
	}
	root, err := filepath.Abs(hostRoot)
	if err != nil {
		return fmt.Errorf("resolve metadata host root: %w", err)
	}

	snapshot := NodeMetadataSnapshot{
		NodeName:    nodeName,
		GeneratedAt: time.Now().UTC(),
		Files:       make(map[string]NodeFileMetadata),
	}
	for _, check := range bench.Checks {
		if !strings.EqualFold(check.Target, "node") || (check.Kind != "host_file_mode" && check.Kind != "host_file_exists") {
			continue
		}
		metadata := NodeFileMetadata{}
		relativePath, ok := approvedNodeMetadataPath(check.Path)
		if !ok {
			metadata.Error = "path is not approved for metadata inspection"
			snapshot.Files[check.Path] = metadata
			continue
		}
		fullPath := filepath.Join(root, relativePath)
		info, statErr := os.Lstat(fullPath)
		if statErr != nil {
			metadata.Error = statErr.Error()
			snapshot.Files[check.Path] = metadata
			continue
		}
		metadata.Symlink = info.Mode()&os.ModeSymlink != 0
		metadata.Regular = info.Mode().IsRegular()
		metadata.Mode = uint32(info.Mode().Perm())
		snapshot.Files[check.Path] = metadata
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode node metadata: %w", err)
	}
	outputPath, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("resolve metadata output path: %w", err)
	}
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create metadata output directory: %w", err)
	}
	temp, err := os.CreateTemp(outputDir, ".kubecis-metadata-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary metadata file: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return fmt.Errorf("set metadata file permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write metadata file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync metadata file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close metadata file: %w", err)
	}
	if err := os.Rename(tempName, outputPath); err != nil {
		return fmt.Errorf("publish metadata file: %w", err)
	}
	return nil
}

// approvedNodeMetadataPath maps benchmark paths to paths beneath the dedicated
// /var/lib/kubelet hostPath mount. Extend deliberately when adding new checks.
func approvedNodeMetadataPath(benchmarkPath string) (string, bool) {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(benchmarkPath)))
	if clean != "var/lib/kubelet/config.yaml" {
		return "", false
	}
	return "config.yaml", true
}

func readNodeMetadataSnapshot(path, expectedNode string) (NodeMetadataSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NodeMetadataSnapshot{}, fmt.Errorf("node metadata unavailable: %w", err)
	}
	var snapshot NodeMetadataSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return NodeMetadataSnapshot{}, fmt.Errorf("decode node metadata: %w", err)
	}
	if snapshot.NodeName != expectedNode {
		return NodeMetadataSnapshot{}, fmt.Errorf("node metadata identity mismatch: got %q, expected %q", snapshot.NodeName, expectedNode)
	}
	age := time.Since(snapshot.GeneratedAt)
	if snapshot.GeneratedAt.IsZero() || age > nodeMetadataMaxAge || age < -time.Minute {
		return NodeMetadataSnapshot{}, fmt.Errorf("node metadata is stale or has an invalid timestamp (age %s)", age)
	}
	if snapshot.Files == nil {
		return NodeMetadataSnapshot{}, fmt.Errorf("node metadata contains no file records")
	}
	return snapshot, nil
}

func evaluateNodeHostCheckFromMetadata(nodeName string, c Check, snapshot NodeMetadataSnapshot) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Remediation: c.Remediation, Scored: c.Scored, Evidence: []Evidence{{Resource: "node", Name: nodeName, Field: "identity", Observed: nodeName, Expected: "Kubernetes node being assessed"}}}
	if _, ok := approvedNodeMetadataPath(c.Path); !ok {
		r.Details = fmt.Sprintf("Host path %q is not approved for metadata inspection.", c.Path)
		return r
	}
	metadata, ok := snapshot.Files[c.Path]
	if !ok {
		r.Details = fmt.Sprintf("Node %s: metadata helper did not report %s.", nodeName, c.Path)
		return r
	}
	if metadata.Error != "" {
		r.Details = fmt.Sprintf("Node %s: host metadata unavailable at %s: %s. This is an access or collection limitation, not a control failure.", nodeName, c.Path, metadata.Error)
		r.Evidence = append(r.Evidence, Evidence{Resource: "host-file", Name: c.Path, Field: "metadata", Observed: "unavailable", Expected: "file metadata"})
		return r
	}
	if metadata.Symlink {
		r.Details = fmt.Sprintf("Host path %s is a symbolic link; target inspection is intentionally not followed.", c.Path)
		r.Evidence = append(r.Evidence, Evidence{Resource: "host-file", Name: c.Path, Field: "type", Observed: "symlink", Expected: "regular file"})
		return r
	}
	if !metadata.Regular {
		r.Details = fmt.Sprintf("Host path %s is not a regular file.", c.Path)
		r.Evidence = append(r.Evidence, Evidence{Resource: "host-file", Name: c.Path, Field: "type", Observed: "non-regular file", Expected: "regular file"})
		return r
	}
	if c.Kind == "host_file_exists" {
		r.Status = "PASS"
		r.Details = "Host configuration file exists: " + c.Path
		r.Evidence = append(r.Evidence, Evidence{Resource: "host-file", Name: c.Path, Observed: "present", Expected: "present"})
		return r
	}
	if metadata.Mode > 0o777 {
		r.Details = fmt.Sprintf("Metadata helper returned invalid permission mode %04o.", metadata.Mode)
		return r
	}
	return evaluateHostFileMode(os.FileMode(metadata.Mode), c, r)
}

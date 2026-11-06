package audit

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// NodeMetadataSnapshot contains allowlisted file metadata and two validated kubelet settings.
// It never includes raw host-file content, secrets, or arbitrary configuration values.
type NodeMetadataSnapshot struct {
	SchemaVersion   int                         `json:"schemaVersion"`
	Benchmark       string                      `json:"benchmark"`
	Target          string                      `json:"target"`
	KubeletSettings map[string]KubeletSetting   `json:"kubeletSettings,omitempty"`
	NodeName        string                      `json:"nodeName"`
	GeneratedAt     time.Time                   `json:"generatedAt"`
	Files           map[string]NodeFileMetadata `json:"files"`
}

// KubeletSetting is bounded, allowlisted, evaluated configuration evidence.
// Only explicit bool(false/true) and a known authorization mode are exported.
type KubeletSetting struct {
	SourcePath string `json:"sourcePath,omitempty"`
	Value      string `json:"value,omitempty"`
	Error      string `json:"error,omitempty"`
}

const nodeSnapshotSchemaVersion = 4

type NodeFileMetadata struct {
	SourcePath string  `json:"sourcePath,omitempty"`
	Mode       uint32  `json:"mode,omitempty"`
	OwnerUID   *uint32 `json:"ownerUid,omitempty"`
	OwnerGID   *uint32 `json:"ownerGid,omitempty"`
	Regular    bool    `json:"regular"`
	Symlink    bool    `json:"symlink"`
	Error      string  `json:"error,omitempty"`
}

const nodeMetadataMaxAge = 2 * time.Minute

// RunNodeMetadataSnapshot inspects only explicitly approved node paths and
// atomically publishes their metadata for the unprivileged auditor container.
func RunNodeMetadataSnapshot(configDir, benchmark, hostRoot, nodeName, outputFile string) error {
	return RunHostMetadataSnapshot(configDir, benchmark, "node", hostRoot, nodeName, outputFile)
}

// RunHostMetadataSnapshot collects metadata for one explicitly selected host target.
// The caller must mount only the target's dedicated host directory read-only.
func RunHostMetadataSnapshot(configDir, benchmark, target, hostRoot, nodeName, outputFile string) error {
	return RunHostMetadataSnapshotWithKubeletRoot(configDir, benchmark, target, hostRoot, "", nodeName, outputFile)
}

// RunHostMetadataSnapshotWithKubeletRoot supports AL2023 nodeadm kubelet config.json
// via a separate narrowly scoped, read-only hostPath mount. The legacy kubeadm
// path is retained for Kind and other supported configurations.
func RunHostMetadataSnapshotWithKubeletRoot(configDir, benchmark, target, hostRoot, kubeletConfigRoot, nodeName, outputFile string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("node metadata helper requires Linux; current OS is %s", runtime.GOOS)
	}
	if strings.TrimSpace(nodeName) == "" {
		return fmt.Errorf("node name is required")
	}
	if strings.TrimSpace(hostRoot) == "" || strings.TrimSpace(outputFile) == "" {
		return fmt.Errorf("host root and output file are required")
	}
	if !strings.EqualFold(target, "node") && !strings.EqualFold(target, "master") {
		return fmt.Errorf("unsupported host metadata target %q (use node or master)", target)
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
		SchemaVersion: nodeSnapshotSchemaVersion,
		Benchmark:     benchmark,
		Target:        strings.ToLower(target),
		NodeName:      nodeName,
		GeneratedAt:   time.Now().UTC(),
		Files:         make(map[string]NodeFileMetadata),
	}
	for _, check := range bench.Checks {
		if !strings.EqualFold(check.Target, target) || (check.Kind != "host_file_mode" && check.Kind != "host_file_exists" && check.Kind != "host_file_owner") {
			continue
		}
		metadata := NodeFileMetadata{}
		relativePath, ok := approvedHostMetadataPath(target, check.Path)
		if !ok {
			metadata.Error = "path is not approved for metadata inspection"
			snapshot.Files[check.Path] = metadata
			continue
		}
		fullPath := filepath.Join(root, relativePath)
		if strings.EqualFold(target, "node") && (check.Path == "var/lib/kubelet/config.yaml" || check.Path == "etc/kubernetes/kubelet/config.json") && kubeletConfigRoot != "" {
			fullPath = filepath.Join(kubeletConfigRoot, "config.json")
			metadata.SourcePath = "etc/kubernetes/kubelet/config.json"
		} else {
			metadata.SourcePath = check.Path
		}
		info, statErr := os.Lstat(fullPath)
		if statErr != nil {
			metadata.Error = statErr.Error()
			snapshot.Files[check.Path] = metadata
			continue
		}
		metadata.Symlink = info.Mode()&os.ModeSymlink != 0
		metadata.Regular = info.Mode().IsRegular()
		metadata.Mode = uint32(info.Mode().Perm())
		if uid, gid, ok := hostFileOwner(info); ok {
			metadata.OwnerUID = &uid
			metadata.OwnerGID = &gid
		}
		snapshot.Files[check.Path] = metadata
	}

	// Read only a mounted kubelet config: kubeadm YAML or explicit AL2023 JSON.
	// Do not export arbitrary configuration contents.
	// Control-plane metadata mode never accesses kubelet configuration.
	if strings.EqualFold(target, "node") {
		if kubeletConfigRoot != "" {
			snapshot.KubeletSettings = collectKubeletSettingsFrom(kubeletConfigRoot, "config.json", "etc/kubernetes/kubelet/config.json")
		} else {
			snapshot.KubeletSettings = collectKubeletSettings(root)
		}
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
	return approvedHostMetadataPath("node", benchmarkPath)
}

// approvedHostMetadataPath is a strict per-target allowlist. The control-plane
// collector sees only the static-pod manifest directory, never the host root.
func approvedHostMetadataPath(target, benchmarkPath string) (string, bool) {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(benchmarkPath)))
	switch strings.ToLower(target) {
	case "node":
		if clean == "var/lib/kubelet/kubeconfig" {
			return "kubeconfig", true
		}
		if clean == "var/lib/kubelet/config.yaml" {
			return "config.yaml", true
		}
		if clean == "etc/kubernetes/kubelet/config.json" {
			return "config.json", true
		}
	case "master", "controlplane", "control-plane":
		if clean == "etc/kubernetes/manifests/kube-apiserver.yaml" {
			return "kube-apiserver.yaml", true
		}
	}
	return "", false
}

func readNodeMetadataSnapshot(path, expectedNode, expectedBenchmark, expectedTarget string) (NodeMetadataSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NodeMetadataSnapshot{}, fmt.Errorf("node metadata unavailable: %w", err)
	}
	var snapshot NodeMetadataSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return NodeMetadataSnapshot{}, fmt.Errorf("decode node metadata: %w", err)
	}
	if snapshot.SchemaVersion != nodeSnapshotSchemaVersion {
		return NodeMetadataSnapshot{}, fmt.Errorf("incompatible node metadata schema %d (expected %d); update metadata helper and auditor together", snapshot.SchemaVersion, nodeSnapshotSchemaVersion)
	}
	if snapshot.Benchmark != expectedBenchmark || snapshot.Target != strings.ToLower(expectedTarget) {
		return NodeMetadataSnapshot{}, fmt.Errorf("node metadata benchmark/target mismatch")
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

func evaluateNodeHostCheckFromMetadata(nodeName string, target string, c Check, snapshot NodeMetadataSnapshot) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Remediation: c.Remediation, Scored: c.Scored, Evidence: []Evidence{{Resource: "node", Name: nodeName, Field: "identity", Observed: nodeName, Expected: "Kubernetes node being assessed"}}}
	if _, ok := approvedHostMetadataPath(target, c.Path); !ok {
		r.Details = fmt.Sprintf("Host path %q is not approved for metadata inspection.", c.Path)
		return r
	}
	metadata, ok := snapshot.Files[c.Path]
	if !ok {
		r.Details = fmt.Sprintf("Node %s: metadata helper did not report %s.", nodeName, c.Path)
		return r
	}
	if metadata.SourcePath != "" && metadata.SourcePath != c.Path && metadata.SourcePath != "etc/kubernetes/kubelet/config.json" {
		r.Details = "Metadata source path is not approved for this check"
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
	if c.Kind == "host_file_owner" {
		if metadata.OwnerUID == nil || metadata.OwnerGID == nil {
			r.Details = "Node owner metadata unavailable; ownership cannot be inferred from permission bits."
			return r
		}
		r.Status = "FAIL"
		if *metadata.OwnerUID == 0 && *metadata.OwnerGID == 0 {
			r.Status = "PASS"
		}
		r.Details = fmt.Sprintf("Node %s: host file %s ownership is %d:%d; expected root:root (0:0).", nodeName, metadata.SourcePath, *metadata.OwnerUID, *metadata.OwnerGID)
		r.Evidence = append(r.Evidence, Evidence{Resource: "host-file", Name: metadata.SourcePath, Field: "owner", Observed: fmt.Sprintf("%d:%d", *metadata.OwnerUID, *metadata.OwnerGID), Expected: "0:0"})
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
	result := evaluateHostFileMode(os.FileMode(metadata.Mode), c, r)
	if metadata.SourcePath != "" {
		for i := range result.Evidence {
			if result.Evidence[i].Resource == "host-file" {
				result.Evidence[i].Name = metadata.SourcePath
			}
		}
	}
	return result
}

// collectKubeletSettings reads only the explicitly mounted kubelet config file.
// Never serialize arbitrary YAML content or parser errors containing source text.
func collectKubeletSettings(root string) map[string]KubeletSetting {
	return collectKubeletSettingsFrom(root, "config.yaml", "var/lib/kubelet/config.yaml")
}

func collectKubeletSettingsFrom(root, filename, sourcePath string) map[string]KubeletSetting {
	names := []string{"authentication.anonymous.enabled", "authorization.mode", "authentication.x509.clientCAFile", "readOnlyPort", "makeIPTablesUtilChains", "eventRecordQPS", "rotateCertificates", "serverTLSBootstrap", "featureGates.RotateKubeletServerCertificate"}
	result := make(map[string]KubeletSetting, len(names))
	fail := func(reason string) map[string]KubeletSetting {
		for _, key := range names {
			result[key] = KubeletSetting{Error: reason}
		}
		return result
	}
	if (filename != "config.yaml" && filename != "config.json") || (sourcePath != "var/lib/kubelet/config.yaml" && sourcePath != "etc/kubernetes/kubelet/config.json") {
		return fail("kubelet configuration source not allowlisted")
	}
	path := filepath.Join(root, filename)
	info, err := os.Lstat(path)
	if err != nil {
		return fail("kubelet configuration unavailable")
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fail("kubelet configuration must be a regular non-symlinked file")
	}
	if info.Size() > 1024*1024 {
		return fail("kubelet configuration exceeds 1 MiB")
	}
	file, err := os.Open(path)
	if err != nil {
		return fail("kubelet configuration unreadable")
	}
	defer file.Close()
	limited, err := io.ReadAll(io.LimitReader(file, 1024*1024+1))
	if err != nil || len(limited) > 1024*1024 {
		return fail("kubelet configuration unreadable or too large")
	}
	var cfg map[string]interface{}
	if err := yaml.Unmarshal(limited, &cfg); err != nil {
		return fail("kubelet configuration is invalid YAML")
	}
	auth, ok := cfg["authentication"].(map[string]interface{})
	if !ok {
		result[names[0]] = KubeletSetting{Error: "authentication block not explicit"}
	} else {
		anonymous, ok := auth["anonymous"].(map[string]interface{})
		if !ok {
			result[names[0]] = KubeletSetting{Error: "anonymous authentication block not explicit"}
		} else {
			enabled, ok := anonymous["enabled"].(bool)
			if !ok {
				result[names[0]] = KubeletSetting{Error: "anonymous.enabled not an explicit boolean"}
			} else {
				result[names[0]] = KubeletSetting{Value: fmt.Sprint(enabled)}
			}
		}
	}
	authorization, ok := cfg["authorization"].(map[string]interface{})
	if !ok {
		result[names[1]] = KubeletSetting{Error: "authorization block not explicit"}
	} else {
		mode, ok := authorization["mode"].(string)
		if !ok || strings.TrimSpace(mode) == "" {
			result[names[1]] = KubeletSetting{Error: "authorization.mode not explicit"}
		} else {
			switch mode {
			case "Webhook", "AlwaysAllow", "AlwaysDeny", "ABAC", "RBAC", "Node":
				result[names[1]] = KubeletSetting{Value: mode}
			default:
				result[names[1]] = KubeletSetting{Error: "authorization.mode is not a recognized allowlisted mode"}
			}
		}
	}
	collectExtendedKubeletSettings(cfg, result)
	for key, setting := range result {
		setting.SourcePath = sourcePath
		result[key] = setting
	}
	return result
}

// evaluateKubeletSettingFromMetadata uses only checked, fresh snapshot evidence.
func evaluateKubeletSettingFromMetadata(nodeName string, c Check, snapshot NodeMetadataSnapshot) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Scored: c.Scored, Remediation: c.Remediation}
	if c.ID >= "3.2.3" && c.ID <= "3.2.8" {
		return evaluateExtendedKubeletSetting(nodeName, c, snapshot)
	}
	if c.Path != "var/lib/kubelet/config.yaml" || (c.Field != "authentication.anonymous.enabled" && c.Field != "authorization.mode") {
		r.Details = "Unapproved kubelet configuration path or field"
		return r
	}
	setting, ok := snapshot.KubeletSettings[c.Field]
	if ok && setting.SourcePath != "" && setting.SourcePath != c.Path && setting.SourcePath != "etc/kubernetes/kubelet/config.json" {
		r.Details = "Kubelet setting source path is not allowlisted"
		return r
	}
	if !ok || setting.Error != "" || setting.Value == "" {
		r.Details = "Kubelet configuration evidence unavailable"
		if setting.Error != "" {
			r.Details += ": " + setting.Error
		}
		return r
	}
	expected := "false"
	if c.Field == "authorization.mode" {
		expected = "Webhook"
	}
	if c.Field == "authentication.anonymous.enabled" && setting.Value != "true" && setting.Value != "false" {
		r.Details = "Invalid anonymous authentication evidence"
		return r
	}
	if c.Field == "authorization.mode" && setting.Value != "Webhook" && setting.Value != "AlwaysAllow" && setting.Value != "AlwaysDeny" && setting.Value != "ABAC" && setting.Value != "RBAC" && setting.Value != "Node" {
		r.Details = "Invalid kubelet authorization evidence"
		return r
	}
	r.Status = "FAIL"
	if setting.Value == expected {
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf("Evaluated kubelet configuration field %s on node %s", c.Field, nodeName)
	source := c.Path
	if setting.SourcePath != "" {
		source = setting.SourcePath
	}
	r.Evidence = []Evidence{{Resource: "host-file", Name: source, Field: c.Field, Observed: setting.Value, Expected: expected}}
	return r
}

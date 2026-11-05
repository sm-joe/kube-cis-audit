// Package node evaluates bounded node-file metadata evidence. It never reads
// host files, executes commands, or follows symlinks. Evidence must be supplied
// by the existing trusted node-metadata helper.
package node

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sm-joe/kube-cis-audit/adapter"
)

const (
	EvidenceKind      = "node_file_metadata"
	KubeletConfigPath = "var/lib/kubelet/config.yaml"
	MaxSnapshotAge    = 2 * time.Minute
)

// Snapshot mirrors the metadata-only payload emitted by KubeCIS's existing
// node helper. It intentionally has no file-content, UID, or GID fields.
type Snapshot struct {
	NodeName    string                  `json:"nodeName"`
	GeneratedAt time.Time               `json:"generatedAt"`
	Files       map[string]FileMetadata `json:"files"`
}

type FileMetadata struct {
	Mode    uint32 `json:"mode,omitempty"`
	Regular bool   `json:"regular"`
	Symlink bool   `json:"symlink"`
	Error   string `json:"error,omitempty"`
}

// ApprovedPaths is deliberately exact-match. Do not replace with prefix or
// path-cleaning logic: a caller cannot expand the helper's inspection scope.
var ApprovedPaths = map[string]struct{}{
	KubeletConfigPath: {},
}

// EvidenceFromSnapshot validates identity, freshness, and exact path before
// creating B2 evidence. It never interprets upstream audit strings as paths.
func EvidenceFromSnapshot(snapshot Snapshot, expectedNode, path string, now time.Time) (adapter.Evidence, error) {
	if _, ok := ApprovedPaths[path]; !ok {
		return adapter.Evidence{}, fmt.Errorf("node metadata path %q is not allowlisted", path)
	}
	if strings.TrimSpace(expectedNode) == "" || snapshot.NodeName != expectedNode {
		return adapter.Evidence{}, fmt.Errorf("node metadata identity mismatch")
	}
	age := now.Sub(snapshot.GeneratedAt)
	if snapshot.GeneratedAt.IsZero() || age > MaxSnapshotAge || age < -time.Minute {
		return adapter.Evidence{}, fmt.Errorf("node metadata snapshot is stale or has invalid timestamp (age %s)", age)
	}
	if snapshot.Files == nil {
		return adapter.Evidence{}, fmt.Errorf("node metadata snapshot contains no file records")
	}
	m, ok := snapshot.Files[path]
	if !ok || m.Error != "" {
		detail := "metadata record is missing"
		if ok && m.Error != "" {
			detail = m.Error
		}
		return adapter.Evidence{Kind: EvidenceKind, Key: path, Found: false, Detail: detail}, nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return adapter.Evidence{}, fmt.Errorf("encode metadata evidence: %w", err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return adapter.Evidence{}, fmt.Errorf("decode metadata evidence: %w", err)
	}
	return adapter.Evidence{Kind: EvidenceKind, Key: path, Value: value, Found: true, Detail: "metadata-only evidence from trusted node helper"}, nil
}

// ModeEvaluator evaluates only the POSIX permission-bit maximum comparator.
// A symlink, non-regular file, invalid metadata, or unsupported comparator is
// NOT_ASSESSED; collection limitations are never converted into FAIL.
type ModeEvaluator struct{}

func (ModeEvaluator) Evaluate(_ context.Context, req adapter.EvaluatorRequest) adapter.Evaluation {
	if req.Target != "node" || req.Evidence.Kind != EvidenceKind {
		return notAssessed("evidence target or kind does not match node metadata evaluator")
	}
	if _, ok := ApprovedPaths[req.Evidence.Key]; !ok {
		return notAssessed("evidence path is not allowlisted")
	}
	if req.Compare.Operator != "bitmask" {
		return notAssessed("only the explicitly supported bitmask comparator is implemented")
	}
	expectedText, ok := req.Compare.Value.(string)
	if !ok || expectedText == "" {
		return notAssessed("expected maximum mode must be an octal string")
	}
	expected, err := strconv.ParseUint(expectedText, 8, 32)
	if err != nil || expected > 0o777 {
		return notAssessed("expected maximum mode is invalid")
	}

	fields, ok := req.Evidence.Value.(map[string]any)
	if !ok {
		return notAssessed("metadata evidence has an unexpected representation")
	}
	regular, rok := fields["regular"].(bool)
	symlink, sok := fields["symlink"].(bool)
	modeValue, mok := numeric(fields["mode"])
	if !rok || !sok || !mok || modeValue > 0o777 {
		return notAssessed("metadata evidence is incomplete or invalid")
	}
	if symlink {
		return notAssessed("path is a symbolic link; target inspection is not followed")
	}
	if !regular {
		return notAssessed("path is not a regular file")
	}
	observed := uint64(modeValue)
	if observed & ^expected != 0 {
		return adapter.Evaluation{Status: adapter.StatusFail, Reason: "file permissions exceed the configured maximum", Details: fmt.Sprintf("observed=%04o expected_maximum=%04o", observed, expected)}
	}
	return adapter.Evaluation{Status: adapter.StatusPass, Reason: "file permission bits satisfy the configured maximum", Details: fmt.Sprintf("observed=%04o expected_maximum=%04o", observed, expected)}
}

func numeric(v any) (uint64, bool) {
	switch n := v.(type) {
	case float64:
		if n < 0 || n != float64(uint64(n)) {
			return 0, false
		}
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	default:
		return 0, false
	}
}
func notAssessed(reason string) adapter.Evaluation {
	return adapter.Evaluation{Status: adapter.StatusNotAssessed, Reason: reason}
}

// NewAdapter wires the single deliberately bounded metadata evaluator and its
// exact mapping. The mapping is for the existing KubeCIS node control contract;
// raw upstream flag/audit strings are not transformed into paths here.
func NewAdapter() (*adapter.Adapter, error) {
	registry := adapter.NewRegistry()
	if err := registry.Register("node-file-mode-v1", ModeEvaluator{}); err != nil {
		return nil, err
	}
	mappings := []adapter.Mapping{{
		ControlID: "4.1.9", Target: "node", SourceKind: "path", SourceValue: KubeletConfigPath,
		ComparatorOps: []string{"bitmask"}, EvaluatorID: "node-file-mode-v1",
		EvidenceKind: EvidenceKind, AllowEvaluation: true,
	}}
	return adapter.New(mappings, registry)
}

// Descriptor constructs a typed descriptor only for the known local contract.
// Callers must not use this helper to translate arbitrary upstream audit text.
func Descriptor(maxMode string) adapter.TestDescriptor {
	return adapter.TestDescriptor{ControlID: "4.1.9", Target: "node", Source: adapter.Source{Kind: "path", Value: KubeletConfigPath}, Compare: adapter.Comparator{Operator: "bitmask", Value: maxMode}}
}

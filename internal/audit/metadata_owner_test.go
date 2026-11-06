package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNodeMetadataOwnerSourceAndFailureSemantics(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux ownership metadata")
	}
	dir := t.TempDir()
	config := filepath.Join(dir, "kubeconfig")
	if err := os.WriteFile(config, []byte("safe-test-placeholder"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(config)
	if err != nil {
		t.Fatal(err)
	}
	uid, gid, ok := hostFileOwner(info)
	if !ok {
		t.Fatal("Linux stat did not provide ownership")
	}
	check := Check{ID: "3.1.2", Target: "node", Path: "var/lib/kubelet/kubeconfig", Kind: "host_file_owner", Scored: true}
	snapshot := NodeMetadataSnapshot{Files: map[string]NodeFileMetadata{
		check.Path: {
			SourcePath: check.Path, Regular: true, Mode: uint32(info.Mode().Perm()), OwnerUID: &uid, OwnerGID: &gid,
		},
	}}
	r := evaluateNodeHostCheckFromMetadata("worker-a", "node", check, snapshot)
	if uid == 0 && gid == 0 {
		if r.Status != "PASS" {
			t.Fatalf("root:root expected PASS, got %+v", r)
		}
	} else if r.Status != "FAIL" {
		t.Fatalf("non-root ownership expected FAIL, got %+v", r)
	}
	delete(snapshot.Files, check.Path)
	r = evaluateNodeHostCheckFromMetadata("worker-a", "node", check, snapshot)
	if r.Status != "NOT_ASSESSED" {
		t.Fatalf("missing file metadata must remain unknown, got %+v", r)
	}
}

func TestNodeOwnershipRejectsMissingUIDGIDAndSymlinks(t *testing.T) {
	check := Check{ID: "3.1.4", Target: "node", Path: "etc/kubernetes/kubelet/config.json", Kind: "host_file_owner"}
	snapshot := NodeMetadataSnapshot{Files: map[string]NodeFileMetadata{
		check.Path: {SourcePath: check.Path, Regular: true, Mode: 0644},
	}}
	if r := evaluateNodeHostCheckFromMetadata("worker-a", "node", check, snapshot); r.Status != "NOT_ASSESSED" {
		t.Fatalf("absent owner evidence should be unknown, got %+v", r)
	}
	snapshot.Files[check.Path] = NodeFileMetadata{SourcePath: check.Path, Symlink: true}
	if r := evaluateNodeHostCheckFromMetadata("worker-a", "node", check, snapshot); r.Status != "NOT_ASSESSED" {
		t.Fatalf("symlink should not be followed, got %+v", r)
	}
}

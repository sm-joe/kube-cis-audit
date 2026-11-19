package audit

import (
	"strings"
	"testing"
)

// The source benchmark's 3.1.1/3.1.2/3.1.3/3.1.4 require a real
// kubeconfig/config.json file with controlled mode and root:root ownership.
// No missing file, symlink or unknown UID/GID may become PASS.
func TestEKS210WorkerHostFileControlsPositiveNegativeAndMissing(t *testing.T) {
	zero, other := uint32(0), uint32(1000)
	cases := []struct {
		id, kind, path, maxMode string
	}{
		{"3.1.1", "host_file_mode", "var/lib/kubelet/kubeconfig", "0644"},
		{"3.1.2", "host_file_owner", "var/lib/kubelet/kubeconfig", ""},
		{"3.1.3", "host_file_mode", "etc/kubernetes/kubelet/config.json", "0644"},
		{"3.1.4", "host_file_owner", "etc/kubernetes/kubelet/config.json", ""},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			check := Check{ID: tc.id, Kind: tc.kind, Target: "node", Path: tc.path, MaxMode: tc.maxMode, Scored: true}
			m := NodeFileMetadata{SourcePath: tc.path, Regular: true, Mode: 0600, OwnerUID: &zero, OwnerGID: &zero}
			snapshot := NodeMetadataSnapshot{Files: map[string]NodeFileMetadata{tc.path: m}}
			assert := func(want string) {
				t.Helper()
				r := evaluateNodeHostCheckFromMetadata("worker", "node", check, snapshot)
				if r.Status != want || r.ID != tc.id {
					t.Fatalf("%s expected %s, got %+v", tc.id, want, r)
				}
			}
			assert("PASS")
			if tc.kind == "host_file_owner" {
				m.OwnerUID = &other
			} else {
				m.Mode = 0666
			}
			snapshot.Files[tc.path] = m
			assert("FAIL")
			m.OwnerUID = nil
			m.OwnerGID = nil
			if tc.kind == "host_file_mode" {
				m.Mode = 0600
			}
			snapshot.Files[tc.path] = m
			if tc.kind == "host_file_owner" {
				assert("NOT_ASSESSED")
			} else {
				assert("PASS")
			}
			m.Symlink = true
			snapshot.Files[tc.path] = m
			assert("NOT_ASSESSED")
			delete(snapshot.Files, tc.path)
			assert("NOT_ASSESSED")
			m.Symlink = false
			m.Error = "node metadata source unavailable"
			snapshot.Files[tc.path] = m
			r := evaluateNodeHostCheckFromMetadata("worker", "node", check, snapshot)
			if r.Status != "NOT_ASSESSED" || !strings.Contains(r.Details, "unavailable") {
				t.Fatalf("missing metadata must be explainable and undecided: %+v", r)
			}
		})
	}
}

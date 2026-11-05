package node

import (
	"context"
	"testing"
	"time"

	"github.com/sm-joe/kube-cis-audit/adapter"
)

func TestEvidenceFromSnapshotValidatesBoundary(t *testing.T) {
	now := time.Now().UTC()
	s := Snapshot{NodeName: "kind-control-plane", GeneratedAt: now, Files: map[string]FileMetadata{KubeletConfigPath: {Mode: 0o600, Regular: true}}}
	ev, err := EvidenceFromSnapshot(s, "kind-control-plane", KubeletConfigPath, now)
	if err != nil || !ev.Found || ev.Kind != EvidenceKind {
		t.Fatalf("unexpected evidence: %#v %v", ev, err)
	}
	if _, err := EvidenceFromSnapshot(s, "other-node", KubeletConfigPath, now); err == nil {
		t.Fatal("expected identity mismatch")
	}
	if _, err := EvidenceFromSnapshot(s, "kind-control-plane", "etc/shadow", now); err == nil {
		t.Fatal("expected path rejection")
	}
	s.GeneratedAt = now.Add(-3 * time.Minute)
	if _, err := EvidenceFromSnapshot(s, "kind-control-plane", KubeletConfigPath, now); err == nil {
		t.Fatal("expected stale snapshot rejection")
	}
}

func TestModeEvaluatorPassFailAndNotAssessed(t *testing.T) {
	tests := []struct {
		name             string
		mode             uint32
		regular, symlink bool
		want             adapter.Status
	}{
		{"pass exact", 0o600, true, false, adapter.StatusPass},
		{"pass restrictive", 0o400, true, false, adapter.StatusPass},
		{"fail broad", 0o640, true, false, adapter.StatusFail},
		{"symlink", 0o600, false, true, adapter.StatusNotAssessed},
		{"non regular", 0o600, false, false, adapter.StatusNotAssessed},
	}
	eval := ModeEvaluator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := adapter.EvaluatorRequest{Target: "node", Evidence: adapter.Evidence{Kind: EvidenceKind, Key: KubeletConfigPath, Value: map[string]any{"mode": float64(tt.mode), "regular": tt.regular, "symlink": tt.symlink}}, Compare: adapter.Comparator{Operator: "bitmask", Value: "0600"}}
			got := eval.Evaluate(context.Background(), req)
			if got.Status != tt.want {
				t.Fatalf("got %s want %s: %s", got.Status, tt.want, got.Reason)
			}
		})
	}
}

func TestAdapterEndToEnd(t *testing.T) {
	a, err := NewAdapter()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	s := Snapshot{NodeName: "node-a", GeneratedAt: now, Files: map[string]FileMetadata{KubeletConfigPath: {Mode: 0o640, Regular: true}}}
	ev, err := EvidenceFromSnapshot(s, "node-a", KubeletConfigPath, now)
	if err != nil {
		t.Fatal(err)
	}
	got := a.Evaluate(context.Background(), Descriptor("0600"), ev)
	if got.Status != adapter.StatusFail {
		t.Fatalf("got %s, want FAIL", got.Status)
	}
	missing := adapter.Evidence{Kind: EvidenceKind, Key: KubeletConfigPath, Found: false}
	got = a.Evaluate(context.Background(), Descriptor("0600"), missing)
	if got.Status != adapter.StatusNotAssessed {
		t.Fatalf("missing evidence got %s", got.Status)
	}
}

func TestAdapterRejectsUnapprovedControlAndComparator(t *testing.T) {
	a, err := NewAdapter()
	if err != nil {
		t.Fatal(err)
	}
	unknown := Descriptor("0600")
	unknown.ControlID = "4.1.10"
	if got := a.Prepare(unknown); got.Request != nil || got.Status != adapter.StatusNotAssessed {
		t.Fatal("unknown control was approved")
	}
	bad := Descriptor("0600")
	bad.Compare.Operator = "exec"
	if got := a.Prepare(bad); got.Request != nil {
		t.Fatal("unsupported comparator was approved")
	}
}

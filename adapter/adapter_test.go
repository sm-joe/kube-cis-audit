package adapter

import (
	"context"
	"testing"
)

type fakeEvaluator struct{ status Status }

func (f fakeEvaluator) Evaluate(context.Context, EvaluatorRequest) Evaluation {
	return Evaluation{Status: f.status, Reason: "fixture result"}
}
func fixture(t *testing.T) (*Adapter, TestDescriptor) {
	t.Helper()
	r := NewRegistry()
	if err := r.Register("node-file-mode", fakeEvaluator{StatusPass}); err != nil {
		t.Fatal(err)
	}
	m := Mapping{ControlID: "1.1.1", Target: "master", SourceKind: "path", SourceValue: "/etc/kubernetes/manifests/kube-apiserver.yaml", ComparatorOps: []string{"eq"}, EvaluatorID: "node-file-mode", EvidenceKind: "file_metadata", AllowEvaluation: true}
	a, err := New([]Mapping{m}, r)
	if err != nil {
		t.Fatal(err)
	}
	d := TestDescriptor{ControlID: "1.1.1", Target: "master", Source: Source{Kind: "path", Value: "/etc/kubernetes/manifests/kube-apiserver.yaml"}, Compare: Comparator{Operator: "eq", Value: 600}}
	return a, d
}
func TestPrepareExactAllowlistedMapping(t *testing.T) {
	a, d := fixture(t)
	got := a.Prepare(d)
	if got.Request == nil || got.EvaluatorID != "node-file-mode" {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestUnknownControlRemainsNotAssessed(t *testing.T) {
	a, d := fixture(t)
	d.ControlID = "9.9.9"
	got := a.Prepare(d)
	if got.Status != StatusNotAssessed || got.Request != nil {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestManualRemainsManual(t *testing.T) {
	a, d := fixture(t)
	d.Manual = true
	got := a.Prepare(d)
	if got.Status != StatusManual {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestComparatorMustBeAllowlisted(t *testing.T) {
	a, d := fixture(t)
	d.Compare.Operator = "gt"
	got := a.Prepare(d)
	if got.Request != nil || got.Status != StatusNotAssessed {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestMissingEvidenceIsNotPass(t *testing.T) {
	a, d := fixture(t)
	got := a.Evaluate(context.Background(), d, Evidence{Kind: "file_metadata", Key: d.Source.Value, Found: false})
	if got.Status != StatusNotAssessed {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestEvidenceContractMustMatch(t *testing.T) {
	a, d := fixture(t)
	got := a.Evaluate(context.Background(), d, Evidence{Kind: "raw_file", Key: d.Source.Value, Found: true})
	if got.Status != StatusNotAssessed {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestTrustedEvidenceInvokesEvaluator(t *testing.T) {
	a, d := fixture(t)
	got := a.Evaluate(context.Background(), d, Evidence{Kind: "file_metadata", Key: d.Source.Value, Found: true, Value: 600})
	if got.Status != StatusPass {
		t.Fatalf("unexpected: %#v", got)
	}
}
func TestUnregisteredEvaluatorRejected(t *testing.T) {
	r := NewRegistry()
	_, err := New([]Mapping{{ControlID: "1", Target: "node", SourceKind: "path", SourceValue: "/x", EvaluatorID: "missing", EvidenceKind: "file_metadata", AllowEvaluation: true}}, r)
	if err == nil {
		t.Fatal("expected unregistered evaluator error")
	}
}

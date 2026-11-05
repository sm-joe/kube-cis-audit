package conformance

import "testing"

func TestCandidateCannotBeReleaseReady(t *testing.T) {
	r := Validate(Inventory{Profiles: []Profile{{Name: "p", BenchmarkVersion: "1.0", Controls: []Mapping{{ControlID: "1.1", Title: "x", Status: "candidate"}}}}})
	if r.ReleaseReady || r.Candidate != 1 {
		t.Fatal(r)
	}
}
func TestVerifiedNeedsProvenance(t *testing.T) {
	r := Validate(Inventory{Profiles: []Profile{{Name: "p", BenchmarkVersion: "1.0", Controls: []Mapping{{ControlID: "1.1", Title: "x", Status: "verified"}}}}})
	if r.ReleaseReady || len(r.Findings) == 0 {
		t.Fatal(r)
	}
}
func TestVerifiedWithProvenance(t *testing.T) {
	r := Validate(Inventory{Profiles: []Profile{{Name: "p", BenchmarkVersion: "1.0", Controls: []Mapping{{ControlID: "1.1", Title: "x", Status: "verified", EvidenceKind: "flag", SourceFile: "cfg.yaml"}}}}})
	if !r.ReleaseReady {
		t.Fatal(r)
	}
}

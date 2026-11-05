package etcd

import (
	"testing"
	"time"
)

func baseEvidence() Evidence {
	return Evidence{
		Component: ComponentEtcd, Profile: "cis-1.24", CollectedAt: time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC),
		Complete: true, SourceComplete: map[string]bool{"flags": true, "config": true},
		Flags:        map[string][]string{"--client-cert-auth": {"true"}, "--auto-tls": {"false"}, "--data-dir": {"/var/lib/etcd"}},
		ConfigValues: map[string]string{"client-transport-security.client-cert-auth": "true"},
	}
}

func TestClientCertAuthPass(t *testing.T) {
	e := baseEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if len(got) != 1 || got[0].Status != StatusPass {
		t.Fatalf("expected PASS: %#v", got)
	}
}

func TestWrongValueFail(t *testing.T) {
	e := baseEvidence()
	e.Flags["--client-cert-auth"] = []string{"false"}
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusFail {
		t.Fatalf("expected FAIL: %#v", got[0])
	}
}

func TestMissingEvidenceWithoutSourceCompletenessNotAssessed(t *testing.T) {
	e := baseEvidence()
	delete(e.Flags, "--client-cert-auth")
	e.SourceComplete["flags"] = false
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED: %#v", got[0])
	}
}

func TestExplicitlyObservedMissingRequiredFlagFails(t *testing.T) {
	e := baseEvidence()
	delete(e.Flags, "--client-cert-auth")
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusFail {
		t.Fatalf("expected FAIL: %#v", got[0])
	}
}

func TestStaleEvidenceNotAssessed(t *testing.T) {
	e := baseEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(16*time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED: %#v", got[0])
	}
}

func TestConfigValuePass(t *testing.T) {
	e := baseEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[5:], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusPass {
		t.Fatalf("expected PASS: %#v", got[0])
	}
}

func TestSensitiveMappingRejected(t *testing.T) {
	m := DefaultMappings()[:1]
	m[0].Sensitive = true
	if err := ValidateMappings(m); err == nil {
		t.Fatal("expected sensitive mapping rejection")
	}
}

func TestSourceErrorNotAssessed(t *testing.T) {
	e := baseEvidence()
	e.SourceErrors = []string{"collector timed out"}
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED: %#v", got[0])
	}
}

package controlplane

import (
	"testing"
	"time"
)

func freshEvidence() Evidence {
	return Evidence{
		Component: ComponentAPIServer, Profile: "cis-1.24", CollectedAt: time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC),
		Complete: true, Flags: map[string][]string{"--anonymous-auth": {"false"}},
	}
}

func TestFlagExpectedValuePass(t *testing.T) {
	e := freshEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if len(got) != 1 || got[0].Status != StatusPass {
		t.Fatalf("expected PASS, got %#v", got)
	}
}

func TestFlagMismatchFails(t *testing.T) {
	e := freshEvidence()
	e.Flags["--anonymous-auth"] = []string{"true"}
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusFail {
		t.Fatalf("expected FAIL, got %#v", got[0])
	}
}

func TestMissingFlagEvidenceNotAssessed(t *testing.T) {
	e := freshEvidence()
	delete(e.Flags, "--anonymous-auth")
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED, got %#v", got[0])
	}
}

func TestIncompleteEvidenceNotAssessed(t *testing.T) {
	e := freshEvidence()
	e.Complete = false
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED, got %#v", got[0])
	}
}

func TestStaleEvidenceNotAssessed(t *testing.T) {
	e := freshEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(16*time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED, got %#v", got[0])
	}
}

func TestAbsentFlagPassesOnlyWhenExplicitlyCollected(t *testing.T) {
	e := freshEvidence()
	e.Flags["--token-auth-file"] = []string{}
	got := EvaluateWithMappings(e, DefaultMappings()[1:2], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusPass {
		t.Fatalf("expected PASS for explicitly observed absence, got %#v", got[0])
	}
	delete(e.Flags, "--token-auth-file")
	got = EvaluateWithMappings(e, DefaultMappings()[1:2], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("missing evidence must not prove absence: %#v", got[0])
	}
}

func TestWrongComponentNotAssessed(t *testing.T) {
	e := freshEvidence()
	e.Component = ComponentScheduler
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if len(got) != 0 {
		t.Fatalf("wrong component must not select mapping: %#v", got)
	}
}

func TestConfigValue(t *testing.T) {
	e := freshEvidence()
	e.ConfigValues = map[string]string{"authentication.anonymous": "false"}
	m := Mapping{ControlID: "T-1", Title: "test config", Component: ComponentAPIServer, Source: SourceConfig, Key: "authentication.anonymous", ExpectedValue: "false"}
	got := EvaluateWithMappings(e, []Mapping{m}, e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusPass {
		t.Fatalf("expected PASS, got %#v", got[0])
	}
}

func TestInvalidDuplicateMappingsRejected(t *testing.T) {
	m := DefaultMappings()[:1]
	if err := ValidateMappings(append(m, m[0])); err == nil {
		t.Fatal("expected duplicate mapping error")
	}
}

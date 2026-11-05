package worker

import (
	"testing"
	"time"
)

func baseEvidence() Evidence {
	return Evidence{
		NodeName: "worker-01", Profile: "cis-1.24",
		CollectedAt: time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC), Complete: true,
		Components: map[Component]ComponentEvidence{
			ComponentKubelet: {
				Flags: map[string][]string{
					"--anonymous-auth": {"false"}, "--authorization-mode": {"Webhook"},
					"--client-ca-file": {"/etc/kubernetes/pki/ca.crt"}, "--read-only-port": {"0"},
					"--protect-kernel-defaults": {"true"}, "--make-iptables-util-chains": {"true"},
				},
				ConfigValues: map[string]string{
					"authentication.anonymous.enabled": "false", "authorization.mode": "Webhook", "readOnlyPort": "0",
				},
				SourceComplete: map[SourceKind]bool{SourceFlag: true, SourceConfig: true},
			},
			ComponentKubeProxy: {
				Flags:          map[string][]string{"--metrics-bind-address": {"127.0.0.1:10249"}},
				SourceComplete: map[SourceKind]bool{SourceFlag: true},
			},
		},
	}
}

func TestKubeletFlagPass(t *testing.T) {
	e := baseEvidence()
	m := DefaultMappings()[:1]
	got := EvaluateWithMappings(e, m, e.CollectedAt.Add(time.Minute))
	if len(got) != 1 || got[0].Status != StatusPass {
		t.Fatalf("expected PASS, got %#v", got)
	}
}

func TestKubeletFlagMismatchFail(t *testing.T) {
	e := baseEvidence()
	ce := e.Components[ComponentKubelet]
	ce.Flags["--anonymous-auth"] = []string{"true"}
	e.Components[ComponentKubelet] = ce
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusFail {
		t.Fatalf("expected FAIL, got %#v", got[0])
	}
}

func TestMissingSourceIncompleteNotAssessed(t *testing.T) {
	e := baseEvidence()
	ce := e.Components[ComponentKubelet]
	delete(ce.Flags, "--anonymous-auth")
	ce.SourceComplete[SourceFlag] = false
	e.Components[ComponentKubelet] = ce
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED, got %#v", got[0])
	}
}

func TestMissingFlagWithCompleteSourceFails(t *testing.T) {
	e := baseEvidence()
	ce := e.Components[ComponentKubelet]
	delete(ce.Flags, "--anonymous-auth")
	e.Components[ComponentKubelet] = ce
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusFail {
		t.Fatalf("expected FAIL, got %#v", got[0])
	}
}

func TestConfigValuePass(t *testing.T) {
	e := baseEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[6:7], e.CollectedAt.Add(time.Minute))
	if got[0].Status != StatusPass {
		t.Fatalf("expected PASS, got %#v", got[0])
	}
}

func TestKubeProxyComponentMapping(t *testing.T) {
	e := baseEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[9:], e.CollectedAt.Add(time.Minute))
	if len(got) != 1 || got[0].Component != ComponentKubeProxy || got[0].Status != StatusPass {
		t.Fatalf("expected kube-proxy PASS, got %#v", got)
	}
}

func TestStaleEvidenceNotAssessed(t *testing.T) {
	e := baseEvidence()
	got := EvaluateWithMappings(e, DefaultMappings()[:1], e.CollectedAt.Add(16*time.Minute))
	if len(got) != 1 || got[0].Status != StatusNotAssessed {
		t.Fatalf("expected NOT_ASSESSED, got %#v", got)
	}
}

func TestInvalidDuplicateMappingRejected(t *testing.T) {
	m := DefaultMappings()[:1]
	if err := ValidateMappings(append(m, m[0])); err == nil {
		t.Fatal("expected duplicate mapping error")
	}
}

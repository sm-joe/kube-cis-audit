package audit

import (
	"testing"
	"time"
)

func ev() Evidence {
	return Evidence{ClusterID: "c", Profile: "p", Source: "test", Complete: true, CollectedAt: time.Now(), MaxAge: time.Hour}
}
func ptr(v bool) *bool { return &v }
func TestAuditMissingConfigNotAssessed(t *testing.T) {
	r := EvaluateConfig(AuditConfig{Evidence: ev()})
	if r[0].Status != NotAssessed {
		t.Fatal(r[0])
	}
}
func TestAuditRotationPass(t *testing.T) {
	a, b, c := 10, 5, 100
	r := EvaluateConfig(AuditConfig{Evidence: ev(), LogMaxAge: &a, LogMaxBackup: &b, LogMaxSize: &c})
	if r[len(r)-1].Status != Pass {
		t.Fatal(r)
	}
}
func TestPolicyNeedsCatchAll(t *testing.T) {
	r := EvaluatePolicy(PolicyInventory{Evidence: ev(), Rules: []PolicyRule{{Level: "Metadata"}}, DefaultLevel: "None"})
	if r[0].Status != Warn {
		t.Fatal(r[0])
	}
}

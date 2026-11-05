package admission

import (
	"testing"
	"time"
)

func ev() Evidence {
	return Evidence{ClusterID: "c", Profile: "p", Source: "test", Complete: true, CollectedAt: time.Now(), MaxAge: time.Hour}
}
func TestPluginMissingIsNotAssessed(t *testing.T) {
	r := EvaluatePlugins(PluginConfig{Evidence: ev(), ConfigKnown: true})
	if r[0].Status != NotAssessed {
		t.Fatal(r[0])
	}
}
func TestDisabledPluginFails(t *testing.T) {
	r := EvaluatePlugins(PluginConfig{Evidence: ev(), ConfigKnown: true, DisabledPlugins: []string{"NodeRestriction"}})
	if r[3].Status != Fail {
		t.Fatal(r[3])
	}
}
func TestPolicyNotReadyWarns(t *testing.T) {
	r := EvaluatePolicies(PolicyInventory{Evidence: ev(), Policies: []Policy{{Name: "p", Ready: false, HasValidation: true, FailurePolicy: "Fail"}}})
	if r[0].Status != Warn {
		t.Fatal(r[0])
	}
}

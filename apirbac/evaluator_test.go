package apirbac

import (
	"testing"
	"time"
)

func good() Evidence {
	return Evidence{ClusterID: "c1", Profile: "candidate", CollectedAt: time.Now(), MaxAge: time.Hour, Complete: true, Source: "test-inventory"}
}
func TestAPIRequiresCompleteEvidence(t *testing.T) {
	e := good()
	e.Complete = false
	r := EvaluateAPI(APIConfig{Evidence: e})
	for _, x := range r {
		if x.Status != NotAssessed {
			t.Fatalf("got %s", x.Status)
		}
	}
}
func TestRBACWildcardWarn(t *testing.T) {
	r := EvaluateRBAC(RBACInventory{Evidence: good(), Roles: []Role{{Name: "wide", Rules: []Rule{{Verbs: []string{"*"}, Resources: []string{"pods"}}}}}})
	if r[0].Status != Warn {
		t.Fatalf("got %s", r[0].Status)
	}
}
func TestRBACIncompleteNotAssessed(t *testing.T) {
	e := good()
	e.Errors = []string{"forbidden"}
	r := EvaluateRBAC(RBACInventory{Evidence: e})
	if r[0].Status != NotAssessed {
		t.Fatalf("got %s", r[0].Status)
	}
}

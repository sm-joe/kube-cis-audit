package admission

import (
	"fmt"
	"strings"
	"time"
)

func gate(e Evidence) string {
	if e.Source == "" || e.ClusterID == "" || e.Profile == "" {
		return "source, cluster identity, or profile missing"
	}
	if !e.Complete {
		return "evidence incomplete"
	}
	if len(e.Errors) > 0 {
		return "collection errors present"
	}
	if e.CollectedAt.IsZero() || e.MaxAge <= 0 || e.CollectedAt.After(time.Now().Add(2*time.Minute)) || time.Since(e.CollectedAt) > e.MaxAge {
		return "evidence timestamp invalid or stale"
	}
	return ""
}
func has(xs []string, s string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}
func EvaluatePlugins(c PluginConfig) []Result {
	names := []string{"NamespaceLifecycle", "LimitRanger", "ServiceAccount", "NodeRestriction"}
	out := []Result{}
	g := gate(c.Evidence)
	for _, n := range names {
		st := NotAssessed
		why := "plugin configuration was not supplied"
		if g != "" {
			why = g
		} else if !c.ConfigKnown {
			why = "plugin configuration is not declared complete"
		} else if has(c.DisabledPlugins, n) {
			st = Fail
			why = "required admission plugin is explicitly disabled"
		} else if has(c.EnabledPlugins, n) {
			st = Pass
			why = "required admission plugin is explicitly enabled"
		} else {
			st = NotAssessed
			why = "plugin is not present in enabled or disabled evidence"
		}
		out = append(out, Result{"B8-PLUGIN-" + strings.ToUpper(n), "Admission plugin " + n, st, why, c.Evidence.Source})
	}
	return out
}
func EvaluatePolicies(inv PolicyInventory) []Result {
	g := gate(inv.Evidence)
	if g != "" {
		return []Result{{"B8-POLICY-INVENTORY", "Admission policy inventory", NotAssessed, g, inv.Evidence.Source}}
	}
	if len(inv.Policies) == 0 {
		return []Result{{"B8-POLICY-INVENTORY", "Admission policy inventory", NotAssessed, "complete inventory contains no policy objects; absence is not interpreted as compliance", inv.Evidence.Source}}
	}
	bad := []string{}
	for _, p := range inv.Policies {
		if !p.Ready {
			bad = append(bad, p.Name+": not ready")
		}
		if p.FailurePolicy != "" && !strings.EqualFold(p.FailurePolicy, "Fail") {
			bad = append(bad, p.Name+": failurePolicy is not Fail")
		}
		if !p.HasValidation && !p.HasMutation {
			bad = append(bad, p.Name+": no validation or mutation behavior declared")
		}
	}
	if len(bad) > 0 {
		return []Result{{"B8-POLICY-READINESS", "Admission policy readiness indicators", Warn, fmt.Sprint(bad), inv.Evidence.Source}}
	}
	return []Result{{"B8-POLICY-READINESS", "Admission policy readiness indicators", Pass, "all supplied policies report ready and declare validation or mutation behavior", inv.Evidence.Source}}
}

package audit

import (
	"fmt"
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
func EvaluateConfig(c AuditConfig) []Result {
	fields := []struct {
		id, title string
		value     *bool
		expected  bool
	}{{"B9-AUDIT-POLICY", "Audit policy is configured", c.PolicyFileConfigured, true}, {"B9-AUDIT-LOGPATH", "Audit log path is configured", c.LogPathConfigured, true}, {"B9-AUDIT-WEBHOOK", "Audit webhook backend is configured", c.WebhookBackendConfigured, false}}
	out := []Result{}
	g := gate(c.Evidence)
	for _, f := range fields {
		st := NotAssessed
		why := "configuration value missing"
		if g != "" {
			why = g
		} else if f.value != nil {
			if *f.value == f.expected {
				st = Pass
				why = "supplied configuration matches candidate predicate"
			} else {
				st = Fail
				why = "supplied configuration does not match candidate predicate"
			}
		}
		out = append(out, Result{f.id, f.title, st, why, c.Evidence.Source})
	}
	if g != "" {
		return out
	}
	if c.LogMaxAge == nil || c.LogMaxBackup == nil || c.LogMaxSize == nil {
		out = append(out, Result{"B9-AUDIT-ROTATION", "Audit log rotation limits configured", NotAssessed, "one or more rotation values are missing", c.Evidence.Source})
	} else if *c.LogMaxAge > 0 && *c.LogMaxBackup > 0 && *c.LogMaxSize > 0 {
		out = append(out, Result{"B9-AUDIT-ROTATION", "Audit log rotation limits configured", Pass, "all supplied rotation limits are positive", c.Evidence.Source})
	} else {
		out = append(out, Result{"B9-AUDIT-ROTATION", "Audit log rotation limits configured", Fail, "rotation limits must be positive", c.Evidence.Source})
	}
	return out
}
func EvaluatePolicy(p PolicyInventory) []Result {
	g := gate(p.Evidence)
	if g != "" {
		return []Result{{"B9-AUDIT-RULES", "Audit policy rule coverage", NotAssessed, g, p.Evidence.Source}}
	}
	if len(p.Rules) == 0 {
		return []Result{{"B9-AUDIT-RULES", "Audit policy rule coverage", NotAssessed, "no policy rules supplied", p.Evidence.Source}}
	}
	if !p.HasCatchAll || p.DefaultLevel == "" {
		return []Result{{"B9-AUDIT-RULES", "Audit policy rule coverage", Warn, "catch-all rule or default audit level not evidenced", p.Evidence.Source}}
	}
	for _, r := range p.Rules {
		if r.Level == "" {
			return []Result{{"B9-AUDIT-RULES", "Audit policy rule coverage", Warn, fmt.Sprintf("rule with verbs %v has no level", r.Verbs), p.Evidence.Source}}
		}
	}
	return []Result{{"B9-AUDIT-RULES", "Audit policy rule coverage", Pass, "supplied rules have levels and a catch-all/default level is evidenced", p.Evidence.Source}}
}

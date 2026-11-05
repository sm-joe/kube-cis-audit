package apirbac

import (
	"fmt"
	"strings"
	"time"
)

func evidenceOK(e Evidence) string {
	if e.Source == "" {
		return "evidence source is not identified"
	}
	if e.ClusterID == "" {
		return "cluster identity is missing"
	}
	if e.Profile == "" {
		return "benchmark profile is missing"
	}
	if !e.Complete {
		return "evidence source is not marked complete"
	}
	if len(e.Errors) > 0 {
		return "evidence collection reported errors"
	}
	if e.CollectedAt.IsZero() {
		return "collection timestamp is missing"
	}
	if e.CollectedAt.After(time.Now().Add(2 * time.Minute)) {
		return "collection timestamp is in the future"
	}
	if e.MaxAge <= 0 || time.Since(e.CollectedAt) > e.MaxAge {
		return "evidence is stale or has no positive freshness limit"
	}
	return ""
}
func result(id, title string, s Status, reason, source string) Result {
	return Result{id, title, s, reason, source}
}

func EvaluateAPI(c APIConfig) []Result {
	checks := []struct {
		id, title string
		ok        bool
		known     bool
		why       string
	}{
		{"B7-API-ANON", "API server anonymous authentication disabled", c.AnonymousAuth != nil && !*c.AnonymousAuth, c.AnonymousAuth != nil, "anonymous-auth must be explicitly false"},
		{"B7-API-INSECURE-BIND", "API server insecure bind address is not exposed", c.InsecureBindAddress == "" || c.InsecureBindAddress == "127.0.0.1", c.InsecureBindAddress != "", "insecure bind address must be loopback or absent"},
		{"B7-API-INSECURE-PORT", "API server insecure port is disabled", c.InsecurePort != nil && *c.InsecurePort == 0, c.InsecurePort != nil, "insecure port must be explicitly zero"},
		{"B7-API-AUTHZ", "API server uses an approved authorization mode", contains(c.AuthorizationModes, "Node") && contains(c.AuthorizationModes, "RBAC"), len(c.AuthorizationModes) > 0, "authorization modes must include Node and RBAC"},
		{"B7-API-AUDIT", "API server audit logging is enabled", c.AuditEnabled != nil && *c.AuditEnabled, c.AuditEnabled != nil, "audit logging must be explicitly enabled"},
	}
	out := make([]Result, 0, len(checks))
	gate := evidenceOK(c.Evidence)
	for _, x := range checks {
		switch {
		case gate != "":
			out = append(out, result(x.id, x.title, NotAssessed, gate, c.Evidence.Source))
		case !x.known:
			out = append(out, result(x.id, x.title, NotAssessed, "required configuration value is absent", c.Evidence.Source))
		case x.ok:
			out = append(out, result(x.id, x.title, Pass, "supplied configuration satisfies candidate predicate", c.Evidence.Source))
		default:
			out = append(out, result(x.id, x.title, Fail, x.why, c.Evidence.Source))
		}
	}
	return out
}
func contains(xs []string, s string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}

// EvaluateRBAC reports broad wildcard permissions and privileged bindings.
// It requires complete Role/RoleBinding/ClusterRole/ClusterRoleBinding inventory.
// It does not infer effective permissions across aggregation rules or external authorizers.
func EvaluateRBAC(inv RBACInventory) []Result {
	gate := evidenceOK(inv.Evidence)
	if gate != "" {
		return []Result{result("B7-RBAC-INVENTORY", "RBAC inventory assessment", NotAssessed, gate, inv.Evidence.Source)}
	}
	var wildcard, privileged []string
	for _, r := range inv.Roles {
		for _, rule := range r.Rules {
			if contains(rule.Verbs, "*") || contains(rule.Resources, "*") || contains(rule.APIGroups, "*") || contains(rule.NonResourceURLs, "*") {
				wildcard = append(wildcard, fmt.Sprintf("%s/%s", r.Namespace, r.Name))
			}
			if contains(rule.Verbs, "*") && (contains(rule.Resources, "secrets") || contains(rule.Resources, "*")) {
				privileged = append(privileged, fmt.Sprintf("%s/%s", r.Namespace, r.Name))
			}
		}
	}
	for _, b := range inv.Bindings {
		if strings.EqualFold(b.RoleKind, "ClusterRole") && strings.EqualFold(b.RoleName, "cluster-admin") {
			for _, s := range b.Subjects {
				if strings.EqualFold(s.Kind, "ServiceAccount") {
					privileged = append(privileged, fmt.Sprintf("binding %s grants cluster-admin to service account %s/%s", b.Name, s.Namespace, s.Name))
				}
			}
		}
	}
	status := Pass
	reason := "complete inventory contains no detected wildcard rules or service-account cluster-admin bindings"
	if len(wildcard) > 0 || len(privileged) > 0 {
		status = Warn
		reason = fmt.Sprintf("broad permission indicators detected; wildcard roles=%v; privileged grants=%v", wildcard, privileged)
	}
	return []Result{result("B7-RBAC-BROAD", "RBAC broad-permission indicators", status, reason, inv.Evidence.Source)}
}

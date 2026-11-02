package audit

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type rbacRoleRef struct {
	kind string
	name string
}

// evaluateRBACRisk inventories role rules and their bindings. It reports
// permission risks, not proof that a subject has successfully exercised them.
func evaluateRBACRisk(ctx context.Context, client kubernetes.Interface, _ Check, o Options, r Result) Result {
	clusterRoles, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list ClusterRoles: "+err.Error()
		return r
	}
	roles, err := client.RbacV1().Roles(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list Roles: "+err.Error()
		return r
	}
	clusterBindings, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list ClusterRoleBindings: "+err.Error()
		return r
	}
	bindings, err := client.RbacV1().RoleBindings(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list RoleBindings: "+err.Error()
		return r
	}

	clusterByName := make(map[string]rbacv1.ClusterRole, len(clusterRoles.Items))
	for _, role := range clusterRoles.Items {
		clusterByName[role.Name] = role
	}
	roleByKey := make(map[string]rbacv1.Role, len(roles.Items))
	for _, role := range roles.Items {
		roleByKey[role.Namespace+"/"+role.Name] = role
	}

	roleRules := func(ref rbacv1.RoleRef, namespace string) ([]rbacv1.PolicyRule, bool) {
		switch ref.Kind {
		case "ClusterRole":
			role, ok := clusterByName[ref.Name]
			return role.Rules, ok
		case "Role":
			role, ok := roleByKey[namespace+"/"+ref.Name]
			return role.Rules, ok
		default:
			return nil, false
		}
	}

	bindingCount, subjectCount, unresolved := 0, 0, 0
	critical, high, tokenAutomountEnabled := 0, 0, 0
	identityInventoryIncomplete := false
	appendEvidence := func(e Evidence) {
		if len(r.Evidence) < 200 {
			r.Evidence = append(r.Evidence, e)
		}
	}
	inspectBinding := func(kind, namespace, name string, ref rbacv1.RoleRef, subjects []rbacv1.Subject) {
		bindingCount++
		rules, ok := roleRules(ref, namespace)
		if !ok {
			unresolved++
			appendEvidence(Evidence{Resource: kind, Namespace: namespace, Name: name, Field: "roleRef", Observed: ref.Kind + ":" + ref.Name + " (unresolved)", Expected: "role reference resolves to an inventoried Role or ClusterRole"})
			return
		}
		for _, subject := range subjects {
			subjectCount++
			appendEvidence(Evidence{Resource: kind, Namespace: namespace, Name: name, Field: "subject", Observed: subjectIdentity(subject), Expected: "review subject ownership and least-privilege requirement"})
		}
		for i, rule := range rules {
			risks := classifyRBACRule(rule)
			for _, risk := range risks {
				if risk.level == 3 {
					critical++
				} else {
					high++
				}
				appendEvidence(Evidence{Resource: ref.Kind, Namespace: namespace, Name: ref.Name, Field: fmt.Sprintf("rules[%d]", i), Observed: risk.description + "; verbs=" + strings.Join(rule.Verbs, ",") + "; apiGroups=" + strings.Join(rule.APIGroups, ",") + "; resources=" + strings.Join(rule.Resources, ",") + "; resourceNames=" + strings.Join(rule.ResourceNames, ","), Expected: "remove unnecessary wildcard, privilege-escalation, impersonation, Secret-read or workload-exec permissions"})
			}
		}
	}
	for _, b := range clusterBindings.Items {
		inspectBinding("ClusterRoleBinding", "", b.Name, b.RoleRef, b.Subjects)
	}
	for _, b := range bindings.Items {
		inspectBinding("RoleBinding", b.Namespace, b.Name, b.RoleRef, b.Subjects)
	}

	// Record workload service-account token automount behavior without reading
	// token Secret contents. Pod-level configuration overrides ServiceAccount.
	serviceAccounts, saErr := client.CoreV1().ServiceAccounts(o.Namespace).List(ctx, metav1.ListOptions{})
	if saErr != nil {
		identityInventoryIncomplete = true
		appendEvidence(Evidence{Resource: "ServiceAccount", Field: "inventory", Observed: "list unavailable: " + saErr.Error(), Expected: "service-account automount configuration inventory"})
	} else {
		saByName := make(map[string]corev1.ServiceAccount, len(serviceAccounts.Items))
		for _, sa := range serviceAccounts.Items {
			saByName[sa.Name] = sa
		}
		pods, podErr := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
		if podErr != nil {
			identityInventoryIncomplete = true
			appendEvidence(Evidence{Resource: "Pod", Field: "inventory", Observed: "list unavailable: " + podErr.Error(), Expected: "workload service-account usage inventory"})
		} else {
			for _, pod := range pods.Items {
				if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
					continue
				}
				name := pod.Spec.ServiceAccountName
				if name == "" {
					name = "default"
				}
				var effective *bool
				source := "Kubernetes default (automount enabled unless overridden)"
				if pod.Spec.AutomountServiceAccountToken != nil {
					effective = pod.Spec.AutomountServiceAccountToken
					source = "Pod spec"
				} else if sa, ok := saByName[name]; ok && sa.AutomountServiceAccountToken != nil {
					effective = sa.AutomountServiceAccountToken
					source = "ServiceAccount spec"
				}
				observed := "true (default inferred); source=" + source
				if effective != nil {
					observed = fmt.Sprintf("%t; source=%s", *effective, source)
				}
				if effective == nil || *effective {
					tokenAutomountEnabled++
				}
				if len(r.Evidence) < 200 {
					r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: "spec.automountServiceAccountToken", Observed: observed + "; serviceAccount=" + name, Expected: "disable token automount when the workload does not require Kubernetes API credentials"})
				}
			}
		}
		for _, sa := range serviceAccounts.Items {
			if len(sa.Secrets) > 0 {
				names := make([]string, 0, len(sa.Secrets))
				for _, ref := range sa.Secrets {
					names = append(names, ref.Name)
				}
				appendEvidence(Evidence{Resource: "ServiceAccount", Namespace: sa.Namespace, Name: sa.Name, Field: "secrets", Observed: "referenced Secret metadata: " + strings.Join(stableStrings(names), ","), Expected: "review legacy service-account token Secret references; token contents are not retrieved"})
			}
		}
	}

	switch {
	case critical > 0:
		r.Status = "FAIL"
	case high > 0:
		r.Status = "WARN"
	case unresolved > 0 || identityInventoryIncomplete:
		r.Status = "NOT_ASSESSED"
	case tokenAutomountEnabled > 0:
		r.Status = "WARN"
	case bindingCount == 0:
		r.Status = "INFO"
	default:
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf("Inventoried %d ClusterRoles, %d Roles, %d ClusterRoleBindings and %d RoleBindings; inspected %d binding(s) and %d subject binding(s). Found %d critical privilege-escalation rule(s), %d high-risk permission rule(s), %d unresolved role reference(s), and %d workload(s) with service-account token automount enabled or defaulted on; identity inventory incomplete=%t. Permission inventory does not prove runtime use or effective authorization in every request context.", len(clusterRoles.Items), len(roles.Items), len(clusterBindings.Items), len(bindings.Items), bindingCount, subjectCount, critical, high, unresolved, tokenAutomountEnabled, identityInventoryIncomplete)
	return r
}

type rbacRisk struct {
	level       int
	description string
}

func classifyRBACRule(rule rbacv1.PolicyRule) []rbacRisk {
	var out []rbacRisk
	verbs := stringSet(rule.Verbs)
	groups := stringSet(rule.APIGroups)
	resources := stringSet(rule.Resources)
	if verbs["*"] || groups["*"] || resources["*"] {
		out = append(out, rbacRisk{3, "wildcard permission grants broad API access"})
	}
	for _, verb := range []string{"impersonate", "bind", "escalate"} {
		if verbs[verb] {
			out = append(out, rbacRisk{3, "privilege-escalation verb=" + verb})
		}
	}
	if groups["*"] || groups[""] {
		if resources["secrets"] && (verbs["*"] || verbs["get"] || verbs["list"] || verbs["watch"]) {
			out = append(out, rbacRisk{2, "read access to Kubernetes Secrets"})
		}
	}
	if groups["*"] || groups[""] || groups["apps"] {
		if resources["pods/exec"] || resources["pods/attach"] || resources["pods/portforward"] {
			out = append(out, rbacRisk{2, "interactive workload access through exec/attach/portforward"})
		}
	}
	if (verbs["create"] || verbs["update"] || verbs["patch"] || verbs["*"]) && (resources["pods"] || resources["deployments"] || resources["daemonsets"] || resources["statefulsets"] || resources["jobs"] || resources["cronjobs"]) {
		out = append(out, rbacRisk{2, "workload creation or mutation can enable workload-identity abuse; review service-account and Pod security constraints"})
	}
	if verbs["*"] && (resources["roles"] || resources["rolebindings"] || resources["clusterroles"] || resources["clusterrolebindings"]) {
		out = append(out, rbacRisk{3, "wildcard mutation may permit RBAC privilege escalation"})
	}
	return dedupeRBACRisks(out)
}

func stringSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
func dedupeRBACRisks(in []rbacRisk) []rbacRisk {
	seen := map[string]bool{}
	out := make([]rbacRisk, 0, len(in))
	for _, risk := range in {
		if !seen[risk.description] {
			seen[risk.description] = true
			out = append(out, risk)
		}
	}
	return out
}

// stableStrings is used by tests and future report renderers to keep output deterministic.
func stableStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

package audit

import (
	"context"
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EvaluateEKSManualEvidence supplies reproducible, read-only observations
// for human controls. No heuristic auto-approves a manual CIS recommendation.
func EvaluateEKSManualEvidence(ctx context.Context, client kubernetes.Interface, c Check) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "MANUAL", Method: "kubernetes_inventory_human_attestation", Scored: false, Remediation: c.Remediation}
	r.Details = c.Description + " Reviewer must record the approval, exception or rejection with a linked evidence reference."
	if strings.HasPrefix(c.ID, "4.1.") {
		roleCheck := func(kind, ns, name string, rules []rbacv1.PolicyRule) {
			for i, rule := range rules {
				interesting := false
				switch c.ID {
				case "4.1.2":
					interesting = hasRBACItem(rule.Resources, "secrets") && hasRBACAny(rule.Verbs, "get", "list", "watch", "*")
				case "4.1.3":
					interesting = hasRBACItem(rule.Verbs, "*") || hasRBACItem(rule.Resources, "*") || hasRBACItem(rule.APIGroups, "*")
				case "4.1.4":
					interesting = hasRBACItem(rule.Resources, "pods") && hasRBACAny(rule.Verbs, "create", "*")
				case "4.1.8":
					interesting = hasRBACAny(rule.Verbs, "bind", "impersonate", "escalate")
				case "4.1.9":
					interesting = hasRBACItem(rule.Resources, "persistentvolumes") && hasRBACAny(rule.Verbs, "create", "*")
				case "4.1.10":
					interesting = hasRBACItem(rule.Resources, "nodes/proxy")
				case "4.1.11":
					interesting = hasRBACAny(rule.Resources, "validatingwebhookconfigurations", "mutatingwebhookconfigurations") && hasRBACAny(rule.Verbs, "create", "update", "patch", "delete", "*")
				case "4.1.12":
					interesting = hasRBACItem(rule.Resources, "serviceaccounts/token") && hasRBACAny(rule.Verbs, "create", "*")
				}
				if interesting {
					r.Evidence = append(r.Evidence, Evidence{Resource: kind, Namespace: ns, Name: name, Field: fmt.Sprintf("rules[%d]", i), Observed: "sensitive RBAC permission candidate", Expected: "approved least-privilege use"})
				}
			}
		}
		if c.ID == "4.1.1" {
			bindings, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
			if err != nil {
				r.Status = "NOT_ASSESSED"
				r.Details = "Cannot list cluster role bindings: " + err.Error()
				return r
			}
			for _, binding := range bindings.Items {
				if binding.RoleRef.Name == "cluster-admin" && binding.RoleRef.Kind == "ClusterRole" {
					r.Evidence = append(r.Evidence, Evidence{Resource: "ClusterRoleBinding", Name: binding.Name, Field: "roleRef", Observed: fmt.Sprintf("cluster-admin; %d subjects", len(binding.Subjects)), Expected: "approved administrative identities"})
				}
			}
		} else {
			roles, err := client.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
			if err != nil {
				r.Status = "NOT_ASSESSED"
				r.Details = "Cannot list Roles: " + err.Error()
				return r
			}
			clusterRoles, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
			if err != nil {
				r.Status = "NOT_ASSESSED"
				r.Details = "Cannot list ClusterRoles: " + err.Error()
				return r
			}
			for _, role := range roles.Items {
				roleCheck("Role", role.Namespace, role.Name, role.Rules)
			}
			for _, role := range clusterRoles.Items {
				roleCheck("ClusterRole", "", role.Name, role.Rules)
			}
		}
		r.Details = fmt.Sprintf("%s Found %d candidate RBAC records. Review bindings, approved principals and legitimate operational exceptions.", c.Title, len(r.Evidence))
		return r
	}
	switch c.ID {
	case "4.3.2":
		policies, err := client.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = "Cannot list NetworkPolicies: " + err.Error()
			return r
		}
		for _, policy := range policies.Items {
			r.Evidence = append(r.Evidence, Evidence{Resource: "NetworkPolicy", Namespace: policy.Namespace, Name: policy.Name, Field: "policyTypes", Observed: fmt.Sprintf("%v", policy.Spec.PolicyTypes), Expected: "ingress and egress rules covering intended Pods"})
		}
		r.Details = fmt.Sprintf("Found %d namespaced NetworkPolicies; map active workloads to selectors and review additional ClusterNetworkPolicy engines before approval.", len(policies.Items))
	case "4.4.2":
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = "Cannot list Pods: " + err.Error()
			return r
		}
		for _, pod := range pods.Items {
			for _, volume := range pod.Spec.Volumes {
				if volume.CSI != nil && strings.Contains(strings.ToLower(volume.CSI.Driver), "secrets-store") {
					r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: "volume/"+volume.Name, Observed: "Secrets Store CSI", Expected: "approved external secret store with rotation"})
				}
			}
		}
		r.Details = fmt.Sprintf("%d active CSI secret-store volume references; review external storage access and rotation for sensitive workloads.", len(r.Evidence))
	case "4.5.1":
		namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = "Cannot list namespaces: " + err.Error()
			return r
		}
		for _, ns := range namespaces.Items {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "pod-security.kubernetes.io/enforce", Observed: firstNonEmpty(ns.Labels["pod-security.kubernetes.io/enforce"], "<unset>"), Expected: "approved administrative boundary"})
		}
		r.Details = fmt.Sprintf("%d namespaces inventoried; verify organization ownership, namespace-level RBAC and cross-namespace network access.", len(namespaces.Items))
	case "5.1.4":
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = "Cannot list Pod image references: " + err.Error()
			return r
		}
		seen := map[string]bool{}
		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				parts := strings.Split(container.Image, "/")
				registry := "docker.io (implicit)"
				if len(parts) > 1 && (strings.ContainsAny(parts[0], ".:") || parts[0] == "localhost") {
					registry = parts[0]
				}
				if !seen[registry] {
					seen[registry] = true
					r.Evidence = append(r.Evidence, Evidence{Resource: "ContainerRegistry", Name: registry, Field: "observed", Observed: "present", Expected: "approved registry with enforced admission restrictions"})
				}
			}
		}
		r.Details = fmt.Sprintf("%d container registries observed; review the allowlist and validation/admission enforcement.", len(seen))
	case "5.4.4":
		services, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = "Cannot list LoadBalancer Services: " + err.Error()
			return r
		}
		for _, svc := range services.Items {
			if string(svc.Spec.Type) == "LoadBalancer" {
				r.Evidence = append(r.Evidence, Evidence{Resource: "Service", Namespace: svc.Namespace, Name: svc.Name, Field: "service.beta.kubernetes.io/aws-load-balancer-ssl-cert", Observed: firstNonEmpty(svc.Annotations["service.beta.kubernetes.io/aws-load-balancer-ssl-cert"], "<unset or TLS passthrough>"), Expected: "valid certificate on TLS-terminating client-facing listeners"})
			}
		}
		ingresses, err := client.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = "Cannot list Ingress resources: " + err.Error()
			return r
		}
		for _, ing := range ingresses.Items {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Ingress", Namespace: ing.Namespace, Name: ing.Name, Field: "alb.ingress.kubernetes.io/certificate-arn", Observed: firstNonEmpty(ing.Annotations["alb.ingress.kubernetes.io/certificate-arn"], "<auto-discovery or unset>"), Expected: "valid listener TLS certificate"})
		}
		r.Details = fmt.Sprintf("%d client-facing Service/Ingress candidates; verify actual ALB/NLB listeners, TLS termination and certificate validity.", len(r.Evidence))
	}
	return r
}

func hasRBACItem(items []string, value string) bool {
	for _, x := range items {
		if x == value {
			return true
		}
	}
	return false
}

func hasRBACAny(items []string, values ...string) bool {
	for _, v := range values {
		if hasRBACItem(items, v) {
			return true
		}
	}
	return false
}

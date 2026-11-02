package audit

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type controlPlaneComponent struct {
	name string
	pod  corev1.Pod
	args []string
}

// evaluateControlPlaneSecurity inspects only visible control-plane Pod command
// arguments. Managed control planes that hide these Pods are explicitly marked
// MANUAL/NOT_ASSESSED rather than inferred to be secure.
func evaluateControlPlaneSecurity(ctx context.Context, client kubernetes.Interface, _ Check, _ Options, r Result) Result {
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to inspect kube-system Pods for control-plane evidence: "+err.Error()
		return r
	}
	components := map[string]controlPlaneComponent{}
	for _, pod := range pods.Items {
		name := strings.ToLower(pod.Name)
		component := ""
		for _, candidate := range []string{"kube-apiserver", "etcd", "kube-controller-manager", "kube-scheduler"} {
			if strings.Contains(name, candidate) {
				component = candidate
				break
			}
		}
		if component == "" {
			continue
		}
		args := []string{}
		for _, c := range pod.Spec.Containers {
			args = append(args, c.Command...)
			args = append(args, c.Args...)
		}
		components[component] = controlPlaneComponent{name: component, pod: pod, args: args}
	}
	if len(components) == 0 {
		r.Status = "MANUAL"
		r.Details = "No control-plane component Pods are visible in kube-system. This is common for managed Kubernetes services; verify control-plane settings through the provider's supported security configuration and evidence. No control-plane PASS is inferred."
		r.Evidence = append(r.Evidence, Evidence{Resource: "ControlPlane", Field: "componentPodInventory", Observed: "no visible API server/etcd/controller-manager/scheduler Pods", Expected: "provider or node-local control-plane configuration evidence"})
		return r
	}

	missing := []string{}
	for _, name := range []string{"kube-apiserver", "etcd", "kube-controller-manager", "kube-scheduler"} {
		if _, ok := components[name]; !ok {
			missing = append(missing, name)
		}
	}
	findings, unknown := 0, 0
	for _, name := range sortedComponentNames(components) {
		component := components[name]
		checks := controlPlaneFlagChecks(name)
		for _, check := range checks {
			value, found := flagValue(component.args, check.flag)
			found = found && strings.TrimSpace(value) != ""
			observed := "not explicitly configured in visible Pod arguments"
			if found {
				observed = value
			}
			if !found {
				unknown++
				r.Evidence = append(r.Evidence, Evidence{Resource: "ControlPlanePod", Namespace: component.pod.Namespace, Name: component.pod.Name, Field: check.flag, Observed: observed, Expected: check.expected + " (effective default/provider configuration requires confirmation)"})
				continue
			}
			bad := check.bad(value)
			if bad {
				findings++
			}
			r.Evidence = append(r.Evidence, Evidence{Resource: "ControlPlanePod", Namespace: component.pod.Namespace, Name: component.pod.Name, Field: check.flag, Observed: value, Expected: check.expected})
		}
	}
	if len(missing) > 0 {
		r.Evidence = append(r.Evidence, Evidence{Resource: "ControlPlane", Field: "missingComponents", Observed: strings.Join(missing, ","), Expected: "visible component Pod or independent provider evidence"})
	}
	switch {
	case findings > 0:
		r.Status = "FAIL"
	case len(missing) > 0 || unknown > 0:
		r.Status = "MANUAL"
	default:
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf("Inspected %d visible control-plane component Pod(s); %d component(s) were not visible, %d configured flag value(s) require review, and %d settings were not explicit in Pod arguments. Pod arguments are configuration evidence only; mounted config files, effective defaults, provider-side controls and runtime behavior are not verified.", len(components), len(missing), findings, unknown)
	return r
}

type flagCheck struct {
	flag, expected string
	bad            func(string) bool
}

func controlPlaneFlagChecks(component string) []flagCheck {
	falseValue := func(v string) bool { return strings.EqualFold(v, "false") }
	trueValue := func(v string) bool { return strings.EqualFold(v, "true") }
	switch component {
	case "kube-apiserver":
		return []flagCheck{
			{flag: "--anonymous-auth", expected: "false", bad: trueValue},
			{flag: "--authorization-mode", expected: "RBAC; do not include AlwaysAllow", bad: func(v string) bool {
				for _, x := range strings.Split(v, ",") {
					if strings.EqualFold(strings.TrimSpace(x), "AlwaysAllow") {
						return true
					}
				}
				return false
			}},
			{flag: "--profiling", expected: "false", bad: trueValue},
		}
	case "etcd":
		return []flagCheck{
			{flag: "--client-cert-auth", expected: "true", bad: falseValue},
			{flag: "--peer-client-cert-auth", expected: "true", bad: falseValue},
			{flag: "--listen-client-urls", expected: "https:// endpoints only", bad: func(v string) bool { return strings.Contains(strings.ToLower(v), "http://") }},
			{flag: "--listen-peer-urls", expected: "https:// endpoints only", bad: func(v string) bool { return strings.Contains(strings.ToLower(v), "http://") }},
		}
	case "kube-controller-manager", "kube-scheduler":
		return []flagCheck{
			{flag: "--bind-address", expected: "127.0.0.1 or a restricted control-plane address", bad: func(v string) bool { return v == "0.0.0.0" || v == "::" }},
			{flag: "--profiling", expected: "false", bad: trueValue},
		}
	default:
		return nil
	}
}

func flagValue(args []string, name string) (string, bool) {
	for i, arg := range args {
		if strings.HasPrefix(arg, name+"=") {
			return strings.TrimPrefix(arg, name+"="), true
		}
		if arg == name {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				return args[i+1], true
			}
			return "", true
		}
	}
	return "", false
}

// evaluateAuditLogging checks visible API-server audit flags. It cannot prove
// file readability, webhook delivery, retention or alert operation.
func evaluateAuditLogging(ctx context.Context, client kubernetes.Interface, _ Check, _ Options, r Result) Result {
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to inspect kube-system Pods for audit configuration: "+err.Error()
		return r
	}
	var apiserver *corev1.Pod
	for i := range pods.Items {
		if strings.Contains(strings.ToLower(pods.Items[i].Name), "kube-apiserver") {
			apiserver = &pods.Items[i]
			break
		}
	}
	if apiserver == nil {
		r.Status = "MANUAL"
		r.Details = "The API server Pod is not visible. Verify audit policy, backend, retention and delivery through provider control-plane evidence; Kubernetes API access alone cannot establish audit logging."
		return r
	}
	args := []string{}
	for _, c := range apiserver.Spec.Containers {
		args = append(args, c.Command...)
		args = append(args, c.Args...)
	}
	policy, hasPolicy := flagValue(args, "--audit-policy-file")
	logPath, hasLog := flagValue(args, "--audit-log-path")
	webhook, hasWebhook := flagValue(args, "--audit-webhook-config-file")
	hasPolicy = hasPolicy && strings.TrimSpace(policy) != ""
	hasLog = hasLog && strings.TrimSpace(logPath) != ""
	hasWebhook = hasWebhook && strings.TrimSpace(webhook) != ""
	backend := hasLog || hasWebhook
	for _, item := range []struct{ flag, value string }{{"--audit-policy-file", policy}, {"--audit-log-path", logPath}, {"--audit-webhook-config-file", webhook}} {
		if item.value != "" {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: apiserver.Namespace, Name: apiserver.Name, Field: item.flag, Observed: item.value, Expected: "configured path with independently verified file contents and operational behavior"})
		} else {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: apiserver.Namespace, Name: apiserver.Name, Field: item.flag, Observed: "not explicitly configured", Expected: "audit policy and at least one supported audit backend"})
		}
	}
	switch {
	case !backend:
		r.Status = "WARN"
		r.Details = "No --audit-log-path or --audit-webhook-config-file flag was visible in the API-server Pod arguments. Audit logging may be configured outside the visible Pod spec; verify provider-managed configuration."
	case !hasPolicy:
		r.Status = "WARN"
		r.Details = "An audit backend flag is configured, but --audit-policy-file was not visible. Audit event coverage and effective policy are unverified."
	default:
		r.Status = "MANUAL"
		r.Details = "API-server audit policy and backend paths are configured in visible Pod arguments. The referenced files, policy rule quality, write permissions, log delivery, retention and alerting cannot be verified through Pod arguments; complete operational validation is required."
	}
	return r
}

func sortedComponentNames(m map[string]controlPlaneComponent) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

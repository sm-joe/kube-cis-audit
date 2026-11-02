package audit

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// evaluateNetworkSecurity evaluates configured policy coverage and Service
// exposure. It does not perform packet probes or assert public reachability.
func evaluateNetworkSecurity(ctx context.Context, client kubernetes.Interface, _ Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list Pods for network assessment: "+err.Error()
		return r
	}
	policies, err := client.NetworkingV1().NetworkPolicies(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list NetworkPolicies: "+err.Error()
		return r
	}
	services, err := client.CoreV1().Services(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list Services: "+err.Error()
		return r
	}
	ingresses, ingressErr := client.NetworkingV1().Ingresses(o.Namespace).List(ctx, metav1.ListOptions{})

	eligiblePods, ingressIsolated, egressIsolated := 0, 0, 0
	broadPolicyRules := 0
	policySelectedCounts := make(map[string]int, len(policies.Items))
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		eligiblePods++
		in, out := false, false
		for _, policy := range policies.Items {
			if policy.Namespace != pod.Namespace {
				continue
			}
			selector, selectorErr := metav1.LabelSelectorAsSelector(&policy.Spec.PodSelector)
			if selectorErr != nil || !selector.Matches(labels.Set(pod.Labels)) {
				continue
			}
			policySelectedCounts[policy.Namespace+"/"+policy.Name]++
			types := policy.Spec.PolicyTypes
			if len(types) == 0 {
				types = []networkingv1.PolicyType{networkingv1.PolicyTypeIngress}
				if len(policy.Spec.Egress) > 0 {
					types = append(types, networkingv1.PolicyTypeEgress)
				}
			}
			for _, typ := range types {
				if typ == networkingv1.PolicyTypeIngress {
					in = true
				}
				if typ == networkingv1.PolicyTypeEgress {
					out = true
				}
			}
		}
		if in {
			ingressIsolated++
		}
		if out {
			egressIsolated++
		}
		if len(r.Evidence) < 150 {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: "networkPolicySelection", Observed: fmt.Sprintf("ingressIsolated=%t; egressIsolated=%t", in, out), Expected: "review both ingress and egress isolation requirements"})
		}
	}

	for _, policy := range policies.Items {
		if policySelectedCounts[policy.Namespace+"/"+policy.Name] == 0 && len(r.Evidence) < 150 {
			r.Evidence = append(r.Evidence, Evidence{Resource: "NetworkPolicy", Namespace: policy.Namespace, Name: policy.Name, Field: "spec.podSelector", Observed: "selects no inspected non-terminal Pod", Expected: "confirm selector labels and whether the policy is intended to select current workloads"})
		}
	}

	for _, policy := range policies.Items {
		for i, rule := range policy.Spec.Ingress {
			issues := []string{}
			if len(rule.From) == 0 {
				issues = append(issues, "empty from list permits all sources")
			}
			if len(rule.Ports) == 0 {
				issues = append(issues, "empty ports list permits all ports")
			}
			if len(issues) > 0 {
				broadPolicyRules++
				if len(r.Evidence) < 150 {
					r.Evidence = append(r.Evidence, Evidence{Resource: "NetworkPolicy", Namespace: policy.Namespace, Name: policy.Name, Field: fmt.Sprintf("spec.ingress[%d]", i), Observed: strings.Join(issues, "; "), Expected: "restrict ingress peers and ports to the required workload paths"})
				}
			}
		}
		for i, rule := range policy.Spec.Egress {
			issues := []string{}
			if len(rule.To) == 0 {
				issues = append(issues, "empty to list permits all destinations")
			}
			if len(rule.Ports) == 0 {
				issues = append(issues, "empty ports list permits all ports")
			}
			if len(issues) > 0 {
				broadPolicyRules++
				if len(r.Evidence) < 150 {
					r.Evidence = append(r.Evidence, Evidence{Resource: "NetworkPolicy", Namespace: policy.Namespace, Name: policy.Name, Field: fmt.Sprintf("spec.egress[%d]", i), Observed: strings.Join(issues, "; "), Expected: "restrict egress destinations and ports to required dependencies"})
				}
			}
		}
	}

	hostPortMappings := 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		for _, container := range podContainers(pod) {
			var ports []corev1.ContainerPort
			switch container.Kind {
			case "container":
				for _, c := range pod.Spec.Containers {
					if c.Name == container.Name {
						ports = c.Ports
					}
				}
			case "initContainer":
				for _, c := range pod.Spec.InitContainers {
					if c.Name == container.Name {
						ports = c.Ports
					}
				}
			case "ephemeralContainer":
				for _, c := range pod.Spec.EphemeralContainers {
					if c.Name == container.Name {
						ports = c.Ports
					}
				}
			}
			for _, port := range ports {
				if port.HostPort > 0 {
					hostPortMappings++
					if len(r.Evidence) < 150 {
						r.Evidence = append(r.Evidence, Evidence{Resource: container.Kind, Namespace: pod.Namespace, Name: pod.Name + "/" + container.Name, Field: "ports.hostPort", Observed: fmt.Sprintf("containerPort=%d hostPort=%d protocol=%s", port.ContainerPort, port.HostPort, port.Protocol), Expected: "avoid hostPort unless node-level exposure is explicitly required and restricted"})
					}
				}
			}
		}
	}
	publicConfigured, nodePortConfigured, externalNameConfigured, externalIPServices := 0, 0, 0, 0
	for _, svc := range services.Items {
		if len(svc.Spec.ExternalIPs) > 0 {
			externalIPServices++
		}
		switch svc.Spec.Type {
		case corev1.ServiceTypeLoadBalancer:
			publicConfigured++
		case corev1.ServiceTypeNodePort:
			nodePortConfigured++
		case corev1.ServiceTypeExternalName:
			externalNameConfigured++
		}
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer || svc.Spec.Type == corev1.ServiceTypeNodePort || svc.Spec.Type == corev1.ServiceTypeExternalName || len(svc.Spec.ExternalIPs) > 0 {
			ports := make([]string, 0, len(svc.Spec.Ports))
			for _, p := range svc.Spec.Ports {
				ports = append(ports, fmt.Sprintf("%s:%d->%d", p.Protocol, p.Port, p.TargetPort.IntValue()))
			}
			sort.Strings(ports)
			observed := fmt.Sprintf("type=%s; ports=%s; externalIPs=%s", svc.Spec.Type, strings.Join(ports, ","), strings.Join(svc.Spec.ExternalIPs, ","))
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				observed += fmt.Sprintf("; loadBalancerIngressEntries=%d", len(svc.Status.LoadBalancer.Ingress))
			}
			if len(r.Evidence) < 150 {
				r.Evidence = append(r.Evidence, Evidence{Resource: "Service", Namespace: svc.Namespace, Name: svc.Name, Field: "spec.type", Observed: observed, Expected: "confirm business justification, source restrictions and external exposure outside this API-only assessment"})
			}
		}
	}

	if ingressErr == nil {
		for _, ing := range ingresses.Items {
			hosts := make([]string, 0, len(ing.Spec.Rules))
			for _, rule := range ing.Spec.Rules {
				if rule.Host != "" {
					hosts = append(hosts, rule.Host)
				}
			}
			sort.Strings(hosts)
			tlsConfigured := len(ing.Spec.TLS) > 0
			if len(r.Evidence) < 150 {
				r.Evidence = append(r.Evidence, Evidence{Resource: "Ingress", Namespace: ing.Namespace, Name: ing.Name, Field: "spec.rules.host/spec.tls", Observed: fmt.Sprintf("hosts=%s; tlsConfigured=%t", strings.Join(hosts, ","), tlsConfigured), Expected: "review ingress class, TLS configuration, authentication and external routing independently"})
			}
		}
	} else {
		r.Evidence = append(r.Evidence, Evidence{Resource: "Ingress", Field: "inventory", Observed: "list unavailable: " + ingressErr.Error(), Expected: "Ingress inventory where networking.k8s.io/v1 is served"})
	}

	missingIngress, missingEgress := eligiblePods-ingressIsolated, eligiblePods-egressIsolated
	status := "PASS"
	if eligiblePods == 0 {
		status = "NOT_ASSESSED"
	} else if missingIngress > 0 || missingEgress > 0 {
		status = "WARN"
	}
	if (publicConfigured+nodePortConfigured+externalIPServices+hostPortMappings+broadPolicyRules) > 0 && status == "PASS" {
		status = "WARN"
	}
	if ingressErr != nil && status == "PASS" {
		status = "NOT_ASSESSED"
	}
	r.Status = status
	r.Details = fmt.Sprintf("Inspected %d non-terminal Pods, %d NetworkPolicies and %d Services. Pod isolation coverage: ingress=%d/%d, egress=%d/%d. Services configured as LoadBalancer=%d, NodePort=%d, ExternalName=%d, externalIP=%d; hostPort mappings=%d; broad/unrestricted policy rules=%d; Ingress inventory=%s. Service types indicate exposure configuration only; this audit does not verify routing, firewall rules, cloud load-balancer reachability, CNI enforcement or actual packet flow.", eligiblePods, len(policies.Items), len(services.Items), ingressIsolated, eligiblePods, egressIsolated, eligiblePods, publicConfigured, nodePortConfigured, externalNameConfigured, externalIPServices, hostPortMappings, broadPolicyRules, ingressInventoryStatus(ingressErr))
	return r
}

func ingressInventoryStatus(err error) string {
	if err != nil {
		return "unavailable"
	}
	return "available"
}

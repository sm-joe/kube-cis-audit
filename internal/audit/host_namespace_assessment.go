package audit

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// evaluatePodHostNamespaces assesses whether non-terminal Pods share the host's
// PID, IPC, or network namespace. These are PodSpec fields, so each Pod is
// evaluated once rather than once per container.
//
// Kubernetes defaults all three fields to false. The API representation uses
// plain bools, so an omitted field and an explicitly configured false value
// cannot be distinguished after decoding; both are reported as effectively
// disabled. This evaluates the Pod specification, not the effective runtime
// namespace configuration.
func evaluatePodHostNamespaces(ctx context.Context, client kubernetes.Interface, _ Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to list Pods for host namespace assessment: " + err.Error()
		return r
	}

	checkedPods := 0
	hostPIDPods := 0
	hostIPCPods := 0
	hostNetworkPods := 0
	appendEvidence := func(e Evidence) {
		if len(r.Evidence) < 100 {
			r.Evidence = append(r.Evidence, e)
		}
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		checkedPods++

		assess := func(field string, enabled bool, counter *int) {
			observed := "false (host namespace sharing disabled; Kubernetes default false)"
			expected := "false; workload must not share the host namespace"
			if enabled {
				(*counter)++
				observed = "true (host namespace sharing enabled)"
			}
			appendEvidence(Evidence{
				Resource:  "Pod",
				Namespace: pod.Namespace,
				Name:      pod.Name,
				Field:     field,
				Observed:  observed,
				Expected:  expected,
			})
		}

		assess("spec.hostPID", pod.Spec.HostPID, &hostPIDPods)
		assess("spec.hostIPC", pod.Spec.HostIPC, &hostIPCPods)
		assess("spec.hostNetwork", pod.Spec.HostNetwork, &hostNetworkPods)
	}

	if checkedPods == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No non-terminal Pods were available for host namespace assessment."
		return r
	}

	totalEnabled := hostPIDPods + hostIPCPods + hostNetworkPods
	if totalEnabled > 0 {
		r.Status = "FAIL"
	} else {
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf(
		"Inspected %d non-terminal Pods. Host namespace sharing enabled: hostPID=%d Pods, hostIPC=%d Pods, hostNetwork=%d Pods. A Pod with any enabled host namespace setting fails this check. The assessment uses Pod-spec values; effective runtime namespace state is not verified.",
		checkedPods, hostPIDPods, hostIPCPods, hostNetworkPods,
	)
	return r
}

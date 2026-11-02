package audit

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// evaluatePodCapabilities assesses whether workloads explicitly drop all
// Linux capabilities and whether they add capabilities back. It does not
// infer the runtime's effective capability set because that depends on the
// container runtime and its configuration.
//
// A container that drops ALL and adds nothing passes. A container that does
// not drop ALL, or adds any capability, is warned for review. Added
// capabilities are not automatically classified as insecure because their
// necessity is workload-specific. CAP_SYS_ADMIN is separately evaluated by
// the privilege-escalation check.
func evaluatePodCapabilities(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to list Pods for Linux capability assessment: " + err.Error()
		return r
	}

	checkedPods := 0
	checkedContainers := 0
	passingContainers := 0
	containersMissingDropAll := 0
	containersWithAddedCapabilities := 0

	appendEvidence := func(e Evidence) {
		if len(r.Evidence) < 60 {
			r.Evidence = append(r.Evidence, e)
		}
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}

		containers := podContainers(pod)
		if len(containers) == 0 {
			continue
		}
		checkedPods++

		for _, container := range containers {
			checkedContainers++
			var added, dropped []string
			if container.SecurityContext != nil && container.SecurityContext.Capabilities != nil {
				for _, capability := range container.SecurityContext.Capabilities.Add {
					added = append(added, normalizeCapability(string(capability)))
				}
				for _, capability := range container.SecurityContext.Capabilities.Drop {
					dropped = append(dropped, normalizeCapability(string(capability)))
				}
			}

			dropsAll := containsCapability(dropped, "ALL")
			if !dropsAll {
				containersMissingDropAll++
				observed := "capabilities.drop unset or does not include ALL"
				if len(dropped) > 0 {
					observed = "capabilities.drop=" + strings.Join(dropped, ",") + " (ALL not dropped)"
				}
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:  pod.Name + "/" + container.Name,
					Field: "securityContext.capabilities.drop", Observed: observed,
					Expected: "ALL",
				})
			} else {
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:  pod.Name + "/" + container.Name,
					Field: "securityContext.capabilities.drop", Observed: "ALL dropped",
					Expected: "ALL",
				})
			}

			if len(added) > 0 {
				containersWithAddedCapabilities++
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:  pod.Name + "/" + container.Name,
					Field: "securityContext.capabilities.add", Observed: "added=" + strings.Join(added, ","),
					Expected: "no added capabilities unless individually justified and approved",
				})
			} else {
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:  pod.Name + "/" + container.Name,
					Field: "securityContext.capabilities.add", Observed: "no added capabilities",
					Expected: "no added capabilities unless individually justified and approved",
				})
			}

			if dropsAll && len(added) == 0 {
				passingContainers++
			}
		}
	}

	if checkedContainers == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No non-terminal Pods with inspectable containers were available for Linux capability assessment."
		return r
	}

	if containersMissingDropAll > 0 || containersWithAddedCapabilities > 0 {
		r.Status = "WARN"
	} else {
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf(
		"Inspected %d containers across %d non-terminal Pods. %d containers drop ALL capabilities and add none; %d do not explicitly drop ALL; %d add one or more capabilities. Added capabilities require workload-specific justification. Effective runtime capabilities are not inferred.",
		checkedContainers, checkedPods, passingContainers, containersMissingDropAll, containersWithAddedCapabilities,
	)
	return r
}

func normalizeCapability(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "CAP_")
	return value
}

func containsCapability(capabilities []string, wanted string) bool {
	for _, capability := range capabilities {
		if normalizeCapability(capability) == wanted {
			return true
		}
	}
	return false
}

package audit

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// evaluatePodPrivilegeSecurity assesses privileged mode and privilege
// escalation for regular, init, and ephemeral containers. These settings are
// container-scoped; there is no PodSecurityContext inheritance for them.
//
// Kubernetes defaults privileged to false and allowPrivilegeEscalation to
// true. The latter is also effectively true for privileged containers and
// containers adding CAP_SYS_ADMIN, even if allowPrivilegeEscalation=false.
func evaluatePodPrivilegeSecurity(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to list Pods for privilege assessment: " + err.Error()
		return r
	}

	checkedPods, checkedContainers := 0, 0
	failures, passed := 0, 0
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
			sc := container.SecurityContext
			privileged := sc != nil && sc.Privileged != nil && *sc.Privileged
			allowEscalation := sc == nil || sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation
			sysAdmin := false
			if sc != nil && sc.Capabilities != nil {
				for _, capability := range sc.Capabilities.Add {
					if strings.EqualFold(string(capability), "SYS_ADMIN") {
						sysAdmin = true
						break
					}
				}
			}

			privObserved := "privileged unset (Kubernetes default false)"
			if sc != nil && sc.Privileged != nil {
				privObserved = fmt.Sprintf("privileged=%t", *sc.Privileged)
			}
			privStatus := "PASS"
			if privileged {
				privStatus = "FAIL"
				failures++
			} else {
				passed++
			}
			appendEvidence(Evidence{
				Resource: container.Kind, Namespace: pod.Namespace,
				Name:  pod.Name + "/" + container.Name,
				Field: "securityContext.privileged", Observed: privObserved + " (" + strings.ToLower(privStatus) + ")",
				Expected: "false; privileged containers are not permitted by this check",
			})

			allowObserved := "allowPrivilegeEscalation unset (Kubernetes default true)"
			if sc != nil && sc.AllowPrivilegeEscalation != nil {
				allowObserved = fmt.Sprintf("allowPrivilegeEscalation=%t", *sc.AllowPrivilegeEscalation)
			}
			effectiveEscalation := allowEscalation || privileged || sysAdmin
			allowStatus := "PASS"
			switch {
			case privileged:
				allowObserved += "; privileged=true forces effective escalation"
				allowStatus = "FAIL"
			case sysAdmin:
				allowObserved += "; CAP_SYS_ADMIN forces effective escalation"
				allowStatus = "FAIL"
			case effectiveEscalation:
				allowStatus = "FAIL"
			case sc == nil || sc.AllowPrivilegeEscalation == nil:
				allowStatus = "FAIL"
			}
			if allowStatus == "FAIL" {
				failures++
			} else {
				passed++
			}
			appendEvidence(Evidence{
				Resource: container.Kind, Namespace: pod.Namespace,
				Name:  pod.Name + "/" + container.Name,
				Field: "securityContext.allowPrivilegeEscalation", Observed: allowObserved + " (" + strings.ToLower(allowStatus) + ")",
				Expected: "false, with neither privileged mode nor CAP_SYS_ADMIN",
			})
		}
	}

	if checkedContainers == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No non-terminal Pods with inspectable containers were available for privilege assessment."
		return r
	}
	if failures > 0 {
		r.Status = "FAIL"
	} else {
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf(
		"Inspected %d containers across %d non-terminal Pods. Found %d failing field observations and %d passing field observations. Missing allowPrivilegeEscalation is treated as effectively true by the Kubernetes default; privileged mode and CAP_SYS_ADMIN also force effective privilege escalation.",
		checkedContainers, checkedPods, failures, passed,
	)
	return r
}

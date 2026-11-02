package audit

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// evaluatePodGroupSecurityContext inventories effective runAsGroup values for
// each regular, init, and ephemeral container, and fsGroup once per Pod.
// It is a KubeCIS informational extension, not an official CIS benchmark check.
func evaluatePodGroupSecurityContext(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to list Pods for group security-context assessment: " + err.Error()
		return r
	}

	checkedPods := 0
	checkedContainers := 0
	runGroupSet, runGroupUnset := 0, 0
	fsGroupSet, fsGroupUnset := 0, 0
	warnings, failures := 0, 0
	configuredFields := 0

	appendEvidence := func(e Evidence) {
		if len(r.Evidence) < 30 {
			r.Evidence = append(r.Evidence, e)
		}
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		checkedPods++

		containers := podContainers(pod)
		for _, container := range containers {
			checkedContainers++
			effective := resolveEffectiveSecurityContext(pod.Spec.SecurityContext, container.SecurityContext)
			assessment := assessGroupID("runAsGroup", effective.RunAsGroup)
			if effective.RunAsGroup == nil {
				runGroupUnset++
			} else {
				runGroupSet++
				configuredFields++
			}
			switch assessment.Status {
			case "WARN":
				warnings++
			case "FAIL":
				failures++
			}
			appendEvidence(Evidence{
				Resource:  container.Kind,
				Namespace: pod.Namespace,
				Name:      pod.Name + "/" + container.Name,
				Field:     assessment.Field,
				Observed:  assessment.Observed,
				Expected:  "review explicit GID; non-zero GID is informational, GID 0 warrants review",
			})
		}

		var fsGroup *int64
		if pod.Spec.SecurityContext != nil {
			fsGroup = pod.Spec.SecurityContext.FSGroup
		}
		fsAssessment := assessFSGroup(fsGroup)
		if fsGroup == nil {
			fsGroupUnset++
		} else {
			fsGroupSet++
			configuredFields++
		}
		switch fsAssessment.Status {
		case "WARN":
			warnings++
		case "FAIL":
			failures++
		}
		appendEvidence(Evidence{
			Resource:  "Pod",
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Field:     fsAssessment.Field,
			Observed:  fsAssessment.Observed,
			Expected:  "review volume ownership requirements; non-zero fsGroup is informational",
		})
	}

	if checkedPods == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No non-terminal Pods were available for group security-context assessment."
		return r
	}

	switch {
	case failures > 0:
		r.Status = "FAIL"
	case warnings > 0:
		r.Status = "WARN"
	case configuredFields > 0:
		r.Status = "INFO"
	default:
		r.Status = "NOT_ASSESSED"
	}

	r.Details = fmt.Sprintf(
		"Inspected %d non-terminal Pods and %d containers. runAsGroup: %d explicitly configured, %d unset; fsGroup: %d explicitly configured, %d unset. GID 0 is reported as a warning; non-zero IDs are informational and do not prove least-privilege permissions or effective volume ownership.",
		checkedPods, checkedContainers, runGroupSet, runGroupUnset, fsGroupSet, fsGroupUnset,
	)
	return r
}

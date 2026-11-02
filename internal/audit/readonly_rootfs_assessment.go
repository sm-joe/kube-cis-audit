package audit

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// evaluatePodReadOnlyRootFilesystem checks the container root filesystem setting
// independently from volume mount writability. A writable volume mount does not
// make the container root filesystem writable, and a read-only root filesystem
// does not make mounted volumes read-only.
//
// Kubernetes treats an omitted readOnlyRootFilesystem field as false. Therefore
// both an explicit false and an unset value fail this hardening check, but the
// evidence distinguishes the two cases. Writable volume mounts are WARN because
// many workloads legitimately need writable data volumes and the API inventory
// cannot determine whether each mount is required.
//
// This KubeCIS extension evaluates Pod specifications, not the effective runtime
// mount namespace. It does not assess volumeDevices or host-level mount state.
func evaluatePodReadOnlyRootFilesystem(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to list Pods for read-only root filesystem assessment: " + err.Error()
		return r
	}

	checkedPods := 0
	checkedContainers := 0
	rootReadOnly := 0
	rootWritable := 0
	rootUnset := 0
	writableMounts := 0
	readOnlyMounts := 0
	failures := 0
	appendEvidence := func(e Evidence) {
		if len(r.Evidence) < 100 {
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
			switch {
			case sc == nil || sc.ReadOnlyRootFilesystem == nil:
				rootUnset++
				failures++
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:     pod.Name + "/" + container.Name,
					Field:    "securityContext.readOnlyRootFilesystem",
					Observed: "readOnlyRootFilesystem unset (Kubernetes default false; writable root filesystem)",
					Expected: "true",
				})
			case !*sc.ReadOnlyRootFilesystem:
				rootWritable++
				failures++
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:     pod.Name + "/" + container.Name,
					Field:    "securityContext.readOnlyRootFilesystem",
					Observed: "readOnlyRootFilesystem=false (writable root filesystem)",
					Expected: "true",
				})
			default:
				rootReadOnly++
				appendEvidence(Evidence{
					Resource: container.Kind, Namespace: pod.Namespace,
					Name:     pod.Name + "/" + container.Name,
					Field:    "securityContext.readOnlyRootFilesystem",
					Observed: "readOnlyRootFilesystem=true",
					Expected: "true",
				})
			}

			// VolumeMount.ReadOnly is a non-pointer bool. An omitted value and
			// explicit false are indistinguishable after API decoding; both mean
			// the mount is writable according to the Pod specification.
			for _, mount := range containerVolumeMounts(pod, container) {
				if mount.ReadOnly {
					readOnlyMounts++
					appendEvidence(Evidence{
						Resource: container.Kind, Namespace: pod.Namespace,
						Name:     pod.Name + "/" + container.Name,
						Field:    "volumeMount.readOnly",
						Observed: fmt.Sprintf("volume=%s mountPath=%s readOnly=true", mount.Name, mount.MountPath),
						Expected: "review mount access; readOnly=true where the workload does not need writes",
					})
				} else {
					writableMounts++
					observed := fmt.Sprintf("volume=%s mountPath=%s readOnly=false (writable mount)", mount.Name, mount.MountPath)
					if mount.SubPath != "" {
						observed += " subPath=" + mount.SubPath
					}
					if mount.SubPathExpr != "" {
						observed += " subPathExpr=" + mount.SubPathExpr
					}
					appendEvidence(Evidence{
						Resource: container.Kind, Namespace: pod.Namespace,
						Name:     pod.Name + "/" + container.Name,
						Field:    "volumeMount.readOnly",
						Observed: observed,
						Expected: "review whether write access is required; set readOnly=true where possible",
					})
				}
			}
		}
	}

	if checkedContainers == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No non-terminal Pods with inspectable containers were available for read-only root filesystem assessment."
		return r
	}

	switch {
	case failures > 0:
		r.Status = "FAIL"
	case writableMounts > 0:
		r.Status = "WARN"
	default:
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf(
		"Inspected %d containers across %d non-terminal Pods. Root filesystem: %d read-only, %d explicitly writable, %d unset (Kubernetes default writable). Volume mounts: %d read-only, %d writable. Writable mounts are reported separately because they do not change the root filesystem setting; volumeDevices and effective runtime mount state are not assessed.",
		checkedContainers, checkedPods, rootReadOnly, rootWritable, rootUnset, readOnlyMounts, writableMounts,
	)
	return r
}

// containerVolumeMounts returns the mounts for a specific podContainer by name
// and kind. The canonical podContainers helper supplies regular, init and
// ephemeral containers; this function retrieves their corresponding mounts.
func containerVolumeMounts(pod corev1.Pod, target podContainer) []corev1.VolumeMount {
	switch target.Kind {
	case "container":
		for _, container := range pod.Spec.Containers {
			if container.Name == target.Name {
				return container.VolumeMounts
			}
		}
	case "initContainer":
		for _, container := range pod.Spec.InitContainers {
			if container.Name == target.Name {
				return container.VolumeMounts
			}
		}
	case "ephemeralContainer":
		for _, container := range pod.Spec.EphemeralContainers {
			if container.Name == target.Name {
				return container.VolumeMounts
			}
		}
	}
	return nil
}

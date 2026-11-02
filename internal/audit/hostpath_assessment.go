package audit

import (
	"context"
	"fmt"
	"path"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type hostPathMount struct {
	Container podContainer
	Mount     *corev1.VolumeMount
	Device    *corev1.VolumeDevice
}

// evaluatePodHostPath inventories Pod HostPath volumes and every matching
// regular/init/ephemeral container mount. It classifies paths lexically; it
// does not resolve symlinks or inspect the host filesystem.
func evaluatePodHostPath(ctx context.Context, client kubernetes.Interface, _ Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to list Pods for HostPath assessment: " + err.Error()
		return r
	}

	checkedPods := 0
	hostPathVolumes := 0
	mountedHostPathVolumes := 0
	writableSensitive := 0
	readOnlySensitive := 0
	writableOther := 0
	readOnlyOther := 0
	unmounted := 0
	severity := 0 // 0=PASS, 1=INFO, 2=WARN, 3=FAIL

	appendEvidence := func(e Evidence) {
		if len(r.Evidence) < 150 {
			r.Evidence = append(r.Evidence, e)
		}
	}
	raise := func(level int) {
		if level > severity {
			severity = level
		}
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		checkedPods++
		containers := podContainers(pod)

		for _, volume := range pod.Spec.Volumes {
			if volume.HostPath == nil {
				continue
			}
			hostPathVolumes++
			rawPath := volume.HostPath.Path
			cleanPath := path.Clean(rawPath)
			sensitive, category := classifyHostPath(cleanPath)

			pathObserved := fmt.Sprintf("path=%q normalized=%q classification=%s", rawPath, cleanPath, category)
			if !path.IsAbs(rawPath) || strings.TrimSpace(rawPath) == "" {
				pathObserved += "; invalid non-absolute HostPath"
				raise(2)
			}
			hostType := "<unset>"
			if volume.HostPath.Type != nil {
				hostType = string(*volume.HostPath.Type)
			}
			appendEvidence(Evidence{
				Resource: "HostPathVolume", Namespace: pod.Namespace,
				Name: pod.Name + "/" + volume.Name, Field: "spec.volumes.hostPath.path",
				Observed: pathObserved + "; type=" + hostType,
				Expected: "avoid HostPath; otherwise use a narrow, documented path and an explicit compatible HostPath type",
			})

			mounts := matchingHostPathMounts(pod, containers, volume.Name)
			if len(mounts) == 0 {
				unmounted++
				raise(2)
				appendEvidence(Evidence{
					Resource: "HostPathVolume", Namespace: pod.Namespace,
					Name: pod.Name + "/" + volume.Name, Field: "volumeMount",
					Observed: "HostPath volume is not referenced by a regular, init, or ephemeral container mount",
					Expected: "remove unused HostPath volume declarations; verify any non-container consumers separately",
				})
				continue
			}

			mountedHostPathVolumes++
			for _, item := range mounts {
				readOnly := item.Mount != nil && item.Mount.ReadOnly
				mountPath := ""
				field := "volumeMount.readOnly"
				mountKind := "volumeMount"
				if item.Mount != nil {
					mountPath = item.Mount.MountPath
				} else if item.Device != nil {
					mountPath = item.Device.DevicePath
					field = "volumeDevices.devicePath"
					mountKind = "volumeDevice"
				}

				mountStatus := "INFO"
				if readOnly {
					if sensitive {
						mountStatus = "WARN"
						readOnlySensitive++
						raise(2)
					} else {
						readOnlyOther++
						raise(1)
					}
				} else if sensitive {
					mountStatus = "FAIL"
					writableSensitive++
					raise(3)
				} else {
					mountStatus = "WARN"
					writableOther++
					raise(2)
				}

				observed := fmt.Sprintf(
					"volume=%s hostPath=%q %s=%q readOnly=%t classification=%s assessment=%s",
					volume.Name, rawPath, mountKind, mountPath, readOnly, category, mountStatus,
				)
				expected := "avoid HostPath; if required, mount the narrowest required path read-only"
				if item.Device != nil {
					expected = "avoid exposing host block devices; use a dedicated storage driver and least-privilege device access"
				}
				appendEvidence(Evidence{
					Resource: item.Container.Kind, Namespace: pod.Namespace,
					Name:  pod.Name + "/" + item.Container.Name,
					Field: field, Observed: observed, Expected: expected,
				})

				if item.Mount != nil && item.Mount.MountPropagation != nil {
					propagation := string(*item.Mount.MountPropagation)
					propagationStatus := "INFO"
					switch *item.Mount.MountPropagation {
					case corev1.MountPropagationBidirectional:
						propagationStatus = "FAIL"
						raise(3)
					case corev1.MountPropagationHostToContainer:
						propagationStatus = "WARN"
						raise(2)
					}
					appendEvidence(Evidence{
						Resource: item.Container.Kind, Namespace: pod.Namespace,
						Name:     pod.Name + "/" + item.Container.Name,
						Field:    "volumeMount.mountPropagation",
						Observed: fmt.Sprintf("volume=%s propagation=%s assessment=%s", volume.Name, propagation, propagationStatus),
						Expected: "unset or None; Bidirectional propagation is not permitted by this assessment",
					})
				}
			}
		}
	}

	if checkedPods == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No non-terminal Pods were available for HostPath assessment."
		return r
	}

	switch severity {
	case 3:
		r.Status = "FAIL"
	case 2:
		r.Status = "WARN"
	case 1:
		r.Status = "INFO"
	default:
		r.Status = "PASS"
	}
	r.Details = fmt.Sprintf(
		"Inspected %d non-terminal Pods and found %d HostPath volume(s), of which %d were mounted. Mounts: %d writable sensitive, %d read-only sensitive, %d writable other, %d read-only other; %d HostPath volume(s) were not mounted. Sensitive writable paths and Bidirectional mount propagation fail; sensitive read-only paths and other writable HostPath mounts require review. Path classification is lexical and does not resolve symlinks or verify runtime mount state.",
		checkedPods, hostPathVolumes, mountedHostPathVolumes, writableSensitive, readOnlySensitive, writableOther, readOnlyOther, unmounted,
	)
	return r
}

func matchingHostPathMounts(pod corev1.Pod, containers []podContainer, volumeName string) []hostPathMount {
	var matches []hostPathMount
	for _, container := range containers {
		for _, mount := range containerVolumeMounts(pod, container) {
			if mount.Name == volumeName {
				mountCopy := mount
				matches = append(matches, hostPathMount{Container: container, Mount: &mountCopy})
			}
		}
		for _, device := range containerVolumeDevices(pod, container) {
			if device.Name == volumeName {
				deviceCopy := device
				matches = append(matches, hostPathMount{Container: container, Device: &deviceCopy})
			}
		}
	}
	return matches
}

func containerVolumeDevices(pod corev1.Pod, target podContainer) []corev1.VolumeDevice {
	switch target.Kind {
	case "container":
		for _, container := range pod.Spec.Containers {
			if container.Name == target.Name {
				return container.VolumeDevices
			}
		}
	case "initContainer":
		for _, container := range pod.Spec.InitContainers {
			if container.Name == target.Name {
				return container.VolumeDevices
			}
		}
	case "ephemeralContainer":
		for _, container := range pod.Spec.EphemeralContainers {
			if container.Name == target.Name {
				return container.VolumeDevices
			}
		}
	}
	return nil
}

// classifyHostPath returns whether a path is sensitive and a human-readable
// category. Matching uses path boundaries, so /etc2 does not match /etc.
func classifyHostPath(hostPath string) (bool, string) {
	clean := path.Clean(hostPath)
	if !path.IsAbs(hostPath) || strings.TrimSpace(hostPath) == "" {
		return false, "invalid"
	}
	if clean == "/" {
		return true, "host-root"
	}
	sensitivePrefixes := []struct {
		path     string
		category string
	}{
		{"/etc", "system-configuration"},
		{"/proc", "process-kernel-interface"},
		{"/sys", "kernel-device-interface"},
		{"/dev", "device-files"},
		{"/run", "runtime-state"},
		{"/var/run", "runtime-state"},
		{"/var/lib/kubelet", "kubernetes-node-state"},
		{"/var/lib/containerd", "container-runtime-state"},
		{"/var/lib/docker", "container-runtime-state"},
		{"/var/log", "host-logs"},
		{"/root", "root-home"},
		{"/home", "host-user-data"},
		{"/boot", "boot-files"},
		{"/usr", "system-binaries"},
	}
	for _, item := range sensitivePrefixes {
		if clean == item.path || strings.HasPrefix(clean, strings.TrimSuffix(item.path, "/")+"/") {
			return true, item.category
		}
	}
	return false, "other-host-path"
}

package audit

import (
	"context"
	"errors"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func hostPathPod(name string, volumes []corev1.Volume, containers []corev1.Container, init []corev1.Container, ephemeral []corev1.EphemeralContainer) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "apps", Name: name},
		Spec: corev1.PodSpec{
			Volumes: volumes, Containers: containers, InitContainers: init, EphemeralContainers: ephemeral,
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
}

func TestEvaluatePodHostPathRiskClassification(t *testing.T) {
	tests := []struct {
		name       string
		hostPath   string
		readOnly   bool
		wantStatus string
		wantText   string
	}{
		{"writable host root is critical", "/", false, "FAIL", "host-root"},
		{"writable etc is critical", "/etc", false, "FAIL", "system-configuration"},
		{"writable kubelet state is critical", "/var/lib/kubelet/pods", false, "FAIL", "kubernetes-node-state"},
		{"read-only etc still warns", "/etc", true, "WARN", "system-configuration"},
		{"writable application data warns", "/srv/app-data", false, "WARN", "other-host-path"},
		{"read-only application data is informational", "/srv/app-data", true, "INFO", "other-host-path"},
		{"path boundary avoids etc2 classification", "/etc2/config", false, "WARN", "other-host-path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := hostPathPod("sample",
				[]corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: tt.hostPath}}}},
				[]corev1.Container{{Name: "app", VolumeMounts: []corev1.VolumeMount{{Name: "host", MountPath: "/host", ReadOnly: tt.readOnly}}}},
				nil, nil)
			result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{Namespace: "apps"}, Result{})
			if result.Status != tt.wantStatus {
				t.Fatalf("status=%s want=%s details=%s", result.Status, tt.wantStatus, result.Details)
			}
			found := false
			for _, evidence := range result.Evidence {
				if strings.Contains(evidence.Observed, tt.wantText) {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected evidence containing %q; got %#v", tt.wantText, result.Evidence)
			}
		})
	}
}

func TestEvaluatePodHostPathIncludesInitAndEphemeralContainers(t *testing.T) {
	volume := corev1.Volume{Name: "host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc"}}}
	pod := hostPathPod("workload", []corev1.Volume{volume}, nil,
		[]corev1.Container{{Name: "init", VolumeMounts: []corev1.VolumeMount{{Name: "host", MountPath: "/host-init", ReadOnly: true}}}},
		[]corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name: "debug", VolumeMounts: []corev1.VolumeMount{{Name: "host", MountPath: "/host-debug", ReadOnly: true}},
		}}})
	result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{}, Result{})
	if result.Status != "WARN" {
		t.Fatalf("status=%s want=WARN details=%s", result.Status, result.Details)
	}
	foundInit, foundEphemeral := false, false
	for _, e := range result.Evidence {
		foundInit = foundInit || (e.Resource == "initContainer" && strings.Contains(e.Name, "/init"))
		foundEphemeral = foundEphemeral || (e.Resource == "ephemeralContainer" && strings.Contains(e.Name, "/debug"))
	}
	if !foundInit || !foundEphemeral {
		t.Fatalf("expected init and ephemeral mount evidence; got %#v", result.Evidence)
	}
}

func TestEvaluatePodHostPathMountPropagation(t *testing.T) {
	bidirectional := corev1.MountPropagationBidirectional
	pod := hostPathPod("propagation",
		[]corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/srv/data"}}}},
		[]corev1.Container{{Name: "app", VolumeMounts: []corev1.VolumeMount{{
			Name: "host", MountPath: "/data", ReadOnly: true, MountPropagation: &bidirectional,
		}}}}, nil, nil)
	result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{}, Result{})
	if result.Status != "FAIL" {
		t.Fatalf("status=%s want=FAIL details=%s", result.Status, result.Details)
	}
	found := false
	for _, e := range result.Evidence {
		if e.Field == "volumeMount.mountPropagation" && strings.Contains(e.Observed, "Bidirectional") {
			found = true
		}
	}
	if !found {
		t.Fatalf("Bidirectional propagation evidence missing: %#v", result.Evidence)
	}
}

func TestEvaluatePodHostPathBlockDeviceMapping(t *testing.T) {
	pod := hostPathPod("device",
		[]corev1.Volume{{Name: "host-device", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{
			Path: "/dev/sda", Type: hostPathTypePtr(corev1.HostPathBlockDev),
		}}}},
		[]corev1.Container{{Name: "app", VolumeDevices: []corev1.VolumeDevice{{Name: "host-device", DevicePath: "/dev/xvdb"}}}},
		nil, nil)
	result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{}, Result{})
	if result.Status != "FAIL" {
		t.Fatalf("status=%s want=FAIL details=%s", result.Status, result.Details)
	}
	found := false
	for _, e := range result.Evidence {
		if e.Field == "volumeDevices.devicePath" && strings.Contains(e.Observed, "/dev/xvdb") {
			found = true
		}
	}
	if !found {
		t.Fatalf("block-device evidence missing: %#v", result.Evidence)
	}
}

func hostPathTypePtr(value corev1.HostPathType) *corev1.HostPathType {
	return &value
}

func TestEvaluatePodHostPathUnmountedVolumeWarns(t *testing.T) {
	pod := hostPathPod("unused",
		[]corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc"}}}},
		[]corev1.Container{{Name: "app"}}, nil, nil)
	result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{}, Result{})
	if result.Status != "WARN" {
		t.Fatalf("status=%s want=WARN details=%s", result.Status, result.Details)
	}
}

func TestEvaluatePodHostPathNoHostPathPasses(t *testing.T) {
	pod := hostPathPod("ordinary", nil, []corev1.Container{{Name: "app"}}, nil, nil)
	result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{}, Result{})
	if result.Status != "PASS" {
		t.Fatalf("status=%s want=PASS details=%s", result.Status, result.Details)
	}
}

func TestEvaluatePodHostPathExcludesTerminalPods(t *testing.T) {
	pod := hostPathPod("completed",
		[]corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc"}}}},
		[]corev1.Container{{Name: "app", VolumeMounts: []corev1.VolumeMount{{Name: "host", MountPath: "/etc"}}}}, nil, nil)
	pod.Status.Phase = corev1.PodSucceeded
	result := evaluatePodHostPath(context.Background(), fake.NewSimpleClientset(pod), Check{}, Options{}, Result{})
	if result.Status != "NOT_ASSESSED" {
		t.Fatalf("status=%s want=NOT_ASSESSED", result.Status)
	}
}

func TestEvaluateDispatchesHostPathBenchmarkKind(t *testing.T) {
	pod := hostPathPod("dispatch",
		[]corev1.Volume{{Name: "host", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc"}}}},
		[]corev1.Container{{Name: "app", VolumeMounts: []corev1.VolumeMount{{Name: "host", MountPath: "/host"}}}},
		nil, nil)
	check := Check{ID: "KUBECIS-2.4.6", Kind: "api_pod_hostpath"}
	result := evaluate(context.Background(), fake.NewSimpleClientset(pod), check, Options{Namespace: "apps"})
	if result.Status != "FAIL" {
		t.Fatalf("status=%s want=FAIL details=%s", result.Status, result.Details)
	}
	if result.ID != check.ID || result.Method != check.Kind {
		t.Fatalf("benchmark metadata not preserved: %#v", result)
	}
}

func TestEvaluatePodHostPathListError(t *testing.T) {
	client := fake.NewSimpleClientset()
	client.PrependReactor("list", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("simulated Pod list failure")
	})
	result := evaluatePodHostPath(context.Background(), client, Check{}, Options{}, Result{})
	if result.Status != "NOT_ASSESSED" || !strings.Contains(result.Details, "simulated Pod list failure") {
		t.Fatalf("unexpected result: %#v", result)
	}
}

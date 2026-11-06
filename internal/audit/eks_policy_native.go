package audit

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

// EvaluateEKSPodAdmissionControl evaluates CIS EKS 2.1.0 controls 4.2.1-4.2.5.
// Namespace PSA enforcement is verified as well as existing workload violations.
// A Pod inventory alone does not establish that admission rejects future Pods.
func EvaluateEKSPodAdmissionControl(ctx context.Context, client kubernetes.Interface, id string) (Result, error) {
	fields := map[string]string{
		"4.2.1": "privileged containers",
		"4.2.2": "host PID namespace",
		"4.2.3": "host IPC namespace",
		"4.2.4": "host network namespace",
		"4.2.5": "privilege escalation",
	}
	label, ok := fields[id]
	if !ok {
		return Result{}, fmt.Errorf("unsupported CIS EKS PSS control %q", id)
	}
	r := Result{ID: id, Title: "Restrict " + label + " at Pod admission", Target: "policies", Status: "NOT_ASSESSED", Method: "kubernetes_psa_and_pods", Scored: true,
		Details: "Pod admission policy and existing Pods must be inspected.", Remediation: "Enforce appropriate Pod Security Standards or independently verify an equivalent admission policy; review approved platform exemptions."}
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, fmt.Errorf("list namespace PSA configuration: %w", err)
	}
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, fmt.Errorf("list Pods for admission control: %w", err)
	}
	if len(namespaces.Items) == 0 {
		r.Details = "No namespaces were returned; cannot establish admission enforcement."
		return r, nil
	}
	required := "baseline"
	if id == "4.2.5" {
		required = "restricted"
	}
	uncovered := 0
	for _, ns := range namespaces.Items {
		level := ns.Labels["pod-security.kubernetes.io/enforce"]
		enforcing := level == "restricted" || (required == "baseline" && level == "baseline")
		if !enforcing {
			uncovered++
		}
		r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "pod-security.kubernetes.io/enforce", Observed: firstNonEmpty(level, "<unset>"), Expected: required + " or stronger"})
	}
	violations := 0
	platformViolations := 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		var findings []string
		if id == "4.2.2" && pod.Spec.HostPID {
			findings = append(findings, "hostPID=true")
		}
		if id == "4.2.3" && pod.Spec.HostIPC {
			findings = append(findings, "hostIPC=true")
		}
		if id == "4.2.4" && pod.Spec.HostNetwork {
			findings = append(findings, "hostNetwork=true")
		}
		scan := func(kind string, containers []corev1.Container) {
			for _, container := range containers {
				sc := container.SecurityContext
				if id == "4.2.1" && sc != nil && sc.Privileged != nil && *sc.Privileged {
					findings = append(findings, kind+"/"+container.Name+":privileged=true")
				}
				if id == "4.2.5" && (sc == nil || sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation) {
					findings = append(findings, kind+"/"+container.Name+":allowPrivilegeEscalation not false")
				}
			}
		}
		scan("container", pod.Spec.Containers)
		scan("initContainer", pod.Spec.InitContainers)
		if id == "4.2.1" || id == "4.2.5" {
			for _, ec := range pod.Spec.EphemeralContainers {
				sc := ec.SecurityContext
				if id == "4.2.1" && sc != nil && sc.Privileged != nil && *sc.Privileged {
					findings = append(findings, "ephemeral/"+ec.Name+":privileged=true")
				}
				if id == "4.2.5" && (sc == nil || sc.AllowPrivilegeEscalation == nil || *sc.AllowPrivilegeEscalation) {
					findings = append(findings, "ephemeral/"+ec.Name+":allowPrivilegeEscalation not false")
				}
			}
		}
		if len(findings) > 0 {
			if pod.Namespace == "kube-system" {
				platformViolations++
			} else {
				violations++
			}
			r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: label, Observed: strings.Join(findings, "; "), Expected: "no violation, or documented/approved platform exception"})
		}
	}
	switch {
	case violations > 0:
		r.Status = "FAIL"
		r.Details = fmt.Sprintf("%d non-platform Pods violate %s; %d platform Pods require exception review; %d namespaces lack confirmed PSA enforcement.", violations, label, platformViolations, uncovered)
	case platformViolations > 0:
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("%d kube-system Pods have configurations requiring an approved platform exception; %d namespaces lack confirmed PSA enforcement. Do not treat required AWS platform agents as automatically non-compliant.", platformViolations, uncovered)
	case uncovered > 0:
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("%d namespaces lack confirmed %s PSA enforce labels; an equivalent admission engine or approved exemptions must be reviewed. No currently violating Pods found.", uncovered, required)
	default:
		r.Status = "PASS"
		r.Details = fmt.Sprintf("All %d namespaces enforce %s or stronger PSA and no existing Pods violate %s.", len(namespaces.Items), required, label)
	}
	return r, nil
}

// EvaluateEKSDefaultServiceAccountUse checks CIS EKS 2.1.0 4.1.5 against
// current Pods plus all built-in workload templates. Kubernetes API admission
// defaults omitted serviceAccountName to "default".
func EvaluateEKSDefaultServiceAccountUse(ctx context.Context, client kubernetes.Interface) (Result, error) {
	r := Result{ID: "4.1.5", Title: "Dedicated workload service accounts", Target: "policies", Status: "NOT_ASSESSED", Method: "kubernetes_workload_identity", Scored: true, Remediation: "Assign dedicated service accounts to application workloads; disable unnecessary auto-mounted API tokens."}
	var uses []Evidence
	record := func(kind, ns, name, sa string) {
		if sa == "" || sa == "default" {
			uses = append(uses, Evidence{Resource: kind, Namespace: ns, Name: name, Field: "serviceAccountName", Observed: firstNonEmpty(sa, "<unset; defaults to default>"), Expected: "dedicated service account"})
		}
	}
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	for _, p := range pods.Items {
		if p.Status.Phase != corev1.PodSucceeded && p.Status.Phase != corev1.PodFailed {
			record("Pod", p.Namespace, p.Name, p.Spec.ServiceAccountName)
		}
	}
	deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	for _, p := range deployments.Items {
		record("Deployment", p.Namespace, p.Name, p.Spec.Template.Spec.ServiceAccountName)
	}
	sts, err := client.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	for _, p := range sts.Items {
		record("StatefulSet", p.Namespace, p.Name, p.Spec.Template.Spec.ServiceAccountName)
	}
	ds, err := client.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	for _, p := range ds.Items {
		record("DaemonSet", p.Namespace, p.Name, p.Spec.Template.Spec.ServiceAccountName)
	}
	jobs, err := client.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	for _, p := range jobs.Items {
		record("Job", p.Namespace, p.Name, p.Spec.Template.Spec.ServiceAccountName)
	}
	cron, err := client.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	for _, p := range cron.Items {
		record("CronJob", p.Namespace, p.Name, p.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName)
	}
	r.Evidence = uses
	if len(uses) > 0 {
		r.Status = "FAIL"
		r.Details = fmt.Sprintf("%d Pods/workload templates rely on the default service account; parent/child duplicates are retained to show deployment sources.", len(uses))
		return r, nil
	}
	r.Status = "MANUAL"
	r.Details = "No built-in workload template or active Pod uses the default service account. Audit default service account RBAC bindings, group privileges and EKS Pod Identity/IRSA associations to complete this control."
	r.Evidence = []Evidence{{Resource: "KubernetesWorkloads", Field: "defaultServiceAccountReferences", Observed: "0", Expected: "0; plus manual review of inherited permissions"}}
	return r, nil
}

// EvaluateEKSServiceAccountTokenMounts checks effective auto-mounting and explicit
// Pod token projections for EKS 4.1.6. Workload API requirements cannot be
// inferred from Pod metadata; token presence alone never proves a security FAIL.
func EvaluateEKSServiceAccountTokenMounts(ctx context.Context, client kubernetes.Interface) (Result, error) {
	r := Result{ID: "4.1.6", Title: "Restrict unnecessary service account tokens", Target: "policies", Status: "NOT_ASSESSED", Method: "kubernetes_service_account_token_inventory", Scored: true, Remediation: "Disable automatic API token mounting for workloads without Kubernetes API access, and review explicit token projections."}
	accounts, err := client.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	byName := map[string]corev1.ServiceAccount{}
	for _, sa := range accounts.Items {
		byName[sa.Namespace+"/"+sa.Name] = sa
	}
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return r, err
	}
	needsReview := 0
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodSucceeded || p.Status.Phase == corev1.PodFailed {
			continue
		}
		name := firstNonEmpty(p.Spec.ServiceAccountName, "default")
		sa, ok := byName[p.Namespace+"/"+name]
		if !ok {
			r.Details = "Referenced service account could not be retrieved: " + p.Namespace + "/" + name
			return r, nil
		}
		automount := true
		if sa.AutomountServiceAccountToken != nil {
			automount = *sa.AutomountServiceAccountToken
		}
		if p.Spec.AutomountServiceAccountToken != nil {
			automount = *p.Spec.AutomountServiceAccountToken
		}
		projected := false
		for _, vol := range p.Spec.Volumes {
			if vol.Projected != nil {
				for _, source := range vol.Projected.Sources {
					if source.ServiceAccountToken != nil {
						projected = true
					}
				}
			}
			if vol.Secret != nil && strings.HasPrefix(vol.Secret.SecretName, "default-token-") {
				projected = true
			}
		}
		if automount || projected {
			needsReview++
			r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: p.Namespace, Name: p.Name, Field: "serviceAccountToken", Observed: fmt.Sprintf("automount=%t explicitProjection=%t", automount, projected), Expected: "only mounts required for approved API or identity access"})
		}
	}
	if needsReview > 0 {
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("%d Pods have automatic or explicit token projections; verify business necessity, RBAC and AWS workload identity.", needsReview)
	}
	if needsReview == 0 {
		r.Status = "PASS"
		r.Details = "No current active Pods have automatically mounted or explicitly projected service account tokens."
		r.Evidence = []Evidence{{Resource: "KubernetesPods", Field: "tokenMountCount", Observed: "0", Expected: "0"}}
	}
	return r, nil
}

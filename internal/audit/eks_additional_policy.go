package audit

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EvaluateEKSAdditionalPolicy implements EKS 2.1.0 Kubernetes-visible controls.
// An incomplete snapshot never becomes a PASS.
func EvaluateEKSAdditionalPolicy(ctx context.Context, client kubernetes.Interface, c Check) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Scored: c.Scored, Remediation: c.Remediation}
	switch c.ID {
	case "4.4.1":
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot list Pods across all namespaces: " + err.Error()
			return r
		}
		violations := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				continue
			}
			checkEnv := func(kind string, containers []corev1.Container) {
				for _, container := range containers {
					for _, env := range container.Env {
						if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
							violations++
							r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: kind + "/" + container.Name + "/env/" + env.Name, Observed: "secretKeyRef", Expected: "mounted secret volume or approved exception"})
						}
					}
					for _, source := range container.EnvFrom {
						if source.SecretRef != nil {
							violations++
							r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: kind + "/" + container.Name + "/envFrom", Observed: "secretRef", Expected: "mounted secret volume or approved exception"})
						}
					}
				}
			}
			checkEnv("container", pod.Spec.Containers)
			checkEnv("initContainer", pod.Spec.InitContainers)
			for _, container := range pod.Spec.EphemeralContainers {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
						violations++
						r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: pod.Namespace, Name: pod.Name, Field: "ephemeral/" + container.Name + "/env/" + env.Name, Observed: "secretKeyRef", Expected: "mounted secret volume or approved exception"})
					}
				}
			}
		}
		if violations > 0 {
			r.Status = "FAIL"
			r.Details = fmt.Sprintf("%d references inject Kubernetes Secrets through environment variables in active Pods; only reference names were collected, never secret values.", violations)
		} else {
			r.Status = "PASS"
			r.Details = fmt.Sprintf("No SecretKeyRef/SecretRef environment injections in %d inventoried Pods; this evaluates observed Pods, not future admission.", len(pods.Items))
		}
		return r
	case "4.5.2":
		pods, err := client.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot inspect the default namespace: " + err.Error()
			return r
		}
		violations := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
				continue
			}
			violations++
			r.Evidence = append(r.Evidence, Evidence{Resource: "Pod", Namespace: "default", Name: pod.Name, Field: "namespace", Observed: "default", Expected: "dedicated application namespace"})
		}
		deployments, err := client.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot inspect Deployments in default: " + err.Error()
			return r
		}
		for _, obj := range deployments.Items {
			violations++
			r.Evidence = append(r.Evidence, Evidence{Resource: "Deployment", Namespace: "default", Name: obj.Name, Field: "namespace", Observed: "default", Expected: "dedicated application namespace"})
		}
		sts, err := client.AppsV1().StatefulSets("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot inspect StatefulSets in default: " + err.Error()
			return r
		}
		for _, obj := range sts.Items {
			violations++
			r.Evidence = append(r.Evidence, Evidence{Resource: "StatefulSet", Namespace: "default", Name: obj.Name, Observed: "default", Expected: "dedicated application namespace"})
		}
		ds, err := client.AppsV1().DaemonSets("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot inspect DaemonSets in default: " + err.Error()
			return r
		}
		for _, obj := range ds.Items {
			violations++
			r.Evidence = append(r.Evidence, Evidence{Resource: "DaemonSet", Namespace: "default", Name: obj.Name, Observed: "default", Expected: "dedicated application namespace"})
		}
		jobs, err := client.BatchV1().Jobs("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot inspect Jobs in default: " + err.Error()
			return r
		}
		for _, obj := range jobs.Items {
			violations++
			r.Evidence = append(r.Evidence, Evidence{Resource: "Job", Namespace: "default", Name: obj.Name, Observed: "default", Expected: "dedicated application namespace"})
		}
		cron, err := client.BatchV1().CronJobs("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.Details = "Cannot inspect CronJobs in default: " + err.Error()
			return r
		}
		for _, obj := range cron.Items {
			violations++
			r.Evidence = append(r.Evidence, Evidence{Resource: "CronJob", Namespace: "default", Name: obj.Name, Observed: "default", Expected: "dedicated application namespace"})
		}
		r.Status = "PASS"
		r.Details = "No active Pods or built-in workload templates in the default namespace."
		if violations > 0 {
			r.Status = "FAIL"
			r.Details = fmt.Sprintf("%d active Pods/workload templates were found in the default namespace; duplicates show controller and Pod sources.", violations)
		}
		return r
	case "5.2.1":
		assessed, err := EvaluateEKSDefaultServiceAccountUse(ctx, client)
		if err != nil {
			r.Details = "Could not inventory dedicated service account usage: " + err.Error()
			return r
		}
		r.Evidence = assessed.Evidence
		r.Status = assessed.Status
		if r.Status == "FAIL" {
			r.Details = "Workloads reference the default service account rather than a dedicated identity. " + assessed.Details
		} else {
			r.Status = "MANUAL"
			r.Details = "No default service account workload references found. Review dedicated service accounts against EKS Pod Identity/IRSA, RBAC and approved workload identity assignments before attestation."
		}
		return r
	case "4.3.1":
		ds, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, "aws-node", metav1.GetOptions{})
		if err != nil {
			r.Details = "Cannot read aws-node DaemonSet: " + err.Error()
			return r
		}
		foundAgent, enabled := false, false
		for _, ctr := range ds.Spec.Template.Spec.Containers {
			if ctr.Name == "aws-network-policy-agent" || strings.Contains(ctr.Image, "network-policy-agent") {
				foundAgent = true
				for _, arg := range ctr.Args {
					if arg == "--enable-network-policy=true" {
						enabled = true
					}
					if arg == "--enable-network-policy=false" {
						r.Status = "FAIL"
						r.Details = "aws-node network policy agent explicitly disabled."
					}
				}
			}
		}
		r.Evidence = append(r.Evidence, Evidence{Resource: "DaemonSet", Namespace: "kube-system", Name: "aws-node", Field: "networkPolicyAgent", Observed: fmt.Sprintf("present=%t enabledArg=%t ready=%d desired=%d", foundAgent, enabled, ds.Status.NumberReady, ds.Status.DesiredNumberScheduled), Expected: "supported version, enabled managed add-on/configmap and agent ready on every eligible node"})
		if r.Status == "FAIL" {
			return r
		}
		if !foundAgent {
			r.Status = "MANUAL"
			r.Details = "No network policy agent detected in aws-node. Confirm managed vpc-cni enableNetworkPolicy and supported add-on version or independently validate an alternative CNI."
		} else if ds.Status.DesiredNumberScheduled == 0 || ds.Status.NumberReady != ds.Status.DesiredNumberScheduled {
			r.Status = "MANUAL"
			r.Details = "aws-node network policy agent is not ready on every scheduled node. Validate managed add-on configuration and agent readiness."
		} else {
			r.Status = "MANUAL"
			r.Details = "aws-node has a ready network policy agent; confirm EKS managed vpc-cni enableNetworkPolicy=true, supported version and actual policy enforcement before PASS."
		}
		return r
	default:
		r.Details = "Unsupported additional EKS control ID"
		return r
	}
}

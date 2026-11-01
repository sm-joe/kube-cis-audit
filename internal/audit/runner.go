package audit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Run(o Options) (Report, error) {
	cfg, err := clusterConfig(o.Kubeconfig)
	if err != nil {
		return Report{}, fmt.Errorf("connect to cluster: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return Report{}, err
	}
	version, err := client.Discovery().ServerVersion()
	if err != nil {
		return Report{}, fmt.Errorf("detect Kubernetes version: %w", err)
	}
	minor := strings.TrimSuffix(version.Minor, "+")
	bench, err := loadBenchmark(o.ConfigDir, o.Benchmark, minor)
	if err != nil {
		return Report{}, err
	}
	report := Report{Tool: "KubeCIS", GeneratedAt: time.Now().UTC(), ClusterVersion: version.GitVersion, Benchmark: bench.ID, CompatibilityStatus: bench.CompatibilityStatus, CompatibilityNote: bench.CompatibilityNote}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for _, c := range bench.Checks {
		if o.Target != "all" && o.Target != "" && !strings.EqualFold(c.Target, o.Target) {
			continue
		}
		r := evaluate(ctx, client, c, o)
		report.Results = append(report.Results, r)
		switch r.Status {
		case "PASS":
			report.Summary.Pass++
		case "FAIL":
			report.Summary.Fail++
		case "WARN":
			report.Summary.Warn++
		case "MANUAL":
			report.Summary.Manual++
		case "NOT_ASSESSED":
			report.Summary.NotAssessed++
		default:
			report.Summary.Info++
		}
	}
	report.Summary.Total = len(report.Results)
	return report, nil
}

func clusterConfig(path string) (*rest.Config, error) {
	// An explicit --kubeconfig always wins. If it is not supplied, honour
	// KUBECONFIG (including its platform-specific list of files) through the
	// standard client-go loading rules.
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve kubeconfig path %q: %w", path, err)
		}
		if _, err := os.Stat(absPath); err != nil {
			return nil, fmt.Errorf("kubeconfig %q is not accessible: %w; pass --kubeconfig <path> or set KUBECONFIG", absPath, err)
		}
		cfg, err := clientcmd.BuildConfigFromFlags("", absPath)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %q: %w", absPath, err)
		}
		return cfg, nil
	}

	// In-cluster credentials are available to Pods through the mounted
	// service-account token and Kubernetes service environment variables.
	// Do not attempt this path on a workstation: it produces noisy client-go
	// warnings when those variables are absent.
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("load in-cluster credentials: %w", err)
		}
		return cfg, nil
	}

	// Outside a Pod, use the normal client-go loading rules. They honour
	// KUBECONFIG and otherwise resolve ~/.kube/config on Windows, macOS and
	// Linux. This avoids BuildConfigFromFlags' in-cluster fallback warning.
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)
	cfg, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("no usable Kubernetes credentials found: %w; run 'kubectl config current-context' to verify your context, or specify --kubeconfig <path> (or set KUBECONFIG). When running as a Kubernetes Job, ensure the Pod uses a ServiceAccount with the required RBAC permissions", err)
	}
	return cfg, nil
}
func loadBenchmark(dir, id, minor string) (Benchmark, error) {
	if id != "auto" && id != "" {
		return readBench(filepath.Join(dir, id, "benchmark.yaml"))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Benchmark{}, fmt.Errorf("read benchmark directory %s: %w", dir, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, er := readBench(filepath.Join(dir, e.Name(), "benchmark.yaml"))
		if er != nil {
			continue
		}
		for _, m := range b.KubernetesMinor {
			if m == minor {
				return b, nil
			}
		}
	}
	return Benchmark{}, fmt.Errorf("no benchmark mapping for Kubernetes minor %s; specify --benchmark explicitly or add a YAML mapping", minor)
}
func readBench(p string) (Benchmark, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return Benchmark{}, fmt.Errorf("load benchmark %s: %w", p, err)
	}
	var b Benchmark
	if err = yaml.Unmarshal(data, &b); err != nil {
		return b, err
	}
	if b.ID == "" || len(b.Checks) == 0 {
		return b, fmt.Errorf("benchmark %s has no id or checks", p)
	}
	return b, nil
}
func evaluate(ctx context.Context, client kubernetes.Interface, c Check, o Options) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Details: "No evaluator produced a result", Method: c.Kind, Remediation: c.Remediation, Scored: c.Scored}
	if c.Manual {
		r.Status = "MANUAL"
		r.Details = "Requires manual verification; the required configuration evidence is not exposed by the Kubernetes API."
		return r
	}
	switch c.Kind {
	case "api_policy_control":
		return evaluatePolicyControl(ctx, client, c, o, r)
	case "api_pod_non_root", "api_pod_seccomp":
		return evaluatePodSecurity(ctx, client, c, o, r)
	case "api_network_policy":
		return evaluateNetworkPolicies(ctx, client, c, o, r)
	case "api_cluster_admin":
		return evaluateClusterAdmin(ctx, client, c, r)

	case "host_file_exists", "host_file_mode":
		p := filepath.Join(o.HostRoot, strings.TrimLeft(filepath.Clean(c.Path), "/"))
		st, err := os.Stat(p)
		if err != nil {
			r.Status = "NOT_ASSESSED"
			r.Details = fmt.Sprintf("Host evidence unavailable at %s: %v. This is an access limitation, not a control failure.", p, err)
			r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Observed: "unavailable", Expected: "readable host file"}}
			return r
		}
		if c.Kind == "host_file_exists" {
			r.Status = "PASS"
			r.Details = "Host configuration file exists: " + c.Path
			r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Observed: "present", Expected: "present"}}
			return r
		}
		// Windows does not expose POSIX permission bits through os.FileMode.
		// Do not turn its synthetic mode (often 0666) into a false Linux finding.
		if runtime.GOOS == "windows" {
			r.Status = "NOT_ASSESSED"
			r.Details = "POSIX host-file permissions cannot be reliably inspected from a Windows process; run KubeCIS on Linux or in the Kubernetes node-audit container."
			r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: "mode", Observed: "unavailable on Windows", Expected: "POSIX permission bits no broader than " + c.MaxMode}}
			return r
		}
		return evaluateHostFileMode(st.Mode(), c, r)
	default:
		r.Status = "NOT_ASSESSED"
		r.Details = "Unsupported check kind: " + c.Kind
	}
	return r
}

func evaluateHostFileMode(mode os.FileMode, c Check, r Result) Result {
	want, err := strconv.ParseUint(c.MaxMode, 8, 32)
	if err != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Invalid maxMode in benchmark YAML: " + err.Error()
		return r
	}
	got := uint64(mode.Perm())
	r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: "mode", Observed: fmt.Sprintf("%04o", got), Expected: "no broader than " + c.MaxMode}}
	if got&^want != 0 {
		r.Status = "FAIL"
		r.Details = fmt.Sprintf("%s permissions are %04o; expected no broader than %s.", c.Path, got, c.MaxMode)
	} else {
		r.Status = "PASS"
		r.Details = fmt.Sprintf("%s permissions are %04o; permission bits satisfy the configured maximum.", c.Path, got)
	}
	return r
}

func evaluatePolicyControl(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to inspect namespaces for Pod Security Admission labels: "+err.Error()
		return r
	}
	total, covered := 0, 0
	for _, ns := range nsList.Items {
		if o.Namespace != "" && ns.Name != o.Namespace {
			continue
		}
		total++
		level := strings.ToLower(ns.Labels["pod-security.kubernetes.io/enforce"])
		if level == "baseline" || level == "restricted" {
			covered++
			r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "pod-security.kubernetes.io/enforce", Observed: level, Expected: "baseline or restricted"})
		} else {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "pod-security.kubernetes.io/enforce", Observed: firstNonEmpty(level, "<unset>"), Expected: "baseline or restricted"})
		}
	}
	if total > 0 && covered == total {
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("Pod Security Admission enforce labels cover all %d inspected namespaces. Confirm the PodSecurity admission plugin is enabled in the API server and review configured exemptions; namespace labels alone do not prove effective admission enforcement.", total)
		return r
	}
	// A missing PSA label is not proof that the cluster has no policy engine.
	// Look for evidence of alternative admission mechanisms, but do not claim
	// they enforce pod security until their rules and match scopes are reviewed.
	var candidates []string
	if bindings, e := client.AdmissionregistrationV1().ValidatingAdmissionPolicyBindings().List(ctx, metav1.ListOptions{}); e == nil {
		policies, pe := client.AdmissionregistrationV1().ValidatingAdmissionPolicies().List(ctx, metav1.ListOptions{})
		policyNames := map[string]bool{}
		if pe == nil {
			for _, p := range policies.Items {
				if len(p.Spec.Validations) > 0 {
					policyNames[p.Name] = true
				}
			}
		}
		for _, b := range bindings.Items {
			if policyNames[b.Spec.PolicyName] && hasDenyAction(b.Spec.ValidationActions) {
				candidates = append(candidates, "ValidatingAdmissionPolicyBinding/"+b.Name+" (Deny)")
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Name: b.Name, Field: "spec.validationActions", Observed: "Deny; policy=" + b.Spec.PolicyName, Expected: "policy rules match pod create/update requests and intended namespaces"})
			}
		}
	}
	if hooks, e := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{}); e == nil {
		for _, h := range hooks.Items {
			name := strings.ToLower(h.Name)
			if strings.Contains(name, "kyverno") || strings.Contains(name, "gatekeeper") || strings.Contains(name, "kubewarden") {
				candidates = append(candidates, "ValidatingWebhookConfiguration/"+h.Name)
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingWebhookConfiguration", Name: h.Name, Field: "webhooks", Observed: fmt.Sprintf("%d webhook(s)", len(h.Webhooks)), Expected: "inspect webhook rules, failurePolicy and policy enforcement"})
			}
		}
	}
	if total == 0 {
		r.Status, r.Details = "NOT_ASSESSED", "No namespaces were available to evaluate policy-control coverage."
	} else if len(candidates) > 0 {
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("Pod Security Admission enforce labels cover %d of %d namespaces. Alternative admission-policy candidates were detected (%s), but their rules, match scope, failure behavior and enforcement must be reviewed manually.", covered, total, strings.Join(candidates, ", "))
	} else if covered > 0 {
		r.Status = "WARN"
		r.Details = fmt.Sprintf("Pod Security Admission enforce labels cover %d of %d namespaces. Uncovered namespaces: %d. No recognized alternative policy candidate was confirmed; inspect admission configuration and policy engines.", covered, total, total-covered)
	} else {
		r.Status = "NOT_ASSESSED"
		r.Details = fmt.Sprintf("No baseline/restricted Pod Security Admission enforce labels were found across %d namespaces. The Kubernetes API cannot prove that no other admission policy mechanism is configured; inspect API-server admission configuration and installed policy engines.", total)
	}
	return r
}

func evaluatePodSecurity(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list Pods: "+err.Error()
		return r
	}
	checkedPods, passed, failed, unknown := 0, 0, 0, 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		checkedPods++
		containers := podContainers(pod)
		for _, container := range containers {
			observed, outcome := evaluateContainerSecurity(c.Kind, pod.Spec.SecurityContext, container.SecurityContext)
			if len(r.Evidence) < 30 {
				r.Evidence = append(r.Evidence, Evidence{Resource: container.Kind, Namespace: pod.Namespace, Name: pod.Name + "/" + container.Name, Field: securityField(c.Kind), Observed: observed + " (" + outcome + ")", Expected: securityExpected(c.Kind)})
			}
			switch outcome {
			case "pass":
				passed++
			case "fail":
				failed++
			default:
				unknown++
			}
		}
	}
	if checkedPods == 0 || passed+failed+unknown == 0 {
		r.Status, r.Details = "NOT_ASSESSED", "No eligible non-terminal Pods with inspectable containers were available."
	} else if failed > 0 {
		r.Status = "FAIL"
		r.Details = fmt.Sprintf("%d of %d inspected containers have an explicit incompatible security setting across %d Pods. Unknown/inherited cases: %d. Evidence is capped at 30 entries.", failed, passed+failed+unknown, checkedPods, unknown)
	} else if unknown > 0 {
		r.Status = "WARN"
		r.Details = fmt.Sprintf("%d containers satisfy the check; %d containers have insufficient explicit Pod-spec evidence across %d Pods. Missing settings are not proof that the image runs as root or that the runtime is unconfined.", passed, unknown, checkedPods)
	} else {
		r.Status = "PASS"
		r.Details = fmt.Sprintf("All %d inspected containers across %d Pods satisfy the configured Pod-spec check.", passed, checkedPods)
	}
	return r
}

type podContainer struct {
	Name, Kind      string
	SecurityContext *corev1.SecurityContext
}

func podContainers(p corev1.Pod) []podContainer {
	out := make([]podContainer, 0, len(p.Spec.InitContainers)+len(p.Spec.Containers)+len(p.Spec.EphemeralContainers))
	for _, c := range p.Spec.InitContainers {
		out = append(out, podContainer{c.Name, "initContainer", c.SecurityContext})
	}
	for _, c := range p.Spec.Containers {
		out = append(out, podContainer{c.Name, "container", c.SecurityContext})
	}
	for _, c := range p.Spec.EphemeralContainers {
		out = append(out, podContainer{c.Name, "ephemeralContainer", c.SecurityContext})
	}
	return out
}

// outcome is pass, fail, or unknown. Pod-level values are inherited unless a
// container-level value explicitly overrides them.
func evaluateContainerSecurity(kind string, podSC *corev1.PodSecurityContext, containerSC *corev1.SecurityContext) (observed, outcome string) {
	if kind == "api_pod_non_root" {
		var runAsNonRoot *bool
		var uid *int64
		if podSC != nil {
			runAsNonRoot, uid = podSC.RunAsNonRoot, podSC.RunAsUser
		}
		if containerSC != nil {
			if containerSC.RunAsNonRoot != nil {
				runAsNonRoot = containerSC.RunAsNonRoot
			}
			if containerSC.RunAsUser != nil {
				uid = containerSC.RunAsUser
			}
		}
		if uid != nil {
			if *uid == 0 {
				return "runAsUser=0", "fail"
			}
			if *uid > 0 {
				return fmt.Sprintf("runAsUser=%d", *uid), "pass"
			}
			return fmt.Sprintf("runAsUser=%d (invalid UID)", *uid), "fail"
		}
		if runAsNonRoot != nil && !*runAsNonRoot {
			return "runAsNonRoot=false; runAsUser unset", "fail"
		}
		if runAsNonRoot != nil && *runAsNonRoot {
			return "runAsNonRoot=true", "pass"
		}
		return "no explicit non-root constraint; image user not inspected", "unknown"
	}
	var profile *corev1.SeccompProfile
	if podSC != nil {
		profile = podSC.SeccompProfile
	}
	if containerSC != nil && containerSC.SeccompProfile != nil {
		profile = containerSC.SeccompProfile
	}
	if profile == nil {
		return "seccompProfile unset", "fail"
	}
	switch profile.Type {
	case corev1.SeccompProfileTypeRuntimeDefault:
		return "seccompProfile.type=RuntimeDefault", "pass"
	case corev1.SeccompProfileTypeLocalhost:
		if profile.LocalhostProfile == nil || strings.TrimSpace(*profile.LocalhostProfile) == "" {
			return "seccompProfile.type=Localhost with empty localhostProfile", "fail"
		}
		return "seccompProfile.type=Localhost; localhostProfile=" + *profile.LocalhostProfile, "pass"
	case corev1.SeccompProfileTypeUnconfined:
		return "seccompProfile.type=Unconfined", "fail"
	default:
		return "seccompProfile.type=" + string(profile.Type), "unknown"
	}
}
func securityField(kind string) string {
	if kind == "api_pod_non_root" {
		return "securityContext.runAsNonRoot/runAsUser"
	}
	return "securityContext.seccompProfile.type"
}
func securityExpected(kind string) string {
	if kind == "api_pod_non_root" {
		return "runAsNonRoot=true or runAsUser>0; never runAsUser=0"
	}
	return "RuntimeDefault or Localhost"
}
func hasDenyAction(items []admissionv1.ValidationAction) bool {
	for _, item := range items {
		if item == admissionv1.Deny {
			return true
		}
	}
	return false
}
func firstNonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func evaluateNetworkPolicies(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	namespace := c.Namespace
	if namespace == "" {
		namespace = o.Namespace
	}
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list namespaces: "+err.Error()
		return r
	}
	policies, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list NetworkPolicies: "+err.Error()
		return r
	}
	covered := map[string]int{}
	for _, p := range policies.Items {
		covered[p.Namespace]++
	}
	missing := 0
	for _, ns := range namespaces.Items {
		if namespace != "" && ns.Name != namespace {
			continue
		}
		if covered[ns.Name] == 0 {
			missing++
			if len(r.Evidence) < 50 {
				r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "NetworkPolicy", Observed: "none", Expected: "at least one NetworkPolicy"})
			}
		} else if len(r.Evidence) < 50 {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "NetworkPolicy", Observed: fmt.Sprintf("%d policy object(s)", covered[ns.Name]), Expected: "at least one NetworkPolicy"})
		}
	}
	if len(namespaces.Items) == 0 {
		r.Status, r.Details = "NOT_ASSESSED", "No namespaces were available to inspect."
		return r
	}
	inspected := len(namespaces.Items)
	if namespace != "" {
		inspected = 0
		for _, ns := range namespaces.Items {
			if ns.Name == namespace {
				inspected++
			}
		}
	}
	if inspected == 0 {
		r.Status, r.Details = "NOT_ASSESSED", "The requested namespace does not exist."
	} else if missing > 0 {
		r.Status = "FAIL"
		r.Details = fmt.Sprintf("%d of %d inspected namespaces have no NetworkPolicy object. Object presence does not prove effective traffic isolation.", missing, inspected)
	} else {
		r.Status = "PASS"
		r.Details = fmt.Sprintf("At least one NetworkPolicy object exists in each of %d inspected namespaces. Rule selectors, ingress/egress isolation and CNI enforcement still require validation.", inspected)
	}
	return r
}

func evaluateClusterAdmin(ctx context.Context, client kubernetes.Interface, c Check, r Result) Result {
	clusterBindings, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list ClusterRoleBindings: "+err.Error()
		return r
	}
	for _, b := range clusterBindings.Items {
		if b.RoleRef.Name != "cluster-admin" {
			continue
		}
		for _, subject := range b.Subjects {
			r.Evidence = append(r.Evidence, Evidence{Resource: "ClusterRoleBinding", Name: b.Name, Field: "subjects", Observed: subjectIdentity(subject), Expected: "only approved, justified administrators"})
		}
	}
	roleBindings, err := client.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list namespaced RoleBindings: "+err.Error()
		return r
	}
	for _, b := range roleBindings.Items {
		if b.RoleRef.Name != "cluster-admin" {
			continue
		}
		for _, subject := range b.Subjects {
			r.Evidence = append(r.Evidence, Evidence{Resource: "RoleBinding", Namespace: b.Namespace, Name: b.Name, Field: "subjects", Observed: subjectIdentity(subject), Expected: "only approved, justified administrators"})
		}
	}
	if len(r.Evidence) > 0 {
		r.Status = "WARN"
		r.Details = fmt.Sprintf("%d cluster-admin subject binding(s) found across ClusterRoleBindings and RoleBindings. The API inventory cannot determine business necessity; review each binding and its owner.", len(r.Evidence))
	} else {
		r.Status = "PASS"
		r.Details = "No subjects are bound to the cluster-admin ClusterRole through ClusterRoleBindings or RoleBindings."
	}
	return r
}

func subjectIdentity(subject rbacv1.Subject) string {
	identity := subject.Kind + ":" + subject.Name
	if subject.Kind == "ServiceAccount" {
		identity += " (namespace=" + subject.Namespace + ")"
	}
	return identity
}

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
	cfg, err := clusterConfig(o.Kubeconfig, o.KubeContext)
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
	provider := DetectCloudProvider(version.GitVersion)
	if o.Benchmark == "cis-eks-2.1.0" {
		if provider != "eks" {
			return Report{}, fmt.Errorf("CIS EKS 2.1.0 profile requires an EKS Kubernetes API server; detected %s", provider)
		}
		if err := validateOfficial210Scope("eks", minor); err != nil {
			return Report{}, err
		}
	} else if o.Benchmark == "cis-2.1.0" {
		if err := validateOfficial210Scope(provider, minor); err != nil {
			return Report{}, err
		}
	}
	bench, err := loadBenchmark(o.ConfigDir, o.Benchmark, minor)
	if err != nil {
		return Report{}, err
	}
	if err := validateBenchmarkMinor(bench, minor); err != nil {
		return Report{}, err
	}
	compatibilityStatus := bench.CompatibilityStatus
	if strings.TrimSpace(compatibilityStatus) == "" {
		compatibilityStatus = "unknown"
	}
	bench = applyProviderProfile(bench, provider)
	report := Report{BuildVersion: BuildVersion, AuditMode: "native", Tool: "KubeCIS Native", GeneratedAt: time.Now().UTC(), ClusterVersion: version.GitVersion, ClusterVersionSource: "kubernetes-api", CloudProvider: provider, BenchmarkProfileCoverageNote: "coveragePercent is decisive PASS/FAIL coverage among currently scored profile checks only; it is NOT the percentage of official CIS EKS controls assessed and is NOT a CIS compliance score.", Benchmark: bench.ID, CompatibilityStatus: compatibilityStatus, CompatibilityNote: benchmarkCompatibilityNote(bench, version.GitVersion, provider)}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var workloadSummaryCheck *Check
	var complianceSummaryCheck *Check
	selectedChecks := make([]Check, 0, len(bench.Checks))
	for i := range bench.Checks {
		c := bench.Checks[i]
		if o.Target != "all" && o.Target != "" && !strings.EqualFold(c.Target, o.Target) {
			continue
		}
		selectedChecks = append(selectedChecks, c)
		if c.Kind == "report_compliance_summary" {
			checkCopy := c
			complianceSummaryCheck = &checkCopy
			continue
		}
		if c.Kind == "report_workload_security_summary" {
			checkCopy := c
			workloadSummaryCheck = &checkCopy
			continue
		}
		// An EKS API scan must not inspect the scanning machine's filesystem
		// when assessing AWS-managed control-plane or EC2 worker-host controls.
		var r Result
		if provider == "eks" && strings.HasPrefix(c.Kind, "host_") && (c.Target == "master" || c.Target == "node") {
			r = Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Scored: c.Scored, Remediation: c.Remediation}
		} else {
			r = evaluate(ctx, client, c, o)
		}
		r = annotateProviderAssessment(provider, c, r)
		r.FindingID = StableFindingID(r)
		report.Results = append(report.Results, r)
		incrementSummary(&report.Summary, r.Status)
	}
	if workloadSummaryCheck != nil {
		r := buildWorkloadSecuritySummary(*workloadSummaryCheck, report.Results)
		r.FindingID = StableFindingID(r)
		report.Results = append(report.Results, r)
		incrementSummary(&report.Summary, r.Status)
	}
	compliance, complianceResult := buildComplianceSummary(selectedChecks, report.Results)
	report.Compliance = compliance
	if complianceSummaryCheck != nil {
		complianceResult.ID = complianceSummaryCheck.ID
		complianceResult.Title = complianceSummaryCheck.Title
		complianceResult.Remediation = complianceSummaryCheck.Remediation
		complianceResult.FindingID = StableFindingID(complianceResult)
		report.Results = append(report.Results, complianceResult)
		incrementSummary(&report.Summary, complianceResult.Status)
	}
	report.Summary.Total = len(report.Results)
	report.enforceNativeCompatibility()
	return report, nil
}

// RunNode performs a node-local, read-only assessment. It deliberately does not
// create a Kubernetes client: the node auditor only inspects the explicitly
// mounted host filesystem and has no need for Kubernetes API credentials.
func RunNode(o Options, nodeName string) (Report, error) {
	if strings.TrimSpace(nodeName) == "" {
		return Report{}, fmt.Errorf("node name is required; set NODE_NAME or pass --node-name")
	}
	hostRoot := ""
	if strings.TrimSpace(o.MetadataFile) == "" {
		if strings.TrimSpace(o.HostRoot) == "" {
			return Report{}, fmt.Errorf("host root is required when metadata-file is not configured")
		}
		var err error
		hostRoot, err = filepath.Abs(o.HostRoot)
		if err != nil {
			return Report{}, fmt.Errorf("resolve host root: %w", err)
		}
	}
	bench, err := readSelectedBenchmark(o.ConfigDir, o.Benchmark)
	if err != nil {
		return Report{}, err
	}
	if o.Provider != "" && o.Provider != "eks" {
		return Report{}, fmt.Errorf("unsupported explicit node provider %q (supported: eks)", o.Provider)
	}
	bench = applyProviderProfile(bench, o.Provider)
	target := strings.TrimSpace(o.Target)
	if target == "" {
		target = "node"
	}
	if !strings.EqualFold(target, "node") && !strings.EqualFold(target, "master") {
		return Report{}, fmt.Errorf("unsupported host audit target %q (use node or master)", target)
	}
	report := Report{
		BuildVersion:         BuildVersion,
		AuditMode:            strings.ToLower(target) + "-local",
		CloudProvider:        o.Provider,
		NodeName:             nodeName,
		Tool:                 "KubeCIS Native",
		GeneratedAt:          time.Now().UTC(),
		ClusterVersion:       "not queried (node-local mode)",
		ClusterVersionSource: "node-local-no-api",
		Benchmark:            bench.ID,
		CompatibilityStatus:  bench.CompatibilityStatus,
		CompatibilityNote:    bench.CompatibilityNote,
	}
	var metadata NodeMetadataSnapshot
	var metadataErr error
	if strings.TrimSpace(o.MetadataFile) != "" {
		metadata, metadataErr = readNodeMetadataSnapshot(o.MetadataFile, nodeName, o.Benchmark, target)
	}
	localChecks := make([]Check, 0, len(bench.Checks))
	for _, c := range bench.Checks {
		// Node-local mode intentionally runs only checks explicitly targeted at
		// worker/control-plane node configuration and backed by host file data.
		// Manual and API-backed checks are outside this report's evidence scope.
		if !strings.EqualFold(c.Target, target) {
			continue
		}
		if c.Kind != "host_file_mode" && c.Kind != "host_file_exists" && c.Kind != "host_kubelet_config_field" && c.Kind != "cis_not_implemented" {
			continue
		}
		localChecks = append(localChecks, c)
		var r Result
		if c.Kind == "cis_not_implemented" {
			r = Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Scored: false, Details: "Official CIS EKS control pending native collector and validated evidence; no PASS/FAIL inferred.", Remediation: c.Remediation}
		} else if strings.TrimSpace(o.MetadataFile) != "" {
			if metadataErr != nil {
				r = Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Remediation: c.Remediation, Scored: c.Scored, Details: "Node metadata snapshot unavailable: " + metadataErr.Error(), Evidence: []Evidence{{Resource: "node", Name: nodeName, Field: "metadata", Observed: "unavailable", Expected: "fresh metadata snapshot from the node metadata helper"}}}
			} else if c.Kind == "host_kubelet_config_field" {
				r = evaluateKubeletSettingFromMetadata(nodeName, c, metadata)
			} else {
				r = evaluateNodeHostCheckFromMetadata(nodeName, target, c, metadata)
			}
		} else if c.Kind == "host_kubelet_config_field" {
			r = evaluateKubeletConfigField(hostRoot, nodeName, c)
		} else {
			r = evaluateNodeHostCheck(hostRoot, nodeName, c)
		}
		r.Scope = nodeName
		r.FindingID = StableFindingID(r)
		report.Results = append(report.Results, r)
		incrementSummary(&report.Summary, r.Status)
	}
	report.Summary.Total = len(report.Results)
	if report.Summary.Total == 0 {
		return Report{}, fmt.Errorf("benchmark %s contains no supported %s-local host checks", bench.ID, target)
	}
	// Compliance is scoped to the host checks actually selected for this
	// node-local report. Do not count unrelated API/manual controls as missing
	// evidence, and do not leave a scored local result out of aggregation.
	report.Compliance, _ = buildComplianceSummary(localChecks, report.Results)
	report.enforceNativeCompatibility()
	return report, nil
}

func readSelectedBenchmark(dir, id string) (Benchmark, error) {
	if id == "" || id == "auto" {
		return Benchmark{}, fmt.Errorf("node-local mode requires an explicit --benchmark because it does not query the Kubernetes API version")
	}
	return readBench(filepath.Join(dir, id, "benchmark.yaml"))
}

func incrementSummary(s *Summary, status string) {
	switch status {
	case "PASS":
		s.Pass++
	case "FAIL":
		s.Fail++
	case "WARN":
		s.Warn++
	case "MANUAL":
		s.Manual++
	case "NOT_ASSESSED":
		s.NotAssessed++
	default:
		s.Info++
	}
}

func evaluateNodeHostCheck(hostRoot, nodeName string, c Check) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Remediation: c.Remediation, Scored: c.Scored, Evidence: []Evidence{{Resource: "node", Name: nodeName, Field: "identity", Observed: nodeName, Expected: "Kubernetes node being assessed"}}}
	clean := filepath.Clean(filepath.FromSlash(c.Path))
	if c.Path == "" || filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		r.Details = fmt.Sprintf("Invalid or unsafe host path %q in benchmark profile; path must be relative to the mounted host root.", c.Path)
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Observed: "rejected unsafe path", Expected: "relative path beneath host root"}}
		return r
	}
	fullPath := filepath.Join(hostRoot, clean)
	rel, err := filepath.Rel(hostRoot, fullPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		r.Details = fmt.Sprintf("Host path %q resolves outside the mounted host root and was rejected.", c.Path)
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Observed: "rejected path escape", Expected: "path beneath host root"}}
		return r
	}
	// Lstat avoids silently following a symlink from the mounted host into a
	// different location. Symlink targets need a separate, explicit policy.
	st, err := os.Lstat(fullPath)
	if err != nil {
		r.Details = fmt.Sprintf("Node %s: host evidence unavailable at %s: %v. This is an access limitation, not a control failure.", nodeName, c.Path, err)
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: "node", Observed: nodeName, Expected: "node-local host file evidence"}, {Resource: "host-file", Name: c.Path, Observed: "unavailable", Expected: "readable host file"}}
		return r
	}
	if st.Mode()&os.ModeSymlink != 0 {
		r.Details = fmt.Sprintf("Host path %s is a symbolic link; target inspection is intentionally not followed.", c.Path)
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: "type", Observed: "symlink", Expected: "regular file or explicitly reviewed target"}}
		return r
	}
	if !st.Mode().IsRegular() {
		r.Details = fmt.Sprintf("Host path %s is not a regular file.", c.Path)
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: "type", Observed: st.Mode().Type().String(), Expected: "regular file"}}
		return r
	}
	if c.Kind == "host_file_exists" {
		r.Status = "PASS"
		r.Details = "Host configuration file exists: " + c.Path
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Observed: "present", Expected: "present"}}
		return r
	}
	if runtime.GOOS == "windows" {
		r.Details = "POSIX host-file permissions cannot be reliably inspected from a Windows process; run node-audit on Linux."
		r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: "mode", Observed: "unavailable on Windows", Expected: "POSIX permission bits no broader than " + c.MaxMode}}
		return r
	}
	return evaluateHostFileMode(st.Mode(), c, r)
}

func clusterConfig(path, contextName string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve kubeconfig path %q: %w", path, err)
		}
		if _, err := os.Stat(absPath); err != nil {
			return nil, fmt.Errorf("kubeconfig %q is not accessible: %w; pass --kubeconfig <path> or set KUBECONFIG", absPath, err)
		}
		loadingRules = &clientcmd.ClientConfigLoadingRules{ExplicitPath: absPath}
	}

	// In-cluster credentials are available to Pods through the mounted
	// service-account token and Kubernetes service environment variables.
	// A kubeconfig context has no meaning inside this authentication mode.
	if path == "" && os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		if contextName != "" {
			return nil, fmt.Errorf("--context cannot be used with in-cluster credentials")
		}
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("load in-cluster credentials: %w", err)
		}
		return cfg, nil
	}

	overrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
	cfg, err := clientConfig.ClientConfig()
	if err != nil {
		if path != "" {
			return nil, fmt.Errorf("load kubeconfig %q (context %q): %w", path, contextName, err)
		}
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
	case "eks_psa_admission":
		if c.ID != "4.2.1" && c.ID != "4.2.2" && c.ID != "4.2.3" && c.ID != "4.2.4" && c.ID != "4.2.5" {
			r.Details = "Unsupported EKS admission control ID"
			return r
		}
		assessed, err := EvaluateEKSPodAdmissionControl(ctx, client, c.ID)
		if err != nil {
			r.Details = "Unable to collect namespace and Pod admission evidence: " + err.Error()
			return r
		}
		return assessed
	case "eks_default_service_account":
		if c.ID != "4.1.5" {
			r.Details = "Unsupported EKS workload identity control ID"
			return r
		}
		assessed, err := EvaluateEKSDefaultServiceAccountUse(ctx, client)
		if err != nil {
			r.Details = "Unable to collect workload identity evidence: " + err.Error()
			return r
		}
		return assessed
	case "eks_token_mounts":
		if c.ID != "4.1.6" {
			r.Details = "Unsupported EKS token mounting control ID"
			return r
		}
		assessed, err := EvaluateEKSServiceAccountTokenMounts(ctx, client)
		if err != nil {
			r.Details = "Unable to collect service account token evidence: " + err.Error()
			return r
		}
		return assessed
	case "api_policy_control":
		return evaluatePolicyControl(ctx, client, c, o, r)
	case "api_pod_non_root", "api_pod_seccomp":
		return evaluatePodSecurity(ctx, client, c, o, r)
	case "api_pod_group_security_context":
		return evaluatePodGroupSecurityContext(ctx, client, c, o, r)
	case "api_network_policy":
		return evaluateNetworkPolicies(ctx, client, c, o, r)
	case "api_cluster_admin":
		return evaluateClusterAdmin(ctx, client, c, r)
	case "api_pod_privilege_security":
		return evaluatePodPrivilegeSecurity(ctx, client, c, o, r)
	case "api_pod_capabilities":
		return evaluatePodCapabilities(ctx, client, c, o, r)
	case "api_pod_read_only_rootfs":
		return evaluatePodReadOnlyRootFilesystem(ctx, client, c, o, r)
	case "api_pod_host_namespaces":
		return evaluatePodHostNamespaces(ctx, client, c, o, r)
	case "api_pod_hostpath":
		return evaluatePodHostPath(ctx, client, c, o, r)
	case "api_rbac_risk":
		return evaluateRBACRisk(ctx, client, c, o, r)
	case "api_network_security":
		return evaluateNetworkSecurity(ctx, client, c, o, r)
	case "api_control_plane_security":
		return evaluateControlPlaneSecurity(ctx, client, c, o, r)
	case "api_audit_logging":
		return evaluateAuditLogging(ctx, client, c, o, r)
	case "report_workload_security_summary":
		r.Status = "NOT_ASSESSED"
		r.Details = "Workload security summary is generated after individual benchmark checks."
		return r
	case "host_kubelet_config_field":
		return evaluateKubeletConfigField(o.HostRoot, "cluster-host", c)
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
	nsList, nsErr := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if nsErr != nil {
		r.Status = "NOT_ASSESSED"
		r.Details = "Unable to inventory namespaces for Pod Security Admission coverage: " + nsErr.Error()
		return r
	}

	totalNamespaces, psaCovered := 0, 0
	for _, ns := range nsList.Items {
		if o.Namespace != "" && ns.Name != o.Namespace {
			continue
		}
		totalNamespaces++
		level := strings.ToLower(ns.Labels["pod-security.kubernetes.io/enforce"])
		if level == "baseline" || level == "restricted" {
			psaCovered++
			r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "pod-security.kubernetes.io/enforce", Observed: level, Expected: "baseline or restricted"})
		} else {
			r.Evidence = append(r.Evidence, Evidence{Resource: "Namespace", Name: ns.Name, Field: "pod-security.kubernetes.io/enforce", Observed: firstNonEmpty(level, "<unset>"), Expected: "baseline or restricted"})
		}
	}
	if totalNamespaces == 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "No namespaces were available in the requested assessment scope."
		return r
	}

	var inventoryErrors []string
	blockingCandidates, advisoryCandidates := 0, 0

	policies, policyErr := client.AdmissionregistrationV1().ValidatingAdmissionPolicies().List(ctx, metav1.ListOptions{})
	bindings, bindingErr := client.AdmissionregistrationV1().ValidatingAdmissionPolicyBindings().List(ctx, metav1.ListOptions{})
	if policyErr != nil {
		inventoryErrors = append(inventoryErrors, "ValidatingAdmissionPolicies: "+policyErr.Error())
		r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicy", Field: "inventory", Observed: "list failed: " + policyErr.Error(), Expected: "successful cluster-wide list"})
	} else {
		r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicy", Field: "inventory", Observed: fmt.Sprintf("list succeeded; count=%d", len(policies.Items)), Expected: "inventory result recorded, including zero objects"})
	}
	if bindingErr != nil {
		inventoryErrors = append(inventoryErrors, "ValidatingAdmissionPolicyBindings: "+bindingErr.Error())
		r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Field: "inventory", Observed: "list failed: " + bindingErr.Error(), Expected: "successful cluster-wide list"})
	} else {
		r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Field: "inventory", Observed: fmt.Sprintf("list succeeded; count=%d", len(bindings.Items)), Expected: "inventory result recorded, including zero objects"})
	}

	if policyErr == nil && bindingErr == nil {
		policyByName := make(map[string]admissionv1.ValidatingAdmissionPolicy, len(policies.Items))
		for _, policy := range policies.Items {
			policyByName[policy.Name] = policy
		}
		for _, binding := range bindings.Items {
			policy, found := policyByName[binding.Spec.PolicyName]
			if !found {
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Name: binding.Name, Field: "spec.policyName", Observed: "unresolved reference: " + binding.Spec.PolicyName, Expected: "existing ValidatingAdmissionPolicy"})
				continue
			}
			if len(policy.Spec.Validations) == 0 || !matchResourcesCanMatchPods(policy.Spec.MatchConstraints, false) {
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Name: binding.Name, Field: "policyMatch", Observed: "policy has no Pod CREATE/UPDATE validation rule", Expected: "validation rule matching core/v1 Pods"})
				continue
			}
			if binding.Spec.MatchResources != nil && !matchResourcesCanMatchPods(binding.Spec.MatchResources, true) {
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Name: binding.Name, Field: "bindingMatch", Observed: "binding resource rules do not match Pod CREATE/UPDATE", Expected: "binding scope intersecting Pod validation rule"})
				continue
			}

			selectors := policySelectorSummary(policy.Spec.MatchConstraints)
			if len(policy.Spec.MatchConditions) > 0 {
				selectors += fmt.Sprintf("; policyMatchConditions=%d", len(policy.Spec.MatchConditions))
			}
			if binding.Spec.MatchResources != nil {
				selectors += "; binding " + policySelectorSummary(binding.Spec.MatchResources)
			}
			if hasDenyAction(binding.Spec.ValidationActions) {
				blockingCandidates++
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Name: binding.Name, Field: "spec.validationActions", Observed: "Deny; policy=" + policy.Name + "; " + selectors + "; policyFailurePolicy=" + failurePolicyValue(policy.Spec.FailurePolicy), Expected: "review CEL expressions, selectors, exclusions, parameter resolution and effective enforcement"})
			} else {
				advisoryCandidates++
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingAdmissionPolicyBinding", Name: binding.Name, Field: "spec.validationActions", Observed: fmt.Sprintf("%v; policy=%s; %s", binding.Spec.ValidationActions, policy.Name, selectors), Expected: "Deny action for blocking validation"})
			}
		}
	}

	webhooks, webhookErr := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if webhookErr != nil {
		inventoryErrors = append(inventoryErrors, "ValidatingWebhookConfigurations: "+webhookErr.Error())
		r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingWebhookConfiguration", Field: "inventory", Observed: "list failed: " + webhookErr.Error(), Expected: "successful cluster-wide list"})
	} else {
		r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingWebhookConfiguration", Field: "inventory", Observed: fmt.Sprintf("list succeeded; count=%d", len(webhooks.Items)), Expected: "inventory result recorded, including zero objects"})
		for _, configuration := range webhooks.Items {
			for _, webhook := range configuration.Webhooks {
				matchesPods := false
				for _, rule := range webhook.Rules {
					if ruleMatchesPodCreateUpdate(rule) {
						matchesPods = true
						break
					}
				}
				if !matchesPods {
					continue
				}
				blockingCandidates++
				selectorInfo := "namespaceSelector=" + labelSelectorState(webhook.NamespaceSelector) + "; objectSelector=" + labelSelectorState(webhook.ObjectSelector)
				if len(webhook.MatchConditions) > 0 {
					selectorInfo += fmt.Sprintf("; matchConditions=%d", len(webhook.MatchConditions))
				}
				r.Evidence = append(r.Evidence, Evidence{Resource: "ValidatingWebhook", Name: configuration.Name + "/" + webhook.Name, Field: "rules/failurePolicy", Observed: fmt.Sprintf("matches Pod CREATE/UPDATE; rules=%d; failurePolicy=%s; %s", len(webhook.Rules), failurePolicyValue(webhook.FailurePolicy), selectorInfo), Expected: "review webhook implementation, selectors, endpoint health and policy enforcement"})
			}
		}
	}

	if len(inventoryErrors) > 0 {
		r.Status = "NOT_ASSESSED"
		r.Details = "Admission-control inventory is incomplete: " + strings.Join(inventoryErrors, "; ")
		if psaCovered > 0 || blockingCandidates > 0 || advisoryCandidates > 0 {
			r.Details += ". Partial evidence was collected, but it is insufficient for a complete assessment."
		}
		return r
	}

	if psaCovered == totalNamespaces {
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("Pod Security Admission enforce labels cover all %d inspected namespaces. Confirm the PodSecurity admission plugin is active and review API-server exemptions; also review discovered admission policies and webhook behavior.", totalNamespaces)
	} else if blockingCandidates > 0 {
		r.Status = "MANUAL"
		r.Details = fmt.Sprintf("Found %d admission mechanism(s) configured to evaluate Pod CREATE/UPDATE requests with potentially blocking behavior. Namespace PSA labels cover %d of %d namespaces. Review CEL expressions or webhook logic, selectors, exclusions, endpoint health and exemptions; configuration inventory alone cannot prove effective enforcement.", blockingCandidates, psaCovered, totalNamespaces)
	} else if advisoryCandidates > 0 {
		r.Status = "WARN"
		r.Details = fmt.Sprintf("Found %d ValidatingAdmissionPolicy binding(s) matching Pod CREATE/UPDATE with only non-blocking validation actions. Namespace PSA labels cover %d of %d namespaces. Warn/Audit actions do not reject requests when a validation evaluates false.", advisoryCandidates, psaCovered, totalNamespaces)
	} else if psaCovered > 0 {
		r.Status = "WARN"
		r.Details = fmt.Sprintf("Pod Security Admission enforce labels cover %d of %d inspected namespaces. Uncovered namespaces require review; no Pod-matching blocking admission policy or validating webhook was identified.", psaCovered, totalNamespaces)
	} else {
		r.Status = "NOT_ASSESSED"
		r.Details = fmt.Sprintf("No Pod Security Admission enforce labels or Pod-matching ValidatingAdmissionPolicy/validating webhook candidates were identified across %d namespaces. The Kubernetes API inventory cannot prove that no external or provider-managed admission control is active.", totalNamespaces)
	}
	return r
}

func evaluatePodSecurity(ctx context.Context, client kubernetes.Interface, c Check, o Options, r Result) Result {
	pods, err := client.CoreV1().Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		r.Status, r.Details = "NOT_ASSESSED", "Unable to list Pods: "+err.Error()
		return r
	}
	checkedPods, passed, failed, missing, unknown := 0, 0, 0, 0, 0
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
			case "missing":
				missing++
			default:
				unknown++
			}
		}
	}
	inspected := passed + failed + missing + unknown
	if checkedPods == 0 || inspected == 0 {
		r.Status, r.Details = "NOT_ASSESSED", "No eligible non-terminal Pods with inspectable containers were available."
		return r
	}
	if failed > 0 || (c.Kind == "api_pod_seccomp" && missing > 0) {
		r.Status = "FAIL"
		if c.Kind == "api_pod_seccomp" {
			r.Details = fmt.Sprintf("%d containers have an explicitly incompatible seccomp setting and %d lack an explicit acceptable seccomp profile across %d Pods; %d containers satisfy the check and %d have unknown settings. The benchmark check evaluates Pod-spec configuration, not the effective runtime profile.", failed, missing, checkedPods, passed, unknown)
		} else {
			r.Details = fmt.Sprintf("%d of %d inspected containers have an explicitly incompatible non-root setting across %d Pods; %d containers satisfy the check and %d have insufficient Pod-spec evidence. Missing settings are not proof that the image runs as root.", failed, inspected, checkedPods, passed, unknown+missing)
		}
	} else if unknown > 0 || missing > 0 {
		r.Status = "WARN"
		r.Details = fmt.Sprintf("%d containers satisfy the check; %d containers have insufficient or unsupported Pod-spec evidence across %d Pods. Missing non-root settings are not proof that the image runs as root.", passed, unknown+missing, checkedPods)
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

// effectiveSecurityContext contains the Pod/container security-context values
// needed by workload security assessments. Nil means the field was not set in
// the Pod specification; this resolver does not infer image or runtime defaults.
type effectiveSecurityContext struct {
	RunAsNonRoot *bool
	RunAsUser    *int64
	RunAsGroup   *int64
	FSGroup      *int64
}

// resolveEffectiveSecurityContext applies Kubernetes Pod-to-container
// inheritance independently for fields supported by both scopes. A
// container-level value overrides only the matching Pod-level field.
// FSGroup is Pod-scoped and is therefore copied only from PodSecurityContext.
func resolveEffectiveSecurityContext(podSC *corev1.PodSecurityContext, containerSC *corev1.SecurityContext) effectiveSecurityContext {
	effective := effectiveSecurityContext{}
	if podSC != nil {
		effective.RunAsNonRoot = podSC.RunAsNonRoot
		effective.RunAsUser = podSC.RunAsUser
		effective.RunAsGroup = podSC.RunAsGroup
		effective.FSGroup = podSC.FSGroup
	}
	if containerSC != nil {
		if containerSC.RunAsNonRoot != nil {
			effective.RunAsNonRoot = containerSC.RunAsNonRoot
		}
		if containerSC.RunAsUser != nil {
			effective.RunAsUser = containerSC.RunAsUser
		}
		if containerSC.RunAsGroup != nil {
			effective.RunAsGroup = containerSC.RunAsGroup
		}
	}
	return effective
}

// outcome is pass, fail, or unknown. Pod-level values are inherited unless a
// container-level value explicitly overrides them.
func evaluateContainerSecurity(kind string, podSC *corev1.PodSecurityContext, containerSC *corev1.SecurityContext) (observed, outcome string) {
	if kind == "api_pod_non_root" {
		effective := resolveEffectiveSecurityContext(podSC, containerSC)
		runAsNonRoot, uid := effective.RunAsNonRoot, effective.RunAsUser
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
		// This check requires an explicit profile in the Pod specification.
		// Missing is reported separately from an explicitly incompatible profile.
		return "seccompProfile unset", "missing"
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

// enforceNativeCompatibility prevents incomplete or provisional native
// assessments from being presented as verified CIS compliance scores.
func (r *Report) enforceNativeCompatibility() {
	status := strings.ToLower(strings.TrimSpace(r.CompatibilityStatus))

	if status == "" {
		status = "unknown"
		r.CompatibilityStatus = status
	}

	c := &r.Compliance

	// A verified profile is still insufficient when the report does not
	// contain complete scored evidence.
	complete := c.ScoredTotal > 0 &&
		c.ScoredPass+c.ScoredFail == c.ScoredTotal &&
		c.ScoredWarn == 0 &&
		c.ScoredManual == 0 &&
		c.ScoredNotAssessed == 0 &&
		c.ScoredOther == 0

	// Node-local reports assess only a subset of cluster controls.
	// They must never claim cluster-wide CIS compliance.
	clusterWide := r.AuditMode == "native"

	if status == "verified" && complete && clusterWide {
		// Retain the compliance summary calculated by
		// buildComplianceSummary.
		return
	}

	// Fail closed: a percentage is not a verified compliance score
	// without a verified, complete assessment.
	c.CompliancePercent = nil

	reason := "withheld: KubeCIS Native assessment is not a verified complete CIS benchmark"

	if status != "verified" {
		reason += "; benchmark compatibility is " + status
	} else if !clusterWide {
		reason += "; node-local assessment is not cluster-wide"
	} else {
		reason += "; scored assessment evidence is incomplete"
	}

	c.ScoreStatus = reason
}

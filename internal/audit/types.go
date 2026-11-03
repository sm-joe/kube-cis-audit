package audit

import "time"

// BuildVersion is embedded in every report so operators can verify which
// auditor implementation actually produced the results.
var BuildVersion = "3.0.0-dev"

type Check struct {
	ID          string `yaml:"id" json:"id"`
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description" json:"description"`
	Target      string `yaml:"target" json:"target"`
	Kind        string `yaml:"kind" json:"kind"`
	Resource    string `yaml:"resource,omitempty" json:"resource,omitempty"`
	Namespace   string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	MaxMode     string `yaml:"maxMode,omitempty" json:"maxMode,omitempty"`
	Field       string `yaml:"field,omitempty" json:"field,omitempty"`
	Expected    string `yaml:"expected,omitempty" json:"expected,omitempty"`
	Remediation string `yaml:"remediation,omitempty" json:"remediation,omitempty"`
	Scored      bool   `yaml:"scored" json:"scored"`
	Manual      bool   `yaml:"manual,omitempty" json:"manual,omitempty"`
}

type Benchmark struct {
	ID                  string   `yaml:"id"`
	Name                string   `yaml:"name"`
	KubernetesMinor     []string `yaml:"kubernetesMinor"`
	CompatibilityStatus string   `yaml:"compatibilityStatus,omitempty"`
	CompatibilityNote   string   `yaml:"compatibilityNote,omitempty"`
	Checks              []Check  `yaml:"checks"`
}

type Evidence struct {
	Resource  string `json:"resource,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Field     string `json:"field,omitempty"`
	Observed  string `json:"observed,omitempty"`
	Expected  string `json:"expected,omitempty"`
}

type Result struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Target      string     `json:"target"`
	Status      string     `json:"status"`
	Details     string     `json:"details"`
	Method      string     `json:"method,omitempty"`
	Evidence    []Evidence `json:"evidence,omitempty"`
	Remediation string     `json:"remediation,omitempty"`
	Scored      bool       `json:"scored"`
}

type Summary struct {
	Pass        int `json:"pass"`
	Fail        int `json:"fail"`
	Warn        int `json:"warn"`
	Info        int `json:"info"`
	Manual      int `json:"manual"`
	NotAssessed int `json:"notAssessed"`
	Total       int `json:"total"`
}

type Report struct {
	BuildVersion        string            `json:"buildVersion"`
	AuditMode           string            `json:"auditMode,omitempty"`
	NodeName            string            `json:"nodeName,omitempty"`
	Tool                string            `json:"tool"`
	GeneratedAt         time.Time         `json:"generatedAt"`
	ClusterVersion      string            `json:"clusterVersion"`
	Benchmark           string            `json:"benchmark"`
	CompatibilityStatus string            `json:"benchmarkCompatibilityStatus,omitempty"`
	CompatibilityNote   string            `json:"benchmarkCompatibilityNote,omitempty"`
	Results             []Result          `json:"results"`
	Summary             Summary           `json:"summary"`
	Compliance          ComplianceSummary `json:"compliance"`
}

type Options struct {
	Kubeconfig, KubeContext, ConfigDir, Benchmark, Target, HostRoot, Namespace string
	MetadataFile                                                               string
}

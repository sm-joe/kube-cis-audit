package worker

import "time"

type Status string

const (
	StatusPass        Status = "PASS"
	StatusFail        Status = "FAIL"
	StatusNotAssessed Status = "NOT_ASSESSED"
)

type Component string

const (
	ComponentKubelet   Component = "kubelet"
	ComponentKubeProxy Component = "kube-proxy"
)

type SourceKind string

const (
	SourceFlag   SourceKind = "flag"
	SourceConfig SourceKind = "config"
)

// Evidence is a point-in-time snapshot produced by a trusted, read-only collector.
// SourceComplete must be true only when the complete relevant source was observed.
type Evidence struct {
	NodeName     string                          `json:"node_name"`
	Profile      string                          `json:"profile"`
	CollectedAt  time.Time                       `json:"collected_at"`
	Complete     bool                            `json:"complete"`
	Components   map[Component]ComponentEvidence `json:"components"`
	SourceErrors []string                        `json:"source_errors,omitempty"`
}

type ComponentEvidence struct {
	Flags          map[string][]string `json:"flags,omitempty"`
	ConfigValues   map[string]string   `json:"config_values,omitempty"`
	SourceComplete map[SourceKind]bool `json:"source_complete,omitempty"`
}

type Mapping struct {
	ControlID   string
	Title       string
	Component   Component
	Source      SourceKind
	Key         string
	Expected    string
	Profiles    []string
	Remediation string
}

type Result struct {
	ControlID      string    `json:"control_id"`
	Title          string    `json:"title"`
	NodeName       string    `json:"node_name,omitempty"`
	Component      Component `json:"component"`
	Status         Status    `json:"status"`
	Expected       string    `json:"expected,omitempty"`
	Observed       []string  `json:"observed,omitempty"`
	EvidenceSource string    `json:"evidence_source,omitempty"`
	Reason         string    `json:"reason"`
	Remediation    string    `json:"remediation,omitempty"`
	CollectedAt    time.Time `json:"collected_at,omitempty"`
}

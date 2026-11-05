package etcd

import "time"

type Status string

const (
	StatusPass        Status = "PASS"
	StatusFail        Status = "FAIL"
	StatusNotAssessed Status = "NOT_ASSESSED"
)

const ComponentEtcd = "etcd"

type Evidence struct {
	Component    string              `json:"component"`
	Profile      string              `json:"profile"`
	CollectedAt  time.Time           `json:"collected_at"`
	Complete     bool                `json:"complete"`
	Flags        map[string][]string `json:"flags,omitempty"`
	ConfigValues map[string]string   `json:"config_values,omitempty"`
	// SourceComplete identifies sources for which absence can be treated as observed absence.
	// Example: map[string]bool{"flags": true, "config": true}.
	SourceComplete map[string]bool `json:"source_complete,omitempty"`
	SourceErrors   []string        `json:"source_errors,omitempty"`
}

type SourceKind string

const (
	SourceFlag   SourceKind = "flag"
	SourceConfig SourceKind = "config"
)

type Mapping struct {
	ControlID   string
	Title       string
	Source      SourceKind
	Key         string
	Expected    string
	Profiles    []string
	Remediation string
	Sensitive   bool // sensitive evidence is deliberately never evaluated or emitted
}

type Result struct {
	ControlID      string    `json:"control_id"`
	Title          string    `json:"title"`
	Status         Status    `json:"status"`
	Expected       string    `json:"expected,omitempty"`
	Observed       []string  `json:"observed,omitempty"`
	EvidenceSource string    `json:"evidence_source,omitempty"`
	Reason         string    `json:"reason"`
	Remediation    string    `json:"remediation,omitempty"`
	CollectedAt    time.Time `json:"collected_at,omitempty"`
}

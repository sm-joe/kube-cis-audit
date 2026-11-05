package controlplane

import "time"

type Status string

const (
	StatusPass        Status = "PASS"
	StatusFail        Status = "FAIL"
	StatusNotAssessed Status = "NOT_ASSESSED"
)

type Component string

const (
	ComponentAPIServer         Component = "kube-apiserver"
	ComponentControllerManager Component = "kube-controller-manager"
	ComponentScheduler         Component = "kube-scheduler"
)

type SourceKind string

const (
	SourceFlag   SourceKind = "flag"
	SourceConfig SourceKind = "config"
)

// Evidence is a point-in-time, read-only snapshot supplied by a trusted collector.
// Complete must be true only when the collector successfully observed all declared sources.
type Evidence struct {
	Component    Component           `json:"component"`
	Profile      string              `json:"profile"`
	CollectedAt  time.Time           `json:"collected_at"`
	Complete     bool                `json:"complete"`
	Flags        map[string][]string `json:"flags,omitempty"`
	ConfigValues map[string]string   `json:"config_values,omitempty"`
	SourceErrors []string            `json:"source_errors,omitempty"`
}

// Mapping is an explicit, static mapping from one benchmark requirement to a typed evidence check.
// Exactly one of ExpectedFlag or ExpectedConfigValue must be set.
type Mapping struct {
	ControlID     string
	Title         string
	Component     Component
	Source        SourceKind
	Key           string
	ExpectedValue string
	AllowMultiple bool
	Remediation   string
	Profiles      []string // empty means profile-independent; exact match otherwise
}

// Result contains a stable, explainable assessment outcome.
type Result struct {
	ControlID      string    `json:"control_id"`
	Title          string    `json:"title"`
	Component      Component `json:"component"`
	Status         Status    `json:"status"`
	Expected       string    `json:"expected,omitempty"`
	Observed       []string  `json:"observed,omitempty"`
	EvidenceSource string    `json:"evidence_source,omitempty"`
	Reason         string    `json:"reason"`
	Remediation    string    `json:"remediation,omitempty"`
	CollectedAt    time.Time `json:"collected_at,omitempty"`
}

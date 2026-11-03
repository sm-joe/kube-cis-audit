// Package adapter translates normalized upstream test metadata into explicitly
// allowlisted, typed evaluator requests. It never executes upstream audit text.
package adapter

import "context"

type Status string

const (
	StatusNotAssessed Status = "NOT_ASSESSED"
	StatusPass        Status = "PASS"
	StatusFail        Status = "FAIL"
	StatusWarn        Status = "WARN"
	StatusInfo        Status = "INFO"
	StatusManual      Status = "MANUAL"
)

// Source describes the evidence locator from an upstream test. Values are data,
// never commands to execute.
type Source struct {
	Kind  string `json:"kind"` // flag, path, env
	Value string `json:"value"`
}

type Comparator struct {
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
	Values   []any  `json:"values,omitempty"`
}

type TestDescriptor struct {
	ControlID string     `json:"control_id"`
	Target    string     `json:"target"`
	Source    Source     `json:"source"`
	Compare   Comparator `json:"compare"`
	BinOp     string     `json:"bin_op,omitempty"`
	Manual    bool       `json:"manual,omitempty"`
}

// Evidence is produced by a trusted collector. Adapter/evaluators must not
// retrieve arbitrary paths or execute arbitrary commands from Source.Value.
type Evidence struct {
	Kind   string `json:"kind"`
	Key    string `json:"key"`
	Value  any    `json:"value,omitempty"`
	Found  bool   `json:"found"`
	Detail string `json:"detail,omitempty"`
}

type Evaluation struct {
	Status  Status `json:"status"`
	Reason  string `json:"reason"`
	Details string `json:"details,omitempty"`
}

type EvaluatorRequest struct {
	ControlID string
	Target    string
	Evidence  Evidence
	Compare   Comparator
	BinOp     string
}

type Evaluator interface {
	Evaluate(context.Context, EvaluatorRequest) Evaluation
}

type Mapping struct {
	ControlID       string   `json:"control_id"`
	Target          string   `json:"target"`
	SourceKind      string   `json:"source_kind"`
	SourceValue     string   `json:"source_value"`
	ComparatorOps   []string `json:"comparator_ops"`
	EvaluatorID     string   `json:"evaluator_id"`
	EvidenceKind    string   `json:"evidence_kind"`
	AllowEvaluation bool     `json:"allow_evaluation"`
}

type AdapterResult struct {
	ControlID   string            `json:"control_id"`
	Status      Status            `json:"status"`
	EvaluatorID string            `json:"evaluator_id,omitempty"`
	Request     *EvaluatorRequest `json:"request,omitempty"`
	Evaluation  *Evaluation       `json:"evaluation,omitempty"`
	Reason      string            `json:"reason"`
}

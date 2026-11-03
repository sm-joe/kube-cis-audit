package adapter

import (
	"context"
	"fmt"
	"strings"
)

type Registry struct{ evaluators map[string]Evaluator }

func NewRegistry() *Registry { return &Registry{evaluators: map[string]Evaluator{}} }
func (r *Registry) Register(id string, e Evaluator) error {
	id = strings.TrimSpace(id)
	if id == "" || e == nil {
		return fmt.Errorf("evaluator id and implementation are required")
	}
	if _, exists := r.evaluators[id]; exists {
		return fmt.Errorf("evaluator %q already registered", id)
	}
	r.evaluators[id] = e
	return nil
}

type Adapter struct {
	mappings []Mapping
	registry *Registry
}

func New(mappings []Mapping, registry *Registry) (*Adapter, error) {
	if registry == nil {
		return nil, fmt.Errorf("evaluator registry is required")
	}
	seen := map[string]bool{}
	for i, m := range mappings {
		if strings.TrimSpace(m.ControlID) == "" || strings.TrimSpace(m.Target) == "" || strings.TrimSpace(m.SourceKind) == "" || strings.TrimSpace(m.SourceValue) == "" || strings.TrimSpace(m.EvaluatorID) == "" || strings.TrimSpace(m.EvidenceKind) == "" {
			return nil, fmt.Errorf("mapping[%d]: required mapping field is empty", i)
		}
		key := mappingKey(m.ControlID, m.Target, m.SourceKind, m.SourceValue)
		if seen[key] {
			return nil, fmt.Errorf("mapping[%d]: duplicate mapping", i)
		}
		seen[key] = true
		if m.AllowEvaluation {
			if _, ok := registry.evaluators[m.EvaluatorID]; !ok {
				return nil, fmt.Errorf("mapping[%d]: evaluator %q is not registered", i, m.EvaluatorID)
			}
		}
	}
	return &Adapter{mappings: append([]Mapping(nil), mappings...), registry: registry}, nil
}

func mappingKey(id, target, kind, value string) string {
	return strings.Join([]string{id, target, kind, value}, "\x00")
}

// Prepare resolves an exact allowlisted mapping. Unknown, manual, ambiguous,
// unsupported comparator, or unapproved definitions remain NOT_ASSESSED.
func (a *Adapter) Prepare(t TestDescriptor) AdapterResult {
	base := AdapterResult{ControlID: t.ControlID, Status: StatusNotAssessed, Reason: "no explicitly approved evaluator mapping"}
	if t.Manual {
		base.Status = StatusManual
		base.Reason = "upstream marks this assessment as manual"
		return base
	}
	var match *Mapping
	for i := range a.mappings {
		m := &a.mappings[i]
		if mappingKey(m.ControlID, m.Target, m.SourceKind, m.SourceValue) == mappingKey(t.ControlID, t.Target, t.Source.Kind, t.Source.Value) {
			if match != nil {
				base.Reason = "multiple mappings match; refusing ambiguous evaluation"
				return base
			}
			match = m
		}
	}
	if match == nil {
		return base
	}
	if !match.AllowEvaluation {
		base.Reason = "mapping exists but evaluation is not approved"
		return base
	}
	if !contains(match.ComparatorOps, t.Compare.Operator) {
		base.Reason = "comparator is not allowlisted for this mapping"
		return base
	}
	if _, ok := a.registry.evaluators[match.EvaluatorID]; !ok {
		base.Reason = "mapped evaluator is not registered"
		return base
	}
	req := &EvaluatorRequest{ControlID: t.ControlID, Target: t.Target, Evidence: Evidence{Kind: match.EvidenceKind, Key: t.Source.Value}, Compare: t.Compare, BinOp: t.BinOp}
	base.EvaluatorID = match.EvaluatorID
	base.Request = req
	base.Reason = "approved typed evaluator request prepared"
	return base
}

// Evaluate only consumes evidence supplied by a trusted collector. This method
// deliberately has no filesystem, process, network, or Kubernetes client access.
func (a *Adapter) Evaluate(ctx context.Context, t TestDescriptor, evidence Evidence) AdapterResult {
	p := a.Prepare(t)
	if p.Request == nil {
		return p
	}
	if evidence.Kind != p.Request.Evidence.Kind || evidence.Key != p.Request.Evidence.Key {
		p.Status = StatusNotAssessed
		p.Reason = "evidence does not match the approved evidence contract"
		return p
	}
	req := *p.Request
	req.Evidence = evidence
	if !evidence.Found {
		p.Status = StatusNotAssessed
		p.Reason = "required evidence is missing"
		return p
	}
	e := a.registry.evaluators[p.EvaluatorID].Evaluate(ctx, req)
	if !validStatus(e.Status) {
		p.Status = StatusNotAssessed
		p.Reason = "evaluator returned an invalid status"
		return p
	}
	p.Status = e.Status
	p.Evaluation = &e
	p.Reason = e.Reason
	return p
}
func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
func validStatus(s Status) bool {
	switch s {
	case StatusPass, StatusFail, StatusWarn, StatusInfo, StatusNotAssessed, StatusManual:
		return true
	}
	return false
}

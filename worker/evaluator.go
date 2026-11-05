package worker

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const maxEvidenceAge = 15 * time.Minute

func Evaluate(e Evidence) []Result {
	return EvaluateWithMappings(e, DefaultMappings(), time.Now())
}

// EvaluateWithMappings permits deterministic tests and controlled profile integration.
func EvaluateWithMappings(e Evidence, mappings []Mapping, now time.Time) []Result {
	if err := ValidateMappings(mappings); err != nil {
		return []Result{{Status: StatusNotAssessed, Reason: "invalid evaluator mapping: " + err.Error()}}
	}
	if strings.TrimSpace(e.NodeName) == "" || strings.TrimSpace(e.Profile) == "" {
		return []Result{{Status: StatusNotAssessed, Reason: "node identity or profile is missing"}}
	}
	if e.CollectedAt.IsZero() || e.CollectedAt.After(now.Add(time.Minute)) || now.Sub(e.CollectedAt) > maxEvidenceAge {
		return []Result{{Status: StatusNotAssessed, Reason: "evidence timestamp is missing, in the future, or stale"}}
	}
	selected := make([]Mapping, 0, len(mappings))
	for _, m := range mappings {
		if profileApplies(m.Profiles, e.Profile) {
			if _, ok := e.Components[m.Component]; ok {
				selected = append(selected, m)
			}
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Component == selected[j].Component {
			return selected[i].ControlID < selected[j].ControlID
		}
		return selected[i].Component < selected[j].Component
	})
	out := make([]Result, 0, len(selected))
	for _, m := range selected {
		out = append(out, evaluateOne(e, e.Components[m.Component], m))
	}
	return out
}

func evaluateOne(e Evidence, component ComponentEvidence, m Mapping) Result {
	r := Result{
		ControlID: m.ControlID, Title: m.Title, NodeName: e.NodeName, Component: m.Component,
		Status: StatusNotAssessed, Expected: m.Expected, Remediation: m.Remediation, CollectedAt: e.CollectedAt,
	}
	if !e.Complete || len(e.SourceErrors) > 0 {
		r.Reason = "node evidence collection is incomplete or contains source errors"
		return r
	}

	var values []string
	var present bool
	switch m.Source {
	case SourceFlag:
		r.EvidenceSource = string(m.Component) + ".flags[" + m.Key + "]"
		values, present = component.Flags[m.Key]
	case SourceConfig:
		r.EvidenceSource = string(m.Component) + ".config_values[" + m.Key + "]"
		var value string
		value, present = component.ConfigValues[m.Key]
		if present {
			values = []string{value}
		}
	default:
		r.Reason = "unsupported evidence source"
		return r
	}
	if !present && !component.SourceComplete[m.Source] {
		r.Reason = "source was not completely observed; missing evidence cannot be assessed"
		return r
	}
	r.Observed = append([]string(nil), values...)

	matched := false
	switch m.Expected {
	case "__PRESENT__":
		matched = present && anyNonBlank(values)
	case "__ABSENT__":
		matched = !present || len(values) == 0
	default:
		matched = len(values) == 1 && strings.TrimSpace(values[0]) == m.Expected
	}
	if matched {
		r.Status = StatusPass
		r.Reason = fmt.Sprintf("complete evidence satisfies expected value %q", m.Expected)
	} else {
		r.Status = StatusFail
		r.Reason = fmt.Sprintf("complete evidence does not satisfy expected value %q", m.Expected)
	}
	return r
}

func profileApplies(profiles []string, profile string) bool {
	if len(profiles) == 0 {
		return true
	}
	for _, p := range profiles {
		if p == profile {
			return true
		}
	}
	return false
}

func anyNonBlank(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

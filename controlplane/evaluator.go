package controlplane

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	expectAbsent   = "__ABSENT__"
	expectPresent  = "__PRESENT__"
	maxEvidenceAge = 15 * time.Minute
)

func Evaluate(e Evidence) []Result {
	return EvaluateWithMappings(e, DefaultMappings(), time.Now())
}

// EvaluateWithMappings is exposed for deterministic tests and controlled profile integration.
// now is supplied by the caller; production integrations should pass time.Now().
func EvaluateWithMappings(e Evidence, mappings []Mapping, now time.Time) []Result {
	if err := ValidateMappings(mappings); err != nil {
		return []Result{{Status: StatusNotAssessed, Reason: "invalid evaluator mapping: " + err.Error()}}
	}
	selected := make([]Mapping, 0, len(mappings))
	for _, m := range mappings {
		if m.Component != e.Component || !profileApplies(m.Profiles, e.Profile) {
			continue
		}
		selected = append(selected, m)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].ControlID < selected[j].ControlID })

	out := make([]Result, 0, len(selected))
	for _, m := range selected {
		out = append(out, evaluateOne(e, m, now))
	}
	return out
}

func evaluateOne(e Evidence, m Mapping, now time.Time) Result {
	r := Result{
		ControlID: m.ControlID, Title: m.Title, Component: m.Component,
		Status: StatusNotAssessed, Expected: m.ExpectedValue,
		Remediation: m.Remediation, CollectedAt: e.CollectedAt,
	}
	if e.Component != m.Component {
		r.Reason = "evidence component does not match mapping"
		return r
	}
	if strings.TrimSpace(e.Profile) == "" {
		r.Reason = "profile identity is missing"
		return r
	}
	if e.CollectedAt.IsZero() || e.CollectedAt.After(now.Add(time.Minute)) || now.Sub(e.CollectedAt) > maxEvidenceAge {
		r.Reason = "evidence timestamp is missing, in the future, or stale"
		return r
	}
	if !e.Complete {
		r.Reason = "evidence collection is incomplete"
		return r
	}
	if len(e.SourceErrors) > 0 {
		r.Reason = "evidence contains collection errors"
		return r
	}

	var observed []string
	switch m.Source {
	case SourceFlag:
		r.EvidenceSource = "component.flags[" + m.Key + "]"
		values, ok := e.Flags[m.Key]
		if !ok {
			r.Reason = "flag evidence was not collected"
			return r
		}
		observed = append([]string(nil), values...)
	case SourceConfig:
		r.EvidenceSource = "component.config_values[" + m.Key + "]"
		value, ok := e.ConfigValues[m.Key]
		if !ok {
			r.Reason = "configuration key evidence was not collected"
			return r
		}
		observed = []string{value}
	default:
		r.Reason = "unsupported evidence source"
		return r
	}
	r.Observed = observed

	matched := false
	switch m.ExpectedValue {
	case expectAbsent:
		matched = len(observed) == 0
	case expectPresent:
		matched = len(observed) > 0 && anyNonBlank(observed)
	default:
		if m.AllowMultiple {
			for _, value := range observed {
				for _, token := range strings.Split(m.ExpectedValue, ",") {
					if strings.TrimSpace(value) == strings.TrimSpace(token) {
						matched = true
					}
				}
			}
		} else {
			matched = len(observed) == 1 && strings.TrimSpace(observed[0]) == m.ExpectedValue
		}
	}
	if matched {
		r.Status = StatusPass
		r.Reason = fmt.Sprintf("complete evidence satisfies expected value %q", m.ExpectedValue)
	} else {
		r.Status = StatusFail
		r.Reason = fmt.Sprintf("complete evidence does not satisfy expected value %q", m.ExpectedValue)
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
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

package etcd

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

func EvaluateWithMappings(e Evidence, mappings []Mapping, now time.Time) []Result {
	if err := ValidateMappings(mappings); err != nil {
		return []Result{{Status: StatusNotAssessed, Reason: "invalid evaluator mapping: " + err.Error()}}
	}
	if e.Component != ComponentEtcd || strings.TrimSpace(e.Profile) == "" {
		return []Result{{Status: StatusNotAssessed, Reason: "missing or mismatched etcd component/profile identity"}}
	}
	selected := make([]Mapping, 0, len(mappings))
	for _, m := range mappings {
		if profileApplies(m.Profiles, e.Profile) {
			selected = append(selected, m)
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].ControlID < selected[j].ControlID })
	out := make([]Result, 0, len(selected))
	for _, m := range selected {
		out = append(out, evaluateOne(e, m, now))
	}
	return out
}

func evaluateOne(e Evidence, m Mapping, now time.Time) Result {
	r := Result{ControlID: m.ControlID, Title: m.Title, Status: StatusNotAssessed,
		Expected: m.Expected, Remediation: m.Remediation, CollectedAt: e.CollectedAt}
	if e.CollectedAt.IsZero() || e.CollectedAt.After(now.Add(time.Minute)) || now.Sub(e.CollectedAt) > maxEvidenceAge {
		r.Reason = "evidence timestamp is missing, in the future, or stale"
		return r
	}
	if !e.Complete || len(e.SourceErrors) > 0 {
		r.Reason = "evidence collection is incomplete or contains source errors"
		return r
	}
	sourceName := "flags"
	var values []string
	var present bool
	switch m.Source {
	case SourceFlag:
		r.EvidenceSource = "etcd.flags[" + m.Key + "]"
		values, present = e.Flags[m.Key]
	case SourceConfig:
		sourceName = "config"
		r.EvidenceSource = "etcd.config_values[" + m.Key + "]"
		var v string
		v, present = e.ConfigValues[m.Key]
		if present {
			values = []string{v}
		}
	default:
		r.Reason = "unsupported evidence source"
		return r
	}
	// Absence is meaningful only when the collector explicitly confirms the entire source was observed.
	if !present && !e.SourceComplete[sourceName] {
		r.Reason = "source was not completely observed; absence cannot be assessed"
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

func anyNonBlank(v []string) bool {
	for _, x := range v {
		if strings.TrimSpace(x) != "" {
			return true
		}
	}
	return false
}

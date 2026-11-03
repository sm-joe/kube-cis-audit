package upstream

import (
	"fmt"
	"strings"
)

type ValidationIssue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func ValidateBenchmark(b *BenchmarkFile) []ValidationIssue {
	var out []ValidationIssue
	if b == nil {
		return []ValidationIssue{{Path: "$", Message: "benchmark is nil"}}
	}
	if len(b.Groups) == 0 {
		out = append(out, ValidationIssue{Path: "$.groups", Message: "at least one group is required"})
	}
	seen := map[string]string{}
	var visit func([]Group, string)
	visit = func(groups []Group, prefix string) {
		for i, g := range groups {
			p := fmt.Sprintf("%s.groups[%d]", prefix, i)
			if strings.TrimSpace(g.ID) == "" {
				out = append(out, ValidationIssue{p + ".id", "group id is required"})
			}
			if strings.TrimSpace(g.Text) == "" {
				out = append(out, ValidationIssue{p + ".text", "group text is required"})
			}
			for j, c := range g.Checks {
				cp := fmt.Sprintf("%s.checks[%d]", p, j)
				id := strings.TrimSpace(c.ID)
				if id == "" {
					out = append(out, ValidationIssue{cp + ".id", "control id is required"})
				} else if prev, ok := seen[id]; ok {
					out = append(out, ValidationIssue{cp + ".id", fmt.Sprintf("duplicate control id %q (first seen at %s)", id, prev)})
				} else {
					seen[id] = cp
				}
				if strings.TrimSpace(c.Text) == "" {
					out = append(out, ValidationIssue{cp + ".text", "control text is required"})
				}
				if c.Scored == nil {
					out = append(out, ValidationIssue{cp + ".scored", "scoring is unspecified; preserve as unknown, do not default"})
				}
				for k, t := range c.Tests {
					if t.Raw.Kind == 0 && t.BinOp == "" && t.Compare == nil && t.Flag == nil && t.Path == nil && t.Env == nil {
						out = append(out, ValidationIssue{fmt.Sprintf("%s.tests[%d]", cp, k), "test item has no recognized fields; inspect preserved raw YAML"})
					}
				}
			}
			visit(g.Subgroups, p)
		}
	}
	visit(b.Groups, "$")
	return out
}

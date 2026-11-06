// Command benchmark-inventory inventories kube-bench CIS control YAML without
// executing any upstream audit commands.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Control struct {
	Profile     string   `json:"profile"`
	Target      string   `json:"target"`
	SourceFile  string   `json:"source_file"`
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Scored      *bool    `json:"scored,omitempty"`
	CheckType   string   `json:"check_type,omitempty"`
	Audit       string   `json:"audit,omitempty"`
	AuditConfig string   `json:"audit_config,omitempty"`
	AuditEnv    string   `json:"audit_env,omitempty"`
	Remediation string   `json:"remediation,omitempty"`
	TestItems   int      `json:"test_items"`
	TestSources []string `json:"test_sources,omitempty"`
	Comparators []string `json:"comparators,omitempty"`
	GroupPath   []string `json:"group_path,omitempty"`
}

type Profile struct {
	Name        string    `json:"name"`
	Files       int       `json:"files"`
	Controls    []Control `json:"controls"`
	Targets     []string  `json:"targets"`
	Scored      int       `json:"scored"`
	NotScored   int       `json:"not_scored"`
	Unspecified int       `json:"scoring_unspecified"`
}

type Inventory struct {
	Source        string    `json:"source"`
	Profiles      []Profile `json:"profiles"`
	TotalFiles    int       `json:"total_files"`
	TotalControls int       `json:"total_controls"`
}

func main() {
	root := flag.String("root", "benchmarks/upstream/kube-bench/cfg", "path to kube-bench cfg directory")
	markdown := flag.String("out", "docs/UPSTREAM-BENCHMARK-INVENTORY.md", "Markdown inventory output")
	jsonPath := flag.String("json", "", "optional JSON inventory output")
	flag.Parse()

	inv, err := buildInventory(*root)
	if err != nil {
		fatal(err)
	}
	if err := writeMarkdown(*markdown, inv); err != nil {
		fatal(err)
	}
	if *jsonPath != "" {
		if err := os.MkdirAll(filepath.Dir(*jsonPath), 0o755); err != nil {
			fatal(err)
		}
		b, err := json.MarshalIndent(inv, "", "  ")
		if err != nil {
			fatal(err)
		}
		if err := os.WriteFile(*jsonPath, append(b, '\n'), 0o644); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("Inventoried %d CIS/Kubernetes distribution profiles, %d YAML files and %d controls.\n", len(inv.Profiles), inv.TotalFiles, inv.TotalControls)
	fmt.Printf("Markdown: %s\n", *markdown)
	if *jsonPath != "" {
		fmt.Printf("JSON: %s\n", *jsonPath)
	}
}

func buildInventory(root string) (Inventory, error) {
	inv := Inventory{Source: filepath.ToSlash(root)}
	entries, err := os.ReadDir(root)
	if err != nil {
		return inv, fmt.Errorf("read cfg directory %q: %w", root, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || !(strings.HasPrefix(entry.Name(), "cis-") || strings.HasPrefix(entry.Name(), "eks-")) {
			continue
		}
		p := Profile{Name: entry.Name()}
		profileDir := filepath.Join(root, entry.Name())
		err := filepath.WalkDir(profileDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !(strings.HasSuffix(d.Name(), ".yaml") || strings.HasSuffix(d.Name(), ".yml")) || d.Name() == "config.yaml" {
				return nil
			}
			controls, target, err := parseControlFile(path, root, entry.Name())
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			p.Files++
			p.Targets = append(p.Targets, target)
			for _, c := range controls {
				p.Controls = append(p.Controls, c)
				if c.Scored == nil {
					p.Unspecified++
				} else if *c.Scored {
					p.Scored++
				} else {
					p.NotScored++
				}
			}
			return nil
		})
		if err != nil {
			return inv, err
		}
		sort.Strings(p.Targets)
		p.Targets = unique(p.Targets)
		sort.Slice(p.Controls, func(i, j int) bool {
			if p.Controls[i].Target == p.Controls[j].Target {
				return p.Controls[i].ID < p.Controls[j].ID
			}
			return p.Controls[i].Target < p.Controls[j].Target
		})
		inv.TotalFiles += p.Files
		inv.TotalControls += len(p.Controls)
		inv.Profiles = append(inv.Profiles, p)
	}
	sort.Slice(inv.Profiles, func(i, j int) bool { return inv.Profiles[i].Name < inv.Profiles[j].Name })
	if len(inv.Profiles) == 0 {
		return inv, fmt.Errorf("no cis-* or eks-* profiles found under %s", root)
	}
	return inv, nil
}

func parseControlFile(path, cfgRoot, profile string) ([]Control, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return nil, "", err
	}
	if len(doc) == 0 {
		return []Control{}, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), nil
	}

	// kube-bench control files use a top-level "controls:" marker followed
	// by sibling metadata keys (id, text, type, groups). The marker itself
	// commonly decodes to nil; the control document is therefore doc, not
	// doc["controls"]. Also accept a wrapped controls mapping for compatibility.
	root := doc
	if controlsValue, exists := doc["controls"]; exists {
		if nested, ok := asMap(controlsValue); ok {
			root = nested
		} else if controlsValue != nil {
			return nil, "", fmt.Errorf("controls must be a mapping marker or mapping, got %T", controlsValue)
		}
	}
	if _, ok := root["groups"]; !ok {
		return nil, "", fmt.Errorf("control document is missing top-level groups")
	}
	target := stringValue(root["type"])
	if target == "" {
		target = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	rel, _ := filepath.Rel(cfgRoot, path)
	var out []Control
	groups, _ := root["groups"].([]interface{})
	walkGroups(groups, nil, func(groupPath []string, raw map[string]interface{}) {
		checks, _ := raw["checks"].([]interface{})
		for _, item := range checks {
			check, ok := asMap(item)
			if !ok {
				continue
			}
			id := stringValue(check["id"])
			if id == "" {
				continue
			}
			c := Control{Profile: profile, Target: target, SourceFile: filepath.ToSlash(rel), ID: id,
				Title: stringValue(check["text"]), CheckType: stringValue(check["type"]),
				Audit: stringValue(check["audit"]), AuditConfig: stringValue(check["audit_config"]),
				AuditEnv: stringValue(check["audit_env"]), Remediation: stringValue(check["remediation"]), GroupPath: append([]string(nil), groupPath...)}
			if v, exists := check["scored"]; exists {
				if b, ok := v.(bool); ok {
					c.Scored = &b
				}
			}
			if tests, ok := asMap(check["tests"]); ok {
				items, _ := tests["test_items"].([]interface{})
				c.TestItems = len(items)
				for _, ti := range items {
					if m, ok := asMap(ti); ok {
						for _, key := range []string{"flag", "path", "env"} {
							if stringValue(m[key]) != "" {
								c.TestSources = append(c.TestSources, key)
							}
						}
						if cmp, ok := asMap(m["compare"]); ok && stringValue(cmp["op"]) != "" {
							c.Comparators = append(c.Comparators, stringValue(cmp["op"]))
						}
					}
				}
			}
			c.TestSources = unique(c.TestSources)
			c.Comparators = unique(c.Comparators)
			out = append(out, c)
		}
	})
	return out, target, nil
}

func walkGroups(groups []interface{}, parents []string, visit func([]string, map[string]interface{})) {
	for _, item := range groups {
		m, ok := asMap(item)
		if !ok {
			continue
		}
		label := stringValue(m["id"])
		if text := stringValue(m["text"]); text != "" {
			if label != "" {
				label += " "
			}
			label += text
		}
		path := append(append([]string(nil), parents...), label)
		visit(path, m)
		children, _ := m["groups"].([]interface{})
		walkGroups(children, path, visit)
	}
}

func asMap(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}
func stringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
func unique(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "benchmark-inventory:", err); os.Exit(1) }

func writeMarkdown(path string, inv Inventory) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# kube-bench benchmark reference inventory\n\nGenerated from `%s`. This is a structural reference inventory of upstream YAML (including versioned EKS profiles). It does not execute audit commands, assert version equivalence with CIS EKS 2.1.0, or certify benchmark correctness.\n\n", inv.Source)
	fmt.Fprintf(&b, "- CIS profiles: **%d**\n- YAML control files: **%d**\n- Control entries: **%d**\n\n", len(inv.Profiles), inv.TotalFiles, inv.TotalControls)
	b.WriteString("## Profile summary\n\n| Profile | Files | Controls | Scored | Not scored | Scoring unspecified | Targets |\n|---|---:|---:|---:|---:|---:|---|\n")
	for _, p := range inv.Profiles {
		fmt.Fprintf(&b, "| `%s` | %d | %d | %d | %d | %d | %s |\n", p.Name, p.Files, len(p.Controls), p.Scored, p.NotScored, p.Unspecified, strings.Join(p.Targets, ", "))
	}
	for _, p := range inv.Profiles {
		fmt.Fprintf(&b, "\n## %s\n\n", p.Name)
		b.WriteString("| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |\n|---|---|---|---|---:|---|---|---|\n")
		for _, c := range p.Controls {
			scored := "unspecified"
			if c.Scored != nil {
				if *c.Scored {
					scored = "yes"
				} else {
					scored = "no"
				}
			}
			title := strings.ReplaceAll(c.Title, "|", "\\|")
			title = strings.ReplaceAll(title, "\n", " ")
			fmt.Fprintf(&b, "| %s | `%s` | %s | %s | %d | %s | %s | `%s` |\n", c.Target, c.ID, title, scored, c.TestItems, strings.Join(c.TestSources, ", "), strings.Join(c.Comparators, ", "), c.SourceFile)
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

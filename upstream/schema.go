// Package upstream parses kube-bench benchmark configuration and control YAML.
// It models upstream data only; it deliberately does not execute audit commands.
package upstream

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Root yaml.Node `yaml:"-"`
}
type BenchmarkFile struct {
	Groups []Group   `yaml:"-"`
	Root   yaml.Node `yaml:"-"`
}

type Group struct {
	ID        string    `yaml:"id"`
	Text      string    `yaml:"text"`
	Node      string    `yaml:"node"`
	Type      string    `yaml:"type"`
	Checks    []Check   `yaml:"checks"`
	Subgroups []Group   `yaml:"groups"`
	Raw       yaml.Node `yaml:"-"`
}

type Check struct {
	ID          string    `yaml:"id"`
	Text        string    `yaml:"text"`
	Type        string    `yaml:"type"`
	Scored      *bool     `yaml:"scored"`
	Audit       string    `yaml:"audit"`
	Tests       []Test    `yaml:"tests"`
	Remediation string    `yaml:"remediation"`
	Profile     string    `yaml:"profile"`
	Raw         yaml.Node `yaml:"-"`
}

// Test retains each upstream test item losslessly in Raw while exposing common fields.
type Test struct {
	BinOp   string      `yaml:"bin_op"`
	Compare *Compare    `yaml:"compare"`
	Flag    *FlagSource `yaml:"flag"`
	Path    *PathSource `yaml:"path"`
	Env     *EnvSource  `yaml:"env"`
	Raw     yaml.Node   `yaml:"-"`
}
type Compare struct {
	Op     string        `yaml:"op"`
	Value  interface{}   `yaml:"value"`
	Values []interface{} `yaml:"values"`
}

// FlagSource accepts both kube-bench's common scalar form (flag: --some-flag)
// and mapping forms used by some benchmark definitions. The original YAML node
// is retained by Test.Raw, so unknown mapping fields are not discarded.
type FlagSource struct {
	RawValue string   `yaml:"-"`
	Name     string   `yaml:"name"`
	Set      bool     `yaml:"set"`
	Value    string   `yaml:"value"`
	Values   []string `yaml:"values"`
}

func (f *FlagSource) UnmarshalYAML(node *yaml.Node) error {
	if f == nil || node == nil {
		return fmt.Errorf("flag source is nil")
	}
	switch node.Kind {
	case yaml.ScalarNode:
		f.RawValue = node.Value
		// Keep the scalar in Name as a convenient normalized representation.
		f.Name = node.Value
		return nil
	case yaml.MappingNode:
		type flagSourceAlias FlagSource
		var decoded flagSourceAlias
		if err := node.Decode(&decoded); err != nil {
			return err
		}
		*f = FlagSource(decoded)
		return nil
	default:
		return fmt.Errorf("flag source must be a scalar or mapping, got YAML kind %d", node.Kind)
	}
}

type PathSource struct {
	Path    string   `yaml:"path"`
	Compare *Compare `yaml:"compare"`
}
type EnvSource struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// ProfileMapping is a normalized mapping independent of upstream config layout.
type ProfileMapping struct {
	MinVersion string `json:"min_version" yaml:"min_version"`
	MaxVersion string `json:"max_version" yaml:"max_version"`
	Version    string `json:"version" yaml:"version"`
	Profile    string `json:"profile" yaml:"profile"`
	Target     string `json:"target,omitempty" yaml:"target,omitempty"`
}

type ProfileSelection struct {
	Profile           string
	MatchedBy         string
	KubernetesVersion string
	Distribution      string
}

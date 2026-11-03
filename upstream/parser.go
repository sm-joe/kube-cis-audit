package upstream

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseConfig(r io.Reader) (*Config, error) {
	var n yaml.Node
	d := yaml.NewDecoder(r)
	if err := d.Decode(&n); err != nil {
		return nil, fmt.Errorf("decode kube-bench config: %w", err)
	}
	if err := ensureSingleDocument(d); err != nil {
		return nil, err
	}
	if len(n.Content) == 0 || n.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("config root must be a YAML mapping")
	}
	return &Config{Root: n}, nil
}
func LoadConfig(path string) (*Config, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return ParseConfig(f)
}

func ParseBenchmark(r io.Reader) (*BenchmarkFile, error) {
	var root yaml.Node
	d := yaml.NewDecoder(r)
	if err := d.Decode(&root); err != nil {
		return nil, fmt.Errorf("decode benchmark YAML: %w", err)
	}
	if err := ensureSingleDocument(d); err != nil {
		return nil, err
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("benchmark root must be a YAML mapping")
	}
	var groups []Group
	if n := lookup(root.Content[0], "groups"); n != nil {
		if err := n.Decode(&groups); err != nil {
			return nil, fmt.Errorf("decode groups: %w", err)
		}
		attachGroupNodes(n, groups)
	}
	return &BenchmarkFile{Groups: groups, Root: root}, nil
}
func LoadBenchmark(path string) (*BenchmarkFile, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return ParseBenchmark(f)
}
func ensureSingleDocument(d *yaml.Decoder) error {
	var extra yaml.Node
	e := d.Decode(&extra)
	if e == io.EOF {
		return nil
	}
	if e != nil {
		return fmt.Errorf("decode trailing YAML document: %w", e)
	}
	return fmt.Errorf("multiple YAML documents are not supported")
}
func lookup(n *yaml.Node, key string) *yaml.Node {
	if n == nil {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}
func attachGroupNodes(n *yaml.Node, gs []Group) {
	for i := range gs {
		if i < len(n.Content) {
			gs[i].Raw = *n.Content[i]
			if c := lookup(n.Content[i], "checks"); c != nil {
				for j := range gs[i].Checks {
					if j < len(c.Content) {
						gs[i].Checks[j].Raw = *c.Content[j]
					}
				}
			}
			if sub := lookup(n.Content[i], "groups"); sub != nil {
				attachGroupNodes(sub, gs[i].Subgroups)
			}
		}
	}
}

// DecodeYAML is provided for callers that need an additional upstream field.
func DecodeYAML(n *yaml.Node, out interface{}) error {
	b, e := yaml.Marshal(n)
	if e != nil {
		return e
	}
	d := yaml.NewDecoder(bytes.NewReader(b))
	return d.Decode(out)
}
func (c *Config) RawYAML() ([]byte, error)        { return yaml.Marshal(&c.Root) }
func (b *BenchmarkFile) RawYAML() ([]byte, error) { return yaml.Marshal(&b.Root) }
func (c *Config) Profiles() []string {
	set := map[string]bool{}
	walk(c.Root.Content[0], func(k string, v *yaml.Node) {
		if strings.Contains(strings.ToLower(k), "benchmark") || strings.EqualFold(k, "profile") {
			if v.Kind == yaml.ScalarNode && v.Value != "" {
				set[v.Value] = true
			}
		}
	})
	out := make([]string, 0, len(set))
	for x := range set {
		out = append(out, x)
	}
	return out
}
func walk(n *yaml.Node, fn func(string, *yaml.Node)) {
	if n == nil {
		return
	}
	if n.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(n.Content); i += 2 {
			fn(n.Content[i].Value, n.Content[i+1])
			walk(n.Content[i+1], fn)
		}
	} else {
		for _, c := range n.Content {
			walk(c, fn)
		}
	}
}

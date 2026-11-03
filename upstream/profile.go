package upstream

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResolveProfile resolves only explicit version mappings present in config.yaml.
// It intentionally refuses to guess a profile from a folder name or nearest version.
func ResolveProfile(c *Config, kubernetesVersion, distribution string) (ProfileSelection, error) {
	if c == nil || len(c.Root.Content) == 0 {
		return ProfileSelection{}, fmt.Errorf("config is nil or empty")
	}
	want, err := parseVersion(kubernetesVersion)
	if err != nil {
		return ProfileSelection{}, err
	}
	var candidates []ProfileMapping
	collectMappings(c.Root.Content[0], &candidates)
	matches := make([]ProfileMapping, 0)
	for _, m := range candidates {
		if m.Profile == "" {
			continue
		}
		if m.Target != "" && distribution != "" && !strings.EqualFold(m.Target, distribution) {
			continue
		}
		if m.Version != "" {
			v, e := parseVersion(m.Version)
			if e == nil && compareVersion(want, v) == 0 {
				matches = append(matches, m)
			}
			continue
		}
		minOK, maxOK := true, true
		if m.MinVersion != "" {
			v, e := parseVersion(m.MinVersion)
			minOK = e == nil && compareVersion(want, v) >= 0
		}
		if m.MaxVersion != "" {
			v, e := parseVersion(m.MaxVersion)
			maxOK = e == nil && compareVersion(want, v) <= 0
		}
		if minOK && maxOK && (m.MinVersion != "" || m.MaxVersion != "") {
			matches = append(matches, m)
		}
	}
	if len(matches) == 0 {
		return ProfileSelection{}, fmt.Errorf("no explicit kube-bench profile mapping for Kubernetes %s (distribution %q)", kubernetesVersion, distribution)
	}
	profiles := map[string]bool{}
	for _, m := range matches {
		profiles[m.Profile] = true
	}
	if len(profiles) != 1 {
		return ProfileSelection{}, fmt.Errorf("ambiguous profile mapping for Kubernetes %s: %v", kubernetesVersion, keys(profiles))
	}
	var selected string
	for p := range profiles {
		selected = p
	}
	return ProfileSelection{Profile: selected, MatchedBy: "explicit config.yaml version mapping", KubernetesVersion: kubernetesVersion, Distribution: distribution}, nil
}
func collectMappings(n *yaml.Node, out *[]ProfileMapping) {
	if n == nil {
		return
	}
	if n.Kind == yaml.MappingNode {
		m := ProfileMapping{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := strings.ToLower(n.Content[i].Value), n.Content[i+1]
			if v.Kind != yaml.ScalarNode {
				continue
			}
			switch k {
			case "profile", "benchmark", "benchmark_version", "benchmark_name":
				m.Profile = v.Value
			case "version", "kubernetes_version", "k8s_version":
				m.Version = v.Value
			case "min_version", "min_kubernetes_version", "min_k8s_version":
				m.MinVersion = v.Value
			case "max_version", "max_kubernetes_version", "max_k8s_version":
				m.MaxVersion = v.Value
			case "distribution", "target", "provider":
				m.Target = v.Value
			}
		}
		if m.Profile != "" && (m.Version != "" || m.MinVersion != "" || m.MaxVersion != "") {
			*out = append(*out, m)
		}
	}
	for _, c := range n.Content {
		collectMappings(c, out)
	}
}

type version struct{ major, minor, patch int }

func parseVersion(s string) (version, error) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "v"))
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return version{}, fmt.Errorf("invalid Kubernetes version %q: expected major.minor[.patch]", s)
	}
	var v version
	vals := []*int{&v.major, &v.minor, &v.patch}
	for i, p := range parts {
		n, e := strconv.Atoi(p)
		if e != nil || n < 0 {
			return version{}, fmt.Errorf("invalid Kubernetes version %q", s)
		}
		*vals[i] = n
	}
	return v, nil
}
func compareVersion(a, b version) int {
	av := []int{a.major, a.minor, a.patch}
	bv := []int{b.major, b.minor, b.patch}
	for i := range av {
		if av[i] < bv[i] {
			return -1
		}
		if av[i] > bv[i] {
			return 1
		}
	}
	return 0
}
func keys(m map[string]bool) []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

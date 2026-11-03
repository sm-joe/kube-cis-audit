package upstream

import (
	"strings"
	"testing"
)

func TestResolveProfileExplicitRange(t *testing.T) {
	c, e := ParseConfig(strings.NewReader("version_mapping:\n  - min_version: '1.24.0'\n    max_version: '1.26.99'\n    benchmark: cis-1.24\n"))
	if e != nil {
		t.Fatal(e)
	}
	got, e := ResolveProfile(c, "v1.25.3", "kubeadm")
	if e != nil {
		t.Fatal(e)
	}
	if got.Profile != "cis-1.24" {
		t.Fatalf("got %q", got.Profile)
	}
}
func TestResolveProfileNoGuess(t *testing.T) {
	c, e := ParseConfig(strings.NewReader("profiles: [cis-1.24]\n"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ResolveProfile(c, "1.25.0", ""); e == nil {
		t.Fatal("expected explicit mapping error")
	}
}
func TestResolveProfileAmbiguous(t *testing.T) {
	c, e := ParseConfig(strings.NewReader("version_mapping:\n - min_version: '1.24.0'\n   max_version: '1.26.0'\n   benchmark: cis-a\n - min_version: '1.25.0'\n   max_version: '1.27.0'\n   benchmark: cis-b\n"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ResolveProfile(c, "1.25.0", ""); e == nil {
		t.Fatal("expected ambiguity")
	}
}

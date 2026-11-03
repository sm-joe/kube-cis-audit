package upstream

import (
	"strings"
	"testing"
)

func TestParseBenchmarkPreservesNestedGroupsAndAudit(t *testing.T) {
	src := `groups:
- id: 1
  text: Master
  checks:
  - id: 1.1.1
    text: Check a file
    scored: true
    audit: /bin/ps -ef
    tests:
    - bin_op: or
      flag: --some-flag
      compare:
        op: eq
        value: true
    remediation: Fix it
  groups:
  - id: 1.2
    text: Nested
    checks:
    - id: 1.2.1
      text: Manual check
      scored: false
      type: manual
`
	b, e := ParseBenchmark(strings.NewReader(src))
	if e != nil {
		t.Fatal(e)
	}
	if len(b.Groups) != 1 || len(b.Groups[0].Subgroups) != 1 {
		t.Fatal("nested groups lost")
	}
	c := b.Groups[0].Checks[0]
	if c.Audit != "/bin/ps -ef" || len(c.Tests) != 1 || c.Tests[0].Compare.Op != "eq" {
		t.Fatalf("control fields lost: %#v", c)
	}
	if got := c.Tests[0].Flag; got == nil || got.Name != "--some-flag" || got.RawValue != "--some-flag" {
		t.Fatalf("scalar flag source was not preserved: %#v", got)
	}
	if got := ValidateBenchmark(b); len(got) != 0 {
		t.Fatalf("unexpected issues: %#v", got)
	}
}
func TestScoredUnknownIsNotDefaulted(t *testing.T) {
	b, e := ParseBenchmark(strings.NewReader("groups:\n- id: 1\n  text: Root\n  checks:\n  - id: 1.1\n    text: Control\n"))
	if e != nil {
		t.Fatal(e)
	}
	if b.Groups[0].Checks[0].Scored != nil {
		t.Fatal("scored should remain unknown")
	}
	if len(ValidateBenchmark(b)) == 0 {
		t.Fatal("expected unknown scoring issue")
	}
}
func TestRejectMultipleDocuments(t *testing.T) {
	_, e := ParseBenchmark(strings.NewReader("groups: []\n---\ngroups: []\n"))
	if e == nil {
		t.Fatal("expected multi-document rejection")
	}
}

func TestParseFlagSourceMapping(t *testing.T) {
	src := `groups:
- id: 1
  text: Master
  checks:
  - id: 1.1.1
    text: Flag mapping
    scored: true
    tests:
    - flag:
        name: --secure-port
        set: true
`
	b, err := ParseBenchmark(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	flag := b.Groups[0].Checks[0].Tests[0].Flag
	if flag == nil || flag.Name != "--secure-port" || !flag.Set {
		t.Fatalf("mapping flag source was not decoded: %#v", flag)
	}
}

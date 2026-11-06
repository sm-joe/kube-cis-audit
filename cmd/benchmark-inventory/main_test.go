package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildInventoryWalksNestedGroupsAndIncludesVersionedEKSProfiles(t *testing.T) {
	root := t.TempDir()
	profile := filepath.Join(root, "cis-test")
	if err := os.MkdirAll(profile, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "eks-1.8.0"), 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `controls:
  id: 1
  text: Control plane
  type: master
  groups:
    - id: 1.1
      text: API server
      checks:
        - id: 1.1.1
          text: Example scored check (Automated)
          scored: true
          audit: stat /some/path
          tests:
            test_items:
              - flag: permissions
                compare:
                  op: bitmask
                  value: "600"
    - id: 1.2
      text: Nested parent
      groups:
        - id: 1.2.1
          text: Nested child
          checks:
            - id: 1.2.1.1
              text: Manual check
              scored: false
`
	if err := os.WriteFile(filepath.Join(profile, "master.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "eks-1.8.0", "node.yaml"), []byte(yaml), 0o600); err != nil { t.Fatal(err) }
	inv, err := buildInventory(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(inv.Profiles) != 2 || inv.TotalControls != 4 {
		t.Fatalf("unexpected inventory: %#v", inv)
	}
	p := inv.Profiles[0]
	if p.Scored != 1 || p.NotScored != 1 {
		t.Fatalf("unexpected scoring counts: %#v", p)
	}
	if p.Controls[0].TestItems != 1 || len(p.Controls[0].Comparators) != 1 || p.Controls[0].Comparators[0] != "bitmask" {
		t.Fatalf("test metadata missing: %#v", p.Controls[0])
	}
}

func TestParseControlFileUsesSiblingFieldsAfterControlsMarker(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "controlplane.yaml")
	fixture := `controls:
id: 1
text: "Control Plane Security Configuration"
type: "controlplane"
groups:
  - id: 1.1
    text: "Control Plane Node Configuration Files"
    checks:
      - id: 1.1.1
        text: "Ensure file permissions are restrictive"
        audit: "stat -c permissions=%a /etc/example"
        tests:
          test_items:
            - flag: "permissions"
              compare:
                op: bitmask
                value: "600"
        remediation: "Set restrictive permissions"
        scored: true
`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	controls, target, err := parseControlFile(path, dir, "cis-test")
	if err != nil {
		t.Fatalf("parseControlFile returned error: %v", err)
	}
	if target != "controlplane" {
		t.Fatalf("target = %q, want controlplane", target)
	}
	if len(controls) != 1 {
		t.Fatalf("got %d controls, want 1", len(controls))
	}
	if controls[0].ID != "1.1.1" || controls[0].CheckType != "" || controls[0].TestItems != 1 {
		t.Fatalf("unexpected parsed control: %#v", controls[0])
	}
	if controls[0].Scored == nil || !*controls[0].Scored {
		t.Fatalf("expected scored=true, got %#v", controls[0].Scored)
	}
}

func TestParseControlFileAcceptsWrappedControlsMapping(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "node.yaml")
	fixture := `controls:
  type: "node"
  groups:
    - id: 4
      text: "Worker Node"
      checks:
        - id: 4.1.1
          text: "Example"
          scored: false
`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	controls, target, err := parseControlFile(path, dir, "cis-test")
	if err != nil {
		t.Fatalf("parseControlFile returned error: %v", err)
	}
	if target != "node" || len(controls) != 1 || controls[0].ID != "4.1.1" {
		t.Fatalf("unexpected parse result: target=%q controls=%#v", target, controls)
	}
}

func TestBuildInventoryRejectsUnrecognizedProfiles(t *testing.T) {
    root := t.TempDir()
    if err := os.MkdirAll(filepath.Join(root, "my-benchmark"), 0o755); err != nil { t.Fatal(err) }
    if _, err := buildInventory(root); err == nil { t.Fatal("accepted an unknown benchmark profile as official CIS/EKS") }
}

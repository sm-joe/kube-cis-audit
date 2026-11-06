package audit

import (
    "strings"
    "testing"
)

func TestValidateBenchmarkMinorFailsClosed(t *testing.T) {
    b := Benchmark{ID: "cis-2.1.0", KubernetesMinor: []string{"31", "37"}}
    for _, minor := range []string{"31", "37"} {
        if err := validateBenchmarkMinor(b, minor); err != nil {
            t.Fatalf("listed version %s rejected: %v", minor, err)
        }
    }
    for _, minor := range []string{"30", "32", "", "137"} {
        err := validateBenchmarkMinor(b, minor)
        if err == nil {
            t.Fatalf("unlisted minor %q was accepted", minor)
        }
    }
}

func TestEKSNodeadmPermissionsRemainASeparateUnscoredExtension(t *testing.T) {
    original := Benchmark{ID: "cis-2.1.0", Checks: []Check{
        {ID: "4.1.9", Title: "generic config mode", Target: "node", Kind: "host_file_mode", Path: "var/lib/kubelet/config.yaml", MaxMode: "0600", Scored: true},
        {ID: "5.2.1", Target: "policies", Kind: "api_policy_control", Scored: true},
    }}
    eks := applyProviderProfile(original, "eks")
    if len(eks.Checks) != 2 {
        t.Fatal("profile checks lost")
    }
    c := eks.Checks[0]
    if c.ID != eksKubeletModeExtensionID || c.MaxMode != "0644" || c.Scored ||
        !strings.Contains(c.Description, "not an official CIS") {
        t.Fatalf("unverified EKS mapping was treated as official: %+v", c)
    }
    if eks.Checks[1].ID != "5.2.1" || !eks.Checks[1].Scored {
        t.Fatal("unrelated benchmark control was modified")
    }
    if original.Checks[0].ID != "4.1.9" || original.Checks[0].MaxMode != "0600" {
        t.Fatal("input benchmark was mutated")
    }
    kind := applyProviderProfile(original, "unknown")
    if kind.Checks[0].ID != "4.1.9" || kind.Checks[0].MaxMode != "0600" ||
        !kind.Checks[0].Scored {
        t.Fatal("Kind kubelet control changed")
    }
}

func TestEKSProvisionalNoteNamesActualServer(t *testing.T) {
    b := Benchmark{CompatibilityNote: "Provisional."}
    note := benchmarkCompatibilityNote(b, "v1.31.14-eks-577a51e", "eks")
    if !strings.Contains(note, "v1.31.14-eks-577a51e") || !strings.Contains(note, "official CIS EKS") {
        t.Fatalf("missing version or scope warning: %s", note)
    }
}

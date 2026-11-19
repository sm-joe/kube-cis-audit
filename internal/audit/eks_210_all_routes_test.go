package audit

import (
	"context"
	"path/filepath"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

// Every factual control ID must be connected to a real evaluator or explicit
// human-evidence review path. This is an implementation-routing test, NOT
// acceptance of its complete CIS audit semantics or a certification claim.
func TestEKS210AllOfficialControlsHaveExecutableRoutes(t *testing.T) {
	catalog, err := LoadCIS210ControlInventory("eks")
	if err != nil {
		t.Fatal(err)
	}
	bench, err := readBench(filepath.Join("..", "..", "benchmarks", "cis-eks-2.1.0", "benchmark.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	providerIDs := map[string]bool{
		"2.1.1": true, "4.1.7": true, "5.1.1": true,
		"5.4.1": true, "5.4.2": true, "5.5.1": true,
	}
	checks := make(map[string]Check)
	for _, c := range bench.Checks {
		if _, exists := checks[c.ID]; exists {
			t.Fatalf("duplicate native EKS check %s", c.ID)
		}
		checks[c.ID] = c
	}
	client := fake.NewSimpleClientset()
	for _, row := range catalog.Controls {
		if !row.Implemented {
			t.Fatalf("%s has a code path but catalog still marks it not implemented", row.ID)
		}
		if providerIDs[row.ID] {
			if _, exists := checks[row.ID]; exists {
				t.Fatalf("%s is provider-owned and must not be duplicated in cluster benchmark", row.ID)
			}
			continue
		}
		c, exists := checks[row.ID]
		if !exists {
			t.Fatalf("%s missing native control", row.ID)
		}
		if c.Kind == "cis_not_implemented" {
			t.Fatalf("%s is routed to unimplemented fallback", row.ID)
		}
		if row.Assessment == "manual" {
			if !c.Manual || c.Kind != "manual" || c.Scored {
				t.Fatalf("%s requires non-scored explicit human review, got %+v", row.ID, c)
			}
			r := evaluate(context.Background(), client, c, Options{Benchmark: "cis-eks-2.1.0"})
			if r.ID != row.ID || (r.Status != "MANUAL" && r.Status != "NOT_ASSESSED") {
				t.Fatalf("%s human-review path must never automatically PASS: %+v", row.ID, r)
			}
			continue
		}
		if c.Manual || !c.Scored {
			t.Fatalf("%s automated control must be scored with evidence: %+v", row.ID, c)
		}
		if c.Target == "node" {
			if c.Kind != "host_file_mode" && c.Kind != "host_file_owner" && c.Kind != "host_kubelet_config_field" {
				t.Fatalf("%s unsupported node evaluator: %s", row.ID, c.Kind)
			}
			if c.Kind == "host_file_mode" || c.Kind == "host_file_owner" {
				if _, ok := approvedHostMetadataPath("node", c.Path); !ok {
					t.Fatalf("%s unapproved host metadata path %q", row.ID, c.Path)
				}
			}
			continue
		}
		switch c.Kind {
		case "eks_default_service_account", "eks_token_mounts", "eks_psa_admission",
			"eks_secret_env", "eks_default_namespace", "eks_vpc_cni_network_policy",
			"eks_dedicated_service_account":
			r := evaluate(context.Background(), client, c, Options{Benchmark: "cis-eks-2.1.0"})
			if r.ID != row.ID || r.Method == "cis_not_implemented" {
				t.Fatalf("%s evaluator did not run: %+v", row.ID, r)
			}
		default:
			t.Fatalf("%s unsupported automated policy kind %s", row.ID, c.Kind)
		}
	}
	if len(checks) != 41 || len(catalog.Controls) != 47 {
		t.Fatalf("expected 41 native + 6 provider, got %d native + %d controls", len(checks), len(catalog.Controls))
	}
	logs := map[string]bool{"api": false, "audit": false, "authenticator": false, "controllerManager": false, "scheduler": false}
	provider := EKSProviderEvidence{
		SchemaVersion: 1, CloudProvider: "eks", AssessmentStatus: "INVENTORY_ONLY",
		ControlPlaneLogging: logs, AccessInventoryStatus: "UNAVAILABLE",
		ECRInventoryStatus: "UNAVAILABLE",
	}
	existing, err := EvaluateEKSCISProviderControls(provider)
	if err != nil {
		t.Fatal(err)
	}
	existing = append(existing, EvaluateEKSAdditionalAWSControls(provider)...)
	seen := map[string]bool{}
	for _, r := range existing {
		if !providerIDs[r.ID] || seen[r.ID] || r.Status == "PASS" {
			t.Fatalf("invalid provider control routing without full evidence: %+v", r)
		}
		seen[r.ID] = true
	}
	if len(seen) != 6 {
		t.Fatalf("only %d of 6 AWS provider control evaluators are wired", len(seen))
	}
}

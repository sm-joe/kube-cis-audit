package audit

import (
	"strings"
	"testing"
)

// IDs and Automated/Manual classes were cross-checked against the operator's
// privately held CIS Amazon EKS Benchmark v2.1.0 PDF, SHA-256:
// c3c9eb01253643a0176b21aafd19cc7c08af0a19872f869712d2d82be8a967e9.
// These are factual identifiers only; no licensed audit text is redistributed.
// Source inventory agreement is NOT evaluator conformance or CIS certification.
func TestPinnedEKS210SourceControlIdentitiesAndClasses(t *testing.T) {
	automated := strings.Fields(`
2.1.1
3.1.1 3.1.2 3.1.3 3.1.4
3.2.1 3.2.2 3.2.3 3.2.4 3.2.5 3.2.6 3.2.7 3.2.8
4.1.5 4.1.6 4.1.7
4.2.1 4.2.2 4.2.3 4.2.4 4.2.5
4.3.1 4.4.1 4.5.2
5.1.1 5.2.1 5.4.1 5.4.2 5.5.1`)
	manual := strings.Fields(`
4.1.1 4.1.2 4.1.3 4.1.4
4.1.8 4.1.9 4.1.10 4.1.11 4.1.12
4.3.2 4.4.2 4.5.1
5.1.2 5.1.3 5.1.4 5.3.1 5.4.3 5.4.4`)
	expected := make(map[string]string, 47)
	for _, id := range automated {
		expected[id] = "automated"
	}
	for _, id := range manual {
		if expected[id] != "" {
			t.Fatalf("duplicate official control %s", id)
		}
		expected[id] = "manual"
	}
	if len(expected) != 47 || len(automated) != 29 || len(manual) != 18 {
		t.Fatalf("pinned CIS EKS inventory incorrect: %d identifiers, %d automated, %d manual", len(expected), len(automated), len(manual))
	}
	catalog, err := LoadCIS210ControlInventory("eks")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range catalog.Controls {
		class, ok := expected[c.ID]
		if !ok || class != c.Assessment {
			t.Fatalf("catalog control %s=%s differs from official source class %s", c.ID, c.Assessment, class)
		}
		delete(expected, c.ID)
	}
	if len(expected) != 0 {
		t.Fatalf("missing official source control IDs: %+v", expected)
	}
}

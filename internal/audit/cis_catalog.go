package audit

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed catalogs/eks-2.1.0.json catalogs/kubernetes-2.1.0.json
var officialControlIndex embed.FS

// CISControlInventory contains only control identifiers, classification,
// provenance pages and explicit implementation state. It contains no licensed
// benchmark text. The PDF is intentionally NOT distributed with KubeCIS.
type CISControlInventory struct {
	Benchmark                string               `json:"benchmark"`
	Status                   string               `json:"status"`
	SupportedKubernetesMinor []string             `json:"supportedKubernetesMinor"`
	Source                   string               `json:"source"`
	Note                     string               `json:"note"`
	Controls                 []CISControlIdentity `json:"controls"`
}
type CISControlIdentity struct {
	ID          string `json:"id"`
	Assessment  string `json:"assessment"`
	PDFPage     int    `json:"pdfPage"`
	Implemented bool   `json:"implemented"`
}

// LoadCIS210ControlInventory enumerates the control inventory from the
// operator-provided 2026 official benchmarks. This inventory's non-implementation
// entries prevent a partial native audit from ever masquerading as a full scan.
func LoadCIS210ControlInventory(flavor string) (CISControlInventory, error) {
	names := map[string]string{"eks": "catalogs/eks-2.1.0.json", "kubernetes": "catalogs/kubernetes-2.1.0.json"}
	file, ok := names[strings.ToLower(flavor)]
	if !ok {
		return CISControlInventory{}, fmt.Errorf("unsupported CIS 2.1.0 flavor %q", flavor)
	}
	b, err := officialControlIndex.ReadFile(file)
	if err != nil {
		return CISControlInventory{}, err
	}
	var inv CISControlInventory
	if err := json.Unmarshal(b, &inv); err != nil {
		return CISControlInventory{}, fmt.Errorf("decode %s catalog: %w", flavor, err)
	}
	expected := map[string]int{"eks": 47, "kubernetes": 134}[strings.ToLower(flavor)]
	if len(inv.Controls) != expected || inv.Status != "inventory" {
		return inv, fmt.Errorf("%s benchmark inventory is incomplete: got %d, expected %d", flavor, len(inv.Controls), expected)
	}
	ids := make(map[string]bool, len(inv.Controls))
	for _, c := range inv.Controls {
		if c.ID == "" || ids[c.ID] || c.PDFPage < 1 || (c.Assessment != "automated" && c.Assessment != "manual") {
			return inv, fmt.Errorf("invalid or duplicate %s CIS control %q", flavor, c.ID)
		}
		ids[c.ID] = true
	}
	return inv, nil
}

// ValidateCIS210FullCoverage rejects full-release status unless all official
// benchmark recommendations are implemented and each required automated
// control has an evidence-supported decisive status. Manual controls require
// an explicit completed reviewer disposition, never a fabricated PASS.
func ValidateCIS210FullCoverage(inv CISControlInventory, results []Result, reviewedManual map[string]bool) error {
	if len(inv.Controls) == 0 {
		return fmt.Errorf("empty CIS control inventory")
	}
	byID := make(map[string][]Result, len(results))
	for _, r := range results {
		byID[r.ID] = append(byID[r.ID], r)
	}
	missing := make([]string, 0)
	for _, c := range inv.Controls {
		if !c.Implemented {
			missing = append(missing, c.ID+":not_implemented")
			continue
		}
		matches := byID[c.ID]
		if len(matches) == 0 {
			missing = append(missing, c.ID+":no_evidence")
			continue
		}
		if c.Assessment == "manual" {
			if !reviewedManual[c.ID] {
				missing = append(missing, c.ID+":manual_review_pending")
			}
			continue
		}
		for _, r := range matches {
			if !r.Scored || (r.Status != "PASS" && r.Status != "FAIL") || len(r.Evidence) == 0 {
				missing = append(missing, c.ID+":unverified_automated_result")
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("CIS %s full release blocked: %d unresolved control items (%s)", inv.Benchmark, len(missing), strings.Join(missing, ","))
	}
	return nil
}

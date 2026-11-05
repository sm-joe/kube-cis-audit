package conformance

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var allowed = map[string]bool{"candidate": true, "verified": true, "disabled": true}

func Decode(r io.Reader) (Inventory, error) {
	var v Inventory
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
func Validate(v Inventory) Report {
	out := Report{Profiles: len(v.Profiles), ReleaseReady: true}
	seenProfiles := map[string]bool{}
	for _, p := range v.Profiles {
		if strings.TrimSpace(p.Name) == "" {
			out.Findings = append(out.Findings, Finding{Profile: p.Name, Code: "PROFILE_NAME_MISSING", Detail: "profile name is required"})
			out.ReleaseReady = false
		}
		if seenProfiles[p.Name] {
			out.Findings = append(out.Findings, Finding{Profile: p.Name, Code: "DUPLICATE_PROFILE", Detail: "profile name is duplicated"})
			out.ReleaseReady = false
		}
		seenProfiles[p.Name] = true
		ids := map[string]bool{}
		for _, m := range p.Controls {
			out.Controls++
			if m.ControlID == "" || m.Title == "" {
				out.Findings = append(out.Findings, Finding{p.Name, m.ControlID, "CONTROL_METADATA_MISSING", "control ID and title are required"})
				out.ReleaseReady = false
			}
			if ids[m.ControlID] {
				out.Findings = append(out.Findings, Finding{p.Name, m.ControlID, "DUPLICATE_CONTROL_ID", "duplicate control ID in profile"})
			}
			ids[m.ControlID] = true
			if !allowed[m.Status] {
				out.Findings = append(out.Findings, Finding{p.Name, m.ControlID, "INVALID_MAPPING_STATUS", fmt.Sprintf("unsupported mapping status %q", m.Status)})
				out.ReleaseReady = false
				continue
			}
			switch m.Status {
			case "candidate":
				out.Candidate++
				out.ReleaseReady = false
			case "verified":
				out.Verified++
				if m.EvidenceKind == "" || m.SourceFile == "" {
					out.Findings = append(out.Findings, Finding{p.Name, m.ControlID, "VERIFICATION_PROVENANCE_MISSING", "verified mappings require evidence_kind and source_file"})
					out.ReleaseReady = false
				}
			case "disabled":
				out.Disabled++
			}
		}
		if p.BenchmarkVersion == "" {
			out.Findings = append(out.Findings, Finding{p.Name, "", "BENCHMARK_VERSION_MISSING", "benchmark version is required"})
			out.ReleaseReady = false
		}
	}
	if out.Verified == 0 {
		out.ReleaseReady = false
	}
	return out
}

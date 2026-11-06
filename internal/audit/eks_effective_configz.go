package audit

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// EnrichNodeWithConfigz applies the live, effective kubelet /configz facts to
// only the three CIS checks whose file-level evidence is often incomplete on
// EKS nodeadm. Caller obtains configz through the authenticated Kubernetes API
// for the same node. No raw configz or credentials are written to audit output.
func EnrichNodeWithConfigz(report *Report, nodeName string, data []byte) error {
	if report == nil || report.AuditMode != "node-local" || report.CloudProvider != "eks" || report.Benchmark != "cis-eks-2.1.0" || report.NodeName != nodeName || nodeName == "" {
		return fmt.Errorf("effective kubelet evidence must match an EKS node-local report and exact node name")
	}
	if len(data) > 1024*1024 {
		return fmt.Errorf("kubelet configz exceeds 1 MiB")
	}
	var root struct {
		KubeletConfig map[string]json.RawMessage `json:"kubeletconfig"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("decode effective kubelet configz: %w", err)
	}
	if len(root.KubeletConfig) == 0 {
		return fmt.Errorf("effective kubelet configz has no kubeletconfig object")
	}
	targets := map[string]string{"3.2.5": "makeIPTablesUtilChains", "3.2.6": "eventRecordQPS", "3.2.7": "rotateCertificates"}
	found := map[string]bool{}
	for i := range report.Results {
		r := &report.Results[i]
		field, ok := targets[r.ID]
		if !ok {
			continue
		}
		if r.Scope != nodeName || r.Target != "node" || r.Method != "host_kubelet_config_field" || found[r.ID] {
			return fmt.Errorf("unexpected or duplicated kubelet control %s scope", r.ID)
		}
		found[r.ID] = true
		raw, present := root.KubeletConfig[field]
		r.EvidenceSource = "kubernetes-node-proxy-configz"
		r.Status = "NOT_ASSESSED"
		r.Details = "Effective kubelet /configz did not expose " + field + "; do not assume defaults."
		r.Evidence = nil
		if !present || string(raw) == "null" {
			continue
		}
		switch r.ID {
		case "3.2.5":
			var value bool
			if err := json.Unmarshal(raw, &value); err != nil {
				r.Details = "Invalid makeIPTablesUtilChains value in /configz"
				continue
			}
			r.Evidence = []Evidence{{Resource: "Node", Name: nodeName, Field: field, Observed: strconv.FormatBool(value), Expected: "true"}}
			r.Status = "FAIL"
			if value {
				r.Status = "PASS"
			}
			r.Details = "Evaluated live effective kubelet iptables chain configuration via /configz."
		case "3.2.6":
			var value int
			if err := json.Unmarshal(raw, &value); err != nil || value < 0 || value > 65535 {
				r.Details = "Invalid effective eventRecordQPS in /configz"
				continue
			}
			r.Evidence = []Evidence{{Resource: "Node", Name: nodeName, Field: field, Observed: strconv.Itoa(value), Expected: "workload-appropriate bounded rate; default 50; 0 is unlimited"}}
			r.Status = "MANUAL"
			r.Details = fmt.Sprintf("Effective eventRecordQPS=%d; human validation of event throughput and throttling is required. Zero disables rate limiting and is not an automatic PASS.", value)
		case "3.2.7":
			var value bool
			if err := json.Unmarshal(raw, &value); err != nil {
				r.Details = "Invalid effective rotateCertificates in /configz"
				continue
			}
			r.Evidence = []Evidence{{Resource: "Node", Name: nodeName, Field: field, Observed: strconv.FormatBool(value), Expected: "true when client certificate authentication is used; otherwise not applicable"}}
			r.Status = "MANUAL"
			r.Details = fmt.Sprintf("Effective rotateCertificates=%t. Inspect active kubelet kubeconfig to determine certificate versus IAM exec-token authentication before deciding PASS/FAIL/NOT_APPLICABLE.", value)
		}
	}
	for id := range targets {
		if !found[id] {
			return fmt.Errorf("node report is missing expected kubelet control %s", id)
		}
	}
	var summary Summary
	var scored ComplianceSummary
	for _, r := range report.Results {
		incrementSummary(&summary, r.Status)
		if r.Scored {
			scored.ScoredTotal++
			switch r.Status {
			case "PASS":
				scored.ScoredPass++
			case "FAIL":
				scored.ScoredFail++
			case "WARN":
				scored.ScoredWarn++
			case "MANUAL":
				scored.ScoredManual++
			case "NOT_ASSESSED":
				scored.ScoredNotAssessed++
			default:
				scored.ScoredOther++
			}
		}
	}
	summary.Total = len(report.Results)
	if scored.ScoredTotal > 0 {
		scored.CoveragePercent = math.Round(float64(scored.ScoredPass+scored.ScoredFail)*10000/float64(scored.ScoredTotal)) / 100
	}
	scored.ScoreStatus = "withheld: incomplete official CIS EKS 2.1.0 control coverage"
	report.Summary = summary
	report.Compliance = scored
	return nil
}

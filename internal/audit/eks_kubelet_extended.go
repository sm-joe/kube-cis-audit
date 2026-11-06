package audit

import (
	"fmt"
	"strconv"
	"strings"
)

// The node helper exports only bounded, allowlisted facts, not raw kubelet
// configuration or secret-bearing file contents.
func collectExtendedKubeletSettings(cfg map[string]interface{}, result map[string]KubeletSetting) {
	specs := map[string]string{
		"authentication.x509.clientCAFile":             "path",
		"readOnlyPort":                                 "integer",
		"makeIPTablesUtilChains":                       "boolean",
		"eventRecordQPS":                               "integer",
		"rotateCertificates":                           "boolean",
		"serverTLSBootstrap":                           "boolean",
		"featureGates.RotateKubeletServerCertificate": "boolean",
	}
	for field, kind := range specs {
		var value interface{} = cfg
		for _, part := range strings.Split(field, ".") {
			node, ok := value.(map[string]interface{})
			if !ok {
				result[field] = KubeletSetting{Error: "field not explicitly configured"}
				value = nil
				break
			}
			next, exists := node[part]
			if !exists {
				result[field] = KubeletSetting{Error: "field not explicitly configured"}
				value = nil
				break
			}
			value = next
		}
		if value == nil {
			continue
		}
		switch kind {
		case "boolean":
			if v, ok := value.(bool); ok {
				result[field] = KubeletSetting{Value: strconv.FormatBool(v)}
			} else {
				result[field] = KubeletSetting{Error: "field is not an explicit boolean"}
			}
		case "integer":
			switch v := value.(type) {
			case int:
				if v >= 0 && v <= 65535 {
					result[field] = KubeletSetting{Value: strconv.Itoa(v)}
				} else {
					result[field] = KubeletSetting{Error: "value outside allowed range"}
				}
			default:
				result[field] = KubeletSetting{Error: "field is not an explicit non-negative integer"}
			}
		case "path":
			if v, ok := value.(string); ok {
				if strings.TrimSpace(v) != "" && len(v) <= 4096 && strings.HasPrefix(v, "/") {
					result[field] = KubeletSetting{Value: "configured"}
				} else {
					result[field] = KubeletSetting{Value: "not-configured"}
				}
			} else {
				result[field] = KubeletSetting{Error: "field is not a string"}
			}
		}
	}
}

func evaluateExtendedKubeletSetting(nodeName string, c Check, snapshot NodeMetadataSnapshot) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Scored: c.Scored, Remediation: c.Remediation}
	fields := map[string]string{
		"3.2.3": "authentication.x509.clientCAFile",
		"3.2.4": "readOnlyPort",
		"3.2.5": "makeIPTablesUtilChains",
		"3.2.6": "eventRecordQPS",
		"3.2.7": "rotateCertificates",
		"3.2.8": "serverTLSBootstrap",
	}
	field, allowed := fields[c.ID]
	if !allowed || field != c.Field || c.Path != "var/lib/kubelet/config.yaml" {
		r.Details = "Control/field/path combination not allowlisted"
		return r
	}
	setting, ok := snapshot.KubeletSettings[field]
	if !ok || setting.Error != "" || setting.Value == "" {
		r.Details = "Kubelet configuration evidence unavailable"
		if setting.Error != "" {
			r.Details += ": " + setting.Error
		}
		return r
	}
	if setting.SourcePath != c.Path && setting.SourcePath != "etc/kubernetes/kubelet/config.json" {
		r.Details = "Kubelet setting source path not allowlisted"
		return r
	}
	expected := ""
	switch c.ID {
	case "3.2.3":
		expected = "configured"
	case "3.2.4":
		expected = "0"
	case "3.2.5", "3.2.7", "3.2.8":
		expected = "true"
	case "3.2.6":
		expected = "0 or justified workload-specific QPS"
	}
	r.Evidence = []Evidence{{Resource: "host-file", Name: setting.SourcePath, Field: field, Observed: setting.Value, Expected: expected}}
	switch c.ID {
	case "3.2.6":
		if setting.Value == "0" {
			r.Status = "PASS"
		} else {
			if _, err := strconv.Atoi(setting.Value); err != nil {
				r.Status = "NOT_ASSESSED"
			} else {
				r.Status = "MANUAL"
			}
		}
	case "3.2.7":
		if setting.Value == "true" {
			r.Status = "PASS"
		} else if setting.Value == "false" {
			r.Status = "MANUAL"
		}
	case "3.2.8":
		gate, ok := snapshot.KubeletSettings["featureGates.RotateKubeletServerCertificate"]
		if !ok || gate.Error != "" || gate.Value == "" || (gate.SourcePath != c.Path && gate.SourcePath != "etc/kubernetes/kubelet/config.json") {
			r.Details = "Kubelet server TLS rotation feature gate evidence unavailable"
			return r
		}
		r.Evidence = append(r.Evidence, Evidence{Resource: "host-file", Name: gate.SourcePath, Field: "featureGates.RotateKubeletServerCertificate", Observed: gate.Value, Expected: "true"})
		if setting.Value == "true" && gate.Value == "true" {
			r.Status = "PASS"
		} else if (setting.Value == "false" || setting.Value == "true") && (gate.Value == "false" || gate.Value == "true") {
			r.Status = "FAIL"
		}
	default:
		r.Status = "FAIL"
		if setting.Value == expected {
			r.Status = "PASS"
		}
	}
	r.Details = fmt.Sprintf("Assessed kubelet field %s on node %s using explicit node metadata; missing fields do not imply secure defaults", field, nodeName)
	if r.Status == "MANUAL" {
		r.Details += "; workload policy and certificate usage require human verification"
	}
	return r
}

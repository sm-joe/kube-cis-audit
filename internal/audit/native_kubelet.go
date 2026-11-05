package audit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// evaluateKubeletConfigField inspects a single allowlisted kubelet configuration
// setting. Missing, unreadable, symlinked and ambiguous files are NOT_ASSESSED.
// No commands or scripts from benchmark data are ever executed.
func evaluateKubeletConfigField(hostRoot, nodeName string, c Check) Result {
	r := Result{ID: c.ID, Title: c.Title, Target: c.Target, Status: "NOT_ASSESSED", Method: c.Kind, Scored: c.Scored, Remediation: c.Remediation}
	const allowedPath = "var/lib/kubelet/config.yaml"
	if runtime.GOOS != "linux" {
		r.Details = "Native kubelet host configuration inspection requires a Linux runtime"
		return r
	}
	if c.Path != allowedPath || (c.Field != "authentication.anonymous.enabled" && c.Field != "authorization.mode") {
		r.Details = "Unapproved kubelet configuration file or field"
		return r
	}
	if hostRoot == "" {
		r.Details = "Host root was not provided"
		return r
	}
	root, err := filepath.Abs(hostRoot)
	if err != nil {
		r.Details = err.Error()
		return r
	}
	// Reject symlinks in every component so parent-directory symlinks cannot
	// escape the intended read-only mount and disclose unrelated host files.
	current := root
	for _, part := range strings.Split(allowedPath, "/") {
		current = filepath.Join(current, part)
		st, e := os.Lstat(current)
		if e != nil {
			r.Details = fmt.Sprintf("Host configuration not accessible: %v", e)
			return r
		}
		if st.Mode()&os.ModeSymlink != 0 {
			r.Details = "Symlink in host configuration path rejected"
			return r
		}
	}
	f, err := os.Open(current)
	if err != nil {
		r.Details = fmt.Sprintf("Cannot read kubelet configuration: %v", err)
		return r
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() > 1024*1024 {
		r.Details = "Kubelet configuration is not a regular file or exceeds 1 MiB"
		return r
	}
	data, err := io.ReadAll(io.LimitReader(f, 1024*1024+1))
	if err != nil || len(data) > 1024*1024 {
		r.Details = "Cannot read bounded kubelet configuration"
		return r
	}
	var rootCfg map[string]interface{}
	if err := yaml.Unmarshal(data, &rootCfg); err != nil {
		r.Details = "Invalid kubelet YAML configuration: " + err.Error()
		return r
	}
	var observed, expected string
	switch c.Field {
	case "authentication.anonymous.enabled":
		auth, ok := rootCfg["authentication"].(map[string]interface{})
		if !ok {
			r.Details = "Kubelet authentication setting not explicit"
			return r
		}
		anonymous, ok := auth["anonymous"].(map[string]interface{})
		if !ok {
			r.Details = "Kubelet anonymous authentication setting not explicit"
			return r
		}
		enabled, ok := anonymous["enabled"].(bool)
		if !ok {
			r.Details = "Kubelet anonymous authentication is not an explicit boolean"
			return r
		}
		observed, expected = fmt.Sprint(enabled), "false"
	case "authorization.mode":
		authorization, ok := rootCfg["authorization"].(map[string]interface{})
		if !ok {
			r.Details = "Kubelet authorization setting not explicit"
			return r
		}
		mode, ok := authorization["mode"].(string)
		if !ok || strings.TrimSpace(mode) == "" {
			r.Details = "Kubelet authorization mode not explicit"
			return r
		}
		observed, expected = mode, "Webhook"
	}
	if observed == expected {
		r.Status = "PASS"
	} else {
		r.Status = "FAIL"
	}
	r.Details = fmt.Sprintf("Evaluated kubelet configuration field %s on node %s", c.Field, nodeName)
	r.Evidence = []Evidence{{Resource: "host-file", Name: c.Path, Field: c.Field, Observed: observed, Expected: expected}}
	return r
}

package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// KubeletClientAuthentication captures only the active kubeconfig's client
// authentication mechanism. No token, certificate, key, command arguments,
// or kubeconfig file content is ever exported in a metadata snapshot.
type KubeletClientAuthentication struct {
	SourcePath string `json:"sourcePath"`
	Mode       string `json:"mode"`
	Error      string `json:"error,omitempty"`
}

const (
	KubeletIAMExecToken = "IAM_EXEC_TOKEN"
	KubeletClientCert   = "CLIENT_CERT"
	KubeletAuthUnknown  = "UNKNOWN"
)

// collectKubeletClientAuthentication only reads the explicitly mounted
// /var/lib/kubelet/kubeconfig file, not a bootstrap kubeconfig or arbitrary
// kubelet arguments. The active kubeconfig context, not unused users, controls
// applicability of CIS EKS 3.2.7.
func collectKubeletClientAuthentication(hostRoot string) KubeletClientAuthentication {
	out := KubeletClientAuthentication{SourcePath: "var/lib/kubelet/kubeconfig", Mode: KubeletAuthUnknown}
	file := filepath.Join(hostRoot, "kubeconfig")
	info, err := os.Lstat(file)
	if err != nil {
		out.Error = "active kubelet kubeconfig not available"
		return out
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		out.Error = "kubeconfig must be a regular file, not a symlink"
		return out
	}
	if info.Size() <= 0 || info.Size() > 256*1024 {
		out.Error = "kubeconfig size outside bounded range"
		return out
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		out.Error = "kubeconfig unreadable"
		return out
	}
	var config struct {
		CurrentContext string `yaml:"current-context"`
		Contexts       []struct {
			Name    string `yaml:"name"`
			Context struct {
				User string `yaml:"user"`
			} `yaml:"context"`
		} `yaml:"contexts"`
		Users []struct {
			Name string `yaml:"name"`
			User struct {
				Exec *struct {
					Command string   `yaml:"command"`
					Args    []string `yaml:"args"`
				} `yaml:"exec"`
				ClientCertificate     string    `yaml:"client-certificate"`
				ClientCertificateData string    `yaml:"client-certificate-data"`
				ClientKey             string    `yaml:"client-key"`
				ClientKeyData         string    `yaml:"client-key-data"`
				Token                 string    `yaml:"token"`
				TokenFile             string    `yaml:"tokenFile"`
				Username              string    `yaml:"username"`
				Password              string    `yaml:"password"`
				AuthProvider          *struct{} `yaml:"auth-provider"`
			} `yaml:"user"`
		} `yaml:"users"`
	}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		out.Error = "kubeconfig format invalid"
		return out
	}
	userName := ""
	contextMatches := 0
	for _, ctx := range config.Contexts {
		if ctx.Name == config.CurrentContext {
			contextMatches++
			userName = ctx.Context.User
		}
	}
	if config.CurrentContext == "" || userName == "" || contextMatches != 1 {
		out.Error = "active kubeconfig context/user not identified"
		return out
	}
	matches := 0
	for _, user := range config.Users {
		if user.Name != userName {
			continue
		}
		matches++
		cert := user.User.ClientCertificate != "" || user.User.ClientCertificateData != ""
		key := user.User.ClientKey != "" || user.User.ClientKeyData != ""
		exec := user.User.Exec
		other := user.User.Token != "" || user.User.TokenFile != "" || user.User.Username != "" || user.User.Password != "" || user.User.AuthProvider != nil
		if exec != nil && !cert && !key && !other {
			cmd := filepath.Base(strings.TrimSpace(exec.Command))
			if cmd == "aws" && len(exec.Args) >= 2 && exec.Args[0] == "eks" && exec.Args[1] == "get-token" {
				out.Mode = KubeletIAMExecToken
			}
			if cmd == "aws-iam-authenticator" && len(exec.Args) > 0 && exec.Args[0] == "token" {
				out.Mode = KubeletIAMExecToken
			}
		} else if exec == nil && cert && key && !other {
			out.Mode = KubeletClientCert
		}
	}
	if matches != 1 {
		out.Mode = KubeletAuthUnknown
		out.Error = "active kubeconfig user missing or duplicated"
	} else if out.Mode == KubeletAuthUnknown {
		out.Error = fmt.Sprintf("active kubelet authentication mechanism is unsupported or ambiguous")
	}
	return out
}

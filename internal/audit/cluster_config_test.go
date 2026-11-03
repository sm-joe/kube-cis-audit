package audit

import (
	"path/filepath"
	"testing"

	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestClusterConfigHonorsExplicitContext(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")

	config := clientcmdapi.NewConfig()
	config.Clusters["first-cluster"] = &clientcmdapi.Cluster{Server: "https://first.example"}
	config.Clusters["second-cluster"] = &clientcmdapi.Cluster{Server: "https://second.example"}
	config.AuthInfos["first-user"] = &clientcmdapi.AuthInfo{Token: "first-token"}
	config.AuthInfos["second-user"] = &clientcmdapi.AuthInfo{Token: "second-token"}
	config.Contexts["first"] = &clientcmdapi.Context{Cluster: "first-cluster", AuthInfo: "first-user"}
	config.Contexts["second"] = &clientcmdapi.Context{Cluster: "second-cluster", AuthInfo: "second-user"}
	config.CurrentContext = "first"

	path := filepath.Join(t.TempDir(), "config")
	if err := clientcmd.WriteToFile(*config, path); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}

	got, err := clusterConfig(path, "second")
	if err != nil {
		t.Fatalf("clusterConfig returned error: %v", err)
	}
	if got.Host != "https://second.example" {
		t.Fatalf("expected second context host, got %q", got.Host)
	}
}

func TestClusterConfigRejectsContextForInClusterCredentials(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
	t.Setenv("KUBERNETES_SERVICE_PORT", "443")

	_, err := clusterConfig("", "workstation")
	if err == nil || err.Error() != "--context cannot be used with in-cluster credentials" {
		t.Fatalf("expected in-cluster context error, got %v", err)
	}
}

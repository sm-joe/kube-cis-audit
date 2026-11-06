package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EKSProviderEvidence is a narrowly scoped, redacted AWS DescribeCluster
// snapshot. It contains configuration inventory, NOT a CIS pass/fail score.
type EKSProviderEvidence struct {
	SchemaVersion       int               `json:"schemaVersion"`
	CloudProvider       string            `json:"cloudProvider"`
	ClusterName         string            `json:"clusterName"`
	Region              string            `json:"region"`
	KubernetesVersion   string            `json:"kubernetesVersion"`
	GeneratedAt         time.Time         `json:"generatedAt"`
	EndpointPrivate     bool              `json:"endpointPrivateAccess"`
	EndpointPublic      bool              `json:"endpointPublicAccess"`
	PublicAccessCIDRs   []string          `json:"publicAccessCidrs,omitempty"`
	ControlPlaneLogging map[string]bool   `json:"controlPlaneLogging"`
	AssessmentStatus    string            `json:"assessmentStatus"`
	AssessmentNote      string            `json:"assessmentNote"`
}

func DecodeEKSProviderEvidence(data []byte, cluster Report, maxAge time.Duration, now time.Time) (EKSProviderEvidence, error) {
	var e EKSProviderEvidence
	if err := json.Unmarshal(data, &e); err != nil {
		return e, fmt.Errorf("decode EKS provider evidence: %w", err)
	}
	if cluster.CloudProvider != "eks" || e.CloudProvider != "eks" || e.SchemaVersion != 1 {
		return e, fmt.Errorf("provider evidence must match EKS cluster and schema v1")
	}
	if strings.TrimSpace(e.ClusterName) == "" || strings.TrimSpace(e.Region) == "" {
		return e, fmt.Errorf("provider evidence missing EKS cluster or region")
	}
	if e.GeneratedAt.IsZero() || e.GeneratedAt.After(now.Add(2*time.Minute)) ||
		now.Sub(e.GeneratedAt) > maxAge {
		return e, fmt.Errorf("EKS provider evidence is stale or future dated")
	}
	// Discovery gives the full EKS build suffix, DescribeCluster gives X.Y.
	// Compare exact major/minor components; never compare just a prefix.
	parts := strings.Split(strings.TrimPrefix(cluster.ClusterVersion, "v"), ".")
	if len(parts) < 2 || parts[0]+"."+parts[1] != e.KubernetesVersion {
		return e, fmt.Errorf("EKS provider evidence version %q does not match cluster %q", e.KubernetesVersion, cluster.ClusterVersion)
	}
	if e.AssessmentStatus != "INVENTORY_ONLY" || len(e.ControlPlaneLogging) == 0 {
		return e, fmt.Errorf("provider evidence must explicitly be unscored configuration inventory")
	}
	for _, logType := range []string{"api", "audit", "authenticator", "controllerManager", "scheduler"} {
		if _, ok := e.ControlPlaneLogging[logType]; !ok {
			return e, fmt.Errorf("EKS provider evidence missing logging field %q", logType)
		}
	}
	return e, nil
}

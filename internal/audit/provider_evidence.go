package audit

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// These optional, bounded read-only inventories allow native AWS controls to
// distinguish verified settings from absent permissions or missing evidence.
type EKSAccessEntryEvidence struct {
	PrincipalARN string   `json:"principalArn"`
	Type         string   `json:"type,omitempty"`
	Groups       []string `json:"kubernetesGroups,omitempty"`
	PolicyScopes []string `json:"policyScopes,omitempty"`
	Error        string   `json:"error,omitempty"`
}

type EKSECRImageScanEvidence struct {
	Registry                   string   `json:"registry"`
	Repository                 string   `json:"repository"`
	Digest                     string   `json:"digest,omitempty"`
	ScanType                   string   `json:"scanType,omitempty"`
	ScanStatus                 string   `json:"scanStatus,omitempty"`
	ScanDigests                []string `json:"scanDigests,omitempty"`
	AttestationScanner         string   `json:"attestationScanner,omitempty"`
	AttestationSignerKeySHA256 string   `json:"attestationSignerKeySha256,omitempty"`
	AttestationScannedAt       string   `json:"attestationScannedAt,omitempty"`
	Error                      string   `json:"error,omitempty"`
}

// EKSAWSManualEvidence is a bounded, read-only human-review inventory.
// It cannot be used to turn MANUAL into PASS without evidence-bound approval.
type EKSAWSManualEvidence struct {
	Status string            `json:"status"`
	Note   string            `json:"note"`
	Items  []json.RawMessage `json:"items"`
}

// EKSProviderEvidence is a narrowly scoped, redacted AWS DescribeCluster
// snapshot. It contains configuration inventory, NOT a CIS pass/fail score.
type EKSProviderEvidence struct {
	SchemaVersion         int                       `json:"schemaVersion"`
	CloudProvider         string                    `json:"cloudProvider"`
	ClusterName           string                    `json:"clusterName"`
	ClusterARN            string                    `json:"clusterArn"`
	Region                string                    `json:"region"`
	KubernetesVersion     string                    `json:"kubernetesVersion"`
	GeneratedAt           time.Time                 `json:"generatedAt"`
	EndpointPrivate       bool                      `json:"endpointPrivateAccess"`
	EndpointPublic        bool                      `json:"endpointPublicAccess"`
	PublicAccessCIDRs     []string                  `json:"publicAccessCidrs,omitempty"`
	ControlPlaneLogging   map[string]bool           `json:"controlPlaneLogging"`
	AssessmentStatus      string                    `json:"assessmentStatus"`
	AssessmentNote        string                    `json:"assessmentNote"`
	AuthenticationMode    string                    `json:"authenticationMode,omitempty"`
	AccessInventoryStatus string                    `json:"accessInventoryStatus,omitempty"`
	AccessEntries         []EKSAccessEntryEvidence  `json:"accessEntries,omitempty"`
	VpcCNIStatus          string                    `json:"vpcCniStatus,omitempty"`
	VpcCNIVersion         string                    `json:"vpcCniVersion,omitempty"`
	VpcCNINetworkPolicy   *bool                     `json:"vpcCniNetworkPolicy,omitempty"`
	ECRInventoryStatus    string                    `json:"ecrInventoryStatus,omitempty"`
	ECRImages             []EKSECRImageScanEvidence `json:"ecrImages,omitempty"`
	NonECRImageCount      int                       `json:"nonEcrImageCount,omitempty"`
	AWSManualEvidence map[string]EKSAWSManualEvidence `json:"awsManualEvidence,omitempty"`
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
	arnPattern := regexp.MustCompile("^arn:aws[a-z-]*:eks:([a-z0-9-]+):([0-9]{12}):cluster/([A-Za-z0-9_-]+)$")
	arnParts := arnPattern.FindStringSubmatch(e.ClusterARN)
	if len(arnParts) != 4 || arnParts[1] != e.Region || arnParts[3] != e.ClusterName {
		return e, fmt.Errorf("EKS provider ARN does not match cluster name and region")
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
	for id, review := range e.AWSManualEvidence {
		switch id {
		case "5.1.2", "5.1.3", "5.3.1", "5.4.3":
		default:
			return e, fmt.Errorf("unrecognized AWS manual-review evidence control %q", id)
		}
		if review.Status != "COLLECTED" && review.Status != "PARTIAL" && review.Status != "UNAVAILABLE" {
			return e, fmt.Errorf("invalid AWS review inventory status for %s", id)
		}
		if len(review.Items) > 40 || len(review.Note) > 1024 {
			return e, fmt.Errorf("oversized AWS manual-review evidence for %s", id)
		}
		for _, item := range review.Items {
			if len(item) > 4096 || !json.Valid(item) {
				return e, fmt.Errorf("invalid AWS manual-review item for %s", id)
			}
		}
	}
	return e, nil
}

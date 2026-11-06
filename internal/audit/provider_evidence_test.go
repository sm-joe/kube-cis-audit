package audit

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDecodeEKSProviderEvidenceRejectsMismatchedAndStaleData(t *testing.T) {
	now := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	cluster := Report{CloudProvider: "eks", ClusterVersion: "v1.31.14-eks-577a51e"}
	candidate := EKSProviderEvidence{
		SchemaVersion: 1, CloudProvider: "eks",
		ClusterName: "kubecis-eks-test", Region: "ap-south-1", ClusterARN: "arn:aws:eks:ap-south-1:123456789012:cluster/kubecis-eks-test",
		KubernetesVersion: "1.31", GeneratedAt: now.Add(-time.Minute),
		EndpointPrivate: true, EndpointPublic: false,
		AssessmentStatus: "INVENTORY_ONLY",
		ControlPlaneLogging: map[string]bool{
			"api": true, "audit": false, "authenticator": true,
			"controllerManager": false, "scheduler": false,
		},
	}
	decode := func(e EKSProviderEvidence) error {
		t.Helper()
		data, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeEKSProviderEvidence(data, cluster, 30*time.Minute, now)
		if err == nil && (got.AssessmentStatus != "INVENTORY_ONLY" || got.ClusterName != candidate.ClusterName) {
			t.Fatal("source data not preserved")
		}
		return err
	}
	if err := decode(candidate); err != nil {
		t.Fatalf("valid EKS provider evidence rejected: %v", err)
	}
	bad := candidate
	bad.KubernetesVersion = "1.3"
	if err := decode(bad); err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("invalid Kubernetes version accepted: %v", err)
	}
	bad = candidate
	bad.GeneratedAt = now.Add(-2 * time.Hour)
	if err := decode(bad); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale EKS provider evidence accepted: %v", err)
	}
	bad = candidate
	bad.AssessmentStatus = "PASS"
	if err := decode(bad); err == nil {
		t.Fatal("a provider configuration inventory was treated as compliance PASS")
	}
	bad = candidate
	bad.ControlPlaneLogging = map[string]bool{"api": true}
	if err := decode(bad); err == nil {
		t.Fatal("incomplete provider logging evidence accepted")
	}
	bad = candidate
	bad.ClusterARN = "arn:aws:eks:us-east-1:123456789012:cluster/kubecis-eks-test"
	if err := decode(bad); err == nil {
		t.Fatal("ARN region mismatch accepted")
	}
	bad = candidate
	bad.ClusterARN = "arn:aws:eks:ap-south-1:123456789012:cluster/other-cluster"
	if err := decode(bad); err == nil {
		t.Fatal("wrong EKS ARN cluster accepted")
	}
	bad = candidate
	bad.CloudProvider = "unknown"
	if err := decode(bad); err == nil {
		t.Fatal("wrong cloud provider evidence accepted")
	}
}

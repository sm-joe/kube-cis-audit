package audit

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Signature verification happens again inside Go, not only in the Bash
// collector. A fabricated provider snapshot must never produce a CIS PASS.
func TestEKSExternalScanCannotPassWithoutPinnedValidSignature(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	path := filepath.Join(t.TempDir(), "approved-scanner.pem")
	if err := os.WriteFile(path, pub, 0600); err != nil {
		t.Fatal(err)
	}
	fp := sha256.Sum256(pub)
	pin := hex.EncodeToString(fp[:])
	t.Setenv("KUBECIS_ATTESTATION_PUBLIC_KEY", path)
	t.Setenv("KUBECIS_ATTESTATION_KEY_SHA256", pin)
	now := time.Now().UTC()
	digest := "sha256:" + strings.Repeat("a", 64)
	child := "sha256:" + strings.Repeat("b", 64)
	item := EKSECRImageScanEvidence{
		Registry: "602401143452.dkr.ecr.us-east-1.amazonaws.com",
		Repository: "eks/coredns", Digest: digest, ScanType: "THIRD_PARTY",
		ScanStatus: "COMPLETE", ScanDigests: []string{child},
		AttestationScanner: "approved-vulnerability-scanner",
		AttestationSignerKeySHA256: pin,
		AttestationScannedAt: now.Format(time.RFC3339),
	}
	var claim externalScanClaim
	claim.SchemaVersion = 1
	claim.Subject.Registry = item.Registry
	claim.Subject.Repository = item.Repository
	claim.Subject.Digest = digest
	claim.Scanner = item.AttestationScanner
	claim.ScanID = "verified-run-101"
	claim.Status = "COMPLETE"
	claim.FindingsAvailable = true
	claim.ScannerReportSHA256 = strings.Repeat("c", 64)
	claim.ScannedAt = item.AttestationScannedAt
	claim.ScannedDigests = []string{child}
	data, err := json.Marshal(claim)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		t.Fatal(err)
	}
	item.AttestationClaimBase64 = base64.StdEncoding.EncodeToString(data)
	item.AttestationSignatureBase64 = base64.StdEncoding.EncodeToString(signature)
	if err := verifyExternalScanEvidence(item, now); err != nil {
		t.Fatalf("trusted scanned manifest rejected: %v", err)
	}
	result := EvaluateEKSAdditionalAWSControls(EKSProviderEvidence{
		ECRInventoryStatus: "COLLECTED", ECRImages: []EKSECRImageScanEvidence{item},
	})
	if len(result) != 1 || result[0].Status != "PASS" {
		t.Fatalf("verified signed scan should allow coverage PASS: %+v", result)
	}
	original := item
	item.AttestationClaimBase64 = base64.StdEncoding.EncodeToString(append(data, 'x'))
	if err := verifyExternalScanEvidence(item, now); err == nil {
		t.Fatal("tampered signed claim accepted")
	}
	item = original
	item.Repository = "eks/kube-proxy"
	if err := verifyExternalScanEvidence(item, now); err == nil {
		t.Fatal("replayed claim for a different image accepted")
	}
	item = original
	item.ScanDigests = []string{"sha256:" + strings.Repeat("d", 64)}
	if err := verifyExternalScanEvidence(item, now); err == nil {
		t.Fatal("altered scanned-child digest accepted")
	}
	item = original
	t.Setenv("KUBECIS_ATTESTATION_KEY_SHA256", strings.Repeat("0", 64))
	result = EvaluateEKSAdditionalAWSControls(EKSProviderEvidence{
		ECRInventoryStatus: "COLLECTED", ECRImages: []EKSECRImageScanEvidence{item},
	})
	if result[0].Status != "MANUAL" {
		t.Fatalf("untrusted signing key must never PASS: %+v", result)
	}
}

func TestEKSECRUnknownScanningModeNeverPasses(t *testing.T) {
	base := EKSECRImageScanEvidence{
		Registry: "123456789012.dkr.ecr.us-east-1.amazonaws.com",
		Repository: "kubecis", Digest: "sha256:" + strings.Repeat("1",64),
		ScanStatus: "COMPLETE",
	}
	for _, mode := range []string{"", "UNKNOWN", "THIRD_PARTY"} {
		base.ScanType = mode
		result := EvaluateEKSAdditionalAWSControls(EKSProviderEvidence{
			ECRInventoryStatus: "COLLECTED", ECRImages: []EKSECRImageScanEvidence{base},
		})
		if result[0].Status != "MANUAL" {
			t.Fatalf("unverified scan mode %q was accepted: %+v", mode, result)
		}
	}
}

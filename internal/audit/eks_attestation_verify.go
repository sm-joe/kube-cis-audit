package audit

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// This is an independent second verification boundary. The shell collector
// checks both the signature and raw scanner report digest. The Go evaluator
// must not accept a hand-edited provider JSON asserting THIRD_PARTY/COMPLETE.
var scanDigestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
var sha256HexPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type externalScanClaim struct {
	SchemaVersion int `json:"schemaVersion"`
	Subject struct {
		Registry string `json:"registry"`
		Repository string `json:"repository"`
		Digest string `json:"digest"`
	} `json:"subject"`
	Scanner string `json:"scanner"`
	ScanID string `json:"scanId"`
	Status string `json:"status"`
	FindingsAvailable bool `json:"findingsAvailable"`
	ScannerReportSHA256 string `json:"scannerReportSha256"`
	ScannedAt string `json:"scannedAt"`
	ScannedDigests []string `json:"scannedDigests"`
}

func verifyExternalScanEvidence(item EKSECRImageScanEvidence, now time.Time) error {
	if item.ScanType != "THIRD_PARTY" || item.ScanStatus != "COMPLETE" || item.Error != "" {
		return fmt.Errorf("external scan is not a complete, error-free claim")
	}
	if !scanDigestPattern.MatchString(item.Digest) || len(item.ScanDigests) < 1 || len(item.ScanDigests) > 8 {
		return fmt.Errorf("invalid deployed image digest or platform manifest evidence")
	}
	keyPath := os.Getenv("KUBECIS_ATTESTATION_PUBLIC_KEY")
	keyPin := os.Getenv("KUBECIS_ATTESTATION_KEY_SHA256")
	if keyPath == "" || !sha256HexPattern.MatchString(keyPin) {
		return fmt.Errorf("operator-pinned external scanner public key is unavailable")
	}
	info, err := os.Lstat(keyPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > 8192 {
		return fmt.Errorf("trusted external scanner public key must be a bounded regular file")
	}
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("trusted external scanner public key could not be read")
	}
	keyHash := sha256.Sum256(keyBytes)
	if hex.EncodeToString(keyHash[:]) != keyPin || item.AttestationSignerKeySHA256 != keyPin {
		return fmt.Errorf("external scanner trust-anchor fingerprint mismatch")
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return fmt.Errorf("external scanner public key must be PKIX PEM")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("invalid external scanner public key")
	}
	rsaKey, ok := pubKey.(*rsa.PublicKey)
	if !ok || rsaKey.N.BitLen() < 2048 {
		return fmt.Errorf("external scanner public key must be at least RSA-2048")
	}
	payload, err := base64.StdEncoding.DecodeString(item.AttestationClaimBase64)
	if err != nil || len(payload) == 0 || len(payload) > 65536 {
		return fmt.Errorf("external scanner signed claim unavailable or too large")
	}
	sig, err := base64.StdEncoding.DecodeString(item.AttestationSignatureBase64)
	if err != nil || len(sig) == 0 || len(sig) > 8192 {
		return fmt.Errorf("external scanner signature unavailable or invalid")
	}
	hash := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, hash[:], sig); err != nil {
		return fmt.Errorf("external scan signature does not verify against approved key")
	}
	var claim externalScanClaim
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claim); err != nil {
		return fmt.Errorf("signed scan claim invalid or contains unsupported fields")
	}
	if claim.SchemaVersion != 1 || claim.Subject.Registry != item.Registry ||
		claim.Subject.Repository != item.Repository || claim.Subject.Digest != item.Digest ||
		claim.Scanner != item.AttestationScanner || strings.TrimSpace(claim.ScanID) == "" ||
		claim.Status != "COMPLETE" || !claim.FindingsAvailable ||
		!sha256HexPattern.MatchString(claim.ScannerReportSHA256) ||
		len(claim.ScannedDigests) != len(item.ScanDigests) {
		return fmt.Errorf("signed scan claim subject, source or findings provenance mismatch")
	}
	timestamp, err := time.Parse(time.RFC3339, claim.ScannedAt)
	if err != nil || claim.ScannedAt != item.AttestationScannedAt ||
		timestamp.After(now.Add(2*time.Minute)) || timestamp.Before(now.Add(-30*24*time.Hour)) {
		return fmt.Errorf("signed scan timestamp invalid, stale or changed")
	}
	seen := make(map[string]bool, len(item.ScanDigests))
	for i, d := range claim.ScannedDigests {
		if !scanDigestPattern.MatchString(d) || seen[d] || item.ScanDigests[i] != d {
			return fmt.Errorf("signed platform digest scope does not match reported scan")
		}
		seen[d] = true
	}
	return nil
}

package audit

import "fmt"

// ApplyEKSAWSManualEvidence adds read-only AWS reviewer observations to the
// corresponding manual CIS EKS findings. The inventory never changes PASS,
// FAIL, MANUAL, or NOT_ASSESSED, and cannot satisfy the official review gate.
func ApplyEKSAWSManualEvidence(out *ConsolidatedReport, e EKSProviderEvidence) {
	if out == nil {
		return
	}
	for i := range out.Results {
		r := &out.Results[i]
		if r.Target != "policies" {
			continue
		}
		review, ok := e.AWSManualEvidence[r.ID]
		if !ok {
			continue
		}
		r.Evidence = append(r.Evidence, Evidence{
			Resource: "AWS::EKS::ManualReviewInventory",
			Name:     e.ClusterARN,
			Field:    "reviewInventoryStatus",
			Observed: fmt.Sprintf("%s; %d bounded records", review.Status, len(review.Items)),
			Expected: "complete, provenance-bound inventory and independent human decision",
		})
		r.Details += fmt.Sprintf(" AWS manual evidence: %s; %d bounded records. %s", review.Status, len(review.Items), review.Note)
		if r.ID == "5.4.3" {
			// A manual benchmark recommendation can have an objective
			// native assessment, but an approved exception remains a
			// separate human decision and never rewrites a FAIL.
			r.Status, r.Details, r.Evidence = assessEKSPrivateNodes(review)
		}
	}
}

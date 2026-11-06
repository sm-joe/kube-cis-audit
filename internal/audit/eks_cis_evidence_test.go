package audit

import "testing"

func TestEvaluateEKSCISProviderControls(t *testing.T) {
    e := EKSProviderEvidence{
        SchemaVersion: 1, CloudProvider: "eks", ClusterARN: "arn:aws:eks:us-east-1:123456789012:cluster/test",
        AssessmentStatus: "INVENTORY_ONLY", EndpointPrivate: true, EndpointPublic: false,
        ControlPlaneLogging: map[string]bool{
            "api": true, "audit": true, "authenticator": true,
            "controllerManager": true, "scheduler": true,
        },
    }
    evaluate := func() []Result {
        t.Helper()
        rs, err := EvaluateEKSCISProviderControls(e)
        if err != nil { t.Fatal(err) }
        if len(rs) != 3 { t.Fatalf("expected 3 EKS controls, got %d", len(rs)) }
        for _, r := range rs {
            if !r.Scored || r.Method != "aws_eks_describe_cluster" || len(r.Evidence) < 2 {
                t.Fatalf("control result lacks official evidence: %+v", r)
            }
        }
        return rs
    }
    r := evaluate()
    for _, x := range r {
        if x.Status != "PASS" { t.Fatalf("%s should PASS: %+v", x.ID, x) }
    }
    e.ControlPlaneLogging["audit"] = false
    r = evaluate()
    if r[0].Status != "FAIL" || r[1].Status != "PASS" || r[2].Status != "PASS" {
        t.Fatalf("disabled audit log must fail only 2.1.1: %+v", r)
    }
    e.EndpointPublic = true
    e.PublicAccessCIDRs = []string{"0.0.0.0/0"}
    r = evaluate()
    if r[1].Status != "FAIL" || r[2].Status != "FAIL" {
        t.Fatalf("world-open endpoint must fail both exposure checks: %+v", r)
    }
    e.PublicAccessCIDRs = []string{"203.0.113.5/32"}
    r = evaluate()
    if r[1].Status != "MANUAL" || r[2].Status != "FAIL" {
        t.Fatalf("unverified approved CIDR should require review; private-only still fails: %+v", r)
    }
    e.EndpointPublic = false
    e.EndpointPrivate = false
    e.PublicAccessCIDRs = nil
    r = evaluate()
    if r[1].Status != "FAIL" || r[2].Status != "FAIL" {
        t.Fatalf("private endpoint disabled must fail: %+v", r)
    }
    e.ControlPlaneLogging = map[string]bool{"api": true}
    if _, err := EvaluateEKSCISProviderControls(e); err == nil {
        t.Fatal("incomplete control-plane logging inventory must fail closed")
    }
}

## CIS EKS 2.1.0 — Control 5.4.3 Worker Node Subnet Assessment

The official benchmark's 5.4.3 recommendation is **Manual**, and its audit procedure requires effective node subnet routing, public IPv4 behavior, and actual worker-node placement. The cluster API endpoint exposure (5.4.1/5.4.2) is unrelated. The source is the operator-held CIS Amazon EKS 2.1.0 PDF, printed pages 215–217.

### Collection and decision

The Linux EKS collector now enriches `awsManualEvidence["5.4.3"]` with:

- Every managed node-group subnet, `MapPublicIpOnLaunch`, VPC ID, and its **effective EC2 route table**. Explicit subnet association is used first; without it, the VPC main route table is resolved.
- Both IPv4 and IPv6 routes, including routes via `igw-*` (a direct Internet Gateway). NAT and egress-only IPv6 gateways are distinct from a directly attached IGW.
- Registered Kubernetes Nodes and their EC2 `providerID` references. EC2 `DescribeInstances` gives the actual subnet and public IPv4 assignment for each; unrecognized or unavailable providers cannot support a PASS.
- No internet routes, credentials, or AWS infrastructure are modified.

The evaluator reports **FAIL** when a verified worker subnet has an active route to an IGW or an observed worker instance has a public IPv4. `MapPublicIpOnLaunch=true` alone does not prove active public placement: absent a confirmed IGW path or observed public interface, it requires **MANUAL** investigation, never an automatic PASS. Explicitly approved public-node exceptions are kept as independent human review, never as a scan PASS. **PASS** requires a complete managed subnet and live node inventory, every worker node mapped to inspected private subnets, no IGW routes, no public-on-launch, and no public node IPv4 addresses. Missing AWS permissions, network snapshots, external node-provider types or unmatched node subnets remain **MANUAL**. `NOT_ASSESSED` is reserved for invalid/missing fundamental collector evidence elsewhere.

The benchmark does not require changing an intentionally public **test** cluster to run the scanner. It requires honest reporting of the deviation. A production remediation is to migrate worker nodes to a replacement managed node group in private subnets and provide NAT or necessary VPC endpoints.

### Read-only AWS permissions

The collector uses `ec2:DescribeSubnets`, `ec2:DescribeRouteTables`, `ec2:DescribeInstances`, `eks:ListNodegroups`, `eks:DescribeNodegroup` and Kubernetes `list nodes`. All queries are read-only. An unauthorized call remains reviewable without inventing a PASS.

### Verify

```bash
git pull origin main
bash scripts/test-eks-node-network.sh
go test ./internal/audit -run 'TestEKS543|TestEKSAWSManualEvidence' -count=1
bash scripts/collect-eks.sh --context "$EKS_CONTEXT" --cluster-name demo-eks \
  --region us-east-1 --binary ./kubecis --output-dir reports/eks
jq '.awsManualEvidence["5.4.3"]' reports/eks/provider.json
```

Because the CIS control remains a manual recommendation, no official benchmark score is released and approved exceptions remain separate evidence-bound decisions.

### Intentional public test clusters: record an exception without changing FAIL

After collecting a consolidated report, use the existing CLI to generate an
immutable-digest review template. A `5.4.3` FAIL appears as an **optional
EXCEPTION-only** candidate; it is never eligible for `APPROVED` or `REJECTED`
as an automated PASS-equivalent decision.

```bash
./kubecis review-init --input reports/eks/consolidated.json \
  --output-file reports/eks/review-template.json
```

Use the actual consolidated-report path produced by the collection workflow.
Create a decision with the exact `reportSha256`, finding ID and scope, the
literal `decision: "EXCEPTION"`, a named reviewer, approved justification,
an independently resolvable `evidenceRef`, and a real timestamp. Then run:

```bash
./kubecis finalize --input reports/eks/consolidated.json \
  --decisions reports/eks/approved-decisions.json \
  --output-dir reports/eks/assessment
```

The final dossier retains `automatedStatus: FAIL` for 5.4.3 and records the
exception separately. The official compliance score remains WITHHELD.
A review template is not itself an approval and no exception is auto-created.


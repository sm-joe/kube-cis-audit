# EKS release evidence and acceptance contract

## Product mode: EKS native preview

This release is a **working native scanner**, not a complete verified CIS Amazon
EKS Benchmark v2.1.0 implementation. The existing `cis-2.1.0` directory is a
provisional starter profile, **not a transcription of the full official EKS
benchmark**. An output percentage from five scored starter-profile controls is
not the official EKS CIS benchmark coverage percentage.

Official benchmark reference:
https://www.cisecurity.org/benchmark/kubernetes

## What the end-to-end acceptance gate proves

- A known kubeconfig context points to an actual EKS API endpoint.
- AWS `eks:DescribeCluster` for a named cluster returns **the same HTTPS API
  endpoint**, with the same ARN region/name and a valid AWS account ID.
- The Kubernetes API reports an EKS distribution and its major.minor version
  matches the AWS DescribeCluster version.
- At least one supported Linux managed EC2 worker is assessed; **every** node
  in the expected inventory has one fresh node-local source.
- AL2023 kubelet file metadata plus two allowlisted fields were read and
  assessed; raw credentials/configuration never enter the snapshot.
- The output contains one native Kubernetes API source, its complete required
  worker sources, and an unscored AWS provider configuration snapshot.
- Reports reject stale sources, incompatible build versions, duplicate nodes,
  incorrect provider identity and incomplete kubelet evidence.
- EKS security posture is reported even when logging is disabled or the API
  endpoint is deliberately public. Such settings may be intentional for a
  nonproduction test cluster and do **not** cause the scanner itself to fail.

The collector runs this preview gate automatically on its consolidated report.

## Two distinct release decisions

**Preview**: `Validate-EKSRelease.ps1 -ReleaseLevel Preview` checks the proven
operational data path and keeps a withheld compliance score.

**Full CIS EKS**: `Validate-EKSRelease.ps1 -ReleaseLevel Full` **fails closed**.
It cannot pass until a verified, complete official CIS EKS benchmark revision
and supported Kubernetes version mapping, control-level evidence, applicability
disposition, end-to-end tests and signed published release are in place.
Changing a flag, adding dummy PASS results or assuming that one verified EKS
cluster exercises every rule cannot satisfy this gate.

## Missing official CIS benchmark implementation

- Obtain the authentic CIS Amazon EKS v2.1.0 control catalogue and verify
  the documented supported Kubernetes versions.
- Map each control ID and audit procedure to an independently testable
  KubeCIS evidence collector, or record a documented `MANUAL`/`NOT_APPLICABLE`
  with justification. Do not infer compliance for provider-managed hosts.
- Create positive/negative/unknown fixtures for **each** control and validate
  them against upstream guidance; validate that unsupported checks reduce
  coverage rather than silently passing.
- Keep official EKS controls separate from the generic CIS Kubernetes benchmark
  and KubeCIS-native extensions.
- Run Linux/Windows, Kind and supported EKS versions end to end; publish
  reproducible builds, SBOM, provenance, and benchmark source/revision matrix.

## Run on a test EKS cluster

Use the extracted Windows release archive (the same commit as the DaemonSet).
Adjust context/name/region to your actual cluster rather than treating context
aliases as AWS cluster identities.

```powershell
kubectl config use-context my-eks-cluster
kubectl apply -k .\deploy\eks
kubectl -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=300s
.\scripts\Collect-ConsolidatedEKSReport.ps1 -Context my-eks-cluster -ClusterName my-eks-cluster -Region us-east-1 -Binary .\kubecis.exe
.\scripts\Validate-EKSRelease.ps1 -ExpectedClusterName my-eks-cluster -ExpectedRegion us-east-1 -ReleaseLevel Preview
```

Use the immutable published image **matching the exact archive commit** for
all DaemonSet containers before collecting evidence.

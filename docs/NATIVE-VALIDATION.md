# Native release validation — Windows PowerShell

All commands run from an extracted source directory. Do not publish as production until all stages pass.

## Build, test and report integrity

```powershell
$ErrorActionPreference = 'Stop'
go version
go mod download
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
New-Item -ItemType Directory -Force .\dist | Out-Null
go build -trimpath -o .\dist\kubecis.exe ./cmd/kubecis
.\dist\kubecis.exe version
kubectl config use-context kind-control-plane
.\dist\kubecis.exe scan --engine native --benchmark auto --target policies --output json --output-file .\dist\scan.json --exit-code-on-fail=false
```

Check the JSON `summary`, `results`, `findingId`, `compliance` and `benchmarkCompatibilityStatus`. Do not infer compliance from a provisional benchmark profile.

## Podman / Kind

```powershell
podman machine start
podman build -t localhost/kubecis-native:rc1 .
podman run --rm localhost/kubecis-native:rc1 version
kind get clusters
kind load docker-image localhost/kubecis-native:rc1 --name control-plane
kubectl config use-context kind-control-plane
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/cluster-scan-job-rc4.yaml
kubectl -n kube-system wait --for=condition=complete job/kubecis-cluster-scan-rc4 --timeout=600s
kubectl -n kube-system logs job/kubecis-cluster-scan-rc4
```

`kind load docker-image` may not import a Podman image if Kind's configured container runtime is different. On Podman-backed Kind, verify `kind get clusters` and runtime setup, and load/export the image using the documented Kind-compatible local image transport before applying the Job. The local cluster scan Job is pinned to `localhost/kubecis-native:rc1` with `imagePullPolicy: Never`, matching the Podman build command. A control-plane node selector requires a labeled control-plane node. Run the node auditors separately after reviewing their hostPath permissions and configuration. Never make an API-only workload privileged solely to get a passing test.

## Gate for release

- Full Go build/test/race/vet pass.
- Comparison and kubelet config regression tests pass.
- Podman image builds; no kube-bench executable in the image.
- Kind cluster scan produces parseable internally consistent JSON.
- Native Linux node audit reads actual allowlisted host config (direct mode), otherwise emits `NOT_ASSESSED`.
- All benchmark compatibility claims and licensing are independently reviewed.
- Root project license chosen and included by the owner.

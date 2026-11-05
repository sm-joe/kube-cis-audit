# One-shot, reproducible local build and Kind deployment for Windows PowerShell.
# Execute from the extracted repository root: .\scripts\Deploy-KindNative.ps1
[CmdletBinding()]
param(
    [string]$ClusterName = 'control-plane',
    [string]$Namespace = 'kube-system',
    [string]$Image = 'localhost/kubecis-native:integrated-rc3.2'
)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
Push-Location $repo
try {
    foreach ($name in @('podman','kind','kubectl')) {
        if (-not (Get-Command $name -ErrorAction SilentlyContinue)) { throw "Missing required command: $name" }
    }
    $env:KIND_EXPERIMENTAL_PROVIDER = 'podman'
    $clusters = @(kind get clusters)
    if ($LASTEXITCODE -ne 0 -or $clusters -notcontains $ClusterName) {
        throw "Kind cluster '$ClusterName' unavailable. Existing clusters: $($clusters -join ', ')"
    }
    Write-Host 'Building KubeCIS Native image with Podman...'
    & podman build --tag $Image --build-arg VERSION=3.1.0-native-integrated-rc3.2 .
    if ($LASTEXITCODE -ne 0) { throw 'Podman build failed; no Kubernetes changes made.' }
    & podman run --rm $Image version
    if ($LASTEXITCODE -ne 0) { throw 'Built image did not start; no Kubernetes changes made.' }
    $archive = Join-Path ([System.IO.Path]::GetTempPath()) 'kubecis-native-integrated.tar'
    try {
        & podman save --format docker-archive --output $archive $Image
        if ($LASTEXITCODE -ne 0) { throw 'Podman image export failed; no Kubernetes changes made.' }
        & kind load image-archive $archive --name $ClusterName
        if ($LASTEXITCODE -ne 0) { throw 'Kind image load failed; no Kubernetes changes made.' }
    } finally { Remove-Item $archive -Force -ErrorAction SilentlyContinue }
    $context = "kind-$ClusterName"
    & kubectl config use-context $context
    if ($LASTEXITCODE -ne 0) { throw "Kubernetes context '$context' unavailable" }
    & kubectl apply -f deploy/rbac.yaml
    if ($LASTEXITCODE -ne 0) { throw 'RBAC apply failed' }
    & kubectl apply -f deploy/kind-native/node-auditor-daemonset.yaml
    if ($LASTEXITCODE -ne 0) { throw 'Node DaemonSet apply failed' }
    & kubectl apply -f deploy/kind-native/controlplane-auditor-daemonset.yaml
    if ($LASTEXITCODE -ne 0) { throw 'Control-plane DaemonSet apply failed' }
    & kubectl -n $Namespace rollout status daemonset/kubecis-node-auditor --timeout=180s
    if ($LASTEXITCODE -ne 0) { throw 'Node DaemonSet rollout failed' }
    & kubectl -n $Namespace rollout status daemonset/kubecis-controlplane-auditor --timeout=180s
    if ($LASTEXITCODE -ne 0) { throw 'Control-plane DaemonSet rollout failed' }
    # Do not patch an immutable old Job template; recreate only our dedicated Job.
    & kubectl -n $Namespace delete job kubecis-native-cluster-scan --ignore-not-found=true --wait=true
    if ($LASTEXITCODE -ne 0) { throw 'Could not replace the dedicated native scan Job' }
    & kubectl apply -f deploy/kind-native/job.yaml
    if ($LASTEXITCODE -ne 0) { throw 'Native scan Job apply failed' }
    & kubectl -n $Namespace wait --for=condition=complete job/kubecis-native-cluster-scan --timeout=180s
    if ($LASTEXITCODE -ne 0) { Write-Warning 'Cluster scan Job did not complete; inspect its Pod logs.' }
    Write-Host '=== Cluster policy audit ==='
    & kubectl -n $Namespace logs job/kubecis-native-cluster-scan --tail=60
    Write-Host '=== Node audit ==='
    $nodePods = & kubectl -n $Namespace get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json | ConvertFrom-Json
    foreach ($pod in $nodePods.items) {
        Write-Host "Node pod: $($pod.metadata.name)"
        & kubectl -n $Namespace logs $pod.metadata.name -c node-auditor --tail=100
        & kubectl -n $Namespace logs $pod.metadata.name -c node-metadata-helper --tail=5
    }
    Write-Host '=== Control-plane audit ==='
    $masterPods = & kubectl -n $Namespace get pods -l app=kubecis-controlplane-auditor -o json | ConvertFrom-Json
    foreach ($pod in $masterPods.items) {
        Write-Host "Control-plane pod: $($pod.metadata.name)"
        & kubectl -n $Namespace logs $pod.metadata.name -c controlplane-auditor --tail=80
    }
    Write-Host 'Deployment and audit commands completed. Verify JSON results and host evidence status.'
} finally { Pop-Location }

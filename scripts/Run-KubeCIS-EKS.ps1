# One-command, version-matched EKS assessment for Windows PowerShell 5.1+.
# Deploys only KubeCIS's own DaemonSet, then collects read-only evidence.
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$ClusterName,
    [Parameter(Mandatory=$true)][string]$Region,
    [string]$OutputDir = 'reports\eks'
)
$ErrorActionPreference = 'Stop'
$homePath = Split-Path -Parent $PSScriptRoot
$original = (Get-Location).Path
try {
    Set-Location $homePath
    if (-not (Test-Path '.\kubecis.exe' -PathType Leaf)) { throw 'Missing matching Windows kubecis.exe' }
    if (-not (Test-Path '.\IMAGE.txt' -PathType Leaf)) { throw 'Missing matching container image reference' }
    if (-not (Test-Path '.\benchmarks\cis-eks-2.1.0\benchmark.yaml')) { throw 'Missing EKS benchmark profile' }

    $image = (Get-Content '.\IMAGE.txt' -Raw).Trim()
    if ($image -notmatch '^ghcr\.io/sm-joe/kubecis:cis210-eks135-([0-9a-f]{40})$') {
        throw "Unexpected image reference in IMAGE.txt: $image"
    }
    $expectedVersion = "3.1.0-native-cis210-eks135-$($Matches[1])"
    $windowsVersion = (& .\kubecis.exe version).Trim()
    if ($LASTEXITCODE -ne 0 -or $windowsVersion -cne $expectedVersion) {
        throw "Windows version mismatch: expected $expectedVersion, got $windowsVersion"
    }
    & aws eks update-kubeconfig --region $Region --name $ClusterName | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Could not authenticate to the requested EKS cluster' }
    $context = (& kubectl config current-context).Trim()
    if ($LASTEXITCODE -ne 0 -or -not $context) { throw 'No current kubeconfig context' }

    & kubectl --context $context apply -k '.\deploy\eks'
    if ($LASTEXITCODE -ne 0) { throw 'Failed to apply KubeCIS node auditor deployment' }

    $imageArgs = @('--context', $context, '-n', 'kube-system', 'set', 'image', 'daemonset/kubecis-node-auditor',
        "node-metadata-init=$image", "node-auditor=$image", "node-metadata-helper=$image")
    & kubectl @imageArgs
    if ($LASTEXITCODE -ne 0) { throw 'Failed to set all three KubeCIS node auditor images' }
    & kubectl --context $context -n kube-system rollout status daemonset/kubecis-node-auditor --timeout=300s
    if ($LASTEXITCODE -ne 0) { throw "KubeCIS node auditor DaemonSet did not roll out with $image" }

    $podsObj = (& kubectl --context $context -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json | ConvertFrom-Json)
    if ($LASTEXITCODE -ne 0 -or -not $podsObj.items) { throw 'No KubeCIS node auditor Pods found' }
    foreach ($pod in @($podsObj.items)) {
        if ($pod.status.phase -ne 'Running') { throw "KubeCIS Pod $($pod.metadata.name) is not Running" }
        $nodeVersion = (& kubectl --context $context -n kube-system exec $pod.metadata.name -c node-auditor -- /usr/local/bin/kubecis version).Trim()
        if ($LASTEXITCODE -ne 0 -or $nodeVersion -cne $expectedVersion) {
            throw "Node auditor version mismatch in $($pod.metadata.name): expected $expectedVersion, got $nodeVersion"
        }
    }
    if ([System.IO.Path]::IsPathRooted($OutputDir)) {
        $reportDir = $OutputDir
    } else {
        $reportDir = Join-Path $homePath $OutputDir
    }
    $scanArgs = @{ Context = $context; ClusterName = $ClusterName; Region = $Region; Binary = '.\kubecis.exe'; ConfigDir = '.\benchmarks'; OutputDir = $reportDir }
    & (Join-Path $PSScriptRoot 'Collect-ConsolidatedEKSReport.ps1') @scanArgs
    $reportPath = Join-Path $reportDir 'consolidated.json'
    if (-not (Test-Path $reportPath)) { throw 'Consolidated report not created' }
    $txt = Join-Path $reportDir 'results.txt'
    & .\kubecis.exe render --input $reportPath --output-file $txt
    if ($LASTEXITCODE -ne 0) { throw 'Failed to render kube-bench-style text report' }
    & .\kubecis.exe review-init --input $reportPath --output-file (Join-Path $reportDir 'review-template.json')
    if ($LASTEXITCODE -ne 0) { throw 'Failed to initialize the human-review template' }
    Write-Host "PASS: completed native EKS preview assessment using matching $expectedVersion"
    Write-Host "Text report: $txt"
    Write-Host "JSON report: $reportPath"
    Write-Host "Review template: $(Join-Path $reportDir 'review-template.json')"
    Write-Warning 'CIS EKS 2.1.0 compliance score remains withheld until human review and official verification are complete.'
} finally {
    Set-Location $original
}

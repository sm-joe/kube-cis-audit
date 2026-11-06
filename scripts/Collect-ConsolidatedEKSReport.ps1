# Collect fresh API and managed-EC2 worker-node evidence from an existing EKS cluster.
# PowerShell 5.1+; requires kubectl and an EKS-authenticated kubeconfig.
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$Context,
    [string]$ClusterName = $Context,
    [string]$Region = '',
    [string]$Namespace = 'kube-system',
    [string]$OutputDir = '.\reports\eks',
    [string]$Binary = '.\kubecis.exe',
    [string]$Benchmark = 'cis-2.1.0',
    [string]$ConfigDir = '.\benchmarks'
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path $Binary -PathType Leaf)) { throw "Missing KubeCIS CLI: $Binary" }
if (-not (Test-Path $ConfigDir -PathType Container)) { throw "Missing benchmark directory: $ConfigDir" }
$active = (& kubectl config current-context)
if ($LASTEXITCODE -ne 0 -or $active.Trim() -ne $Context) {
    throw "Current context must be '$Context' (got '$active'); refusing ambiguous EKS scan"
}
# Check the server to avoid accidentally auditing a Kind cluster under an EKS-like context.
$server = (& kubectl --context $Context version -o json | ConvertFrom-Json).serverVersion.gitVersion
if ($LASTEXITCODE -ne 0 -or $server -notmatch '-eks-') {
    throw "Expected an Amazon EKS Kubernetes server version; observed '$server'"
}
$nodesObj = (& kubectl --context $Context get nodes -l kubernetes.io/os=linux -o json | ConvertFrom-Json)
if ($LASTEXITCODE -ne 0) { throw 'Cannot list Linux worker nodes' }
$nodes = @($nodesObj.items | Where-Object {
    $_.metadata.labels.PSObject.Properties.Name -contains 'eks.amazonaws.com/nodegroup'
} | ForEach-Object { $_.metadata.name } | Sort-Object -Unique)
if ($nodes.Count -eq 0) { throw 'No EKS managed EC2 Linux worker nodes found' }
# Explicitly reject other Linux node types until their auditor support is verified.
$allNodes = @($nodesObj.items | ForEach-Object { $_.metadata.name })
if ($nodes.Count -ne $allNodes.Count) {
    throw 'Mixed EKS compute modes detected: only EKS managed EC2 Linux nodes are supported for complete EKS consolidation'
}
$podsObj = (& kubectl --context $Context -n $Namespace get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json | ConvertFrom-Json)
if ($LASTEXITCODE -ne 0) { throw 'Cannot list EKS node auditor Pods' }
New-Item -Path $OutputDir -ItemType Directory -Force | Out-Null
$dir = (Resolve-Path $OutputDir).Path
$benchmarkDir = (Resolve-Path $ConfigDir).Path
$utf8 = New-Object System.Text.UTF8Encoding($false)
$inputs = New-Object 'System.Collections.Generic.List[string]'
function Write-JsonReport([string]$Path, [string[]]$Lines) {
    $raw = ($Lines -join "`n").Trim()
    if (-not $raw.StartsWith('{')) { throw "Auditor did not return JSON for $Path" }
    $parsed = $raw | ConvertFrom-Json
    if (-not $parsed.generatedAt -or -not $parsed.auditMode) { throw "Incomplete KubeCIS report at $Path" }
    [System.IO.File]::WriteAllText($Path, $raw + "`n", $utf8)
}
$clusterPath = Join-Path $dir 'cluster.json'
# In-cluster Job logs are never reused. The API scan runs freshly from the CLI,
# with an explicit context and policies-only scope required by consolidation.
& $Binary scan --context $Context --config-dir $benchmarkDir --benchmark $Benchmark --target policies --engine native --output json --output-file $clusterPath --exit-code-on-fail=false
if ($LASTEXITCODE -ne 0) { throw 'Fresh EKS API audit failed' }
$clusterReport = Get-Content $clusterPath -Raw | ConvertFrom-Json
if ($clusterReport.cloudProvider -ne 'eks') { throw 'Cluster report did not identify EKS' }
$inputs.Add($clusterPath)
foreach ($node in $nodes) {
    $matching = @($podsObj.items | Where-Object {
        $_.spec.nodeName -eq $node -and $_.status.phase -eq 'Running'
    })
    if ($matching.Count -ne 1) { throw "Expected one Running node-auditor Pod on $node; found $($matching.Count)" }
    $pod = $matching[0].metadata.name
    $data = @(& kubectl --context $Context -n $Namespace exec $pod -c node-auditor -- /usr/local/bin/kubecis node-audit --provider eks --benchmark $Benchmark --target node --metadata-file /metadata/snapshot.json --node-name $node --output json --exit-code-on-fail=false)
    if ($LASTEXITCODE -ne 0) { throw "Node-local audit failed: $node" }
    $safe = $node -replace '[^a-zA-Z0-9_.-]', '_'
    $path = Join-Path $dir "node-$safe.json"
    Write-JsonReport $path $data
    $inputs.Add($path)
}
$providerPath = Join-Path $dir 'provider.json'
& (Join-Path $PSScriptRoot 'Collect-EKSProviderEvidence.ps1') -ClusterName $ClusterName -Region $Region -OutputFile $providerPath
if ($LASTEXITCODE -ne 0 -or -not (Test-Path $providerPath)) { throw 'EKS provider DescribeCluster collection failed' }
$argsList = @('consolidate','--provider-evidence',$providerPath,'--cluster',$clusterPath,'--expected-nodes',($nodes -join ','),'--max-age','30m','--output-file',(Join-Path $dir 'consolidated.json'))
foreach ($path in $inputs) {
    if ($path -ne $clusterPath) { $argsList += @('--local',$path) }
}
& $Binary @argsList
if ($LASTEXITCODE -ne 0) { throw 'EKS consolidation rejected: stale, incomplete, or inconsistent evidence' }
Write-Host "PASS: Validated EKS consolidated report at $(Join-Path $dir 'consolidated.json')"

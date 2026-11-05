# Collect a fresh, validated aggregate from the already deployed Kind cluster.
# Requires KubeCIS Native image including the `consolidate` command in local Go build.
[CmdletBinding()]
param(
    [string]$Namespace = 'kube-system',
    [string]$ClusterName = 'control-plane',
    [string]$OutputDir = '.\reports',
    [string]$Binary = '.\dist\kubecis.exe'
)
$ErrorActionPreference = 'Stop'
$env:KIND_EXPERIMENTAL_PROVIDER = 'podman'
if (-not (Test-Path $Binary)) { throw "Compile native CLI first: go build -o $Binary ./cmd/kubecis" }
$context = "kind-$ClusterName"
$active = (& kubectl config current-context).Trim()
if ($LASTEXITCODE -ne 0 -or $active -ne $context) { throw "Expected context '$context'; actual '$active'. Switch deliberately with kubectl config use-context $context" }
$nodesObj = (& kubectl get nodes -l kubernetes.io/os=linux -o json | ConvertFrom-Json)
if ($LASTEXITCODE -ne 0) {throw 'Failed to retrieve expected Linux nodes'}
$nodes = @($nodesObj.items | ForEach-Object { $_.metadata.name })
if ($nodes.Count -lt 1) {throw 'No Linux nodes found'}
$masters = @($nodesObj.items | Where-Object { $_.metadata.labels.PSObject.Properties.Name -contains 'node-role.kubernetes.io/control-plane' } | ForEach-Object { $_.metadata.name })
$nodePods = (& kubectl -n $Namespace get pods -l app.kubernetes.io/name=kubecis-node-auditor -o json | ConvertFrom-Json)
if ($LASTEXITCODE -ne 0) {throw 'Could not list node auditor Pods'}
$masterPods = (& kubectl -n $Namespace get pods -l app=kubecis-controlplane-auditor -o json | ConvertFrom-Json)
if ($LASTEXITCODE -ne 0) {throw 'Could not list control plane auditor Pods'}
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
$directory = (Resolve-Path $OutputDir).Path
$inputs = New-Object 'System.Collections.Generic.List[string]'
$utf8 = New-Object System.Text.UTF8Encoding($false)
function Write-JsonResult([string]$Path, [string[]]$Content) {
    $raw = ($Content -join "`n").Trim()
    if (-not $raw.StartsWith('{')) {throw "Expected JSON output for $Path; received: $($raw.Substring(0,[Math]::Min(120,$raw.Length)))"}
    $null = $raw | ConvertFrom-Json
    [System.IO.File]::WriteAllText($Path, $raw + "`n", $utf8)
}
$clusterPath = Join-Path $directory 'cluster.json'
$clusterJson = @(& kubectl -n $Namespace logs job/kubecis-native-cluster-scan -c kubecis)
if ($LASTEXITCODE -ne 0) {throw 'Cluster audit Job logs unavailable. Redeploy the scan Job.'}
Write-JsonResult $clusterPath $clusterJson
foreach ($node in $nodes) {
    $matching = @($nodePods.items | Where-Object { $_.spec.nodeName -eq $node -and $_.status.phase -eq 'Running' })
    if ($matching.Count -ne 1) {throw "Expected exactly one Running node auditor Pod on '$node'; found $($matching.Count)"}
    $pod = $matching[0].metadata.name
    $data = @(& kubectl -n $Namespace exec $pod -c node-auditor -- /usr/local/bin/kubecis node-audit --benchmark cis-2.1.0 --target node --metadata-file /metadata/snapshot.json --node-name $node --output json --exit-code-on-fail=false)
    if ($LASTEXITCODE -ne 0) {throw "Node audit failed on $node"}
    $safe = $node -replace '[^a-zA-Z0-9_.-]', '_'
    $path = Join-Path $directory "node-$safe.json"
    Write-JsonResult $path $data
    $inputs.Add($path)
}
foreach ($node in $masters) {
    $matching = @($masterPods.items | Where-Object { $_.spec.nodeName -eq $node -and $_.status.phase -eq 'Running' })
    if ($matching.Count -ne 1) {throw "Expected exactly one Running control-plane auditor Pod on '$node'; found $($matching.Count)"}
    $pod = $matching[0].metadata.name
    $data = @(& kubectl -n $Namespace exec $pod -c controlplane-auditor -- /usr/local/bin/kubecis node-audit --benchmark cis-2.1.0 --target master --metadata-file /metadata/snapshot.json --node-name $node --output json --exit-code-on-fail=false)
    if ($LASTEXITCODE -ne 0) {throw "Control-plane audit failed on $node"}
    $safe = $node -replace '[^a-zA-Z0-9_.-]', '_'
    $path = Join-Path $directory "master-$safe.json"
    Write-JsonResult $path $data
    $inputs.Add($path)
}
$argsList = @('consolidate','--cluster',$clusterPath,'--expected-nodes',($nodes -join ','),'--max-age','30m','--output-file',(Join-Path $directory 'consolidated.json'))
if ($masters.Count -gt 0) { $argsList += @('--expected-control-planes',($masters -join ',')) }
foreach ($path in $inputs) {$argsList += @('--local', $path)}
& $Binary @argsList
if ($LASTEXITCODE -ne 0) {throw 'Consolidation validation failed; reports retained for inspection, no aggregated report accepted'}
Write-Host "Validated consolidated report: $(Join-Path $directory 'consolidated.json')"

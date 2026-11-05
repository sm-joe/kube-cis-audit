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
# Create a fresh cluster scan from the already deployed Job template.
# The new Job retains the existing container, service account and
# security configuration while avoiding stale logs from older scans.

$clusterPath = Join-Path $directory 'cluster.json'

$templateJobName = 'kubecis-native-cluster-scan'
$runId = [DateTimeOffset]::UtcNow.ToString('yyyyMMddHHmmss')
$newJobName = "kubecis-native-e2e-$runId"

$templateJob = & kubectl -n $Namespace get job $templateJobName -o json |
    ConvertFrom-Json

if ($LASTEXITCODE -ne 0 -or $null -eq $templateJob) {
    throw "Cluster scan template Job '$templateJobName' not found"
}

# Build a new Job using the original Pod template.
$newJob = @{
    apiVersion = 'batch/v1'
    kind       = 'Job'
    metadata   = @{
        name      = $newJobName
        namespace = $Namespace
        labels    = @{
            'app.kubernetes.io/name' = 'kubecis-native-e2e'
        }
    }
    spec       = @{
        backoffLimit = 0
        template     = $templateJob.spec.template
    }
}

# The original Job's Pod template may contain generated metadata
# that must not be reused on a newly created Job.
$podTemplate = $templateJob.spec.template

if ($null -ne $podTemplate.metadata) {
    $podTemplate.metadata.PSObject.Properties.Remove('creationTimestamp')
    $podTemplate.metadata.PSObject.Properties.Remove('uid')
    $podTemplate.metadata.PSObject.Properties.Remove('resourceVersion')

    # Remove labels injected by the Kubernetes Job controller.
    if ($null -ne $podTemplate.metadata.labels) {
        $podTemplate.metadata.labels.PSObject.Properties.Remove('controller-uid')
        $podTemplate.metadata.labels.PSObject.Properties.Remove('batch.kubernetes.io/controller-uid')
        $podTemplate.metadata.labels.PSObject.Properties.Remove('job-name')
        $podTemplate.metadata.labels.PSObject.Properties.Remove('batch.kubernetes.io/job-name')
    }
}

$newJob.spec.template = $podTemplate

$jobJson = $newJob | ConvertTo-Json -Depth 100

$jobJson | kubectl -n $Namespace apply -f -

if ($LASTEXITCODE -ne 0) {
    throw "Failed to create fresh cluster scan Job '$newJobName'"
}

# Wait for the newly created scan to finish successfully.
kubectl -n $Namespace wait `
    --for=condition=complete `
    "job/$newJobName" `
    --timeout=180s

if ($LASTEXITCODE -ne 0) {
    kubectl -n $Namespace describe job $newJobName
    kubectl -n $Namespace logs "job/$newJobName" -c kubecis --tail=100

    throw "Fresh cluster scan Job '$newJobName' failed"
}

# Read only this run's logs, not the old template Job's logs.
$clusterJson = @(
    & kubectl -n $Namespace logs "job/$newJobName" -c kubecis
)

if ($LASTEXITCODE -ne 0) {
    throw "Failed to collect fresh cluster report from '$newJobName'"
}

Write-JsonResult $clusterPath $clusterJson

# Validate freshness before proceeding to local reports.
$clusterReport = Get-Content $clusterPath -Raw |
    ConvertFrom-Json

$generated = [DateTimeOffset]::Parse($clusterReport.generatedAt)
$age = [DateTimeOffset]::UtcNow - $generated

if ($age.TotalMinutes -lt 0 -or $age.TotalMinutes -gt 30) {
    throw "Fresh cluster Job returned invalid report timestamp: $($clusterReport.generatedAt)"
}

Write-Host "Fresh cluster audit collected from: $newJobName"
Write-Host "Cluster report timestamp: $($clusterReport.generatedAt)"
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

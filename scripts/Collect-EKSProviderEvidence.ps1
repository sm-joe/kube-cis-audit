# Read-only Amazon EKS DescribeCluster configuration inventory, NOT a CIS score.
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$ClusterName,
    [string]$Region = "",
    [Parameter(Mandatory=$true)][string]$ExpectedAPIServer,
    [Parameter(Mandatory=$true)][string]$OutputFile
)
$ErrorActionPreference = 'Stop'
$awsArgs = @('eks', 'describe-cluster', '--name', $ClusterName, '--output', 'json')
if (-not [string]::IsNullOrWhiteSpace($Region)) { $awsArgs += @('--region', $Region) }
$raw = & aws @awsArgs
if ($LASTEXITCODE -ne 0 -or -not $raw) {
    throw "Read-only AWS DescribeCluster failed for $ClusterName. Verify AWS region and eks:DescribeCluster permission."
}
$cluster = ($raw -join "`n" | ConvertFrom-Json).cluster
if ($null -eq $cluster -or $cluster.status -ne 'ACTIVE') {
    throw "Amazon EKS cluster is missing or not ACTIVE"
}
if ($cluster.name -cne $ClusterName -or -not $cluster.arn -or -not $cluster.endpoint) {
    throw 'AWS DescribeCluster returned missing or unexpected cluster identity'
}
# Compare the actual Kubernetes API server URL from the selected context to
# this exact AWS cluster endpoint. Matching Kubernetes versions is NOT enough.
try {
    $expectedUri = [Uri]::new($ExpectedAPIServer)
    $actualUri = [Uri]::new([string]$cluster.endpoint)
} catch { throw 'Invalid Kubernetes/EKS API server URL for identity verification' }
if ($expectedUri.Scheme -ne 'https' -or $actualUri.Scheme -ne 'https' -or
    $expectedUri.Host -ine $actualUri.Host -or $expectedUri.Port -ne $actualUri.Port -or
    $expectedUri.AbsolutePath.TrimEnd('/') -cne $actualUri.AbsolutePath.TrimEnd('/')) {
    throw "EKS cluster identity mismatch: Kubernetes context does not point to AWS DescribeCluster endpoint for '$ClusterName'"
}
$vpc = $cluster.resourcesVpcConfig
if ($null -eq $vpc -or $null -eq $vpc.endpointPrivateAccess -or $null -eq $vpc.endpointPublicAccess) {
    throw 'EKS provider endpoint configuration is incomplete'
}
$types = @('api','audit','authenticator','controllerManager','scheduler')
$logging = @{}
foreach ($type in $types) {
    $matching = @($cluster.logging.clusterLogging | Where-Object {
        @($_.types) -contains $type
    })
    if ($matching.Count -ne 1 -or $null -eq $matching[0].enabled) {
        throw "EKS DescribeCluster response is missing control-plane logging evidence for $type"
    }
    $logging[$type] = [bool]$matching[0].enabled
}
$regionValue = $Region
if ([string]::IsNullOrWhiteSpace($regionValue)) {
    $regionValue = [string]$env:AWS_REGION
    if (-not $regionValue) { $regionValue = [string]$env:AWS_DEFAULT_REGION }
}
if (-not $regionValue -and $cluster.arn -match '^arn:aws[^:]*:eks:([^:]+):') {
    $regionValue = $Matches[1]
}
if (-not $regionValue) { throw 'Unable to determine the EKS AWS region' }
if ($cluster.arn -notmatch '^arn:aws[^:]*:eks:([^:]+):([0-9]{12}):cluster/(.+)
$evidence = [ordered]@{
    schemaVersion = 1
    cloudProvider = 'eks'
    clusterName = [string]$cluster.name
    clusterArn = [string]$cluster.arn
    region = $regionValue
    kubernetesVersion = [string]$cluster.version
    generatedAt = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ')
    endpointPrivateAccess = [bool]$vpc.endpointPrivateAccess
    endpointPublicAccess = [bool]$vpc.endpointPublicAccess
    publicAccessCidrs = @($vpc.publicAccessCidrs)
    controlPlaneLogging = $logging
    assessmentStatus = 'INVENTORY_ONLY'
    assessmentNote = 'Read-only AWS configuration evidence. Does not certify effective security or map to any official CIS EKS scored control.'
}
$parent = Split-Path -Parent $OutputFile
if ($parent) { New-Item -Path $parent -ItemType Directory -Force | Out-Null }
$utf8 = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText(
    [System.IO.Path]::GetFullPath($OutputFile),
    ($evidence | ConvertTo-Json -Depth 8) + "`n",
    $utf8
)
Write-Host "Collected read-only EKS provider inventory for $($cluster.name) ($regionValue)"
) {
    throw 'AWS DescribeCluster returned malformed EKS cluster ARN'
}
if ($Matches[1] -ne $regionValue -or $Matches[3] -cne $ClusterName) {
    throw 'AWS DescribeCluster ARN name/region does not match the requested target'
}
$evidence = [ordered]@{
    schemaVersion = 1
    cloudProvider = 'eks'
    clusterName = [string]$cluster.name
    region = $regionValue
    kubernetesVersion = [string]$cluster.version
    generatedAt = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffffffZ')
    endpointPrivateAccess = [bool]$vpc.endpointPrivateAccess
    endpointPublicAccess = [bool]$vpc.endpointPublicAccess
    publicAccessCidrs = @($vpc.publicAccessCidrs)
    controlPlaneLogging = $logging
    assessmentStatus = 'INVENTORY_ONLY'
    assessmentNote = 'Read-only AWS configuration evidence. Does not certify effective security or map to any official CIS EKS scored control.'
}
$parent = Split-Path -Parent $OutputFile
if ($parent) { New-Item -Path $parent -ItemType Directory -Force | Out-Null }
$utf8 = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText(
    [System.IO.Path]::GetFullPath($OutputFile),
    ($evidence | ConvertTo-Json -Depth 8) + "`n",
    $utf8
)
Write-Host "Collected read-only EKS provider inventory for $($cluster.name) ($regionValue)"

# Read-only Amazon EKS DescribeCluster configuration inventory, NOT a CIS score.
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][string]$ClusterName,
    [string]$Region = "",
    [Parameter(Mandatory=$true)][string]$ExpectedAPIServer,
    [Parameter(Mandatory=$true)][string]$OutputFile,
    [string]$Context = ""
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
    throw 'AWS DescribeCluster returned missing or unexpected identity'
}
try {
    $expectedUri = [Uri]::new($ExpectedAPIServer)
    $actualUri = [Uri]::new([string]$cluster.endpoint)
} catch { throw 'Invalid EKS API server URI' }
if ($expectedUri.Scheme -ne 'https' -or $actualUri.Scheme -ne 'https' -or
    $expectedUri.Host -ine $actualUri.Host -or $expectedUri.Port -ne $actualUri.Port -or
    $expectedUri.AbsolutePath.TrimEnd('/') -cne $actualUri.AbsolutePath.TrimEnd('/')) {
    throw 'EKS cluster identity mismatch: selected kubeconfig endpoint differs from AWS DescribeCluster endpoint'
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
if ($cluster.arn -notmatch '^arn:aws[a-z-]*:eks:([^:]+):([0-9]{12}):cluster/(.+)\z' -or
    $Matches[1] -cne $regionValue -or $Matches[3] -cne $ClusterName) {
    throw 'AWS EKS ARN name/region mismatch'
}
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
# Record access-entry evidence without modifying IAM or Kubernetes state.
$authMode = [string]$cluster.accessConfig.authenticationMode
$evidence.authenticationMode = $authMode
$evidence.accessInventoryStatus = 'UNAVAILABLE'
$evidence.accessEntries = @()
if ($authMode -eq 'CONFIG_MAP') {
    $evidence.accessInventoryStatus = 'LEGACY_ONLY'
} elseif ($authMode -eq 'API' -or $authMode -eq 'API_AND_CONFIG_MAP') {
    try {
        $accessArgs = @('eks', 'list-access-entries', '--cluster-name', $ClusterName, '--region', $regionValue, '--output', 'json')
        $accessRaw = & aws @accessArgs 2>$null
        $accessCode = $LASTEXITCODE
        if ($accessCode -eq 0 -and $accessRaw) {
            $accessObj = $accessRaw -join "`n" | ConvertFrom-Json
            if ($null -ne $accessObj.accessEntries) {
                $accessList = @($accessObj.accessEntries)
                if ($accessList.Count -le 100) {
                    $collected = @()
                    $allComplete = $true
                    foreach ($principal in $accessList) {
                        $item = [ordered]@{ principalArn = [string]$principal; type = ''; kubernetesGroups = @(); policyScopes = @(); error = '' }
                        try {
                            $detailRaw = & aws eks describe-access-entry --cluster-name $ClusterName --principal-arn $principal --region $regionValue --output json 2>$null
                            if ($LASTEXITCODE -ne 0 -or -not $detailRaw) { throw 'entry not readable' }
                            $detail = ($detailRaw -join "`n" | ConvertFrom-Json).accessEntry
                            if ($null -eq $detail) { throw 'missing access entry details' }
                            $item.type = [string]$detail.type
                            $item.kubernetesGroups = @($detail.kubernetesGroups | Where-Object { $_ })
                            $policyRaw = & aws eks list-associated-access-policies --cluster-name $ClusterName --principal-arn $principal --region $regionValue --output json 2>$null
                            if ($LASTEXITCODE -ne 0 -or -not $policyRaw) { throw 'associated policies not readable' }
                            $associated = ($policyRaw -join "`n" | ConvertFrom-Json).associatedAccessPolicies
                            foreach ($policy in @($associated | Where-Object { $null -ne $_ })) {
                                $scopeType = [string]$policy.accessScope.type
                                $namespaces = @($policy.accessScope.namespaces | Where-Object { $_ })
                                $item.policyScopes += "$([string]$policy.policyArn):$scopeType:$($namespaces -join ',')"
                            }
                        } catch {
                            $item.error = 'incomplete access entry or policy evidence'
                            $allComplete = $false
                        }
                        $collected += $item
                    }
                    $evidence.accessEntries = @($collected)
                    if ($allComplete) { $evidence.accessInventoryStatus = 'COLLECTED' }
                }
            }
        }
    } catch {
        $evidence.accessInventoryStatus = 'UNAVAILABLE'
    }
}

# EKS managed VPC CNI add-on. Missing permissions or self-managed CNI are
# explicitly unknown, never interpreted as secure defaults.
$evidence.vpcCniStatus = 'UNAVAILABLE'
$evidence.vpcCniVersion = ''
try {
    $addonRaw = & aws eks describe-addon --cluster-name $ClusterName --addon-name vpc-cni --region $regionValue --output json 2>$null
    if ($LASTEXITCODE -eq 0 -and $addonRaw) {
        $addon = ($addonRaw -join "`n" | ConvertFrom-Json).addon
        if ($null -ne $addon -and $addon.status) {
            $evidence.vpcCniStatus = [string]$addon.status
            $evidence.vpcCniVersion = [string]$addon.addonVersion
            if ($addon.configurationValues) {
                $addonCfg = ([string]$addon.configurationValues | ConvertFrom-Json)
                $flag = $addonCfg.enableNetworkPolicy
                if ($null -ne $flag) {
                    if ($flag -is [bool]) { $evidence.vpcCniNetworkPolicy = [bool]$flag }
                    elseif ([string]$flag -eq 'true') { $evidence.vpcCniNetworkPolicy = $true }
                    elseif ([string]$flag -eq 'false') { $evidence.vpcCniNetworkPolicy = $false }
                }
            }
        }
    }
} catch {
    $evidence.vpcCniStatus = 'UNAVAILABLE'
}

# Collect only names, digests and scan states for images actually deployed in
# Pods. Raw image-scan vulnerability reports and credentials are not persisted.
if ($Context) {
    $evidence.ecrInventoryStatus = 'UNAVAILABLE'
    $evidence.ecrImages = @()
    $evidence.nonEcrImageCount = 0
    try {
        $podRaw = & kubectl --context $Context get pods --all-namespaces -o json 2>$null
        if ($LASTEXITCODE -eq 0 -and $podRaw) {
            $podObj = $podRaw -join "`n" | ConvertFrom-Json
            if ($null -ne $podObj.items) {
                $ecrRefs = @{}
                $others = @{}
                foreach ($pod in @($podObj.items)) {
                    foreach ($container in @($pod.spec.containers)) {
                        $image = [string]$container.image
                        if ($image -match '^([0-9]{12})\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com(?:\.cn)?/([^@:]+)') {
                            $registry = "$($Matches[1]).dkr.ecr.$($Matches[2]).amazonaws.com"
                            $imageRegion = $Matches[2]
                            $repository = $Matches[3]
                            $digest = ''
                            $statusMatch = @($pod.status.containerStatuses | Where-Object { $_.name -eq $container.name })
                            if ($statusMatch.Count -gt 0 -and [string]$statusMatch[0].imageID -match 'sha256:[a-fA-F0-9]{64}') {
                                $digest = $Matches[0].ToLowerInvariant()
                            }
                            $key = "$registry/$repository@$digest"
                            $ecrRefs[$key] = @{ registry=$registry; region=$imageRegion; repository=$repository; digest=$digest }
                        } else {
                            $others[$image] = $true
                        }
                    }
                }
                $evidence.nonEcrImageCount = $others.Count
                if ($ecrRefs.Count -le 80) {
                    $images = @()
                    foreach ($record in $ecrRefs.Values) {
                        $entry = [ordered]@{ registry=$record.registry; repository=$record.repository; digest=$record.digest; scanType=''; scanStatus=''; error='' }
                        try {
                            $scanArgs = @('ecr','get-registry-scanning-configuration','--region',$record.region,'--output','json')
                            $scanRaw = & aws @scanArgs 2>$null
                            if ($LASTEXITCODE -ne 0 -or -not $scanRaw) { throw 'scanning configuration unavailable' }
                            $scanCfg = ($scanRaw -join "`n" | ConvertFrom-Json).scanningConfiguration
                            if ($null -eq $scanCfg -or -not $scanCfg.scanType) { throw 'missing ECR scan configuration' }
                            $entry.scanType = [string]$scanCfg.scanType
                            if (-not $record.digest) { throw 'deployed image digest unavailable' }
                            $scanResultRaw = & aws ecr describe-image-scan-findings --repository-name $record.repository --image-id "imageDigest=$($record.digest)" --region $record.region --output json 2>$null
                            if ($LASTEXITCODE -ne 0 -or -not $scanResultRaw) { throw 'image scan status unavailable' }
                            $scanResult = $scanResultRaw -join "`n" | ConvertFrom-Json
                            if ($scanResult.imageScanStatus.status) {
                                $entry.scanStatus = [string]$scanResult.imageScanStatus.status
                            } else {
                                throw 'image scan status not returned'
                            }
                        } catch {
                            $entry.error = 'unable to verify image scan status'
                        }
                        $images += $entry
                    }
                    $evidence.ecrImages = @($images)
                    $evidence.ecrInventoryStatus = 'COLLECTED'
                }
            }
        }
    } catch {
        $evidence.ecrInventoryStatus = 'UNAVAILABLE'
    }
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

# Fail-closed verification of an EKS consolidated KubeCIS Native report.
[CmdletBinding()]
param(
    [string]$ReportPath = '.\reports\eks\consolidated.json',
    [int]$MaxAgeMinutes = 30,
    [ValidateSet('Preview', 'Full')][string]$ReleaseLevel = 'Preview',
    [string]$ExpectedClusterName = '',
    [string]$ExpectedRegion = ''
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path $ReportPath -PathType Leaf)) { throw "Consolidated EKS report missing: $ReportPath" }
$r = Get-Content $ReportPath -Raw | ConvertFrom-Json
if ($r.auditMode -ne 'consolidated') { throw "Expected consolidated audit; got $($r.auditMode)" }
if ($r.cloudProvider -ne 'eks') { throw "Expected EKS; got $($r.cloudProvider)" }
if ($r.benchmarkCompatibilityStatus -ne 'provisional') { throw 'This gate expects the explicitly provisional benchmark profile' }
if ($null -ne $r.compliance.compliancePercent) { throw 'Provisional EKS assessment must not publish compliance percentage' }
if ($r.compliance.scoreStatus -notlike 'withheld:*') { throw 'Compliance score was not withheld' }
$nodes = @($r.expectedNodes | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
if ($nodes.Count -eq 0) { throw 'No expected EKS managed EC2 worker nodes' }
if (@($r.expectedControlPlaneNodes | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }).Count -ne 0) { throw 'Unexpected master-local expectations on AWS-managed EKS control plane' }
$sources = @($r.sources)
if ($sources.Count -ne (1 + $nodes.Count)) { throw "Expected $($nodes.Count + 1) sources, got $($sources.Count)" }
if (@($sources | Where-Object { $_.auditMode -eq 'native' -and -not $_.nodeName }).Count -ne 1) {
    throw 'Expected one fresh native Kubernetes API report'
}
if (@($sources | Where-Object { $_.auditMode -eq 'master-local' }).Count -gt 0) {
    throw 'Unexpected EKS master-local report'
}
foreach ($node in $nodes) {
    if (@($sources | Where-Object { $_.auditMode -eq 'node-local' -and $_.nodeName -eq $node }).Count -ne 1) {
        throw "Expected exactly one node-local report for $node"
    }
}
$utcNow = [DateTimeOffset]::UtcNow
foreach ($src in $sources) {
    $time = [DateTimeOffset]::Parse($src.generatedAt)
    $age = ($utcNow - $time).TotalMinutes
    if ($age -lt -2 -or $age -gt $MaxAgeMinutes) {
        throw "Stale or future source $($src.auditMode)/$($src.nodeName): age $([math]::Round($age,2)) minutes"
    }
}
if ($r.summary.total -ne @($r.results).Count) { throw 'Report summary and findings count mismatch' }
if ($null -eq $r.providerEvidence) { throw 'Missing read-only AWS EKS provider configuration inventory' }
$p = $r.providerEvidence
if ($p.cloudProvider -ne 'eks' -or $p.assessmentStatus -ne 'INVENTORY_ONLY' -or
    $p.schemaVersion -ne 1 -or -not $p.clusterName -or -not $p.region) {
    throw 'Invalid EKS provider evidence metadata'
}
if ($r.clusterVersion -notmatch '^v(\d+)\.(\d+)\.') { throw 'Invalid cluster Kubernetes version' }
$expectedMinor = "$($Matches[1]).$($Matches[2])"
if ($p.kubernetesVersion -ne $expectedMinor) {
    throw "AWS provider Kubernetes version does not match API server: $($p.kubernetesVersion) vs $expectedMinor"
}
$providerAge = ([DateTimeOffset]::UtcNow - [DateTimeOffset]::Parse($p.generatedAt)).TotalMinutes
if ($providerAge -gt $MaxAgeMinutes -or $providerAge -lt -2) {
    throw 'AWS EKS provider evidence is stale or future dated'
}
foreach ($type in @('api','audit','authenticator','controllerManager','scheduler')) {
    if ($p.controlPlaneLogging.PSObject.Properties.Name -notcontains $type) {
        throw "Missing AWS EKS logging inventory field $type"
    }
}
if ($ExpectedClusterName -and $p.clusterName -cne $ExpectedClusterName) {
    throw "Provider cluster name mismatch: expected '$ExpectedClusterName', got '$($p.clusterName)'"
}
if ($ExpectedRegion -and $p.region -cne $ExpectedRegion) {
    throw "Provider cluster region mismatch: expected '$ExpectedRegion', got '$($p.region)'"
}
if ($p.clusterArn -notmatch '^arn:aws[a-z-]*:eks:([^:]+):([0-9]{12}):cluster/(.+)\z' -or
    $Matches[1] -cne $p.region -or $Matches[3] -cne $p.clusterName) {
    throw 'Provider evidence ARN name/region mismatch'
}
if ($ReleaseLevel -eq 'Full') {
    throw 'Full CIS EKS release NOT ACCEPTED: the profile is provisional; the official complete control inventory and version applicability are not verified.'
}
if ($r.compliance.scoredTotal -lt 1) { throw 'No scored checks were included' }
if ($r.compliance.scoredPass + $r.compliance.scoredFail -gt $r.compliance.scoredTotal) {
    throw 'Decisive statuses exceed total scored controls'
}
# A consolidated report is not an acceptable EKS node audit if every kubelet
# configuration result is NOT_ASSESSED. Require explicit verified AL2023
# source evidence for all three essential worker checks on every expected node.
$requiredNodeChecks = @('KUBECIS-EKS-NODE-CONFIG-MODE', 'KUBECIS-NATIVE-NODE-1', 'KUBECIS-NATIVE-NODE-2')
foreach ($node in $nodes) {
    foreach ($checkId in $requiredNodeChecks) {
        $found = @($r.results | Where-Object { $_.id -eq $checkId -and $_.scope -eq $node })
        if ($found.Count -ne 1) {
            throw "Missing or duplicate EKS kubelet control $checkId for worker $node"
        }
        $finding = $found[0]
        if (@('PASS','FAIL') -notcontains $finding.status) {
            throw "EKS node evidence incomplete: $checkId on $node is $($finding.status). $($finding.details)"
        }
        $sourceEvidence = @($finding.evidence | Where-Object {
            $_.name -eq 'etc/kubernetes/kubelet/config.json'
        })
        if ($sourceEvidence.Count -eq 0) {
            throw "EKS node control $checkId on $node lacks AL2023 nodeadm config.json evidence"
        }
    }
}
Write-Host "PASS (PREVIEW ONLY): EKS API + $($nodes.Count) managed EC2 node source(s) verified"
Write-Host "Benchmark: $($r.benchmark) ($($r.benchmarkCompatibilityStatus)); score withheld as required"
Write-Host "Scored evidence coverage: $($r.compliance.coveragePercent)% within configured starter profile"

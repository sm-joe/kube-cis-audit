# Fail-closed verification of an EKS consolidated KubeCIS Native report.
[CmdletBinding()]
param(
    [string]$ReportPath = '.\reports\eks\consolidated.json',
    [int]$MaxAgeMinutes = 30
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path $ReportPath -PathType Leaf)) { throw "Consolidated EKS report missing: $ReportPath" }
$r = Get-Content $ReportPath -Raw | ConvertFrom-Json
if ($r.auditMode -ne 'consolidated') { throw "Expected consolidated audit; got $($r.auditMode)" }
if ($r.cloudProvider -ne 'eks') { throw "Expected EKS; got $($r.cloudProvider)" }
if ($r.benchmarkCompatibilityStatus -ne 'provisional') { throw 'This gate expects the explicitly provisional benchmark profile' }
if ($null -ne $r.compliance.compliancePercent) { throw 'Provisional EKS assessment must not publish compliance percentage' }
if ($r.compliance.scoreStatus -notlike 'withheld:*') { throw 'Compliance score was not withheld' }
$nodes = @($r.expectedNodes)
if ($nodes.Count -eq 0) { throw 'No expected EKS managed EC2 worker nodes' }
if (@($r.expectedControlPlaneNodes).Count -ne 0) { throw 'Unexpected master-local expectations on AWS-managed EKS control plane' }
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
if ($r.compliance.scoredTotal -lt 1) { throw 'No scored checks were included' }
if ($r.compliance.scoredPass + $r.compliance.scoredFail -gt $r.compliance.scoredTotal) {
    throw 'Decisive statuses exceed total scored controls'
}
Write-Host "PASS: EKS API + $($nodes.Count) managed EC2 node source(s) verified"
Write-Host "Benchmark: $($r.benchmark) ($($r.benchmarkCompatibilityStatus)); score withheld as required"
Write-Host "Scored evidence coverage: $($r.compliance.coveragePercent)% within configured starter profile"

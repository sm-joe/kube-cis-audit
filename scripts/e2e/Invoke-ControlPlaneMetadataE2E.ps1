[CmdletBinding()]
param(
    [string]$Context = "kind-kind",
    [string]$Image = "kubecis:e2e-3-0-0-rc3",
    [string]$Artifacts = ".\artifacts\controlplane-e2e"
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$Manifest = Join-Path $RepoRoot "e2e\controlplane-auditor-daemonset.yaml"
if (-not (Test-Path $Manifest)) { throw "Control-plane DaemonSet manifest not found: $Manifest" }
New-Item -ItemType Directory -Path $Artifacts -Force | Out-Null
$Rendered = Join-Path $Artifacts "controlplane-auditor-daemonset.generated.yaml"

$source = Get-Content -Raw -Path $Manifest
$matches = [regex]::Matches($source, '(?m)^\s*image:\s*kubecis:e2e\s*$')
if ($matches.Count -ne 3) { throw "Expected exactly 3 generic image references; found $($matches.Count)." }
$source = [regex]::Replace($source, '(?m)^(\s*image:\s*)kubecis:e2e\s*$', ('$1' + $Image))
Set-Content -Path $Rendered -Value $source -Encoding utf8

function Invoke-Kubectl([string[]]$Arguments, [string]$FailureMessage) {
    $output = & kubectl --context $Context @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) { throw "$FailureMessage`n$($output -join "`n")" }
    return $output
}

Write-Host "Applying only the KubeCIS control-plane auditor DaemonSet to context '$Context'..."
Invoke-Kubectl @("apply", "-f", $Rendered) "Applying control-plane auditor DaemonSet failed" | Out-Host
Invoke-Kubectl @("-n", "kube-system", "rollout", "status", "daemonset/kubecis-e2e-controlplane-auditor", "--timeout=180s") "Control-plane auditor did not become ready" | Out-Host

$podsJson = Invoke-Kubectl @("-n", "kube-system", "get", "pods", "-l", "app=kubecis-e2e-controlplane-auditor", "-o", "json") "Could not list control-plane auditor Pods"
$pods = (($podsJson -join "`n") | ConvertFrom-Json).items
if (@($pods).Count -lt 1) { throw "No control-plane auditor Pod was scheduled. Confirm the node has label node-role.kubernetes.io/control-plane and /etc/kubernetes/manifests exists." }
$pod = $pods[0].metadata.name

$helper = Invoke-Kubectl @("-n", "kube-system", "logs", $pod, "-c", "controlplane-metadata-helper", "--tail=20") "Could not retrieve control-plane metadata helper logs"
$helper | Set-Content -Encoding utf8 (Join-Path $Artifacts "controlplane-metadata-helper.log")
$auditor = Invoke-Kubectl @("-n", "kube-system", "logs", $pod, "-c", "controlplane-auditor", "--tail=100") "Could not retrieve control-plane auditor logs"
$auditor | Set-Content -Encoding utf8 (Join-Path $Artifacts "controlplane-audit.log")

$joined = $auditor -join "`n"
$jsonStart = $joined.IndexOf('{')
$jsonEnd = $joined.LastIndexOf('}')
if ($jsonStart -lt 0 -or $jsonEnd -le $jsonStart) { throw "No JSON report found. See $Artifacts\controlplane-audit.log" }
$report = $joined.Substring($jsonStart, $jsonEnd - $jsonStart + 1) | ConvertFrom-Json
if ($report.auditMode -ne "master-local") { throw "Unexpected audit mode '$($report.auditMode)'." }
$report | ConvertTo-Json -Depth 100 | Set-Content -Encoding utf8 (Join-Path $Artifacts "controlplane-audit.json")
if ($report.results.Count -ne 1 -or $report.results[0].id -ne "1.1.1") { throw "Expected exactly control 1.1.1 in control-plane report." }
$result = $report.results[0]
if ($result.status -notin @("PASS", "FAIL")) { throw "Control 1.1.1 did not receive file-mode evidence: status=$($result.status), details=$($result.details)" }
$modeEvidence = $result.evidence | Where-Object { $_.field -eq "mode" -and $_.observed }
if ($null -eq $modeEvidence) { throw "Control 1.1.1 report is missing observed permission-mode evidence." }
if ($modeEvidence.observed -notmatch '^[0-7]{4}$') { throw "Control 1.1.1 observed mode is not a four-digit octal permission value: $($modeEvidence.observed)" }
if ($report.nodeName -ne $pods[0].spec.nodeName) { throw "Report node identity does not match scheduled control-plane node." }

# The master-local profile currently has one scored host-file check. Assert
# report accounting so a PASS/FAIL result cannot be emitted with empty scoring.
if ($report.compliance.scoredTotal -ne 1) { throw "Expected one scored control in master-local scope; got $($report.compliance.scoredTotal)." }
if (($report.compliance.scoredPass + $report.compliance.scoredFail) -ne 1) { throw "Scored PASS/FAIL counters do not reconcile with the single assessed control." }
if ($report.compliance.coveragePercent -ne 100) { throw "Expected 100% evidence coverage for the single assessed control; got $($report.compliance.coveragePercent)." }
if ($report.compliance.scoreStatus -notlike "complete:*") { throw "Unexpected master-local score status: $($report.compliance.scoreStatus)" }
Write-Host "Control 1.1.1 assessed on $($report.nodeName): $($result.status)."
Write-Host "Observed evidence: $($result.evidence | Where-Object { $_.field -eq 'mode' } | ConvertTo-Json -Compress)"
Write-Host "Artifacts: $Artifacts"

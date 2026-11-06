# Evidence-bound interactive review for a completed KubeCIS Native EKS scan.
# Reviewers must actively choose APPROVED, REJECTED or EXCEPTION. No default approvals.
[CmdletBinding()]
param(
    [string]$ReportDir = 'reports\eks',
    [Parameter(Mandatory=$true)][string]$Reviewer,
    [string]$DecisionsFile = '',
    [switch]$ShowEvidence
)
$ErrorActionPreference='Stop'
$root=Split-Path -Parent $PSScriptRoot
if (-not [IO.Path]::IsPathRooted($ReportDir)) {$ReportDir=Join-Path $root $ReportDir}
$source=Join-Path $ReportDir 'consolidated.json'
$templatePath=Join-Path $ReportDir 'review-template.json'
if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {throw "Missing consolidated report: $source"}
if (-not (Test-Path -LiteralPath $templatePath -PathType Leaf)) {throw "Missing review template: $templatePath"}
if ([string]::IsNullOrWhiteSpace($Reviewer)) {throw 'Reviewer identity must not be blank'}
$sourceHash=(Get-FileHash -LiteralPath $source -Algorithm SHA256).Hash.ToLowerInvariant()
$template=Get-Content -LiteralPath $templatePath -Raw | ConvertFrom-Json
if ([string]$template.reportSha256 -cne $sourceHash) {throw 'Review template is stale or belongs to another scan. Refusing review.'}
if (-not $DecisionsFile){$DecisionsFile=Join-Path $ReportDir 'review-decisions.json'}
elseif (-not [IO.Path]::IsPathRooted($DecisionsFile)) {$DecisionsFile=Join-Path (Get-Location).Path $DecisionsFile}
$decisions=@()
if (Test-Path -LiteralPath $DecisionsFile -PathType Leaf) {
    $saved=Get-Content -LiteralPath $DecisionsFile -Raw | ConvertFrom-Json
    if ([string]$saved.reportSha256 -cne $sourceHash){throw "Saved decisions belong to another source report: $DecisionsFile"}
    $decisions=@($saved.decisions | Where-Object {$null -ne $_})
}
$report=Get-Content -LiteralPath $source -Raw | ConvertFrom-Json
foreach ($candidate in @($template.pendingReviews)) {
    $id=[string]$candidate.id
    $scope=[string]$candidate.scope
    $matches=@($decisions | Where-Object {$_.id -ceq $id -and [string]$_.scope -ceq $scope})
    if ($matches.Count -gt 1){throw "Duplicate saved decisions for $id / $scope"}
    if ($matches.Count -eq 1){continue}
    Write-Host ''
    Write-Host "Control $id $scope : $($candidate.title)" -ForegroundColor Cyan
    Write-Host ([string]$candidate.details)
    $detail=@($report.results | Where-Object {$_.id -ceq $id -and [string]$_.scope -ceq $scope})
    if ($detail.Count -ne 1){throw "Expected unique manual control $id / $scope in source report"}
    if ($detail[0].status -cne 'MANUAL'){throw "Review attempt on a non-MANUAL control $id"}
    if ($ShowEvidence) {
        foreach($ev in @($detail[0].evidence | Where-Object {$null -ne $_})){
            Write-Host ("  Evidence: {0} {1}/{2} {3}: {4} (expected {5})" -f $ev.resource,$ev.namespace,$ev.name,$ev.field,$ev.observed,$ev.expected)
        }
    }
    Write-Host 'Choose A=APPROVED, R=REJECTED, E=EXCEPTION, S=SKIP, Q=QUIT'
    $choice=(Read-Host 'Decision').Trim().ToUpperInvariant()
    if($choice -eq 'Q'){break}
    if($choice -eq 'S' -or $choice -eq ''){continue}
    $label=switch($choice){
        'A' {'APPROVED'}
        'R' {'REJECTED'}
        'E' {'EXCEPTION'}
        default {throw "Invalid decision choice: $choice"}
    }
    $evidence=(Read-Host 'Evidence reference (ticket, URL, reviewed document)').Trim()
    $why=(Read-Host 'Specific human justification').Trim()
    if (-not $evidence -or -not $why){throw "No evidence or justification for $id; decision not saved"}
    $decisions+= [ordered]@{id=$id;scope=$scope;decision=$label;reviewer=$Reviewer;evidenceRef=$evidence;justification=$why;reviewedAt=[DateTime]::UtcNow.ToString('o')}
    $json=[ordered]@{reportSha256=$sourceHash;decisions=@($decisions)} | ConvertTo-Json -Depth 15
    $tmp=$DecisionsFile+'.tmp'
    [IO.File]::WriteAllText($tmp,$json,[Text.UTF8Encoding]::new($false))
    Move-Item -LiteralPath $tmp -Destination $DecisionsFile -Force
}
if (-not (Test-Path -LiteralPath $DecisionsFile -PathType Leaf)) {
    [IO.File]::WriteAllText($DecisionsFile,([ordered]@{reportSha256=$sourceHash;decisions=@()} | ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false))
}
$exe=Join-Path $root 'kubecis.exe'
& $exe review --input $source --decisions $DecisionsFile --output-file (Join-Path $ReportDir 'review-validated.json')
if ($LASTEXITCODE -ne 0){throw 'Human decisions validation failed; source report untouched'}
& $exe finalize --input $source --decisions $DecisionsFile --output-dir $ReportDir
if ($LASTEXITCODE -ne 0){throw 'Final assessment generation failed; source report untouched'}
Write-Host "Final assessment: $(Join-Path $ReportDir 'final-assessment.txt')"

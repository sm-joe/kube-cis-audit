[CmdletBinding()]
param(
    [string]$ClusterName = "kind"
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$Artifacts = Join-Path $RepoRoot "artifacts\e2e"
New-Item -ItemType Directory -Force -Path $Artifacts | Out-Null
$ImageName = "kubecis:e2e"
$ImageArchive = Join-Path $Artifacts "kubecis-e2e-image.tar"
$ExpectedBuildVersion = "2.9.0-consolidated-security-assessment"
$OriginalContext = $null

function Require-Command {
    param([Parameter(Mandatory = $true)][string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found in PATH. This script requires Podman, kubectl and kind."
    }
}

# Capture native stdout/stderr without allowing PowerShell 5.1 to turn the
# Kind/Podman provider's informational stderr messages into terminating errors.
function Invoke-NativeCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Command,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [switch]$AllowFailure
    )

    $previousPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = "Continue"
        $nativeOutput = @(& $Command @Arguments 2>&1)
        $nativeExitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousPreference
    }

    $outputText = ($nativeOutput | ForEach-Object { $_.ToString() }) -join [Environment]::NewLine
    $result = [pscustomobject]@{
        ExitCode = $nativeExitCode
        Output   = $outputText.Trim()
    }

    if ($nativeExitCode -ne 0 -and -not $AllowFailure) {
        throw "$Command $($Arguments -join ' ') failed with exit code $nativeExitCode.`n$outputText"
    }

    return $result
}

function Invoke-Kubectl {
    param(
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [switch]$AllowFailure
    )
    return Invoke-NativeCommand -Command "kubectl" -Arguments $Arguments -AllowFailure:$AllowFailure
}

Require-Command "podman"
Require-Command "kubectl"
Require-Command "kind"

try {
    Write-Host "[1/8] Verifying the existing Kind cluster '$ClusterName'..."

    $kindResult = Invoke-NativeCommand -Command "kind" -Arguments @("get", "clusters")
    $clusters = @(
        $kindResult.Output -split "`r?`n" |
            ForEach-Object { $_.Trim() } |
            Where-Object { $_ -and $_ -notmatch '^using\s+podman\s+due\s+to\s+KIND_EXPERIMENTAL_PROVIDER$' -and $_ -notmatch '^enabling\s+experimental\s+podman\s+provider$' }
    )

    if ($clusters -notcontains $ClusterName) {
        throw "Kind cluster '$ClusterName' was not found. Existing clusters: $($clusters -join ', '). This script will NOT create or delete a cluster."
    }

    $contextName = "kind-$ClusterName"
    $contextsResult = Invoke-Kubectl -Arguments @("config", "get-contexts", "-o", "name")
    $contexts = @($contextsResult.Output -split "`r?`n" | ForEach-Object { $_.Trim() } | Where-Object { $_ })
    if ($contexts -notcontains $contextName) {
        throw "The Kind cluster exists, but kubeconfig context '$contextName' is missing. Refusing to create or replace the cluster."
    }

    $currentContextResult = Invoke-Kubectl -Arguments @("config", "current-context") -AllowFailure
    if ($currentContextResult.ExitCode -eq 0) {
        $OriginalContext = $currentContextResult.Output.Trim()
    }

    $null = Invoke-Kubectl -Arguments @("config", "use-context", $contextName)
    $nodesResult = Invoke-Kubectl -Arguments @("get", "nodes", "-o", "json")
    $nodes = $nodesResult.Output | ConvertFrom-Json
    if (@($nodes.items).Count -lt 1) {
        throw "The existing cluster '$ClusterName' returned no Kubernetes nodes."
    }
    $notReadyNodes = @(
        foreach ($node in $nodes.items) {
            $readyCondition = $node.status.conditions | Where-Object { $_.type -eq "Ready" } | Select-Object -First 1
            if (-not $readyCondition -or $readyCondition.status -ne "True") { $node.metadata.name }
        }
    )
    if ($notReadyNodes.Count -gt 0) {
        throw "The existing cluster has NotReady nodes: $($notReadyNodes -join ', '). Fix cluster health before running E2E."
    }
    Write-Host "Reusing context '$contextName' with $(@($nodes.items).Count) Ready node(s). No cluster creation will be attempted."

    Write-Host "[2/8] Building KubeCIS image with Podman..."
    $buildResult = Invoke-NativeCommand -Command "podman" -Arguments @("build", "--pull", "-t", $ImageName, $RepoRoot)
    Write-Host $buildResult.Output

    $userResult = Invoke-NativeCommand -Command "podman" -Arguments @("image", "inspect", $ImageName, "--format", "{{.Config.User}}")
    if ($userResult.Output.Trim() -ne "65532:65532") {
        throw "Runtime image must default to UID/GID 65532:65532; got '$($userResult.Output)'"
    }

    $versionResult = Invoke-NativeCommand -Command "podman" -Arguments @("run", "--rm", $ImageName, "version")
    if ($versionResult.Output.Trim() -ne $ExpectedBuildVersion) {
        throw "Image version command returned '$($versionResult.Output)'; expected '$ExpectedBuildVersion'."
    }

    Write-Host "[3/8] Exporting the Podman image and loading it into the existing Kind cluster..."
    if (Test-Path $ImageArchive) { Remove-Item -Force $ImageArchive }
    $null = Invoke-NativeCommand -Command "podman" -Arguments @("save", "--format", "docker-archive", "--output", $ImageArchive, $ImageName)
    $loadResult = Invoke-NativeCommand -Command "kind" -Arguments @("load", "image-archive", $ImageArchive, "--name", $ClusterName)
    Write-Host $loadResult.Output
    if (Test-Path $ImageArchive) { Remove-Item -Force $ImageArchive }

    Write-Host "[4/8] Resetting only the previous KubeCIS E2E resources, then applying RBAC and fixtures..."
    # These names are dedicated to this test. The Kind cluster itself is never modified/deleted.
    $null = Invoke-Kubectl -Arguments @("delete", "-f", (Join-Path $RepoRoot "e2e\cluster-job.yaml"), "--ignore-not-found=true") -AllowFailure
    $null = Invoke-Kubectl -Arguments @("delete", "-f", (Join-Path $RepoRoot "e2e\node-auditor-daemonset.yaml"), "--ignore-not-found=true") -AllowFailure
    $null = Invoke-Kubectl -Arguments @("delete", "-f", (Join-Path $RepoRoot "e2e\fixtures.yaml"), "--ignore-not-found=true") -AllowFailure

    $rbacResult = Invoke-Kubectl -Arguments @("apply", "-f", (Join-Path $RepoRoot "deploy\rbac.yaml"))
    Write-Host $rbacResult.Output
    $fixturesResult = Invoke-Kubectl -Arguments @("apply", "-f", (Join-Path $RepoRoot "e2e\fixtures.yaml"))
    Write-Host $fixturesResult.Output

    Write-Host "[5/8] Running cluster/API audit Job..."
    $jobApplyResult = Invoke-Kubectl -Arguments @("apply", "-f", (Join-Path $RepoRoot "e2e\cluster-job.yaml"))
    Write-Host $jobApplyResult.Output

    $deadline = (Get-Date).AddMinutes(4)
    $jobPhase = ""
    while ((Get-Date) -lt $deadline) {
        $jobResult = Invoke-Kubectl -Arguments @("get", "job", "kubecis-e2e-cluster-audit", "-n", "kube-system", "-o", "json")
        $job = $jobResult.Output | ConvertFrom-Json
        if ([int]$job.status.succeeded -gt 0) { $jobPhase = "Complete"; break }
        if ([int]$job.status.failed -gt 0) { $jobPhase = "Failed"; break }
        Start-Sleep -Seconds 3
    }
    if (-not $jobPhase) { throw "Cluster audit Job did not terminate within four minutes." }

    $clusterLogsResult = Invoke-Kubectl -Arguments @("logs", "job/kubecis-e2e-cluster-audit", "-n", "kube-system")
    $clusterLogs = $clusterLogsResult.Output
    $clusterLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "cluster-audit.log")
    $jsonMatch = [regex]::Match($clusterLogs, '(?s)\{.*\}')
    if (-not $jsonMatch.Success) { throw "Cluster audit did not emit a JSON report. See artifacts/e2e/cluster-audit.log" }
    $clusterReport = $jsonMatch.Value | ConvertFrom-Json
    $clusterReport | ConvertTo-Json -Depth 100 | Set-Content -Encoding utf8 (Join-Path $Artifacts "cluster-audit.json")
    if ($clusterReport.buildVersion -ne $ExpectedBuildVersion) {
        throw "Unexpected cluster build version: $($clusterReport.buildVersion)"
    }

    $expectedIds = @("KUBECIS-2.4.7", "KUBECIS-2.5.1", "KUBECIS-2.6.1", "KUBECIS-2.7.1", "KUBECIS-2.8.1", "KUBECIS-2.9.1")
    $actualIds = @($clusterReport.results | ForEach-Object { $_.id })
    foreach ($id in $expectedIds) {
        if ($actualIds -notcontains $id) { throw "Cluster report is missing expected check '$id'" }
    }

    $checksById = @{}
    foreach ($result in $clusterReport.results) { $checksById[$result.id] = $result }
    foreach ($id in @("KUBECIS-2.4.2", "KUBECIS-2.4.5", "KUBECIS-2.5.1")) {
        if (-not $checksById.ContainsKey($id) -or $checksById[$id].status -ne "FAIL") {
            $actualStatus = if ($checksById.ContainsKey($id)) { $checksById[$id].status } else { "MISSING" }
            throw "Expected insecure fixture finding '$id' to be FAIL; got '$actualStatus'"
        }
    }
    Write-Host "Cluster report verified. Job phase: $jobPhase; checks: $($clusterReport.results.Count); FAIL: $($clusterReport.summary.fail); NOT_ASSESSED: $($clusterReport.summary.notAssessed)"

    Write-Host "[6/8] Running node-local DaemonSet with metadata-only helper..."
    $dsApplyResult = Invoke-Kubectl -Arguments @("apply", "-f", (Join-Path $RepoRoot "e2e\node-auditor-daemonset.yaml"))
    Write-Host $dsApplyResult.Output
    $rolloutResult = Invoke-Kubectl -Arguments @("rollout", "status", "daemonset/kubecis-e2e-node-auditor", "-n", "kube-system", "--timeout=180s")
    Write-Host $rolloutResult.Output
    Start-Sleep -Seconds 40

    $nodePodsResult = Invoke-Kubectl -Arguments @("get", "pods", "-n", "kube-system", "-l", "app=kubecis-e2e-node-auditor", "-o", "json")
    $nodePodsResponse = $nodePodsResult.Output | ConvertFrom-Json
    $nodePods = @($nodePodsResponse.items)
    if ($nodePods.Count -lt 1) { throw "No node auditor Pods were created." }
    $nodePod = $nodePods[0].metadata.name

    $helperLogsResult = Invoke-Kubectl -Arguments @("logs", $nodePod, "-n", "kube-system", "-c", "node-metadata-helper", "--tail=20")
    $nodeLogsResult = Invoke-Kubectl -Arguments @("logs", $nodePod, "-n", "kube-system", "-c", "node-auditor", "--tail=200")
    $helperLogs = $helperLogsResult.Output
    $nodeLogs = $nodeLogsResult.Output
    $helperLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-metadata-helper.log")
    $nodeLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-audit.log")
    if ($helperLogs -notmatch "snapshot refreshed") { throw "Metadata helper did not confirm a refreshed snapshot. See node-metadata-helper.log" }
    if ($nodeLogs -notmatch '"auditMode":\s*"node-local"') { throw "Node auditor did not emit a node-local JSON report. See node-audit.log" }
    if ($nodeLogs -notmatch [regex]::Escape('"buildVersion": "' + $ExpectedBuildVersion + '"')) { throw "Node report has an unexpected build version." }

    Write-Host "[7/8] Verifying runtime security boundaries..."
    $nodePodResult = Invoke-Kubectl -Arguments @("get", "pod", $nodePod, "-n", "kube-system", "-o", "json")
    $nodePodJson = $nodePodResult.Output | ConvertFrom-Json
    $mainContainer = $nodePodJson.spec.containers | Where-Object { $_.name -eq "node-auditor" } | Select-Object -First 1
    $helperContainer = $nodePodJson.spec.containers | Where-Object { $_.name -eq "node-metadata-helper" } | Select-Object -First 1
    if (-not $mainContainer -or -not $mainContainer.securityContext.readOnlyRootFilesystem -or $mainContainer.securityContext.runAsUser -ne 65532) {
        throw "Node auditor main container security context does not match the non-root/read-only design."
    }
    if (-not $helperContainer -or $helperContainer.securityContext.runAsUser -ne 0 -or $helperContainer.securityContext.capabilities.add -notcontains "DAC_READ_SEARCH") {
        throw "Metadata helper does not have the narrowly scoped DAC_READ_SEARCH exception."
    }
    if ($nodePodJson.spec.automountServiceAccountToken -ne $false) { throw "Node DaemonSet must not mount a service-account token." }

    Write-Host "[8/8] E2E checks completed successfully."
    Write-Host "Artifacts: $Artifacts"
    Write-Host "Cluster: $ClusterName (reused; not created or deleted)"
    Write-Host "Cluster report: $($clusterReport.summary | ConvertTo-Json -Compress)"
    Write-Host "Job phase '$jobPhase' is acceptable when fixture findings cause the CLI's findings-based non-zero exit; the report is validated above."
}
finally {
    if ($OriginalContext) {
        $restoreResult = Invoke-Kubectl -Arguments @("config", "use-context", $OriginalContext) -AllowFailure
        if ($restoreResult.ExitCode -eq 0) {
            Write-Host "Restored original kubectl context '$OriginalContext'."
        } else {
            Write-Warning "Could not restore original kubectl context '$OriginalContext': $($restoreResult.Output)"
        }
    }
    if (Test-Path $ImageArchive) { Remove-Item -Force $ImageArchive -ErrorAction SilentlyContinue }
    Write-Host "The existing Kind cluster '$ClusterName' was not created, deleted, or otherwise cleaned up."
}

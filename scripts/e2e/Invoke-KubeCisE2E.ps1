[CmdletBinding()]
param(
    [string]$ClusterName = "kubecis-e2e",
    [switch]$KeepCluster
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$Artifacts = Join-Path $RepoRoot "artifacts\e2e"
New-Item -ItemType Directory -Force -Path $Artifacts | Out-Null

function Require-Command {
    param([string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found in PATH. Install Docker Desktop, kubectl and kind, then retry."
    }
}

Require-Command "docker"
Require-Command "kubectl"
Require-Command "kind"

# Native tools such as Kind/Podman can write informational diagnostics to stderr
# while returning exit code 0. Do not let PowerShell promote that stderr stream
# into terminating errors; the script checks $LASTEXITCODE after native commands.
if (Test-Path Variable:PSNativeCommandUseErrorActionPreference) {
    $PSNativeCommandUseErrorActionPreference = $false
}
$clusters = @(kind get clusters 2>$null)
$kindGetClustersExitCode = $LASTEXITCODE
if ($kindGetClustersExitCode -ne 0) {
    throw "kind get clusters failed ($kindGetClustersExitCode). Verify the Kind/Podman provider and container runtime."
}
if ($clusters -contains $ClusterName) {
    throw "Kind cluster '$ClusterName' already exists. Delete it or choose another -ClusterName; this test expects a disposable cluster."
}
$ClusterCreated = $false

try {
    Write-Host "[1/8] Creating disposable Kind cluster '$ClusterName'..."
    kind create cluster --name $ClusterName
    if ($LASTEXITCODE -ne 0) { throw "kind create cluster failed ($LASTEXITCODE)" }
    $ClusterCreated = $true

    Write-Host "[2/8] Building KubeCIS image..."
    docker build --pull -t kubecis:e2e $RepoRoot
    if ($LASTEXITCODE -ne 0) { throw "Docker image build failed ($LASTEXITCODE)" }
    $imageUser = docker image inspect kubecis:e2e --format '{{.Config.User}}'
    if ($LASTEXITCODE -ne 0 -or $imageUser.Trim() -ne "65532:65532") { throw "Runtime image must default to UID/GID 65532; got '$imageUser'" }
    $imageVersion = docker run --rm kubecis:e2e version
    if ($LASTEXITCODE -ne 0 -or $imageVersion.Trim() -ne "2.9.0-consolidated-security-assessment") { throw "Image version command failed or returned an unexpected build identifier: '$imageVersion'" }

    Write-Host "[3/8] Loading image into Kind..."
    kind load docker-image kubecis:e2e --name $ClusterName
    if ($LASTEXITCODE -ne 0) { throw "kind load docker-image failed ($LASTEXITCODE)" }

    kubectl config use-context "kind-$ClusterName" | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "Could not select Kind context" }

    Write-Host "[4/8] Applying read-only auditor RBAC and intentionally insecure test fixtures..."
    kubectl apply -f (Join-Path $RepoRoot "deploy\rbac.yaml")
    if ($LASTEXITCODE -ne 0) { throw "Applying auditor RBAC failed" }
    kubectl apply -f (Join-Path $RepoRoot "e2e\fixtures.yaml")
    if ($LASTEXITCODE -ne 0) { throw "Applying E2E fixtures failed" }

    Write-Host "[5/8] Running cluster/API audit Job..."
    kubectl apply -f (Join-Path $RepoRoot "e2e\cluster-job.yaml")
    if ($LASTEXITCODE -ne 0) { throw "Creating cluster audit Job failed" }

    $deadline = (Get-Date).AddMinutes(4)
    $jobPhase = ""
    while ((Get-Date) -lt $deadline) {
        $job = kubectl get job kubecis-e2e-cluster-audit -n kube-system -o json | ConvertFrom-Json
        if ($job.status.succeeded -gt 0) { $jobPhase = "Complete"; break }
        if ($job.status.failed -gt 0) { $jobPhase = "Failed"; break }
        Start-Sleep -Seconds 3
    }
    if (-not $jobPhase) { throw "Cluster audit Job did not terminate within four minutes" }

    $clusterLogs = kubectl logs job/kubecis-e2e-cluster-audit -n kube-system 2>&1 | Out-String
    $clusterLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "cluster-audit.log")
    $jsonMatch = [regex]::Match($clusterLogs, '(?s)\{.*\}')
    if (-not $jsonMatch.Success) { throw "Cluster audit did not emit a JSON report. See artifacts/e2e/cluster-audit.log" }
    $clusterReport = $jsonMatch.Value | ConvertFrom-Json
    $clusterReport | ConvertTo-Json -Depth 100 | Set-Content -Encoding utf8 (Join-Path $Artifacts "cluster-audit.json")
    if ($clusterReport.buildVersion -ne "2.9.0-consolidated-security-assessment") {
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
        if ($checksById[$id].status -ne "FAIL") {
            throw "Expected insecure fixture finding '$id' to be FAIL; got '$($checksById[$id].status)'"
        }
    }
    Write-Host "Cluster audit report verified. Job phase: $jobPhase; checks: $($clusterReport.results.Count); FAIL: $($clusterReport.summary.fail); NOT_ASSESSED: $($clusterReport.summary.notAssessed)"

    Write-Host "[6/8] Running node-local DaemonSet with metadata-only helper..."
    kubectl apply -f (Join-Path $RepoRoot "e2e\node-auditor-daemonset.yaml")
    if ($LASTEXITCODE -ne 0) { throw "Creating node auditor DaemonSet failed" }
    kubectl rollout status daemonset/kubecis-e2e-node-auditor -n kube-system --timeout=180s
    if ($LASTEXITCODE -ne 0) { throw "Node auditor DaemonSet did not become ready" }
    Start-Sleep -Seconds 40

    $nodePodsResponse = kubectl get pods -n kube-system -l app=kubecis-e2e-node-auditor -o json | ConvertFrom-Json
    $nodePods = @($nodePodsResponse.items)
    if ($nodePods.Count -lt 1) { throw "No node auditor Pods were created" }
    $nodePod = $nodePods[0].metadata.name
    $helperLogs = kubectl logs $nodePod -n kube-system -c node-metadata-helper --tail=20 2>&1 | Out-String
    $nodeLogs = kubectl logs $nodePod -n kube-system -c node-auditor --tail=200 2>&1 | Out-String
    $helperLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-metadata-helper.log")
    $nodeLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-audit.log")
    if ($helperLogs -notmatch "snapshot refreshed") { throw "Metadata helper did not confirm a refreshed snapshot. See node-metadata-helper.log" }
    if ($nodeLogs -notmatch '"auditMode": "node-local"') { throw "Node auditor did not emit a node-local JSON report. See node-audit.log" }
    if ($nodeLogs -notmatch '"buildVersion": "2.9.0-consolidated-security-assessment"') { throw "Node report has an unexpected build version" }

    Write-Host "[7/8] Verifying runtime security boundaries..."
    $nodePodJson = kubectl get pod $nodePod -n kube-system -o json | ConvertFrom-Json
    $mainContainer = $nodePodJson.spec.containers | Where-Object { $_.name -eq "node-auditor" }
    $helperContainer = $nodePodJson.spec.containers | Where-Object { $_.name -eq "node-metadata-helper" }
    if (-not $mainContainer.securityContext.readOnlyRootFilesystem -or $mainContainer.securityContext.runAsUser -ne 65532) {
        throw "Node auditor main container security context does not match the non-root/read-only design"
    }
    if ($helperContainer.securityContext.runAsUser -ne 0 -or $helperContainer.securityContext.capabilities.add -notcontains "DAC_READ_SEARCH") {
        throw "Metadata helper does not have the narrowly scoped DAC_READ_SEARCH exception"
    }
    if ($nodePodJson.spec.automountServiceAccountToken -ne $false) { throw "Node DaemonSet must not mount a service-account token" }

    Write-Host "[8/8] E2E checks completed. Artifacts: $Artifacts"
    Write-Host "Cluster report: $($clusterReport.summary | ConvertTo-Json -Compress)"
    Write-Host "Cluster job phase '$jobPhase' is acceptable when the fixture findings cause the CLI's findings-based non-zero exit. The report is still validated above."
}
finally {
    if ($ClusterCreated -and -not $KeepCluster) {
        Write-Host "Deleting disposable Kind cluster '$ClusterName'..."
        kind delete cluster --name $ClusterName
    } elseif ($ClusterCreated -and $KeepCluster) {
        Write-Host "Keeping Kind cluster '$ClusterName' as requested."
    }
}

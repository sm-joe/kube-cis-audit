[CmdletBinding()]
param(
    [string]$Context = "kind-kind",
    [string]$Version = "3.1.0-native-integrated-rc3.2"
)

$ErrorActionPreference = "Stop"

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$Artifacts = Join-Path $RepoRoot "artifacts\e2e"
$ImageArchive = Join-Path $Artifacts "kubecis-e2e-image.tar"
$Image = "kubecis:e2e-$($Version.Replace('.', '-'))"
$KubernetesImage = "localhost/$Image"
$RenderedJobManifest = Join-Path $Artifacts "cluster-job.generated.yaml"
$RenderedNodeDaemonSetManifest = Join-Path $Artifacts "node-auditor-daemonset.generated.yaml"

# Kind context names use the form kind-<cluster-name>.
if (-not $Context.StartsWith("kind-")) {
    throw "Context '$Context' is not a Kind context. Expected a context such as 'kind-kind'."
}
$ClusterName = $Context.Substring(5)
if ([string]::IsNullOrWhiteSpace($ClusterName)) {
    throw "Could not derive the Kind cluster name from context '$Context'."
}

New-Item -ItemType Directory -Force -Path $Artifacts | Out-Null

function Require-Command {
    param([Parameter(Mandatory)][string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' was not found in PATH."
    }
}

function Invoke-Native {
    param(
        [Parameter(Mandatory)][string]$Command,
        [Parameter(Mandatory)][string[]]$Arguments,
        [string]$FailureMessage = "$Command failed"
    )

    $previousErrorActionPreference = $ErrorActionPreference
    $previousNativePreference = $null
    $hasNativePreference = Test-Path Variable:PSNativeCommandUseErrorActionPreference

    if ($hasNativePreference) {
        $previousNativePreference = $PSNativeCommandUseErrorActionPreference
        $PSNativeCommandUseErrorActionPreference = $false
    }

    try {
        # Kind/Podman can write informational messages to stderr. Capture both
        # streams, but use the native process exit code to determine success.
        $ErrorActionPreference = "Continue"
        $output = & $Command @Arguments 2>&1
        $exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
        if ($hasNativePreference) {
            $PSNativeCommandUseErrorActionPreference = $previousNativePreference
        }
    }

    if ($exitCode -ne 0) {
        $detail = ($output | ForEach-Object { "$_" }) -join [Environment]::NewLine
        throw "$FailureMessage (exit $exitCode). $detail"
    }

    return $output
}

Require-Command "podman"
Require-Command "kind"
Require-Command "kubectl"

# This script is specifically for the existing Podman-backed Kind cluster.
$env:KIND_EXPERIMENTAL_PROVIDER = "podman"

Write-Host "[1/8] Verifying the existing Kind cluster and context '$Context'..."

$clusters = @(
    Invoke-Native -Command "kind" `
        -Arguments @("get", "clusters") `
        -FailureMessage "kind get clusters failed" |
        ForEach-Object { "$_".Trim() } |
        Where-Object { $_ -ne "" -and $_ -notmatch "^using podman due to" -and $_ -notmatch "^enabling experimental podman provider$" }
)

if ($clusters -notcontains $ClusterName) {
    throw "Expected existing Kind cluster '$ClusterName' was not found. This script never creates or deletes clusters. Found: $($clusters -join ', ')"
}

$contextNames = @(
    Invoke-Native -Command "kubectl" `
        -Arguments @("config", "get-contexts", "-o", "name") `
        -FailureMessage "Could not read kubeconfig contexts" |
        ForEach-Object { "$_".Trim() }
)

if ($contextNames -notcontains $Context) {
    throw "Kubeconfig context '$Context' was not found. Available contexts: $($contextNames -join ', ')"
}

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "get", "nodes") `
    -FailureMessage "Existing cluster is not reachable" |
    Out-Host

Write-Host "[2/8] Building and verifying the version-specific Linux container image..."

Invoke-Native -Command "podman" `
    -Arguments @("build", "--pull", "--build-arg", "VERSION=$Version", "-t", $Image, $RepoRoot) `
    -FailureMessage "Podman image build failed" |
    Out-Host

$imageUser = (
    Invoke-Native -Command "podman" `
        -Arguments @("image", "inspect", "--format", "{{.Config.User}}", $Image) `
        -FailureMessage "Could not inspect image" |
        Out-String
).Trim()

if ($imageUser -ne "65532:65532") {
    throw "Runtime image must default to UID/GID 65532:65532; got '$imageUser'."
}

$imageVersion = (
    Invoke-Native -Command "podman" `
        -Arguments @("run", "--rm", $Image, "version") `
        -FailureMessage "Container version command failed" |
        Out-String
).Trim()

if ($imageVersion -ne $Version) {
    throw "Container version mismatch. Expected '$Version', got '$imageVersion'."
}

Write-Host "Verified image: $Image; version: $imageVersion; runtime user: $imageUser"

Write-Host "[3/8] Loading the Podman-built image into the existing Kind cluster..."

if (Test-Path $ImageArchive) {
    Remove-Item -Force $ImageArchive
}

Invoke-Native -Command "podman" `
    -Arguments @("save", "--format", "docker-archive", "-o", $ImageArchive, $Image) `
    -FailureMessage "Could not export Podman image" |
    Out-Host

Invoke-Native -Command "kind" `
    -Arguments @("load", "image-archive", $ImageArchive, "--name", $ClusterName) `
    -FailureMessage "Could not load Podman image into existing Kind cluster" |
    Out-Host

# Podman archive imports this image into Kind containerd with the localhost/
# namespace. Kubernetes must use that exact name when imagePullPolicy is Never.
# Use ctr's quiet output: it returns complete image references, whereas
# the tabular output from crictl/ctr splits repository and tag into columns.
$nodeImageReferences = @(
    Invoke-Native -Command "podman" `
        -Arguments @("exec", "kind-control-plane", "ctr", "-n", "k8s.io", "images", "list", "-q") `
        -FailureMessage "Could not inspect image references in the existing Kind node" |
        ForEach-Object { "$_".Trim() } |
        Where-Object { $_ -ne "" }
)

if ($nodeImageReferences -notcontains $KubernetesImage) {
    $nodeImageList = (
        Invoke-Native -Command "podman" `
            -Arguments @("exec", "kind-control-plane", "ctr", "-n", "k8s.io", "images", "list") `
            -FailureMessage "Could not retrieve the Kind containerd image table" |
            Out-String
    )
    throw "Image '$KubernetesImage' is not present in Kind containerd. Available references:`n$($nodeImageReferences -join [Environment]::NewLine)`nContainerd image table:`n$nodeImageList"
}
Write-Host "Verified exact Kind containerd image reference: $KubernetesImage"

# Render a temporary Job manifest with the exact version-specific image tag.
# The tracked e2e/cluster-job.yaml remains unchanged.
$jobManifestPath = Join-Path $RepoRoot "e2e\cluster-job.yaml"
if (-not (Test-Path $jobManifestPath)) {
    throw "Cluster audit Job manifest was not found: $jobManifestPath"
}

$jobManifest = Get-Content -Raw -Path $jobManifestPath
$genericImageLine = 'image:\s*kubecis:e2e(?=\s*(#.*)?$)'
if (-not [regex]::IsMatch($jobManifest, $genericImageLine, [System.Text.RegularExpressions.RegexOptions]::Multiline)) {
    throw "Expected image reference 'image: kubecis:e2e' was not found in e2e\cluster-job.yaml. No manifest was applied."
}

$jobManifest = [regex]::Replace(
    $jobManifest,
    $genericImageLine,
    "image: $KubernetesImage",
    [System.Text.RegularExpressions.RegexOptions]::Multiline
)
Set-Content -Path $RenderedJobManifest -Value $jobManifest -Encoding utf8

Write-Host "[4/8] Applying read-only auditor RBAC and intentionally insecure fixtures..."

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "apply", "-f", (Join-Path $RepoRoot "deploy\rbac.yaml")) `
    -FailureMessage "Applying auditor RBAC failed" |
    Out-Host

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "apply", "-f", (Join-Path $RepoRoot "e2e\fixtures.yaml")) `
    -FailureMessage "Applying E2E fixtures failed" |
    Out-Host

Write-Host "[5/8] Running cluster/API audit Job..."

# Jobs are immutable; replace only this named E2E Job, never the cluster or
# unrelated workloads.
Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "-n", "kube-system", "delete", "job", "kubecis-e2e-cluster-audit", "--ignore-not-found=true") `
    -FailureMessage "Could not remove the previous KubeCIS E2E Job" |
    Out-Host

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "apply", "-f", $RenderedJobManifest) `
    -FailureMessage "Creating cluster audit Job failed" |
    Out-Host

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "-n", "kube-system", "wait", "--for=condition=complete", "job/kubecis-e2e-cluster-audit", "--timeout=5m") `
    -FailureMessage "Cluster audit Job did not complete" |
    Out-Host

$clusterLogs = (
    Invoke-Native -Command "kubectl" `
        -Arguments @("--context", $Context, "-n", "kube-system", "logs", "job/kubecis-e2e-cluster-audit") `
        -FailureMessage "Could not retrieve cluster audit logs" |
        Out-String
)

$clusterLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "cluster-audit.log")

$jsonMatch = [regex]::Match($clusterLogs, '(?s)\{.*\}')
if (-not $jsonMatch.Success) {
    throw "Cluster audit did not emit a JSON report. See artifacts\e2e\cluster-audit.log"
}

$clusterReport = $jsonMatch.Value | ConvertFrom-Json
$clusterReport | ConvertTo-Json -Depth 100 | Set-Content -Encoding utf8 (Join-Path $Artifacts "cluster-audit.json")

if ($clusterReport.buildVersion -ne $Version) {
    throw "Unexpected cluster build version: $($clusterReport.buildVersion); expected $Version. Confirm the Job used image '$Image'."
}

$expectedIds = @(
    "KUBECIS-2.4.7",
    "KUBECIS-2.5.1",
    "KUBECIS-2.6.1",
    "KUBECIS-2.7.1",
    "KUBECIS-2.8.1",
    "KUBECIS-2.9.1"
)

$actualIds = @($clusterReport.results | ForEach-Object { $_.id })
foreach ($id in $expectedIds) {
    if ($actualIds -notcontains $id) {
        throw "Cluster report is missing expected check '$id'"
    }
}

$checksById = @{}
foreach ($result in $clusterReport.results) {
    $checksById[$result.id] = $result
}

foreach ($id in @("KUBECIS-2.4.2", "KUBECIS-2.4.5", "KUBECIS-2.5.1")) {
    if ($checksById[$id].status -ne "FAIL") {
        throw "Expected insecure fixture finding '$id' to be FAIL; got '$($checksById[$id].status)'"
    }
}

Write-Host "Cluster report verified. Checks: $($clusterReport.results.Count); FAIL: $($clusterReport.summary.fail); NOT_ASSESSED: $($clusterReport.summary.notAssessed)"

Write-Host "[6/8] Running node-local DaemonSet with metadata-only helper..."

# Render the E2E DaemonSet with the same version-specific image as the Job.
# Both the auditor and metadata-helper containers must use the freshly built image.
$nodeDaemonSetPath = Join-Path $RepoRoot "e2e\node-auditor-daemonset.yaml"
if (-not (Test-Path $nodeDaemonSetPath)) {
    throw "Node auditor DaemonSet manifest was not found: $nodeDaemonSetPath"
}

$nodeDaemonSetManifest = Get-Content -Raw -Path $nodeDaemonSetPath
$nodeImagePattern = '(?m)^([ \t]*image:[ \t]*)kubecis:e2e([ \t]*(?:#.*)?)$'
$nodeImageMatches = [regex]::Matches($nodeDaemonSetManifest, $nodeImagePattern)
if ($nodeImageMatches.Count -ne 3) {
    throw "Expected exactly three 'image: kubecis:e2e' entries in e2e\node-auditor-daemonset.yaml (node-metadata-init, node-auditor and node-metadata-helper); found $($nodeImageMatches.Count). No DaemonSet manifest was applied."
}

$nodeDaemonSetManifest = [regex]::Replace(
    $nodeDaemonSetManifest,
    $nodeImagePattern,
    ('$1' + $KubernetesImage + '$2')
)
Set-Content -Path $RenderedNodeDaemonSetManifest -Value $nodeDaemonSetManifest -Encoding utf8

# Fail before apply if either container does not reference the intended image.
$renderedNodeImageLines = @(
    Get-Content -Path $RenderedNodeDaemonSetManifest |
        Where-Object { $_.Trim() -ceq "image: $KubernetesImage" }
)
$renderedNodeImageCount = $renderedNodeImageLines.Count
if ($renderedNodeImageCount -ne 3) {
    throw "Generated node DaemonSet does not contain exactly three references to '$KubernetesImage'."
}

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "apply", "-f", $RenderedNodeDaemonSetManifest) `
    -FailureMessage "Applying node auditor DaemonSet failed" |
    Out-Host

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "-n", "kube-system", "rollout", "restart", "daemonset/kubecis-e2e-node-auditor") `
    -FailureMessage "Could not restart the E2E node auditor DaemonSet" |
    Out-Host

Invoke-Native -Command "kubectl" `
    -Arguments @("--context", $Context, "-n", "kube-system", "rollout", "status", "daemonset/kubecis-e2e-node-auditor", "--timeout=180s") `
    -FailureMessage "Node auditor DaemonSet did not become ready" |
    Out-Host

$nodePodsJson = (
    Invoke-Native -Command "kubectl" `
        -Arguments @("--context", $Context, "-n", "kube-system", "get", "pods", "-l", "app=kubecis-e2e-node-auditor", "-o", "json") `
        -FailureMessage "Could not list node auditor Pods" |
        Out-String
)

$nodePods = @((ConvertFrom-Json $nodePodsJson).items)
if ($nodePods.Count -lt 1) {
    throw "No node auditor Pods were created."
}

$nodePod = $nodePods[0].metadata.name
$nodePodJson = $null
$initSucceeded = $false
for ($attempt = 0; $attempt -lt 60; $attempt++) {
    $podJsonText = (
        Invoke-Native -Command "kubectl" `
            -Arguments @("--context", $Context, "-n", "kube-system", "get", "pod", $nodePod, "-o", "json") `
            -FailureMessage "Could not inspect node auditor Pod readiness" |
            Out-String
    )
    $nodePodJson = $podJsonText | ConvertFrom-Json
    $initStatus = $nodePodJson.status.initContainerStatuses |
        Where-Object { $_.name -eq "node-metadata-init" } |
        Select-Object -First 1

    if ($initStatus -and $initStatus.state.terminated) {
        if ($initStatus.state.terminated.exitCode -ne 0) {
            throw "node-metadata-init failed with exit code $($initStatus.state.terminated.exitCode). See node-metadata-init.log."
        }
        $initSucceeded = $true
        break
    }
    Start-Sleep -Seconds 2
}
if (-not $initSucceeded) {
    throw "Timed out waiting for node-metadata-init to publish the initial snapshot."
}

# Init-container success proves the first snapshot was atomically published
# before node-auditor started. Capture init logs and the periodic helper logs.
$initLogs = (
    Invoke-Native -Command "kubectl" `
        -Arguments @("--context", $Context, "-n", "kube-system", "logs", $nodePod, "-c", "node-metadata-init") `
        -FailureMessage "Could not retrieve initial metadata helper logs" |
        Out-String
)
$helperLogs = (
    Invoke-Native -Command "kubectl" `
        -Arguments @("--context", $Context, "-n", "kube-system", "logs", $nodePod, "-c", "node-metadata-helper", "--tail=50") `
        -FailureMessage "Could not retrieve metadata helper logs" |
        Out-String
)
$nodeLogs = (
    Invoke-Native -Command "kubectl" `
        -Arguments @("--context", $Context, "-n", "kube-system", "logs", $nodePod, "-c", "node-auditor", "--tail=300") `
        -FailureMessage "Could not retrieve node auditor logs" |
        Out-String
)

$initLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-metadata-init.log")
$helperLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-metadata-helper.log")
$nodeLogs | Set-Content -Encoding utf8 (Join-Path $Artifacts "node-audit.log")

if ($nodeLogs -notmatch '"auditMode": "node-local"') {
    throw "Node auditor did not emit a node-local JSON report. See node-audit.log"
}
if ($nodeLogs -notmatch ('"buildVersion": "' + [regex]::Escape($Version) + '"')) {
    throw "Node report has an unexpected build version."
}
# The first assessment must contain actual file-mode evidence. NOT_ASSESSED
# here means the snapshot handoff failed and is an E2E failure, not a security
# finding to be silently accepted.
$nodeControlEvidence = [regex]::Match(
    $nodeLogs,
    '(?s)"id"\s*:\s*"4\.1\.9".*?"status"\s*:\s*"(PASS|FAIL|NOT_ASSESSED)"'
)
if (-not $nodeControlEvidence.Success) {
    throw "Node report did not contain control 4.1.9. See node-audit.log"
}
if ($nodeControlEvidence.Groups[1].Value -eq "NOT_ASSESSED") {
    throw "Node control 4.1.9 was NOT_ASSESSED; initial metadata snapshot was not consumed. See node-metadata-init.log and node-audit.log."
}
if ($nodeLogs -notmatch '"observed"\s*:\s*"0[0-7]{3}"') {
    throw "Node report did not include an observed octal file mode for control 4.1.9."
}

Write-Host "[7/8] Verifying runtime security boundaries..."

$mainContainer = $nodePodJson.spec.containers | Where-Object { $_.name -eq "node-auditor" }
$helperContainer = $nodePodJson.spec.containers | Where-Object { $_.name -eq "node-metadata-helper" }
$initContainer = $nodePodJson.spec.initContainers | Where-Object { $_.name -eq "node-metadata-init" }

if (-not $initContainer -or $initContainer.securityContext.runAsUser -ne 0 -or $initContainer.securityContext.capabilities.add -notcontains "DAC_READ_SEARCH") {
    throw "Initial metadata collector does not have the expected narrowly scoped DAC_READ_SEARCH exception."
}
if (-not $mainContainer.securityContext.readOnlyRootFilesystem -or $mainContainer.securityContext.runAsUser -ne 65532) {
    throw "Node auditor main container security context does not match the non-root/read-only design."
}
if ($helperContainer.securityContext.runAsUser -ne 0 -or $helperContainer.securityContext.capabilities.add -notcontains "DAC_READ_SEARCH") {
    throw "Metadata helper does not have the narrowly scoped DAC_READ_SEARCH exception."
}
if ($nodePodJson.spec.automountServiceAccountToken -ne $false) {
    throw "Node DaemonSet must not mount a service-account token."
}

Write-Host "[8/8] E2E checks completed. Existing cluster '$Context' was reused and left running."
Write-Host "Image: $Image"
Write-Host "Artifacts: $Artifacts"
Write-Host "Cluster report summary: $($clusterReport.summary | ConvertTo-Json -Compress)"
Write-Host "No Kind cluster was created, switched, or deleted by this script."

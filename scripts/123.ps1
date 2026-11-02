# ============================================================
# KubeCIS - Node Metadata Initialization Container
# Kubernetes context : kind-kind
# Namespace          : kube-system
# DaemonSet          : kubecis-node-auditor
# ============================================================

$ErrorActionPreference = "Stop"

$Context    = "kind-kind"
$Namespace  = "kube-system"
$DaemonSet  = "kubecis-node-auditor"
$InitName   = "node-metadata-init"
$PatchFile  = Join-Path $env:TEMP "kubecis-init-container-patch.json"

# ------------------------------------------------------------
# 1. Verify Kubernetes context
# ------------------------------------------------------------

Write-Host "`n[1/7] Checking Kubernetes context..." -ForegroundColor Cyan

$CurrentContext = kubectl config current-context

if ($LASTEXITCODE -ne 0) {
    throw "Unable to determine current Kubernetes context."
}

if ($CurrentContext -ne $Context) {
    throw "Expected context '$Context', but current context is '$CurrentContext'."
}

Write-Host "Context: $CurrentContext" -ForegroundColor Green

# ------------------------------------------------------------
# 2. Verify existing DaemonSet
# ------------------------------------------------------------

Write-Host "`n[2/7] Checking existing DaemonSet..." -ForegroundColor Cyan

$DaemonSetJson = kubectl --context $Context get daemonset $DaemonSet `
    -n $Namespace -o json

if ($LASTEXITCODE -ne 0) {
    throw "Unable to retrieve DaemonSet '$DaemonSet'."
}

try {
    $DaemonSetObject = $DaemonSetJson | ConvertFrom-Json -ErrorAction Stop
}
catch {
    throw "Unable to parse DaemonSet JSON: $($_.Exception.Message)"
}

Write-Host "DaemonSet found: $DaemonSet" -ForegroundColor Green

# ------------------------------------------------------------
# 3. Define metadata initialization container
# ------------------------------------------------------------

Write-Host "`n[3/7] Preparing metadata initialization container..." -ForegroundColor Cyan

$InitContainer = @{
    name            = $InitName
    image           = "ghcr.io/sm-joe/kubecis:milestone-2-step-2.2-metadata-helper"
    imagePullPolicy = "IfNotPresent"

    args = @(
        "node-metadata",
        "--benchmark", "cis-2.1.0",
        "--host-root", "/host-kubelet",
        "--output-file", "/shared/node-metadata.json",
        "--node-name", '$(NODE_NAME)',
        "--interval", "0s"
    )

    env = @(
        @{
            name = "NODE_NAME"
            valueFrom = @{
                fieldRef = @{
                    fieldPath = "spec.nodeName"
                }
            }
        }
    )

    securityContext = @{
        allowPrivilegeEscalation = $false
        capabilities = @{
            add  = @("DAC_READ_SEARCH")
            drop = @("ALL")
        }
        readOnlyRootFilesystem = $true
        runAsGroup = 0
        runAsNonRoot = $false
        runAsUser = 0
        seccompProfile = @{
            type = "RuntimeDefault"
        }
    }

    volumeMounts = @(
        @{
            name      = "kubelet-directory"
            mountPath = "/host-kubelet"
            readOnly  = $true
        },
        @{
            name      = "node-metadata"
            mountPath = "/shared"
        }
    )

    resources = @{
        requests = @{
            cpu    = "10m"
            memory = "16Mi"
        }
        limits = @{
            cpu    = "50m"
            memory = "64Mi"
        }
    }
}

# ------------------------------------------------------------
# 4. Apply patch only if init container is absent
# ------------------------------------------------------------

Write-Host "`n[4/7] Checking init container..." -ForegroundColor Cyan

$ExistingInitContainers = @(
    $DaemonSetObject.spec.template.spec.initContainers |
        ForEach-Object { $_.name }
)

if ($ExistingInitContainers -contains $InitName) {

    Write-Host "$InitName already exists. Skipping patch." -ForegroundColor Yellow

}
else {

    Write-Host "$InitName does not exist. Applying patch..." -ForegroundColor Cyan

    $PatchOperations = @(
        @{
            op    = "add"
            path  = "/spec/template/spec/initContainers"
            value = @($InitContainer)
        }
    )

    $Patch = ConvertTo-Json `
        -InputObject $PatchOperations `
        -Depth 30 `
        -Compress

    try {
        $null = ConvertFrom-Json -InputObject $Patch -ErrorAction Stop
    }
    catch {
        throw "Generated JSON Patch is invalid: $($_.Exception.Message)"
    }

    if (-not $Patch.TrimStart().StartsWith("[")) {
        throw "JSON Patch must be a JSON array."
    }

    [System.IO.File]::WriteAllText(
        $PatchFile,
        $Patch,
        [System.Text.UTF8Encoding]::new($false)
    )

    try {

        kubectl --context $Context patch daemonset $DaemonSet `
            -n $Namespace `
            --type=json `
            --patch-file $PatchFile

        if ($LASTEXITCODE -ne 0) {
            throw "Failed to patch KubeCIS DaemonSet."
        }

        Write-Host "Init container added successfully." -ForegroundColor Green

    }
    finally {
        Remove-Item $PatchFile -Force -ErrorAction SilentlyContinue
    }
}

# ------------------------------------------------------------
# 5. Verify deployed init container configuration
# ------------------------------------------------------------

Write-Host "`n[5/7] Verifying deployed configuration..." -ForegroundColor Cyan

$InitContainerNames = kubectl --context $Context get daemonset $DaemonSet `
    -n $Namespace `
    -o jsonpath="{.spec.template.spec.initContainers[*].name}"

if ($LASTEXITCODE -ne 0) {
    throw "Unable to verify init container configuration."
}

if (($InitContainerNames -split "\s+") -notcontains $InitName) {
    throw "Expected init container '$InitName' was not found."
}

Write-Host "Init container present: $InitContainerNames" -ForegroundColor Green

# ------------------------------------------------------------
# 6. Wait for DaemonSet rollout
# ------------------------------------------------------------

Write-Host "`n[6/7] Waiting for DaemonSet rollout..." -ForegroundColor Cyan

kubectl --context $Context rollout status "daemonset/$DaemonSet" `
    -n $Namespace `
    --timeout=180s

if ($LASTEXITCODE -ne 0) {

    Write-Host "`nDaemonSet rollout did not complete." -ForegroundColor Red

    kubectl --context $Context get pods `
        -n $Namespace `
        -l app.kubernetes.io/name=kubecis-node-auditor `
        -o wide

    throw "KubeCIS DaemonSet rollout failed. Inspect init container state."
}

Write-Host "DaemonSet rollout completed." -ForegroundColor Green

# ------------------------------------------------------------
# 7. Display init container and application container states
# ------------------------------------------------------------

Write-Host "`n[7/7] Checking KubeCIS Pod status..." -ForegroundColor Cyan

$Pods = kubectl --context $Context get pods `
    -n $Namespace `
    -l app.kubernetes.io/name=kubecis-node-auditor `
    -o json

if ($LASTEXITCODE -ne 0) {
    throw "Unable to retrieve KubeCIS Pods."
}

$PodObjects = $Pods | ConvertFrom-Json

foreach ($Pod in $PodObjects.items) {

    Write-Host "`nPod: $($Pod.metadata.name)" -ForegroundColor Cyan
    Write-Host "Phase: $($Pod.status.phase)"

    Write-Host "`nInit containers:" -ForegroundColor Yellow

    foreach ($Container in $Pod.status.initContainerStatuses) {

        Write-Host "  Name: $($Container.name)"
        Write-Host "  Ready: $($Container.ready)"
        Write-Host "  Restarts: $($Container.restartCount)"

        if ($Container.state.terminated) {
            Write-Host "  State: Terminated"
            Write-Host "  Exit code: $($Container.state.terminated.exitCode)"
            Write-Host "  Reason: $($Container.state.terminated.reason)"
        }
        elseif ($Container.state.waiting) {
            Write-Host "  State: Waiting"
            Write-Host "  Reason: $($Container.state.waiting.reason)"
            Write-Host "  Message: $($Container.state.waiting.message)"
        }
        elseif ($Container.state.running) {
            Write-Host "  State: Running"
        }
    }

    Write-Host "`nApplication containers:" -ForegroundColor Yellow

    foreach ($Container in $Pod.status.containerStatuses) {
        Write-Host "  Name: $($Container.name)"
        Write-Host "  Ready: $($Container.ready)"
        Write-Host "  Restarts: $($Container.restartCount)"
    }
}

Write-Host "`nKubeCIS DaemonSet validation finished." -ForegroundColor Green
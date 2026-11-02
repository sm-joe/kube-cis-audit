[CmdletBinding()]
param(
    [string]$Context = "kind-kind",
    [string]$Namespace = "kube-system",
    [string]$ServiceAccount = "kubecis",
    [string]$ClusterRole = "kubecis-auditor",
    [string]$ClusterRoleBinding = "kubecis-auditor",
    [string]$TestPod = "kubecis-m32-rbac-check",
    [string]$TestImage = "registry.k8s.io/kubectl:v1.37.0",
    [string]$ReportPath,
    [string]$BaselineReportPath,
    [string]$AuditJob = "kubecis-e2e-cluster-audit"
)

$ErrorActionPreference = "Stop"
$script:Passed = 0
$script:Failed = 0
$script:Skipped = 0

function Write-Result {
    param([string]$Name, [ValidateSet("PASS","FAIL","SKIP","INFO")][string]$Status, [string]$Detail = "")
    $color = switch ($Status) {
        "PASS" { "Green" }
        "FAIL" { "Red" }
        "SKIP" { "Yellow" }
        default { "Cyan" }
    }
    Write-Host ("[{0}] {1}{2}" -f $Status, $Name, $(if ($Detail) { " - $Detail" } else { "" })) -ForegroundColor $color
    switch ($Status) {
        "PASS" { $script:Passed++ }
        "FAIL" { $script:Failed++ }
        "SKIP" { $script:Skipped++ }
    }
}

function Invoke-Kubectl {
    param([string[]]$Arguments)
    $output = & kubectl --context $Context @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "kubectl $($Arguments -join ' ') failed: $output"
    }
    return $output
}

function Read-Report {
    param([string]$Path)
    if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) {
        throw "Report file not found: $Path"
    }
    try {
        return Get-Content -LiteralPath $Path -Raw | ConvertFrom-Json
    } catch {
        throw "Invalid JSON report '$Path': $($_.Exception.Message)"
    }
}


function ConvertTo-PlainHashtable {
    param([Parameter(ValueFromPipeline=$true)]$InputObject)
    process {
        if ($null -eq $InputObject) {
            return $null
        }
        if ($InputObject -is [System.Management.Automation.PSCustomObject]) {
            $converted = @{}
            foreach ($property in $InputObject.PSObject.Properties) {
                $converted[$property.Name] = ConvertTo-PlainHashtable $property.Value
            }
            return $converted
        }
        if ($InputObject -is [System.Collections.IDictionary]) {
            $converted = @{}
            foreach ($key in $InputObject.Keys) {
                $converted[$key] = ConvertTo-PlainHashtable $InputObject[$key]
            }
            return $converted
        }
        if ($InputObject -is [System.Array]) {
            $converted = @()
            foreach ($item in $InputObject) {
                $converted += ,(ConvertTo-PlainHashtable $item)
            }
            return ,$converted
        }
        return $InputObject
    }
}

function New-AuditJobFromTemplate {
    param([string]$SourceJob, [string]$NewJobName)

    $raw = & kubectl --context $Context get job $SourceJob -n $Namespace -o json
    if ($LASTEXITCODE -ne 0) { throw "Cannot read source audit Job '$SourceJob'." }
    # ConvertFrom-Json -AsHashtable is unavailable in Windows PowerShell 5.1.
    # Convert recursively so optional Kubernetes metadata fields can be removed safely.
    $jobObject = $raw | ConvertFrom-Json
    $job = ConvertTo-PlainHashtable $jobObject

    $job.metadata["name"] = $NewJobName
    foreach ($field in @("resourceVersion","uid","creationTimestamp","generation","managedFields","ownerReferences","annotations")) {
        $job.metadata.Remove($field)
    }
    $job.Remove("status")
    $job.spec.Remove("selector")
    $job.spec.Remove("manualSelector")
    if ($job.spec.template.metadata) {
        $job.spec.template.metadata["labels"] = @{}
        $job.spec.template.metadata.Remove("annotations")
    }

    $json = $job | ConvertTo-Json -Depth 100
    $json | & kubectl --context $Context apply -f -
    if ($LASTEXITCODE -ne 0) { throw "Failed to create audit Job '$NewJobName'." }
}

function Invoke-AuditJobAndSaveReport {
    param([string]$JobName, [string]$OutputPath)

    $null = Invoke-Kubectl -Arguments @("wait","-n",$Namespace,"--for=condition=complete","job/$JobName","--timeout=240s")
    $logs = & kubectl --context $Context logs -n $Namespace "job/$JobName" --all-containers=true
    if ($LASTEXITCODE -ne 0) { throw "Could not retrieve logs from '$JobName'." }

    $rawText = ($logs -join "`n").Trim()
    $parsed = $null
    try {
        $parsed = $rawText | ConvertFrom-Json
    } catch {
        # Some images emit informational lines before the JSON document.
        $start = $rawText.IndexOf("{")
        $end = $rawText.LastIndexOf("}")
        if ($start -lt 0 -or $end -le $start) { throw "Job '$JobName' did not emit a JSON report." }
        $parsed = $rawText.Substring($start, $end - $start + 1) | ConvertFrom-Json
    }

    $parsed | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $OutputPath -Encoding utf8
    return $parsed
}

function Get-ReportSignature {
    param($Report)
    $checks = @($Report.results | ForEach-Object {
        [PSCustomObject]@{
            id = $_.id
            status = $_.status
            scored = $_.scored
        }
    } | Sort-Object id)
    [ordered]@{
        benchmark = $Report.benchmark
        clusterVersion = $Report.clusterVersion
        checks = $checks
        summary = $Report.summary
        compliance = $Report.compliance
    } | ConvertTo-Json -Depth 100 -Compress
}

Write-Host "`nKubeCIS Milestone 3.2 validation" -ForegroundColor Cyan
Write-Host "Context: $Context | Auditor: $Namespace/$ServiceAccount`n"

try {
    # 1. Confirm the target cluster is reachable.
    $null = Invoke-Kubectl -Arguments @("cluster-info")
    Write-Result "Kubernetes context reachable" "PASS" $Context
} catch {
    Write-Result "Kubernetes context reachable" "FAIL" $_.Exception.Message
    throw
}

# 2. ClusterRole and binding shape.
try {
    $role = (Invoke-Kubectl -Arguments @("get","clusterrole",$ClusterRole,"-o","json") | ConvertFrom-Json)
    $binding = (Invoke-Kubectl -Arguments @("get","clusterrolebinding",$ClusterRoleBinding,"-o","json") | ConvertFrom-Json)

    $boundCorrectly = @($binding.subjects | Where-Object {
        $_.kind -eq "ServiceAccount" -and $_.name -eq $ServiceAccount -and $_.namespace -eq $Namespace
    }).Count -gt 0
    if ($binding.roleRef.kind -ne "ClusterRole" -or $binding.roleRef.name -ne $ClusterRole -or -not $boundCorrectly) {
        Write-Result "Auditor ClusterRoleBinding" "FAIL" "Binding does not target the expected ServiceAccount and ClusterRole."
    } else {
        Write-Result "Auditor ClusterRoleBinding" "PASS" "$ClusterRoleBinding -> $Namespace/$ServiceAccount"
    }

    $crbs = (Invoke-Kubectl -Arguments @("get","clusterrolebindings","-o","json") | ConvertFrom-Json).items
    $rbs  = (Invoke-Kubectl -Arguments @("get","rolebindings","-A","-o","json") | ConvertFrom-Json).items
    $extraBindings = @()
    foreach ($b in $crbs) {
        foreach ($s in $b.subjects) {
            if ($s.kind -eq "ServiceAccount" -and $s.name -eq $ServiceAccount -and $s.namespace -eq $Namespace -and $b.metadata.name -ne $ClusterRoleBinding) {
                $extraBindings += "ClusterRoleBinding/$($b.metadata.name)"
            }
        }
    }
    foreach ($b in $rbs) {
        foreach ($s in $b.subjects) {
            if ($s.kind -eq "ServiceAccount" -and $s.name -eq $ServiceAccount -and $s.namespace -eq $Namespace) {
                $extraBindings += "RoleBinding/$($b.metadata.namespace)/$($b.metadata.name)"
            }
        }
    }
    if ($extraBindings.Count -eq 0) {
        Write-Result "No additional RBAC bindings" "PASS"
    } else {
        Write-Result "No additional RBAC bindings" "FAIL" ($extraBindings -join ", ")
    }

    $secretRules = @($role.rules | Where-Object {
        $_.resources -contains "*" -or $_.resources -contains "secrets"
    })
    if ($secretRules.Count -eq 0) {
        Write-Result "ClusterRole does not grant Secret access" "PASS"
    } else {
        Write-Result "ClusterRole does not grant Secret access" "FAIL" "A rule includes secrets or wildcard resources."
    }

    $writeVerbs = @("create","update","patch","delete","deletecollection","*")
    $writeRules = @($role.rules | Where-Object {
        ($_.resources -contains "*" -or @($_.resources | Where-Object { $_ -notin @("selfsubjectreviews","selfsubjectaccessreviews","selfsubjectrulesreviews") }).Count -gt 0) -and
        @($_.verbs | Where-Object { $_ -in $writeVerbs }).Count -gt 0
    })
    if ($writeRules.Count -eq 0) {
        Write-Result "ClusterRole has no workload/resource write verbs" "PASS"
    } else {
        Write-Result "ClusterRole has no workload/resource write verbs" "FAIL" "Write verbs found in ClusterRole rules."
    }

    $nodeRules = @($role.rules | Where-Object { $_.resources -contains "*" -or $_.resources -contains "nodes" })
    if ($nodeRules.Count -eq 0) {
        Write-Result "ClusterRole does not grant Node inventory access" "PASS"
    } else {
        Write-Result "ClusterRole does not grant Node inventory access" "FAIL" "A rule includes nodes or wildcard resources."
    }
} catch {
    Write-Result "ClusterRole and binding checks" "FAIL" $_.Exception.Message
}

# 3. Effective permissions, evaluated from a Pod running as the auditor ServiceAccount.
$podCreated = $false
try {
    $existing = & kubectl --context $Context get pod $TestPod -n $Namespace --ignore-not-found -o name 2>$null
    if ($existing) {
        $null = Invoke-Kubectl -Arguments @("delete","pod",$TestPod,"-n",$Namespace,"--wait=true")
    }

    $podYaml = @"
apiVersion: v1
kind: Pod
metadata:
  name: $TestPod
  namespace: $Namespace
  labels:
    app.kubernetes.io/name: kubecis-m32-rbac-check
spec:
  serviceAccountName: $ServiceAccount
  restartPolicy: Never
  automountServiceAccountToken: true
  containers:
    - name: check
      image: $TestImage
      imagePullPolicy: IfNotPresent
      command: ["kubectl", "auth", "can-i", "--list"]
      env:
        - name: HOME
          value: /tmp
      volumeMounts:
        - name: tmp
          mountPath: /tmp
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        runAsNonRoot: true
        runAsUser: 65532
        runAsGroup: 65532
        capabilities:
          drop: ["ALL"]
  volumes:
    - name: tmp
      emptyDir:
        sizeLimit: 16Mi
"@
    $podYaml | & kubectl --context $Context apply -f -
    if ($LASTEXITCODE -ne 0) { throw "Failed to create temporary RBAC test Pod." }
    $podCreated = $true

    $null = Invoke-Kubectl -Arguments @("wait","-n",$Namespace,"--for=jsonpath={.status.phase}=Succeeded","pod/$TestPod","--timeout=120s")
    $effective = (Invoke-Kubectl -Arguments @("logs","-n",$Namespace,$TestPod)) -join "`n"

    $resourceLines = @{}
    foreach ($line in ($effective -split "`r?`n")) {
        if ($line -match '^\s*(\S+)\s+\[\]\s+\[\]\s+\[(.*?)\]\s*$') {
            $resourceLines[$Matches[1]] = @($Matches[2] -split '\s+')
        }
    }

    $requiredResources = @(
        "pods","namespaces","serviceaccounts","services",
        "networkpolicies.networking.k8s.io","ingresses.networking.k8s.io",
        "roles.rbac.authorization.k8s.io","rolebindings.rbac.authorization.k8s.io",
        "clusterroles.rbac.authorization.k8s.io","clusterrolebindings.rbac.authorization.k8s.io",
        "validatingadmissionpolicies.admissionregistration.k8s.io",
        "validatingadmissionpolicybindings.admissionregistration.k8s.io",
        "validatingwebhookconfigurations.admissionregistration.k8s.io"
    )
    $missing = @($requiredResources | Where-Object { -not $resourceLines.ContainsKey($_) })
    if ($missing.Count -eq 0) {
        Write-Result "Effective required read permissions" "PASS" "$($requiredResources.Count) resource types present"
    } else {
        Write-Result "Effective required read permissions" "FAIL" ("Missing: " + ($missing -join ", "))
    }

    $secretGrant = @($resourceLines.GetEnumerator() | Where-Object {
        $_.Key -eq "*" -or $_.Key -match '(^|\.)secrets($|\.)'
    } | Where-Object { @($_.Value | Where-Object { $_ -in @("get","list","watch","*") }).Count -gt 0 })
    if ($secretGrant.Count -eq 0) {
        Write-Result "Effective ServiceAccount Secret access denied" "PASS"
    } else {
        Write-Result "Effective ServiceAccount Secret access denied" "FAIL" "Effective permission list includes Secret read access."
    }

    $nodeGrant = @($resourceLines.GetEnumerator() | Where-Object {
        $_.Key -eq "*" -or $_.Key -match '(^|\.)nodes($|\.)'
    } | Where-Object { @($_.Value | Where-Object { $_ -in @("get","list","watch","*") }).Count -gt 0 })
    if ($nodeGrant.Count -eq 0) {
        Write-Result "Effective ServiceAccount Node access denied" "PASS"
    } else {
        Write-Result "Effective ServiceAccount Node access denied" "FAIL" "Effective permission list includes Node read access."
    }

    $effectiveWrite = @($resourceLines.GetEnumerator() | Where-Object {
        $_.Key -notmatch '^selfsubject(review|accessreview|rulesreview)s?\.'
    } | Where-Object { @($_.Value | Where-Object { $_ -in $writeVerbs }).Count -gt 0 })
    if ($effectiveWrite.Count -eq 0) {
        Write-Result "Effective ServiceAccount workload write access denied" "PASS"
    } else {
        Write-Result "Effective ServiceAccount workload write access denied" "FAIL" "Write permissions found in effective permission list."
    }

    if ($resourceLines.ContainsKey("clustertrustbundles.certificates.k8s.io")) {
        Write-Result "ClusterTrustBundle permission review" "INFO" "Effective get/list/watch permission exists; it is not in the four declared ClusterRole rules and should be traced separately."
    } else {
        Write-Result "ClusterTrustBundle permission review" "PASS" "No unexpected ClusterTrustBundle permission observed."
    }
} catch {
    Write-Result "Effective ServiceAccount permission test" "FAIL" $_.Exception.Message
} finally {
    if ($podCreated) {
        & kubectl --context $Context delete pod $TestPod -n $Namespace --wait=true --ignore-not-found | Out-Null
    }
}

# 4. Existing E2E Job status.
try {
    $job = & kubectl --context $Context get job $AuditJob -n $Namespace -o json 2>$null
    if ($LASTEXITCODE -eq 0 -and $job) {
        $jobObject = $job | ConvertFrom-Json
        if ([int]$jobObject.status.succeeded -ge 1) {
            Write-Result "Cluster audit E2E Job" "PASS" "$AuditJob succeeded"
        } else {
            Write-Result "Cluster audit E2E Job" "FAIL" "$AuditJob has not succeeded"
        }
    } else {
        Write-Result "Cluster audit E2E Job" "SKIP" "Job '$AuditJob' not found in namespace '$Namespace'; set -AuditJob if its name differs."
    }
} catch {
    Write-Result "Cluster audit E2E Job" "FAIL" $_.Exception.Message
}

# 5. Report structure, summary consistency, scored-control consistency.
if ($ReportPath) {
    try {
        $report = Read-Report -Path $ReportPath
        $results = @($report.results)
        if ($results.Count -eq 0) { throw "Report results array is empty." }

        if ([int]$report.summary.total -ne $results.Count) {
            Write-Result "Report summary total matches results" "FAIL" "summary.total=$($report.summary.total), results=$($results.Count)"
        } else {
            Write-Result "Report summary total matches results" "PASS" "$($results.Count) results"
        }

        $duplicateIds = @($results | Group-Object id | Where-Object { $_.Count -gt 1 })
        if ($duplicateIds.Count -eq 0) {
            Write-Result "Report check IDs are unique" "PASS"
        } else {
            Write-Result "Report check IDs are unique" "FAIL" (($duplicateIds.Name) -join ", ")
        }

        $summaryMap = @{
            pass = "PASS"; fail = "FAIL"; warn = "WARN"; info = "INFO"
            manual = "MANUAL"; notAssessed = "NOT_ASSESSED"
        }
        $summaryMismatch = @()
        foreach ($key in $summaryMap.Keys) {
            $actual = @($results | Where-Object { $_.status -eq $summaryMap[$key] }).Count
            $declared = [int]$report.summary.$key
            if ($actual -ne $declared) { $summaryMismatch += "$key declared=$declared actual=$actual" }
        }
        if ($summaryMismatch.Count -eq 0) {
            Write-Result "Report status summary matches result records" "PASS"
        } else {
            Write-Result "Report status summary matches result records" "FAIL" ($summaryMismatch -join "; ")
        }

        $scored = @($results | Where-Object { $_.scored -eq $true })
        $scoredMap = @{
            scoredTotal = $null; scoredPass = "PASS"; scoredFail = "FAIL"; scoredWarn = "WARN"
            scoredManual = "MANUAL"; scoredNotAssessed = "NOT_ASSESSED"
        }
        $scoredMismatch = @()
        foreach ($key in $scoredMap.Keys) {
            if ($key -eq "scoredTotal") {
                $actual = $scored.Count
            } else {
                $actual = @($scored | Where-Object { $_.status -eq $scoredMap[$key] }).Count
            }
            $declared = [int]$report.compliance.$key
            if ($actual -ne $declared) { $scoredMismatch += "$key declared=$declared actual=$actual" }
        }
        if ($scoredMismatch.Count -eq 0) {
            Write-Result "Report scored-control totals are consistent" "PASS"
        } else {
            Write-Result "Report scored-control totals are consistent" "FAIL" ($scoredMismatch -join "; ")
        }

        if ($report.benchmarkCompatibilityStatus -eq "provisional") {
            Write-Result "Benchmark compatibility disclosure" "PASS" "Report explicitly marks mapping provisional."
        } else {
            Write-Result "Benchmark compatibility disclosure" "INFO" "Compatibility status: $($report.benchmarkCompatibilityStatus)"
        }
    } catch {
        Write-Result "Report integrity validation" "FAIL" $_.Exception.Message
    }
} else {
    Write-Result "Report integrity validation" "SKIP" "Supply -ReportPath <report.json>."
}

# 6. Repeatability: clone the existing successful audit Job and compare stable report fields.
$repeatJob = "kubecis-m32-repeat"
$stressJob = "kubecis-m32-stress"
$repeatPath = Join-Path (Get-Location) "kubecis-m32-repeat-report.json"
$stressPath = Join-Path (Get-Location) "kubecis-m32-stress-report.json"
$fixtureNamespace = "kubecis-m32-load"
$fixtureCreated = $false

try {
    if (-not $ReportPath) { throw "A baseline -ReportPath is required for repeatability and stress validation." }
    $baseline = Read-Report -Path $ReportPath

    $null = & kubectl --context $Context delete job $repeatJob -n $Namespace --ignore-not-found --wait=true
    New-AuditJobFromTemplate -SourceJob $AuditJob -NewJobName $repeatJob
    $repeat = Invoke-AuditJobAndSaveReport -JobName $repeatJob -OutputPath $repeatPath

    if ((Get-ReportSignature $baseline) -ceq (Get-ReportSignature $repeat)) {
        Write-Result "Repeatability (independent second scan)" "PASS" "Check IDs, statuses, summary and compliance are identical."
    } else {
        Write-Result "Repeatability (independent second scan)" "FAIL" "Stable check statuses, summary or compliance changed between scans."
    }
} catch {
    Write-Result "Repeatability (independent second scan)" "FAIL" $_.Exception.Message
}

# 7. Large inventory: create 50 isolated Pod fixtures, scan, verify completion/report, then clean up.
try {
    $null = & kubectl --context $Context delete namespace $fixtureNamespace --ignore-not-found --wait=true
    $null = Invoke-Kubectl -Arguments @("create","namespace",$fixtureNamespace)

    $documents = for ($i = 1; $i -le 50; $i++) {
        $podName = "inventory-pod-{0:D2}" -f $i
        @"
apiVersion: v1
kind: Pod
metadata:
  name: $podName
  namespace: $fixtureNamespace
  labels:
    app.kubernetes.io/name: kubecis-inventory-fixture
spec:
  restartPolicy: Never
  automountServiceAccountToken: false
  containers:
    - name: fixture
      image: registry.k8s.io/pause:3.10
      imagePullPolicy: IfNotPresent
      securityContext:
        privileged: true
"@
    }
    ($documents -join "`n---`n") | & kubectl --context $Context apply -f -
    if ($LASTEXITCODE -ne 0) { throw "Failed to create the 50-Pod inventory fixture." }
    $fixtureCreated = $true

    $createdPods = (Invoke-Kubectl -Arguments @("get","pods","-n",$fixtureNamespace,"-o","json") | ConvertFrom-Json).items.Count
    if ($createdPods -ne 50) { throw "Expected 50 fixture Pods, found $createdPods." }

    $null = & kubectl --context $Context delete job $stressJob -n $Namespace --ignore-not-found --wait=true
    New-AuditJobFromTemplate -SourceJob $AuditJob -NewJobName $stressJob
    $stress = Invoke-AuditJobAndSaveReport -JobName $stressJob -OutputPath $stressPath

    if (@($stress.results).Count -gt 0 -and [int]$stress.summary.total -gt 0) {
        Write-Result "Large-inventory scan (50 fixture Pods)" "PASS" "Audit Job completed and emitted $($stress.summary.total) report results."
    } else {
        Write-Result "Large-inventory scan (50 fixture Pods)" "FAIL" "Stress scan report was empty."
    }
} catch {
    Write-Result "Large-inventory scan (50 fixture Pods)" "FAIL" $_.Exception.Message
} finally {
    if ($fixtureCreated) {
        & kubectl --context $Context delete namespace $fixtureNamespace --wait=true --ignore-not-found | Out-Null
    }
    & kubectl --context $Context delete job $repeatJob -n $Namespace --wait=true --ignore-not-found | Out-Null
    & kubectl --context $Context delete job $stressJob -n $Namespace --wait=true --ignore-not-found | Out-Null
}

Write-Host "Repeatability report: $repeatPath"
Write-Host "Stress report:        $stressPath"

Write-Host "`nSummary: PASS=$script:Passed FAIL=$script:Failed SKIP=$script:Skipped" -ForegroundColor Cyan
if ($script:Failed -gt 0) { exit 1 }
exit 0

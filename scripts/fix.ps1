# KubeCIS - Correct ClusterRoleBinding ServiceAccount

$Context = "kind-kind"
$Namespace = "kube-system"
$Binding = "kubecis-auditor"
$ClusterRole = "kubecis-auditor"
$ServiceAccount = "kubecis"

Write-Host "`nChecking ServiceAccount..." -ForegroundColor Cyan

kubectl --context $Context get serviceaccount $ServiceAccount -n $Namespace

if ($LASTEXITCODE -ne 0) {
    throw "KubeCIS ServiceAccount not found."
}

Write-Host "`nUpdating ClusterRoleBinding..." -ForegroundColor Cyan

kubectl --context $Context patch clusterrolebinding $Binding `
    --type=merge `
    -p '{"subjects":[{"kind":"ServiceAccount","name":"kubecis","namespace":"kube-system"}]}'

if ($LASTEXITCODE -ne 0) {
    throw "Failed to update KubeCIS ClusterRoleBinding."
}

Write-Host "`nVerifying ClusterRoleBinding..." -ForegroundColor Cyan

kubectl --context $Context get clusterrolebinding $Binding -o yaml

if ($LASTEXITCODE -ne 0) {
    throw "Unable to verify ClusterRoleBinding."
}

Write-Host "`nChecking effective permissions..." -ForegroundColor Cyan

$Identity = "system:serviceaccount:kube-system:kubecis"

$Checks = @(
    @{ Verb = "get"; Resource = "pods" },
    @{ Verb = "list"; Resource = "pods" },
    @{ Verb = "get"; Resource = "namespaces" },
    @{ Verb = "list"; Resource = "namespaces" },
    @{ Verb = "get"; Resource = "services" },
    @{ Verb = "list"; Resource = "services" },
    @{ Verb = "get"; Resource = "secrets" },
    @{ Verb = "list"; Resource = "secrets" }
)

foreach ($Check in $Checks) {

    $Result = kubectl --context $Context auth can-i `
        $Check.Verb $Check.Resource --as=$Identity

    if ($LASTEXITCODE -ne 0) {
        throw "Authorization check could not be completed."
    }

    Write-Host ("{0,-6} {1,-15} {2}" -f `
        $Check.Verb, $Check.Resource, $Result)
}

Write-Host "`nRBAC validation completed." -ForegroundColor Green
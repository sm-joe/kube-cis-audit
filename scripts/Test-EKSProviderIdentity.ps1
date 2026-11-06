# PowerShell contract regression: reject wrong EKS endpoints, ARN, and region.
$ErrorActionPreference = 'Stop'
$root = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root -Force | Out-Null
try {
    function aws {
        $global:LASTEXITCODE = 0
        $response = [ordered]@{
            cluster = [ordered]@{
                name = 'kubecis-eks-test'
                arn = 'arn:aws:eks:ap-south-1:123456789012:cluster/kubecis-eks-test'
                endpoint = 'https://ABCDEF.eks.ap-south-1.amazonaws.com'
                status = 'ACTIVE'
                version = '1.31'
                resourcesVpcConfig = @{ endpointPrivateAccess = $true; endpointPublicAccess = $true; publicAccessCidrs = @('0.0.0.0/0') }
                logging = @{ clusterLogging = @(
                    @{ types = @('api', 'audit', 'authenticator', 'controllerManager', 'scheduler'); enabled = $false }
                ) }
            }
        }
        return ($response | ConvertTo-Json -Depth 10)
    }
    $collector = Join-Path $PSScriptRoot 'Collect-EKSProviderEvidence.ps1'
    $output = Join-Path $root 'provider.json'
    & $collector -ClusterName 'kubecis-eks-test' -Region 'ap-south-1' -ExpectedAPIServer 'https://ABCDEF.eks.ap-south-1.amazonaws.com' -OutputFile $output
    $report = Get-Content $output -Raw | ConvertFrom-Json
    if ($report.clusterArn -ne 'arn:aws:eks:ap-south-1:123456789012:cluster/kubecis-eks-test') {
        throw 'EKS ARN missing from provider evidence'
    }
    if ((Get-Content $output -Raw) -match 'ABCDEF.eks.ap-south-1.amazonaws.com') {
        throw 'EKS API endpoint leaked into redacted provider report'
    }
    $rejected = $false
    try {
        & $collector -ClusterName 'kubecis-eks-test' -Region 'ap-south-1' -ExpectedAPIServer 'https://OTHER.eks.ap-south-1.amazonaws.com' -OutputFile $output | Out-Null
    } catch {
        $rejected = $_.Exception.Message -match 'identity mismatch'
    }
    if (-not $rejected) { throw 'Mismatched kubeconfig EKS API endpoint was accepted' }
    $rejected = $false
    try {
        & $collector -ClusterName 'kubecis-eks-test' -Region 'us-east-1' -ExpectedAPIServer 'https://ABCDEF.eks.ap-south-1.amazonaws.com' -OutputFile $output | Out-Null
    } catch {
        $rejected = $_.Exception.Message -match 'ARN name/region'
    }
    if (-not $rejected) { throw 'Mismatched AWS region was accepted' }
    Write-Host 'PASS: EKS context endpoint and ARN identity contract'
} finally {
    Remove-Item -Path $root -Recurse -Force -ErrorAction SilentlyContinue
}

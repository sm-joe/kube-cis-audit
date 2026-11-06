# Regression contract: preview release gate must accept the full six AWS
# findings, reject mismatched provider counts, and never upgrade to full CIS.
$ErrorActionPreference = 'Stop'
$root = Join-Path ([System.IO.Path]::GetTempPath()) ('kubecis-release-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root -Force | Out-Null
try {
    $node = 'ip-10-10-1-65.ec2.internal'
    $when = [DateTimeOffset]::UtcNow.ToString('o')
    $arn = 'arn:aws:eks:us-east-1:123456789012:cluster/synthetic-contract'
    $results = @()
    foreach ($id in @('2.1.1','5.4.1','5.4.2')) {
        $results += [ordered]@{ id=$id; method='aws_eks_describe_cluster'; status='FAIL'; scored=$true; evidence=@(@{resource='AWS::EKS::Cluster';name=$arn;field='x';observed='false';expected='true'}) }
    }
    foreach ($id in @('4.1.7','5.5.1')) {
        $results += [ordered]@{ id=$id; method='aws_eks_access_entries'; status='MANUAL'; scored=$true; evidence=@() }
    }
    $results += [ordered]@{ id='5.1.1'; method='aws_ecr_image_scanning'; status='MANUAL'; scored=$true; evidence=@() }
    foreach ($id in @('3.1.3','3.2.1','3.2.2')) {
        $results += [ordered]@{ id=$id; scope=$node; method='host_kubelet_config_field'; target='node'; status='PASS'; scored=$true; evidence=@(@{name='etc/kubernetes/kubelet/config.json';resource='host-file';observed='true';expected='true'}) }
    }
    $report = [ordered]@{
        auditMode='consolidated'; cloudProvider='eks'; benchmark='cis-eks-2.1.0'
        benchmarkCompatibilityStatus='provisional';clusterVersion='v1.35.9-eks-577a51e'
        compliance=@{compliancePercent=$null;scoreStatus='withheld: incomplete official CIS EKS 2.1.0 control coverage';scoredTotal=9;scoredPass=3;scoredFail=3;coveragePercent=66.67}
        expectedNodes=@($node); expectedControlPlaneNodes=@()
        sources=@(
            @{auditMode='native';generatedAt=$when;findings=0}
            @{auditMode='node-local';nodeName=$node;generatedAt=$when;findings=3}
            @{auditMode='aws-provider';generatedAt=$when;findings=6}
        )
        results=$results; summary=@{total=$results.Count}
        providerEvidence=@{schemaVersion=1;cloudProvider='eks';assessmentStatus='INVENTORY_ONLY';generatedAt=$when
            clusterName='synthetic-contract';region='us-east-1';clusterArn=$arn;kubernetesVersion='1.35'
            controlPlaneLogging=@{api=$false;audit=$false;authenticator=$false;controllerManager=$false;scheduler=$false}
            accessInventoryStatus='COLLECTED';ecrInventoryStatus='COLLECTED'
        }
    }
    $path = Join-Path $root 'report.json'
    function Write-Report { $report | ConvertTo-Json -Depth 14 | Set-Content -LiteralPath $path -Encoding UTF8 }
    $validator = Join-Path $PSScriptRoot 'Validate-EKSRelease.ps1'
    Write-Report
    & $validator -ReportPath $path -ReleaseLevel Preview -ExpectedClusterName synthetic-contract -ExpectedRegion us-east-1 | Out-Null
    # Extra unmatched AWS sources should not be accepted.
    $report.sources[2].findings = 5
    Write-Report
    $rejected = $false
    try { & $validator -ReportPath $path -ReleaseLevel Preview -ExpectedClusterName synthetic-contract -ExpectedRegion us-east-1 | Out-Null } catch {
        $rejected = $_.Exception.Message -like '*provider source integrity mismatch*'
    }
    if (-not $rejected) { throw 'Release gate accepted incorrect AWS source findings count' }
    $report.sources[2].findings = 6
    # Duplicate the official AWS ID while preserving the length.
    $report.results[3].id = '2.1.1'
    Write-Report
    $rejected = $false
    try { & $validator -ReportPath $path -ReleaseLevel Preview -ExpectedClusterName synthetic-contract -ExpectedRegion us-east-1 | Out-Null } catch {
        $rejected = $_.Exception.Message -like '*duplicate official EKS AWS provider control*'
    }
    if (-not $rejected) { throw 'Release gate accepted duplicate AWS control ID' }
    $report.results[3].id = '4.1.7'
    Write-Report
    $rejected = $false
    try { & $validator -ReportPath $path -ReleaseLevel Full -ExpectedClusterName synthetic-contract -ExpectedRegion us-east-1 | Out-Null } catch {
        $rejected = $_.Exception.Message -like '*Full CIS EKS release NOT ACCEPTED*'
    }
    if (-not $rejected) { throw 'Release gate incorrectly certified official CIS completeness' }
    Write-Host 'PASS: EKS release contract accepts valid six-control AWS source; rejects mismatches, duplicates and false full certification'
} finally {
    Remove-Item -LiteralPath $root -Force -Recurse -ErrorAction SilentlyContinue
}

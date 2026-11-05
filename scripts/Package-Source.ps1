[CmdletBinding()]
param(
  [string]$Version = '3.1.0-native-integrated-rc3.2',
  [string]$OutputDirectory = '.\dist'
)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$out = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($OutputDirectory)
$stage = Join-Path ([System.IO.Path]::GetTempPath()) ('kubecis-source-' + [guid]::NewGuid().ToString('N'))
$archive = Join-Path $out ("KubeCIS-Native-$Version-source.zip")
$topFiles = @('.dockerignore','.gitignore','go.mod','go.sum','Dockerfile','Makefile','README.md','CHANGELOG.md','ATTRIBUTION.md','install.ps1','RELEASE-READINESS-NATIVE-RC3.md','SOURCE-CLEANUP.md')
$packages = @('adapter','admission','apirbac','audit','conformance','controlplane','etcd','worker','node','internal/audit','cmd/kubecis','testdata','e2e')
$allowedFiles = @(
  'benchmarks/cis-2.0.1/benchmark.yaml','benchmarks/cis-2.1.0/benchmark.yaml',
  '.github/workflows/ci.yml','.github/workflows/release.yml',
  'scripts/Build-KubeCIS.ps1','scripts/Collect-ConsolidatedKindReport.ps1',
  'scripts/Deploy-KindNative.ps1','scripts/Package-Source.ps1',
  'scripts/e2e/Invoke-KubeCisE2E.ps1','scripts/e2e/Invoke-ControlPlaneMetadataE2E.ps1',
  'deploy/rbac.yaml','deploy/job.yaml','deploy/kustomization.yaml',
  'deploy/create-ghcr-pull-secret.ps1','deploy/node-auditor-daemonset.yaml',
  'deploy/controlplane-auditor-daemonset.yaml',
  'deploy/kind-native/job.yaml','deploy/kind-native/node-auditor-daemonset.yaml',
  'deploy/kind-native/controlplane-auditor-daemonset.yaml',
  'docs/CONSOLIDATED-RELEASE.md','docs/CONTROLPLANE-METADATA-COLLECTOR.md',
  'docs/INTEGRATED-KIND-RELEASE.md','docs/NATIVE-VALIDATION.md',
  'docs/RELEASE-SCOPE.md','docs/SECURITY.md',
  'docs/PRODUCTION-CIS-RELEASE-BLOCKERS.md'
)
try {
  New-Item -ItemType Directory -Force -Path $stage,$out | Out-Null
  $files = Get-ChildItem -LiteralPath $repo -Recurse -File -Force
  foreach ($file in $files) {
    $rel = [System.IO.Path]::GetRelativePath($repo,$file.FullName).Replace('\','/')
    $approved = $topFiles -contains $rel -or $allowedFiles -contains $rel
    if (-not $approved) {
      foreach ($pkg in $packages) {
        if ($rel.StartsWith("$pkg/",[System.StringComparison]::OrdinalIgnoreCase)) {
          $approved = $true
          break
        }
      }
    }
    if (-not $approved) { continue }
    if ($rel -match '(?i)(?:^|/)(?:kubebench[^/]*\.go|.*\.exe|.*\.tar|.*\.zip|.*\.tfstate|.*\.tfvars)$') { continue }
    $dest = Join-Path $stage ($rel.Replace('/',[System.IO.Path]::DirectorySeparatorChar))
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dest) | Out-Null
    Copy-Item -LiteralPath $file.FullName -Destination $dest
  }
  if (Test-Path $archive) { Remove-Item -LiteralPath $archive -Force }
  Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $archive -CompressionLevel Optimal
  $sha = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
  Set-Content -LiteralPath "$archive.sha256" -Value "$sha  $(Split-Path -Leaf $archive)" -Encoding ascii
  Write-Host "Created $archive"
  Write-Host "SHA-256: $sha"
} finally {
  if (Test-Path $stage) { Remove-Item -LiteralPath $stage -Recurse -Force }
}

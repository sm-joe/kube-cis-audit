$ErrorActionPreference = 'Stop'

$Version = 'v0.16.0'
$RepoRoot = Split-Path -Parent $PSScriptRoot
$BinDir = Join-Path $RepoRoot 'bin'
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$PreviousGobin = $env:GOBIN
try {
    $env:GOBIN = $BinDir
    go install "github.com/aquasecurity/kube-bench@$Version"
    if ($LASTEXITCODE -ne 0) {
        throw "go install github.com/aquasecurity/kube-bench@$Version failed with exit code $LASTEXITCODE"
    }
}
finally {
    $env:GOBIN = $PreviousGobin
}

$ModuleCache = (go env GOMODCACHE).Trim()
$ModuleDir = Join-Path $ModuleCache 'github.com/aquasecurity/kube-bench@v0.16.0'
$UpstreamConfig = Join-Path $ModuleDir 'cfg'
$Destination = Join-Path $RepoRoot 'benchmarks/upstream/kube-bench/cfg'
if (-not (Test-Path (Join-Path $UpstreamConfig 'config.yaml'))) {
    throw "Pinned upstream config not found at $UpstreamConfig"
}
if (Test-Path $Destination) {
    Remove-Item -Recurse -Force $Destination
}
Copy-Item -Recurse $UpstreamConfig $Destination

$SourceDestination = Join-Path $RepoRoot 'third_party/kube-bench/source'
if (Test-Path $SourceDestination) {
    Remove-Item -Recurse -Force $SourceDestination
}
New-Item -ItemType Directory -Force -Path $SourceDestination | Out-Null
Get-ChildItem -LiteralPath $ModuleDir -Force | Copy-Item -Destination $SourceDestination -Recurse -Force

$LicenseDestination = Join-Path $RepoRoot 'benchmarks/upstream/kube-bench'
Copy-Item (Join-Path $ModuleDir 'LICENSE') $LicenseDestination -Force
if (Test-Path (Join-Path $ModuleDir 'NOTICE')) {
    Copy-Item (Join-Path $ModuleDir 'NOTICE') $LicenseDestination -Force
}

Write-Host "Installed kube-bench $Version to $BinDir"
Write-Host "Imported matching upstream cfg/ into $Destination"
Write-Host "Vendored pinned upstream source into $SourceDestination"
$BinaryName = if ($env:OS -eq 'Windows_NT') { 'kube-bench.exe' } else { 'kube-bench' }
& (Join-Path $BinDir $BinaryName) --help | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "The installed kube-bench binary failed its --help check with exit code $LASTEXITCODE"
}

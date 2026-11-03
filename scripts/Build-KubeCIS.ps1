[CmdletBinding()]
param(
    [string]$Version = "3.0.0-dev",
    [string]$OutputDirectory = ".\dist"
)

$ErrorActionPreference = "Stop"
$RepositoryRoot = Split-Path -Parent $PSScriptRoot
Push-Location $RepositoryRoot
try {
    $GoVersionOutput = & go version
    if ($LASTEXITCODE -ne 0) { throw "Go is not installed or is not available on PATH." }
    $VersionMatch = [regex]::Match(($GoVersionOutput -join " "), "go(\d+\.\d+(?:\.\d+)?)")
    if (-not $VersionMatch.Success -or [version]$VersionMatch.Groups[1].Value -lt [version]"1.26.0") {
        throw "KubeCIS requires Go 1.26.0 or later. Detected: $GoVersionOutput"
    }
    Write-Host "Using $GoVersionOutput" -ForegroundColor Cyan

    $GoFiles = @(Get-ChildItem -Path .\cmd, .\internal -Filter *.go -Recurse | ForEach-Object { $_.FullName })
    if ($GoFiles.Count -gt 0) {
        & gofmt -w $GoFiles
        if ($LASTEXITCODE -ne 0) { throw "gofmt failed." }
    }

    Write-Host "Running go vet..." -ForegroundColor Cyan
    & go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "go vet failed." }

    Write-Host "Running complete Go test suite..." -ForegroundColor Cyan
    & go test ./... -count=1
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed." }

    $ResolvedOutputDirectory = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($OutputDirectory)
    New-Item -ItemType Directory -Path $ResolvedOutputDirectory -Force | Out-Null
    $BinaryPath = Join-Path $ResolvedOutputDirectory "kubecis.exe"
    $LdFlags = "-s -w -X github.com/sm-joe/kube-cis-audit/internal/audit.BuildVersion=$Version"
    Write-Host "Building $Version..." -ForegroundColor Cyan
    & go build -trimpath -ldflags $LdFlags -o $BinaryPath .\cmd\kubecis
    if ($LASTEXITCODE -ne 0) { throw "KubeCIS build failed." }

    # Cobra's root --version output includes the application name
    # (for example: "kubecis version 3.0.0-dev"). The explicit
    # "version" subcommand intentionally returns only the version value.
    $BuiltVersion = (& $BinaryPath version | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or $BuiltVersion -ne $Version) {
        throw "Built binary version mismatch. Expected '$Version', got '$BuiltVersion'."
    }

    Write-Host "KubeCIS build and validation succeeded." -ForegroundColor Green
    Write-Host "Binary: $BinaryPath"
    Write-Host "Run: & `"$BinaryPath`" scan --benchmark auto --target all --output json --output-file .\kubecis-report.json"
} finally {
    Pop-Location
}

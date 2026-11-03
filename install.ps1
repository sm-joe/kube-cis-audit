[CmdletBinding()]
param(
    [string]$Version = "latest",
    [string]$InstallDir = (Join-Path $HOME ".kubecis\bin")
)

$ErrorActionPreference = "Stop"
$Repository = "sm-joe/kube-cis-audit"
$AssetName = "kubecis_windows_amd64.zip"

if ($Version -eq "latest") {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest" -Headers @{ "User-Agent" = "KubeCIS-Installer" }
} else {
    $Tag = $Version
    if (-not $Tag.StartsWith("v")) { $Tag = "v$Tag" }
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/tags/$Tag" -Headers @{ "User-Agent" = "KubeCIS-Installer" }
}

$Asset = $Release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1
if (-not $Asset) {
    throw "Release '$($Release.tag_name)' does not contain $AssetName. Check the GitHub release assets."
}

$TempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString("N"))
$ArchivePath = Join-Path $TempRoot $AssetName
$ExtractPath = Join-Path $TempRoot "extracted"
try {
    New-Item -ItemType Directory -Path $ExtractPath -Force | Out-Null
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Invoke-WebRequest -Uri $Asset.browser_download_url -OutFile $ArchivePath -UseBasicParsing

    $ChecksumAsset = $Release.assets | Where-Object { $_.name -eq "SHA256SUMS" } | Select-Object -First 1
    if (-not $ChecksumAsset) { throw "Release is missing SHA256SUMS; refusing to install an unverifiable archive." }
    $ChecksumPath = Join-Path $TempRoot "SHA256SUMS"
    Invoke-WebRequest -Uri $ChecksumAsset.browser_download_url -OutFile $ChecksumPath -UseBasicParsing
    $ChecksumLine = Get-Content -LiteralPath $ChecksumPath | Where-Object { $_ -match "\s+$([regex]::Escape($AssetName))$" } | Select-Object -First 1
    if (-not $ChecksumLine -or $ChecksumLine -notmatch "^([A-Fa-f0-9]{64})\s+\*?.+$") {
        throw "No valid SHA-256 entry found for $AssetName."
    }
    $ExpectedHash = $Matches[1].ToLowerInvariant()
    $ActualHash = (Get-FileHash -LiteralPath $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($ActualHash -ne $ExpectedHash) { throw "SHA-256 verification failed for $AssetName." }

    Expand-Archive -LiteralPath $ArchivePath -DestinationPath $ExtractPath -Force

    $Binary = Join-Path $ExtractPath "kubecis.exe"
    $Benchmarks = Join-Path $ExtractPath "benchmarks"
    if (-not (Test-Path $Binary) -or -not (Test-Path $Benchmarks)) {
        throw "Release archive is missing kubecis.exe or the benchmarks directory."
    }

    Copy-Item -LiteralPath $Binary -Destination (Join-Path $InstallDir "kubecis.exe") -Force
    if (Test-Path (Join-Path $InstallDir "benchmarks")) {
        Remove-Item (Join-Path $InstallDir "benchmarks") -Recurse -Force
    }
    Copy-Item -LiteralPath $Benchmarks -Destination $InstallDir -Recurse -Force
    if (Test-Path (Join-Path $ExtractPath "README.md")) {
        Copy-Item (Join-Path $ExtractPath "README.md") -Destination $InstallDir -Force
    }

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $PathEntries = @($UserPath -split ";" | Where-Object { $_ })
    if ($PathEntries -notcontains $InstallDir) {
        $NewPath = (@($PathEntries) + $InstallDir) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    }
    if (($env:Path -split ";") -notcontains $InstallDir) {
        $env:Path = "$InstallDir;$env:Path"
    }

    Write-Host "KubeCIS $($Release.tag_name) installed to $InstallDir" -ForegroundColor Green
    Write-Host "Run: kubecis --version"
    Write-Host "Run: kubecis scan --benchmark auto --target all --output json --output-file .\kubecis-report.json"
} finally {
    if (Test-Path $TempRoot) { Remove-Item $TempRoot -Recurse -Force }
}

[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $Root
try {
    Write-Host '== KubeCIS B1: format =='
    gofmt -w ./upstream ./cmd/kubecis-b1-validate
    if ($LASTEXITCODE -ne 0) { throw "gofmt failed with exit code $LASTEXITCODE" }

    # Deliberately test only B1 packages. Do not use `go test ./...` here:
    # when B1 is placed in an existing KubeCIS checkout, that would also
    # compile the existing CLI/audit engine and its unrelated dependencies.
    Write-Host '== KubeCIS B1: tests =='
    go test ./upstream ./cmd/kubecis-b1-validate
    if ($LASTEXITCODE -ne 0) { throw "B1 Go tests failed with exit code $LASTEXITCODE" }

    Write-Host '== KubeCIS B1: build =='
    go build ./upstream ./cmd/kubecis-b1-validate
    if ($LASTEXITCODE -ne 0) { throw "B1 Go build failed with exit code $LASTEXITCODE" }

    Write-Host 'B1 standalone implementation: PASS'
}
finally { Pop-Location }

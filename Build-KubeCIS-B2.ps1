$ErrorActionPreference = 'Stop'
$env:GOTOOLCHAIN = 'local'
Write-Host '== KubeCIS B2: format =='
go fmt ./...
if ($LASTEXITCODE -ne 0) { throw "gofmt failed with exit code $LASTEXITCODE" }
Write-Host '== KubeCIS B2: tests =='
go test ./...
if ($LASTEXITCODE -ne 0) { throw "Go tests failed with exit code $LASTEXITCODE" }
Write-Host '== KubeCIS B2: vet =='
go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet failed with exit code $LASTEXITCODE" }
Write-Host 'B2 validation completed successfully.'

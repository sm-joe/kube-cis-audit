# Milestone 1 build fix 4

This patch aligns the Go module path with the repository import path:
`github.com/sm-joe/kube-cis-audit`.

It also removes an unused `path/filepath` import from `internal/audit/runner_test.go`.

After extracting, run commands from the directory containing this `go.mod`:

```powershell
go clean -testcache
go test ./...
go build -o kubecis.exe ./cmd/kubecis
```

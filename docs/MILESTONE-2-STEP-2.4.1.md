# Milestone 2.4.1 — Effective Security-Context Resolution

## Scope

This step introduces a pure resolver for the following Pod/container security-context fields:

- `runAsNonRoot`
- `runAsUser`
- `runAsGroup`
- `fsGroup`

It does not add a benchmark check, change benchmark YAML, or change report output.

## Resolution rules

- `runAsNonRoot`, `runAsUser`, and `runAsGroup` are initialized from `PodSecurityContext`.
- A non-nil container-level value overrides only the corresponding Pod-level value.
- `fsGroup` is copied from `PodSecurityContext` only. It is not a container-level field and is not inferred.
- Nil values remain nil. The resolver does not infer image defaults, runtime identities, or filesystem ownership.
- Explicit zero values are preserved; callers decide how those values should be assessed.

## Implementation

`internal/audit/runner.go` defines `effectiveSecurityContext` and
`resolveEffectiveSecurityContext()`. The existing non-root evaluator now uses
this resolver for `runAsNonRoot` and `runAsUser`. Existing seccomp resolution
and evaluation are unchanged.

`internal/audit/runner_test.go` tests nil contexts, Pod inheritance, independent
container overrides, partial overrides, Pod-only `fsGroup`, and explicit zero
values overriding Pod defaults. Existing non-root and seccomp tests remain.

## Validation

Run from the repository root with Go 1.26.0:

```powershell
gofmt -w .\internal\audit\runner.go .\internal\audit\runner_test.go
go test ./internal/audit
go test ./...
```

This package was prepared in an environment with Go 1.23.2, while `go.mod`
requires Go 1.26.0. Therefore the Go test commands could not be executed here.
Run them in the project's Go 1.26.0 Windows environment before merging.

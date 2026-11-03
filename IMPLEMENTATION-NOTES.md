# B1 implementation notes

## Contract

The parser keeps the complete YAML document in `yaml.Node` (`Config.Root`, `BenchmarkFile.Root`) and also exposes commonly used typed fields. This is intentional: kube-bench configuration evolves, and B1 must not discard unknown source definitions.

## Not implemented in B1

- Execution of `audit` strings or test sources
- Comparator evaluation
- Kubernetes API or node access
- Mapping upstream controls to KubeCIS evaluators
- Changes to KubeCIS report types or scan status

## Integration seam

A future integration can call `upstream.LoadConfig`, `upstream.ResolveProfile`, and `upstream.LoadBenchmark`, then translate the parsed upstream structures into an internal normalized representation. That translation belongs in a later integration step; this module does not alter the existing audit engine.

## Corrective release (B1 parser/build isolation)

- `flag` test sources now accept both scalar (`flag: --name`) and mapping forms.
- The PowerShell build script tests/builds only `./upstream` and `./cmd/kubecis-b1-validate`.
- It intentionally does not use `go test ./...`, because that would compile unrelated existing KubeCIS packages when this implementation is placed in the repository root.

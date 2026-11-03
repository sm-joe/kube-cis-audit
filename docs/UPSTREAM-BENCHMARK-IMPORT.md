# Upstream kube-bench benchmark import

KubeCIS uses the versioned configuration files from the upstream
[`aquasecurity/kube-bench`](https://github.com/aquasecurity/kube-bench)
repository as its benchmark-definition source. KubeCIS remains an independent
Go audit engine.

## Pinned source

The import script currently pins:

- Repository: `https://github.com/aquasecurity/kube-bench.git`
- Release: `v0.16.0`
- Expected commit prefix: `5c6c22d`
- Imported source: `cfg/`
- License: upstream `LICENSE`, copied alongside the imported configuration

The pin is deliberate. Do not change the release or commit as part of an
ordinary scan. Review upstream release notes, configuration changes, and
license/attribution requirements before updating the pin.

## Import and inventory (Linux)

Run from the KubeCIS repository root:

```bash
./scripts/update-kube-bench-benchmarks.sh
```

The script requires Git and the Go version declared in `go.mod`. It checks out
the pinned release, verifies the commit, imports `cfg/` and the upstream
license into `benchmarks/upstream/kube-bench/`, and generates:

- `docs/UPSTREAM-BENCHMARK-INVENTORY.md`
- `docs/UPSTREAM-BENCHMARK-INVENTORY.json`

Existing KubeCIS profiles under `benchmarks/cis-*` are not overwritten.

To regenerate the inventory without fetching upstream:

```bash
go run ./cmd/benchmark-inventory \
  --root benchmarks/upstream/kube-bench/cfg \
  --out docs/UPSTREAM-BENCHMARK-INVENTORY.md \
  --json docs/UPSTREAM-BENCHMARK-INVENTORY.json

go test ./cmd/benchmark-inventory -count=1
```

## Safety boundary

The importer and inventory generator only read YAML. They do not execute
`audit`, `audit_config`, `audit_env`, remediation text, shell commands, or any
other upstream command string.

Upstream YAML is benchmark metadata, not trusted executable code. KubeCIS must
translate supported test definitions into native, reviewed Go evaluators.
Unsupported or unimplemented checks must remain explicitly unsupported or
`NOT_ASSESSED`; they must never be reported as `PASS` merely because an
upstream command exists.

## Inventory scope

The inventory tool currently includes directories named `cis-*` and YAML files
other than each profile's `config.yaml`. It traverses nested groups and
records control IDs, titles, target, scoring metadata, source file, audit
command presence, test-item source types, comparison operators and remediation.
It deliberately excludes EKS/GKE/AKS and other vendor profiles from this
initial CIS-only inventory. Those require a separate scope and applicability
decision.

The inventory is structural. It does not prove that the upstream controls
match every licensed CIS benchmark requirement, that a control is applicable
to a particular distribution, or that KubeCIS can safely implement the test.

## Next phase

Use the generated inventory to build the control-to-evaluator compatibility
matrix. Only after that review should KubeCIS add the upstream parser/normalized
model and native evaluators. Do not replace the current runtime benchmark
loader in this import phase.


## Empty upstream target files

Some kube-bench CIS profiles contain target YAML files with an empty YAML
document because that target has no controls in that profile. The inventory
generator treats an empty document as a valid target with zero controls.
A non-empty YAML document that omits the required top-level `controls`
mapping is still rejected as a schema error.

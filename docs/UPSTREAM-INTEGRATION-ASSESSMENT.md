# KubeCIS Upstream Benchmark Integration Assessment

> **Historical assessment (2026-11-03).** This document describes the pre-execution-engine
> state of the earlier `kube-cis-audit(5).zip` snapshot. The pinned upstream engine
> integration is documented in [UPSTREAM-ENGINE-INTEGRATION.md](UPSTREAM-ENGINE-INTEGRATION.md).
> Do not treat the implementation recommendations below as the current integration status.

**Assessment date:** 2026-11-03  
**Repository snapshot:** `kube-cis-audit(5).zip`  
**Assessment type:** Source-code and benchmark-format compatibility review  
**Implementation status:** Assessment only; no Go source, benchmark, deployment, or workflow files changed.

## 1. Executive summary

KubeCIS has a functioning, deliberately narrow Go audit engine and a custom YAML benchmark format. It is **not currently compatible with kube-bench's benchmark YAML format**. The two formats differ structurally and semantically:

- KubeCIS expects one flat `checks` list with KubeCIS-specific `kind` values.
- kube-bench uses `controls -> groups -> checks`, with target files such as `master.yaml`, `controlplane.yaml`, `node.yaml`, `etcd.yaml`, and `policies.yaml`.
- kube-bench check definitions can contain `audit`, `audit_config`, `audit_env`, `tests`, `test_items`, `bin_op`, `set`, `compare`, `type`, `scored`, and remediation data.
- kube-bench's global and profile-specific `config.yaml` files provide component binary, configuration, service and kubeconfig paths, plus version and target mappings.

**Conclusion:** Do not copy upstream YAML into the current `benchmarks/<id>/benchmark.yaml` location and expect it to run. KubeCIS needs a dedicated upstream loader/parser and a safe compatibility/evaluation layer.

The recommended design is to vendor a pinned upstream kube-bench `cfg/` snapshot as benchmark data, parse it into an internal normalized model, and evaluate only operations implemented natively by KubeCIS. Arbitrary upstream shell commands must never be executed by KubeCIS.

## 2. Current repository inventory

### 2.1 Existing benchmark files

The archive contains:

- `benchmarks/cis-2.0.1/benchmark.yaml`
- `benchmarks/cis-2.1.0/benchmark.yaml`

The current KubeCIS format is defined by `internal/audit/types.go`:

- `Benchmark`: `id`, `name`, `kubernetesMinor`, compatibility metadata, `checks`.
- `Check`: `id`, `title`, `description`, `target`, `kind`, optional resource/path/field/expected/remediation fields, `scored`, and `manual`.

The current 2.1.0 profile has 20 entries: 8 numbered CIS-looking entries and 12 `KUBECIS-*` extensions. The 2.0.1 profile has 6 entries. These are starter profiles, not a complete benchmark catalogue.

### 2.2 Current evaluator dispatch

`internal/audit/runner.go` dispatches the following KubeCIS-specific kinds:

- `api_policy_control`
- `api_pod_non_root`
- `api_pod_seccomp`
- `api_pod_group_security_context`
- `api_network_policy`
- `api_cluster_admin`
- `api_pod_privilege_security`
- `api_pod_capabilities`
- `api_pod_read_only_rootfs`
- `api_pod_host_namespaces`
- `api_pod_hostpath`
- `api_rbac_risk`
- `api_network_security`
- `api_control_plane_security`
- `api_audit_logging`
- `host_file_exists`
- `host_file_mode`
- reporting-only summary kinds
- `manual`

This is a KubeCIS-specific evaluator contract. It does not interpret kube-bench `audit` commands or its `tests` expression model.

### 2.3 Existing node-audit boundary

`RunNode` only executes checks with target `node`, and currently only supports `host_file_mode` and `host_file_exists`. The metadata helper further restricts collection to the allowlisted path `var/lib/kubelet/config.yaml` through `approvedNodeMetadataPath`.

This is an intentional security boundary. Upstream integration must not widen host access or introduce arbitrary command execution to make more upstream tests appear supported.

### 2.4 Existing cluster-audit boundary

`Run` uses the Kubernetes API client. It evaluates KubeCIS-native API checks and does not execute commands on control-plane hosts. The control-plane and audit-logging extension checks inspect visible Pod arguments, which is not equivalent to executing all kube-bench host-level audit commands or reading all host configuration files.

## 3. Upstream kube-bench format and implications

The upstream documentation describes controls as hierarchical YAML:

```yaml
controls:
  id: 1
  text: "Master Node Security Configuration"
  type: master
  groups:
    - id: 1.1
      text: API Server
      checks:
        - id: 1.1.1
          text: "... (Scored)"
          audit: "..."
          tests:
            bin_op: and
            test_items:
              - flag: "--example-flag"
                set: true
          remediation: "..."
          scored: true
```

The upstream format supports multiple evidence/extraction mechanisms:

- `audit` output interpreted as flags (`flag` test item)
- `audit_config` output interpreted as configuration paths (`path` test item)
- `audit_env` output interpreted as environment variables (`env` test item)
- `bin_op` boolean combination of test items
- `set` presence/absence tests
- `compare` operations including `eq`, `noteq`, `gt`, `gte`, `lt`, `lte`, `has`, `nothave`, `regex`, and `bitmask`
- `type: skip` for intentionally omitted controls
- global and benchmark-specific configuration for binaries, config files, services, kubeconfigs, component lists, version mappings, and target mappings

These semantics cannot be represented by the current `Check` struct without a new model and parser.

## 4. Compatibility matrix

| Upstream feature | Current KubeCIS support | Assessment |
|---|---|---|
| Hierarchical `controls/groups/checks` | No | Requires parser and normalization |
| Check IDs and text | Partial | Current fields can store ID/title; preserve original text separately or map without loss |
| Scored flag | Yes | Current `Scored` field exists |
| Remediation | Yes | Current field exists |
| `audit` shell command | No; intentionally not executed | Must map known patterns to safe native evidence collectors; otherwise unsupported |
| `audit_config` | No generic support | Requires safe allowlisted file/config reader and parser |
| `audit_env` | No generic support | Requires narrowly scoped process/environment evidence source; otherwise unsupported |
| `flag` test items | Partial | Existing control-plane checks inspect selected visible Pod arguments, not generic host process arguments |
| `path` test items | No generic support | Requires structured config-path extraction, not arbitrary shell `cat` |
| `env` test items | No | Requires native evidence source or unsupported status |
| `bin_op` (`and`/`or`) | No generic model | Requires normalized expression evaluator |
| `set` true/false | No generic model | Requires normalized expression evaluator |
| `compare` operators | No generic model | Requires typed comparison engine and tests |
| `type: skip` | No explicit upstream type | Should become INFO/SKIPPED with source reason, not PASS |
| Target mapping | Partial | KubeCIS targets currently use `master`, `node`, `policies`, and `report`; upstream has additional target files/sections including `controlplane`, `etcd`, `managedservices` |
| Profile/global `config.yaml` | No | Requires profile resolver and path/component config model |
| Kubernetes version mapping | Partial | KubeCIS scans directory `kubernetesMinor` values; does not parse upstream `version_mapping` |
| Distribution profiles | No general mapping | Requires explicit supported profile/platform selection |

## 5. Version and profile mapping finding

The upstream `cfg/config.yaml` currently documents a `version_mapping` and `target_mapping` mechanism. The inspected upstream mapping includes Kubernetes 1.34 and 1.35 mapped to `cis-2.0`; it does not establish a `cis-2.1.0` mapping for Kubernetes 1.37.

Therefore:

- KubeCIS's existing `cis-2.1.0` profile and provisional Kubernetes 1.37 mapping cannot be replaced by an assumed upstream equivalent.
- Do not silently relabel upstream `cis-2.0` as `cis-2.1.0`.
- Do not assume Kubernetes minor version and CIS benchmark version have a one-to-one relationship.
- The importer must preserve the upstream profile identifier and mapping as supplied.
- If no upstream mapping exists, auto-selection must fail clearly or require an explicitly supported profile; it must not guess.

The exact upstream commit must be pinned before integration. The upstream `main` branch is useful for assessment, but must not be the reproducible source used by normal builds.

## 6. Recommended target architecture

```text
benchmarks/
  upstream/
    kube-bench/
      UPSTREAM_VERSION
      UPSTREAM_COMMIT
      cfg/
        config.yaml
        cis-*/
          config.yaml
          master.yaml
          controlplane.yaml
          node.yaml
          etcd.yaml
          policies.yaml
          ...
  extensions/
    kubecis-*.yaml

internal/
  benchmark/
    upstream_types.go       # kube-bench YAML schema
    parser.go               # parse controls/groups/checks
    resolver.go             # profile/version/target resolution
    normalize.go            # upstream -> KubeCIS normalized checks
    validate.go             # schema, duplicate IDs, unsupported forms
  audit/
    ... existing native evaluators ...
```

The precise package/file layout should be finalized after a focused implementation design review; this is a target architecture, not an instruction to create all files immediately.

### Normalized control model

The normalized representation should preserve at least:

- original upstream profile ID and source file
- upstream repository URL, release/tag and immutable commit SHA
- target and component/group hierarchy
- control ID and original title/text
- scored classification
- remediation text
- original audit/test metadata for traceability
- normalized evidence requirements
- evaluator implementation status
- applicability status and reason

Do not flatten away source metadata needed to prove where a control came from.

## 7. Safe execution model

**Do not execute upstream `audit` strings.** They are shell commands and may contain pipelines, substitutions, file reads, or other arbitrary operations. They are not safe declarative policy expressions.

Instead, implement a constrained interpreter over recognized upstream patterns:

1. Parse the YAML as data only.
2. Recognize a finite set of audit patterns and map them to native evidence collectors.
3. Resolve file paths only through explicit distribution-specific allowlists and the existing read-only host mounts.
4. Parse YAML/JSON configuration using native Go parsers; do not shell out to `cat`, `grep`, `ps`, `stat`, or `sh`.
5. Evaluate test items using a typed expression model with explicit `and`/`or` semantics.
6. Treat unsupported command patterns, missing evidence, denied access, hidden managed control planes, and unknown values as `NOT_ASSESSED` (or the appropriate manual/applicability state), never PASS.
7. Preserve the current no-Secrets-access requirement and node-auditor privilege boundary.

A future evaluator may collect process arguments or service metadata only if its collection path is explicitly designed, allowlisted, read-only, tested, and does not require broadening the helper's privilege boundary without a separate security review.

## 8. Proposed integration phases

### Phase A — Pin and inventory upstream

- Select a kube-bench release/tag and immutable commit SHA.
- Vendor only the relevant upstream `cfg/` source as data.
- Inventory all CIS profiles separately from vendor-specific and STIG profiles.
- Extract profile IDs, target files, controls, scored flags, audit mechanisms and test operators.
- Record license and attribution requirements before redistribution.

**Exit criteria:** reproducible upstream snapshot and generated inventory; no runtime dependency on network or kube-bench executable.

### Phase B — Parser and normalized model

- Implement YAML structs for the upstream schema.
- Parse nested controls/groups/checks and all relevant target files.
- Parse profile/global `config.yaml` and mappings.
- Validate IDs, duplicate controls, required fields, target references and unknown schema features.
- Preserve upstream source metadata.

**Exit criteria:** parser tests against every vendored CIS profile; no evaluator changes yet.

### Phase C — Compatibility report

- Classify each upstream check as `NATIVE_SUPPORTED`, `MAPPABLE_WITH_NEW_COLLECTOR`, `MANUAL`, `UNSUPPORTED`, or `NOT_APPLICABLE`.
- Identify the exact source audit/test shape behind each classification.
- Map checks to existing KubeCIS evaluators only when semantics are equivalent.
- Do not count KubeCIS extensions as official CIS controls.

**Exit criteria:** complete generated control coverage/gap matrix for selected profiles.

### Phase D — Safe native evaluator expansion

- Add native evidence collectors for the most common and safely supportable upstream test patterns.
- Add typed comparison operators and boolean expression evaluation.
- Keep shell commands non-executable.
- Add positive, negative, missing-evidence, unsupported-pattern, and applicability tests for each mapped control.

**Exit criteria:** every implemented control has test-backed evidence and correct indeterminate handling.

### Phase E — Version/platform selection and reports

- Use upstream version and target mappings without guessing.
- Model self-managed versus managed control-plane visibility.
- Report unsupported and manually assessed controls explicitly.
- Maintain reproducible benchmark provenance in scan reports.

**Exit criteria:** repeatable profile selection and transparent coverage/scoring for each supported platform.

## 9. Immediate blockers and risks

1. **No upstream parser:** current loader only reads `benchmark.yaml` into KubeCIS's custom schema.
2. **No upstream expression engine:** no generic `test_items`, `bin_op`, `set`, or `compare` implementation exists.
3. **No generic config collector:** host metadata helper currently supports only a single allowlisted kubelet path.
4. **No generic process/service collector:** cluster API checks cannot substitute for host process or service evidence.
5. **No upstream target resolver:** current target filter does not model all upstream target types.
6. **Version mapping mismatch:** existing provisional `cis-2.1.0`/Kubernetes 1.37 profile is not validated by the inspected upstream mapping.
7. **Security risk if naively ported:** executing upstream shell audit strings would violate KubeCIS's safe, read-only design and introduce command-injection/privilege risks.
8. **Source licensing/attribution:** verify the upstream repository license and required notices before bundling its YAML in KubeCIS releases.

## 10. Decisions recommended

- Adopt kube-bench `cfg/` as the source of benchmark definition data.
- Pin an immutable upstream tag/commit; never track `main` at runtime.
- Keep KubeCIS's Go CLI and native audit engine.
- Build an upstream parser/normalizer before changing evaluators.
- Never execute arbitrary upstream audit shell commands.
- Preserve upstream control IDs, text, scoring, remediation and provenance.
- Separate official upstream CIS controls from KubeCIS extensions.
- Do not claim full coverage until the generated matrix accounts for every control in each supported profile.
- Do not modify GitHub workflows, existing node privilege boundaries, report schema, or CLI during the initial parser/inventory phase.

## 11. References

- kube-bench repository: https://github.com/aquasecurity/kube-bench
- Upstream global config and version/target mappings: https://github.com/aquasecurity/kube-bench/blob/main/cfg/config.yaml
- Upstream controls YAML schema and test semantics: https://github.com/aquasecurity/kube-bench/blob/main/docs/controls.md
- Upstream architecture: https://aquasecurity.github.io/kube-bench/latest/architecture/
- Upstream running/version selection: https://github.com/aquasecurity/kube-bench/blob/main/docs/running.md

---

**Assessment result:** KubeCIS can adopt kube-bench as its benchmark-definition source, but integration requires a parser, normalized control model, profile resolver, and safe native test interpreter. The current engine is a useful base for selected API and host-file checks; it is not a drop-in kube-bench YAML runner.

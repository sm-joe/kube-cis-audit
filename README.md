# KubeCIS

A Go-based, read-only Kubernetes security configuration auditor. Checks are described in YAML and results are emitted as human-readable terminal output or JSON. It can run with a local kubeconfig, in a container, or as a Kubernetes Job.

> **Important scope statement:** this repository is a working starter implementation, not a complete or certified CIS assessment. The included `cis-2.0.1` YAML is a small example profile, not the full CIS Benchmark. CIS Benchmark content is copyrighted/licensed; obtain the official benchmark and translate the recommendations you are entitled to use into your own YAML profile. A PASS means only that the implemented test condition passed. It is not a CIS certification.

## Features

- Detects the Kubernetes server version through discovery and selects a benchmark YAML by the configured Kubernetes minor-version mapping.
- Targets: `master`, `node`, and `policies`; select `all` to include every configured target.
- Read-only Kubernetes API checks for workload security, RBAC permissions and service-account token automount, NetworkPolicy coverage, Service/Ingress/hostPort exposure configuration, and visible control-plane/audit flags.
- Consolidated benchmark coverage and scoring metadata; the compliance percentage is withheld when scored controls have unresolved statuses.
- Consolidated workload security reporting across KubeCIS extensions 2.4.1–2.4.6.
- Milestones 2.5–2.9 add RBAC risk inventory, network configuration assessment, visible control-plane posture, audit logging configuration evidence, and benchmark coverage reporting.
- Host filesystem checks for node/control-plane configuration files when the host root is mounted at `/host`.
- PASS / FAIL / WARN / MANUAL / NOT_ASSESSED / INFO terminal output, structured JSON evidence, remediation text, and a summary.
- Credential selection: explicit `--kubeconfig` first; otherwise in-cluster ServiceAccount credentials when the Kubernetes service environment is present; otherwise standard client-go kubeconfig loading (`KUBECONFIG`, then `~/.kube/config`). If none is usable, the error explains how to configure access.

## Build and run

Requires Go 1.26+ and network access to download Go modules. Kubernetes 1.37 is paired with client-go/apimachinery/api v0.37.2; client-go v0.37 requires Go 1.26.0.

```sh
go mod tidy
go test ./...
go build -o kubecis ./cmd/kubecis

# Uses ~/.kube/config (Windows, macOS, Linux) and the current context
./kubecis scan --config-dir ./benchmarks --benchmark auto --target all

# Or provide an explicit kubeconfig
./kubecis scan --kubeconfig "$HOME/.kube/config" --target all

# Windows PowerShell
.\kubecis.exe scan --kubeconfig "$env:USERPROFILE\.kube\config" --target all

# Machine-readable output
./kubecis scan --output json > kubecis-report.json

# Select one target
./kubecis scan --target policies
```

`--benchmark auto` requires an explicit minor-version mapping in a benchmark's `kubernetesMinor` list. If no mapping exists, the tool fails closed rather than silently applying an unrelated profile.

A provisional `cis-2.1.0` starter profile is included for Kubernetes 1.37. Its mapping is explicitly marked `compatibilityStatus: provisional`: the CIS public catalogue lists v2.1.0 as current, but does not publish a Kubernetes 1.37 compatibility matrix. This is **not** confirmation of official compatibility and the YAML contains only this project's illustrative starter checks, not the complete CIS benchmark. Validate the official benchmark and your distribution before using it for compliance evidence. The CLI and JSON report expose this provisional status.

## Container

```sh
docker build -t kubecis:0.1.0 .
docker run --rm -v "$HOME/.kube:/home/nonroot/.kube:ro" \
  -e KUBECONFIG=/home/nonroot/.kube/config kubecis:0.1.0 scan --output text
```

The default image runs non-root and has a read-only root filesystem when deployed with the supplied Job. For host-level checks, mount the host filesystem read-only at `/host` and set `--host-root /host`. Do not use privileged mode or host PID namespace for these file metadata checks.

## Kubernetes Job

Build and publish the image, then replace the example image in `deploy/job.yaml` with your registry path:

```sh
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/job.yaml
kubectl -n kube-system logs job/kubecis-audit
```

The supplied Job runs API/policy checks only and uses a minimal read-only ClusterRole. For node/control-plane file checks, schedule a separate per-node workload with a read-only host-root mount and suitable tolerations/affinity. Control-plane nodes in managed services are generally not accessible; those controls should report NOT_ASSESSED or MANUAL, not PASS.

## YAML check schema

Each benchmark directory contains `benchmark.yaml`. A check has an ID, title, target, evaluator kind, scored flag, and optional remediation. The evaluator kind selects Go logic; YAML controls which checks are enabled and their benchmark metadata.

Supported evaluator kinds:

| Kind | Scope | Behaviour |
|---|---|---|
| `api_policy_control` | policies | Inventories Pod Security Admission labels, Pod-matching ValidatingAdmissionPolicy bindings, and validating webhooks by actual resource rules. Distinguishes Deny from Warn/Audit and reports incomplete API inventory as NOT_ASSESSED; discovered candidates remain MANUAL pending effective-enforcement review. |
| `api_pod_non_root` | policies | Inspects regular, init and ephemeral containers, resolving container-level overrides over Pod-level `runAsNonRoot` and `runAsUser`. Missing explicit identity evidence is WARN, not a claim that the image runs as root. |
| `api_pod_seccomp` | policies | Inspects effective Pod/container seccomp settings. Explicit `RuntimeDefault` and valid `Localhost` profiles pass; `Unconfined` or missing profiles fail this pod-spec check. |
| `api_network_policy` | policies | Checks for at least one NetworkPolicy object in every namespace, including system namespaces. Presence does not prove effective isolation or CNI enforcement. |
| `api_cluster_admin` | policies | Lists User, Group and ServiceAccount subjects bound to `cluster-admin` through both ClusterRoleBindings and RoleBindings. Bindings are WARN for necessity review. |
| `api_rbac_risk` | policies | Inventories roles/bindings, wildcard and sensitive permissions, escalation verbs, and service-account token automount metadata. Never reads Secret contents. |
| `api_network_security` | policies | Assesses Pod ingress/egress policy selection and inventories Service, Ingress, hostPort and broad policy-rule configuration. Does not prove reachability or CNI enforcement. |
| `api_control_plane_security` | policies | Inspects visible control-plane Pod arguments for selected security flags. Hidden managed control planes and omitted flags require MANUAL verification. |
| `api_audit_logging` | policies | Inspects visible API-server audit policy/backend flags. Does not prove policy file contents, log delivery, retention or alerting. |
| `report_compliance_summary` | report | Adds scored-control counts and coverage; withholds compliance percentage if any scored control is unresolved. |
| `host_file_exists` | master/node | Checks a file beneath `--host-root`. |
| `host_file_mode` | master/node | Checks Unix permission bits against `maxMode` (octal string), using a bitmask comparison. |
| `manual: true` | any | Emits MANUAL to indicate auditor verification is required. |

Unknown evaluator kinds and unavailable evidence emit `NOT_ASSESSED`, not INFO or PASS. `target` is a filter/group label; it does not infer whether a node is a control-plane node. The initial profile is deliberately limited and illustrative.

## Audit limitations and next steps

- Managed control planes (EKS/AKS/GKE control plane) do not expose host configuration to customers. Use provider-specific CIS benchmark profiles and mark inaccessible checks as not applicable/manual.
- Pod checks inspect live Pod objects, including regular, init and ephemeral containers. They do not inspect controller templates that have not yet produced Pods, image config metadata, admission audit logs, or effective runtime state.
- NetworkPolicy existence is not equivalent to a default-deny policy or verified dataplane enforcement.
- Host file checks verify metadata only; they do not parse kubelet flags, API server arguments, etcd TLS configuration, audit policy content, or systemd settings.
- Expand the engine with typed evaluators for each YAML check family, unit tests using fake Kubernetes clients, evidence collection, exclusions, and signed report output before production compliance use.

## References

- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes) — CIS lists Kubernetes Benchmark v2.1.0 as current in its catalogue. CIS currently lists v2.1.0. The included Kubernetes 1.37 mapping is provisional because the public catalogue does not publish an explicit Kubernetes-version compatibility matrix. The included YAML is only a starter profile.
- [Aqua Security kube-bench](https://github.com/aquasecurity/kube-bench) — established reference implementation using YAML controls and host-level checks.
- [kube-bench controls format](https://github.com/aquasecurity/kube-bench/blob/main/docs/controls.md)

This project does not bundle or reproduce the full CIS Benchmark.


### CIS 2.1.0 starter-profile limitations

The `cis-2.1.0` directory is a limited starter profile, not a complete benchmark implementation. It currently includes selected checks only. Pod checks evaluate non-terminal Pods in all namespaces and effective pod/container security contexts; they do not inspect controller templates or image metadata. Pod Security Admission coverage is inferred from namespace `pod-security.kubernetes.io/enforce` labels and does not prove the admission plugin is enabled or that Kyverno/Gatekeeper policies are effective. Host filesystem checks require the relevant node filesystem to be mounted at `--host-root`; a single Job only inspects the node on which that Job runs. Use a DaemonSet or a dedicated per-node collection design for complete worker-node coverage. Managed control planes may not expose host files at all.


## Audit result semantics (Milestone 1)

KubeCIS distinguishes a verified violation from missing evidence:

- `PASS`: the configured evaluator found evidence satisfying the check.
- `FAIL`: the evaluator found an explicit violation of the configured check.
- `WARN`: partial coverage or an ambiguous configuration needs review.
- `MANUAL`: evidence suggests a control is configured, but effective enforcement cannot be established through the Kubernetes API alone.
- `NOT_ASSESSED`: the required API or host evidence was unavailable, or the evaluator does not support the check.
- `INFO`: informational result.

Each result can include structured `evidence` entries in JSON and corresponding evidence lines in text output. Missing kubelet/control-plane host mounts are `NOT_ASSESSED`, not `FAIL`. Pod security evaluation includes regular, init and ephemeral containers and honours container-level overrides over Pod-level security context. A missing `runAsNonRoot`/`runAsUser` declaration is reported as insufficient explicit evidence rather than a claim that the image is running as root. NetworkPolicy presence is checked across all namespaces; presence alone is not proof of isolation or CNI enforcement.

The `api_policy_control` evaluator does not claim a policy engine is effective merely because a webhook or policy object exists. It matches validating webhooks and ValidatingAdmissionPolicy bindings against Pod CREATE/UPDATE requests, distinguishes Deny from Warn/Audit actions, and surfaces incomplete inventory. Discovered candidates and full PSA namespace-label coverage remain MANUAL pending review of CEL expressions, selectors, exclusions, parameters, webhook endpoint behaviour, API-server configuration and exemptions. The Job ClusterRole includes read-only access to the admissionregistration resources used for this discovery. See `docs/MILESTONE-2-STEP-2.3.md`.

## Node-level host assessment (Linux)

The regular `kubecis scan` command audits Kubernetes API-visible configuration. It does not automatically have access to the filesystem of each Kubernetes node. For Linux node host-file checks, use the separate `node-audit` command through the read-only DaemonSet:

1. Build and publish the image, then replace the placeholder image in `deploy/node-auditor-daemonset.yaml`.
2. Review the host-root mount and your cluster's hostPath security policy.
3. Apply the DaemonSet:

   ```powershell
   kubectl apply -f .\deploy\node-auditor-daemonset.yaml
   ```

4. Inspect per-node results:

   ```powershell
   kubectl -n kube-system get pods -l app.kubernetes.io/name=kubecis-node-auditor -o wide
   kubectl -n kube-system logs -l app.kubernetes.io/name=kubecis-node-auditor --all-containers=true --prefix=true
   ```

The node auditor requires an explicit benchmark because it does not query the Kubernetes API version:

```text
kubecis node-audit --benchmark cis-2.1.0 --host-root /host --node-name worker-01 --output json
```

The DaemonSet repeats the assessment every 24 hours. It runs as non-root with all Linux capabilities dropped, has no service-account token, uses RuntimeDefault seccomp, and mounts the host root read-only. If a host path cannot be traversed, it reports `NOT_ASSESSED`; do not make the container privileged to force access. This initial node mode runs only checks with `target: node` and supported host-file kinds. It does not assess managed control-plane files or Windows nodes. See `docs/MILESTONE-2-STEP-2.2.md` for the security model and limitations.


### Verify the running build

Run `kubecis version` to print the embedded build identifier. Cluster JSON reports include the same value in `buildVersion`. For the admission-inventory verification package, the expected identifier is `2.3.2-admission-inventory`.

## Docker image and Kubernetes deployment

See [Docker and E2E deployment guide](docs/DOCKER-AND-E2E.md) for the multi-stage Distroless image, the read-only cluster audit Job, the two-container node-local DaemonSet, GHCR publishing, and the disposable Kind E2E test.

The node-local DaemonSet separates the unprivileged auditor from a metadata-only helper. The helper's narrowly scoped UID 0 + `CAP_DAC_READ_SEARCH` exception is restricted to a read-only `/var/lib/kubelet` mount and an explicit file-path allowlist; it does not read host file contents. Do not replace this boundary with a privileged container.

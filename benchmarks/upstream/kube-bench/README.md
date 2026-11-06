# kube-bench EKS benchmark reference snapshot (NOT an executable dependency)

This directory contains an unchanged copy of `cfg/eks-1.8.0` from
[Aqua Security kube-bench](https://github.com/aquasecurity/kube-bench),
for **control inventory and implementation design only**. KubeCIS does not
run, execute or import kube-bench audits. See the source repository's Apache
License 2.0 and retain its attribution and notices.

Original revision: `main` upstream reference files pinned by their Git blob
SHA-1 identities below (Git hashes refer to file contents, not to a CIS
publication or an endorsement):

| File | Upstream blob SHA |
|---|---|
| config.yaml | 17301a751a53161bd248d6b0f450ab1e6a5e1a86 |
| controlplane.yaml | d6d674bd8ce6762ff5f701af82c9f69c9ee52e89 |
| managedservices.yaml | 4f6560c4552b73456eb7e42b7105287a7e083afa |
| master.yaml | 76444a46d376958dbbf19af61d22bf1d87a847f5 |
| node.yaml | e81fa66ffd6d3081f639c316b48fd13a85f74056 |
| policies.yaml | a7b05add85a67e4da117c0ccda6f3b71ebe2a61d |

**IMPORTANT:** `eks-1.8.0` is the *upstream kube-bench EKS 1.8.0
configuration*, NOT the CIS Amazon EKS Benchmark v2.1.0. Neither this
directory nor the provisional local `benchmarks/cis-2.1.0` starter profile
establishes current official CIS benchmark coverage. Do not merge control IDs
across editions: e.g. EKS 1.8.0 control 3.1.3 differs from the generic
Kubernetes 4.1.9 mapping.

To inventory this reference locally:

```shell
go run ./cmd/benchmark-inventory -root benchmarks/upstream/kube-bench/cfg -out docs/UPSTREAM-BENCHMARK-INVENTORY.md -json dist/eks-reference-inventory.json
```

The inventory is structural; it does not execute audit shell snippets or
transmute any MANUAL check into PASS.

# KubeCIS release process

> Stable product releases do not require independent CIS certification. The bundled provisional benchmark profiles retain their stated limitations and withheld official scoring.

A version tag is the release trigger. The workflow builds platform binaries with the benchmark profiles, publishes SHA-256 checksums, and builds/pushes a multi-architecture GHCR image.

## Before tagging

Review [RC4-PATCH-NOTES.md](RC4-PATCH-NOTES.md) for the candidate hardening changes and validation limitations.

Run locally with the Go version pinned in go.mod:

```bash
go vet ./...
go test ./... -count=1
CGO_ENABLED=0 go build -trimpath -o ./kubecis ./cmd/kubecis
./kubecis version
./kubecis --help
```

Review the included benchmark compatibility note and all three deployment modes (`deploy/job.yaml`, `deploy/node-auditor-daemonset.yaml`, and `deploy/controlplane-auditor-daemonset.yaml`). Do not claim complete CIS coverage or certification.

The release workflow matches every tag beginning with `v`. Tags containing a hyphen (for example `v3.0.0-rc4`) are published as GitHub prereleases and receive only their versioned GHCR tag. They do **not** move `latest`. Only stable tags without a hyphen publish `latest`. The release workflow now runs formatting, `go vet`, the complete Go test suite, and a CLI build before producing release assets.

## Publish

Merge the release-ready changes to `main`, then create and push a semantic version tag:

```bash
git tag v3.2.0
git push origin v3.2.0
```

The GitHub Actions Release workflow publishes:

- `kubecis_windows_amd64.zip`
- `kubecis_linux_amd64.tar.gz`
- `kubecis_linux_arm64.tar.gz`
- `kubecis_darwin_amd64.tar.gz`
- `kubecis_darwin_arm64.tar.gz`
- `SHA256SUMS`
- `ghcr.io/sm-joe/kubecis:v3.2.0` and `:latest` for a stable release; prereleases receive only their versioned tag

The Windows installer downloads the release asset and verifies its SHA-256 checksum before installation.

## First-release checklist

- Confirm GitHub Actions CI passes on the release commit.
- Confirm all five binary assets and `SHA256SUMS` exist on the GitHub Release.
- Confirm GHCR contains both `linux/amd64` and `linux/arm64` manifests.
- Test the Windows installer on a clean Windows account.
- Test `kubecis scan` against a designated cluster with the least-privilege RBAC manifest.
- Confirm the generated report identifies the binary build version and provisional benchmark compatibility.
- Make the GHCR package public if anonymous pulls are intended.

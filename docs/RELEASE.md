# Release process

Run Go tests, vet, Linux collection and deployment tests, and review the security workflow. Merge the release-ready PR only after checks pass.

```bash
git switch main
git pull --ff-only
go vet ./...
go test ./... -count=1
git tag v1.0.0
git push origin v1.0.0
```

The Release Pipeline builds, verifies and signs Linux archives, publishes multi-architecture container images, and attaches checksums and an immutable release manifest. Verify the assets and run the documented EKS collection from a clean extracted archive.

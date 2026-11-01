# Milestone 1 build fix 3

The previous host-file permission test assumed `os.Chmod(0644)` followed by `os.Stat().Mode().Perm()` returns POSIX mode `0644`. Windows does not support arbitrary POSIX mode bits through Go's filesystem API and may report a synthetic mode such as `0666`.

The permission-mask logic is now isolated in `evaluateHostFileMode` and tested directly with POSIX `os.FileMode` values (0644 must fail against 0600; 0600 must pass). On Windows, runtime host-file mode checks now return `NOT_ASSESSED` rather than reporting synthetic Windows mode bits as a Linux/Kubernetes security failure. Run node host-file checks on Linux or in the node-audit container for real POSIX evidence.

Source was formatted with gofmt. Tests were not run in the packaging environment: it has Go 1.23.2, while go.mod requires Go 1.26.0. Run `go test ./...` with Go 1.26.0 or later.

# B3 fixture contract

The unit tests construct metadata-only snapshots in memory. The fixture shape
matches the existing KubeCIS node helper JSON: `nodeName`, `generatedAt`, and
`files[path]` with `mode`, `regular`, `symlink`, and optional `error`.
No file contents, UID, or GID are collected or represented.

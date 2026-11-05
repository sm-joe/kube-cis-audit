# Security model

The cluster CLI uses read-only Kubernetes API calls; its ServiceAccount must not have write verbs. Local node assessments require Linux and only narrow, read-only host mounts. The main container runs as non-root with dropped capabilities. Metadata helpers export only allowlisted file metadata and the selected kubelet authentication/authorization values; arbitrary configuration content is not exposed. Checks with unavailable evidence return `NOT_ASSESSED`.

Kubelet content checks only accept `var/lib/kubelet/config.yaml`, allowlisted field keys, a regular non-symlink path, and files no larger than 1 MiB. Do not mount Kubernetes secrets or entire host roots merely to satisfy audit checks. Symlinks and unreadable evidence never count as PASS.

KubeCIS does not execute shell scripts or arbitrary commands provided by benchmark YAML. Never claim full CIS compliance for provisional or unsupported mappings. A `FAIL` is an evaluated insecure setting, while `NOT_ASSESSED` denotes unavailable/incomplete evidence. Missing current findings do not prove resolution.

Report files can contain sensitive resource names and configuration details: protect access and avoid committing reports. No local credentials, kubeconfigs, secrets or Terraform state should be included in release archives.

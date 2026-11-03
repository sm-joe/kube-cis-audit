# kube-bench CIS benchmark inventory

Generated from `benchmarks/upstream/kube-bench/cfg`. This is a structural inventory of upstream YAML; it does not execute audit commands or certify benchmark correctness.

- CIS profiles: **13**
- YAML control files: **65**
- Control entries: **1647**

## Profile summary

| Profile | Files | Controls | Scored | Not scored | Scoring unspecified | Targets |
|---|---:|---:|---:|---:|---:|---|
| `cis-1.10` | 5 | 130 | 64 | 66 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.11` | 5 | 133 | 65 | 68 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.12` | 5 | 131 | 65 | 66 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.20` | 5 | 123 | 69 | 54 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.23` | 5 | 125 | 67 | 58 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.24` | 5 | 124 | 64 | 60 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.24-microk8s` | 5 | 124 | 58 | 66 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.5` | 5 | 122 | 87 | 35 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.6` | 5 | 122 | 69 | 53 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.6-k3s` | 5 | 122 | 66 | 56 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.7` | 5 | 131 | 61 | 70 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.8` | 5 | 130 | 63 | 67 | 0 | controlplane, etcd, master, node, policies |
| `cis-1.9` | 5 | 130 | 64 | 66 | 0 | controlplane, etcd, master, node, policies |

## cis-1.10

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.10/controlplane.yaml` |
| controlplane | `3.1.2` | Service account token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.10/controlplane.yaml` |
| controlplane | `3.1.3` | Bootstrap token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.10/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.10/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.10/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.10/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.10/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.10/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.10/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.10/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.10/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.10/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.13` | Ensure that the default administrative credential file permissions are set to 600 (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.14` | Ensure that the default administrative credential file ownership is set to root:root (Automated) | yes | 1 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.10/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.10/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.10/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.10/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.10/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.10/master.yaml` |
| master | `1.2.15` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.2.16` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.17` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.10/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.10/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.10/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.20` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.10/master.yaml` |
| master | `1.2.21` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.2.22` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.23` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.24` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.25` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.26` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.27` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.28` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.10/master.yaml` |
| master | `1.2.29` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.10/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is set (Manual) | no | 1 | flag | has | `cis-1.10/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.10/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.10/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.10/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.10/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.10/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.10/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.10/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.10/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.10/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.10/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.10/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.10/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.10/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.10/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.10/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.10/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.10/node.yaml` |
| node | `4.2.10` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.10/node.yaml` |
| node | `4.2.11` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.10/node.yaml` |
| node | `4.2.12` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.10/node.yaml` |
| node | `4.2.13` | Ensure that a limit is set on pod PIDs (Manual) | no | 1 | flag, path |  | `cis-1.10/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.10/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.10/node.yaml` |
| node | `4.2.4` | Verify that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.10/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.10/node.yaml` |
| node | `4.2.6` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.10/node.yaml` |
| node | `4.2.7` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.10/node.yaml` |
| node | `4.2.8` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.10/node.yaml` |
| node | `4.2.9` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.10/node.yaml` |
| node | `4.3.1` | Ensure that the kube-proxy metrics service is bound to localhost (Automated) | yes | 2 | flag, path | has | `cis-1.10/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Automated) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.1.10` | Minimize access to the proxy sub-resource of nodes (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.1.11` | Minimize access to the approval sub-resource of certificatesigningrequests objects (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.1.12` | Minimize access to webhook configuration objects (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.1.13` | Minimize access to the service account token creation (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Automated) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Automated) | no | 2 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Automated) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used (Automated) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Automated) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.1.9` | Minimize access to create persistent volumes (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Manual) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Manual) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Manual) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Manual) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Manual) | no | 1 | flag | eq | `cis-1.10/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.10/policies.yaml` |

## cis-1.11

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.11/controlplane.yaml` |
| controlplane | `3.1.2` | Service account token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.11/controlplane.yaml` |
| controlplane | `3.1.3` | Bootstrap token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.11/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.11/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.11/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.11/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.11/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.11/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.11/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.11/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.11/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.11/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.13` | Ensure that the default administrative credential file permissions are set to 600 (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.14` | Ensure that the default administrative credential file ownership is set to root:root (Automated) | yes | 1 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.11/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.11/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.11/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.11/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.11/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.11/master.yaml` |
| master | `1.2.15` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.2.16` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.17` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.11/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.11/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.11/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.20` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.11/master.yaml` |
| master | `1.2.21` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.2.22` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.23` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.24` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.25` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.26` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.27` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.28` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.11/master.yaml` |
| master | `1.2.29` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.11/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is set (Manual) | no | 1 | flag | has | `cis-1.11/master.yaml` |
| master | `1.2.30` | Ensure that the --service-account-extend-token-expiration parameter is set to false (Automated) | yes | 1 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.11/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.11/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.11/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.11/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.11/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.11/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.11/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.11/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.11/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.11/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.11/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.11/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.11/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.11/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.11/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.11/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.11/node.yaml` |
| node | `4.2.10` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.11/node.yaml` |
| node | `4.2.11` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.11/node.yaml` |
| node | `4.2.12` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.11/node.yaml` |
| node | `4.2.13` | Ensure that a limit is set on pod PIDs (Manual) | no | 1 | flag, path |  | `cis-1.11/node.yaml` |
| node | `4.2.14` | Ensure that the --seccomp-default parameter is set to true (Manual) | no | 1 | flag, path |  | `cis-1.11/node.yaml` |
| node | `4.2.15` | Ensure that the --IPAddressDeny is set to any (Manual) | no | 1 | flag, path |  | `cis-1.11/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.11/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.11/node.yaml` |
| node | `4.2.4` | Verify that if defined, the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.11/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.11/node.yaml` |
| node | `4.2.6` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.11/node.yaml` |
| node | `4.2.7` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.11/node.yaml` |
| node | `4.2.8` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.11/node.yaml` |
| node | `4.2.9` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.11/node.yaml` |
| node | `4.3.1` | Ensure that the kube-proxy metrics service is bound to localhost (Automated) | yes | 2 | flag, path | has | `cis-1.11/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.1.10` | Minimize access to the proxy sub-resource of nodes (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.1.11` | Minimize access to the approval sub-resource of certificatesigningrequests objects (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.1.12` | Minimize access to webhook configuration objects (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.1.13` | Minimize access to the service account token creation (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 2 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.1.9` | Minimize access to create persistent volumes (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Manual) | no | 1 | flag | eq | `cis-1.11/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.6.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.6.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.6.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |
| policies | `5.6.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.11/policies.yaml` |

## cis-1.12

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.12/controlplane.yaml` |
| controlplane | `3.1.2` | Service account token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.12/controlplane.yaml` |
| controlplane | `3.1.3` | Bootstrap token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.12/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.12/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.12/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.12/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.12/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.12/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.12/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.12/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.12/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.12/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.13` | Ensure that the default administrative credential file permissions are set to 600 (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.14` | Ensure that the default administrative credential file ownership is set to root:root (Automated) | yes | 1 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.12/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.12/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.12/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.12/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.12/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.12/master.yaml` |
| master | `1.2.15` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.2.16` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.17` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.12/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.12/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.12/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.20` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.12/master.yaml` |
| master | `1.2.21` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.2.22` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.23` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.24` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.25` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.26` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.27` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.28` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.12/master.yaml` |
| master | `1.2.29` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.12/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is set (Manual) | no | 1 | flag | has | `cis-1.12/master.yaml` |
| master | `1.2.30` | Ensure that the --service-account-extend-token-expiration parameter is set to false (Automated) | yes | 1 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.12/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.12/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.12/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.12/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.12/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.12/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.12/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.12/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.12/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.12/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.12/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.12/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.12/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.12/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.12/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.12/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.12/node.yaml` |
| node | `4.2.10` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.12/node.yaml` |
| node | `4.2.11` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.12/node.yaml` |
| node | `4.2.12` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.12/node.yaml` |
| node | `4.2.13` | Ensure that a limit is set on pod PIDs (Manual) | no | 1 | flag, path |  | `cis-1.12/node.yaml` |
| node | `4.2.14` | Ensure that the --seccomp-default parameter is set to true (Manual) | no | 1 | flag, path |  | `cis-1.12/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.12/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.12/node.yaml` |
| node | `4.2.4` | Verify that if defined, the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.12/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.12/node.yaml` |
| node | `4.2.6` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.12/node.yaml` |
| node | `4.2.7` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.12/node.yaml` |
| node | `4.2.8` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.12/node.yaml` |
| node | `4.2.9` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.12/node.yaml` |
| node | `4.3.1` | Ensure that the kube-proxy metrics service is bound to localhost (Automated) | yes | 2 | flag, path | has | `cis-1.12/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.1.10` | Minimize access to the proxy sub-resource of nodes (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.1.11` | Minimize access to the approval sub-resource of certificatesigningrequests objects (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.1.12` | Minimize access to webhook configuration objects (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.1.13` | Minimize access to the service account token creation (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 2 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.1.9` | Minimize access to create persistent volumes (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Manual) | no | 1 | flag | eq | `cis-1.12/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.6.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.6.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.6.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |
| policies | `5.6.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.12/policies.yaml` |

## cis-1.20

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.20/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.20/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.20/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.20/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.20/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.20/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.20/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.20/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.20/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.20/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.20/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.20/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.20/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.20/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.20/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.20/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin PodSecurityPolicy is set (Automated) | yes | 1 | flag | has | `cis-1.20/master.yaml` |
| master | `1.2.16` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.20/master.yaml` |
| master | `1.2.17` | Ensure that the --insecure-bind-address argument is not set (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.18` | Ensure that the --insecure-port argument is set to 0 (Automated) | yes | 1 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.2.19` | Ensure that the --secure-port argument is not set to 0 (Automated) | yes | 2 | flag | gt | `cis-1.20/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.20` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.2.21` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.22` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.20/master.yaml` |
| master | `1.2.23` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.20/master.yaml` |
| master | `1.2.24` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.20/master.yaml` |
| master | `1.2.25` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.20/master.yaml` |
| master | `1.2.26` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.2.27` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.28` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.29` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.3` | Ensure that the --kubelet-https argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.2.30` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.31` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.32` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.33` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.20/master.yaml` |
| master | `1.2.34` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.20/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.20/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.20/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.20/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.20/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.20/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.20/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.20/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.20/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.20/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/node.yaml` |
| node | `4.1.10` | Ensure that the kubelet --config configuration file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.20/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.20/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.20/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.20/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.20/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.20/node.yaml` |
| node | `4.1.9` | Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.20/node.yaml` |
| node | `4.2.1` | Ensure that the anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.20/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.20/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.20/node.yaml` |
| node | `4.2.12` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.20/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.20/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.20/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.20/node.yaml` |
| node | `4.2.4` | Ensure that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.20/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.20/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Automated) | yes | 1 | flag, path | eq | `cis-1.20/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.20/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.20/node.yaml` |
| node | `4.2.9` | Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture (Manual) | no | 1 | flag, path | eq | `cis-1.20/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.1` | Minimize the admission of privileged containers (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of containers wishing to share the host process ID namespace (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host IPC namespace (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host network namespace (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers with allowPrivilegeEscalation (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of root containers (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of containers with the NET_RAW capability (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with added capabilities (Automated) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports Network Policies (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have Network Policies defined (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.4.1` | Prefer using secrets as files over secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your pod definitions (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.7.3` | Apply Security Context to Your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.20/policies.yaml` |

## cis-1.23

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.23/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.23/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.23/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.23/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.23/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.23/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.23/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.23/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.23/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.23/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.23/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.23/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.23/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.23/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.23/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.23/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.23/master.yaml` |
| master | `1.2.16` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.23/master.yaml` |
| master | `1.2.17` | Ensure that the --secure-port argument is not set to 0 (Automated) | yes | 2 | flag | gt | `cis-1.23/master.yaml` |
| master | `1.2.18` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.20` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.23/master.yaml` |
| master | `1.2.21` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.23/master.yaml` |
| master | `1.2.22` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.23/master.yaml` |
| master | `1.2.23` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.23/master.yaml` |
| master | `1.2.24` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.2.25` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.26` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.27` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.28` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.29` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is not set (Automated) | yes | 2 | flag | nothave | `cis-1.23/master.yaml` |
| master | `1.2.30` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.31` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.23/master.yaml` |
| master | `1.2.32` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.23/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-https argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.6` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.23/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.23/master.yaml` |
| master | `1.2.9` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.23/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.23/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.23/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.23/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.23/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.23/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/node.yaml` |
| node | `4.1.10` | Ensure that the kubelet --config configuration file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.23/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.23/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.23/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.23/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.23/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.23/node.yaml` |
| node | `4.1.9` | Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.23/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.23/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.23/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.23/node.yaml` |
| node | `4.2.12` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.23/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.23/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.23/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.23/node.yaml` |
| node | `4.2.4` | Ensure that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.23/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.23/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Automated) | yes | 1 | flag, path | eq | `cis-1.23/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.23/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.23/node.yaml` |
| node | `4.2.9` | Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture (Manual) | no | 1 | flag, path | eq | `cis-1.23/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Automated) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.23/policies.yaml` |

## cis-1.24

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.24/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.24/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.24/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.24/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.24/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.24/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.24/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.24/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.24/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.24/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.24/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.24/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.24/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.24/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.24/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.24/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.24/master.yaml` |
| master | `1.2.16` | Ensure that the --secure-port argument is not set to 0 (Automated) | yes | 2 | flag | gt | `cis-1.24/master.yaml` |
| master | `1.2.17` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.24/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.24/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.20` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.24/master.yaml` |
| master | `1.2.21` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.24/master.yaml` |
| master | `1.2.22` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.24/master.yaml` |
| master | `1.2.23` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.24/master.yaml` |
| master | `1.2.24` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.25` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.26` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.27` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.28` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.29` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is not set (Automated) | yes | 2 | flag | nothave | `cis-1.24/master.yaml` |
| master | `1.2.30` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.24/master.yaml` |
| master | `1.2.31` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.24/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.24/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.24/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.24/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.24/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.24/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.24/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.24/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 1 | flag | eq | `cis-1.24/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.24/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 1 | flag | eq | `cis-1.24/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.24/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.24/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.24/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.24/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.24/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.24/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.24/node.yaml` |
| node | `4.2.12` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.24/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.24/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.24/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.24/node.yaml` |
| node | `4.2.4` | Verify that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.24/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.24/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Automated) | yes | 1 | flag, path | eq | `cis-1.24/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.24/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.24/node.yaml` |
| node | `4.2.9` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.24/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Automated) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.24/policies.yaml` |

## cis-1.24-microk8s

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.24-microk8s/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.24-microk8s/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | no | 2 | env, flag |  | `cis-1.24-microk8s/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | no | 1 | env, flag | eq | `cis-1.24-microk8s/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | no | 2 | env, flag | eq | `cis-1.24-microk8s/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.24-microk8s/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | no | 2 | env, flag | eq | `cis-1.24-microk8s/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.24-microk8s/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 2 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.16` | Ensure that the --secure-port argument is not set to 0 (Automated) | yes | 2 | flag | gt | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.17` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.20` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.21` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.22` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.23` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.24` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.25` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | no | 2 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.26` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.27` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.28` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | no | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.29` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is not set (Automated) | yes | 2 | flag | nothave | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.30` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.31` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.24-microk8s/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.24-microk8s/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.24-microk8s/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.24-microk8s/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.12` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.4` | Verify that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Automated) | yes | 1 | flag, path | eq | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.24-microk8s/node.yaml` |
| node | `4.2.9` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 1 | flag, path | eq | `cis-1.24-microk8s/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Automated) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.24-microk8s/policies.yaml` |

## cis-1.5

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Not Scored) | no | 0 |  |  | `cis-1.5/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Scored) | yes | 1 | flag |  | `cis-1.5/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Not Scored) | no | 0 |  |  | `cis-1.5/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Scored) | yes | 2 | flag |  | `cis-1.5/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Scored) | yes | 1 | flag | eq | `cis-1.5/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Scored) | yes | 2 | flag | eq | `cis-1.5/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Scored) | yes | 2 | flag |  | `cis-1.5/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Scored) | yes | 1 | flag | eq | `cis-1.5/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Scored) | yes | 2 | flag | eq | `cis-1.5/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Not Scored) | no | 1 | flag |  | `cis-1.5/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Not Scored) | no | 0 |  |  | `cis-1.5/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Not Scored) | no | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Not Scored) | no | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 644 or more restrictive (Not Scored) | no | 0 |  |  | `cis-1.5/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Not Scored) | no | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin EventRateLimit is set (Not Scored) | no | 1 | flag | has | `cis-1.5/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysAdmit is not set (Scored) | yes | 2 | flag | nothave | `cis-1.5/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin AlwaysPullImages is set (Not Scored) | no | 1 | flag | has | `cis-1.5/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Not Scored) | no | 2 | flag | has | `cis-1.5/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin ServiceAccount is set (Scored) | yes | 2 | flag | nothave | `cis-1.5/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NamespaceLifecycle is set (Scored) | yes | 2 | flag | nothave | `cis-1.5/master.yaml` |
| master | `1.2.16` | Ensure that the admission control plugin PodSecurityPolicy is set (Scored) | yes | 1 | flag | has | `cis-1.5/master.yaml` |
| master | `1.2.17` | Ensure that the admission control plugin NodeRestriction is set (Scored) | yes | 1 | flag | has | `cis-1.5/master.yaml` |
| master | `1.2.18` | Ensure that the --insecure-bind-address argument is not set (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.19` | Ensure that the --insecure-port argument is set to 0 (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.2.2` | Ensure that the --basic-auth-file argument is not set (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.20` | Ensure that the --secure-port argument is not set to 0 (Scored) | yes | 2 | flag | gt | `cis-1.5/master.yaml` |
| master | `1.2.21` | Ensure that the --profiling argument is set to false (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.2.22` | Ensure that the --audit-log-path argument is set (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.23` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Scored) | yes | 1 | flag | gte | `cis-1.5/master.yaml` |
| master | `1.2.24` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Scored) | yes | 1 | flag | gte | `cis-1.5/master.yaml` |
| master | `1.2.25` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Scored) | yes | 1 | flag | gte | `cis-1.5/master.yaml` |
| master | `1.2.26` | Ensure that the --request-timeout argument is set as appropriate (Scored) | yes | 0 |  |  | `cis-1.5/master.yaml` |
| master | `1.2.27` | Ensure that the --service-account-lookup argument is set to true (Scored) | yes | 2 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.2.28` | Ensure that the --service-account-key-file argument is set as appropriate (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.29` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Scored) | yes | 2 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.3` | Ensure that the --token-auth-file parameter is not set (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.30` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Scored) | yes | 2 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.31` | Ensure that the --client-ca-file argument is set as appropriate (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.32` | Ensure that the --etcd-cafile argument is set as appropriate (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.33` | Ensure that the --encryption-provider-config argument is set as appropriate (Not Scored) | no | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.34` | Ensure that encryption providers are appropriately configured (Not Scored) | no | 0 |  |  | `cis-1.5/master.yaml` |
| master | `1.2.35` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Not Scored) | no | 1 | flag | valid_elements | `cis-1.5/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-https argument is set to true (Scored) | yes | 2 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Scored) | yes | 2 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.6` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Scored) | yes | 1 | flag | nothave | `cis-1.5/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes Node (Scored) | yes | 1 | flag | has | `cis-1.5/master.yaml` |
| master | `1.2.9` | Ensure that the --authorization-mode argument includes RBAC (Scored) | yes | 1 | flag | has | `cis-1.5/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Not Scored) | no | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Scored) | yes | 1 | flag | noteq | `cis-1.5/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Scored) | yes | 1 | flag |  | `cis-1.5/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Scored) | yes | 2 | flag | nothave | `cis-1.5/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Scored) | yes | 2 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Scored) | yes | 1 | flag | eq | `cis-1.5/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Scored) | yes | 2 | flag | eq | `cis-1.5/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/node.yaml` |
| node | `4.1.10` | Ensure that the kubelet configuration file ownership is set to root:root (Scored) | yes | 1 | flag |  | `cis-1.5/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.5/node.yaml` |
| node | `4.1.3` | Ensure that the proxy kubeconfig file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/node.yaml` |
| node | `4.1.4` | Ensure that the proxy kubeconfig file ownership is set to root:root (Scored) | yes | 1 | flag |  | `cis-1.5/node.yaml` |
| node | `4.1.5` | Ensure that the kubelet.conf file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/node.yaml` |
| node | `4.1.6` | Ensure that the kubelet.conf file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Scored) | yes | 1 | flag | eq | `cis-1.5/node.yaml` |
| node | `4.1.9` | Ensure that the kubelet configuration file has permissions set to 644 or more restrictive (Scored) | yes | 1 | flag | bitmask | `cis-1.5/node.yaml` |
| node | `4.2.1` | Ensure that the anonymous-auth argument is set to false (Scored) | yes | 1 | flag, path | eq | `cis-1.5/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Scored) | yes | 2 | flag, path |  | `cis-1.5/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Scored) | yes | 2 | flag, path | eq | `cis-1.5/node.yaml` |
| node | `4.2.12` | Ensure that the RotateKubeletServerCertificate argument is set to true (Scored) | yes | 2 | flag, path | nothave | `cis-1.5/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Not Scored) | no | 1 | flag, path | valid_elements | `cis-1.5/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Scored) | yes | 1 | flag, path | nothave | `cis-1.5/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Scored) | yes | 1 | flag, path |  | `cis-1.5/node.yaml` |
| node | `4.2.4` | Ensure that the --read-only-port argument is set to 0 (Scored) | yes | 2 | flag, path | eq | `cis-1.5/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Scored) | yes | 2 | flag, path | noteq | `cis-1.5/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Scored) | yes | 1 | flag, path | eq | `cis-1.5/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Scored) | yes | 2 | flag, path | eq | `cis-1.5/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Not Scored) | no | 1 | flag |  | `cis-1.5/node.yaml` |
| node | `4.2.9` | Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture (Not Scored) | no | 1 | flag, path | eq | `cis-1.5/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.1` | Minimize the admission of privileged containers (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of containers wishing to share the host process ID namespace (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host IPC namespace (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host network namespace (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers with allowPrivilegeEscalation (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of root containers (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of containers with the NET_RAW capability (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with added capabilities (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with capabilities assigned (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports Network Policies (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have Network Policies defined (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.4.1` | Prefer using secrets as files over secrets as environment variables (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your pod definitions (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.7.3` | Apply Security Context to Your Pods and Containers (Not Scored) | no | 0 |  |  | `cis-1.5/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Scored) | yes | 0 |  |  | `cis-1.5/policies.yaml` |

## cis-1.6

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.6/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.6/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.6/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.6/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.6/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.6/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.6/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.6/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.6/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.6/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.6/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.6/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.6/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.6/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.6/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.6/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.6/master.yaml` |
| master | `1.2.16` | Ensure that the admission control plugin PodSecurityPolicy is set (Automated) | yes | 1 | flag | has | `cis-1.6/master.yaml` |
| master | `1.2.17` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.6/master.yaml` |
| master | `1.2.18` | Ensure that the --insecure-bind-address argument is not set (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.19` | Ensure that the --insecure-port argument is set to 0 (Automated) | yes | 1 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.2.2` | Ensure that the --basic-auth-file argument is not set (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.20` | Ensure that the --secure-port argument is not set to 0 (Automated) | yes | 2 | flag | gt | `cis-1.6/master.yaml` |
| master | `1.2.21` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.2.22` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.23` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.6/master.yaml` |
| master | `1.2.24` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.6/master.yaml` |
| master | `1.2.25` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.6/master.yaml` |
| master | `1.2.26` | Ensure that the --request-timeout argument is set as appropriate (Automated) | yes | 0 |  |  | `cis-1.6/master.yaml` |
| master | `1.2.27` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.2.28` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.29` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.3` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.30` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.31` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.32` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.33` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.34` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.6/master.yaml` |
| master | `1.2.35` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.6/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-https argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.6` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.6/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.6/master.yaml` |
| master | `1.2.9` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.6/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.6/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.6/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.6/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/node.yaml` |
| node | `4.1.10` | Ensure that the kubelet --config configuration file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.6/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 644 or more restrictive (Manual) | no | 2 | flag | bitmask | `cis-1.6/node.yaml` |
| node | `4.1.4` | Ensure that the proxy kubeconfig file ownership is set to root:root (Manual) | no | 2 | flag |  | `cis-1.6/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.6/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.6/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.6/node.yaml` |
| node | `4.1.9` | Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6/node.yaml` |
| node | `4.2.1` | Ensure that the anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.6/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.6/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Manual) | no | 2 | flag, path | eq | `cis-1.6/node.yaml` |
| node | `4.2.12` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.6/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.6/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.6/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.6/node.yaml` |
| node | `4.2.4` | Ensure that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.6/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.6/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Automated) | yes | 1 | flag, path | eq | `cis-1.6/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.6/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.6/node.yaml` |
| node | `4.2.9` | Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture (Manual) | no | 1 | flag, path | eq | `cis-1.6/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.1` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of containers wishing to share the host process ID namespace (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host IPC namespace (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host network namespace (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers with allowPrivilegeEscalation (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of root containers (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of containers with the NET_RAW capability (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with added capabilities (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports Network Policies (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have Network Policies defined (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.4.1` | Prefer using secrets as files over secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your pod definitions (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.7.3` | Apply Security Context to Your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.6/policies.yaml` |

## cis-1.6-k3s

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.6-k3s/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.6-k3s/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate if use etcd as database (Automated) | yes | 2 | flag |  | `cis-1.6-k3s/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6-k3s/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | flag |  | `cis-1.6-k3s/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive if etcd is used (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd if etcd is used (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.13` | Ensure that the admin.kubeconfig file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/master.yaml` |
| master | `1.1.14` | Ensure that the admin.kubeconfig file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.kubeconfig file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.kubeconfig file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.17` | Ensure that the cloud-controller.kubeconfig file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/master.yaml` |
| master | `1.1.18` | Ensure that the $controllermanagerkubeconfig file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | no | 1 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.6-k3s/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.6-k3s/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.6-k3s/master.yaml` |
| master | `1.2.16` | Ensure that the admission control plugin PodSecurityPolicy is set (Automated) | yes | 1 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.2.17` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.2.18` | Ensure that the --insecure-bind-address argument is not set (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.19` | Ensure that the --insecure-port argument is set to 0 (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.2.2` | Ensure that the --basic-auth-file argument is not set (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.20` | Ensure that the --secure-port argument is not set to 0 (Automated) | yes | 2 | flag | gt | `cis-1.6-k3s/master.yaml` |
| master | `1.2.21` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.2.22` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.23` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.6-k3s/master.yaml` |
| master | `1.2.24` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.6-k3s/master.yaml` |
| master | `1.2.25` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.6-k3s/master.yaml` |
| master | `1.2.26` | Ensure that the --request-timeout argument is set as appropriate (Automated) | yes | 2 | flag | lte | `cis-1.6-k3s/master.yaml` |
| master | `1.2.27` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.2.28` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.29` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.3` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.30` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.31` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.32` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.33` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.34` | Ensure that encryption providers are appropriately configured (Manual) | yes | 0 |  |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.35` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.6-k3s/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-https argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.6` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.6-k3s/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.2.9` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.6-k3s/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.6-k3s/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.6-k3s/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.1.10` | Ensure that the kubelet --config configuration file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.1.3` | If proxy kubeproxy.kubeconfig file exists ensure permissions are set to 644 or more restrictive (Automated) | yes | 2 | flag | bitmask | `cis-1.6-k3s/node.yaml` |
| node | `4.1.4` | Ensure that the proxy kubeconfig file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.6-k3s/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 644 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.6-k3s/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/node.yaml` |
| node | `4.1.9` | Ensure that the kubelet --config configuration file has permissions set to 644 or more restrictive (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.1` | Ensure that the anonymous-auth argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/node.yaml` |
| node | `4.2.10` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.11` | Ensure that the --rotate-certificates argument is not set to false (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.12` | Verify that the RotateKubeletServerCertificate argument is set to true (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.13` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.6-k3s/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.4` | Ensure that the --read-only-port argument is set to 0 (Automated) | no | 2 | flag | eq | `cis-1.6-k3s/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Automated) | yes | 2 | flag | noteq | `cis-1.6-k3s/node.yaml` |
| node | `4.2.6` | Ensure that the --protect-kernel-defaults argument is set to true (Automated) | yes | 1 | flag | eq | `cis-1.6-k3s/node.yaml` |
| node | `4.2.7` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.6-k3s/node.yaml` |
| node | `4.2.8` | Ensure that the --hostname-override argument is not set (Not Applicable) | no | 0 |  |  | `cis-1.6-k3s/node.yaml` |
| node | `4.2.9` | Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture (Manual) | no | 1 | flag | eq | `cis-1.6-k3s/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.1` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of containers wishing to share the host process ID namespace (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host IPC namespace (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host network namespace (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers with allowPrivilegeEscalation (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of root containers (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of containers with the NET_RAW capability (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with added capabilities (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports Network Policies (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have Network Policies defined (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.4.1` | Prefer using secrets as files over secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your pod definitions (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.7.3` | Apply Security Context to Your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.6-k3s/policies.yaml` |

## cis-1.7

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.7/controlplane.yaml` |
| controlplane | `3.1.2` | Service account token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.7/controlplane.yaml` |
| controlplane | `3.1.3` | Bootstrap token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.7/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.7/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.7/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.7/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.7/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.7/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.7/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.7/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.7/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.7/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.7/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.7/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.7/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.7/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.7/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.7/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.7/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.7/master.yaml` |
| master | `1.2.16` | Ensure that the --secure-port argument is not set to 0 - NoteThis recommendation is obsolete and will be deleted per the consensus process (Manual) | no | 2 | flag | gt | `cis-1.7/master.yaml` |
| master | `1.2.17` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.7/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.7/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.20` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.7/master.yaml` |
| master | `1.2.21` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.7/master.yaml` |
| master | `1.2.22` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.7/master.yaml` |
| master | `1.2.23` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.7/master.yaml` |
| master | `1.2.24` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.25` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.26` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.27` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.28` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.29` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is set (Manual) | no | 1 | flag | has | `cis-1.7/master.yaml` |
| master | `1.2.30` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.7/master.yaml` |
| master | `1.2.31` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.7/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.7/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.7/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.7/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.7/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.7/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.7/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.7/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.7/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.7/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.7/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.7/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.7/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.7/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.7/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.7/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.7/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.7/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.7/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.7/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.7/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.7/node.yaml` |
| node | `4.2.10` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.7/node.yaml` |
| node | `4.2.11` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.7/node.yaml` |
| node | `4.2.12` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.7/node.yaml` |
| node | `4.2.13` | Ensure that a limit is set on pod PIDs (Manual) | no | 1 | flag, path |  | `cis-1.7/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.7/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.7/node.yaml` |
| node | `4.2.4` | Verify that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.7/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.7/node.yaml` |
| node | `4.2.6` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.7/node.yaml` |
| node | `4.2.7` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.7/node.yaml` |
| node | `4.2.8` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.7/node.yaml` |
| node | `4.2.9` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.7/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.10` | Minimize access to the proxy sub-resource of nodes (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.11` | Minimize access to the approval sub-resource of certificatesigningrequests objects (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.12` | Minimize access to webhook configuration objects (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.13` | Minimize access to the service account token creation (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.1.9` | Minimize access to create persistent volumes (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Automated) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.7/policies.yaml` |

## cis-1.8

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.8/controlplane.yaml` |
| controlplane | `3.1.2` | Service account token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.8/controlplane.yaml` |
| controlplane | `3.1.3` | Bootstrap token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.8/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.8/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.8/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.8/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.8/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.8/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.8/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.8/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.8/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.8/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.13` | Ensure that the admin.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.14` | Ensure that the admin.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.8/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.8/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.8/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.8/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual) | no | 2 | flag | has | `cis-1.8/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.8/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.8/master.yaml` |
| master | `1.2.15` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.8/master.yaml` |
| master | `1.2.16` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.8/master.yaml` |
| master | `1.2.17` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.8/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.8/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.20` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.8/master.yaml` |
| master | `1.2.21` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.8/master.yaml` |
| master | `1.2.22` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.8/master.yaml` |
| master | `1.2.23` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.24` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.25` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.26` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.27` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.28` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.29` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.8/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is set (Manual) | no | 1 | flag | has | `cis-1.8/master.yaml` |
| master | `1.2.30` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.8/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.8/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.8/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.8/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.8/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.8/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.8/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.8/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.8/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.8/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.8/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.8/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.8/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.8/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.8/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.8/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.8/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.8/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.8/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.8/node.yaml` |
| node | `4.2.10` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.8/node.yaml` |
| node | `4.2.11` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.8/node.yaml` |
| node | `4.2.12` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.8/node.yaml` |
| node | `4.2.13` | Ensure that a limit is set on pod PIDs (Manual) | no | 1 | flag, path |  | `cis-1.8/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.8/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.8/node.yaml` |
| node | `4.2.4` | Verify that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.8/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.8/node.yaml` |
| node | `4.2.6` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.8/node.yaml` |
| node | `4.2.7` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.8/node.yaml` |
| node | `4.2.8` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.8/node.yaml` |
| node | `4.2.9` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.8/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.10` | Minimize access to the proxy sub-resource of nodes (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.11` | Minimize access to the approval sub-resource of certificatesigningrequests objects (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.12` | Minimize access to webhook configuration objects (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.13` | Minimize access to the service account token creation (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used. (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.1.9` | Minimize access to create persistent volumes (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Automated) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.8/policies.yaml` |

## cis-1.9

| Target | Control ID | Title | Scored | Test items | Sources | Comparators | Source file |
|---|---|---|---|---:|---|---|---|
| controlplane | `3.1.1` | Client certificate authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.9/controlplane.yaml` |
| controlplane | `3.1.2` | Service account token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.9/controlplane.yaml` |
| controlplane | `3.1.3` | Bootstrap token authentication should not be used for users (Manual) | no | 0 |  |  | `cis-1.9/controlplane.yaml` |
| controlplane | `3.2.1` | Ensure that a minimal audit policy is created (Manual) | no | 1 | flag |  | `cis-1.9/controlplane.yaml` |
| controlplane | `3.2.2` | Ensure that the audit policy covers key security concerns (Manual) | no | 0 |  |  | `cis-1.9/controlplane.yaml` |
| etcd | `2.1` | Ensure that the --cert-file and --key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.9/etcd.yaml` |
| etcd | `2.2` | Ensure that the --client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.9/etcd.yaml` |
| etcd | `2.3` | Ensure that the --auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.9/etcd.yaml` |
| etcd | `2.4` | Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated) | yes | 2 | env, flag |  | `cis-1.9/etcd.yaml` |
| etcd | `2.5` | Ensure that the --peer-client-cert-auth argument is set to true (Automated) | yes | 1 | env, flag | eq | `cis-1.9/etcd.yaml` |
| etcd | `2.6` | Ensure that the --peer-auto-tls argument is not set to true (Automated) | yes | 2 | env, flag | eq | `cis-1.9/etcd.yaml` |
| etcd | `2.7` | Ensure that a unique Certificate Authority is used for etcd (Manual) | no | 1 | env, flag |  | `cis-1.9/etcd.yaml` |
| master | `1.1.1` | Ensure that the API server pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.10` | Ensure that the Container Network Interface file ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.11` | Ensure that the etcd data directory permissions are set to 700 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.12` | Ensure that the etcd data directory ownership is set to etcd:etcd (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.13` | Ensure that the default administrative credential file permissions are set to 600 (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.14` | Ensure that the default administrative credential file ownership is set to root:root (Automated) | yes | 1 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.1.15` | Ensure that the scheduler.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.16` | Ensure that the scheduler.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.17` | Ensure that the controller-manager.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.18` | Ensure that the controller-manager.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.19` | Ensure that the Kubernetes PKI directory and file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.2` | Ensure that the API server pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.20` | Ensure that the Kubernetes PKI certificate file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.21` | Ensure that the Kubernetes PKI key file permissions are set to 600 (Manual) | no | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.3` | Ensure that the controller manager pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.4` | Ensure that the controller manager pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.5` | Ensure that the scheduler pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.6` | Ensure that the scheduler pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.7` | Ensure that the etcd pod specification file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.1.8` | Ensure that the etcd pod specification file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.1.9` | Ensure that the Container Network Interface file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.9/master.yaml` |
| master | `1.2.1` | Ensure that the --anonymous-auth argument is set to false (Manual) | no | 1 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.2.10` | Ensure that the admission control plugin AlwaysAdmit is not set (Automated) | yes | 2 | flag | nothave | `cis-1.9/master.yaml` |
| master | `1.2.11` | Ensure that the admission control plugin AlwaysPullImages is set (Manual) | no | 1 | flag | has | `cis-1.9/master.yaml` |
| master | `1.2.12` | Ensure that the admission control plugin ServiceAccount is set (Automated) | yes | 2 | flag | nothave | `cis-1.9/master.yaml` |
| master | `1.2.13` | Ensure that the admission control plugin NamespaceLifecycle is set (Automated) | yes | 2 | flag | nothave | `cis-1.9/master.yaml` |
| master | `1.2.14` | Ensure that the admission control plugin NodeRestriction is set (Automated) | yes | 1 | flag | has | `cis-1.9/master.yaml` |
| master | `1.2.15` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.2.16` | Ensure that the --audit-log-path argument is set (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.17` | Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.9/master.yaml` |
| master | `1.2.18` | Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.9/master.yaml` |
| master | `1.2.19` | Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated) | yes | 1 | flag | gte | `cis-1.9/master.yaml` |
| master | `1.2.2` | Ensure that the --token-auth-file parameter is not set (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.20` | Ensure that the --request-timeout argument is set as appropriate (Manual) | no | 0 |  |  | `cis-1.9/master.yaml` |
| master | `1.2.21` | Ensure that the --service-account-lookup argument is set to true (Automated) | yes | 2 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.2.22` | Ensure that the --service-account-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.23` | Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.24` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.25` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.26` | Ensure that the --etcd-cafile argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.27` | Ensure that the --encryption-provider-config argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.28` | Ensure that encryption providers are appropriately configured (Manual) | no | 1 | flag | valid_elements | `cis-1.9/master.yaml` |
| master | `1.2.29` | Ensure that the API Server only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag | valid_elements | `cis-1.9/master.yaml` |
| master | `1.2.3` | Ensure that the --DenyServiceExternalIPs is set (Manual) | no | 1 | flag | has | `cis-1.9/master.yaml` |
| master | `1.2.4` | Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated) | yes | 2 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.5` | Ensure that the --kubelet-certificate-authority argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.2.6` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag | nothave | `cis-1.9/master.yaml` |
| master | `1.2.7` | Ensure that the --authorization-mode argument includes Node (Automated) | yes | 1 | flag | has | `cis-1.9/master.yaml` |
| master | `1.2.8` | Ensure that the --authorization-mode argument includes RBAC (Automated) | yes | 1 | flag | has | `cis-1.9/master.yaml` |
| master | `1.2.9` | Ensure that the admission control plugin EventRateLimit is set (Manual) | no | 1 | flag | has | `cis-1.9/master.yaml` |
| master | `1.3.1` | Ensure that the --terminated-pod-gc-threshold argument is set as appropriate (Manual) | no | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.3.2` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.3.3` | Ensure that the --use-service-account-credentials argument is set to true (Automated) | yes | 1 | flag | noteq | `cis-1.9/master.yaml` |
| master | `1.3.4` | Ensure that the --service-account-private-key-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.3.5` | Ensure that the --root-ca-file argument is set as appropriate (Automated) | yes | 1 | flag |  | `cis-1.9/master.yaml` |
| master | `1.3.6` | Ensure that the RotateKubeletServerCertificate argument is set to true (Automated) | yes | 2 | flag | nothave | `cis-1.9/master.yaml` |
| master | `1.3.7` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.4.1` | Ensure that the --profiling argument is set to false (Automated) | yes | 1 | flag | eq | `cis-1.9/master.yaml` |
| master | `1.4.2` | Ensure that the --bind-address argument is set to 127.0.0.1 (Automated) | yes | 2 | flag | eq | `cis-1.9/master.yaml` |
| node | `4.1.1` | Ensure that the kubelet service file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/node.yaml` |
| node | `4.1.10` | If the kubelet config.yaml configuration file is being used validate file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/node.yaml` |
| node | `4.1.2` | Ensure that the kubelet service file ownership is set to root:root (Automated) | yes | 2 | flag |  | `cis-1.9/node.yaml` |
| node | `4.1.3` | If proxy kubeconfig file exists ensure permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.9/node.yaml` |
| node | `4.1.4` | If proxy kubeconfig file exists ensure ownership is set to root:root (Manual) | no | 1 | flag |  | `cis-1.9/node.yaml` |
| node | `4.1.5` | Ensure that the --kubeconfig kubelet.conf file permissions are set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/node.yaml` |
| node | `4.1.6` | Ensure that the --kubeconfig kubelet.conf file ownership is set to root:root (Automated) | yes | 1 | flag |  | `cis-1.9/node.yaml` |
| node | `4.1.7` | Ensure that the certificate authorities file permissions are set to 600 or more restrictive (Manual) | no | 1 | flag | bitmask | `cis-1.9/node.yaml` |
| node | `4.1.8` | Ensure that the client certificate authorities file ownership is set to root:root (Manual) | no | 1 | flag | eq | `cis-1.9/node.yaml` |
| node | `4.1.9` | If the kubelet config.yaml configuration file is being used validate permissions set to 600 or more restrictive (Automated) | yes | 1 | flag | bitmask | `cis-1.9/node.yaml` |
| node | `4.2.1` | Ensure that the --anonymous-auth argument is set to false (Automated) | yes | 1 | flag, path | eq | `cis-1.9/node.yaml` |
| node | `4.2.10` | Ensure that the --rotate-certificates argument is not set to false (Automated) | yes | 2 | flag, path | eq | `cis-1.9/node.yaml` |
| node | `4.2.11` | Verify that the RotateKubeletServerCertificate argument is set to true (Manual) | no | 2 | flag, path | nothave | `cis-1.9/node.yaml` |
| node | `4.2.12` | Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers (Manual) | no | 1 | flag, path | valid_elements | `cis-1.9/node.yaml` |
| node | `4.2.13` | Ensure that a limit is set on pod PIDs (Manual) | no | 1 | flag, path |  | `cis-1.9/node.yaml` |
| node | `4.2.2` | Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated) | yes | 1 | flag, path | nothave | `cis-1.9/node.yaml` |
| node | `4.2.3` | Ensure that the --client-ca-file argument is set as appropriate (Automated) | yes | 1 | flag, path |  | `cis-1.9/node.yaml` |
| node | `4.2.4` | Verify that the --read-only-port argument is set to 0 (Manual) | no | 2 | flag, path | eq | `cis-1.9/node.yaml` |
| node | `4.2.5` | Ensure that the --streaming-connection-idle-timeout argument is not set to 0 (Manual) | no | 2 | flag, path | noteq | `cis-1.9/node.yaml` |
| node | `4.2.6` | Ensure that the --make-iptables-util-chains argument is set to true (Automated) | yes | 2 | flag, path | eq | `cis-1.9/node.yaml` |
| node | `4.2.7` | Ensure that the --hostname-override argument is not set (Manual) | no | 1 | flag |  | `cis-1.9/node.yaml` |
| node | `4.2.8` | Ensure that the eventRecordQPS argument is set to a level which ensures appropriate event capture (Manual) | no | 2 | flag, path | gte | `cis-1.9/node.yaml` |
| node | `4.2.9` | Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate (Manual) | no | 2 | flag, path |  | `cis-1.9/node.yaml` |
| node | `4.3.1` | Ensure that the kube-proxy metrics service is bound to localhost (Automated) | yes | 2 | flag, path | has | `cis-1.9/node.yaml` |
| policies | `5.1.1` | Ensure that the cluster-admin role is only used where required (Automated) | no | 1 | flag | eq | `cis-1.9/policies.yaml` |
| policies | `5.1.10` | Minimize access to the proxy sub-resource of nodes (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.1.11` | Minimize access to the approval sub-resource of certificatesigningrequests objects (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.1.12` | Minimize access to webhook configuration objects (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.1.13` | Minimize access to the service account token creation (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.1.2` | Minimize access to secrets (Automated) | no | 1 | flag | eq | `cis-1.9/policies.yaml` |
| policies | `5.1.3` | Minimize wildcard use in Roles and ClusterRoles (Automated) | no | 2 | flag | eq | `cis-1.9/policies.yaml` |
| policies | `5.1.4` | Minimize access to create pods (Automated) | no | 1 | flag | eq | `cis-1.9/policies.yaml` |
| policies | `5.1.5` | Ensure that default service accounts are not actively used (Automated) | no | 1 | flag | eq | `cis-1.9/policies.yaml` |
| policies | `5.1.6` | Ensure that Service Account Tokens are only mounted where necessary (Automated) | no | 1 | flag | eq | `cis-1.9/policies.yaml` |
| policies | `5.1.7` | Avoid use of system:masters group (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.1.8` | Limit use of the Bind, Impersonate and Escalate permissions in the Kubernetes cluster (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.1.9` | Minimize access to create persistent volumes (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.1` | Ensure that the cluster has at least one active policy control mechanism in place (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.10` | Minimize the admission of containers with capabilities assigned (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.11` | Minimize the admission of Windows HostProcess containers (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.12` | Minimize the admission of HostPath volumes (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.13` | Minimize the admission of containers which use HostPorts (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.2` | Minimize the admission of privileged containers (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.3` | Minimize the admission of containers wishing to share the host process ID namespace (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.4` | Minimize the admission of containers wishing to share the host IPC namespace (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.5` | Minimize the admission of containers wishing to share the host network namespace (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.6` | Minimize the admission of containers with allowPrivilegeEscalation (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.7` | Minimize the admission of root containers (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.8` | Minimize the admission of containers with the NET_RAW capability (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.2.9` | Minimize the admission of containers with added capabilities (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.3.1` | Ensure that the CNI in use supports NetworkPolicies (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.3.2` | Ensure that all Namespaces have NetworkPolicies defined (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.4.1` | Prefer using Secrets as files over Secrets as environment variables (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.4.2` | Consider external secret storage (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.5.1` | Configure Image Provenance using ImagePolicyWebhook admission controller (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.7.1` | Create administrative boundaries between resources using namespaces (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.7.2` | Ensure that the seccomp profile is set to docker/default in your Pod definitions (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.7.3` | Apply SecurityContext to your Pods and Containers (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |
| policies | `5.7.4` | The default namespace should not be used (Manual) | no | 0 |  |  | `cis-1.9/policies.yaml` |

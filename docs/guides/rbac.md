---
title: rbac | Stash
description: rbac of Stash
menu:
  product_stash_0.5.1:
    identifier: rbac-stash
    name: RBAC
    parent: getting-started
    weight: 45
product_name: stash
menu_name: product_stash_0.5.1
section_menu_id: getting-started
url: /products/stash/0.5.1/getting-started/rbac/
aliases:
  - /products/stash/0.5.1/rbac/
---
> New to Stash? Please start [here](/docs/guides/README.md).

# Configuring RBAC

To use Stash in a RBAC enabled cluster, [install Stash](/docs/setup/install.md) with RBAC options. This creates a ClusterRole named `stash-sidecar`.

Sidecar container added to workloads makes various calls to Kubernetes api. ServiceAccounts used with Deployment, ReplicaSet, DaemonSet and ReplicationController workloads are automatically bound to `stash-sidecar` ClusterRole by Stash operator. Users should manually add the following RoleBinding to service accounts used with StatefulSet workloads to authorize these api calls.

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: <statefulset-name>-stash-sidecar
  namespace: <statefulset-namespace>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: stash-sidecar
subjects:
- kind: ServiceAccount
  name: <statefulset-sa>
  namespace: <statefulset-namespace>
```

You can find full working examples [here](/docs/guides/workloads.md).

## Next Steps

- Learn how to use Stash to backup a Kubernetes deployment [here](/docs/guides/backup.md).
- Learn about the details of Restic CRD [here](/docs/concepts/restic.md).
- To restore a backup see [here](/docs/guides/restore.md).
- Learn about the details of Recovery CRD [here](/docs/concepts/recovery.md).
- To run backup in offline mode see [here](/docs/guides/offline_backup.md)
- See the list of supported backends and how to configure them [here](/docs/guides/backends.md).
- See working examples for supported workload types [here](/docs/guides/workloads.md).
- Thinking about monitoring your backup operations? Stash works [out-of-the-box with Prometheus](/docs/guides/monitoring.md).
- Learn about how to configure Stash operator as workload initializer [here](/docs/guides/initializer.md).
- Want to hack on Stash? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
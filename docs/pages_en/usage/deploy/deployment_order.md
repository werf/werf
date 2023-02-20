---
title: Deployment order
permalink: usage/deploy/deployment_order.html
---

## Deployment stages

Kubernetes resources are deployed in the following stages:

1. Deploying `CustomResourceDefinitions` from the `crds` directories of the included charts.

2. Deploying `pre-install`, `pre-upgrade` or `pre-rollback` hooks one hook at a time, starting with hooks with less weight to greater weight. If a hook is dependent on an external resource, it will only be deployed once that resource is ready.

3. Deploying basic resources: the resources with the same weight are combined into groups (resources without a specified weight are assigned a weight of 0) and deployed one group at a time, starting with groups with lower-weight resources to groups with higher-weight resources. If a resource in a group depends on an external resource, the group will be deployed only after that resource is ready.

4. Deploying `post-install`, `post-upgrade` or `post-rollback` hooks one hook at a time, starting with hooks with less weight to ones with greater weight. If a hook is dependent on an external resource, it will be deployed only after that resource is ready.

## Deploying CustomResourceDefinitions

To deploy CustomResourceDefinitions, put the CRD manifests in the `crds/*.yaml` non-template files in any of the included charts. During the following deployment, these CRDs will be deployed first, with hooks and core resources following them.

Example:

```yaml
# .helm/crds/crontab.yaml:
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
# ...
spec:
  names:
    kind: CronTab
```

```
# .helm/templates/crontab.yaml:
apiVersion: example.org/v1
kind: CronTab
# ...
```

```shell
werf converge
```

In this case, the CRD for the CronTab resource will be deployed first, followed by the CronTab resource.

## Changing the order in which resources are deployed (werf only)

By default, werf combines all the main resources ("main" means those are not hooks or CRDs from `crds/*.yaml`) into one group, creates resources for that group, and then tracks their readiness.

The resources for the group are created in the following order:

- Namespace;
- NetworkPolicy;
- ResourceQuota;
- LimitRange;
- PodSecurityPolicy;
- PodDisruptionBudget;
- ServiceAccount;
- Secret;
- SecretList;
- ConfigMap;
- StorageClass;
- PersistentVolume;
- PersistentVolumeClaim;
- CustomResourceDefinition;
- ClusterRole;
- ClusterRoleList;
- ClusterRoleBinding;
- ClusterRoleBindingList;
- Role;
- RoleList;
- RoleBinding;
- RoleBindingList;
- Service;
- DaemonSet;
- Pod;
- ReplicationController;
- ReplicaSet;
- Deployment;
- HorizontalPodAutoscaler;
- StatefulSet;
- Job;
- CronJob;
- Ingress;
- APIService.

Readiness tracking is enabled for all resources in the group simultaneously as soon as *all* resources in the group are created.

To change the order in which resources are deployed, you can create *new resource groups* by assigning resources a *weight* other than the default `0`. All resources with the same weight are combined into their respective groups, and then the resource groups are deployed sequentially starting with the group with the lesser weight and progressing to the greater weight, for example:

```
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
  annotations:
    werf.io/weight: "-1"
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app1
  annotations:
    werf.io/weight: "1"
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2
  annotations:
    werf.io/weight: "1"
# ...
```

```shell
werf converge
```

In this case, the `database` resource was deployed first, followed by `database-migrations`, and then `app1` and `app2` were deployed in parallel.

## Running tasks before/after installing, upgrading, rolling back or deleting a release

To deploy certain resources before or after installing, upgrading, rolling back or deleting a release, convert the resource to a *hook* by adding the `helm.sh/hook` annotation to it, for example:

```
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-initialization
  annotations:
    helm.sh/hook: pre-install
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
```

```shell
werf converge
```

In this case, the `database-initialization` resource will be deployed during the first-time *installation* of the release, while the `myapp` resource will be deployed whenever the release is installed, upgraded, or rolled back.

The `helm.sh/hook` annotation sets the resource as a hook and defines the conditions on which the resource should be deployed (you can specify several comma-separated conditions). The available conditions for deploying a hook are listed below:

* `pre-install` — during the release installation *before* installing the main resources;

* `pre-upgrade` — when upgrading the release *before* upgrading the main resources;

* `pre-rollback` — when rolling back the release *before* rolling back the main resources;

* `pre-delete` — when deleting the release *before* deleting the main resources;

* `post-install` — when installing the release *after* installing the main resources;

* `post-upgrade` — when upgrading the release *after* upgrading the main resources;

* `post-rollback` — when rolling back the release *after* rolling back the main resources;

* `post-delete` — when deleting the release *after* deleting the main resources;

To set the order in which hooks are deployed, assign them different *weights* (the default is `0`), so that hooks are deployed sequentially starting with a hook with a lower weight and proceeding to ones with a higher weight, for example:

```
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: first
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "-1"
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: second
  annotations:
    helm.sh/hook: pre-install
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: third
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "1"
# ...
```

```shell
werf converge
```

In this case, the `first` hook will be deployed first, followed by the `second` hook, and then the `third` hook.

**By default, redeploying the same hook will delete the old hook in the cluster just before deploying the new one.** You can change the deletion policy for the old hook using the `helm.sh/hook-delete-policy` annotation. It can have the following values:

- `hook-succeeded` — delete the new hook once it has been successfully deployed; do not delete it if it has failed to deploy;

- `hook-failed` — delete the new hook once it has failed to deploy; do not delete it if it has been deployed successfully;

- `before-hook-creation` — (default) delete the old hook right before creating a new one.

## Waiting for non-release resources to be ready (werf only)

The resources deployed as part of the current release may depend on resources that do not belong to this release. werf can wait for these external resources to be ready — you just need to add the `<name>.external-dependency.werf.io/resource` annotation as follows:

```
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
# ...
```

```shell
werf converge
```

The `myapp` deployment will start deploying only after the `my-dynamic-vault-secret` (it is created automatically by the operator in the cluster) has been created and is ready.

And here is how you can configure werf to wait for multiple external resources to be ready:

```
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
    database.external-dependency.werf.io/resource: statefulset/my-database
# ...
```

By default, werf looks for the external resource in the release Namespace (unless, of course, the resource is a cluster-wide). You can change the Namespace of the external resource by attaching the `<name>.external-dependency.werf.io/namespace` annotation to it:

```
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
    secret.external-dependency.werf.io/namespace: my-namespace 
```

*Note that all other release resources with the same weight will also be waiting for the external resource to be ready, since resources are grouped by weight and deployed as groups.*

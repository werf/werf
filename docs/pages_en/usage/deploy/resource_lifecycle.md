---
title: Resource lifecycle
permalink: usage/deploy/resource_lifecycle.html
---

## Conditional resource deployment

To deploy certain resources only for a specific deploy type (install, upgrade, rollback, uninstall) or at a specific deployment stage (pre, main, post stages), use the `werf.io/deploy-on` annotation, which is inspired by the `helm.sh/hook` annotation.

Available values for `werf.io/deploy-on` are:
* `pre-install`, `install`, `post-install` — render only on release installation
* `pre-upgrade`, `upgrade`, `post-upgrade` — render only on release upgrade
* `pre-rollback`, `rollback`, `post-rollback` — render only on release rollback
* `pre-delete`, `delete`, `post-delete` — render only on release deletion

The default value for general resources is `install,upgrade,rollback`, while for hooks it is populated from `helm.sh/hook`.

Example:

```yaml
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-initialization
  annotations:
    werf.io/deploy-on: pre-install
    werf.io/delete-policy: before-creation
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
# ...
```

In this case, the `database-initialization` resource will be deployed during the first-time installation of the release and before `myapp`, while the `myapp` resource will be deployed in the main deployment stage whenever the release is installed, upgraded, or rolled back.

## Resource ownership

The `werf.io/ownership` annotation defines how resource deletions are handled and how release annotations are managed. Allowed values are:
 * `release`: the resource is deleted if removed from the chart or when the release is uninstalled, and release annotations of the resource are applied/validated during deploy.
 * `anyone`: the opposite of `release`, the resource is never deleted on uninstall or when removed from the chart, and release annotations are not applied/validated during deploy.

General resources have `release` ownership by default, while hooks and CRDs from the `crds` directory have `anyone` ownership.

Example:

```yaml
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
  annotations:
    werf.io/ownership: "anyone"
# ...
```

Here, ownership `anyone` makes this Job behave a lot like a Helm hook, meaning that it will not be deleted on uninstall or when removed from the chart, and its release annotations will not be applied/validated during deploy.

## Resource delete policies

### werf.io/delete-policy

The `werf.io/delete-policy` annotation controls resource deletions during its deployment and is inspired by `helm.sh/hook-delete-policy`. Allowed values:
* `before-creation`: the resource is always recreated
* `before-creation-if-immutable`: the resource is recreated only if we got the `field is immutable` error trying to update the resource
* `succeeded`: the resource is deleted after its readiness check succeeds
* `failed`: the resource is deleted if its readiness check fails

Multiple values can be specified at once. By default, general resources have no delete policy, while hooks have values from `helm.sh/hook-delete-policy` mapped to `werf.io/delete-policy`.

Example:

```yaml
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
  annotations:
    werf.io/delete-policy: before-creation,succeeded
# ...
```

Here, the `database-migrations` Job is always recreated and then deleted after it becomes ready.

### werf.io/delete-propagation

The `werf.io/delete-propagation` annotation controls the propagation policy used when deleting resources. Allowed values are:
* `Foreground`: delete the resource after deleting all of its dependents.
* `Background`: delete the resource immediately, and delete all of its dependents in the background.
* `Orphan`: delete the resource, but leave all of its dependents untouched.

Example:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/delete-propagation: Background
# ...
```

Here, when the `myapp` Deployment is deleted, its dependents (ReplicaSets, Pods, etc.) will be deleted in the background.

By default, resources are deleted with the `Foreground` propagation policy.

### helm.sh/resource-policy

The annotation `helm.sh/resource-policy: keep` forbids any resource deletion from happening. The resource can never be deleted for any reason when this annotation is present. This annotation is also respected on the resource in the cluster, even if it is not present in the chart.

Example:

```yaml
# .helm/templates/example.yaml:
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  annotations:
    helm.sh/resource-policy: keep
# ...
```

Here, the `my-pvc` PersistentVolumeClaim will never be deleted for any reason.

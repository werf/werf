---
title: Deployment order
permalink: usage/deploy/deployment_order.html
---

## Deployment stages

Kubernetes resources are deployed in the following stages:

1. Deploy `CustomResourceDefinitions`.
1. Deploy `pre-install`, `pre-upgrade`, `pre-rollback` resources, from low to high weight (`werf.io/weight`, `helm.sh/hook-weight`).
1. Deploy `install`, `upgrade`, `rollback` resources, from low to high weight.
1. Deploy `post-install`, `post-upgrade`, `post-rollback` resources, from low to high weight.

Resources with the same weight are grouped and deployed concurrently. Default weight is 0. Resources with the `werf.io/deploy-dependency-<name>` annotation are deployed as soon as their dependencies are satisfied, but within their stage (pre, main or post).

## Deploying CustomResourceDefinitions

To deploy CustomResourceDefinitions, put their manifests into `crds/*.yaml` non-template files in any of the included charts. During deployment, CRDs are always deployed before any other resources.

Example:

```yaml
# .helm/crds/crontab.yaml:
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  names:
    kind: CronTab
# ...
```

```yaml
# .helm/templates/crontab.yaml:
apiVersion: example.org/v1
kind: CronTab
# ...
```

In this case, the CRD for the CronTab resource will be deployed first, followed by the CronTab resource.

## Resource ordering

### Ordering via weights (werf-only)

The annotation `werf.io/weight` can be used to set resource ordering during deployment. Resources have weight 0 by default. Resources with a lower weight are deployed before resources with a higher weight. If resources have the same weight, they are deployed in parallel.

Take a look at the example:

```yaml
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

In this case, the `database` resource is deployed first, followed by `database-migrations`, and then `app1` and `app2` are deployed in parallel.

### Ordering via dependencies (werf only)

The annotation `werf.io/deploy-dependency-<name>` can be used to set resource ordering during deployment. The resource with such an annotation will be deployed as soon as all its dependencies are satisfied, but within its stage (pre, main or post). The resource weight is ignored when this annotation is used.

For example:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
  annotations:
    werf.io/deploy-dependency-db: state=ready,kind=StatefulSet,name=database
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app1
  annotations:
    werf.io/deploy-dependency-migrations: state=ready,kind=Job,name=database-migrations
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2
  annotations:
    werf.io/deploy-dependency-migrations: state=ready,kind=Job,name=database-migrations
# ...
```

In this case, the `database` resource is deployed first, followed by `database-migrations`, and then `app1` and `app2` are deployed in parallel.

Check out all capabilities of this annotation [here]({{ "/reference/deploy_annotations.html#resource-dependencies" | true_relative_url }}).

This is a more flexible and effective way to set the order of resource deployments in comparison to `werf.io/weight` and other methods, as it allows you to deploy resources in a graph-like order.

### Deletion ordering via dependencies (werf only)

Annotation `werf.io/delete-dependency-<name>` can be used to set resource ordering during deletion. The resource with such an annotation will be deleted only after all its dependencies are deleted.

For example:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
# ...
---
apiVersion: v1
kind: Service
metadata:
  name: app
  annotations:
    werf.io/delete-dependency-ingress: state=absent,kind=Ingress,group=networking.k8s.io,version=v1,name=app
# ...
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
  annotations:
    werf.io/delete-dependency-service: state=absent,kind=Service,version=v1,name=app
# ...
```

In this case, the `Ingress` will be deleted first, then the `Service`, and only then the `Deployment`.

Check out all capabilities of this annotation [here]({{ "/reference/deploy_annotations.html#delete-dependencies" | true_relative_url }}).

## Waiting for non-release resources to be ready (werf only)

The resources deployed as part of the current release may depend on resources that do not belong to this release. werf can wait for these external resources to be ready — you just need to add the `<name>.external-dependency.werf.io/resource` annotation as follows:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
# ...
```

The `myapp` deployment will start deploying only after the `my-dynamic-vault-secret` (which is created automatically by the operator in the cluster) has been created and is ready.

Here is how you can configure werf to wait for multiple external resources to be ready:

```yaml
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

By default, werf looks for the external resource in the release Namespace (unless, of course, the resource is cluster-wide). You can change the Namespace of the external resource by attaching the `<name>.external-dependency.werf.io/namespace` annotation to it:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
    secret.external-dependency.werf.io/namespace: my-namespace
```

---
title: Deployment order
permalink: usage/deploy/deployment_order.html
---

When running the `werf converge` command, werf starts the deployment process that includes the following steps:

1. Rendering chart templates into a single list of Kubernetes resource manifests and validating them.
2. Sequentially running `pre-install` or `pre-upgrade` hooks sorted by their weight and tracking each hook to completion while printing its logs in the process.
3. Grouping non-hook Kubernetes resources by their weight and deploying each group sequentially according to its weight: creating/updating/deleting resources and tracking them until they are ready while printing logs in the process.
4. Running `post-install` or `post-upgrade` hooks the same way as `pre-install` and `pre-upgrade` hooks.

## Resource weight

By default, resources are applied and tracked simultaneously because they all have the same initial ([werf.io/weight: 0]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})) (which means that you haven't changed it or deliberately set it to "0"). However, you can change the order in which resources are applied and tracked by setting different weights for them.

Before the deployment phase, werf groups the resources based on their weight, then applies the group with the lowest associated weight and waits until the resources from that group are ready. After all the resources from that group have been successfully deployed, werf proceeds to deploy the resources from the group with the next lowest weight. The process continues until all release resources have been deployed.

> Note that "werf.io/weight" works only for non-Hook resources. For Hooks, use "helm.sh/hook-weight".

Let's look at the example:
```yaml
kind: Job
metadata:
  name: db-migration
---
kind: StatefulSet
metadata:
  name: postgres
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app
```

All of the above resources will be deployed simultaneously because they all have the same default weight (0). But what if the database must be deployed before migrations, and the application should only start after the migrations are completed? Well, there is a solution to this problem! Try this:
```yaml
kind: StatefulSet
metadata:
  name: postgres
  annotations:
    werf.io/weight: "-2"
---
kind: Job
metadata:
  name: db-migration
  annotations:
    werf.io/weight: "-1"
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app
```

In the above example, werf will first deploy the database and wait for it to become ready, then run migrations and wait for them to complete, and then deploy the application and the related service.

Reference:
* [`werf.io/weight`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})

## Helm hooks

The helm hook is an arbitrary Kubernetes resource marked with the `helm.sh/hook` annotation. For example:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```

A lot of various helm hooks come into play during the deploy process. We have already discussed `pre|post-install|upgrade` hooks during the [deploy process steps]({{ "/usage/deploy/deployment_order.html" | true_relative_url }}). These hooks are often used to perform tasks such as migrations (in the case of `pre-upgrade` hooks) or some post-deploy actions. The full list of available hooks can be found in the [helm docs](https://helm.sh/docs/topics/charts_hooks/).

Hooks are sorted in the ascending order specified by the `helm.sh/hook-weight` annotation (hooks with the same weight are sorted by the name). After that, hooks are created and executed sequentially. werf by default recreates the Kubernetes resource for each hook if that resource already exists in the cluster. Hook resources are remained existing in the Kubernetes cluster after execution.

Created hooks resources will not be deleted after completion, unless there is [special annotation `"helm.sh/hook-delete-policy": hook-succeeded,hook-failed`](https://helm.sh/docs/topics/charts_hooks/).

## External dependencies

To make one release resource dependent on another, you can change its [deployment order]({{ "/usage/deploy/deployment_order.html" | true_relative_url }}) so that the dependent one is deployed after the primary one has been successfully deployed. But what if you want a release resource to depend on a resource that is not part of that release or is not even managed by werf (e.g., created by some operator)?

In this case, the [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }}) annotation can help you set the dependency on an external resource. The annotated resource will not be deployed until the defined external dependency is created and ready.

Example:
```yaml
kind: Deployment
metadata:
  name: app
  annotations:
    secret.external-dependency.werf.io/resource: secret/dynamic-vault-secret
```

In the example above, werf will wait for the `dynamic-vault-secret` to be created before proceeding to deploy the `app` deployment. We assume that `dynamic-vault-secret` is created by the operator from your Vault instance and is not managed by werf.

Let's take a look at another example:
```yaml
kind: Deployment
metadata:
  name: service1
  annotations:
    service2.external-dependency.werf.io/resource: deployment/service2
    service2.external-dependency.werf.io/namespace: service2-production
```

Here, werf will not start deploying `service1` until the deployment of `service2` in another namespace has been successfully completed. Note that `service2` can either be deployed as part of another werf release or be managed by some other CI/CD software (i.e., outside of werf's control).

Reference:
* [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }})
* [`<name>.external-dependency.werf.io/namespace`]({{ "/reference/deploy_annotations.html#external-dependency-namespace" | true_relative_url }})

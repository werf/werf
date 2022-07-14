---
title: External dependencies
permalink: advanced/helm/deploy_process/external_dependencies.html
---

To make one release resource dependent on another, you can change its [deployment order]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}) so that the dependent one is deployed after the primary one has been successfully deployed. But what if you want a release resource to depend on a resource that is not part of that release or is not even managed by werf (e.g., created by some operator)?

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

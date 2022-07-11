---
title: External dependencies
permalink: advanced/helm/deploy_process/external_dependencies.html
---

If you want to make one release resource depend on another you can change its [deployment order]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}) to make the latter deploy only after the former is deployed successfully. But what if we want to make a release resource depend on a resource that is not managed as part of our release, or even not managed by werf at all (e.g. created by some operator)?

[`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }}) annotation can define an external dependency on another resource. Until the defined external dependency is created and ready, the annotated resource will not be deployed.

Example:
```yaml
kind: Deployment
metadata:
  name: app
  annotations:
    secret.external-dependency.werf.io/resource: secret/dynamic-vault-secret
```

In this example the `app` deployment will not be deployed and will wait until Secret `dynamic-vault-secret` is created. Lets assume that `dynamic-vault-secret` is created by the operator from your Vault instance and is not managed by werf.

Another example:
```yaml
kind: Deployment
metadata:
  name: service1
  annotations:
    service2.external-dependency.werf.io/resource: deployment/service2
    service2.external-dependency.werf.io/namespace: service2-production
```

Here the deployment of `service1` will not be started until deployment `service2` in another namespace is well and ready. `service2` deployment might be deployed as part of another werf release, or not be deployed by werf at all, but managed using different CI/CD software.

Reference:
* [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }})
* [`<name>.external-dependency.werf.io/namespace`]({{ "/reference/deploy_annotations.html#external-dependency-namespace" | true_relative_url }})

---
title: Deployment order
permalink: advanced/helm/deploy_process/deployment_order.html
---

By default, all resources have [`werf.io/weight: 0`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }}), thus will be applied and tracked concurrently. If you want to specify the order in which resources will be applied and tracked you can specify different weights for your resources.

Before deployment phase resources are grouped based on their weight, then the group with the lowest associated weight is applied to the cluster and tracked until resources in this group are ready. After all resources in group are successfully deployed, the deployment of the resources with the next lowest weight is started. This continues until all release resources are deployed.

> "werf.io/weight" will work only for non-Hook resources. For Hooks use "helm.sh/hook-weight".

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

All of these resources will be deployed simultaneously, since they have default weight 0. But what if we need to deploy database before migrations run and also do not run the app until migrations finished? Try this:
```yaml
kind: StatefulSet
metadata:
  name: postgres
  annotations:
    werf.io/weight: "10"
---
kind: Job
metadata:
  name: db-migration
  annotations:
    werf.io/weight: "20"
---
kind: Deployment
metadata:
  name: app
  annotations:
    werf.io/weight: "30"
---
kind: Service
metadata:
  name: app
  annotations:
    werf.io/weight: "30"
```

This will deploy the database and wait until it's ready, then perform the migrations, and only then deploy the application and its service.

Reference:
* [`werf.io/weight`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})

---
title: Deployment order
permalink: advanced/helm/deploy_process/deployment_order.html
---

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
